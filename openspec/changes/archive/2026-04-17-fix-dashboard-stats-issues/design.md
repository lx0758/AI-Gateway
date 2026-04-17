## Context

仪表盘资产统计存在 GORM 软删除机制使用不一致的问题。部分统计使用 `Model(&Entity{})` 方法（自动过滤软删除），部分使用 `Table()` 或 `Raw()` 方法（不自动过滤软删除），导致统计结果不一致。

```
┌─────────────────────────────────────────────────────────────┐
│                    GORM 软删除机制                           │
├─────────────────────────────────────────────────────────────┤
│  ✅ Model(&Entity{}) → 自动添加 WHERE deleted_at IS NULL   │
│  ❌ Table("table_name") → 不添加软删除条件                  │
│  ❌ Raw("SELECT ...") → 不添加软删除条件                    │
└─────────────────────────────────────────────────────────────┘
```

受影响的实体：`Provider`、`ProviderModel`、`MCP`、`MCPTool`、`MCPResource`、`MCPPrompt` 均定义了 `DeletedAt gorm.DeletedAt` 字段。

## Goals / Non-Goals

**Goals:**
- 确保所有资产统计查询正确过滤软删除记录
- 统一仪表盘资产卡片颜色方案，提升视觉一致性

**Non-Goals:**
- 不改变统计的业务逻辑
- 不修改数据库结构

## Decisions

### Decision 1: 使用 GORM Model 方法替代 Table/Raw

**选择**: 将 `Table()` 调用改为 `Model(&Entity{})`，将 `Raw()` SQL 中添加 `deleted_at IS NULL` 条件

**理由**:
- `Model()` 方法更符合 GORM 最佳实践
- 保持代码风格一致性
- 自动获得软删除过滤能力

**替代方案**:
- 在所有 Raw SQL 中手动添加 `deleted_at IS NULL` 条件
  - 缺点：容易遗漏，维护成本高

### Decision 2: 颜色方案按功能分组

**选择**: 
- primary（蓝）→ AI 模型链路（厂商、厂商模型、模型）
- success（绿）→ 认证鉴权（Keys）
- warning（橙）→ MCP 生态（服务、工具、资源、提示词）
- danger（红）→ 不用于正常统计

**理由**:
- 颜色语义与功能匹配
- danger 颜色应保留给警告状态，避免视觉疲劳

## Risks / Trade-offs

**风险**: 修改后统计数据会减少（排除软删除记录）

**缓解**: 这是预期行为，修复后数据才准确
