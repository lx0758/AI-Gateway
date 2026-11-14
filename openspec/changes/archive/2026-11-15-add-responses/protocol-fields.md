# 协议字段研究：Responses ↔ Chat ↔ Messages 转换基准

> 本文是 design.md D5（转换器三件套）的字段级展开，供 `tasks.md` 第 4 组实现时逐格对照。信息源：OpenAI Responses 官方 API 参考、Anthropic Messages 官方规范、仓库既有 OC↔AM 转换代码（`provider_anthropic.go`/`provider_openai.go`，下称"既有转换"）。
>
> **协议记号**：**OC** = OpenAI Chat Completions，**AM** = Anthropic Messages，**OR** = OpenAI Responses；`X2Y` 表示客户端协议 X → 上游方言 Y（如 `OR2OC` = Responses 客户端请求 Chat 上游）。

## 一、三协议定位速览

| | Chat Completions (OC) | Anthropic Messages (AM) | OpenAI Responses (OR) |
|---|---|---|---|
| 端点 | `POST /chat/completions` | `POST /messages` | `POST /responses` |
| 消息载体 | `messages[]` 角色消息 | `messages[]` 角色消息 | **`input`：string 或扁平 item 数组** |
| 系统提示 | `role:"system"` 消息 | 顶层 `system` 字段 | 顶层 `instructions` 字段 |
| 输出结构 | `choices[].message` | `content[]` 内容块 | **`output[]` item 数组** |
| 多轮上下文 | 客户端全量回放 | 客户端全量回放 | 全量回放（本网关强制 `store:false`） |
| 流式 | `data:{...}` chunk + `data:[DONE]`，**无事件名** | `event:` 命名事件 | `event:` 命名事件 + **每事件 `sequence_number`** |

## 二、顶层请求参数映射

| Responses (OR) | Chat (OC) | Messages (AM) | 转换说明 |
|---|---|---|---|
| `model` | `model` | `model` | 直传（网关改写为上游 ID） |
| `input` | `messages` | `messages` | 结构映射，见 §三 |
| `instructions` | 首条 `role:"system"` 消息 | 顶层 `system` | OR→OC 前置 system 消息；OR→AM 写 `system` |
| `max_output_tokens` | `max_tokens` | `max_tokens`（**必填**） | 三方改名；OC 侧注意 o 系列要 `max_completion_tokens`，既有 OC2AM 用 `max_tokens`，沿用 |
| `tools[]` 平铺 | `tools[]` **嵌套** `function:{}` | `tools[]` 平铺 | 见 §四 |
| `tool_choice` | `tool_choice` | `tool_choice` | `auto/none/required` ↔ `auto/none/any`；`{type:function}` ↔ `{type:function,name}` |
| `stream` | `stream` | `stream` | 直传 |
| `temperature`/`top_p` | 同名 | 同名 | **OR→AM 不传**（沿用既有 OC2AM 最小字段集先例；且 Anthropic 新模型已弃用采样参数，传了反而 400）；OR→OC 直传 |
| `reasoning{effort}` | （无对应/`reasoning_effort`） | `thinking{budget_tokens}` | **不映射**：effort→budget 无换算表；沿用既有转换不传推理配置的先例，由上游默认决定 |
| `metadata` | `metadata` | `metadata` | 直传（如存在） |
| `store`/`previous_response_id`/`conversation`/`background` | — | — | **网关拦截层已处理**（400 或改写 `store:false`），转换层永不见到 |
| （OC 侧无） | `stream_options{include_usage}` | — | **新格中上游为 OC 的（OR→OC）必须注入** `stream_options:{include_usage:true}` 才能从 chat 上游拿到 usage；既有 AM→OC 维持现状 |

**既有转换的字段集基准**（OC2AM 请求 struct 实测）：只取 `model/max_tokens/messages/tools/stream`——`temperature`、`top_p`、`stop`、`seed`、`response_format` **全部丢弃**。四个新转换格统一沿用此最小集（OR→OC 可多传 `temperature/top_p`，因为 chat 上游普遍支持且 Codex 会带）。

## 三、消息与内容结构映射

### 3.1 文本与图片

