# Spec Delta

## ADDED Requirements

### Requirement: 映射对话框切换厂商清空模型选择

系统 SHALL 在映射新增/编辑对话框中，当管理员切换厂商时立即清空已选模型并重新加载该厂商的模型列表，防止陈旧的模型选择触发不匹配拒绝。

#### Scenario: 切换厂商清空已选模型
- **WHEN** 管理员在映射对话框中将厂商从 A 切换为 B
- **THEN** 已选模型被清空且模型下拉展示厂商 B 的模型列表

#### Scenario: 未选厂商时模型选择器为空
- **WHEN** 管理员清空厂商选择
- **THEN** 模型选择器为空且不加载任何模型

#### Scenario: 清空后必须重新选择模型才能保存
- **WHEN** 管理员切换厂商后未重新选择模型即提交
- **THEN** 表单提示模型为必填，不发起保存请求

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
- **AND** model_info 包含 {context_window, max_output}

#### Scenario: 同一 ProviderModel 可多次映射
- **WHEN** 管理员为 Model "gpt-4" 创建两个 ModelMapping
- **AND** 两个映射的 provider_model_id 相同
- **THEN** 两个映射都创建成功
- **AND** 可以为两个映射设置不同的权重

### Requirement: API 层排除禁用 Provider 的映射

系统 SHALL 在所有 API 响应和统计计算中，过滤掉与禁用 Provider（Provider.enabled=false）关联的 ModelMappings。

#### Scenario: Model 列表排除禁用 Provider 映射
- **WHEN** 管理员请求 GET /api/v1/models
- **AND** Model 有 ModelMappings，其中部分有 Provider.enabled=false
- **THEN** 响应 mappings 数组仅包含 Provider.enabled=true 的映射
- **AND** mapping_count 仅反映启用的 Provider 映射
- **AND** min_context_window、min_max_output 仅从启用的 Provider 映射计算

#### Scenario: Model 详情排除禁用 Provider 映射
- **WHEN** 管理员请求 GET /api/v1/models/:id
- **AND** Model 有 Provider.enabled=false 的 ModelMappings
- **THEN** 响应 mappings 数组排除这些映射
- **AND** 所有映射相关统计仅反映启用的 Provider 映射

#### Scenario: Model 映射列表排除禁用 Provider 映射
- **WHEN** 管理员请求 GET /api/v1/models/:id/mappings
- **AND** Model 有 Provider.enabled=false 的 ModelMappings
- **THEN** 响应 mappings 数组排除这些映射

#### Scenario: Key 模型列表排除禁用 Provider 映射
- **WHEN** 管理员请求 GET /api/v1/keys/:id/models
- **AND** Model 有 Provider.enabled=false 的 ModelMappings
- **THEN** 响应 mapping_count、min_context_window、min_max_output 仅从启用的 Provider 映射计算

#### Scenario: 统计计算忽略禁用 Provider 映射
- **WHEN** 系统计算 min_context_window 或 min_max_output
- **AND** ModelMapping 有 Provider.enabled=false
- **THEN** 该映射从计算中排除

#### Scenario: Model 更新响应排除禁用 Provider 映射
- **WHEN** 管理员通过 PUT /api/v1/models/:id 更新 Model
- **AND** Model 有 Provider.enabled=false 的 ModelMappings
- **THEN** 响应 mappings 数组排除这些映射

## REMOVED Requirements

### Requirement: ModelMapping 支持 Provider 关联
**Reason**: 该要求的"UI 在详情页显示模型能力"场景随 capabilities 全局移除而废弃，要求重写为仅描述 Provider 信息返回。
**Migration**: 由 ADDED 要求"ModelMapping 响应包含 Provider 信息"承接 Preload Provider 与 UI 展示场景。

### Requirement: 统计计算排除失效模型的映射
**Reason**: 该要求包含能力交集统计场景，随能力交集计算删除而废弃，要求重写为仅描述 Token 统计。
**Migration**: 由 ADDED 要求"Token 统计排除失效模型的映射"承接 Token 统计排除与空统计场景。

## ADDED Requirements

### Requirement: ModelMapping 响应包含 Provider 信息

系统 SHALL 为 ModelMapping 查询 Preload Provider 关系。

#### Scenario: 映射包含 Provider 信息
- **WHEN** API 返回 ModelMapping
- **THEN** 映射包含 provider 对象，带 {id, name, openai_base_url, anthropic_base_url}
- **AND** UI 可以显示 Provider 名称和类型标签

### Requirement: Token 统计排除失效模型的映射

系统 SHALL 在计算 min_context_window、min_max_output 时，除排除禁用 Provider 的映射外，同样排除关联模型 `availability="unavailable"`（失效）的映射，使统计与真实路由能力一致。

#### Scenario: Token 统计排除失效模型映射
- **WHEN** 系统计算 min_context_window 或 min_max_output
- **AND** 某启用映射关联的 ProviderModel `availability="unavailable"`
- **THEN** 该映射从计算中排除

#### Scenario: 全部映射指向失效模型时统计为空
- **WHEN** 模型的所有启用映射均关联 `availability="unavailable"` 的 ProviderModel
- **THEN** Token Summary 显示 "-"
