## Why

当前 `ModelMapping` 使用字符串字段 `ProviderModelName` 匹配 `ProviderModel.ModelID`，存在两个问题：

1. **无法修改厂商模型 ID**：修改 `ProviderModel.ModelID` 会导致映射"断裂"，路由查询失败
2. **删除厂商模型产生孤儿数据**：删除 `ProviderModel` 不会级联删除相关的 `ModelMapping`，导致路由静默失效

将字符串匹配改为外键关联，解决数据一致性问题。

## What Changes

- `ModelMapping.ProviderModelName` (string) → `ProviderModelID` (uint, 外键)
- `ProviderModel.ModelID` 字段可编辑
- 删除 `ProviderModel` 时级联删除相关 `ModelMapping`
- **BREAKING** API 请求体：`provider_model_name` → `provider_model_id`

## Capabilities

### New Capabilities

- `model-mapping-reference`: ModelMapping 通过外键关联 ProviderModel，确保数据一致性

### Modified Capabilities

- `provider-model-management`: ProviderModel 的 ModelID 字段可编辑

## Impact

### 数据库
- `model_mappings` 表结构变更（字段迁移）
- 需要数据迁移脚本

### API
- `POST /models/:id/mappings`: `provider_model_name` → `provider_model_id`
- `PUT /models/:id/mappings/:mid`: `provider_model_name` → `provider_model_id`
- `PUT /providers/:id/models/:mid`: 允许修改 `model_id`

### 代码
- `internal/model/db.go`: ModelMapping 结构体
- `internal/handler/model.go`: Mapping 相关 CRUD 逻辑
- `internal/handler/provider_model.go`: Update 方法
- `internal/router/router.go`: 路由查询逻辑
- `internal/handler/model_testing.go`: 测试逻辑
