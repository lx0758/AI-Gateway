## 1. 后端 API

- [x] 1.1 在 `model_testing.go` 中新增 `TestCustomModel` handler 方法：接收 Provider ID + model_id 请求体，构造临时 ProviderModel，复用 `executeTest` 执行测试
- [x] 1.2 在 `server/cmd/server/main.go` 中注册路由 `POST /providers/:id/test-custom`

## 2. 前端 UI

- [x] 2.1 在 `Providers/Detail.vue` 的 actions 区域新增"自定义测试"按钮
- [x] 2.2 实现 `testCustomModel` 函数：使用 `ElMessageBox.prompt` 获取模型 ID，调用 `POST /providers/:id/test-custom`，复用现有测试结果对话框展示结果
- [x] 2.3 在 i18n 文件中添加自定义测试相关的翻译文案
