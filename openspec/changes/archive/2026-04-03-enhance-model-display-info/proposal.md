## Why

当前存在两个主要问题：

1. **信息展示不直观**：模型列表和别名映射界面的信息展示不够清晰，用户难以快速了解模型的核心特性（token上下文限制、能力支持等），影响配置别名映射时的决策效率。

2. **界面风格不一致**：Aliases 页面使用折叠面板布局，与 Providers 页面的扁平表格风格不一致，用户体验差异明显。Mappings 嵌套在折叠面板中，空间受限，难以完整展示模型信息。

## What Changes

### 1. Aliases 主界面改版

- 将折叠面板布局改为扁平表格布局（与 Providers 界面风格一致）
- 新增批量删除功能
- 新增 Token 上下文汇总列（显示所有有效 mappings 的最小 token 组合）
- 新增能力特性列（显示所有有效 mappings 的能力交集）
- 表格列：选择框、别名名称、映射数量、能力特性、Token汇总、状态开关、操作按钮（编辑、详情、删除）
- Token 汇总增加鼠标悬停提示显示原始值

### 2. Aliases 详情页面新增

- 新增独立的详情页面（`/aliases/:id`），用于管理 mappings
- mappings 表格支持拖放排序，拖放后自动重新计算权重（位置从上到下权重递减）
- mappings 表格显示完整信息：提供商、API类型、模型名称、能力特性、上下文窗口、权重、状态、操作
- 能力特性列显示在上下文窗口列之前
- 上下文窗口增加鼠标悬停提示显示原始值
- 支持批量删除 mappings
- 显示别名名称和状态标签（移除编辑/删除按钮）

### 3. Providers Detail 页面增强

- 格式化显示 token 上下文信息（保留一位小数，如 "128K"、"153.6K"、"1.5M"）
- 显式展示能力特性标签（Vision、Tools、Stream）
- 能力特性列显示在上下文窗口列之前
- 上下文窗口增加鼠标悬停提示显示原始值

### 4. 后端 API 扩展

- GET `/aliases` 返回 alias 对象新增字段：
  - `min_context_window`、`min_max_output`（汇总统计最小值）
  - `supports_vision`、`supports_tools`、`supports_stream`（能力交集）
- GET `/aliases/:id` 返回 mappings 时包含模型详细信息（token、能力）
- 新增 PUT `/aliases/:id/mappings/order` API，用于拖放排序后的权重更新

## Capabilities

### New Capabilities

- `alias-detail-page`: 新增别名详情页面，用于管理单个 alias 的 mappings，支持拖放排序、批量操作、完整信息展示

### Modified Capabilities

- `manual-model-ui`: 需求变更 - 增强模型列表的显示要求，显式展示能力特性和格式化的 token 信息
- `alias-mapping`: 需求变更 - AliasMapping 应包含 ProviderModel 的详细信息（token、能力），支持拖放排序调整权重
- `model-alias`: 需求变更 - Alias 主列表应显示 token 汇总信息，界面风格改为扁平表格布局

## Impact

- **前端代码**：
  - `web/src/views/Aliases/index.vue`: 改版为扁平表格布局，增加批量删除、Token汇总列、能力交集显示、鼠标提示
  - `web/src/views/Aliases/Detail.vue`: **新增**详情页面，实现拖放排序、完整信息展示、别名显示
  - `web/src/views/Providers/Detail.vue`: 增强 token 和能力显示（格式化、鼠标提示）
  - `web/src/utils/format.ts`: 新增 token 格式化函数（保留一位小数）
  - `web/src/router/index.ts`: 新增 `/aliases/:id` 路由
  - `web/src/locales/*.ts`: 新增国际化文本
  
- **后端代码**：
  - `server/internal/handler/alias.go`: 
    - GET `/aliases` 新增 min_context_window、min_max_output、supports_vision、supports_tools、supports_stream 字段
    - GET `/aliases/:id` mappings 包含模型详细信息
    - 新增 PUT `/aliases/:id/mappings/order` 处理拖放排序
    - 新增 `calculateCapabilitiesIntersection` 函数计算能力交集
  - `server/internal/model/alias.go`: 新增汇总统计和能力交集计算方法
  
- **API 变更**：
  - GET `/aliases` 返回对象新增字段：
    - `min_context_window`、`min_max_output`（token 汇总）
    - `supports_vision`、`supports_tools`、`supports_stream`（能力交集）
  - GET `/aliases/:id` 返回的 mapping 对象新增字段：`model_info` 包含 `{context_window, max_output, supports_vision, supports_tools, supports_stream}`
  - 新增 PUT `/aliases/:id/mappings/order` 接口，请求体：`{ "order": [mapping_id_1, mapping_id_2, ...] }`
  
- **数据库**：
  - 无变更，使用现有的 Alias、AliasMapping、ProviderModel 表数据
  
- **依赖**：
  - 前端拖放功能使用 sortablejs 库
  - 鼠标提示使用 Element Plus 的 el-tooltip 组件