# 提案: 修复使用量统计问题

## 概述

修复三个相关的统计问题：
1. 流式 API Token 统计始终为 0
2. 缺少模型调用次数统计（映射模型 vs 实际模型）
3. 密钥 `used_quota` 统计错误（流式请求导致的连锁问题）

## 问题分析

### 问题 1: 流式 Token 统计为 0

**现状**: `handleStreamResponse` 中直接透传响应，token 始终记录为 0。

**影响**:
- Usage 统计不准确
- API Key 的 `used_quota` 不更新
- 无法正确计费

### 问题 2: 模型调用次数统计

**现状**: `UsageLog.Model` 只记录一个值，原始别名被实际模型名覆盖。

**需求**: 需要同时知道用户请求的模型别名和实际调用的上游模型。

### 问题 3: 密钥 used_quota 错误

**根因**: 问题 1 的连锁影响。流式请求 token=0，`used_quota` 累加无效。

**现象**: Key 2 的 used_quota=0，但数据库中有该 Key 的调用记录。

## 解决方案

### 方案 1: 流式 Token 统计

使用 OpenAI 的 `stream_options.include_usage` 参数强制上游返回 usage 信息。

**✅ 已验证通过** - OpenRouter 完美支持此特性：

```
请求: {"stream": true, "stream_options": {"include_usage": true}}

响应 (最后一个 chunk):
data: {"choices":[], "usage":{
  "prompt_tokens": 18,
  "completion_tokens": 20,
  "total_tokens": 38,
  "cost": 0,
  "prompt_tokens_details": {...},
  "completion_tokens_details": {"reasoning_tokens": 16, ...}
}}
data: [DONE]
```

**实现要点**:
```go
type OpenAIRequest struct {
    // ...
    StreamOptions *StreamOptions `json:"stream_options,omitempty"`
}

type StreamOptions struct {
    IncludeUsage bool `json:"include_usage,omitempty"`
}
```

在请求时自动注入：
```go
if req.Stream {
    req.StreamOptions = &StreamOptions{IncludeUsage: true}
}
```

解析流式响应中最后一个带 `usage` 字段的 chunk 并记录。

### 方案 2: 模型调用次数

扩展 `UsageLog` 结构：

```go
type UsageLog struct {
    // ... existing fields
    Model       string  // 用户请求的模型别名 (保持)
    ActualModel string  // 新增: 实际调用的上游模型 ID
}
```

在 `proxy.go` 中：
```go
alias := req.Model           // 保存原始别名
req.Model = result.ActualModel
// ...
h.logUsage(c, providerID, alias, result.ActualModel, ...)
```

### 方案 3: 密钥统计修复

问题 1 修复后自动解决。同时增加调用次数统计：

```go
// APIKey 结构增加
type APIKey struct {
    // ...
    UsedQuota   int64  // 已用 token (现有)
    UsedCount   int64  // 新增: 调用次数
}
```

## 范围

**包含**:
- 后端流式 token 统计修复
- 数据库 schema 变更 (新增字段)
- API 接口扩展 (返回映射模型和实际模型)
- 前端 Usage 页面增强

**不包含**:
- 历史数据迁移 (新字段允许为空)
- 其他统计功能

## 风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 数据库迁移 | 低 | 新字段允许 NULL，不影响现有数据 |
| 上游不支持 include_usage | ✅ 已验证 | OpenRouter/OpenAI 支持，Anthropic 有独立方案 |
| 前端兼容性 | 低 | 新字段为可选，旧 API 继续工作 |

## 预估工作量

- 后端修改: 2-3 小时
- 数据库迁移: 30 分钟
- 前端更新: 1-2 小时
- 测试验证: 1 小时

**总计**: 约 5-6 小时
