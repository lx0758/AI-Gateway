## Context

当前系统中，当 Provider 被禁用（`enabled=false`）时：
- **路由层**（`router.go`）已正确过滤禁用的 Provider，不会参与实际请求路由
- **API 层**（`model.go`、`key.go`）在查询 ModelMapping 时未过滤禁用的 Provider

这导致 API 返回的模型映射列表包含不可用的映射，统计数据不准确，用户体验混乱。

## Goals / Non-Goals

**Goals:**
- API 层查询 ModelMapping 时过滤掉 `Provider.enabled=false` 的记录
- 统计计算函数正确处理禁用的 Provider
- 确保前后端数据一致，用户只看到可用的映射

**Non-Goals:**
- 不修改路由层逻辑（已正确实现）
- 不添加软删除或级联禁用功能
- 不修改数据库 schema

## Decisions

### 决策 1: 过滤方式 - 数据库层 JOIN 过滤

**选择:** 使用数据库 JOIN 过滤 Provider.enabled

**理由:**
- 性能更好：数据库层面过滤，减少传输数据量
- 代码简洁：一次查询完成过滤
- 一致性：所有相关查询统一使用相同方式

**替代方案:**
- 应用层过滤：查询所有 ModelMapping 后在代码中过滤
  - 缺点：传输冗余数据，性能较差

### 决策 2: 统计函数处理方式

**选择:** 在 `calculateMinTokens` 和 `calculateCapabilitiesIntersection` 函数中增加 Provider.enabled 检查

**理由:**
- 这些函数已存在，只需增加过滤逻辑
- 保持函数职责清晰：计算已启用映射的统计数据

**实现方式:**
- 遍历 mappings 时，检查 `m.Provider.Enabled`
- 需确保调用这些函数前已 Preload Provider

## Risks / Trade-offs

**风险 1: Preload 缺失导致空指针**
- **风险:** 如果调用统计函数时未 Preload Provider，`m.Provider` 为 nil
- **缓解:** 检查 `m.Provider != nil` 后再访问 `Enabled` 字段

**风险 2: 已有数据依赖禁用 Provider 的映射**
- **风险:** 用户可能依赖 API 返回的禁用 Provider 映射进行某些操作
- **缓解:** 这是预期行为改变，符合业务逻辑。禁用的 Provider 本就不应显示

**权衡:**
- 完全隐藏 vs 标记显示：选择完全隐藏，避免用户困惑
- 性能 vs 灵活性：选择性能（数据库层过滤），牺牲一些灵活性
