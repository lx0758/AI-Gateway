## 1. Token 统计修复

- [x] 1.1 修复 `copyAnthropicStreaming` 函数中的 token 统计逻辑
  - `message_start`: 记录 `input_tokens` 和 `output_tokens`
  - `message_delta`: 累加 `input_tokens` 和 `output_tokens`
- [x] 1.2 修复 `streamAnthropicToOpenAI` 函数中的 token 统计
  - 修正 JSON 类型断言（`float64` 而非 `int`）
  - 在 `message_delta` 时发送包含 `usage` 字段的 OpenAI chunk
  - 正确处理缓存 token 和初始 output_tokens

## 2. 流式响应转换修复

- [x] 2.1 修复 tool_calls 格式
  - 在 `content_block_start` 时发送包含 `id` 的初始事件
  - 解决 "Expected 'id' to be a string" 错误
- [x] 2.2 修复 stop_reason 提取
  - 从 `event.Delta["stop_reason"]` 提取
- [x] 2.3 添加 thinking 支持
  - `thinking_delta` → `reasoning_content` 转换

## 3. 非流式响应转换修复

- [x] 3.1 修复 `convertAnthropicResponseToOpenAI`
  - `arguments` 序列化为 JSON 字符串
  - 添加 `thinking` → `reasoning_content` 转换

## 4. 请求格式转换修复

- [x] 4.1 修复 `convertOpenAIToolResultToAnthropic`
  - 修正 tool_result 格式，`type` 和 `tool_use_id` 放入 content block

## 5. 测试验证

### 5.1 Anthropic 厂商测试用例 (17 个)

**基础测试**
- [x] `TestAnthropicUsage_Total`: token 计算测试 (4 子测试)
- [x] `TestParseDataURL`: 图片 URL 解析测试 (3 子测试)

**流式透传测试**
- [x] `TestCopyAnthropicStreaming_TokenCounting`: 透传流式 token 测试
- [x] `TestCopyAnthropicStreaming_WithOutputTokensAtStart`: 初始 output_tokens 场景

**流式转换测试**
- [x] `TestStreamAnthropicToOpenAI_TokenCounting`: 转换流式 token 测试
- [x] `TestStreamAnthropicToOpenAI_ToolCalls`: 工具调用转换测试
- [x] `TestStreamAnthropicToOpenAI_Thinking`: thinking 内容测试
- [x] `TestStreamAnthropicToOpenAI_Usage`: usage 字段输出测试
- [x] `TestStreamAnthropicToOpenAI_UsageWithOutputTokensAtStart`: 初始 output_tokens 场景
- [x] `TestStreamAnthropicToOpenAI_UsageWithCacheTokens`: 缓存 token 场景

**非流式转换测试**
- [x] `TestConvertAnthropicResponseToOpenAI`: 非流式响应测试
- [x] `TestConvertAnthropicResponseToOpenAI_ToolUse`: 非流式工具调用测试
- [x] `TestConvertAnthropicResponseToOpenAI_WithThinking`: thinking block 转换测试
- [x] `TestConvertOpenAIToolResultToAnthropic`: tool_result 格式测试

### 5.2 OpenAI Compatible 厂商测试用例 (25 个)

**基础测试**
- [x] `TestOpenAIUsage_Total`: token 计算测试 (4 子测试)
- [x] `TestExtractSystemContent`: system 内容提取测试 (3 子测试)

**流式透传测试**
- [x] `TestCopyOpenAIStreaming_TokenCounting`: 透传流式 token 测试
- [x] `TestCopyOpenAIStreaming_WithReasoning`: reasoning_tokens 测试
- [x] `TestCopyOpenAIStreaming_WithReasoningTokens`: 详细 reasoning 测试
- [x] `TestCopyOpenAIStreaming_WithCachedTokens`: cached_tokens 测试

**非流式透传测试**
- [x] `TestCopyOpenAIResponse_TokenCounting`: 透传非流式测试
- [x] `TestCopyOpenAIResponse_WithReasoningTokens`: reasoning_tokens 测试

**请求转换测试**
- [x] `TestConvertAnthropicMessageToOpenAI_Text`: 文本消息转换
- [x] `TestConvertAnthropicMessageToOpenAI_Blocks`: 多内容块转换
- [x] `TestConvertAnthropicMessageToOpenAI_ToolUse`: 工具调用转换
- [x] `TestConvertAnthropicToolResultToOpenAI`: tool_result 转换
- [x] `TestConvertAnthropicToolToOpenAI`: tool 定义转换

**响应转换测试**
- [x] `TestConvertOpenAIResponseToAnthropic`: 非流式响应转换
- [x] `TestConvertOpenAIResponseToAnthropic_ToolCalls`: 工具调用转换
- [x] `TestConvertOpenAIResponseToAnthropic_WithReasoningContent`: reasoning_content 转换
- [x] `TestConvertOpenAIResponseToAnthropic_EmptyContent`: 空内容处理

**流式转换测试**
- [x] `TestStreamOpenAIToAnthropic_Text`: 文本流式转换
- [x] `TestStreamOpenAIToAnthropic_Thinking`: thinking 流式转换
- [x] `TestStreamOpenAIToAnthropic_ToolCalls`: 工具调用流式转换
- [x] `TestStreamOpenAIToAnthropic_WithReasoningTokens`: reasoning_tokens 场景
- [x] `TestStreamOpenAIToAnthropic_ThinkingAndText`: thinking + text 混合场景
- [x] `TestStreamOpenAIToAnthropic_UsageWithReasoning`: usage 计算场景

### 5.3 手动测试验证

- [x] 普通文本对话 ✓
- [x] 流式响应 ✓
- [x] 工具调用 ✓
- [x] Thinking 内容 ✓
- [x] Token 统计 ✓

## 测试统计

| 厂商 | 测试用例数 | 子测试数 | 状态 |
|------|-----------|---------|------|
| Anthropic | 13 | 7 | ✓ 全部通过 |
| OpenAI Compatible | 19 | 7 | ✓ 全部通过 |
| **总计** | **32** | **14** | **✓ 全部通过** |
