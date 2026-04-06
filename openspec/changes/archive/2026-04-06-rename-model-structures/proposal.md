## Why

当前系统中 "alias"（别名）的命名不够直观，用户难以理解其代表"模型"的概念。为了提高系统的可读性和用户体验，需要将这些结构重命名为更清晰的名称。

## What Changes

**BREAKING** 数据库表重命名：
- `aliases` → `models`（模型）
- `alias_mappings` → `model_mappings`（模型映射）
- `usage_logs` → `model_logs`（模型日志）

数据库列名更新：
- `model_mappings.alias_id` → `model_mappings.model_id`
- `key_models.alias_id` → `key_models.model_id`（间接依赖）

**BREAKING** 后端代码重命名：
- `Alias` → `Model`（PO 结构体）
- `AliasMapping` → `ModelMapping`（PO 结构体）
- `UsageLog` → `ModelLog`（PO 结构体）
- `handler/alias.go` → `handler/model.go`
- 所有相关 API 路径和函数名

**BREAKING** API 路径简化：
- `/api-keys` → `/keys`（API Key 路径简化）

冗余代码清理：
- 删除未使用的 `Mapping` PO（server/internal/model/db.go）

**BREAKING** 前端代码重命名：
- `views/Aliases/` → `views/Models/`
- 路由路径 `/aliases` → `/models`
- 所有相关组件、变量和函数名

国际化：
- 更新中英文翻译文件中的相关文本

## Capabilities

### New Capabilities

无需新增能力，本次变更为纯重构。

### Modified Capabilities

- `model-alias`: 重命名为 `model-management`，更新所有需求描述中的术语
- `alias-mapping`: 重命名为 `model-mapping`，更新所有需求描述中的术语
- `usage-tracking`: 更新日志记录相关需求，`usage_logs` → `model_logs`

## Impact

**数据库影响**：
- 需要执行数据迁移脚本重命名表
- 外键约束需要更新
- 列名需要更新（alias_id → model_id）

**后端影响**：
- `server/internal/model/db.go` - PO 定义（含删除 Mapping PO）
- `server/internal/handler/alias.go` → `model.go`
- `server/internal/handler/usage.go` - 日志记录
- `server/internal/handler/key.go` - KeyModel 关联（间接依赖）
- `server/internal/handler/proxy*.go` - 代理处理
- `server/internal/router/router.go` - 路由定义
- 所有引用这些结构的代码

**前端影响**：
- `web/src/views/Aliases/` → `web/src/views/Models/`
- `web/src/router/index.ts` - 路由配置
- `web/src/locales/zh.ts` - 中文翻译
- `web/src/locales/en.ts` - 英文翻译
- `web/src/components/layout/MainLayout.vue` - 导航菜单

**文档影响**：
- `openspec/specs/model-alias/` → `openspec/specs/model-management/`
- `openspec/specs/alias-mapping/` → `openspec/specs/model-mapping/`
- `openspec/specs/usage-tracking/` - 更新表名引用
