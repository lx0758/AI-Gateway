## Project Positioning

**项目名称：** AI Proxy（AI 代理）

**定位：** 统一的 AI 服务代理平台，当前聚焦于 AI 模型 API 代理，未来将扩展支持：
- **MCP (Model Context Protocol)** 代理
- **ACP (Agent Communication Protocol)** 代理

## Why

系统存在三处实现缺陷和三个功能缺失：
1. 厂商编辑时前端未加载现有数据，导致无法修改创建时的信息
2. 模型映射创建后无法修改 Alias、Provider、Model 等关键字段
3. API Key 的模型权限控制功能未实现，且当前使用 JSON 存储不够灵活
4. 同步厂商模型时会覆盖手动创建的模型配置
5. 仪表盘数据始终为0，后端返回 snake_case 字段名但前端期望 camelCase
6. 刷新页面后跳转登录页，路由守卫在 fetchUser 执行前就检查了登录状态

## What Changes

- **FIX** 厂商编辑：前端编辑对话框加载并填充现有厂商数据
- **FIX** 仪表盘数据：统一后端 API 返回字段命名为 camelCase
- **FIX** 刷新登录：路由守卫中先尝试恢复用户状态再判断登录
- **FIX** API 类型显示：前端显示格式改为 `@ai-sdk/openai-compatible` 和 `@ai-sdk/anthropic`
- **NEW** 模型映射编辑：支持修改 Alias、Provider、ProviderModel
- **NEW** API Key 模型权限：使用关联表替代 JSON 存储，支持配置允许访问的模型
- **NEW** 手动模型保护：同步时跳过 `source=manual` 的模型，避免覆盖手动配置

## Capabilities

### New Capabilities

- `api-key-model-permission`: API Key 模型访问权限控制，使用关联表管理 Key 可访问的模型列表
- `manual-model-protection`: 手动创建模型的同步保护机制，同步时保留手动配置不被覆盖

### Modified Capabilities

- `model-mapping`: 新增修改映射功能，支持修改 Alias、Provider、ProviderModel
- `provider-management`: 修正前端实现，编辑时正确加载现有数据
- `dashboard`: 修正后端 API 返回字段命名，使用 camelCase 与前端保持一致
- `auth-session`: 修正路由守卫逻辑，刷新页面时正确恢复登录状态

## Impact

**前端变更：**
- `web/src/views/Providers/index.vue` - 编辑对话框加载现有数据，API 类型显示新格式
- `web/src/views/Models/index.vue` - 新增编辑对话框
- `web/src/views/APIKeys/index.vue` - 新增模型权限配置字段
- `web/src/router/index.ts` - 路由守卫先恢复用户状态再判断登录
- `web/src/App.vue` - 移除冗余的 fetchUser 调用
- `web/src/locales/*.ts` - 标题改为 "AI Proxy" / "AI 代理"

**后端变更：**
- `server/go.mod` - module 名称改为 `ai-proxy`
- `server/internal/model/models.go` - 新增 `APIKeyModel` 关联表
- `server/internal/handler/model_mapping.go` - 扩展 Update 支持更多字段
- `server/internal/handler/api_key.go` - 新增 Update API，支持模型权限
- `server/internal/handler/provider_model.go` - 同步逻辑跳过手动模型
- `server/internal/handler/usage.go` - Dashboard API 返回字段改用 camelCase
- `server/internal/middleware/auth.go` - session 名称改为 `ai-proxy-session`

**数据库变更：**
- 新增 `api_key_models` 关联表
