# API Key MCP 资源

## ADDED Requirements

### Requirement: 工具权限配置
系统 SHALL 允许管理员配置每个 API Key 可以访问哪些 MCP 工具。

#### Scenario: 授予工具访问权限
- **WHEN** 管理员授予 API Key 对 MCP 工具的访问权限
- **THEN** 系统创建 KeyMCPTool 关联

#### Scenario: 撤销工具访问权限
- **WHEN** 管理员移除 API Key 对 MCP 工具的访问权限
- **THEN** 系统删除 KeyMCPTool 关联

#### Scenario: 列出 Key 工具
- **WHEN** 管理员请求 API Key 的工具
- **THEN** 系统返回该 Key 有权访问的所有工具

### Requirement: 资源权限配置
系统 SHALL 允许管理员配置每个 API Key 可以访问哪些 MCP 资源。

#### Scenario: 授予资源访问权限
- **WHEN** 管理员授予 API Key 对 MCP 资源的访问权限
- **THEN** 系统创建 KeyMCPResource 关联

#### Scenario: 撤销资源访问权限
- **WHEN** 管理员移除 API Key 对 MCP 资源的访问权限
- **THEN** 系统删除 KeyMCPResource 关联

### Requirement: 提示词权限配置
系统 SHALL 允许管理员配置每个 API Key 可以访问哪些 MCP 提示词。

#### Scenario: 授予提示词访问权限
- **WHEN** 管理员授予 API Key 对 MCP 提示词的访问权限
- **THEN** 系统创建 KeyMCPPrompt 关联

#### Scenario: 撤销提示词访问权限
- **WHEN** 管理员移除 API Key 对 MCP 提示词的访问权限
- **THEN** 系统删除 KeyMCPPrompt 关联

### Requirement: RESTful API 设计
系统 SHALL 提供一致的 RESTful 端点用于配置 MCP 资源权限。

#### Scenario: GET Key 的 MCP 工具
- **WHEN** 管理员调用 GET /api/v1/api-keys/:id/mcp-tools
- **THEN** 系统返回该 Key 有权访问的工具 ID 列表

#### Scenario: PUT Key 的 MCP 工具
- **WHEN** 管理员调用 PUT /api/v1/api-keys/:id/mcp-tools，附带工具 ID 数组
- **THEN** 系统用新列表替换 Key 的工具权限

#### Scenario: GET Key 的 MCP 资源
- **WHEN** 管理员调用 GET /api/v1/api-keys/:id/mcp-resources
- **THEN** 系统返回该 Key 有权访问的资源 ID 列表

#### Scenario: PUT Key 的 MCP 资源
- **WHEN** 管理员调用 PUT /api/v1/api-keys/:id/mcp-resources，附带资源 ID 数组
- **THEN** 系统用新列表替换 Key 的资源权限

#### Scenario: GET Key 的 MCP 提示词
- **WHEN** 管理员调用 GET /api/v1/api-keys/:id/mcp-prompts
- **THEN** 系统返回该 Key 有权访问的提示词 ID 列表

#### Scenario: PUT Key 的 MCP 提示词
- **WHEN** 管理员调用 PUT /api/v1/api-keys/:id/mcp-prompts，附带提示词 ID 数组
- **THEN** 系统用新列表替换 Key 的提示词权限

### Requirement: 权限执行
系统 SHALL 在客户端调用 MCP 端点时执行 MCP 资源权限。

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

### Requirement: Initialize 响应过滤
系统 SHALL 基于 API Key 权限过滤 `initialize` 响应中的资源。

#### Scenario: 过滤的工具列表
- **WHEN** 客户端使用 API Key 调用 initialize
- **THEN** 系统仅返回 Key 有权限的工具，带命名空间前缀

#### Scenario: 过滤的资源列表
- **WHEN** 客户端使用 API Key 调用 initialize
- **THEN** 系统仅返回 Key 有权限的资源，带修改的 URI

#### Scenario: 过滤的提示词列表
- **WHEN** 客户端使用 API Key 调用 initialize
- **THEN** 系统仅返回 Key 有权限的提示词，带命名空间前缀

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