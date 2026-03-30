# Proposal: Add Anthropic API Endpoint

## Why

目前系统仅支持 OpenAI 风格的 `/openai/v1/chat/completions` 接口。使用 Anthropic SDK 的客户端（如 Claude Code、Cursor 等)无法直接连接到本代理服务。

## What

新增 `/anthropic/v1/messages` 端点，提供与 Anthropic Messages API 兼容的接口。

## Architecture

采用 **Manufacturer 接口模式**：

```
Handler → Factory.Create(provider) → Manufacturer
                                       ├── ExecuteOpenAIRequest(ctx, model)
                                       ├── ExecuteAnthropicRequest(ctx, model)
                                       ├── Name()
                                       └── SyncModels(provider)
```

## Request Flow

| 请求格式 | 后端类型 | 行为 |
|---------|---------|------|
| OpenAI | OpenAI | Passthrough |
| OpenAI | Anthropic | OpenAI → Anthropic 转换 |
| Anthropic | Anthropic | Passthrough |
| Anthropic | OpenAI | Anthropic → OpenAI 转换 |

## Interface

```go
type Manufacturer interface {
    Name() string
    ExecuteOpenAIRequest(ctx *gin.Context, model *model.ProviderModel) (int, error)
    ExecuteAnthropicRequest(ctx *gin.Context, model *model.ProviderModel) (int, error)
    SyncModels(provider *model.Provider) ([]model.ProviderModel, error)
}
```

返回值 `(int, error)` 中的 int 为 token 使用量。

## Implementations

| Manufacturer | ExecuteOpenAIRequest | ExecuteAnthropicRequest |
|-------------|---------------------|------------------------|
| OpenAICompatibleManufacturer | Passthrough | Anthropic → OpenAI 转换 |
| AnthropicManufacturer | OpenAI → Anthropic 转换 | Passthrough |

## Format Conversion

### Anthropic → OpenAI

- `system` → `messages[0].role="system"`
- `max_tokens` → `max_tokens`
- `messages` → `messages`
- `tools` → `tools`
- `stream` → `stream` + `stream_options.include_usage`

### OpenAI → Anthropic

- `messages[].role="system"` → `system`
- `messages` → `messages` (移除 system)
- `max_tokens` → `max_tokens`
- `tools` → `tools`
- `stream` → `stream`

## Directory Structure

```
server/internal/
├── manufacturer/
│   ├── manufacturer.go              # Interface 定义
│   ├── factory.go                   # Factory + ProviderType 常量
│   ├── manufacturer_openai_compatible.go
│   └── manufacturer_anthropic.go
├── handler/
│   ├── proxy_openai.go             # 使用 manufacturer.ExecuteOpenAIRequest
│   ├── proxy_anthropic.go          # 使用 manufacturer.ExecuteAnthropicRequest
│   └── provider_model.go           # 使用 manufacturer.SyncModels
```

## Dependencies

- 复用现有的 `ModelRouter` 进行模型路由
- 复用现有的 `model.DB` 进行数据访问
- 删除旧的 `transformer/` 包

## Capabilities

| Capability | Type | Description |
|------------|------|-------------|
| anthropic-api | 新增 | 支持 Anthropic Messages API 格式的代理接口 |
| api-authentication | 修改 | 扩展支持 `x-api-key` header 认证方式 |
| manufacturer-pattern | 新增 | 抽象厂商逻辑的接口模式 |
| bidirectional-transform | 新增 | 支持 OpenAI ↔ Anthropic 双向格式转换 |
