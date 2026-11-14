# Design

## Context

现有链路：`handler/proxy_openai_chat_completions.go` 读 body → 解析 model → `VerifyKeyID` 白名单 → `router.Route(model)` → `Provider.Execute*` 转发 → `ModelLog` 落库。`Provider` 接口有两个按方言命名的方法（`ExecuteOpenAIRequest`/`ExecuteAnthropicRequest`），OC↔AM 转换约 1000 行集中在 `provider_anthropic.go`，**仓库无任何测试**——所有改动必须以"既有转换代码零逻辑改动"为前提。动机与范围见 proposal.md（Why）；行为契约见 `specs/openai-responses-api/spec.md` 与 `specs/provider-multi-style/spec.md`。

## Goals / Non-Goals

**Goals:**

- 9 格矩阵每格可分发、可验证；新代码集中在 Responses 相关文件，既有 chat/messages 链路仅改名
- 拦截层与管理端点的错误应答共用一个错误体构造器，保证 `stateless_gateway` 契约处处一致
- 转换器的 SSE 事件类型定义只写一份，供四个转换格复用

**Non-Goals:**

- 不做状态存储、ID 映射、404 运行时回退（见 proposal 范围排除）
- 不做 Responses 端点的模型测试功能扩展（model_testing 维持 chat 现状）
- 不动 codec 注册表、不动 `Execute(spec)` 抽象

## Decisions

### D1：接口与文件按外部产品名定名，文件 = 接口方法

接口三方法：`ExecuteOpenAIChatCompletionsRequest` / `ExecuteAnthropicMessagesRequest` / `ExecuteOpenAIResponsesRequest`。文件组织采用唯一规则——**一个文件 = 一个接口方法 = 一个客户端协议入口的全部行为，转换与触发它的入口同文件**：

| 文件 | 内容 | 性质 |
|---|---|---|
| `provider_openai_chat_completions.go` | `ExecuteOpenAIChatCompletionsRequest` 全部实现：OpenAI 透传、既有 OC2AM（自 `provider_anthropic.go` 搬入）、新增 OC→OR | 改名 + 搬入 |
| `provider_anthropic_messages.go` | `ExecuteAnthropicMessagesRequest` 全部实现：Anthropic 透传、既有 AM2OC（自 `provider_openai.go` 搬入）、新增 AM→OR | 改名 + 搬入 |
| `provider_openai_responses.go` | `ExecuteOpenAIResponsesRequest` 全部实现（原生透传、OR→OC、OR→AM 的分发）+ Responses SSE 事件类型与解析器 + usage 抓取 | 新增 |
| `provider_automated.go` | 三方法的委托分发（分发器不属于三协议之一，不改名），增加 responses 分支 | 修改 |

- **不设独立 conversion 文件**：OR→OC/OR→AM 与分发它们的 `ExecuteOpenAIResponsesRequest` 同居 responses 文件；OC→OR、AM→OR 分别随 chat/messages 入口就位；Responses SSE 事件类型定义放 responses 文件（Responses 协议的"家"），其余文件同包直接引用
- **既有代码的搬运**：OC2AM、AM2OC 各一次纯剪切粘贴（同包、编译器兜底、零逻辑改动），diff 变大但语义零风险；共享 helper（`copyOpenAIResponse`/`copyOpenAIStreaming`）留在 chat 文件供其余文件跨文件调用
- 方法可分布在不同文件（Go 同包合法），一个 struct 的方法按"入口协议"归文件，不再按 struct 归文件

**替代方案（曾考虑，被否）**：按"转换目标方言"摆放（OC2AM 留在 anthropic 文件、OR→OC 放 chat 文件……），可零搬运、与既有代码物理位置一致——但会出现"Responses 入口不在 responses 文件"的错位；另一替代是独立 `conversion_*.go` 按转换对分文件——同包内文件无模块边界，只增寻址成本。二选一后确定"文件 = 方法"，与接口命名法同构。

### D2：Config 携带全部三个 BaseURL，分发在 Execute 内部

`provider.Config` 增加 `OpenAIResponsesBaseURL`，Router 为三种 requestStyle 创建的实例都携带该 Provider **全部已配置**的 BaseURL（而非只带本风格的那一个）——因为 OR→OC 回退需要同时看到 `OpenAIResponsesBaseURL`（空则不原生）与 `OpenAIChatCompletionsBaseURL`（转换目标）。分发规则统一实现为：

