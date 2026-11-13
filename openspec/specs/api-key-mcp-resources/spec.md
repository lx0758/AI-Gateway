# api-key-mcp-resources 规格

## Purpose

定义 API Key 对 MCP 能力的细粒度权限控制：工具、资源、提示词三级权限配置、MCP 调用时的权限执行与响应过滤，及级联删除。

## Requirements

### Requirement: 工具权限配置

系统 SHALL 为每个 API Key 提供独立的 MCP 工具白名单开关：关闭时允许全部工具并清除关联记录，开启时仅允许勾选的工具。

#### Scenario: 授予工具访问权限
- **WHEN** 管理员授予 API Key 对 MCP 工具的访问权限
- **THEN** 系统创建 KeyMCPTool 关联

#### Scenario: 撤销工具访问权限
- **WHEN** 管理员移除 API Key 对 MCP 工具的访问权限
- **THEN** 系统删除 KeyMCPTool 关联

#### Scenario: 列出 Key 工具
- **WHEN** 管理员请求 API Key 的工具
- **THEN** 系统返回全部工具及勾选状态与白名单开关状态

#### Scenario: 开启工具白名单
- **WHEN** 管理员开启 API Key 的工具白名单开关
- **THEN** 系统弹出"将拒绝所有调用"的确认警示
- **AND** 开启后仅勾选的工具被允许

#### Scenario: 关闭工具白名单
- **WHEN** 管理员关闭 API Key 的工具白名单开关
- **THEN** 系统删除该 Key 的全部 KeyMCPTool 关联
- **AND** 该 Key 可以访问所有工具

#### Scenario: 白名单关闭时工具勾选开关显示为开启且禁用
- **WHEN** 管理员打开工具白名单开关关闭的 Key 的工具列表
- **THEN** 每个工具的勾选开关显示为开启状态且不可操作，提示"白名单已关闭，全部允许"

### Requirement: 资源权限配置

系统 SHALL 为每个 API Key 提供独立的 MCP 资源白名单开关：关闭时允许全部资源并清除关联记录，开启时仅允许勾选的资源。

#### Scenario: 授予资源访问权限
- **WHEN** 管理员授予 API Key 对 MCP 资源的访问权限
- **THEN** 系统创建 KeyMCPResource 关联

#### Scenario: 撤销资源访问权限
- **WHEN** 管理员移除 API Key 对 MCP 资源的访问权限
- **THEN** 系统删除 KeyMCPResource 关联

#### Scenario: 开启资源白名单
- **WHEN** 管理员开启 API Key 的资源白名单开关
- **THEN** 系统弹出"将拒绝所有调用"的确认警示
- **AND** 开启后仅勾选的资源被允许

#### Scenario: 关闭资源白名单
- **WHEN** 管理员关闭 API Key 的资源白名单开关
- **THEN** 系统删除该 Key 的全部 KeyMCPResource 关联
- **AND** 该 Key 可以访问所有资源

#### Scenario: 白名单关闭时资源勾选开关显示为开启且禁用
- **WHEN** 管理员打开资源白名单开关关闭的 Key 的资源列表
- **THEN** 每个资源的勾选开关显示为开启状态且不可操作，提示"白名单已关闭，全部允许"

### Requirement: 提示词权限配置

系统 SHALL 为每个 API Key 提供独立的 MCP 提示词白名单开关：关闭时允许全部提示词并清除关联记录，开启时仅允许勾选的提示词。

#### Scenario: 授予提示词访问权限
- **WHEN** 管理员授予 API Key 对 MCP 提示词的访问权限
- **THEN** 系统创建 KeyMCPPrompt 关联

#### Scenario: 撤销提示词访问权限
- **WHEN** 管理员移除 API Key 对 MCP 提示词的访问权限
- **THEN** 系统删除 KeyMCPPrompt 关联

#### Scenario: 开启提示词白名单
- **WHEN** 管理员开启 API Key 的提示词白名单开关
- **THEN** 系统弹出"将拒绝所有调用"的确认警示
- **AND** 开启后仅勾选的提示词被允许

#### Scenario: 关闭提示词白名单
- **WHEN** 管理员关闭 API Key 的提示词白名单开关
- **THEN** 系统删除该 Key 的全部 KeyMCPPrompt 关联
- **AND** 该 Key 可以访问所有提示词

#### Scenario: 白名单关闭时提示词勾选开关显示为开启且禁用
- **WHEN** 管理员打开提示词白名单开关关闭的 Key 的提示词列表
- **THEN** 每个提示词的勾选开关显示为开启状态且不可操作，提示"白名单已关闭，全部允许"

### Requirement: RESTful API 设计

系统 SHALL 提供一致的 RESTful 端点用于配置 MCP 资源白名单。

#### Scenario: GET Key 的 MCP 工具
- **WHEN** 管理员调用 GET /keys/:id/mcp-tools
- **THEN** 系统返回该 Key 的工具列表及勾选状态

#### Scenario: PUT Key 的 MCP 工具
- **WHEN** 管理员调用 PUT /keys/:id/mcp-tools，附带工具 ID 数组
- **THEN** 系统用新列表替换 Key 的工具勾选记录

