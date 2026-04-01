## 1. 数据模型变更

- [x] 1.1 修改 model/db.go Provider 结构体，删除 Type 字段，添加 OpenAIBaseURL 和 AnthropicBaseURL 字段
- [x] 1.2 编写数据库迁移逻辑，将现有 Provider 的 type 和 base_url 映射到新字段
- [ ] 1.3 测试迁移脚本，验证 OpenAI 和 Anthropic 类型的 Provider 正确迁移
- [ ] 1.4 验证 ProviderModel 和 ModelMapping 数据关联保持正确

## 2. Provider 实现类调整

- [x] 2.1 修改 provider/config.go，删除 ProviderType 字段，保留 BaseURL 和 APIKey
- [x] 2.2 修改 provider/provider.go，确保 Provider 接口不变，构造函数接受 Config
- [x] 2.3 修改 OpenAIProvider 构造函数，接受 Config 而不依赖 model.Provider
- [x] 2.4 修改 AnthropicProvider 构造函数，接受 Config 而不依赖 model.Provider
- [x] 2.5 验证 Provider 实例不持有 model.Provider 引用，避免循环依赖

## 3. Router 调整

- [x] 3.1 修改 router/router.go RouteResult 结构体，改为返回 []Provider 实例列表
- [x] 3.2 修改 router/router.go Route() 方法签名，添加 requestStyle 参数
- [x] 3.3 Router 内部增加 provider 包依赖（创建实例需要）
- [x] 3.4 实现路由优先级逻辑：按 weight 排序，优先匹配支持 requestStyle 的 Provider
- [x] 3.5 实现分组逻辑：支持同格式的 Provider 放前面，不支持的放后面（降级候选）
- [x] 3.6 实现实例创建逻辑：根据 BaseURL 和 requestStyle 创建对应的 Provider 实例
- [x] 3.7 测试路由逻辑：验证 weight 优先、格式匹配优先、降级场景
- [x] 3.8 验证循环依赖：Router → Provider → Model 不形成循环

## 4. Factory 清理

- [x] 4.1 删除 provider/factory.go 文件
- [x] 4.2 搜索并删除所有 Factory.Create 等调用
- [x] 4.3 更新相关测试文件，删除 Factory 相关测试

## 5. Handler 调整

- [x] 5.1 修改 handler/provider.go createProviderRequest 结构体，删除 type 字段，添加 openai_base_url 和 anthropic_base_url
- [x] 5.2 修改 handler/provider.go updateProviderRequest 结构体，添加 openai_base_url 和 anthropic_base_url 字段
- [x] 5.3 修改 handler/provider.go providerResponse 结构体，添加 openai_base_url 和 anthropic_base_url 字段
- [x] 5.4 修改创建和更新逻辑，验证至少一个 BaseURL 不为空
- [x] 5.5 修改 handler/proxy_openai.go，调用路由时传递 requestStyle="openai"
- [x] 5.6 修改 handler/proxy_openai.go，删除 Factory 创建逻辑，直接使用 Router 返回的实例列表第一个元素
- [x] 5.7 修改 handler/proxy_anthropic.go，调用路由时传递 requestStyle="anthropic"
- [x] 5.8 修改 handler/proxy_anthropic.go，删除 Factory 创建逻辑，直接使用 Router 返回的实例列表第一个元素
- [x] 5.9 添加空列表检查：Router 返回空列表时返回 404 错误

## 6. 前端调整

- [x] 6.1 修改前端厂商编辑表单，删除 type 下拉框
- [x] 6.2 添加 OpenAI BaseURL 输入框（可选）
- [x] 6.3 添加 Anthropic BaseURL 输入框（可选）
- [x] 6.4 添加表单验证：至少一个 BaseURL 必填
- [x] 6.5 修改厂商列表显示，展示支持的 API 风格（基于 BaseURL 是否为空）
- [x] 6.6 更新国际化文件，添加新的字段翻译

## 7. 测试和验证

- [ ] 7.1 编写单元测试：Provider 数据模型迁移
- [ ] 7.2 编写单元测试：Router.Route 带 requestStyle 参数的各种场景
- [ ] 7.3 编写单元测试：Router 返回 Provider 实例列表的正确性
- [ ] 7.4 编写集成测试：完整的请求流程（OpenAI/Anthropic 请求 → 路由 → 调用）
- [ ] 7.5 手动测试：创建同时支持两种风格的 Provider
- [ ] 7.6 手动测试：验证模型不再重复存储
- [ ] 7.7 手动测试：验证路由优先使用同格式调用
- [ ] 7.8 手动测试：验证多个 Provider 实例正确排序（支持未来负载均衡）

## 8. 文档和清理

- [x] 8.1 更新 README.md，说明新的 Provider 配置方式
- [x] 8.2 更新 API 文档，说明 Provider 创建/更新接口字段变更
- [x] 8.3 添加迁移指南，说明从旧版本升级的步骤
- [x] 8.4 标记 Breaking Change，删除 Provider.Type 和 Factory 相关代码和文档
- [ ] 8.5 清理测试文件中的 LSP 错误（如 provider_openai_compatible_test.go）