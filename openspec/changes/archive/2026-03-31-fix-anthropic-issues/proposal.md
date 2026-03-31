## Why

Anthropic 厂商实现存在多个问题：
1. Anthropic 客户端请求 Anthropic 后端时 token 统计不准确
2. OpenAI 客户端请求 Anthropic 后端完全不可用
3. Tool calls 格式转换错误导致客户端报错
4. Thinking 内容未正确传递给客户端
5. 缓存 token 和初始 output_tokens 场景处理不当

这些问题影响了系统的可靠性和跨协议兼容性，需要立即修复以确保生产环境的稳定性。

## What Changes

### Token 统计修复
- `copyAnthropicStreaming`: 修正 token 统计逻辑，正确处理 `message_start` 和 `message_delta` 中的 token 数据
- `streamAnthropicToOpenAI`: 修正 JSON 类型断言（`float64` 而非 `int`），在 `message_delta` 时发送包含 `usage` 字段的 chunk
- 支持缓存 token 场景（`message_delta` 中 `input_tokens` 不为 0）
- 支持初始 output_tokens 场景（`message_start` 中 `output_tokens` 不为 0，如缓存 thinking）

### 流式响应转换修复
- `streamAnthropicToOpenAI`: 在 `content_block_start` 时立即发送包含 `id` 的 tool_call 初始事件（解决 "Expected 'id' to be a string" 错误）
- `streamAnthropicToOpenAI`: 修正 `stop_reason` 从 `event.Delta` 中提取（而非根级别）
- `streamAnthropicToOpenAI`: 添加 `thinking_delta` → `reasoning_content` 转换支持

### 非流式响应转换修复
- `convertAnthropicResponseToOpenAI`: 修正 `arguments` 字段序列化为 JSON 字符串（而非对象）
- `convertAnthropicResponseToOpenAI`: 添加 `thinking` → `reasoning_content` 转换支持

### 请求格式转换修复
- `convertOpenAIToolResultToAnthropic`: 修正 tool_result 格式，将 `type: "tool_result"` 和 `tool_use_id` 放入 content block 内部

### 测试覆盖
- 新增 `manufacturer_anthropic_test.go`：17 个测试用例
- 新增 `manufacturer_openai_compatible_test.go`：25 个测试用例
- 覆盖场景：
  - Token 统计准确性（含缓存 token、初始 output_tokens）
  - Tool calls 转换
  - Thinking 内容转换
  - Usage 字段输出
  - Reasoning tokens 计算

## Capabilities

### New Capabilities

无新增能力。

### Modified Capabilities

- `anthropic-api`: 修正 Anthropic 厂商实现的正确性和跨协议兼容性，支持 thinking 内容传递

## Impact

- **代码影响**: `server/internal/manufacturer/manufacturer_anthropic.go`
  - `copyAnthropicStreaming`: token 统计逻辑修正
  - `streamAnthropicToOpenAI`: 流式转换完整重构
  - `convertAnthropicResponseToOpenAI`: 非流式响应转换修正
  - `convertOpenAIToolResultToAnthropic`: tool_result 格式修正
- **新增文件**: 
  - `server/internal/manufacturer/manufacturer_anthropic_test.go` (17 测试用例)
  - `server/internal/manufacturer/manufacturer_openai_compatible_test.go` (25 测试用例)
- **API 影响**: `/v1/chat/completions` 端点请求 Anthropic 后端模型
- **用户影响**: 使用 OpenAI SDK 调用 Anthropic 模型的用户将能够正常使用服务，包括工具调用和 thinking 内容
