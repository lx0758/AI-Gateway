## MODIFIED Requirements

### Requirement: Provider 模型测试 API

系统 SHALL 提供 API 端点来测试特定的 Provider 模型，同时支持对数据库中不存在的模型进行自定义测试。

```
POST /api/v1/providers/:id/models/:model_id/test
POST /api/v1/providers/:id/test-custom
```

#### Scenario: 使用 OpenAI 协议测试 Provider 模型

- **WHEN** Provider 配置了 OpenAIBaseURL
- **THEN** 系统使用 OpenAI 协议执行测试
- **AND** 返回测试结果，包含延迟、Token 数、响应内容

#### Scenario: 使用 Anthropic 协议测试 Provider 模型

- **WHEN** Provider 配置了 AnthropicBaseURL
- **THEN** 系统使用 Anthropic 协议执行测试
- **AND** 返回测试结果，包含延迟、Token 数、响应内容

#### Scenario: 同时测试两种协议的 Provider 模型

- **WHEN** Provider 同时配置了 OpenAIBaseURL 和 AnthropicBaseURL
- **THEN** 系统执行两次测试，每个协议一次
- **AND** 返回两个测试结果

#### Scenario: 测试失败返回错误

- **WHEN** 测试请求失败（连接错误、超时、API 错误）
- **THEN** 系统返回 success=false 并附带错误消息

#### Scenario: Provider 不存在

- **WHEN** Provider ID 不存在
- **THEN** 系统返回 404 错误

#### Scenario: Provider 模型不存在

- **WHEN** 该 Provider 的模型 ID 不存在
- **THEN** 系统返回 404 错误
