## MODIFIED Requirements

### Requirement: ModelMapping 引用 Provider 和 ProviderModel

系统 SHALL 将 ModelMapping 与 Provider（通过 provider_id）和 ProviderModel（通过 provider_model_id 外键）关联，确保数据一致性。

#### Scenario: 映射使用有效 ProviderModel
- **WHEN** 管理员创建 ModelMapping，provider_id=5，provider_model_id=100
- **THEN** 系统验证 ProviderModel.id=100 存在
- **AND** 系统验证 ProviderModel.provider_id=5 匹配
- **AND** 创建成功

#### Scenario: 映射使用无效 ProviderModel
- **WHEN** 管理员创建 ModelMapping，provider_model_id=999
- **THEN** 创建失败，"Provider 模型未找到"错误

#### Scenario: 映射 ProviderModel 的 Provider 不匹配
- **WHEN** 管理员创建 ModelMapping，provider_id=5，provider_model_id=100
- **AND** ProviderModel.id=100 的 provider_id=3
- **THEN** 创建失败，"Provider 不匹配"错误

#### Scenario: 删除 ProviderModel 级联删除映射
- **WHEN** 管理员删除 ProviderModel.id=100
- **AND** 存在 ModelMapping.provider_model_id=100
- **THEN** 系统自动删除所有关联的 ModelMapping
- **AND** 不留下孤立的映射

#### Scenario: 修改 ProviderModel.ModelID 不影响映射
- **WHEN** 管理员将 ProviderModel.ModelID 从 "gpt-4" 改为 "gpt-4-turbo"
- **AND** 存在 ModelMapping.provider_model_id 关联此 ProviderModel
- **THEN** ModelMapping 关联仍然有效
- **AND** 路由查询正常工作

#### Scenario: 映射在 API 响应中包含模型信息
- **WHEN** API 在详情页返回 ModelMapping
- **THEN** 每个映射包含 provider_model_id
- **AND** 每个映射包含 provider_model_name（从关联的 ProviderModel.ModelID 获取）
- **AND** 每个映射包含 model_info 对象
- **AND** model_info 包含 {context_window, max_output, supports_vision, supports_tools, supports_stream}

#### Scenario: 同一 ProviderModel 可多次映射
- **WHEN** 管理员为 Model "gpt-4" 创建两个 ModelMapping
- **AND** 两个映射的 provider_model_id 相同
- **THEN** 两个映射都创建成功
- **AND** 可以为两个映射设置不同的权重

## ADDED Requirements

### Requirement: ProviderModel.ModelID 可编辑

系统 SHALL 允许管理员修改 ProviderModel.ModelID 字段。

#### Scenario: 修改 ModelID
- **WHEN** 管理员通过 PUT /api/v1/providers/:id/models/:mid 请求
- **AND** 请求体包含 model_id 字段
- **THEN** 系统更新 ProviderModel.ModelID
- **AND** 关联的 ModelMapping 自动生效（使用外键，不受影响）

#### Scenario: ModelID 唯一性约束
- **WHEN** 管理员将 ProviderModel.ModelID 改为已存在的值
- **AND** 同一 Provider 下已有相同 ModelID
- **THEN** 更新失败，"ModelID 已存在"错误
