## Why

当前 KeyModel 表存储模型别名名称（字符串），当别名重命名时，API Key 的模型绑定关系无法自动跟随更新，导致数据不一致。改为存储 AliasID 后，通过数据库外键关联，别名重命名时 API Key 自动显示新名称，保证数据一致性。

## What Changes

**BREAKING 变更：**

- **数据库变更**：KeyModel 表字段 `Model string` 改为 `AliasID uint`，添加外键约束
- **API 变更**：
  - 创建/更新 API Key：`models` 参数从 `string[]` 改为 `uint[]`（AliasID）
  - 响应格式：keyModel 新增 `alias_id` 和 `alias_name` 字段
  - **删除 API**：移除 `POST /api-keys/:id/models` 和 `DELETE /api-keys/:id/models/:model_alias` 两个端点
- **前端变更**：模型选择改为使用 AliasID，回显和显示使用 `alias_id` 和 `alias_name`

## Capabilities

### New Capabilities
<!-- 无新增能力 -->

### Modified Capabilities
- `api-key-management`: API Key 创建/更新时使用 AliasID 绑定模型，查询响应包含 AliasID 和 Alias 名称

## Impact

- **代码变更**：
  - `server/internal/model/db.go`：KeyModel 结构修改
  - `server/internal/handler/key.go`：请求/响应格式变更，删除 AddModel/RemoveModel
  - `server/cmd/server/main.go`：删除两个路由
  - `web/src/views/Keys/index.vue`：前端选择器改用 AliasID
  
- **数据迁移**：用户手动处理现有数据（将名称转换为 AliasID）

- **兼容性**：
  - 前端需要同步更新（无法兼容旧版本）
  - API 调用方需要改用 AliasID
  - 数据库 schema 变化需要手动迁移