```
原生方言存在 → 透传
否则按 Chat > Anthropic > Responses 顺位找第一个存在的 BaseURL → 转换
否则 → 明确错误（"provider 无可用端点"，理论不可达，因创建校验要求至少一个）
```

**替代方案**：Router 在创建时就选好"透传 or 转换"并构造不同实例——被否：分发规则会散落在 Router 与两个 Provider 里三处，收敛在 Execute 内只写一次，且与 spec 的"Router 创建实例 + 实例按权重分发"一致。

### D3：拦截层放在 handler，body 只解析一次

`Responses` handler 流程：

```
读 body → JSON 解析为 map
  ├─ 含 previous_response_id / conversation / background → 400 stateless_gateway（不进路由）
  ├─ 强制 store = false（缺省也补）
  ├─ 解析 model → VerifyKeyID → Route（复用现有链路，零改动）
  └─ 改写后的 body 字节下传 Execute（不二次读取）
```

- **错误体构造器**：新 `gatewayError(c, status, code, message)` helper，输出 OpenAI 风格 `{"error":{message,type,code}}`；拦截层 400、管理端点 400 共用，保证契约一致
- **改写方式**：unmarshal → 改 map → re-marshal。代价是字段顺序与空白变化（JSON 语义无损）；被否的替代是"字符串级正则改写"（对嵌套/缺省 `store` 不可靠）
- 解析为 map 后透传，天然保留 `input`、`tools`、`reasoning` 等全部未知字段

### D4：管理端点 = 注册但只出错误

`GET/DELETE /responses/{id}`、`POST /responses/{id}/cancel`、`GET /responses/{id}/input_items`、`/conversations` 族：挂进现有 `RequireAPIKey` 组（未鉴权先 401，与全站一致），handler 体只有一行 `gatewayError(400, "stateless_gateway", ...)`，**不解析、不路由、不触达 Provider**。

### D5：转换器三件套结构 + SSE 类型一份

> **字段级映射基准见 `protocol-fields.md`**（三协议参数/消息/工具/响应/SSE 全对照 + 10 条陷阱 + 既有资产复用矩阵），本节只定结构。

每个转换格 = 请求翻译 + 响应翻译 + SSE 事件流翻译：

- **请求翻译**：unmarshal 双方结构体 → 字段映射 → marshal。`input`（string/Item 数组）↔ `messages`；`function_call`↔`tool_calls`↔`tool_use` 按既有 OC↔AM 的成熟映射对齐
- **reasoning 信封（B 方案，字段级细节见 `protocol-fields.md` §3.3）**：格 OR→AM 响应把 Anthropic `{thinking, signature}` 装入 reasoning item 的 `encrypted_content`（`gw1:`+base64，含上游 BaseURL 标识）；请求侧解信封、按原序重建带原样签名的 thinking 块——**跨轮零丢失、无重思考**；信封缺失/损坏/标识不匹配一律降级为丢弃（= 不带信封的旧行为），**永不因信封 400，B 恒不劣于 A**；发往 OR 上游的请求永不携带信封
- **SSE 翻译**：逐事件转换。Responses 侧事件形状统一定义在 `provider_openai_responses.go`，chat/messages 文件中的 OC→OR、AM→OR 直接引用同一套类型——四格不各抄一遍事件结构
- **流式串联**：转换格内是"上游 SSE → 解析为事件 → 翻译 → 写回客户端"，复用现有 `copyOpenAIStreaming` 的管道骨架，仅替换事件翻译层
- 既有 OC↔AM 转换（约 1000 行）**零改动**，只是被既有方法名改名后继续调用

### D6：usage 抓取（双路径）

| 路径 | 抓取点 | 映射 |
|---|---|---|
| 非流式 | 响应体 `usage.input_tokens/output_tokens` | → `ModelLog` 的 prompt/completion token 字段 |
| 流式 | tee 一条解析支路扫 SSE，`response.completed` 事件里带 usage | 同上；事件流本体原样写客户端，解析失败只丢统计不丢流 |

透传与转换两条路径都走 D6——转换路径的 usage 由翻译层顺带归一。

### D7：三字段按产品名词表全链改名 + `OpenAIResponsesBaseURL` 贯通

**词表（DB 列 = Go 字段 = API JSON = 前端表单，一处词表四层同名）**：

