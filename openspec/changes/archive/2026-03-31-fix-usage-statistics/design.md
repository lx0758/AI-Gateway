## Context

Provider 重构后，`proxy_openai.go` 和 `proxy_anthropic.go` 中的 TODO 注释未实现，导致 `usage_logs` 表无数据写入，所有统计功能失效。

当前 `UsageLog` 结构包含 `PromptTokens` 和 `CompletionTokens` 两个 int 字段，但不同厂商返回的 usage 信息格式不一，难以统一收集精确数据。

前端 Dashboard 和 Usage 页面的统计展示维度不完整，缺少厂商和 Key 维度的统计，也缺少耗时统计。

## Goals / Non-Goals

**Goals:**
- 修复数据写入逻辑，确保每次 API 调用都记录到 `usage_logs` 表
- 简化字段结构，使用 `TotalTokens` 代替分开的 prompt/completion tokens
- 新增厂商和 Key 维度的统计
- 新增耗时统计（AVG/MAX/MIN）
- 优化前端展示，增加统计表格和卡片

**Non-Goals:**
- 不处理历史数据迁移（项目未上线）
- 不支持精确的 prompt/completion tokens 分别统计
- 不做实时统计（基于数据库查询即可）

## Decisions

### 1. 字段简化

**决定**: 移除 `PromptTokens`/`CompletionTokens`，新增 `TotalTokens`

**理由**: 
- 不同厂商返回的 usage 格式差异大（OpenAI 有 prompt/completion，Anthropic 有 input/output，其他厂商可能只有 total）
- 用户主要关心总消耗，分开统计价值不大
- 简化 Provider 接口，只需返回总数

**类型选择**: 使用 `int64` 而非 `int`
- 单次请求 tokens 可能很大
- SUM 聚合后可能溢出
- 与 `APIKey.Quota`/`UsedQuota` 保持一致

### 2. 数据写入位置

**决定**: 在 Handler 层写入，而非 Provider 层

**理由**:
- Handler 层已有所有需要的信息（apiKeyID, providerID, model, tokens, err）
- Handler 可自行计算耗时（请求前后计时）
- Provider 层保持简单，只负责请求转发和响应转换

### 3. 统计查询

**决定**: 基于现有 `usage_logs` 表，增加新的查询 API

**理由**:
- 数据量可控，不需要额外的聚合表
- 实时查询满足需求
- 复用现有表结构

## Risks / Trade-offs

**风险**: tokens 数据不够精确
→ **可接受**: 用户已确认不需要精确到 prompt/completion 级别

**风险**: 数据库写入增加延迟
→ **缓解**: 异步写入或批量写入（当前规模不需要）

**风险**: 前端改动较大
→ **缓解**: 保持向后兼容，逐步优化
