## 1. 数据模型

- [x] 1.1 在 server/internal/model/db.go 中定义 MCPLog 结构体
- [x] 1.2 在 autoMigrate() 中添加 MCPLog 自动迁移
- [x] 1.3 为 MCPLog 添加 TableName() 和 String() 方法

## 2. 工具调用日志记录

- [x] 2.1 在 UsageHandler 中创建 NewMCPLog 辅助函数
- [x] 2.2 在 handleToolsCall 中添加日志记录逻辑：
  - [x] 2.2.1 记录调用开始时间
  - [x] 2.2.2 计算 latency_ms
  - [x] 2.2.3 计算 input_size（arguments JSON大小）
  - [x] 2.2.4 计算 output_size（response JSON大小）
  - [x] 2.2.5 调用 NewMCPLog 创建日志并写入数据库

## 3. 资源读取日志记录

- [x] 3.1 在 handleResourcesRead 中添加日志记录逻辑：
  - [x] 3.1.1 记录调用开始时间
  - [x] 3.1.2 计算 latency_ms
  - [x] 3.1.3 设置 input_size 为 0
  - [x] 3.1.4 计算 output_size（response JSON大小）
  - [x] 3.1.5 调用 NewMCPLog 创建日志并写入数据库

## 4. 提示词获取日志记录

- [x] 4.1 在 handlePromptsGet 中添加日志记录逻辑：
  - [x] 4.1.1 记录调用开始时间
  - [x] 4.1.2 计算 latency_ms
  - [x] 4.1.3 计算 input_size（arguments JSON大小）
  - [x] 4.1.4 计算 output_size（response JSON大小）
  - [x] 4.1.5 调用 NewMCPLog 创建日志并写入数据库

## 5. 后端 API

- [x] 5.1 在 UsageHandler 中新增 MCPLogs() 方法：查询 MCP 日志，支持时间范围筛选
- [x] 5.2 在 UsageHandler 中新增 mcpLogResponse DTO 结构体
- [x] 5.3 在 main.go 中注册路由：GET /api/v1/usage/mcp-logs

## 6. 前端页面

- [x] 6.1 创建 web/src/views/MCPUsage/index.vue 页面
- [x] 6.2 实现顶部统计卡片（总请求数、成功率、总数据量、平均耗时）
- [x] 6.3 实现时间范围选择器（日期时间范围）
- [x] 6.4 实现 7 个分组统计表格：
  - [x] 6.4.1 接入点统计（source）
  - [x] 6.4.2 IP 统计（client_ips）
  - [x] 6.4.3 调用类型统计（call_type）
  - [x] 6.4.4 Key 统计（key_name）
  - [x] 6.4.5 MCP 服务统计（mcp_name）
  - [x] 6.4.6 MCP 服务类型统计（mcp_type）
  - [x] 6.4.7 MCP 服务+调用类型统计（mcp_name + call_type）
- [x] 6.5 实现日志明细表格（包含所有字段）
- [x] 6.6 实现数据大小格式化（字节 → KB/MB）
- [x] 6.7 实现耗时格式化（毫秒 → 秒）
- [x] 6.8 实现 IP 链显示和 tooltip
- [x] 6.9 实现错误信息复制功能

## 7. 前端路由和菜单

- [x] 7.1 在 web/src/router/index.ts 添加 /mcp_usage 路由
- [x] 7.2 在 web/src/locales/zh.ts 添加 mcpUsage 翻译键
- [x] 7.3 在 MainLayout.vue 添加 MCP 用量菜单项

## 8. 测试验证

- [ ] 8.1 启动服务，验证 mcp_logs 表自动创建（包含索引）
- [ ] 8.2 测试工具调用成功场景，验证日志记录正确
- [ ] 8.3 测试工具调用失败场景，验证错误日志记录正确
- [ ] 8.4 测试资源读取和提示词获取的日志记录
- [ ] 8.5 验证 initialize/list/ping 方法不产生日志记录
- [ ] 8.6 测试前端页面：时间范围选择、分组统计显示、日志明细显示
- [ ] 8.7 测试 API：GET /api/v1/usage/mcp-logs 返回正确数据