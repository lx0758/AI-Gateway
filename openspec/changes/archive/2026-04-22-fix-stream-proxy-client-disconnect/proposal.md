## Why

当 Gateway 代理流式请求时，如果客户端提前断开连接（例如用户取消请求、浏览器关闭），Gateway 无法及时感知到这一情况。这导致后端 API 会继续发送完整的响应数据，造成资源浪费（带宽、GPU tokens、计算资源）。

## What Changes

- 修改 `copyOpenAIStreaming`、`copyAnthropicStreaming`、`streamOpenAIToAnthropic`、`streamAnthropicToOpenAI` 四个函数，添加 context 参数支持
- 在流复制循环中，使用 `goroutine + channel` 包装阻塞读取，goroutine 内部先检查 context 是否取消
- 后端 HTTP 请求使用 `req.WithContext(c.Request.Context())` 传递 context，使后端请求也能响应取消
- 客户端断开时 goroutine 立即退出（不等待后端数据），主协程收到信号后立即返回

## Capabilities

### New Capabilities

- `stream-cancellation`: 流代理取消机制 - 当客户端断开时及时取消后端请求

### Modified Capabilities

- (无)

## Impact

**受影响的代码**:
- `server/internal/provider/provider_openai.go`: `ExecuteOpenAIRequest`, `copyOpenAIStreaming`, `streamOpenAIToAnthropic`
- `server/internal/provider/provider_anthropic.go`: `ExecuteAnthropicRequest`, `copyAnthropicStreaming`, `streamAnthropicToOpenAI`

**受影响的文件数**: 2
