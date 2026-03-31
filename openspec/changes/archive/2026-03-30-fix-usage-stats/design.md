# 设计文档: 修复使用量统计

## 验证结果

**stream_options.include_usage 特性验证** ✅

| Provider | 支持状态 | 返回格式 |
|----------|----------|----------|
| OpenRouter | ✅ 支持 | 最后一个 chunk 包含完整 usage |
| OpenAI | ✅ 支持 | 官方特性 |
| Anthropic | ⚠️ 不同格式 | message_start + message_delta 事件 |
| DeepSeek | 待验证 | 需要 Key |

**OpenRouter 实际响应示例**:
```json
data: {"id":"gen-xxx", "choices":[{"delta":{"content":"Hi!"}, "finish_reason":"length"}]}

data: {"choices":[], "usage":{
  "prompt_tokens": 18,
  "completion_tokens": 20,
  "total_tokens": 38,
  "prompt_tokens_details": {"cached_tokens": 0},
  "completion_tokens_details": {"reasoning_tokens": 16}
}}

data: [DONE]
```

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                        请求处理流程                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────┐    ┌─────────────┐    ┌─────────────────────────┐│
│  │ Client  │───▶│ ProxyHandler│───▶│ Transformer             ││
│  │ Request │    │             │    │ (注入 stream_options)   ││
│  └─────────┘    └─────────────┘    └─────────────────────────┘│
│                                              │                  │
│                                              ▼                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    上游 Provider                             ││
│  │  ┌─────────────────────────────────────────────────────┐   ││
│  │  │  Stream Response:                                    │   ││
│  │  │  data: {...chunk...}                                 │   ││
│  │  │  data: {...chunk...}                                 │   ││
│  │  │  data: {"choices":[],"usage":{"prompt_tokens":10,...}}│  ││
│  │  │  data: [DONE]                                        │   ││
│  │  └─────────────────────────────────────────────────────┘   ││
│  └─────────────────────────────────────────────────────────────┘│
│                                              │                  │
│                                              ▼                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  TransformStream (解析 usage)                               ││
│  │  ┌───────────────┐    ┌───────────────┐    ┌─────────────┐ ││
│  │  │ 读取 SSE 行   │───▶│ 提取 usage   │───▶│ 返回给调用者│ ││
│  │  └───────────────┘    └───────────────┘    └─────────────┘ ││
│  └─────────────────────────────────────────────────────────────┘│
│                                              │                  │
│                                              ▼                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  logUsage(model, actualModel, promptTokens, completionTok...││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 数据模型变更

### UsageLog 扩展

```go
type UsageLog struct {
    ID               uint      `gorm:"primaryKey" json:"id"`
    APIKeyID         uint      `gorm:"index" json:"key_id"`        // 移除 not null
    ProviderID       uint      `gorm:"index" json:"provider_id"`
    Model            string    `gorm:"size:128;not null" json:"model"`           // 用户请求的别名
    ActualModel      string    `gorm:"size:128" json:"actual_model"`             // 新增: 实际模型
    PromptTokens     int       `gorm:"default:0" json:"prompt_tokens"`
    CompletionTokens int       `gorm:"default:0" json:"completion_tokens"`
    LatencyMs        int       `gorm:"default:0" json:"latency_ms"`
    Status           string    `gorm:"size:32;not null" json:"status"`
    ErrorMsg         string    `gorm:"type:text" json:"error_msg"`
    CreatedAt        time.Time `gorm:"index" json:"created_at"`
}
```

### APIKey 扩展

```go
type APIKey struct {
    // ... existing fields
    UsedCount   int64  `gorm:"default:0" json:"used_count"`  // 新增: 调用次数
}
```

## 接口设计

### Transformer 接口变更

```go
type StreamResult struct {
    Usage         *Usage // 提取的 token 信息
    Error         error  // 流处理错误
}

type Transformer interface {
    TransformRequest(req *OpenAIRequest) (interface{}, error)
    TransformResponse(body []byte) (*OpenAIResponse, error)
    // 新签名: 返回提取的 usage
    TransformStream(reader io.Reader, writer io.Writer) *StreamResult
}
```

### API 接口扩展

**GET /usage/stats** - 增加模型维度

```json
{
  "totalRequests": 100,
  "totalTokens": 5000,
  "modelStats": [
    {
      "model": "gpt-4",           // 映射模型
      "actualModel": "gpt-4-turbo", // 实际模型
      "count": 50,
      "tokens": 2500
    }
  ]
}
```

**GET /usage/logs** - 增加字段

```json
{
  "logs": [
    {
      "id": 1,
      "model": "gpt-4",
      "actual_model": "gpt-4-turbo",
      "prompt_tokens": 100,
      "completion_tokens": 50
    }
  ]
}
```

**GET /api-keys** - 增加调用次数

```json
{
  "keys": [
    {
      "id": 1,
      "used_quota": 4591,
      "used_count": 42
    }
  ]
}
```

## 实现细节

### 1. PassThroughTransformer

```go
func (t *PassThroughTransformer) TransformStream(reader io.Reader, writer io.Writer) *StreamResult {
    result := &StreamResult{}
    scanner := bufio.NewScanner(reader)
    // 增加 buffer 大小以处理大响应
    scanner.Buffer(make([]byte, 64*1024), 1024*1024)
    
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Fprintln(writer, line)
        if f, ok := writer.(interface{ Flush() }); ok {
            f.Flush()
        }
        
        // 解析 SSE 数据
        if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                break
            }
            
            var chunk map[string]interface{}
            if err := json.Unmarshal([]byte(data), &chunk); err != nil {
                continue
            }
            
            // 提取 usage (最后一个带 usage 的 chunk 会覆盖前面的)
            if usage, ok := chunk["usage"].(map[string]interface{}); ok {
                result.Usage = &Usage{
                    PromptTokens:     int(usage["prompt_tokens"].(float64)),
                    CompletionTokens: int(usage["completion_tokens"].(float64)),
                }
            }
        }
    }
    
    result.Error = scanner.Err()
    return result
}
```

### 2. AnthropicTransformer

Anthropic 在 `message_start` 和 `message_delta` 事件中分别提供 token 信息：

```go
func (t *OpenAIToAnthropicTransformer) TransformStream(reader io.Reader, writer io.Writer) *StreamResult {
    result := &StreamResult{}
    // ... 现有逻辑
    
    // message_start 事件
    case "message_start":
        result.Usage.PromptTokens = msg.Message.Usage.InputTokens
    
    // message_delta 事件  
    case "message_delta":
        result.Usage.CompletionTokens = delta.Usage.OutputTokens
    
    return result
}
```

### 3. ProxyHandler 调整

```go
func (h *ProxyHandler) handleStreamResponse(c *gin.Context, resp *http.Response, trans transformer.Transformer, providerID uint, alias string, actualModel string, startTime time.Time) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    result := trans.TransformStream(resp.Body, c.Writer)
    
    usage := result.Usage
    if usage == nil {
        usage = &transformer.Usage{} // fallback
    }
    
    status := "success"
    if result.Error != nil {
        status = "error"
    }
    
    h.logUsage(c, providerID, alias, actualModel, usage.PromptTokens, usage.CompletionTokens, 
               int(time.Since(startTime).Milliseconds()), status, errorMsg)
}
```

### 4. 数据库迁移

```go
func Migrate(db *gorm.DB) error {
    // 添加新列
    if !db.Migrator().HasColumn(&UsageLog{}, "actual_model") {
        db.Exec("ALTER TABLE usage_logs ADD COLUMN actual_model VARCHAR(128)")
    }
    if !db.Migrator().HasColumn(&APIKey{}, "used_count") {
        db.Exec("ALTER TABLE keys ADD COLUMN used_count INTEGER DEFAULT 0")
    }
    
    // 修改 key_id 允许 NULL（管理员请求）
    // SQLite 不支持 ALTER COLUMN，需要重建表
    // 或者保持 not null，管理员请求不记录 usage
    
    return nil
}
```

## 测试计划

1. **单元测试**
   - Transformer 的 `TransformStream` 返回正确的 usage
   - 边界情况：无 usage 的流式响应

2. **集成测试**
   - 流式请求后 `UsageLog` 有正确的 token 值
   - 流式请求后 `APIKey.used_quota` 正确累加
   - 映射模型和实际模型都正确记录

3. **手动验证**
   - 调用流式 API，检查 usage_logs 表
   - 检查 API Keys 页面的 used_quota 和 used_count
