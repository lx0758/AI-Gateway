## Why

当前 Key 页面的编辑弹窗混合了"基础信息编辑"和"权限详情管理"，交互不够清晰。用户需要在一个弹窗中切换4个TAB来管理模型、MCP工具、MCP资源、MCP提示词的权限，操作繁琐且容易混淆。

本次改造将"编辑"和"详情管理"分离，提供更清晰的交互层次：
- 列表只展示基础信息和权限概览（不限制/仅允许N个）
- 编辑弹窗只保留名称修改
- 新增独立的详情页，用单选框交互管理各类组件权限

## What Changes

- **列表页改造**：
  - 允许模型列改为显示"不限制"或"仅允许 N 个"
  - 新增 MCP工具/MCP资源/MCP提示词列，显示逻辑同上
  - 新增"详情"按钮，跳转到详情页
  - 编辑按钮改为只修改名称的弹窗

- **新增 Key 详情页**：
  - 参考 MCP 详情页结构
  - 基础信息卡片：名称、Key、状态、过期时间等
  - 4个TAB：Models、MCP工具、MCP资源、MCP提示词
  - 每个TAB用单选框（默认/仅允许）控制权限
  - 每个TAB提供"全部允许"按钮清空关联

- **后端接口改造**：
  - GET `/keys/:id/models` 返回全量模型列表 + selected 状态
  - GET `/keys/:id/mcp-tools` 返回全量工具列表 + selected 状态（过滤禁用组件）
  - 同理 MCP Resources 和 MCP Prompts
  - 新增 POST `/keys/:id/models/:model_id` 单项添加关联
  - 新增 DELETE `/keys/:id/models/:model_id` 单项删除关联
  - 新增 DELETE `/keys/:id/models` 清空所有模型关联
  - 同理其他3类组件的单项操作接口
  - GET `/keys` 增加 mcp_tools_count/mcp_resources_count/mcp_prompts_count 返回
  - 新增 GET `/keys/:id` 返回单个 Key 基础信息

- **路由配置**：
  - 新增 `/keys/:id` 路由指向详情页

## Capabilities

### New Capabilities

- `key-detail-page`: Key 详情页的前端交互设计，包括单选框权限管理、全部允许按钮等UI交互规范

### Modified Capabilities

- `api-key-management`: 
  - GET `/keys/:id/models` 接口改造：返回全量模型列表 + selected 状态
  - 新增 POST `/keys/:id/models/:model_id` 单项添加模型关联
  - 新增 DELETE `/keys/:id/models/:model_id` 单项删除模型关联
  - 新增 DELETE `/keys/:id/models` 清空所有模型关联
  - GET `/keys` 接口增加 mcp_tools_count/mcp_resources_count/mcp_prompts_count
  - 新增 GET `/keys/:id` 获取单个 Key 基础信息

- `api-key-mcp-resources`:
  - GET `/keys/:id/mcp-tools` 接口改造：返回全量工具列表（过滤禁用） + selected 状态
  - 新增 POST `/keys/:id/mcp-tools/:tool_id` 单项添加工具关联
  - 新增 DELETE `/keys/:id/mcp-tools/:tool_id` 单项删除工具关联
  - 新增 DELETE `/keys/:id/mcp-tools` 清空所有工具关联
  - 同理 MCP Resources 和 MCP Prompts 的接口改造

## Impact

- **前端代码**：
  - `web/src/views/Keys/index.vue` - 列表页改造
  - `web/src/views/Keys/Detail.vue` - 新增详情页
  - `web/src/router/index.ts` - 新增路由配置

- **后端代码**：
  - `server/internal/handler/key.go` - 接口改造和新增
  - `server/internal/router/router.go` - 路由注册

- **API 接口**：
  - 改造 2 个 GET 接口（models 和 mcp-tools）
  - 新增 8 个 POST/DELETE 单项操作接口
  - 新增 4 个 DELETE 清空接口
  - 改造 1 个 GET /keys 列表接口
  - 新增 1 个 GET /keys/:id 详情接口

- **数据库**：无变更，现有关联表已支持