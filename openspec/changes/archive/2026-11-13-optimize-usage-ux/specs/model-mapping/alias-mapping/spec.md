# Spec Delta

## MODIFIED Requirements

### Requirement: AliasMapping 引用 Provider 和 ProviderModel

系统 SHALL 将 AliasMapping 与 Provider（通过 provider_id）关联，并通过 model_id 字符串引用 ProviderModel，包含 Token 信息。

#### Scenario: 映射使用有效 Provider
- **WHEN** 管理员创建 AliasMapping，provider_id=5，provider_model_name="gpt-4-turbo"
- **THEN** 系统验证 Provider.id=5 存在
- **AND** 系统验证 ProviderModel(provider_id=5, model_id="gpt-4-turbo") 存在
- **AND** 创建成功

#### Scenario: 映射使用无效 Provider
- **WHEN** 管理员创建 AliasMapping，provider_id=999
- **THEN** 创建失败，"Provider 未找到"错误

#### Scenario: 映射使用无效模型
- **WHEN** 管理员创建 AliasMapping，provider_id=5，provider_model_name="nonexistent"
- **THEN** 创建失败，"Provider 模型未找到"错误

#### Scenario: 映射在 API 响应中包含模型信息
- **WHEN** API 在详情页返回 AliasMapping
- **THEN** 每个映射包含 model_info 对象
- **AND** model_info 包含 {context_window, max_output}
- **AND** 值从 ProviderModel 表实时检索

## REMOVED Requirements

### Requirement: AliasMapping 支持 Provider 关联
**Reason**: 该要求的"UI 在详情页显示模型能力"场景随 capabilities 全局移除而废弃，要求重写为仅描述 Provider 信息返回。
**Migration**: 由 ADDED 要求"AliasMapping 响应包含 Provider 信息"承接 Preload Provider 与 UI 展示场景。

## ADDED Requirements

### Requirement: AliasMapping 响应包含 Provider 信息

系统 SHALL 为 AliasMapping 查询 Preload Provider 关系。

#### Scenario: 映射包含 Provider 信息
- **WHEN** API 返回 AliasMapping
- **THEN** 映射包含 provider 对象，带 {id, name, openai_base_url, anthropic_base_url}
- **AND** UI 可以显示 Provider 名称和类型标签
