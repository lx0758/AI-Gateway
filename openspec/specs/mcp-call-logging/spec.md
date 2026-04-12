## ADDED Requirements

### Requirement: 记录 MCP 工具调用

系统 SHALL 将每次 MCP 工具调用记录到 `mcp_logs` 表，包含以下信息：
- `source`: "mcp-proxy"
- `client_ips`: 客户端 IP 和转发链（逗号分隔）
- `key_id`: 请求使用的 API Key ID
- `key_name`: API Key 名称
- `mcp_id`: MCP 服务 ID
- `mcp_name`: MCP 服务名称
- `mcp_type`: MCP 服务类型（remote/local）
- `call_type`: "tool"
- `call_method`: "call"
- `call_target`: 工具模块名称
- `input_size`: 输入参数大小（字节）（int64）
- `output_size`: 输出内容大小（字节）（int64）
- `latency_ms`: 调用延迟（毫秒）（int64）
- `status`: "success" 或 "error"
- `error_msg`: 失败时的错误消息，成功时为空
- `created_at`: 日志创建时间戳

#### Scenario: 成功的工具调用

- **WHEN** 通过 tools/call 方法成功完成 MCP 工具调用
- **THEN** 系统记录 MCP 日志，call_type="tool"、call_method="call"、status="success"、input_size 来自输入参数、output_size 来自响应内容、latency_ms 为计算值

#### Scenario: 失败的工具调用

- **WHEN** MCP 工具调用因错误失败
- **THEN** 系统记录 MCP 日志，call_type="tool"、call_method="call"、status="error"、error_msg 包含错误内容、input_size 和 output_size 为 0

### Requirement: 记录 MCP 资源读取

系统 SHALL 将每次 MCP 资源读取记录到 `mcp_logs` 表，包含以下信息：
- 与工具调用相同的字段，但：
- `call_type`: "resource"
- `call_method`: "read"
- `call_target`: 资源模块名称

#### Scenario: 成功的资源读取

- **WHEN** 通过 resources/read 方法成功完成 MCP 资源读取
- **THEN** 系统记录 MCP 日志，call_type="resource"、call_method="read"、status="success"、input_size 为 0、output_size 来自响应内容、latency_ms 为计算值

#### Scenario: 失败的资源读取

- **WHEN** MCP 资源读取因错误失败
- **THEN** 系统记录 MCP 日志，call_type="resource"、call_method="read"、status="error"、error_msg 包含错误内容、input_size 和 output_size 为 0

### Requirement: 记录 MCP 提示词获取

系统 SHALL 将每次 MCP 提示词获取记录到 `mcp_logs` 表，包含以下信息：
- 与工具调用相同的字段，但：
- `call_type`: "prompt"
- `call_method`: "get"
- `call_target`: 提示词模块名称

#### Scenario: 成功的提示词获取

- **WHEN** 通过 prompts/get 方法成功完成 MCP 提示词获取
- **THEN** 系统记录 MCP 日志，call_type="prompt"、call_method="get"、status="success"、input_size 来自参数、output_size 来自响应内容、latency_ms 为计算值

#### Scenario: 失败的提示词获取

- **WHEN** MCP 提示词获取因错误失败
- **THEN** 系统记录 MCP 日志，call_type="prompt"、call_method="get"、status="error"、error_msg 包含错误内容、input_size 和 output_size 为 0

### Requirement: 不记录元数据查询

系统 SHALL NOT 记录元数据查询调用的 MCP 日志，包括：
- initialize 方法
- tools/list 方法
- resources/list 方法
- prompts/list 方法
- ping 方法

#### Scenario: 元数据查询不记录

- **WHEN** 执行 MCP 元数据查询（initialize、list 方法或 ping）
- **THEN** 系统不创建 MCP 日志条目

### Requirement: 计算输入和输出大小

系统 SHALL 为每次 MCP 调用计算并记录 input_size 和 output_size：
- `input_size`: 输入数据的大小（字节）（tools/call 和 prompts/get 的参数，resources/read 为 0）
- `output_size`: 输出内容的大小（字节）（来自 MCP 服务器的响应内容）

#### Scenario: 计算工具调用的输入大小

- **WHEN** 记录工具调用时
- **THEN** 系统计算 input_size 为 JSON 编码的输入参数的字节大小

#### Scenario: 计算资源读取的输出大小

- **WHEN** 记录资源读取时
- **THEN** 系统计算 output_size 为响应内容的字节大小

### Requirement: 计算延迟

系统 SHALL 通过测量从调用开始到完成的时间来为每次 MCP 调用计算 latency_ms。

#### Scenario: 计算延迟

- **WHEN** MCP 调用开始时
- **THEN** 系统记录开始时间
- **AND** 调用完成后，系统计算 latency_ms 为完成时间与开始时间的差值（毫秒）

### Requirement: 数据库索引

系统 SHALL 为 `mcp_logs` 表在以下列创建数据库索引：
- `key_id`: 用于按 API Key 查询日志
- `mcp_id`: 用于按 MCP 服务查询日志
- `call_type`: 用于按调用类型（tool/resource/prompt）查询日志
- `created_at`: 用于时间范围查询

#### Scenario: 按 Key 查询日志

