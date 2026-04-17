## Why

仪表盘资产概览存在三个问题：
1. **数据统计错误**：部分统计查询未正确过滤软删除记录，导致"可用数量 > 总数量"的异常
2. **UI 视觉混乱**：资产卡片颜色分配无规律，danger（红色）用于正常统计不合理，影响用户体验
3. **格式化不一致**：MCP 数据量显示 2 位小数，而其他统计项均为 1 位小数

## What Changes

### 后端修复

- 修复 `activeProviderModels` 统计：Raw SQL 查询添加 `deleted_at IS NULL` 条件
- 修复 `totalMCPTools/Resources/Prompts` 统计：将 `Table()` 调用改为 `Model()` 方法，确保软删除过滤
- 修复 `activeMCPTools/Resources/Prompts` 统计：Raw SQL 查询添加 `deleted_at IS NULL` 条件

### 前端优化

- 统一资产卡片颜色方案，按功能分组：
  - **primary（蓝色）**：厂商、厂商模型、模型（AI 模型链路）
  - **success（绿色）**：Keys（认证鉴权）
  - **warning（橙色）**：MCP服务、MCP工具、MCP资源、MCP提示词（MCP 生态）
  - **danger（红色）**：仅用于警告/错误状态
- 统一 MCP 数据量格式化为 1 位小数，与其他统计项保持一致

## Capabilities

### New Capabilities

无新增能力

### Modified Capabilities

- `usage-tracking`: 修改仪表盘资产统计需求，明确要求统计应排除软删除的记录

## Impact

**后端**：
- `server/internal/handler/usage.go`：Dashboard 方法中的统计查询

**前端**：
- `web/src/views/Dashboard/index.vue`：资产概览卡片的颜色类名、formatSize 函数

**数据库**：无变更
