# anthropic-api 规格

## Purpose

提供与 Anthropic Messages API 兼容的代理接口，允许使用 Anthropic SDK 的客户端（如 Claude Code、Cursor）直接连接到本代理服务；涵盖端点行为、x-api-key 认证、跨协议请求与响应转换、Token 统计与错误处理。

## Requirements

### Requirement: Token 统计准确性

系统 SHALL 在所有请求场景下准确统计 input_tokens 和 output_tokens。

#### Scenario: Anthropic 客户端请求 Anthropic 后端（流式）
- **WHEN** Anthropic 客户端发送流式请求到 Anthropic 后端
- **THEN** 系统 SHALL 正确解析 `message_start` 和 `message_delta` 事件中的 token 数据
- **AND** 系统 SHALL 返回准确的 token 统计数量

#### Scenario: OpenAI 客户端请求 Anthropic 后端（流式）
- **WHEN** OpenAI 客户端发送流式请求到 Anthropic 后端
- **THEN** 系统 SHALL 正确转换 Anthropic SSE 格式为 OpenAI SSE 格式
- **AND** 系统 SHALL 在最后的 chunk 中包含 `usage` 字段
- **AND** 系统 SHALL 返回准确的 `prompt_tokens`、`completion_tokens` 和 `total_tokens`

#### Scenario: OpenAI 客户端请求 Anthropic 后端（非流式）
- **WHEN** OpenAI 客户端发送非流式请求到 Anthropic 后端
- **THEN** 系统 SHALL 正确转换 Anthropic 响应为 OpenAI 格式
- **AND** 系统 SHALL 返回准确的 token 统计数量

### Requirement: 跨协议请求兼容性

系统 SHALL 支持 OpenAI 客户端请求 Anthropic 后端模型的完整流程。

#### Scenario: 基本文本对话
- **WHEN** OpenAI 客户端请求 Anthropic 模型进行文本对话
- **THEN** 系统 SHALL 正确转换请求格式从 OpenAI 到 Anthropic
- **AND** 系统 SHALL 正确转换响应格式从 Anthropic 到 OpenAI

#### Scenario: 工具调用
- **WHEN** OpenAI 客户端请求 Anthropic 模型使用工具
- **THEN** 系统 SHALL 正确转换 `tools` 参数
- **AND** 系统 SHALL 在 `content_block_start` 时发送包含 `id` 的 tool_call chunk
- **AND** 系统 SHALL 正确转换 `tool_calls` 响应
- **AND** `arguments` 字段 SHALL 为 JSON 字符串格式

#### Scenario: 工具结果返回
- **WHEN** OpenAI 客户端返回工具调用结果
- **THEN** 系统 SHALL 正确转换为 Anthropic 的 `tool_result` 格式
- **AND** `type: "tool_result"` 和 `tool_use_id` SHALL 在 content block 内部

#### Scenario: 多模态输入
- **WHEN** OpenAI 客户端发送包含图片的请求到 Anthropic 后端
- **THEN** 系统 SHALL 正确转换图片格式（base64 或 URL）

### Requirement: Thinking 内容传递

系统 SHALL 正确传递 Anthropic 的 thinking 内容到 OpenAI 客户端。

#### Scenario: 流式 thinking 内容
- **WHEN** Anthropic 后端返回 `thinking_delta` 事件
- **THEN** 系统 SHALL 转换为 OpenAI 格式的 `reasoning_content` 字段

#### Scenario: 非流式 thinking 内容
- **WHEN** Anthropic 后端返回包含 `thinking` block 的响应
- **THEN** 系统 SHALL 在 OpenAI 响应的 `message.reasoning_content` 字段中传递内容

## 参考附录：协议与端点说明

### 端点与认证

- `POST /anthropic/v1/messages`：Anthropic Messages API 兼容端点，使用 `x-api-key: <api-key>` header 认证
- 请求体关键字段：`model`、`max_tokens`、`system`、`messages`、`stream`、`tools`
- 非流式响应为 Anthropic Message 结构（含 `usage.input_tokens` / `usage.output_tokens`）；流式响应为 SSE 事件序列：`message_start` → `content_block_start` → `content_block_delta` → `content_block_stop` → `message_delta` → `message_stop`

### 请求路由

1. 从请求 body 提取 `model` 字段，经 `ModelRouter.Route(model)` 解析模型别名，获取 `ProviderModel` 与 `Provider`
2. 权限检查：API Key 有效、未过期、配额未超限；Key 关联模型权限列表时校验请求模型在允许列表中

### 请求转换映射

| Anthropic → OpenAI（后端为 OpenAI） | 说明 |
|---------------|-------------|
| `system` → `messages[0].role="system"` | 插入到 messages 首位 |
| `messages` → `messages` | 保留非 system 消息 |
| `max_tokens` → `max_tokens` | 直接映射 |
| `stream` → `stream` + `stream_options.include_usage` | 添加 usage 选项 |
| `tools[].input_schema` → `tools[].function.parameters` | 重命名 |

内容块转换：`text` → `content` 字符串；`image`（source.media_type/data）→ `image_url` data URI；`tool_use` → `tool_calls[].function`；`tool_result` → `role: "tool"` message。

反向（OpenAI → Anthropic，后端为 Anthropic）：`messages[role="system"]` 提取为独立 `system` 字段，其余直接映射。

### 流式响应转换

- OpenAI → Anthropic：`delta.content` → `text_delta`；`delta.reasoning_content` → `thinking_delta`；`delta.tool_calls` → `input_json_delta` + tool_use block
- Anthropic → OpenAI：`text_delta` → `delta.content`；`thinking_delta` → `delta.reasoning_content`；`input_json_delta` → `tool_calls[].function.arguments`；`tool_use` block → `tool_calls[]`（含 id / type / function）
- Token 统计：`message_start.message.usage.input_tokens` → `usage.prompt_tokens`；`message_delta.usage.output_tokens` → `usage.completion_tokens`；在 `message_delta` 时发送包含完整 `usage` 的 chunk

### 错误处理

| HTTP 状态码 | 场景 |
|------------|------|
| 400 | 请求格式错误 |
| 401 | API Key 无效或缺失 |
| 403 | 模型权限不足 |
| 404 | 模型未找到 |
| 429 | 配额超限 |
| 500 | 内部错误 / 后端错误 |

### 与 OpenAI 端点的差异

| 方面 | OpenAI 端点 | Anthropic 端点 |
|------|-------------|---------------|
| 路径 | `/openai/v1/chat/completions` | `/anthropic/v1/messages` |
| 认证 Header | `Authorization: Bearer <key>` | `x-api-key: <key>` |
| 请求格式 | OpenAI Chat Completions | Anthropic Messages |
| 响应格式 | OpenAI 格式 | Anthropic 格式 |
