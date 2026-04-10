## Context

当前系统架构：

```
HTTP Request → Gin Context → Provider.ExecuteXxxRequest() → HTTP Response
```

测试功能需要复用现有的 Provider 执行逻辑，而非重新实现。通过构造 `httptest.ResponseRecorder` 和 `gin.CreateTestContext()`，可以在测试场景下调用现有代码。

现有 Provider 接口支持双协议：
- `ExecuteOpenAIRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error`
- `ExecuteAnthropicRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error`

## Goals / Non-Goals

**Goals:**
- 厂商模型测试：验证 Provider 下某个具体 ProviderModel 是否可用
- 虚拟模型测试：验证 Model（Alias）的完整路由链路
- 支持双协议测试（OpenAI + Anthropic）
- 复用现有 Provider 执行代码

**Non-Goals:**
- 不实现流式测试（非流式响应更易解析）
- 不持久化测试结果
- 不实现批量并发测试
- 不实现 Provider 连通性测试（另行处理）

## Decisions

### 1. 测试请求格式

**决定：固定非流式请求，消息 "Hi"，max_tokens 100**

理由：
- 非流式响应完整，易于解析和展示
- 简单消息足够验证连通性
- 限制 token 数量节省费用

### 2. 协议测试策略

**决定：Provider 支持什么协议就测什么协议**

```
Provider.OpenAIBaseURL != "" → 测试 OpenAI 协议
Provider.AnthropicBaseURL != "" → 测试 Anthropic 协议
```

双协议 Provider 会执行两次测试，分别验证两种协议。

### 3. 执行方式

**决定：顺序执行**

理由：
- 避免对上游 API 造成压力
- 结果按 mapping weight 排序展示
- 实现简单

### 4. 代码复用方式

**决定：使用 httptest 模拟 Gin Context**

```go
w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)
c.Request = httptest.NewRequest("POST", "/", bytes.NewReader(bodyBytes))

providerImpl.ExecuteOpenAIRequest(c, pm, usage)

respBody, _ := io.ReadAll(w.Body)
```

这样完全复用现有 `provider_*.go` 的执行逻辑，包括协议转换。

### 5. API 设计

**厂商模型测试：**
```
POST /api/v1/providers/:id/models/:model_id/test
Response: { tests: [{protocol, success, latency_ms, tokens, response, error}] }
```

**虚拟模型测试：**
```
POST /api/v1/models/:id/test
Response: { tests: [{mapping_id, weight, provider, provider_model, protocol_tests: [...]}] }
```

## Risks / Trade-offs

| 风险 | 缓解措施 |
|-----|---------|
| 测试消耗 API 费用 | 限制 max_tokens=100，文档说明 |
| 测试超时 | 设置 30 秒超时，显示超时错误 |
| 上游 API 限流 | 顺序执行，不并发 |
| 响应解析失败 | 返回原始响应体，便于调试 |
