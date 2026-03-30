# Design: Add Anthropic API Endpoint

## Context

AI-Proxy 是一个 AI 模型代理服务。系统使用 Gin 框架, 采用分层架构:

- `handler/` - HTTP 请求处理
- `manufacturer/` - 厂商抽象接口 (原 transformer/)
- `router/` - 模型路由
- `middleware/` - 认证等中间件
- `model/` - 数据模型

## Goals / Non-Goals

**Goals:**
- 实现 `/anthropic/v1/messages` 端点
- 支持 `x-api-key` header 认证
- 支持 Anthropic 请求到 OpenAI/Anthropic 后端
- 支持 OpenAI 请求到 Anthropic 后端
- 采用 Manufacturer 接口模式封装厂商逻辑

**Non-Goals:**
- 不实现其他 Anthropic API 端点
- 不修改现有 `/openai/v1` 接口

## Architecture: Manufacturer Pattern

### 核心思想

每个 Manufacturer 实现负责：
1. **认证**: 决定使用什么 header (Bearer / x-api-key)
2. **端点**: 决定请求发送到哪个 URL
3. **格式**: 根据后端类型决定是否转换
4. **模型同步**: 封装厂商特定的模型列表获取逻辑

### 接口定义

```go
type Manufacturer interface {
    Name() string
    ExecuteOpenAIRequest(ctx *gin.Context, model *model.ProviderModel) (int, error)
    ExecuteAnthropicRequest(ctx *gin.Context, model *model.ProviderModel) (int, error)
    SyncModels(provider *model.Provider) ([]model.ProviderModel, error)
}
```

返回值 `(int, error)` 中的 int 为 token 使用量：
- OpenAI 格式: `usage.total_tokens + completion_tokens_details.reasoning_tokens`
- Anthropic 格式: `usage.input_tokens + usage.output_tokens`

### 请求流程

```
Client Request
     │
     ▼
┌─────────────────────────────────┐
│         Handler                 │
│  1. Parse request              │
│  2. Route model                 │
│  3. Create manufacturer         │
│  4. Call Execute*Request()      │
└─────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────┐
│      Manufacturer               │
│  1. Check backend type         │
│  2. If different format: convert│
│  3. Set auth headers           │
│  4. Forward to backend          │
│  5. Copy response               │
└─────────────────────────────────┘
```

### 请求格式矩阵

| 请求格式 | 后端 Provider | 行为 |
|---------|--------------|------|
| OpenAI | OpenAI | Passthrough |
| OpenAI | Anthropic | 转换为 Anthropic 格式 |
| Anthropic | Anthropic | Passthrough |
| Anthropic | OpenAI | 转换为 OpenAI 格式 |

## Implementations

### OpenAICompatibleManufacturer (后端是 OpenAI)

```go
func (m *OpenAICompatibleManufacturer) ExecuteOpenAIRequest(c *gin.Context, pm *model.ProviderModel) error {
    // Passthrough - 直接转发
}

func (m *OpenAICompatibleManufacturer) ExecuteAnthropicRequest(c *gin.Context, pm *model.ProviderModel) error {
    // Anthropic → OpenAI 转换后转发
}
```

### AnthropicManufacturer (后端是 Anthropic)

```go
func (m *AnthropicManufacturer) ExecuteOpenAIRequest(c *gin.Context, pm *model.ProviderModel) error {
    // OpenAI → Anthropic 转换后转发
}

func (m *AnthropicManufacturer) ExecuteAnthropicRequest(c *gin.Context, pm *model.ProviderModel) error {
    // Passthrough - 直接转发
}
```

## Format Conversion Details

### Anthropic → OpenAI

**已实现转换:**
- `system` → `messages[0].role="system"` (支持字符串和数组格式)
- `max_tokens` → `max_tokens`
- `messages` → `messages`
- `tools` → `tools` (input_schema → parameters)
- `stream` → `stream` + `stream_options.include_usage`

**多模态内容:**
- `content[].type="text"` → `content` 字符串或数组
- `content[].type="image"` → `content[].type="image_url"` (base64 或 URL)
- `content[].type="tool_use"` → `tool_calls[]`
- `content[].type="tool_result"` → `role="tool"` 消息

### OpenAI → Anthropic

**已实现转换:**
- `messages[].role="system"` → `system` (从 messages 移除)
- `messages` → `messages` (移除 system 消息)
- `max_tokens` → `max_tokens`
- `tools` → `tools`
- `stream` → `stream`

**流式响应 (OpenAI → Anthropic):**
- SSE 事件格式转换
- thinking 块支持 (reasoning_content → thinking_delta)
- tool_use 流式构建

## File Structure

```
server/internal/
├── manufacturer/
│   ├── manufacturer.go              # Interface: Manufacturer
│   ├── factory.go                   # Factory, Config, ProviderType constants
│   ├── manufacturer_openai_compatible.go
│   └── manufacturer_anthropic.go
│
├── handler/
│   ├── proxy_openai.go             # ProxyHandler.ChatCompletions()
│   ├── proxy_anthropic.go          # AnthropicProxyHandler.Messages()
│   └── provider_model.go           # ProviderModelHandler.Sync()
│
└── transformer/                     # DELETED
```

## Deleted: Transformer Package

旧的 Transformer 模式被删除，因为：
1. Handler 需要知道选择哪个 transformer
2. 每个 transformer 只支持单向转换
3. 不符合单一职责原则

Manufacturer 模式替代方案：
- 接口更简洁
- 每个实现知道自己的后端类型
- 自动选择 passthrough 或转换

## Risks / Trade-offs

### 1. 转换限制

**风险**: 复杂场景（如多模态）需要完整测试。

**已实现**: 
- 多模态内容转换 (image、thinking)
- 工具调用转换 (tool_use/tool_result)
- 流式响应转换 (OpenAI → Anthropic)

### 2. 流式响应

**状态**: 已实现 Anthropic 格式的流式响应转换

- `OpenAICompatibleManufacturer.ExecuteAnthropicRequest`: 实现了完整的 OpenAI → Anthropic 流式转换
  - message_start / message_delta / message_stop 事件
  - content_block_start / content_block_delta / content_block_stop 事件
  - thinking 内容块支持 (reasoning_content)
  - tool_use 工具调用流式支持

- `AnthropicManufacturer.ExecuteOpenAIRequest`: 目前直接透传流式响应
  - TODO: 实现 Anthropic → OpenAI 流式转换