| OR `input` item | OC `messages` | AM `messages` |
|---|---|---|
| `input:string` | `{role:"user", content:[{type:"text",text}]}` | `{role:"user", content:[{type:"text",text}]}` |
| `{type:"message", content:[{type:"input_text",text}]}` | `{type:"text",text}` | `{type:"text",text}` |
| `{type:"input_image", image_url}` | `{type:"image_url",{url}}` | `{type:"image", source:{type:base64/url}}`（data URL 拆解，既有 `parseDataURL` 复用） |
| 输出侧 `{type:"output_text",text}` | `message.content` 字符串（多段 `\n` 拼接） | `{type:"text",text}` 块 |

### 3.2 工具调用（三方对照，关键表）

| 语义 | OR | OC | AM |
|---|---|---|---|
| 模型发起调用 | item `{type:"function_call", **call_id**, name, arguments(JSON串)}` | `message.tool_calls[] {id, type:"function", function:{name, arguments}}` | 块 `{type:"tool_use", **id**, name, input(对象)}` |
| 客户端回传结果 | item `{type:"function_call_output", **call_id**, output}` | `{role:"tool", **tool_call_id**, content}` | `{role:"user", content:[{type:"tool_result", **tool_use_id**, content}]}` |
| 参数形态 | JSON **字符串** | JSON **字符串** | **对象**（流式时为 `partial_json` 增量） |
| 关联键 | `call_id` | `tool_calls[].id` / `tool_call_id` | `tool_use.id` / `tool_use_id` |

- **`call_id` 是唯一贯穿键**：OR→X 转换必须原样保留 `call_id` 作为目标协议的 id（X→OR 反向同理），否则下一轮回传结果对不上。
- **`arguments` 字符串 ↔ `input` 对象**互转：既有 `json.Unmarshal`/`json.Marshal` 模式直接复用（OC2AM/AM2OC 双向均已实现）。
- **多 tool 结果聚合**：AM 协议要求 `tool_result` 块必须在 `role:"user"` 消息里、紧跟对应 `tool_use` 之后——既有 `pendingToolResults` + `flushPendingToolResults` 模式是唯一正确解，OR→AM/OC→AM 必须复用同构逻辑（OR/OC 侧 tool 消息是独立消息，AM 侧要合并成一条 user 消息）。
- **流式索引体系**：OC 用 `tool_calls[].index`（delta 内），AM 用内容块 `index`（事件级），OR 用 `output_index` + `content_index`。三套索引互转需维护映射表——既有 `toolCallState`（id/name/blockIdx/startSent）就是 OC↔AM 的现成状态机，OR 侧转换器需同构实现（加 `output_index` 维度）。

### 3.3 reasoning / thinking（按格拆解与信封机制）

> **B 方案（信封保真）已确认**：Anthropic 的 `thinking` 块不能凭空合成（无合法签名回传必 400），但可以**把原样签名随思考全文一起带走**——OR 的 `encrypted_content` 字段语义即"不透明、客户端原样回传"，正好当信封用。

| 格（客户端→上游） | 流向 | 映射 | 说明 |
|---|---|---|---|
| **OR→AM**（上游 Anthropic） | 响应（AM→OR） | `thinking{t,s}` → `{type:"reasoning", summary:[节选], encrypted_content:"gw1:"+base64({p:AnthropicBaseURL, s:signature, t:thinking})}` | **创建信封**：summary 供客户端展示，全文+签名进信封（`gw1:` 前缀与 OpenAI 真密文区分）；客户端（Codex 等）按协议原样回传 `encrypted_content` |
| **OR→AM** | 请求（OR→AM） | 信封解码成功**且 `p` == 当前 Anthropic 上游 BaseURL** → 按原序重建 `{type:"thinking", thinking:t, signature:s}`；否则 **丢弃该 item** | **消费信封**：原样回传 ≠ modified，验签通过、跨轮零丢失；`p` 不匹配（failover 换上游）→ 丢弃，绝不 400 |
| **OR→OC** 请求 / **OC→OR** 响应 | OR→OC 转换 | `reasoning` 文本 → assistant `reasoning_content`（扩展字段）：**格 OR→OC 请求**中信封存在则解出全文 `t`（比 summary 节选全），否则用 summary 文本；外来密文（非 `gw1:`，如 OC→OR 响应里上游 OR 的真密文）恒丢弃 | chat 不验签；与既有 OC2AM 的 `thinking→reasoning_content`（`provider_anthropic.go:503`）方向一致 |
| **OR→OC** 响应 / **OC→OR** 请求 | OC→OR 转换 | `reasoning_content`/`reasoning` → `{type:"reasoning", summary:[...]}` item | **不建信封**（chat 无签名可装）；既有 AM2OC 流式已从 OC delta 识别这两字段（`provider_openai.go:790-794`），读取逻辑可复用 |
| **AM→OR**（休眠，上游 OR） | 请求（AM→OR） | `thinking{t,s}` → `{type:"reasoning", summary:[t]}`，**不建信封** | 信封规则③：永不发往 OR 上游；OR 不验签，signature 丢弃无害；客户端原 thinking 块仍在自己消息里，下轮回 Anthropic 原生透传时签名完好 |
| **AM→OR**（休眠） | 响应（OR→AM） | 上游 OR 的 reasoning → **丢弃**（v1） | OpenAI 密文变不出 Anthropic 合法签名；文本透传（`thinking_delta`）留作后续优化 |

