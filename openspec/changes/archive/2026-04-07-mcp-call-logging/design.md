## Context

当前系统已实现 ModelLog 用于记录 AI 模型调用日志，包括 Token 消耗、延迟、状态等信息。随着 MCP 协议支持的增加（MCPProxyHandler），需要类似的日志机制来监控 MCP 服务调用。

**参考模式**：
- ModelLog 结构设计（字段、索引）
- UsageHandler 的日志创建模式（NewModelLog 函数）
- 数据库自动迁移机制

**约束**：
- 不记录输入输出内容（节省存储，保护隐私）
- 不记录元数据查询调用（initialize、list 类方法）
- 前端页面参考 Usage 页面设计，保持一致性

## Goals / Non-Goals

**Goals:**

- 实现 MCPLog 数据模型，记录 MCP 实际调用（tool call、resource read、prompt get）
- 在 MCPProxyHandler 关键调用点记录日志
- 确保日志记录不影响主流程性能（同步写入）
- 数据模型兼容 SQLite 和 PostgreSQL
- 实现 MCPUsageHandler，提供日志查询和统计 API
- 实现前端 MCPUsage 页面，展示多维度统计和日志明细

**Non-Goals:**

- 不在 Dashboard 增加 MCP 统计概览
- 不记录元数据查询调用（initialize、tools/list、resources/list、prompts/list）
- 不记录输入输出内容完整数据（仅记录大小）

## Decisions

### D1: MCPLog 字段设计

**决策**：采用精简字段设计，只记录必要信息。

**字段列表**：
- `ID`: 主键
- `Source`: 来源（如 "mcp-proxy"）
- `ClientIPs`: 客户端 IP（复用现有 IP 获取逻辑）
- `KeyID/KeyName`: Key 信息
- `MCPID/MCPName/MCPType`: MCP 服务信息
- `CallType`: 调用类型（tool/resource/prompt）
- `CallTarget`: 调用目标模块名称（tool_name/resource_name/prompt_name）
- `CallMethod`: 调用方法（call/read/get）
- `InputSize`: 输入大小（字节）
- `OutputSize`: 输出大小（字节）
- `LatencyMs`: 调用耗时（毫秒）
- `Status`: 状态（success/error）
- `ErrorMsg`: 错误信息
- `CreatedAt`: 创建时间

**理由**：
- 参考 ModelLog 设计，保持一致性
- 不记录 InputArgs/OutputContent（节省存储，保护隐私）
- CallType/CallTarget/CallMethod 组合可完整描述调用类型
- CallTarget 统一存放模块名称，便于统计和展示

**替代方案**：
- 记录完整输入输出内容 → 存储成本高，隐私风险大，性能影响
- 不记录 InputSize/OutputSize → 无法分析数据量趋势

### D2: 日志记录点选择

**决策**：只在实际调用点记录日志，不记录元数据查询。

**记录点**：
- `handleToolsCall`: 记录 tool call（CallType="tool", CallMethod="call"）
- `handleResourcesRead`: 记录 resource read（CallType="resource", CallMethod="read"）
- `handlePromptsGet`: 记录 prompt get（CallType="prompt", CallMethod="get"）

**不记录点**：
- `handleInitialize`: 元数据查询，频率高但价值低
- `handleToolsList`: 列表查询，不涉及实际调用
- `handleResourcesList`: 列表查询
- `handlePromptsList`: 列表查询
- `handlePing`: 心跳检测，无业务价值

**理由**：
- 实际调用才是真正的业务行为，值得记录
- 元数据查询频率高，记录会显著增加日志量和写入开销

### D3: 日志写入时机

**决策**：在调用完成后同步写入日志。

**流程**：
```
开始调用 → 记录 startTime → 执行调用 → 计算 latency → 创建 MCPLog → DB.Create() → 返回响应
```

**理由**：
- 参考 ModelLog 的同步写入模式
- MCP 调用频率低于 Model 调用，同步写入性能影响小
- 同步写入确保日志完整性，避免异步丢失

**替代方案**：
- 异步写入（goroutine） → 可能丢失日志，问题排查困难
- 批量写入 → 增加复杂度，实时性差

### D4: 索引设计

**决策**：为常用查询维度建立索引。

**索引列表**：
- `key_id`: 按 Key 统计查询
- `mcp_id`: 按 MCP 服务统计查询
- `call_type`: 按调用类型统计查询
- `created_at`: 时间范围查询（必需）

**理由**：
- 参考 ModelLog 的索引设计
- 时间范围查询是日志查询的基本需求
- Key、MCP、CallType 是最常用的统计维度

**替代方案**：
- 不建索引 → 查询性能差，大数据量下问题严重
- 建更多索引（如 status、mcp_type） → 增加写入开销，暂不需要

### D5: Handler 设计

**决策**：复用 UsageHandler，新增 MCPLogs() 方法和 NewMCPLog() 辅助函数。