| 旧 | 新 |
|---|---|
| `openai_base_url` / `OpenAIBaseURL` | `openai_chat_completions_base_url` / `OpenAIChatCompletionsBaseURL` |
| `anthropic_base_url` / `AnthropicBaseURL` | `anthropic_messages_base_url` / `AnthropicMessagesBaseURL` |
| —（新增） | `openai_responses_base_url` / `OpenAIResponsesBaseURL` |

- **零兼容层**：gorm 用新列名直读，不保留旧列名 tag、不做双名查找；`AutoMigrate` 只会新增 `openai_responses_base_url`（默认空），**不会**重命名既有两列——存量两列的 `ALTER TABLE ... RENAME COLUMN` 由运维**上线前手动执行**，顺序强制"先改列、后部署"
- **新列语义**：空字符串 = 无原生 Responses 能力 → 走协议转换（存量 Provider 行为不变，安全降级）
- **贯通面**：`model/db.go`（gorm tag）、`handler/provider.go`（DTO 三件套 + updates map）、`handler/model.go`、`router/route_result.go`、`model_testing.go`、前端 `Providers/index.vue`（字段 + 标签判定 + 新增输入框）
- URL 校验与"至少一个"校验沿用现有逻辑，扩展到第三字段；`SyncModels` 端点顺位 OpenAI → Anthropic → Responses

## Risks / Trade-offs

- **[仓库无测试，chat/messages 主路径被改名波及]** → 改名与逻辑改动分两个提交；回归清单全量手工跑（见 tasks）：chat 流式/非流式、messages 流式/非流式、model 测试
- **[字段改名无兼容层，部署顺序错误即启动读不到列]** → 迁移手册写死"先 RENAME DDL、后部署"；回滚对称（列名改回）；启动时 gorm 会因缺列直接报错——快速失败优于静默错数据
- **[SSE 转换是全新事件形状，四格 × 流式 = 最大出错面]** → 事件类型单份定义（`provider_openai_responses.go`）；先做 OR↔OC 流式（覆盖 DeepSeek/Kimi 主流量）再做 OR→AM；每格手工验证"终止事件送达 + 无半截 chunk"
- **[unmarshal/re-marshal 改变字段顺序]** → JSON 语义无损；上游与 SDK 均不依赖键序；风险接受
- **[OR→AM/OR↔OC 转换丢信息（web_search_call 等托管工具 item 无对应物）]** → 与既有 OC↔AM 的取舍一致：能映射的（reasoning/tool_calls/input）全保真，无对应物的丢弃并在 SSE 翻译里静默跳过；不为此发明私有字段
- **[`store` 强制改写与"纯透传"哲学的张力]** → 已在 spec 固化为强制无状态契约；改写面收敛为仅此一个字段

## Migration Plan

1. **阶段一（零行为变化）**：`git mv` 两文件 + 接口方法改名 + OC2AM/AM2OC 搬移 + 调用点同步 → 编译通过 + 回归 chat/messages
2. **阶段二（配置层）**：三字段全链改名 + 新增 `openai_responses_base_url` 列 + PO/DTO/表单 + Config 扩展 → **上线前运维手动执行两列 RENAME DDL（顺序：先改列、后部署）** → 验证存量数据读取正常、新列空字符串、关联不动
3. **阶段三（新端点）**：拦截层 + 管理端点 + 原生透传 + usage → curl 验证 400 契约与透传
4. **阶段四（转换）**：OR→OC → OR→AM → OC→OR/AM→OR 休眠格（SSE 类型定义在 `provider_openai_responses.go`）→ 逐格流式/非流式验证

**回滚**：阶段一为纯机械改动，直接回滚提交即可；阶段二回滚需把两列名 DDL 改回旧名（手动，与上线对称）；阶段三/四新端点无存量依赖，回滚不影响既有流量。`openai_responses_base_url` 列留空无副作用，无需回滚。

## Open Questions

- 硅基流动等"标准路径但响应字段疑似非标准"的上游是否需要兼容分支——待阶段三实测响应体后决定，不影响 spec（spec 约定"原样透传"，非标准即按非标准透传）
- **既有 AM2OC 非流式响应丢 `reasoning_content` 而流式保留**（`convertOpenAIResponseToAnthropic` 无此字段处理，已核代码，不对称、疑似既有 bug）——本 change 遵守"既有 OC↔AM 零改动"约束不修；是否后续用与 gw1 信封同源的机制补齐（含既有 OC2AM/AM2OC 请求侧思考传递），留待独立 change 评估