**信封规则（四条收敛）**：① 创建**仅在格 OR→AM 的响应方向**；② 消费在格 OR→AM 的请求方向（全量重建）与格 OR→OC 的请求方向（仅取文本）；③ 发往 OR 上游的请求**永不携带**信封；④ 客户端剥离 `encrypted_content`、信封损坏、`p` 不匹配 → 一律降级为丢弃（= 旧方案 A），**B 恒不劣于 A，永不因信封 400**。

**round-trip 自洽性（格 OR→AM）**：Anthropic 思考 →（格 OR→AM 响应）信封装入 `encrypted_content` → 客户端全量回放 →（格 OR→AM 请求）解信封、按原序重建 thinking 块 → Anthropic 验签通过。思考上下文**跨轮零丢失、无重思考**。

**关于既有 OC↔AM 的更正（原"同 precedent"说法作废）**：既有转换请求侧两侧都丢思考（OC2AM 丢 `reasoning_content`，`convertOpenAIRequestToAnthropic` 无此字段处理；AM2OC 丢 thinking 块，`convertAnthropicMessageToOpenAI` 无 `thinking` 分支，均已核代码）——但这是**协议约束所致**：OC 协议没有承载 Anthropic 签名的字段，丢弃是唯一不 400 的选择；**B 方案恰恰用 `encrypted_content` 把这个约束绕开了**，所以"既有也丢"不构成 OR→AM 丢弃的先例。既有真正的缺陷是 **AM2OC 非流式响应丢 `reasoning_content` 而流式保留**（`convertOpenAIResponseToAnthropic` 无此字段处理，不对称、疑似 bug），受"既有 OC↔AM 零改动"约束本 change 不动，已列入 design Open Questions。

## 四、tools 定义映射

```jsonc
// OR（平铺）                          // OC（嵌套）                          // AM（平铺）
{"type":"function",                  {"type":"function",                   {"name":"get_weather",
 "name":"get_weather",                "function":{"name":"get_weather",     "description":"...",
 "description":"...",                  "description":"...",                  "input_schema":{...}}
 "parameters":{...}}                  "parameters":{...}}}
```

- **OR→AM**：只差 `parameters` → `input_schema` 改名 + 删 `type`——**最浅的一格**；既有 `convertOpenAIToolToAnthropic` 已实现 OC 嵌套→AM，OR 平铺→AM 是它的简化版。
- **OR→OC**：平铺 → 嵌套（包一层 `function`）。
- **OC→OR / AM→OR**：反向拆解（既有 `convertAnthropicToolToOpenAI` 的镜像 + 去嵌套）。
- `tool_choice`：`required`（OR/OC）→ `{type:"any"}`（AM）；`none` → `{type:"none"}`；`auto` → `{type:"auto"}`。

## 五、响应体映射

### 5.1 结构

