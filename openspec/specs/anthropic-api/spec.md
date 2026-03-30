# Spec: Anthropic API Endpoint

## 概述

提供与 Anthropic Messages API 兼容的代理接口，允许使用 Anthropic SDK 的客户端（如 Claude Code、Cursor）直接连接到本代理服务。

## 端点

### POST /anthropic/v1/messages

Anthropic Messages API 兼容端点。

#### 认证

使用 `x-api-key` header 进行认证：

```
x-api-key: <api-key>
```

#### 请求格式

```json
{
  "model": "claude-3-opus-20240229",
  "max_tokens": 4096,
  "system": "You are a helpful assistant.",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "stream": false,
  "tools": []
}
```

#### 响应格式

**非流式响应：**

```json
{
  "id": "msg_xxx",
  "type": "message",
  "role": "assistant",
  "model": "claude-3-opus-20240229",
  "content": [
    {"type": "text", "text": "Hello! How can I help you today?"}
  ],
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 10,
    "output_tokens": 20
  }
}
```

**流式响应：**

SSE 事件序列：
- `message_start` - 消息开始
- `content_block_start` - 内容块开始
- `content_block_delta` - 内容增量
- `content_block_stop` - 内容块结束
- `message_delta` - 消息增量（包含 stop_reason）
- `message_stop` - 消息结束

## 请求路由

### 模型解析

1. 从请求 body 中提取 `model` 字段
2. 通过 `ModelRouter.Route(model)` 解析模型别名
3. 获取 `ProviderModel` 和 `Provider` 信息
4. 根据 `Provider.APIType` 选择对应的 Manufacturer

### 权限检查

1. 检查 API Key 是否有效、未过期、配额未超限
2. 如果 API Key 关联了模型权限列表，检查请求模型是否在允许列表中

## 格式转换

### Anthropic → OpenAI (后端为 OpenAI)

当后端 Provider 类型为 `openai` 时，需要将 Anthropic 请求转换为 OpenAI 格式。

| Anthropic 字段 | OpenAI 字段 | 说明 |
|---------------|-------------|------|
| `system` | `messages[0].role="system"` | 插入到 messages 首位 |
| `messages` | `messages` | 保留非 system 消息 |
| `max_tokens` | `max_tokens` | 直接映射 |
| `stream` | `stream` + `stream_options.include_usage` | 添加 usage 选项 |
| `tools[].input_schema` | `tools[].function.parameters` | 重命名 |

**内容块转换：**

| Anthropic content type | OpenAI format |
|------------------------|---------------|
| `type: "text"` | `content: string` 或 `content[].type: "text"` |
| `type: "image"` + `source.media_type/data` | `content[].type: "image_url"` + `image_url.url: "data:..."` |
| `type: "tool_use"` | `tool_calls[].function` |
| `type: "tool_result"` | `role: "tool"` message |

### OpenAI → Anthropic (后端为 Anthropic)

当后端 Provider 类型为 `anthropic` 时，需要将 OpenAI 请求转换为 Anthropic 格式。

| OpenAI 字段 | Anthropic 字段 | 说明 |
|-------------|---------------|------|
| `messages[role="system"]` | `system` | 提取为独立字段 |
| `messages` (非 system) | `messages` | 保留 |
| `max_tokens` | `max_tokens` | 直接映射 |
| `stream` | `stream` | 直接映射 |

## 流式响应转换

### OpenAI → Anthropic (OpenAICompatibleManufacturer)

将 OpenAI SSE 格式转换为 Anthropic SSE 格式：

```
OpenAI: data: {"choices":[{"delta":{"content":"Hi"}}]}
    ↓
Anthropic: event: content_block_delta
          data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hi"}}
```

**支持的内容类型：**
- `delta.content` → `text_delta`
- `delta.reasoning_content` → `thinking_delta`
- `delta.tool_calls` → `input_json_delta` + tool_use block

### Anthropic → OpenAI (AnthropicManufacturer)

当前实现：直接透传 Anthropic SSE 流

TODO: 实现 Anthropic → OpenAI 流式转换

## 错误处理

| HTTP 状态码 | 场景 |
|------------|------|
| 400 | 请求格式错误 |
| 401 | API Key 无效或缺失 |
| 403 | 模型权限不足 |
| 404 | 模型未找到 |
| 429 | 配额超限 |
| 500 | 内部错误 / 后端错误 |

## 与 OpenAI 端点的差异

| 方面 | OpenAI 端点 | Anthropic 端点 |
|------|-------------|---------------|
| 路径 | `/openai/v1/chat/completions` | `/anthropic/v1/messages` |
| 认证 Header | `Authorization: Bearer <key>` | `x-api-key: <key>` |
| 请求格式 | OpenAI Chat Completions | Anthropic Messages |
| 响应格式 | OpenAI 格式 | Anthropic 格式 |

## 依赖

- `ModelRouter` - 模型路由解析
- `Manufacturer` 接口 - 厂商抽象
- `model.DB` - 数据访问（API Key、模型权限）
