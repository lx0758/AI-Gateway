## Why

API Key 泄露或安全风险时，用户需要重新生成 Key 值而不删除整个 API Key 记录。当前系统只能删除并重新创建 API Key，这会丢失所有关联配置（名称、模型绑定、过期时间等），增加了管理负担和安全风险窗口。

## What Changes

- 新增 API Key 重置功能：生成新的 Key 值，保留原有配置
- 新增 REST API 端点：`POST /api-keys/:id/reset`
- 重置后返回新的 raw_key，旧的 Key 值立即失效
- 重置操作记录在审计日志中（如果系统有审计功能）

## Capabilities

### New Capabilities
- `api-key-reset`: API Key 重置功能，包括安全策略、权限控制和使用场景定义

### Modified Capabilities
<!-- 无现有规范需要修改 -->

## Impact

- **代码变更**：
  - `server/internal/handler/key.go`: 新增 `Reset` 方法
  - `server/cmd/server/main.go`: 新增路由 `POST /api-keys/:id/reset`
  - 可能需要更新前端 UI 以支持重置操作
  
- **API 变更**：
  - 新增端点：`POST /api-keys/:id/reset`
  - 返回格式：与创建 API Key 相同（包含 masked key 和 raw_key）

- **依赖影响**：
  - 无数据库 schema 变更（Key 模型已有 Key 字段）
  - 不影响现有认证中间件（使用最新的 Key 值进行验证）

- **安全考虑**：
  - 重置后的旧 Key 立即失效，所有使用旧 Key 的请求将返回 401
  - 需要确保用户有权访问新的 raw_key（仅在重置时返回一次）