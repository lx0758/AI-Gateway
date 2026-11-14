# Spec Delta

## Purpose

提供与 OpenAI Responses API 兼容的代理端点，让 Codex CLI 等仅说 Responses 协议的客户端接入本网关；涵盖强制无状态拦截、9 格协议分发与跨协议转换、管理端点错误契约、流式转发与 Token 统计。

## ADDED Requirements

### Requirement: Responses 端点鉴权与按模型路由

系统 SHALL 提供 `POST /openai/v1/responses` 端点，要求 Bearer API Key 鉴权，并按请求体中的 `model` 字段执行 Key 模型白名单校验与上游路由，与现有对话端点共用同一套 Key、白名单与路由体系。

#### Scenario: 鉴权通过并路由到上游
- **WHEN** 客户端携带有效 API Key 且 `model` 在该 Key 白名单内向 `POST /openai/v1/responses` 发起请求
- **THEN** 系统按 model 解析出目标 Provider 并将请求交由其执行

#### Scenario: 模型不在 Key 白名单
- **WHEN** 请求的 `model` 不在该 API Key 的模型白名单中
- **THEN** 系统拒绝请求并返回与现有对话端点一致的鉴权错误

### Requirement: 强制无状态请求拦截

系统 SHALL 在路由之前拦截携带 `previous_response_id`、`conversation` 或 `background` 字段的 Responses 请求，返回 HTTP 400 与结构化错误体（`error.code` 为 `stateless_gateway`），且 SHALL NOT 转发至任何上游。

#### Scenario: 携带 previous_response_id
- **WHEN** 请求体包含非空 `previous_response_id`
- **THEN** 系统返回 400，错误体 `error.code` 为 `stateless_gateway`，message 指示改用全量回放输入

#### Scenario: 携带 conversation 或 background
- **WHEN** 请求体包含 `conversation` 或 `background`
- **THEN** 系统返回 400 与相同的 `stateless_gateway` 错误契约

#### Scenario: 纯无状态请求放行
- **WHEN** 请求体不含任何上述状态字段
- **THEN** 请求正常进入白名单校验与路由流程

### Requirement: store 字段强制改写

系统 SHALL 在转发前将请求体中的 `store` 字段无条件改写为 `false`（含字段缺省的情况），且该改写 SHALL NOT 影响其余字段的原样透传。

#### Scenario: 客户端显式要求存储
- **WHEN** 请求体携带 `store: true`
- **THEN** 上游收到的请求体中 `store` 为 `false`
- **AND** 客户端收到的响应回显 `store` 为 `false`

#### Scenario: store 字段缺省
- **WHEN** 请求体未包含 `store` 字段
- **THEN** 上游收到的请求体中 `store` 为 `false`

### Requirement: 原生 Responses 透传

当目标 Provider 配置了 `OpenAIResponsesBaseURL` 时，系统 SHALL 将请求原样透传至 `{OpenAIResponsesBaseURL}/responses`，请求与响应全字段保留；上游非 2xx 响应 SHALL 原样透传给客户端。

#### Scenario: 上游支持原生 Responses
- **WHEN** Provider 配置了 OpenAIResponsesBaseURL 且上游正常处理请求
- **THEN** 客户端收到与直连上游一致的响应体（或一致的 SSE 事件流）

#### Scenario: 上游返回错误
- **WHEN** 上游对透传请求返回非 2xx 状态码
- **THEN** 状态码与错误体原样透传给客户端

#### Scenario: 上游未配置 ResponsesBaseURL
- **WHEN** Provider 的 OpenAIResponsesBaseURL 为空
- **THEN** 系统不发起原生透传，转而按方言分发权重执行跨协议转换

### Requirement: 方言分发权重

系统 SHALL 在 9 格矩阵（3 种客户端协议 × 3 种上游方言）内按"原生方言优先，必须转换时按 Chat > Anthropic > Responses 顺位选择转换目标"分发请求，使三种客户端协议均可访问三种上游方言。