| OR 响应 | OC 响应 | AM 响应 |
|---|---|---|
| `id`（`resp_` 前缀） | `id`（`chatcmpl-` 前缀） | `id`（`msg_` 前缀） |
| `object:"response"` | `object:"chat.completion"` | `type:"message"` |
| `status:"completed"` | （隐含） | （隐含） |
| `output[]`：`{type:"message",content:[output_text]}` | `choices[0].message{role:"assistant",content}` | `content[]:{type:"text"}` |
| `output[]`：`{type:"function_call"}` | `choices[0].message.tool_calls[]` | `content[]:{type:"tool_use"}` |
| `status:"incomplete"` + `incomplete_details.reason:"max_output_tokens"` | `finish_reason:"length"` | `stop_reason:"max_tokens"` |
| 有 function_call 时 | `finish_reason:"tool_calls"` | `stop_reason:"tool_use"` |
| 正常结束 | `finish_reason:"stop"` | `stop_reason:"end_turn"` |
| `usage{input_tokens,output_tokens,total_tokens}` | `usage{prompt_tokens,completion_tokens,total_tokens}` | `usage{input_tokens,output_tokens}` |
| `error` / `incomplete_details.reason:"content_filter"` | `finish_reason:"content_filter"` | `stop_reason:"refusal"`（近似） |

**id 前缀合成**：跨协议转换时 id 必须重写为目标协议格式（既有 `generateID()` 两处已有，OR 侧合成 `resp_`/`rs_` 前缀）。

### 5.2 usage 映射（结论：只有 chat 需要改名）

```
OR.input_tokens    ⇄ AM.input_tokens     （同名，直拷）
OR.output_tokens   ⇄ AM.output_tokens    （同名，直拷）
OR.input_tokens    → OC.prompt_tokens    （改名）
OR.output_tokens   → OC.completion_tokens（改名）
OR.total_tokens    → OC.total_tokens     （同名；AM 无 total，既有转换也不造）
OR.input_tokens_details.cached_tokens / output_tokens_details.reasoning_tokens → OC 无对应，丢弃
```

## 六、SSE 事件流映射（四格的核心工程量）

### 6.1 三协议事件形状

```
OC:  data:{choices:[{delta:{role/content/tool_calls}}]}  ×N → data:{choices:[{finish_reason}]} → data:[DONE]
    （无事件名、无序号；usage 需 stream_options.include_usage 才在末尾 chunk 出现）

AM:  event:message_start → [content_block_start → content_block_delta×N → content_block_stop]×块
    → message_delta{stop_reason,usage} → message_stop     （另有 ping、error）

OR:  event:response.created → response.in_progress
    → [output_item.added → content_part.added → output_text.delta×N → content_part.done → output_item.done]
    → response.completed{response:{...全量output,usage}}   （每事件带 sequence_number；另有 incomplete/failed/error）
```

> **方向约定**：小节标题按**格**（客户端→上游，与 tasks/9 格矩阵一致）命名；表格列头显式标注上游/客户端方向。**请求转换方向 == 格名；响应转换方向 == 格名反向**——这是本文档此前表格装反的根源。

### 6.2 格 OR→OC（Codex → DeepSeek/Kimi 类，**本期主流量**）

> 响应流：上游 OC chunk → 客户端 OR 事件

| 上游 OC chunk | 客户端 OR 事件 |
|---|---|
| 首 chunk `role` | `response.created` + `in_progress` + `output_item.added`(message) + `content_part.added` |
| `delta.content` | `output_text.delta`（带自增 `sequence_number`） |
| `delta.tool_calls[].index` | 映射为 `output_index`：`output_item.added`(function_call) + `function_call_arguments.delta` |
| `finish_reason` chunk | `content_part.done` + `output_item.done` |
| usage chunk（请求注入 `stream_options.include_usage` 捕获） | 暂存，归入 `response.completed`（`prompt/completion` → `input/output` 改名） |
| **`[DONE]`** | **`response.completed`**（拼全量 `output` 数组 + 暂存 usage） |
| 上游中断/异常（无 `[DONE]`） | 合成 `response.failed` 或 `error` 事件——OR 客户端依赖终态事件，不能静默断流 |

### 6.3 格 OC→OR（休眠格）

> 响应流：上游 OR 事件 → 客户端 OC chunk