#### Scenario: GET Key 的 MCP 资源
- **WHEN** 管理员调用 GET /keys/:id/mcp-resources
- **THEN** 系统返回该 Key 的资源列表及勾选状态

#### Scenario: PUT Key 的 MCP 资源
- **WHEN** 管理员调用 PUT /keys/:id/mcp-resources，附带资源 ID 数组
- **THEN** 系统用新列表替换 Key 的资源勾选记录

#### Scenario: GET Key 的 MCP 提示词
- **WHEN** 管理员调用 GET /keys/:id/mcp-prompts
- **THEN** 系统返回该 Key 的提示词列表及勾选状态

#### Scenario: PUT Key 的 MCP 提示词
- **WHEN** 管理员调用 PUT /keys/:id/mcp-prompts，附带提示词 ID 数组
- **THEN** 系统用新列表替换 Key 的提示词勾选记录

#### Scenario: 切换白名单开关
- **WHEN** 管理员调用 PUT /keys/:id/whitelist/tools、/whitelist/resources 或 /whitelist/prompts 且 enabled=false
- **THEN** 对应开关关闭且该 Key 该类型的全部关联记录被清除
- **WHEN** 管理员调用同一端点且 enabled=true
- **THEN** 对应开关开启且关联记录保持为空，等待管理员勾选

#### Scenario: 勾选与取消勾选单项
- **WHEN** 管理员调用 POST 或 DELETE /keys/:id/mcp-tools/:tool_id（资源、提示词同构）
- **THEN** 系统创建或删除对应的关联记录

### Requirement: 权限执行

系统 SHALL 在客户端调用 MCP 端点时按对应类型的白名单开关执行资源权限。

#### Scenario: 访问已授权工具
- **WHEN** 客户端调用有权限的工具
- **THEN** 系统将请求路由到 MCP 服务

#### Scenario: 访问未授权工具
- **WHEN** 客户端调用无权限的工具
- **THEN** 系统返回 JSON-RPC 错误，代码 -32602

#### Scenario: 访问已授权资源
- **WHEN** 客户端读取有权限的资源
- **THEN** 系统将请求路由到 MCP 服务

#### Scenario: 访问未授权资源
- **WHEN** 客户端读取无权限的资源
- **THEN** 系统返回 JSON-RPC 错误，代码 -32602

#### Scenario: 白名单关闭时全部放行
- **WHEN** 客户端通过某类型白名单开关为关闭的 Key 发起调用
- **THEN** 系统将请求路由到 MCP 服务

#### Scenario: 白名单开启时按勾选状态执行
- **WHEN** 客户端调用白名单开启的 Key 下未勾选的工具、资源或提示词
- **THEN** 系统返回 JSON-RPC 错误，代码 -32602

### Requirement: Initialize 响应过滤

系统 SHALL 基于 API Key 白名单开关与勾选状态过滤 `initialize` 响应中的资源。

#### Scenario: 过滤的工具列表
- **WHEN** 客户端使用工具白名单开启的 API Key 调用 initialize
- **THEN** 系统仅返回该 Key 已勾选的工具，带命名空间前缀

#### Scenario: 过滤的资源列表
- **WHEN** 客户端使用资源白名单开启的 API Key 调用 initialize
- **THEN** 系统仅返回该 Key 已勾选的资源，带修改的 URI

#### Scenario: 过滤的提示词列表
- **WHEN** 客户端使用提示词白名单开启的 API Key 调用 initialize
- **THEN** 系统仅返回该 Key 已勾选的提示词，带命名空间前缀

#### Scenario: 白名单关闭时不过滤
- **WHEN** 客户端使用某类型白名单开关为关闭的 API Key 调用 initialize
- **THEN** 系统返回该类型的全部工具、资源或提示词（仍带命名空间前缀）

### Requirement: 命名空间前缀添加
系统 SHALL 向所有返回给客户端的资源标识符添加命名空间前缀。

#### Scenario: 工具名称前缀
- **WHEN** 系统在 initialize 响应中返回工具
- **THEN** 工具名称前缀为 `{symbol}.{original_name}`

#### Scenario: 资源 URI 前缀
- **WHEN** 系统在 initialize 响应中返回资源
- **THEN** 资源 URI 前缀为 `mcp://{symbol}/{original_uri}`

#### Scenario: 提示词名称前缀
- **WHEN** 系统在 initialize 响应中返回提示词
- **THEN** 提示词名称前缀为 `{symbol}.{original_name}`

### Requirement: 级联删除
系统 SHALL 在资源被删除时处理权限清理。

#### Scenario: 工具从服务中删除
- **WHEN** MCP 工具从数据库中移除（服务删除）
- **THEN** 系统删除该工具的所有 KeyMCPTool 关联

#### Scenario: API Key 被删除
- **WHEN** API Key 被删除
- **THEN** 系统删除该 Key 的所有 KeyMCPTool、KeyMCPResource、KeyMCPPrompt 关联
