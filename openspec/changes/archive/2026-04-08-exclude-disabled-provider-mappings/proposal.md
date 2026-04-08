## Why

当 Provider 被禁用（`enabled=false`）时，其关联的 ModelMapping 仍然会在 API 响应中显示，导致用户看到不可用的映射，统计数据（MinContextWindow、MinMaxOutput、能力交集）也不准确。虽然路由层已正确过滤禁用的 Provider，但 API 层未同步此逻辑，造成数据展示与实际行为不一致。

## What Changes

- 在查询 ModelMapping 时，过滤掉 `Provider.enabled=false` 的记录
- 统计计算函数（`calculateMinTokens`、`calculateCapabilitiesIntersection`）也需要过滤禁用的 Provider
- 完全隐藏禁用 Provider 的映射，避免用户困惑

## Capabilities

### New Capabilities

无

### Modified Capabilities

- `model-mapping`: 增加 Provider.enabled 状态检查要求，确保 API 返回的映射仅包含启用的 Provider

## Impact

**受影响的文件：**
- `server/internal/handler/model.go`: List、Get、Update、ListMappings 函数
- `server/internal/handler/model.go`: calculateMinTokens、calculateCapabilitiesIntersection 函数
- `server/internal/handler/key.go`: ListModels 函数

**受影响的 API：**
- `GET /api/v1/models` - 模型列表
- `GET /api/v1/models/:id` - 模型详情
- `GET /api/v1/models/:id/mappings` - 模型映射列表
- `PUT /api/v1/models/:id` - 模型更新
- `GET /api/v1/keys/:id/models` - Key 的模型权限列表

**不受影响：**
- 路由层（`router.go`）已正确过滤，无需修改
