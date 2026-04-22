## 1. 修改流复制函数签名

- [x] 1.1 修改 `copyOpenAIStreaming` 添加 `ctx context.Context` 参数
- [x] 1.2 修改 `copyAnthropicStreaming` 添加 `ctx context.Context` 参数
- [x] 1.3 修改 `streamOpenAIToAnthropic` 添加 `ctx context.Context` 参数
- [x] 1.4 修改 `streamAnthropicToOpenAI` 添加 `ctx context.Context` 参数

## 2. 实现可取消的读取模式

- [x] 2.1 在 `copyOpenAIStreaming` 中实现 `goroutine + channel` 读取模式，goroutine 内部先检查 context
- [x] 2.2 在 `copyAnthropicStreaming` 中实现 `goroutine + channel` 读取模式，goroutine 内部先检查 context
- [x] 2.3 在 `streamOpenAIToAnthropic` 中实现 `goroutine + channel` 读取模式，goroutine 内部先检查 context
- [x] 2.4 在 `streamAnthropicToOpenAI` 中实现 `goroutine + channel` 读取模式，goroutine 内部先检查 context

## 3. 后端请求传递 context

- [x] 3.1 在 `ExecuteOpenAIRequest` 中为后端请求添加 `req = req.WithContext(c.Request.Context())`
- [x] 3.2 在 `ExecuteAnthropicRequest` 中为后端请求添加 `req = req.WithContext(c.Request.Context())`

## 4. 更新调用处

- [x] 4.1 更新 `ExecuteOpenAIRequest` 中对 `copyOpenAIStreaming` 的调用，传入 `c.Request.Context()`
- [x] 4.2 更新 `ExecuteOpenAIRequest` 中对 `streamOpenAIToAnthropic` 的调用，传入 `c.Request.Context()`
- [x] 4.3 更新 `ExecuteAnthropicRequest` 中对 `streamAnthropicToOpenAI` 的调用，传入 `c.Request.Context()`
- [x] 4.4 更新 `ExecuteAnthropicRequest` 中对 `copyAnthropicStreaming` 的调用，传入 `c.Request.Context()`
