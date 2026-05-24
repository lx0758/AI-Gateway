## Why

当前系统删除策略不统一，部分实体使用软删除但有无效的级联删除约束，部分关联表本应硬删除却有软删除字段。需要统一删除策略，确保数据一致性和级联删除正确执行。

## What Changes

- **BREAKING** 移除 `ProviderModel` 的软删除，改为硬删除
- **BREAKING** 移除 `ModelMapping` 的软删除，改为硬删除
- **BREAKING** 移除 `MCPTool`, `MCPResource`, `MCPPrompt` 的软删除，改为硬删除
- 移除无效的 `OnDelete:CASCADE` 约束（软删除不触发数据库级联）
- 在代码层面实现级联删除逻辑

## Capabilities

### New Capabilities

- `delete-strategy`: 定义系统删除策略，区分软删除实体和硬删除实体

### Modified Capabilities

无（这是内部实现变更，不影响外部 API 行为）

## Impact

### 数据模型

| 模型 | 变更 |
|------|------|
| ProviderModel | 移除 DeletedAt 字段 |
| ModelMapping | 移除 DeletedAt 字段 |
| MCPTool | 移除 DeletedAt 字段 |
| MCPResource | 移除 DeletedAt 字段 |
| MCPPrompt | 移除 DeletedAt 字段 |

### 级联删除逻辑

| 删除操作 | 级联删除 |
|----------|----------|
| 删除 Provider | 硬删除 ProviderModel → 硬删除 ModelMapping |
| 删除 ProviderModel | 硬删除 ModelMapping |
| 删除 Model | 硬删除 ModelMapping |
| 删除 MCP | 硬删除 MCPTool, MCPResource, MCPPrompt |
| 删除 Key | 硬删除 KeyModel, KeyMCPTool, KeyMCPResource, KeyMCPPrompt |

### 代码文件

- `internal/model/db.go`: 移除 DeletedAt 字段
- `internal/handler/provider.go`: 删除时级联删除 ProviderModel 和 ModelMapping
- `internal/handler/provider_model.go`: 删除时级联删除 ModelMapping
- `internal/handler/model.go`: 删除时级联删除 ModelMapping
- `internal/handler/key.go`: 删除时级联删除关联表
