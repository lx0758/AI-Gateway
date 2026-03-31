# Remove Key Quota and Limit Features

## Why

当前 Key 限额系统存在以下问题：

1. **限额功能从未生效** — `UsedQuota` 字段从未被更新，导致限额检查永远通过
2. **RateLimit 未实现** — 字段存在但从未使用
3. **维护成本高** — 对于小型项目，完整的限额系统过于复杂
4. **统计页面已覆盖** — 使用统计在 `/usage` 页面已有完整展示，Key 列表页面的统计信息冗余

对于当前规模的项目，简化的 Key 管理更合适。保留 `ExpiresAt` 过期时间和 `Enabled` 启用/禁用开关即可满足基本需求。

## What Changes

移除 Key 相关的限额和使用统计功能：

- **移除字段**: `RateLimit`, `Quota`, `UsedQuota`, `UsedCount`
- **移除检查逻辑**: auth 中间件中的 quota 超限检查
- **移除 API 参数**: 创建/更新 Key 时的限额相关参数
- **移除 Key 列表统计**: `TotalTokens`, `AvgLatency` 及相关查询

保留：
- `ExpiresAt` — Key 过期时间
- `Enabled` — Key 启用/禁用状态
- `UsageLog` — 使用日志（供统计页面使用）

## Capabilities

### New Capabilities

无

### Modified Capabilities

无（这是移除功能，不涉及 spec 级别的行为变更）

## Impact

**后端代码:**
- `server/internal/model/models.go` — Key 模型字段
- `server/internal/middleware/auth.go` — 移除 quota 检查
- `server/internal/handler/key.go` — 移除限额相关参数和统计

**数据库:**
- 需要迁移：删除 `keys` 表的 `rate_limit`, `quota`, `used_quota`, `used_count` 列

**API:**
- `POST /api/v1/api-keys` — 移除 `rate_limit`, `quota` 参数
- `PUT /api/v1/api-keys/:id` — 移除 `rate_limit`, `quota` 参数
- `GET /api/v1/api-keys` — 移除返回值中的 `quota`, `used_quota`, `used_count`, `total_tokens`, `avg_latency`

**前端:**
- Key 管理页面移除限额相关表单字段和统计显示
