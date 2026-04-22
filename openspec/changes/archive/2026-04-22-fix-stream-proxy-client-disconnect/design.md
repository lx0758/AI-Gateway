## Context

当前流式请求代理实现存在客户端断开检测问题。当用户取消请求或连接中断时，Gateway 继续从后端读取完整响应，造成资源浪费。

```
┌─────────────────────────────────────────────────────────────────────┐
│                      当前问题流程                                    │
└─────────────────────────────────────────────────────────────────────┘

  客户端              Gateway                  后端 API
    │                   │                         │
    │── POST /chat ────▶│── POST /chat ─────────▶│
    │                   │                         │
    │◀─ SSE Stream ─────│◀─ SSE Stream ──────────│
    │                   │                         │
    │  [断开连接]        │                         │
    │                   │  继续读取数据...         │  继续发送数据...
    │                   │  fmt.Fprint() 不会报错   │  不知道客户端断开
    │                   │                         │
    │                   │  读取完整响应后才退出     │  整个响应已发送
    │                   │                         │
```

受影响的函数：
- `provider_openai.go`: `copyOpenAIStreaming`, `streamOpenAIToAnthropic`
- `provider_anthropic.go`: `copyAnthropicStreaming`, `streamAnthropicToOpenAI`

## Goals / Non-Goals

**Goals:**
- 客户端断开时立即停止从后端读取数据
- 后端 HTTP 请求能够响应 context 取消
- 最小化实现复杂度，不引入新依赖

**Non-Goals:**
- 不修改后端 API 的行为
- 不实现重试或恢复机制
- 不处理非流式请求的特殊情况（它们已经可以正常工作）

## Decisions

### Decision 1: 使用 Context 取消机制

**选择**: 在流复制函数中添加 `context.Context` 参数，通过 `select + channel` 检测取消。

**理由**:
- Gin 框架原生支持 `c.Request.Context()`，当客户端断开时 context 会被取消
- 这是 Go 语言处理取消的标准模式
- 无需引入新依赖

**替代方案考虑**:
- `net.Conn.SetReadDeadline`: 需要从 `io.Reader` 提取底层连接，实现复杂
- 检测写入错误: 有效但不完整，无法取消后端请求

### Decision 2: 使用 Channel 包装阻塞读取（改进版）

**选择**: 使用 `goroutine + channel` 将阻塞的 `ReadString` 调用包装为可取消操作，goroutine 内部先检查 context 是否取消。

```go
type readResult struct {
    line string
    err  error
}
readCh := make(chan readResult, 1)

go func() {
    select {
    case <-ctx.Done():
        readCh <- readResult{err: ctx.Err()}
    default:
        line, err := reader.ReadString('\n')
        readCh <- readResult{line: line, err: err}
    }
}()

select {
case <-ctx.Done():
    return ctx.Err()
case result := <-readCh:
    if result.err == context.Canceled {
        return ctx.Err()
    }
    // 处理结果
}
```

**理由**:
- `bufio.Reader.ReadString` 是阻塞调用，无法直接响应 context 取消
- goroutine 内部先检查 context，收到取消信号后立即返回，不等待后端数据
- 避免 goroutine 泄漏：客户端断开时 goroutine 立即退出
- channel receive 可以同时等待数据到达和 context 取消

### Decision 3: 后端请求传递 context

**选择**: `req = req.WithContext(c.Request.Context())`

**理由**:
- 部分 HTTP 后端支持 context 取消（如 OpenAI API）
- 即使后端不支持，取消读取后 connection 也会被关闭
- 改动最小，仅需一行

## Risks / Trade-offs

[Risk] goroutine 泄漏
→ [Mitigation] goroutine 内部先检查 context，取消时立即退出，不等待后端数据

[Risk] 引入 goroutine 增加复杂度
→ [Mitigation] 每个流复制函数只启动一个 goroutine，生命周期清晰

[Risk] 性能略有下降（channel 通信开销）
→ [Mitigation] 相比节省的后端资源，开销可忽略

## Migration Plan

1. 修改 `copyOpenAIStreaming` 添加 context 参数和取消检测
2. 修改 `copyAnthropicStreaming` 添加 context 参数和取消检测
3. 修改 `streamOpenAIToAnthropic` 添加 context 参数和取消检测
4. 修改 `streamAnthropicToOpenAI` 添加 context 参数和取消检测
5. 修改各 `Execute*Request` 函数，传递 context 给后端请求
6. 更新所有调用处，传入 `c.Request.Context()`

## Open Questions

- 无
