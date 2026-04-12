# MCP 协议端点

## ADDED Requirements

### Requirement: JSON-RPC 2.0 端点
系统 SHALL 在 `/mcp/v1` 提供 JSON-RPC 2.0 端点，接受 POST 请求和 JSON-RPC 2.0 格式消息。

#### Scenario: 有效 JSON-RPC 请求
- **WHEN** 客户端发送 POST 请求到 `/mcp/v1`，附带有效的 JSON-RPC 2.0 消息
- **THEN** 系统返回 JSON-RPC 2.0 响应，包含 result 或 error

#### Scenario: 无效 JSON-RPC 请求
- **WHEN** 客户端发送格式错误的 JSON-RPC 请求
- **THEN** 系统返回 JSON-RPC 错误响应，代码 -32700（解析错误）

### Requirement: API Key 认证
系统 SHALL 使用现有 API Key 系统认证 MCP 请求。客户端 MUST 通过 `Authorization: Bearer sk-xxx` Header 提供 API Key。

#### Scenario: 有效 API Key
- **WHEN** 客户端在 Authorization Header 中提供有效 API Key
- **THEN** 系统处理 MCP 请求

#### Scenario: 缺少 API Key
- **WHEN** 客户端未提供 API Key
- **THEN** 系统返回 HTTP 401 Unauthorized

#### Scenario: 无效 API Key
- **WHEN** 客户端提供无效或禁用的 API Key
- **THEN** 系统返回 HTTP 401 Unauthorized

### Requirement: Initialize 方法
系统 SHALL 实现 `initialize` 方法，基于 API Key 的权限返回服务器能力和可用资源。

#### Scenario: 带权限的 Initialize
- **WHEN** 客户端使用有效 API Key 调用 `initialize`
- **THEN** 系统返回能力对象，包含 API Key 有权访问的工具、资源和提示词

#### Scenario: 使用过期 API Key Initialize
- **WHEN** 客户端使用过期 API Key 调用 `initialize`
- **THEN** 系统返回 HTTP 401 Unauthorized

### Requirement: Tools List 方法
系统 SHALL 实现 `tools/list` 方法，返回 API Key 有权使用的所有工具，使用命名空间前缀名称。

#### Scenario: 列出可用工具
- **WHEN** 客户端使用有效 API Key 调用 `tools/list`
- **THEN** 系统返回工具数组，名称前缀为 `{symbol}.{tool_name}`

#### Scenario: 无权限列出工具
- **WHEN** 客户端使用无工具权限的 API Key 调用 `tools/list`
- **THEN** 系统返回空数组

### Requirement: Tools Call 方法
系统 SHALL 实现 `tools/call` 方法，将工具执行路由到适当的 MCP 服务。

#### Scenario: 带有效权限调用工具
- **WHEN** 客户端使用工具名称 `{symbol}.{tool_name}` 和有效参数调用 `tools/call`
- **THEN** 系统将请求路由到由 symbol 标识的 MCP 服务并返回结果

#### Scenario: 无权限调用工具
- **WHEN** 客户端调用 `tools/call`，使用无权限的工具名称
- **THEN** 系统返回 JSON-RPC 错误，代码 -32602（无效参数）

#### Scenario: 调用不存在工具
- **WHEN** 客户端调用 `tools/call`，使用不存在的工具名称
- **THEN** 系统返回 JSON-RPC 错误，代码 -32602（无效参数）

### Requirement: Resources List 方法
系统 SHALL 实现 `resources/list` 方法，返回 API Key 有权访问的所有资源。

#### Scenario: 列出可用资源
- **WHEN** 客户端使用有效 API Key 调用 `resources/list`
- **THEN** 系统返回资源数组，URI 前缀为 `mcp://{symbol}/{original_uri}`

### Requirement: Resources Read 方法
系统 SHALL 实现 `resources/read` 方法，从适当的 MCP 服务读取资源内容。

#### Scenario: 带有效权限读取资源
- **WHEN** 客户端使用 URI `mcp://{symbol}/{original_uri}` 调用 `resources/read`
- **THEN** 系统将请求路由到 MCP 服务并返回资源内容

### Requirement: Prompts List 方法
系统 SHALL 实现 `prompts/list` 方法，返回 API Key 有权使用的所有提示词。

#### Scenario: 列出可用提示词
- **WHEN** 客端使用有效 API Key 调用 `prompts/list`
- **THEN** 系统返回提示词数组，名称前缀为 `{symbol}.{prompt_name}`

### Requirement: Prompts Get 方法
系统 SHALL 实现 `prompts/get` 方法，从适当的 MCP 服务检索提示词模板。

#### Scenario: 带有效权限获取提示词
- **WHEN** 客户端使用提示词名称 `{symbol}.{prompt_name}` 调用 `prompts/get`
- **THEN** 系统将请求路由到 MCP 服务并返回提示词模板

### Requirement: SSE 传输支持
系统 SHALL 支持 Server-Sent Events (SSE) 传输，供偏好实时流式传输的客户端使用。

#### Scenario: SSE 连接请求
- **WHEN** 客户端发送 GET 请求到 `/mcp/v1`，附带 `Accept: text/event-stream` Header
- **THEN** 系统为双向通信建立 SSE 连接

#### Scenario: SSE 消息格式
- **WHEN** 系统通过 SSE 发送 JSON-RPC 消息
- **THEN** 消息格式为 `data: {json-rpc-message}\n\n`

### Requirement: 错误处理
系统 SHALL 为所有错误条件返回标准 JSON-RPC 2.0 错误代码。

#### Scenario: 服务不可用
- **WHEN** 工具调用期间 MCP 服务不可用
- **THEN** 系统返回 JSON-RPC 错误，代码 -32603（内部错误）和描述性消息

#### Scenario: 超时
- **WHEN** MCP 服务在超时期间未响应
- **THEN** 系统返回 JSON-RPC 错误，代码 -32603（内部错误）和超时消息