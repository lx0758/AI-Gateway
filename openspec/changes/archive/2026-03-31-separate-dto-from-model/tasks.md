## 改造计划

### 1. APIKey Handler

- [x] 在 `key.go` 末尾添加 DTO 定义
- [x] 重构 List/Create/Update 方法，使用 DTO 替代 model 作为响应

### 2. Provider Handler

- [x] 在 `provider.go` 末尾添加 DTO 定义
- [x] 重构 List/Create/Update 方法，使用 DTO 替代 model 作为响应

### 3. ModelMapping Handler

- [x] 在 `model_mapping.go` 中整理现有 DTO（如有）
- [x] 重构 List/Create 方法，使用 DTO 替代 model 作为响应

### 4. ProviderModel Handler

- [x] 在 `provider_model.go` 末尾添加 DTO 定义
- [x] 重构 List/Create 方法，使用 DTO 替代 model 作为响应

### 5. 数据库迁移

- [x] 填充 providers 表 NULL 的 created_at/updated_at
- [x] 填充 provider_models 表 NULL 的 created_at/updated_at
- [x] 填充 model_mappings 表 NULL 的 created_at
- [x] 填充 keys 表 NULL 的 created_at
- [x] 填充 key_models 表 NULL 的 created_at
- [x] 填充 usage_logs 表 NULL 的 created_at

### 6. 验证

- [x] 运行 `go build` 确保编译通过
- [ ] 手动测试各接口返回正常