- **WHEN** 用户按 key_id 查询 MCP 日志
- **THEN** 查询使用 key_id 索引进行高效查找

#### Scenario: 按时间范围查询日志

- **WHEN** 用户按 created_at 时间范围查询 MCP 日志
- **THEN** 查询使用 created_at 索引进行高效查找

### Requirement: 查询 MCP 日志 API

系统 SHALL 提供 API 端点查询 MCP 日志，支持时间范围过滤。

#### Scenario: 使用默认时间范围查询日志

- **WHEN** 用户请求 GET /api/v1/usage/mcp-logs 且不带参数
- **THEN** 系统返回今天的 MCP 日志（从 00:00:00 到 23:59:59）

#### Scenario: 使用自定义时间范围查询日志

- **WHEN** 用户请求 GET /api/v1/usage/mcp-logs 并带 start_date 和 end_date 参数
- **THEN** 系统返回指定时间范围内的 MCP 日志
- **AND** 响应包含按 created_at 降序排序的日志

#### Scenario: 查询日志响应格式

- **WHEN** 用户成功查询 MCP 日志
- **THEN** 响应包含 "logs" 数组，每个日志包含：id、source、client_ips、key_id、key_name、mcp_id、mcp_name、mcp_type、call_type、call_target、call_method、input_size、output_size、latency_ms、status、error_msg、created_at

### Requirement: 显示 MCP 使用统计

系统 SHALL 在 MCPUsage 页面显示 MCP 使用统计，包含以下组件：

#### Scenario: 显示汇总统计

- **WHEN** 用户打开 MCPUsage 页面
- **THEN** 系统显示汇总卡片：总请求数、成功率、总数据大小（输入 + 输出）、平均延迟

#### Scenario: 显示按来源统计

- **WHEN** 用户查看 MCPUsage 页面
- **THEN** 系统显示按来源分组的统计表（接入点统计）
- **AND** 每行显示：来源、调用次数、输入大小、输出大小、平均延迟

#### Scenario: 显示按 IP 统计

- **WHEN** 用户查看 MCPUsage 页面
- **THEN** 系统显示按 client_ips 分组的统计表（IP 统计）
- **AND** 每行显示：客户端 IP（链中第一个 IP）、完整 IP 链、调用次数、输入大小、输出大小、平均延迟

#### Scenario: 显示按调用类型统计

- **WHEN** 用户查看 MCPUsage 页面
- **THEN** 系统显示按 call_type 分组的统计表（调用类型统计）
- **AND** 每行显示：调用类型（tool/resource/prompt）、调用次数、输入大小、输出大小、平均延迟

#### Scenario: 显示按 Key 统计

- **WHEN** 用户查看 MCPUsage 页面
- **THEN** 系统显示按 key_name 分组的统计表（Key 统计）
- **AND** 每行显示：Key 名称、调用次数、输入大小、输出大小、平均延迟

#### Scenario: 显示按 MCP 服务统计

- **WHEN** 用户查看 MCPUsage 页面
- **THEN** 系统显示按 mcp_name 分组的统计表（MCP 服务统计）
- **AND** 每行显示：MCP 服务名称、调用次数、输入大小、输出大小、平均延迟

#### Scenario: 显示按 MCP 服务类型统计

- **WHEN** 用户查看 MCPUsage 页面
- **THEN** 系统显示按 mcp_type 分组的统计表（MCP 服务类型统计）
- **AND** 每行显示：MCP 服务类型（remote/local）、调用次数、输入大小、输出大小、平均延迟

#### Scenario: 显示按 MCP 服务和调用类型统计

- **WHEN** 用户查看 MCPUsage 页面
- **THEN** 系统显示按 mcp_name 和 call_type 分组的统计表（MCP 服务+调用类型统计）
- **AND** 每行显示：MCP 服务名称、调用类型、调用次数、输入大小、输出大小、平均延迟

### Requirement: 显示 MCP 日志详情

系统 SHALL 在 MCPUsage 页面以详细表格显示 MCP 日志。

#### Scenario: 显示日志表格

- **WHEN** 用户查看 MCPUsage 页面
- **THEN** 系统显示 MCP 日志表格
- **AND** 列包含：时间、来源、客户端 IP、Key 名称、MCP 服务名称、调用类型、调用目标、输入大小、输出大小、延迟、状态、错误消息

#### Scenario: 格式化数据大小

- **WHEN** 在日志表格中显示 input_size 和 output_size
- **THEN** 系统将字节格式化为人类可读格式（例如 1024 字节 → "1 KB")

#### Scenario: 格式化延迟

- **WHEN** 在日志表格中显示 latency_ms
- **THEN** 系统将毫秒格式化为人类可读格式（例如 1500 ms → "1.5 s")

#### Scenario: 显示 IP 链 Tooltip

- **WHEN** 日志在 client_ips 字段中有多个 IP
- **THEN** 系统在表格单元格中显示第一个 IP
- **AND** 悬停时显示完整 IP 链的 Tooltip

#### Scenario: 显示错误消息

- **WHEN** 日志有 status="error"
- **THEN** 系统在错误列显示 error_msg
- **AND** 提供复制按钮以复制错误消息