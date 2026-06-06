## Why

当前模型测试功能只能测试数据库中已存在的模型，用户无法在添加模型前验证某个模型 ID 是否可用。这导致用户需要先添加模型、再测试、如果不行再删除，操作繁琐且产生无效数据。

## What Changes

- 新增 API 端点，允许用户通过 Provider ID + 自定义模型 ID（字符串）直接测试，无需该模型在数据库中存在
- 在 Provider 详情页新增"自定义测试"按钮，点击后弹出输入框让用户输入模型 ID，提交后走测试流程并展示结果

## Capabilities

### New Capabilities
- `custom-model-testing`: 允许用户输入任意模型 ID 对指定 Provider 执行测试，无需预先在数据库中创建模型记录

### Modified Capabilities
- `model-testing`: 新增自定义模型测试场景，扩展测试 API 以支持临时模型 ID

## Impact

- 后端: `server/internal/handler/model_testing.go` 新增 handler 方法，`server/cmd/server/main.go` 注册新路由
- 前端: `web/src/views/Providers/Detail.vue` 新增自定义测试按钮和对话框
- API: 新增 `POST /api/v1/providers/:id/test-custom` 端点
