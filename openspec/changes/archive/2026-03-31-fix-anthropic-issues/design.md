## Context

AI-Proxy 支持 OpenAI 和 Anthropic 两种协议格式，允许客户端使用各自的 SDK 调用任意后端模型。当前 Anthropic 厂商实现存在以下问题：

1. **Token 统计不准确**: `copyAnthropicStreaming` 和 `streamAnthropicToOpenAI` 函数 token 统计逻辑错误
2. **OpenAI 客户端请求 Anthropic 后端不可用**: 多个转换问题
3. **Tool calls 格式错误**: `id` 字段未在正确的时机发送
4. **Thinking 内容丢失**: 未将 Anthropic 的 `thinking_delta` 转换为 OpenAI 的 `reasoning_content`
5. **缓存 token 场景**: 未正确处理缓存读取和初始 output_tokens

### 约束

- 必须保持与 Anthropic API 的完全兼容
- 不能影响现有的 OpenAI → OpenAI 和 Anthropic → Anthropic 直通流程
- 需要正确处理流式和非流式两种模式

## Goals / Non-Goals

**Goals:**
- 修复 Anthropic 客户端请求 Anthropic 后端时的 token 统计问题
- 修复 OpenAI 客户端请求 Anthropic 后端的完整请求流程
- 确保 tool_calls 格式符合 OpenAI 规范
- 支持 thinking 内容的正确传递
- 确保 token 统计在所有场景下准确无误（含缓存 token）

**Non-Goals:**
- 不修改 API 接口定义
- 不添加新的功能特性
- 不修改其他厂商的实现

## Decisions

### 1. Token 统计修复方案

**问题**: Anthropic API 的 token 统计逻辑复杂
- `message_start.message.usage.input_tokens` 是初始输入 token 数
- `message_start.message.usage.output_tokens` 可能为 0 或非 0（如缓存 thinking）
- `message_delta.usage.input_tokens` 可能包含缓存读取的 token（非 0）
- `message_delta.usage.output_tokens` 是最终的输出 token 数

**方案**:
- `message_start`: 记录 `input_tokens` 和 `output_tokens`
- `message_delta`: 累加 `input_tokens` 和 `output_tokens`
- 在 `message_delta` 时发送包含 `usage` 字段的 OpenAI chunk
- 最终 `prompt_tokens` = 初始 input + delta input
- 最终 `completion_tokens` = 初始 output + delta output

### 2. Tool Calls 格式修复

**问题**: OpenAI 流式格式要求 `id` 在 tool_call 首次出现时就发送

**方案**:
- 在 `content_block_start` 事件（block 类型为 `tool_use`）时立即发送包含 `id`、`type`、`function.name` 的初始 chunk
- 后续 `content_block_delta` 事件只发送 `function.arguments` 增量

### 3. Stop Reason 提取修复

**问题**: `stop_reason` 嵌套在 `delta` 对象内，而非根级别

**方案**:
- 从 `event.Delta["stop_reason"]` 提取，而非 `event.StopReason`

### 4. Thinking 内容支持

**问题**: Anthropic 的 `thinking_delta` 需要转换为 OpenAI 的 `reasoning_content`

**方案**:
- 流式: `thinking_delta` → `delta.reasoning_content`
- 非流式: `thinking` block → `message.reasoning_content`

### 5. Tool Result 格式修复

**问题**: Anthropic 的 tool_result 要求特定格式

**方案**:
```go
{
  "role": "user",
  "content": [{
    "type": "tool_result",
    "tool_use_id": "xxx",
    "content": "result"
  }]
}
```

### 6. Arguments 序列化修复

**问题**: OpenAI 的 `tool_calls[].function.arguments` 必须是 JSON 字符串

**方案**:
- 将 `block["input"]` 对象序列化为 JSON 字符串后再赋值给 `arguments`

### 7. OpenAI Compatible 测试覆盖

**问题**: OpenAI Compatible 厂商同样需要测试覆盖

**方案**:
- 新增 `manufacturer_openai_compatible_test.go`
- 覆盖流式透传、非流式透传、请求转换、响应转换、流式转换等场景
- 特别覆盖 `reasoning_tokens` 和 `cached_tokens` 场景

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| 类型断言失败导致 token 统计为 0 | 使用 `float64` 类型断言，匹配 Go 的 JSON 解析行为 |
| 流式响应格式不兼容 | 添加 42 个单元测试覆盖各种场景 |
| Thinking 内容格式差异 | 使用 `reasoning_content` 字段，与 OpenAI 扩展格式一致 |
| 缓存 token 计算错误 | 添加专门测试用例验证缓存 token 场景 |
