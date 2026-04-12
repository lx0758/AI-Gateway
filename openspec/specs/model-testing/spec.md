## ADDED Requirements

### Requirement: Provider 模型测试 API

系统 SHALL 提供 API 端点来测试特定的 Provider 模型。

```
POST /api/v1/providers/:id/models/:model_id/test
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

### Requirement: 虚拟模型测试 API

系统 SHALL 提供 API 端点来测试虚拟模型（别名）及其所有映射。

```
POST /api/v1/models/:id/test
```

#### Scenario: 测试具有单个映射的虚拟模型

- **WHEN** 模型有一个启用的映射
- **THEN** 系统执行该映射的测试
- **AND** 返回测试结果，包含 Provider 和模型详细信息

#### Scenario: 测试具有多个映射的虚拟模型

- **WHEN** 模型有多个启用的映射
- **THEN** 系统按权重顺序为每个映射执行测试
- **AND** 返回所有测试结果，按权重排序

#### Scenario: 测试具有禁用映射的虚拟模型

- **WHEN** 模型有禁用的映射
- **THEN** 系统跳过禁用的映射
- **AND** 仅测试启用的映射

#### Scenario: 虚拟模型不存在

- **WHEN** 模型 ID 不存在
- **THEN** 系统返回 404 错误

#### Scenario: 虚拟模型没有启用的映射

- **WHEN** 模型没有启用的映射
- **THEN** 系统返回空的测试数组

### Requirement: 测试执行行为

#### Scenario: 测试使用固定消息

- **WHEN** 执行测试时
- **THEN** 系统向模型发送消息 "Hi"
- **AND** 设置 max_tokens 为 100
- **AND** 设置 stream 为 false

#### Scenario: 测试测量延迟

- **WHEN** 执行测试时
- **THEN** 系统记录延迟（毫秒）

#### Scenario: 测试提取 Token 使用量

- **WHEN** 测试响应包含 Token 使用量
- **THEN** 系统提取 input_tokens 和 output_tokens

#### Scenario: 测试提取响应内容

- **WHEN** 测试成功
- **THEN** 系统提取响应文本内容

#### Scenario: 测试超时

- **WHEN** 测试请求超过 30 秒
- **THEN** 系统返回超时错误

### Requirement: 测试代码复用现有的 Provider 逻辑

#### Scenario: 测试使用 httptest 上下文

- **WHEN** 执行测试时
- **THEN** 系统使用 httptest.NewRecorder 和 gin.CreateTestContext 创建 gin.Context
- **AND** 调用现有的 Provider.ExecuteOpenAIRequest 或 ExecuteAnthropicRequest