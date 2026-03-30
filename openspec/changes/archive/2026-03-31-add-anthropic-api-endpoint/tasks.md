# Tasks: Add Anthropic API Endpoint

## 1. Manufacturer 接口实现

- [x] 1.1 创建 `manufacturer/manufacturer.go` - 定义 Manufacturer 接口
- [x] 1.2 创建 `manufacturer/factory.go` - Factory + ProviderType 常量
- [x] 1.3 创建 `manufacturer/manufacturer_openai_compatible.go` 
- [x] 1.4 创建 `manufacturer/manufacturer_anthropic.go`

## 2. Handler 实现

- [x] 2.1 创建 `handler/proxy_openai.go` - OpenAI 代理处理器
- [x] 2.2 创建 `handler/proxy_anthropic.go` - Anthropic 代理处理器
- [x] 2.3 实现模型权限检查
- [x] 2.4 实现 API Key 权限检查

## 3. ProviderModelHandler 改造

- [x] 3.1 重构 `ProviderModelHandler.Sync()` 使用 manufacturer.SyncModels()
- [x] 3.2 删除 `syncOpenAIModels()` 方法
- [x] 3.3 删除 `syncAnthropicModels()` 方法

## 4. 格式转换实现

- [x] 4.1 实现 Anthropic → OpenAI 转换 (OpenAICompatibleManufacturer)
- [x] 4.2 实现 OpenAI → Anthropic 转换 (AnthropicManufacturer)

## 5. 清理

- [x] 5.1 删除 `transformer/` 包
- [x] 5.2 更新 provider_model.go 使用 manufacturer.ProviderType* 常量

## 6. 路由注册

- [x] 6.1 修改 `cmd/server/main.go` - 注册 `/anthropic/v1` 路由组
- [x] 6.2 添加 Anthropic 路由组

## 7. 中间件

- [x] 7.1 修改 `middleware/auth.go` - 添加 `x-api-key` header 支持
- [x] 7.2 严格使用 x-api-key 认证（Anthropic）

## 8. 待完成任务

- [ ] 8.1 编写单元测试
- [ ] 8.2 编写集成测试
- [ ] 8.3 测试 Anthropic 请求到 OpenAI 后端
- [ ] 8.4 测试 OpenAI 请求到 Anthropic 后端
- [ ] 8.5 测试流式响应转换
- [ ] 8.6 测试错误处理
- [ ] 8.7 更新 API 文档
- [ ] 8.8 运行现有测试套件
- [ ] 8.9 提交代码审查