#### Scenario: Responses 客户端访问 chat-only 上游
- **WHEN** 客户端请求 Responses 端点，Provider 仅配置 OpenAIChatCompletionsBaseURL（如 DeepSeek 类上游）
- **THEN** 系统执行 OR→OC 转换后调用上游 `chat/completions`

#### Scenario: Responses 客户端访问 Anthropic-only 上游
- **WHEN** 客户端请求 Responses 端点，Provider 仅配置 AnthropicMessagesBaseURL
- **THEN** 系统执行 OR→AM 转换后调用上游 `messages`

#### Scenario: Chat 客户端访问 responses-only 上游
- **WHEN** 客户端请求 Chat 端点，Provider 仅配置 OpenAIResponsesBaseURL
- **THEN** 系统执行 OC→OR 转换后调用上游 `responses`

#### Scenario: Anthropic 客户端访问 responses-only 上游
- **WHEN** 客户端请求 Messages 端点，Provider 仅配置 OpenAIResponsesBaseURL
- **THEN** 系统执行 AM→OR 转换后调用上游 `responses`

### Requirement: 跨协议转换完整性

跨协议转换 SHALL 完整传递对话输入、输出内容、工具定义与工具调用结果，并在 SSE 场景下保持事件流语义等价（流式与非流式均适用）。

#### Scenario: 多轮工具调用经 OR→OC 转换
- **WHEN** Responses 客户端携带 function_call 输出发起下一轮请求，且上游为 chat 方言
- **THEN** 转换后上游收到等价的 `tool_calls` 结构
- **AND** 上游返回的工具调用被转换回 Responses 的 `function_call` output item

#### Scenario: 流式事件转换
- **WHEN** 转换路径上上游以 SSE 返回增量内容
- **THEN** 客户端收到语义等价的本协议 SSE 事件流，且终止事件正常送达

#### Scenario: Anthropic thinking 跨轮保真（信封往返）
- **WHEN** Responses 客户端请求 Anthropic 上游，上游返回带签名的 thinking 块，客户端将 reasoning item（含 `encrypted_content`）原样回传
- **THEN** 系统解出信封并在请求中按原序重建携带原样签名的 thinking 块，上游验签通过
- **AND** 信封缺失、损坏或与当前上游不匹配时，系统丢弃该 reasoning item 并继续请求，不返回错误

### Requirement: 流式与用量统计

系统 SHALL 转发 Responses SSE 事件流，并从 `response.completed` 事件（流式）或响应体（非流式）中提取 `input_tokens` 与 `output_tokens`，映射为现有调用日志的 token 字段记录。

#### Scenario: 非流式请求统计
- **WHEN** 请求以非流式完成且响应含 usage
- **THEN** 系统将 `input_tokens` 记为输入 token、`output_tokens` 记为输出 token 写入调用日志

#### Scenario: 流式请求统计
- **WHEN** 请求以流式完成
- **THEN** 系统在 `response.completed` 事件中提取 usage 并写入调用日志
- **AND** 事件流本身原样送达客户端

### Requirement: 管理端点结构化错误

系统 SHALL 注册 `GET/DELETE /responses/{id}`、`POST /responses/{id}/cancel`、`GET /responses/{id}/input_items` 及 `/conversations` 族路由，全部由网关直接应答 HTTP 400 与 `stateless_gateway` 错误体，SHALL NOT 转发至上游。

#### Scenario: 客户端检索历史响应
- **WHEN** 客户端携带有效 API Key 请求 `GET /responses/{id}`
- **THEN** 系统返回 400，错误体 `error.code` 为 `stateless_gateway`，message 说明网关不存储响应并指导改用 `store: false` 全量回放

#### Scenario: 取消后台任务请求
- **WHEN** 客户端携带有效 API Key 请求 `POST /responses/{id}/cancel`
- **THEN** 系统返回 400 与相同的 `stateless_gateway` 错误契约，且上游未收到任何请求

#### Scenario: 未鉴权访问管理端点
- **WHEN** 请求未携带有效 API Key
- **THEN** 系统返回与其余端点一致的鉴权错误