| 上游 OR 事件 | 客户端 OC chunk |
|---|---|
| `response.created` / `in_progress` | 首个 `delta:{role:"assistant"}` chunk（可选但推荐，SDK 依赖） |
| `output_item.added`(message) | （并入首 chunk） |
| `output_text.delta{delta}` | `delta:{content: delta}` |
| `output_item.added`(function_call) | `delta:{tool_calls:[{index, id:call_id, type:"function", function:{name, arguments:""}}]}` |
| `function_call_arguments.delta{delta}` | `delta:{tool_calls:[{index, function:{arguments: delta}}]}` |
| `output_item.done`(function_call) | （无，或补 `function:{arguments}` 定稿） |
| `response.completed` | ① `finish_reason`（无工具→`stop`；有工具→`tool_calls`；incomplete+max_output→`length`）② 若客户端请求带 `stream_options.include_usage` → 补 usage chunk（`input/output`→`prompt/completion` 改名；OR 的 completed 自带 usage，**无需注入**）③ **`data:[DONE]`** |
| `response.failed` / `error` | 终止 chunk（`finish_reason:"content_filter"` 或直接断流）——**OC 客户端在等 [DONE]，必须给终止信号** |

### 6.4 格 OR→AM（Codex → Anthropic-only 上游）

> 响应流：上游 AM 事件 → 客户端 OR 事件

| 上游 AM 事件 | 客户端 OR 事件 |
|---|---|
| `message_start`（`usage.input_tokens/output_tokens` 暂存） | `response.created` + `in_progress` + `output_item.added`(message) + `content_part.added` |
| `content_block_start{thinking}` | `output_item.added`(reasoning) + `reasoning_summary_part.added` |
| `content_block_delta{thinking_delta}` | `reasoning_summary_text.delta`（思考全文入缓冲） |
| `content_block_delta{signature_delta}` | 签名入缓冲（不单独发事件） |
| `content_block_stop`(thinking) | `reasoning_summary_part.done` + `output_item.done`（**`encrypted_content` = gw1 信封**，§3.3） |
| `content_block_delta{text_delta}` | `output_text.delta` |
| `content_block_stop`(text) | `content_part.done` + `output_item.done` |
| `content_block_start{tool_use}` | `output_item.added`(function_call，`call_id`=tool_use.id) |
| `content_block_delta{input_json_delta}` | `function_call_arguments.delta` |
| `content_block_stop`(tool_use) | `output_item.done` |
| `message_delta{stop_reason,usage}` | `response.completed`（拼全量 `output`：`input_tokens` 来自 message_start 暂存、`output_tokens` 来自本事件，**同名直拷**） |
| `message_stop` | （并入 completed） |
| `ping` | 忽略不转发 |
| `error` | `event:error` 或 `response.failed` |

### 6.5 格 AM→OR（休眠格）

> 响应流：上游 OR 事件 → 客户端 AM 事件

| 上游 OR 事件 | 客户端 AM 事件 |
|---|---|
| `response.created` | `message_start`（`content:[]`；**`usage.input_tokens` 只能置 0**——OR 的 usage 到 `completed` 才有，见 §七-4） |
| `output_item.added`(message) + `content_part.added` | `content_block_start{index, content_block:{type:"text",text:""}}` |
| `output_text.delta` | `content_block_delta{index, delta:{type:"text_delta",text}}` |
| `content_part.done` | `content_block_stop{index}` |
| `output_item.added`(function_call) | `content_block_start{index, content_block:{type:"tool_use",id:call_id,name,input:{}}}` |
| `function_call_arguments.delta` | `content_block_delta{delta:{type:"input_json_delta",partial_json: delta}}` |
| `output_item.done`(function_call) | `content_block_stop{index}` |
| `reasoning` 相关事件 | **丢弃**（§3.3 格 AM→OR 响应行，v1）——不产生任何 AM 事件 |
| `response.completed` | `message_delta{delta:{stop_reason 映射}, usage:{output_tokens}}` + **`message_stop`** |
| `response.failed` / `error` | `event:error` + `data:{type:"error",error:{...}}` |
| `sequence_number` | 消费后不转发（AM 事件无此字段） |

## 七、陷阱清单（实现时逐条核对）

> 方向记号统一为**格**（客户端→上游）；"响应/请求"指该格内的流向。

