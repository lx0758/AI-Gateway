## ADDED Requirements

### Requirement: Stream proxy cancellation on client disconnect
当 Gateway 代理流式请求时，如果客户端断开连接，系统 SHALL 立即停止从后端读取数据并退出流复制循环。

#### Scenario: Client disconnects during OpenAI stream proxy
- **WHEN** 客户端通过 OpenAI 兼容接口请求流式响应，且在收到部分数据后断开连接
- **THEN** Gateway SHALL 在检测到客户端断开后立即退出 `copyOpenAIStreaming` 循环

#### Scenario: Client disconnects during Anthropic stream proxy
- **WHEN** 客户端通过 Anthropic 兼容接口请求流式响应，且在收到部分数据后断开连接
- **THEN** Gateway SHALL 在检测到客户端断开后立即退出 `copyAnthropicStreaming` 循环

#### Scenario: Client disconnects during OpenAI to Anthropic conversion
- **WHEN** 客户端请求将 OpenAI 流转换为 Anthropic 格式，且在转换过程中断开连接
- **THEN** Gateway SHALL 在检测到客户端断开后立即退出 `streamOpenAIToAnthropic` 循环

#### Scenario: Client disconnects during Anthropic to OpenAI conversion
- **WHEN** 客户端请求将 Anthropic 流转换为 OpenAI 格式，且在转换过程中断开连接
- **THEN** Gateway SHALL 在检测到客户端断开后立即退出 `streamAnthropicToOpenAI` 循环

### Requirement: Backend request respects context cancellation
当客户端断开连接时，发往后端的 HTTP 请求 SHALL 能够被取消或最终终止。

#### Scenario: Backend request uses request context
- **WHEN** Gateway 向后端发起 HTTP 请求时
- **THEN** 请求 SHALL 使用 `req.WithContext(c.Request.Context())` 绑定客户端的 context

### Requirement: Context passed to stream copy functions
所有流复制函数 SHALL 接收 `context.Context` 参数以检测客户端断开。

#### Scenario: Stream functions receive context
- **WHEN** 调用流复制函数时
- **THEN** 调用方 SHALL 传递 `c.Request.Context()` 作为第一个参数

### Requirement: Goroutine checks context before blocking read
Goroutine 在调用阻塞的 `ReadString` 前 SHALL 先检查 context 是否已取消，避免goroutine 泄漏。

#### Scenario: Goroutine exits immediately on context cancellation
- **WHEN** 客户端断开连接，context 被取消
- **THEN** Goroutine SHALL 立即收到 ctx.Done() 信号并返回，不等待后端数据
