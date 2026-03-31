## MODIFIED Requirements

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
