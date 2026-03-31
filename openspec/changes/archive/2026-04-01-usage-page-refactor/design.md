## Context

当前 Usage 页面加载时发起 3 个独立请求：
- `GET /usage/stats` - 返回总请求数、成功率、Tokens、延迟、modelStats
- `GET /usage/key-stats` - 返回 keyStats
- `GET /usage/logs` - 返回原始日志（限制 1000 条）

三个接口各自执行数据库聚合查询，存在计算冗余。1000 条限制在数据量大时导致统计不准确。

## Goals / Non-Goals

**Goals:**
- 简化后端接口，移除重复聚合逻辑
- 移除日志数量限制，返回全量数据
- 增强前端分析能力，支持多维度聚合（接入点、厂商、厂商模型）
- 统一时间范围控制

**Non-Goals:**
- 不改变数据库 schema
- 不引入新的外部依赖
- 不改变 Dashboard 页面（保持现有逻辑）

## Decisions

### 1. 单一日志接口 + 前端聚合

**决定**: 废弃 `/usage/stats` 和 `/usage/key-stats`，扩展 `/usage/logs` 返回全量数据，前端负责聚合。

**原因**:
- 减少后端重复计算
- 前端聚合更灵活，可动态调整统计维度
- 全量数据支持更多分析场景

**替代方案考虑**:
- 方案 B：后端保留聚合接口，只去掉 LIMIT → 仍存在重复计算
- 方案 C：后端新增一个聚合接口 → 增加接口数量，不简化

### 2. 日志接口时间范围

**决定**: `/usage/logs` 使用前端传来的 `start_date` 和 `end_date` 参数，无 LIMIT 限制。

**原因**: 前端已有日期选择器，直接复用最简单。

### 3. 前端聚合计算

**前端聚合计算字段**:

| 统计类型 | 聚合维度 | 计算字段 |
|---------|---------|---------|
| 基础统计 | - | count, success_count, sum(tokens), avg(latency) |
| Key 统计 | key_name | count, sum(tokens), avg(latency) |
| 模型统计 | model | count, sum(tokens), avg(latency) |
| 接入点统计 | source | count, sum(tokens), avg(latency) |
| 厂商统计 | provider_name | count, sum(tokens), avg(latency) |
| 厂商模型统计 | provider_name + model | count, sum(tokens), avg(latency) |

### 4. 卡片顺序

```
┌──────────────────────────────────────────────────────┐
│  Usage 页面卡片顺序                                    │
├──────────────────────────────────────────────────────┤
│  1. 接入点统计 (el-table: source, count, tokens)     │
│  2. Key 统计 (el-table: key_name, count, tokens)     │
│  3. 模型统计 (el-table: model, count, tokens)        │
│  4. 厂商统计 (el-table: provider_name, count, tokens)│
│  5. 厂商模型统计 (el-table: provider+model, count)   │
│  6. 日志详情 (el-table: 原始日志列表)                 │
└──────────────────────────────────────────────────────┘
```

## Risks / Trade-offs

**[风险] 大数据量下前端聚合性能**

→ ** Mitigation**: 浏览器处理数万条日志 JS 聚合通常在 100ms 内完成。如后续遇性能问题，可考虑前端分页或 Web Worker。

**[风险] 日志数据过大导致网络传输慢**

→ **Mitigation**: 可增加后端分页或限制最大时间范围（如 90 天）。当前方案下，日期范围由前端控制，默认查询近 1 天数据。

**[权衡] 后端聚合 vs 前端聚合**

→ 选择前端聚合以换取更好的灵活性和代码简化，代价是前端需要多做一些计算。对于管理后台级别数据量，这个 trade-off 是可接受的。

## Open Questions

1. 日志接口是否需要分页返回？（如返回超过 10000 条）
2. 是否需要保留 Dashboard 页面的独立接口？（本次只改 Usage 页面）
