## RENAMED Requirements

FROM: `AliasMapping` → TO: `ModelMapping`
FROM: `alias_mappings` → TO: `model_mappings`

---

## MODIFIED Requirements

### Requirement: ModelMapping 引用 Provider 和 ProviderModel

系统 SHALL 将 ModelMapping 与 Provider（通过 provider_id）关联，并通过 model_id 字符串引用 ProviderModel，包含模型能力和 Token 信息。

#### Scenario: 映射使用有效 Provider
- **WHEN** 管理员创建 ModelMapping，provider_id=5，provider_model_name="gpt-4-turbo"
- **THEN** 系统验证 Provider.id=5 存在
- **AND** 系统验证 ProviderModel(provider_id=5, model_id="gpt-4-turbo") 存在
- **AND** 创建成功

#### Scenario: 映射使用无效 Provider
- **WHEN** 管理员创建 ModelMapping，provider_id=999
- **THEN** 创建失败，"Provider 未找到"错误

#### Scenario: 映射使用无效模型
- **WHEN** 管理员创建 ModelMapping，provider_id=5，provider_model_name="nonexistent"
- **THEN** 创建失败，"Provider 模型未找到"错误

#### Scenario: 映射在 API 响应中包含模型信息
- **WHEN** API 在详情页返回 ModelMapping
- **THEN** 每个映射包含 model_info 对象
- **AND** model_info 包含 {context_window, max_output, supports_vision, supports_tools, supports_stream}
- **AND** 值从 ProviderModel 表实时检索

### Requirement: ModelMapping 属于 Model

系统 SHALL 强制 ModelMapping.model_id 和 Model.id 之间的外键关系。

#### Scenario: 创建映射需要现有模型
- **WHEN** 管理员创建 ModelMapping，model_id=1
- **THEN** 系统验证 Model.id=1 存在
- **AND** 如果模型存在，创建成功
- **AND** 如果模型不存在，创建失败，"模型未找到"错误

#### Scenario: 映射与模型一起检索
- **WHEN** 路由器查询 ModelMapping
- **THEN** 系统可以 Preload Model 关系
- **AND** mapping.model.name 字段可访问

### Requirement: ModelMapping 有权重用于负载均衡

系统 SHALL 使用 ModelMapping.weight 字段用于路由优先级。

#### Scenario: 更高权重先路由
- **WHEN** Model "gpt-4" 有权重为 [10, 50, 30] 的 ModelMappings
- **THEN** 路由器按权重 DESC 返回 Providers：[50, 30, 10]

#### Scenario: 默认权重为 1
- **WHEN** 管理员创建 ModelMapping 而未指定权重
- **THEN** 系统默认设置 weight=1

#### Scenario: 通过拖拽排序更新权重
- **WHEN** 管理员在详情页将映射拖拽到位置 1
- **THEN** 系统设置 weight = total_mappings - 1
- **WHEN** 管理员将映射拖拽到最后位置
- **THEN** 系统设置 weight = 0

### Requirement: ModelMapping 可以启用或禁用

系统 SHALL 允许启用/禁用单个 ModelMapping 而不影响同级映射。

#### Scenario: 禁用的映射从路由排除
- **WHEN** ModelMapping.enabled=false
- **THEN** 路由器从 Provider 列表排除此映射

#### Scenario: 禁用的映射在 UI 中仍可见
- **WHEN** ModelMapping.enabled=false
- **THEN** 管理员可以在详情页看到映射
- **AND** 管理员可以通过状态开关重新启用映射

#### Scenario: UI 为禁用的映射显示模型信息
- **WHEN** ModelMapping.enabled=false
- **THEN** 管理员可以看到模型 Token 和能力信息
- **AND** 信息帮助管理员决定重新启用或删除

### Requirement: ModelMapping 支持 Provider 关联

系统 SHALL 为 ModelMapping 查询 Preload Provider 关系。

#### Scenario: 映射包含 Provider 信息
- **WHEN** API 返回 ModelMapping
- **THEN** 映射包含 provider 对象，带 {id, name, openai_base_url, anthropic_base_url}
- **AND** UI 可以显示 Provider 名称和类型标签

#### Scenario: UI 在详情页显示模型能力
- **WHEN** 管理员查看模型详情页映射表
- **THEN** 系统显示 Capabilities 列，显示 Vision、Tools、Stream 标签
- **AND** 标签使用颜色显示：Vision（绿色）、Tools（橙色）、Stream（蓝色）
- **AND** 标签基于 model_info 值

### Requirement: ModelMapping 支持拖拽重新排序

系统 SHALL 允许在详情页通过拖拽重新排序 ModelMappings，自动更新权重。

#### Scenario: 拖拽线性更新权重
- **WHEN** 管理员拖拽并放下映射以重新排序
- **THEN** 系统计算新权重：位置 1 = total - 1，位置 2 = total - 2，...，最后 = 0
- **AND** 系统调用 PUT `/models/:id/mappings/order` API
- **AND** API 在数据库中更新所有映射权重

#### Scenario: 拖拽 API 接收顺序数组
- **WHEN** 前端调用 PUT `/models/:id/mappings/order`
- **THEN** 请求体包含 `{ "order": [mapping_id_1, mapping_id_2, ...] }`
- **AND** 系统基于数组位置索引更新权重

#### Scenario: 拖拽保留其他属性
- **WHEN** 系统通过拖拽更新权重
- **THEN** model_id、provider_id、provider_model_name、enabled 保持不变
- **AND** 仅权重值被修改

### Requirement: API 层排除禁用 Provider 的映射

系统 SHALL 在所有 API 响应和统计计算中，过滤掉与禁用 Provider（Provider.enabled=false）关联的 ModelMappings。

#### Scenario: Model 列表排除禁用 Provider 映射
- **WHEN** 管理员请求 GET /api/v1/models
- **AND** Model 有 ModelMappings，其中部分有 Provider.enabled=false
- **THEN** 响应 mappings 数组仅包含 Provider.enabled=true 的映射
- **AND** mapping_count 仅反映启用的 Provider 映射
- **AND** min_context_window、min_max_output、capabilities 仅从启用的 Provider 映射计算

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
- **THEN** 响应 mapping_count、min_context_window、min_max_output 和 capabilities 仅从启用的 Provider 映射计算

#### Scenario: 统计计算忽略禁用 Provider 映射
- **WHEN** 系统计算 min_context_window 或 min_max_output
- **AND** ModelMapping 有 Provider.enabled=false
- **THEN** 该映射从计算中排除
- **WHEN** 系统计算能力交集
- **AND** ModelMapping 有 Provider.enabled=false
- **THEN** 该映射从交集中排除

#### Scenario: Model 更新响应排除禁用 Provider 映射
- **WHEN** 管理员通过 PUT /api/v1/models/:id 更新 Model
- **AND** Model 有 Provider.enabled=false 的 ModelMappings
- **THEN** 响应 mappings 数组排除这些映射