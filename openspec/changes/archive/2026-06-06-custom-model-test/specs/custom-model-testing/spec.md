## ADDED Requirements

### Requirement: 自定义模型测试 API

系统 SHALL 提供 API 端点，允许用户使用自定义模型 ID 对指定 Provider 执行测试，无需该模型在数据库中存在。

```
POST /api/v1/providers/:id/test-custom
```

请求体:
```json
{
  "model_id": "gpt-4-turbo"
}
```

#### Scenario: 使用自定义模型 ID 测试 OpenAI 协议

- **WHEN** Provider 配置了 OpenAIBaseURL 且用户提交 model_id
- **THEN** 系统使用 OpenAI 协议以该 model_id 执行测试
- **AND** 返回测试结果，包含延迟、Token 数、响应内容

#### Scenario: 使用自定义模型 ID 测试 Anthropic 协议

- **WHEN** Provider 配置了 AnthropicBaseURL 且用户提交 model_id
- **THEN** 系统使用 Anthropic 协议以该 model_id 执行测试
- **AND** 返回测试结果，包含延迟、Token 数、响应内容

#### Scenario: 同时测试两种协议

- **WHEN** Provider 同时配置了 OpenAIBaseURL 和 AnthropicBaseURL 且用户提交 model_id
- **THEN** 系统执行两次测试，每个协议一次
- **AND** 返回两个测试结果

#### Scenario: Provider 不存在

- **WHEN** Provider ID 不存在
- **THEN** 系统返回 404 错误

#### Scenario: model_id 为空

- **WHEN** 请求体中 model_id 为空字符串
- **THEN** 系统返回 400 错误

#### Scenario: 测试失败返回错误

- **WHEN** 测试请求失败（连接错误、超时、API 错误、模型不存在）
- **THEN** 系统返回 success=false 并附带错误消息

### Requirement: 自定义模型测试 UI

系统 SHALL 在 Provider 详情页提供自定义模型测试入口。

#### Scenario: 点击自定义测试按钮

- **WHEN** 用户在 Provider 详情页点击"自定义测试"按钮
- **THEN** 系统弹出输入框，提示用户输入模型 ID

#### Scenario: 提交自定义模型 ID 进行测试

- **WHEN** 用户输入模型 ID 并确认
- **THEN** 系统调用自定义测试 API
- **AND** 在测试结果对话框中展示测试结果

#### Scenario: 取消自定义测试输入

- **WHEN** 用户取消输入
- **THEN** 系统关闭输入框，不执行任何操作

#### Scenario: 测试结果展示

- **WHEN** 自定义测试完成
- **THEN** 系统以与现有模型测试相同的格式展示结果（协议、成功/失败、延迟、Token、响应内容、错误信息）
