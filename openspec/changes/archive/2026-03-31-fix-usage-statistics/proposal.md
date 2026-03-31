## Why

Provider 重构后，统计数据未写入 `usage_logs` 表，导致所有统计功能失效。需要修复数据写入逻辑，并优化统计维度和前端展示。

## What Changes

- **修复数据写入**: 在 `proxy_openai.go` 和 `proxy_anthropic.go` 中添加 UsageLog 写入逻辑
- **简化字段结构**: 移除 `PromptTokens`/`CompletionTokens`，新增 `TotalTokens` (int64)，统一使用 int64 类型
- **新增统计维度**: 
  - 厂商统计（调用次数、Tokens、平均耗时）
  - Key 统计（调用次数、Tokens、平均耗时）
- **新增耗时统计**: 各维度增加 `LatencyMs` 统计（AVG/MAX/MIN）
- **前端优化**: Dashboard 新增厂商统计表，Usage 页面新增 Key 统计表，API Key 页面增加统计列

## Capabilities

### New Capabilities

- `usage-tracking`: 使用量追踪能力，记录每次 API 调用的详细信息（厂商、模型、Key、Tokens、耗时、状态）

### Modified Capabilities

- 无（这是新增能力，不修改现有 spec）

## Impact

**后端**:
- `server/internal/model/models.go` - UsageLog 结构调整
- `server/internal/handler/proxy_openai.go` - 数据写入
- `server/internal/handler/proxy_anthropic.go` - 数据写入
- `server/internal/handler/usage.go` - 统计查询优化

**前端**:
- `web/src/views/Dashboard/index.vue` - 新增统计卡片和厂商统计表
- `web/src/views/Usage/index.vue` - 新增 Key 统计表，调整字段
- `web/src/views/Keys/index.vue` - 新增统计列

**数据库**:
- `usage_logs` 表结构变更（移除字段、新增字段、类型变更）