**设计要点**：
- `UsageHandler` 新增 `MCPLogs(c *gin.Context)` 方法，处理 `/api/v1/usage/mcp-logs` 路由
- `UsageHandler` 新增 `NewMCPLog()` 辅助函数，类似现有的 `NewModelLog()` 函数
- 新增 `mcpLogResponse` DTO 结构体，用于返回 MCP 日志数据
- 不创建新的 MCPUsageHandler，保持代码结构一致性

**理由**：
- Usage 页面和 MCPUsage 页面功能相似，复用 Handler 更合理
- 遵循现有架构模式（UsageHandler 包含 ModelLogs 和 MCPLogs）
- 减少代码重复，降低维护成本

**替代方案**：
- 创建独立的 MCPUsageHandler → 功能重复，维护成本高
- 在 MCPProxyHandler 中处理日志查询 → 职责混乱（代理和查询不应混合）

### D6: 路由设计

**决策**：MCP 日志查询路由使用 `/api/v1/usage/mcp-logs`。

**路由映射**：
- `GET /api/v1/usage/logs` → `UsageHandler.ModelLogs()`（现有的模型日志）
- `GET /api/v1/usage/mcp-logs` → `UsageHandler.MCPLogs()`（新增的 MCP 日志）
- `GET /api/v1/usage/dashboard` → `UsageHandler.Dashboard()`（现有的仪表盘）

**理由**：
- 统一放在 `/api/v1/usage` 路径下，语义清晰
- 与前端 Usage 页面和 MCPUsage 页面结构对应
- 便于后续扩展（如添加更多类型的日志）

### D7: 前端页面设计

**决策**：创建独立的 MCPUsage 页面，参考 Usage 页面设计。

**页面结构**：
```
顶部：统计卡片（总请求数、成功率、总数据量、平均耗时）
时间范围：日期时间选择器

分组统计卡片（按优先级排列）：
1. 接入点统计（source）
2. IP 统计（client_ips）
3. 调用类型统计（call_type）
4. Key 统计（key_name）
5. MCP 服务统计（mcp_name）
6. MCP 服务类型统计（mcp_type）
7. MCP 服务+调用类型统计（mcp_name + call_type）

日志明细表格：
- 时间、接入点、IP、Key、MCP服务、调用类型
- 调用目标、输入大小、输出大小、耗时、状态、错误信息
```

**理由**：
- 参考 Usage 页面的成熟设计，保持一致性
- 多维度统计帮助用户全面了解 MCP 使用情况
- 独立页面避免与 Usage 页面混淆

**替代方案**：
- 合并到 Usage 页面 → 功能混杂，不够清晰
- 只显示日志明细 → 缺少统计汇总，不够直观

### D8: 分组统计维度设计

**决策**：提供 7 个维度的分组统计。

**维度列表**：
1. **接入点统计**（source）：按来源分组（如 mcp-proxy）
2. **IP 统计**（client_ips）：按客户端 IP 分组，显示完整 IP 链
3. **调用类型统计**（call_type）：按 tool/resource/prompt 分组
4. **Key 统计**（key_name）：按 API Key 分组
5. **MCP 服务统计**（mcp_name）：按 MCP 服务名称分组
6. **MCP 服务类型统计**（mcp_type）：按 remote/local 分组
7. **MCP 服务+调用类型统计**（mcp_name + call_type）：组合维度，类似厂商模型统计

**统计指标**：
- 调用次数（count）
- 输入大小（input_size）
- 输出大小（output_size）
- 平均耗时（avg_latency）

**理由**：
- 参考 Usage 页面的统计维度
- 多维度覆盖不同分析需求（谁调用了、调用了什么、性能如何）
- 前端聚合计算，后端只提供原始日志（简化 API）

**替代方案**：
- 后端提供聚合 API → 增加后端复杂度，灵活性差
- 减少统计维度 → 信息不够全面

## Risks / Trade-offs

### R1: 日志写入延迟

**风险**：每条 MCP 调用增加一次数据库写入，影响响应延迟。

**缓解**：
- MCP 调用频率低于 Model 调用（通常每会话只调用几次）
- 日志写入时间通常 <10ms，对整体延迟影响小
- SQLite 单连接限制下，日志写入顺序执行，但 MCP 调用本身也较慢

**监控**：可后续观察 LatencyMs 指标，如果异常可考虑异步写入。

### R2: 存储增长

**风险**：mcp_logs 表会持续增长，长期运行后可能占用大量存储。

**缓解**：
- 不记录输入输出内容，大幅减少单条日志大小（约100-200字节）
- 参考 model_logs 表的管理模式（暂无清理策略，后续可添加）
- PostgreSQL 下可使用分区表或定期清理

**建议**：后续可添加日志清理策略（如保留30天）。

### R3: 错误日志丢失

**风险**：调用失败时，错误信息可能包含敏感数据，但无法过滤。

**缓解**：
- ErrorMsg 字段记录原始错误，便于排查问题
- 错误信息通常不包含用户数据，而是 MCP 服务端的错误描述
- 如果后续发现敏感数据泄露，可添加错误信息截断或过滤

## Open Questions

无（设计方案已完整）。