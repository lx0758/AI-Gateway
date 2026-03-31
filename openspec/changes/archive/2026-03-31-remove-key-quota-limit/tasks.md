## 1. 后端模型修改

- [x] 1.1 移除 Key 模型中的限额字段 (`RateLimit`, `Quota`, `UsedQuota`, `UsedCount`)
- [x] 1.2 移除 KeyModel 中的相关引用（如有）— 无需修改，KeyModel 未引用这些字段

## 2. 中间件修改

- [x] 2.1 移除 `RequireAPIKey` 中的 quota 超限检查逻辑
- [x] 2.2 移除 `RequireAPIKeyForAnthropic` 中的 quota 超限检查逻辑

## 3. Handler 修改

- [ ] 3.1 移除 `CreateAPIKeyRequest` 中的 `RateLimit` 和 `Quota` 字段
- [ ] 3.2 移除 `UpdateAPIKeyRequest` 中的 `RateLimit` 和 `Quota` 字段
- [ ] 3.3 移除 `List` 方法中的 `UsedQuota`, `UsedCount`, `TotalTokens`, `AvgLatency` 字段
- [ ] 3.4 移除 `List` 方法中的统计查询逻辑
- [ ] 3.5 移除 `Create` 方法中的 `RateLimit` 和 `Quota` 赋值
- [ ] 3.6 移除 `Update` 方法中的限额更新逻辑

## 4. 验证

- [ ] 4.1 运行 `go build` 确保编译通过
- [ ] 4.2 运行现有测试确保无回归
