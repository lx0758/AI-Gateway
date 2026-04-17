## 1. 后端修复

- [x] 1.1 修复 `activeProviderModels` 统计查询，添加 `deleted_at IS NULL` 条件
- [x] 1.2 修复 `totalMCPTools` 统计，将 `Table("mcp_tools")` 改为 `Model(&MCPTool{})`
- [x] 1.3 修复 `totalMCPResources` 统计，将 `Table("mcp_resources")` 改为 `Model(&MCPResource{})`
- [x] 1.4 修复 `totalMCPPrompts` 统计，将 `Table("mcp_prompts")` 改为 `Model(&MCPPrompt{})`
- [x] 1.5 修复 `activeMCPTools` 统计查询，添加 `deleted_at IS NULL` 条件
- [x] 1.6 修复 `activeMCPResources` 统计查询，添加 `deleted_at IS NULL` 条件
- [x] 1.7 修复 `activeMCPPrompts` 统计查询，添加 `deleted_at IS NULL` 条件

## 2. 前端优化

- [x] 2.1 修改"厂商"卡片颜色为 primary（移除默认样式，显式添加 primary）
- [x] 2.2 修改"厂商模型"卡片颜色为 primary（移除 warning）
- [x] 2.3 修改"MCP服务"卡片颜色为 warning（移除 danger）
- [x] 2.4 修改"MCP工具"卡片颜色为 warning
- [x] 2.5 修改"MCP资源"卡片颜色为 warning（移除 primary）
- [x] 2.6 修改"MCP提示词"卡片颜色为 warning（移除 danger）
- [x] 2.7 统一 MCP 数据量格式化为 1 位小数，与其他统计项保持一致

## 3. 验证

- [x] 3.1 启动服务，验证仪表盘统计数据正常
- [x] 3.2 创建并软删除一条 ProviderModel 记录，验证统计正确
- [x] 3.3 验证资产卡片颜色显示正确