1. **`sequence_number` 必须单调**：**合成 OR 事件的格（OR→OC、OR→AM 的响应方向）**每个事件带自增序号；**消费 OR 事件的格（OC→OR、AM→OR 的响应方向）**只读不转发，无需维护。
2. **`call_id` 全程原样**：它是下一轮 `function_call_output`/`tool_call_id`/`tool_use_id` 的关联键，任何环节改写即断链。
3. **OC 流没有终止事件名**：**格 OC→OR**（响应）在 `response.completed` 后必须补 `data:[DONE]`；**格 OR→OC**（响应）必须把上游 `[DONE]` 消费掉、换成 `response.completed`（上游 OC 断流无 `[DONE]` 时合成 `response.failed`）。
4. **格 AM→OR 流式 `message_start.input_tokens` 恒为 0**：OR 的 usage 只在 `completed` 事件给出，时序上晚于 `message_start`——不可解，接受（客户端显示 0，`ModelLog` 仍从 completed 拿准确值）。**格 OR→AM 无此问题**：AM 的 `message_start` 自带 `input_tokens`，暂存后归入 completed。
5. **OC 流 usage 要主动要**：格 OR→OC 向 chat 上游请求注入 `stream_options:{include_usage:true}`；上游不支持则 usage 为 0，降级接受（既有 AM→OC 行为维持现状，不在本 change 范围）。格 OC→OR **无需注入**（OR 的 completed 自带 usage），仅在客户端请求带 `include_usage` 时补发 chunk。
6. **AM 要求 `max_tokens` 必填**：格 OR→AM 若客户端没给 `max_output_tokens`，需给默认值（OR 客户端如 Codex 通常必给）。
7. **AM 的 tool_result 聚合**：格 OR→AM（及既有 OC→AM）多个连续 tool 结果必须合并进**一条** `role:"user"` 消息（Anthropic 硬约束），复用 `flushPendingToolResults` 同构逻辑。
8. **OR 的 item 扁平 vs 消息嵌套**：OR 一轮的 `input` 可能是 `[message, function_call, function_call_output, reasoning, message...]` 混排——X→OR 转换（请求方向）要按语义切分成对应 item 序列，不能按"一条消息一个 item"想当然。
9. **未知 item/event 优雅跳过**：OR 有 53 种事件、item 还有 `image_generation_call`/`code_interpreter_call` 等托管工具形态——转换器对不认识的类型**跳过不崩**（既有转换对未知 content 块的处理同此原则）。
10. **`stop_sequences`/`stop`**：OR `stop` ↔ AM `stop_sequences` 改名（若传）；既有最小字段集不传，维持。
11. **gw1 信封跨原生透传的已知边缘**：格 OR→AM 创建的信封若下一轮经**原生透传**到达 OR 上游（同一虚拟模型同时映射 Anthropic 与 OR 原生两类上游、权重中途切换），上游可能视 `encrypted_content` 为非法密文而 400——发生条件苛刻，接受为已知边缘（错误原样透传）；若实测常见，在原生透传分支剥离 `gw1:` 前缀的 encrypted_content 即可（前缀可识别，剥离成本 O(input)）。

## 八、与既有转换的复用矩阵

| 既有资产 | 位置 | 新格复用方式 |
|---|---|---|
| `pendingToolResults` 聚合模式 | `provider_anthropic.go:217` | OR→AM、OC→AM 同构重写（AM 侧约束相同） |
| `arguments` 字符串 ↔ `input` 对象 | OC2AM/AM2OC 双向 | 直接抽 helper 复用 |
| `parseDataURL` | `provider_anthropic.go:423` | OR→AM 图片块复用 |
| `toolCallState` 状态机 | `provider_openai.go:676` | OC→OR 流式加 `output_index` 维度同构实现 |
| `reasoning_content` 识别 | `provider_openai.go:790` | OC→OR 镜像 |
| `generateID` / SSE writer | 两文件各一份 | OR 侧合成 `resp_` 前缀 id + `sequence_number` 维度 |
| `copyOpenAIStreaming` 管道骨架 | `provider_openai.go` | 四格流式转换统一走"读事件→翻译→写回"三段式 |
