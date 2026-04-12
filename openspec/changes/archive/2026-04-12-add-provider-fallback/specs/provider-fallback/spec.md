## ADDED Requirements

### Requirement: 冷却状态管理

系统 SHALL 维护 Provider+Model 组合的冷却状态，包括连续 429 计数和冷却结束时间。

冷却状态 SHALL 以 ProviderID 和 ModelName 组合为 key。

系统 SHALL 在内存中存储冷却状态，不持久化。

#### Scenario: 初始状态为无冷却

- **WHEN** 系统启动
- **THEN** 所有 Provider+Model 组合的冷却状态为空（consecutive429=0，cooldownUntil=null）

#### Scenario: 冷却状态 key 格式

- **WHEN** Provider ID=5，ModelName="gpt-4o"
- **THEN** 冷却状态 key 为 "5:gpt-4o"

### Requirement: 连续 429 触发冷却

系统 SHALL 在 Provider+Model 组合连续 3 次返回 429 后，将其加入冷却期（30 分钟）。

系统 SHALL 在请求成功时重置该组合的连续 429 计数。

冷却期内，系统 SHALL 不增加连续 429 计数（因为该组合不会被路由选中）。

#### Scenario: 第 1 次 429

- **WHEN** Provider A 的 Model M 返回 429，且该组合当前 consecutive429=0
- **THEN** 该组合的 consecutive429 变为 1，cooldownUntil 保持 null

#### Scenario: 第 2 次 429

- **WHEN** Provider A 的 Model M 返回 429，且该组合当前 consecutive429=1
- **THEN** 该组合的 consecutive429 变为 2，cooldownUntil 保持 null

#### Scenario: 第 3 次 429 触发冷却

- **WHEN** Provider A 的 Model M 返回 429，且该组合当前 consecutive429=2
- **THEN** 该组合的 consecutive429 变为 3，cooldownUntil 设置为当前时间 + 30 分钟

#### Scenario: 成功请求重置计数

- **WHEN** Provider A 的 Model M 请求成功，且该组合当前 consecutive429=2
- **THEN** 该组合的 consecutive429 重置为 0，cooldownUntil 保持 null

#### Scenario: 成功请求不改变冷却状态

- **WHEN** Provider A 的 Model M 请求成功，且该组合当前 cooldownUntil 不为 null（正在冷却）
- **THEN** 该组合的 cooldownUntil 保持不变（冷却继续），consecutive429 重置为 0

### Requirement: 路由过滤冷却中的 Provider

系统 SHALL 在路由时过滤掉处于冷却期的 Provider+Model 组合。

系统 SHALL 在所有 Provider+Model 组合都处于冷却期时，选择最早结束冷却的那个。

#### Scenario: 正常路由

- **WHEN** 用户请求 Model M，且 Provider A、B、C 都可用（无冷却）
- **THEN** 系统按 weight 排序，返回所有 Provider

#### Scenario: 过滤冷却中的 Provider

- **WHEN** 用户请求 Model M，且 Provider A 正在冷却，Provider B、C 可用
- **THEN** 系统只返回 Provider B、C（按 weight 排序）

#### Scenario: 所有 Provider 都在冷却期

- **WHEN** 用户请求 Model M，且 Provider A 冷却结束时间为 T1，Provider B 为 T2（T1 < T2）
- **THEN** 系统返回 Provider A（最早恢复）

#### Scenario: 冷却期结束自动恢复

- **WHEN** 当前时间超过 Provider A 的 cooldownUntil
- **THEN** Provider A 自动恢复可用，cooldownUntil 重置为 null

### Requirement: 非 429 错误不触发冷却

系统 SHALL 只在 429 状态码时触发冷却逻辑，其他错误状态码 SHALL NOT 影响冷却状态。

#### Scenario: 500 错误不触发冷却

- **WHEN** Provider A 的 Model M 返回 500，且该组合当前 consecutive429=1
- **THEN** 该组合的 consecutive429 保持 1，cooldownUntil 保持 null

#### Scenario: 网络错误不触发冷却

- **WHEN** Provider A 的 Model M 请求超时或网络错误
- **THEN** 该组合的冷却状态不变