## Why

当前管理后台的列表展示缺乏合理的排序和分组，用户查找特定厂商或模型时需要手动浏览。模型映射表结构设计为平铺单表，同一 alias 在多条记录中重复，管理时需要逐条编辑，无法整体查看同一别名下的所有厂商映射。用量日志的错误信息无法便捷复制，排查问题效率低。这些设计和体验问题降低了管理效率。

## What Changes

- **BREAKING**: 数据模型重构，删除 `ModelMapping` 单表，新增 `Alias` + `AliasMapping` 一对多结构
- 厂商列表按名称排序（前端）
- 厂商详情页模型列表按模型名称排序（前端）
- 模型别名管理页面重构：按 alias 分组展示，内嵌映射列表，支持折叠/展开
- 用量日志错误信息列增加复制按钮

## Capabilities

### New Capabilities

- `alias`: 模型别名管理能力，支持创建/编辑/删除别名，别名作为用户调用 API 时使用的模型名称标识
- `alias-mapping`: 别名映射管理能力，支持为别名配置多个厂商模型映射，实现负载均衡和故障转移

### Modified Capabilities

无。原 `model-mapping` 能力已被 `alias` + `alias-mapping` 完全替代。

## Impact

- **后端数据模型**: 新增 `aliases` 和 `alias_mappings` 表，删除 `model_mappings` 表
- **后端 API**: 新增 `/aliases` 和 `/aliases/:id/mappings` 嵌套资源接口，删除 `/model-mappings` 接口
- **前端**: 新增 `views/Aliases/index.vue` 别名管理页面，删除 `views/Models/`
- **代理逻辑**: 更新模型查找逻辑，从 `aliases` + `alias_mappings` 表获取映射配置
- **数据**: 不迁移旧数据，新表从空状态开始