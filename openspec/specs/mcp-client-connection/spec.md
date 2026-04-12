# MCP 客户端连接

## ADDED Requirements

### Requirement: 按需连接策略
系统 SHALL 仅在需要时（工具调用、资源读取、提示词获取或同步）与 MCP 服务建立连接，而非维护持久连接。

#### Scenario: 第一次工具调用
- **WHEN** 客户端首次调用 MCP 服务的工具
- **THEN** 系统建立连接，执行工具，并保持连接到空闲超时时间

#### Scenario: 超时内的后续调用
- **WHEN** 客户端在空闲超时内再次调用同一 MCP 服务
- **THEN** 系统复用现有连接

#### Scenario: 超时后的调用
- **WHEN** 客户端在空闲超时过期后调用工具
- **THEN** 系统建立新连接

### Requirement: 远程 MCP 客户端（HTTP/SSE）
系统 SHALL 实现 HTTP 和 SSE 客户端用于连接远程 MCP 服务。

#### Scenario: HTTP 连接
- **WHEN** 通过 HTTP 连接远程 MCP 服务
- **THEN** 系统通过 POST 发送 JSON-RPC 请求并接收 JSON-RPC 响应

#### Scenario: SSE 连接
- **WHEN** 连接支持 SSE 的远程 MCP 服务
- **THEN** 系统建立 SSE 连接并流式传输 JSON-RPC 消息

#### Scenario: 自定义 Headers
- **WHEN** 远程 MCP 服务需要自定义 Headers
- **THEN** 系统在 HTTP 请求中包含配置的 Headers

#### Scenario: 连接超时
- **WHEN** 远程服务在连接超时（5秒）内未响应
- **THEN** 系统向调用者返回超时错误

### Requirement: 本地 MCP 客户端（stdio）
系统 SHALL 实现 stdio 客户端，通过进程执行连接本地 MCP 服务。

#### Scenario: 进程启动
- **WHEN** 连接本地 MCP 服务
- **THEN** 系统使用环境变量执行配置的命令

#### Scenario: stdio 通信
- **WHEN** 与本地 MCP 进程通信
- **THEN** 系统通过 stdin 发送 JSON-RPC 并通过 stdout 接收 JSON-RPC

#### Scenario: 进程错误处理
- **WHEN** 本地 MCP 进程写入 stderr
- **THEN** 系统记录 stderr 输出用于调试

#### Scenario: 进程启动超时
- **WHEN** 本地 MCP 进程在启动超时（10秒）内未响应
- **THEN** 系统终止进程并向调用者返回错误

### Requirement: 进程生命周期管理
系统 SHALL 使用空闲超时管理本地 MCP 进程生命周期，但不进行监控。

#### Scenario: 进程空闲超时
- **WHEN** 本地 MCP 进程空闲 5 分钟
- **THEN** 系统终止该进程

#### Scenario: 进程崩溃
- **WHEN** 本地 MCP 进程在请求期间崩溃
- **THEN** 系统向客户端返回错误且不自动重启

#### Scenario: 崩溃后的手动重启
- **WHEN** 客户端在进程崩溃后发送新请求
- **THEN** 系统启动新的进程实例

### Requirement: 请求超时
系统 SHALL 对所有 MCP 操作强制请求超时（30秒）。

#### Scenario: 工具调用超时
- **WHEN** 工具执行超过 30 秒
- **THEN** 系统向客户端返回超时错误

#### Scenario: 资源读取超时
- **WHEN** 资源读取超过 30 秒
- **THEN** 系统向客户端返回超时错误

### Requirement: 连接错误传播
系统 SHALL 在连接失败时返回清晰的 JSON-RPC 错误。

#### Scenario: 服务不可用
- **WHEN** MCP 服务不可访问
- **THEN** 系统返回 JSON-RPC 错误，代码 -32603，消息 "MCP 服务不可用: {symbol}"

#### Scenario: 认证失败
- **WHEN** 远程 MCP 服务因无效凭据拒绝连接
- **THEN** 系统返回 JSON-RPC 错误，代码 -32603，消息 "MCP 服务认证失败"

### Requirement: JSON-RPC 协议处理
系统 SHALL 对所有 MCP 通信正确实现 JSON-RPC 2.0 协议。

#### Scenario: 请求 ID 关联
- **WHEN** 系统发送带 ID 的 JSON-RPC 请求
- **THEN** 系统关联相同 ID 的响应

#### Scenario: 通知处理
- **WHEN** MCP 服务发送通知（无 ID）
- **THEN** 系统处理通知而不期待响应

#### Scenario: 批量请求支持
- **WHEN** 客户端发送 JSON-RPC 请求数组
- **THEN** 系统处理每个请求并返回响应数组