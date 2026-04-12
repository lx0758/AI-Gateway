# MCP 服务管理

## ADDED Requirements

### Requirement: 创建 MCP 服务
系统 SHALL 允许管理员创建 MCP 服务配置，包含名称、标识符、类型（远程/本地）和连接参数。

#### Scenario: 创建远程服务
- **WHEN** 管理员创建类型为 "remote" 的 MCP 服务，包含 URL 和可选的自定义 Headers
- **THEN** 系统将服务配置存储到数据库

#### Scenario: 创建本地服务
- **WHEN** 管理员创建类型为 "local" 的 MCP 服务，包含命令字符串
- **THEN** 系统将服务配置存储到数据库

#### Scenario: 标识符重复
- **WHEN** 管理员创建 MCP 服务时使用的标识符已存在
- **THEN** 系统返回 HTTP 400 错误，消息为 "标识符已存在"

### Requirement: 标识符验证
系统 SHALL 验证服务标识符仅包含允许的字符 `[0-9a-zA-Z_-]`，且长度在 2-200 字符之间。

#### Scenario: 有效标识符
- **WHEN** 管理员创建标识符为 "filesystem" 的服务
- **THEN** 系统接受该标识符

#### Scenario: 无效的标识符字符
- **WHEN** 管理员创建标识符为 "file@system" 的服务
- **THEN** 系统返回 HTTP 400 错误，附带验证消息

#### Scenario: 标识符太短
- **WHEN** 管理员创建标识符为 "a" 的服务
- **THEN** 系统返回 HTTP 400 错误，附带验证消息

### Requirement: 列出 MCP 服务
系统 SHALL 提供端点列出所有配置的 MCP 服务及其元数据。

#### Scenario: 列出所有服务
- **WHEN** 管理员请求服务列表
- **THEN** 系统返回所有服务，包含名称、标识符、类型、启用状态和最后同步时间

### Requirement: 更新 MCP 服务
系统 SHALL 允许管理员更新 MCP 服务配置。

#### Scenario: 更新服务配置
- **WHEN** 管理员更新服务的 URL 或命令
- **THEN** 系统存储更新后的配置

#### Scenario: 更新服务标识符
- **WHEN** 管理员将服务标识符更新为新的唯一值
- **THEN** 系统更新标识符，所有关联资源反映新标识符

### Requirement: 删除 MCP 服务
系统 SHALL 允许管理员删除 MCP 服务配置。

#### Scenario: 删除服务
- **WHEN** 管理员删除 MCP 服务
- **THEN** 系统移除服务及所有缓存的工具/资源/提示词

#### Scenario: 删除带有权限的服务
- **WHEN** 管理员删除配置了 API Key 权限的 MCP 服务
- **THEN** 系统移除服务、缓存资源和所有权限关联

### Requirement: 测试 MCP 服务连接
系统 SHALL 提供测试端点以验证 MCP 服务连接性。

#### Scenario: 测试远程服务成功
- **WHEN** 管理员测试可访问的远程 MCP 服务
- **THEN** 系统返回成功，附带服务能力

#### Scenario: 测试本地服务成功
- **WHEN** 管理员测试成功执行的本地 MCP 服务命令
- **THEN** 系统返回成功，附带服务能力

#### Scenario: 测试服务失败
- **WHEN** 管理员测试不可访问或失败的 MCP 服务
- **THEN** 系统返回错误，附带失败原因

### Requirement: 同步 MCP 服务资源
系统 SHALL 提供同步端点，从 MCP 服务获取并缓存工具、资源和提示词。

#### Scenario: 同步服务成功
- **WHEN** 管理员触发 MCP 服务的同步
- **THEN** 系统连接到服务，获取所有资源，更新数据库缓存

#### Scenario: 同步检测到新工具
- **WHEN** 同步发现数据库中不存在的新工具
- **THEN** 系统创建新的 MCPTool 记录

#### Scenario: 同步检测到移除的工具
- **WHEN** 同步发现数据库中有工具但服务中已不存在
- **THEN** 系统将这些工具标记为不可用（软删除）

### Requirement: 服务启用/禁用
系统 SHALL 允许管理员启用或禁用 MCP 服务，而不删除它们。

#### Scenario: 禁用服务
- **WHEN** 管理员禁用 MCP 服务
- **THEN** 系统将服务标记为禁用，它不包含在客户端资源列表中

#### Scenario: 启用服务
- **WHEN** 管理员启用先前禁用的 MCP 服务
- **THEN** 系统将服务标记为启用，它包含在客户端资源列表中