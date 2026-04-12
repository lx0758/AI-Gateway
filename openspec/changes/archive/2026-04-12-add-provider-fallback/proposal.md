## Why

当 Provider 返回 429（限流）时，系统当前没有自动切换机制。高峰期或达到调用阈值时，用户需要手动在控制台切换 Provider，影响服务可用性。

需要一个被动式的故障转移机制：当某个 Provider 模型连续返回 429 后，系统自动在后续请求中避开它，尝试其他 Provider。

## What Changes

- 添加 Provider 故障转移机制（被动式，不主动探测）
- 当 Provider+ProviderModel 组合连续 3 次返回 429（每次间隔 ≥5秒），进入 30 分钟冷却期
- 冷却期内，路由自动避开该组合，选择其他可用 Provider
- 请求成功时，重置连续 429 计数
- 所有 Provider 都在冷却期时，选择最早恢复的那个尝试
- Provider 配置更新（启用、APIKey、BaseURL）时，自动清除相关冷却状态
- Router 单例化，全局共享冷却状态

## Capabilities

### New Capabilities

- `provider-fallback`: Provider 故障转移能力，包括冷却状态管理、路由过滤、状态更新、配置刷新

### Modified Capabilities

无（这是新功能，不改变现有行为）

## Impact

- **代码改动**：
  - `router/router.go`: Route() 返回单个最优结果，增加冷却过滤逻辑，单例化
  - `router/cooldown.go`: 新建，管理冷却状态（内存 map，带缓冲间隔）
  - `router/route_result.go`: 新建，RouteResult 结构体
  - `provider/provider_error.go`: 新建，ProviderError 结构化错误类型
  - `provider/provider_openai.go`: 返回结构化错误
  - `provider/provider_anthropic.go`: 返回结构化错误
  - `handler/proxy_openai.go`: 解析错误，更新冷却状态
  - `handler/proxy_anthropic.go`: 同上
  - `handler/provider.go`: 配置更新时清除冷却状态
- **无 API 变化**：对外接口保持不变
- **无数据模型变化**：冷却状态存储在内存，不持久化
- **无依赖变化**：不需要 Redis 或其他外部存储