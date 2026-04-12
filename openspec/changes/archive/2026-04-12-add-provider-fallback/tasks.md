## 1. 基础设施

- [x] 1.1 创建 `router/cooldown.go`，实现冷却状态管理结构和方法
  - CooldownState 结构体：Consecutive429, CooldownUntil, Last429Time
  - CooldownManager：内存 map + mutex + 计数/冷却/重置方法
  - IsCooldown()：判断是否在冷却期
  - Record429()：记录 429（5秒缓冲），触发冷却逻辑
  - RecordSuccess()：重置计数
  - GetEarliestCooldownEnd()：获取最早恢复时间
  - ClearCooldown()：清除单个冷却状态
  - ClearAllForProvider()：清除 Provider 下所有冷却状态

- [x] 1.2 创建 `router/route_result.go`，剥离 RouteResult 结构体

- [x] 1.3 创建 `provider/provider_error.go`，定义 ProviderError 结构体
  - StatusCode int
  - Message string
  - 实现 error 接口
  - IsRateLimitError() 辅助函数

## 2. Provider 层改动

- [x] 2.1 修改 `provider/provider_openai.go`，返回结构化错误
  - 将 `fmt.Errorf("%d - %s", ...)` 改为 `&ProviderError{StatusCode, Message}`
  - 保持响应写入逻辑不变（当前请求仍返回给客户端）

- [x] 2.2 修改 `provider/provider_anthropic.go`，返回结构化错误
  - 同上

## 3. Router 层改动

- [x] 3.1 修改 `router/router.go`
  - Router 单例化（globalRouter）
  - Route() 返回 `*RouteResult`（单个最优结果）
  - 过滤掉冷却中的 Provider+ProviderModel 组合
  - 所有 Provider 都冷却时，返回最早恢复的
  - 新增全局函数：RecordRateLimit(), RecordSuccess(), ClearAllCooldownsForProvider()

## 4. Handler 层改动

- [x] 4.1 修改 `handler/proxy_openai.go` 的 ChatCompletions()
  - 调用 `router.RecordRateLimit()` 和 `router.RecordSuccess()`

- [x] 4.2 修改 `handler/proxy_anthropic.go` 的 Messages()
  - 同上

- [x] 4.3 修改 `handler/provider.go` 的 Update()
  - Provider 启用/APIKey/BaseURL 更新时调用 `router.ClearAllCooldownsForProvider()`

## 5. 测试验证

- [ ] 5.1 手动测试冷却触发逻辑
  - 模拟连续 3 次 429（间隔 ≥5秒），验证冷却状态变化
  - 验证路由过滤生效
  - 验证 5 秒缓冲生效（瞬间多次 429 只计一次）

- [ ] 5.2 手动测试冷却恢复逻辑
  - 等待冷却期结束，验证路由恢复
  - 测试成功请求重置计数
  - 测试配置更新清除冷却