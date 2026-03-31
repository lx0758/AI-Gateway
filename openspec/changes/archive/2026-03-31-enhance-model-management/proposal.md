## Why

部分厂商（如自定义 API 端点、新厂商）未完整实现模型列表接口，无法通过自动同步获取模型信息。后端 API 已支持手动添加模型，但前端 UI 缺少相关功能。

**更重要的问题**: 当厂商改名模型时（如 gpt-4 → gpt-4-turbo），ModelMapping 仍指向旧的 ProviderModel 记录，导致路由失效。原因是 ModelMapping 用 ProviderModel.ID（数据库主键）关联，而非 model_id 字符串。

**效率问题**: 各列表页只能逐个删除，批量清理时效率低下。

## What Changes

- **BREAKING**: ModelMapping 改用 `provider_model_name` 字段存储模型 ID，替代 `provider_model_id` 外键
- **新增**: Provider 详情页支持手动添加/编辑/删除 ProviderModel 的 UI
- **新增**: 数据迁移脚本，将现有 `provider_model_id` 转换为 `provider_model_name`
- **新增**: 所有列表页支持多选批量删除功能
  - 厂商列表：批量删除厂商
  - 模型映射列表：批量删除映射
  - API 密钥列表：批量删除密钥
  - Provider 模型列表：批量删除模型
- **优化**: Provider 模型列表添加 Loading 状态，提升大量数据加载体验

## Capabilities

### New Capabilities

- `manual-model-ui`: Provider 详情页的手动模型管理界面
- `batch-delete`: 列表页多选批量删除功能

### Modified Capabilities

- `model-mapping`: **BREAKING** - 改用 model_id 字符串关联，替代数据库 ID 外键

## Impact

- **数据库**: ModelMapping 表结构变更，需要迁移
- **后端**: `handler/model_mapping.go` 查询逻辑修改
- **后端**: `router/model_router.go` 路由逻辑修改
- **前端**: 所有列表页添加多选和批量删除功能
  - `views/Providers/index.vue`
  - `views/Providers/Detail.vue`
  - `views/Models/index.vue`
  - `views/Keys/index.vue`
