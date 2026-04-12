# MCP 资源同步

## ADDED Requirements

### Requirement: 资源发现
系统 SHALL 连接到 MCP 服务，使用 MCP 协议方法发现可用的工具、资源和提示词。

#### Scenario: 发现工具
- **WHEN** 系统与 MCP 服务同步
- **THEN** 系统调用 `tools/list` 方法并缓存所有工具定义

#### Scenario: 发现资源
- **WHEN** 系统与 MCP 服务同步
- **THEN** 系统调用 `resources/list` 方法并缓存所有资源定义

#### Scenario: 发现提示词
- **WHEN** 系统与 MCP 服务同步
- **THEN** 系统调用 `prompts/list` 方法并缓存所有提示词定义

### Requirement: 能力检测
系统 SHALL 在同步期间使用 `initialize` 方法检测 MCP 服务能力。

#### Scenario: 检测支持的能力
- **WHEN** 系统与 MCP 服务初始化连接
- **THEN** 系统存储服务能力（工具、资源、提示词支持标志）

#### Scenario: 能力不匹配
- **WHEN** 服务不支持某项能力（例如无资源）
- **THEN** 系统将该能力标记为不支持并跳过相关同步步骤

### Requirement: 工具缓存
系统 SHALL 在数据库中缓存工具定义，附带服务关联。

#### Scenario: 缓存新工具
- **WHEN** 同步发现数据库中不存在的新工具
- **THEN** 系统创建 MCPTool 记录，包含名称、描述和输入 Schema

#### Scenario: 更新现有工具
- **WHEN** 同步发现同名但 Schema 不同的工具
- **THEN** 系统更新现有的 MCPTool 记录

#### Scenario: 移除不可用的工具
- **WHEN** 同步发现数据库中有工具但服务响应中不存在
- **THEN** 系统将工具标记为不可用而非删除

### Requirement: 资源缓存
系统 SHALL 在数据库中缓存资源定义，附带服务关联。

#### Scenario: 缓存新资源
- **WHEN** 同步发现数据库中不存在的新资源
- **THEN** 系统创建 MCPResource 记录，包含 URI、名称、描述和 MIME 类型

#### Scenario: 更新现有资源
- **WHEN** 同步发现同 URI 但元数据不同的资源
- **THEN** 系统更新现有的 MCPResource 记录

### Requirement: 提示词缓存
系统 SHALL 在数据库中缓存提示词定义，附带服务关联。

#### Scenario: 缓存新提示词
- **WHEN** 同步发现数据库中不存在的新提示词
- **THEN** 系统创建 MCPPrompt 记录，包含名称、描述和参数 Schema

#### Scenario: 更新现有提示词
- **WHEN** 同步发现同名但参数不同的提示词
- **THEN** 系统更新现有的 MCPPrompt 记录

### Requirement: 同步元数据跟踪
系统 SHALL 为每个 MCP 服务跟踪最后同步时间。

#### Scenario: 更新同步时间戳
- **WHEN** 同步成功完成
- **THEN** 系统更新服务的 LastSyncAt 时间戳

#### Scenario: 同步失败
- **WHEN** 同步因连接错误失败
- **THEN** 系统记录错误并保留之前的 LastSyncAt 值

### Requirement: 手动同步触发
系统 SHALL 允许管理员手动触发单个服务的同步。

#### Scenario: 通过 API 手动同步
- **WHEN** 管理员调用服务的同步端点
- **THEN** 系统立即执行同步并返回结果

#### Scenario: 同步正在进行中
- **WHEN** 在前一次同步仍在运行时触发同步
- **THEN** 系统返回 HTTP 409 Conflict 或队列请求

### Requirement: 同步错误处理
系统 SHALL 优雅处理同步错误，不影响其他服务。

#### Scenario: 一个服务失败
- **WHEN** 一个 MCP 服务的同步失败
- **THEN** 系统记录错误但继续从其他服务提供缓存资源

#### Scenario: 所有服务失败
- **WHEN** 所有服务的同步失败
- **THEN** 系统继续提供先前缓存的资源