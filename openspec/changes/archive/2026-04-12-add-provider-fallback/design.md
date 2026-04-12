## Context

当前架构中，`ModelRouter.Route()` 返回所有可用 Provider（按 weight 排序），但 `proxy_openai.go` 和 `proxy_anthropic.go` 只取 `results[0]` 执行。

Provider 返回错误时（包括 429），请求直接失败，返回给客户端。没有重试或切换逻辑。

**关键约束**：
- 单实例部署，不需要跨实例共享状态
- 被动式：不主动探测健康状态
- 流式请求：响应头已发送时无法切换（当前请求不切换）

## Goals / Non-Goals

**Goals:**
- Provider+ProviderModel 连续 3 次 429（间隔 ≥5秒）后自动进入冷却期（30 分钟）
- 冷却期内路由自动避开该组合
- 请求成功时重置计数
- 所有 Provider 都在冷却期时，选最早恢复的尝试
- Provider 配置更新时自动清除冷却状态

**Non-Goals:**
- 不做主动健康检查或探测
- 不做成本优化或价格感知
- 不做时间窗口内的 429 计数（只计连续，带缓冲间隔）
- 不持久化冷却状态（进程重启后清空）
- 当前请求不切换（只影响下次请求）

## Decisions

### 1. 冷却状态存储：内存 map

**选择**: 进程内内存 map，不持久化

**理由**:
- 单实例部署，无需跨实例共享
- 进程重启本身是罕见事件，重启后状态清空可接受
- 避免引入 Redis 等额外依赖

**替代方案**:
- SQLite/PostgreSQL 持久化 → 增加复杂度，收益有限
- Redis → 需要额外依赖，单实例不需要

### 2. 冷却 Key 设计：ProviderID + ProviderModelID

**选择**: `ProviderID:ProviderModelID` 作为 key（数据库主键）

**理由**:
- 同一 Provider 不同模型的限流策略可能不同
- 需要精确到模型级别
- 使用数据库主键比字符串 ModelID 更精确

**示例**: `5:12` 表示 Provider ID=5 的 ProviderModel ID=12

### 3. 触发逻辑：连续 3 次 429（带缓冲间隔）

**选择**: 累计连续 429，成功则重置，达到 3 次触发冷却

**缓冲间隔**: 5 秒内同一 Provider+Model 的多次 429 只计一次

**理由**:
- 高并发场景下，瞬间多个请求返回 429，若无缓冲会快速触发冷却
- 5 秒缓冲避免"瞬间判死刑"
- 成功请求证明 Provider 恢复正常，重置计数合理

**状态结构**:
```go
type CooldownState struct {
    Consecutive429  int        // 当前连续计数
    CooldownUntil   *time.Time // 冷却结束时间
    Last429Time     *time.Time // 最后一次 429 时间（用于缓冲判断）
}
```

### 4. 所有 Provider 冷却时的处理

**选择**: 选择最早恢复冷却的那个

**理由**:
- 避免直接返回 503，给用户一个尝试机会
- 最早恢复的 Provider 最可能已经恢复

### 5. Provider 错误类型：结构化

**选择**: 新增 `ProviderError` 结构体，包含 StatusCode

```go
type ProviderError struct {
    StatusCode int
    Message    string
}

func IsRateLimitError(err error) bool
```

**理由**:
- Handler 层需要区分 429 和其他错误
- 当前是字符串拼接 `fmt.Errorf("%d - %s", ...)`，无法解析

### 6. Router 单例化

**选择**: 全局单例 `ModelRouter`

```go
var globalRouter = &ModelRouter{
    cooldownManager: NewCooldownManager(),
}

func GetRouter() *ModelRouter
func RecordRateLimit(providerID, providerModelID uint)
func RecordSuccess(providerID, providerModelID uint)
func ClearAllCooldownsForProvider(providerID uint)
```

**理由**:
- OpenAI 和 Anthropic Handler 共享冷却状态
- 简化调用方式

### 7. Route() 返回单个结果

**选择**: `Route(name string) (*RouteResult, error)` 返回最优的单个结果

**理由**:
- Handler 只使用第一个结果，返回切片无意义
- Router 层完成所有决策逻辑（冷却过滤、最早恢复选择）

### 8. 配置更新时清除冷却

**选择**: Provider 配置更新时自动清除冷却状态

**触发场景**:
- Provider 从禁用变为启用
- Provider APIKey 更新
- Provider BaseURL 更新

**理由**:
- 配置变更意味着用户主动干预，应重置冷却状态
- 无需用户额外操作

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 进程重启后冷却状态丢失 | 可接受。重启是罕见事件，且 429 状态本身是临时的 |
| 冷却期过长（30 分钟）可能导致 Provider 已恢复但仍被避开 | 配置更新时自动清除；用户也可在控制台手动禁用/启用 |
| 非 429 错误（500、超时）不触发冷却 | 设计意图。只针对限流，其他错误可能是临时问题 |
| 流式请求中途失败无法切换 | 设计意图。当前请求不切换，只影响下次请求 |
| 高并发瞬间触发冷却 | 5 秒缓冲间隔，同一组合 5 秒内只计一次 |