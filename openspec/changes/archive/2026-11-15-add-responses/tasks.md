# Tasks

## 1. 阶段一：接口与文件改名（零行为变化）

- [x] 1.1 `git mv` 两文件：`provider_openai.go` → `provider_openai_chat_completions.go`、`provider_anthropic.go` → `provider_anthropic_messages.go`，验证 `go build ./...` 通过且 `git diff --stat` 仅显示改名（无内容改动）
- [x] 1.2 接口方法改名：`provider.go` 中 `ExecuteOpenAIRequest` → `ExecuteOpenAIChatCompletionsRequest`、`ExecuteAnthropicRequest` → `ExecuteAnthropicMessagesRequest`，同步 `proxy_openai_chat_completions.go`、`proxy_anthropic_messages.go`、`model_testing.go`、`provider_automated.go` 全部调用点与实现，验证 `go build ./...` 通过（编译器兜底找全三实现三调用点）
- [x] 1.3 搬移既有转换：OC2AM 转换块从 `provider_anthropic_messages.go` 剪切到 `provider_openai_chat_completions.go`、AM2OC 从 `provider_openai_chat_completions.go` 剪切到 `provider_anthropic_messages.go`（纯移动），验证 `git diff` 中搬移块内容逐行一致（零逻辑改动）且 `go build ./...` 通过
- [x] 1.4 手工回归既有链路：chat 流式/非流式、messages 流式/非流式、model 测试功能各跑通一次，验证行为与改名前一致

## 2. 阶段二：字段词表改名 + ResponsesBaseURL 贯通

- [x] 2.1 `model/db.go` PO 字段与 gorm tag 改名（`OpenAIChatCompletionsBaseURL`/`AnthropicMessagesBaseURL`）+ 新增 `OpenAIResponsesBaseURL`（`openai_responses_base_url`），验证 `go build ./...` 通过且 `AutoMigrate` 启动时新增空列、不改名既有列
- [x] 2.2 同步改名波及点：`handler/provider.go`（DTO 三件套、updates map、"至少一个 BaseURL"校验扩到三字段、URL TrimSuffix）、`handler/model.go`、`router/route_result.go`、`router/router.go`、`handler/provider_model.go`、`model_testing.go`，验证 `go build ./...` 通过且 grep 仓库无残留 `openai_base_url`/`anthropic_base_url`（除 openspec 归档）
- [x] 2.3 `provider.Config` 扩展为三 BaseURL 字段，Router 创建实例时携带全部已配置值，验证 `go build ./...` 通过
- [x] 2.4 前端 `web/src/views/Providers/index.vue`：字段改名（表单、标签判定、回显）+ 新增 Responses BaseURL 输入框 + "至少一个"校验同步，验证 `npm run build` 通过且页面上三输入框可保存回显
- [x] 2.5 编写并执行上线迁移手册（存 `openspec/changes/add-responses/` 下）：两列 `ALTER TABLE ... RENAME COLUMN` 语句 + "先改列、后部署"顺序说明，验证在本地库副本上执行后新代码启动读取数据正常

## 3. 阶段三：Responses 主端点 + 拦截层 + 管理端点

- [x] 3.1 `gatewayError(c, status, code, message)` 错误体构造器（OpenAI 风格 `{"error":{...}}`），验证单元 curl 可打出 400 `stateless_gateway` 标准体
- [x] 3.2 `Responses` handler：读 body → 拦截 `previous_response_id`/`conversation`/`background`（400）→ 强制 `store=false` → 解析 model → `VerifyKeyID` → `Route` → Execute → `ModelLog`，验证 curl：三状态字段各得 400 且上游未收到请求、无状态字段请求正常放行、`store:true` 请求上游收到 `false`
- [x] 3.3 管理端点 handler：`GET/DELETE /responses/{id}`、`POST /responses/{id}/cancel`、`GET /responses/{id}/input_items`、`/conversations` 族注册路由并只返回 400 `stateless_gateway`，验证 curl：带 key 得 400 标准体、不带 key 得 401、上游无请求
- [x] 3.4 `ExecuteOpenAIResponsesRequest` 原生透传实现（`provider_openai_responses.go`：OpenAIProvider 分发 + 三分发规则），验证 curl 指向千问类上游：流式/非流式 200 透传、上游 404 原样透传
- [x] 3.5 usage 抓取双路径（非流式读响应体、流式 tee 扫 `response.completed`），验证 `model_logs` 记录 `input_tokens`/`output_tokens` 映射值，且流式事件流本体逐字节不受解析支路影响
- [x] 3.6 `main.go` 注册全部新路由（1 主端点 + 5 类管理端点），验证启动无路由冲突、`/openai/v1` 组中间件（RequireAPIKey）对新端点生效

## 4. 阶段四：跨协议转换（OR↔OC、OR→AM、休眠格）

> 全组实现前先读 `protocol-fields.md`：字段映射表（§二~§五）、SSE 事件映射（§六）、陷阱清单（§七，10 条逐条核对）、既有资产复用（§八）。

- [x] 4.1 Responses SSE 事件类型定义与解析器（`provider_openai_responses.go` 内，`response.created`/`output_text.delta`/`response.completed` 等），验证可解析千问类上游真实事件流并提取 usage
- [x] 4.2 OR→OC 转换（请求/响应/SSE 三件套），验证 curl：Codex 形态请求打 DeepSeek 类 chat-only 上游流式/非流式跑通、多轮 `function_call` ↔ `tool_calls` 往返正确、usage 落库
- [x] 4.3 OC→OR 转换（休眠格），验证构造仅配 `openai_responses_base_url` 的 Provider，chat 客户端请求可转换跑通
- [x] 4.4 OR→AM 转换（请求/响应/SSE 三件套 + gw1 信封，见 protocol-fields §3.3），验证 Responses 客户端打 Anthropic-only 上游流式/非流式跑通、thinking 信封往返（响应侧装入 `encrypted_content`、请求侧按原序重建带签名 thinking 块、上游验签通过不 400）、信封被剥离/损坏时降级丢弃且请求不失败
- [x] 4.5 AM→OR 转换（休眠格），验证 Anthropic 客户端请求仅配 responses 端点的 Provider 可转换跑通

## 5. 集成回归

- [x] 5.1 全量手工回归：chat 流式/非流式、messages 流式/非流式、model 测试、管理后台 Provider 增改查，验证全部与改动前行为一致（仓库无测试，此为上线门槛）
- [x] 5.2 9 格矩阵逐格冒烟：3 透传格 + 4 新转换格 + 2 既有转换格各一条请求，验证每格请求成功且 `model_logs` 有正确记录
- [x] 5.3 错误契约冒烟：拦截层三字段 400、管理端点 400、上游错误原样透传三类场景，验证错误体格式与 spec 场景逐条一致
- [x] 5.4 运行 `openspec validate add-responses --strict` 并确认通过，spec 场景与实现逐条对照无偏差

## Workflow follow-up

- 实现完成后运行 `openspec verify add-responses` 核对实现与工件一致。
- 评审通过后归档 change（`openspec archive add-responses`），确认 delta 合入主 specs。
