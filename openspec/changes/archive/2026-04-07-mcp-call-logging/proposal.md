## Why

当前系统缺少 MCP 调用统计功能，无法追踪 MCP 服务使用情况、识别性能瓶颈或进行调用审计。随着 MCP 协议支持的增加，需要与 ModelLog 类似的日志记录机制来监控 MCP 调用的健康状态和性能指标。

## What Changes

- 新增 MCPLog 数据模型，记录 MCP 调用日志（工具调用、资源读取、提示词获取）
- MCPProxyHandler 在关键调用点记录日志：tools/call、resources/read、prompts/get
- 不记录输入输出内容（只记录大小）
- 不记录元数据查询调用（initialize、tools/list、resources/list、prompts/list）
- UsageHandler 新增 MCPLogs() 方法，处理 /api/v1/usage/mcp-logs 路由
- UsageHandler 新增 NewMCPLog() 辅助函数，用于创建 MCP 日志
- 前端新增 MCPUsage 页面，展示 MCP 调用统计和日志明细（参考 ModelUsage 页面设计）
- 多维度分组统计：接入点、IP、调用类型、Key、MCP服务、MCP服务类型、MCP服务+调用类型

## Capabilities

### New Capabilities

- `mcp-call-logging`: MCP 调用日志记录能力，包括：
  - 记录工具调用、资源读取、提示词获取三类实际调用
  - 记录调用基本信息（Key、MCP服务、调用类型、调用目标）
  - 记录性能指标（输入输出大小、耗时）
  - 记录状态信息（成功/失败、错误信息）
  - 不记录输入输出内容，仅记录大小（节省存储，保护隐私）
  - 不记录元数据查询调用（initialize、list类方法）
  - 提供按时间范围查询日志的 API
  - 提供按多个维度统计的 API（接入点、IP、调用类型、Key、MCP服务、MCP服务类型、MCP服务+调用类型）
  - 前端 MCPUsage 页面展示统计卡片和日志明细

### Modified Capabilities

无

## Impact

### 后端影响

- `server/internal/model/db.go`: 新增 MCPLog 结构体，autoMigrate() 添加 MCPLog
- `server/internal/handler/usage.go`: 新增 MCPLogs() 方法、NewMCPLog() 辅助函数、mcpLogResponse DTO
- `server/internal/handler/proxy_mcp.go`: 在 handleToolsCall、handleResourcesRead、handlePromptsGet 中调用 NewMCPLog() 记录日志
- `server/cmd/server/main.go`: 注册路由 GET /api/v1/usage/mcp-logs

### 前端影响

- `web/src/views/MCPUsage/index.vue`: 新增 MCPUsage 页面（统计卡片、分组统计表格、日志明细）
- `web/src/router/index.ts`: 新增 /mcp-usage 路由
- `web/src/locales/zh.ts`: 新增 mcpUsage 翻译键
- `web/src/components/layout/MainLayout.vue`: 新增 MCP 用量菜单项

### 数据库影响

- SQLite/PostgreSQL: 新增 mcp_logs 表，包含索引（key_id、mcp_id、call_type、created_at）

### API 影响

- 新增 GET /api/v1/usage/mcp-logs: 查询 MCP 调用日志（支持时间范围筛选）