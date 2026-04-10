## Why

当前系统缺少模型可用性验证功能。用户配置 Provider 和 Model 后，无法快速验证配置是否正确、模型是否能正常响应。现有的 Provider test 接口是一个未实现的存根，而 Model（虚拟模型）层面完全没有测试能力。

用户需要一种方式来：
- 验证厂商模型配置是否正确（API Key、BaseURL、模型 ID）
- 验证虚拟模型的路由链路是否通畅（Alias → Mapping → Provider → ProviderModel）
- 快速定位配置问题

## What Changes

- 新增厂商模型测试 API：`POST /api/v1/providers/:id/models/:model_id/test`
- 新增虚拟模型测试 API：`POST /api/v1/models/:id/test`
- Provider Detail 页面增加模型测试按钮和结果展示
- Model Detail 页面增加映射测试按钮和结果展示
- 测试功能复用现有 `provider.ExecuteOpenAIRequest` / `ExecuteAnthropicRequest` 逻辑

## Capabilities

### New Capabilities

- `model-testing`: 模型测试能力，包括厂商模型测试和虚拟模型测试

### Modified Capabilities

无

## Impact

- 后端新增 `handler/model_test.go`
- 前端修改 `views/Providers/Detail.vue` 和 `views/Models/Detail.vue`
- 新增 i18n 翻译条目
