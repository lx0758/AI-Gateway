# model-mapping 规格

## Purpose

定义模型映射（ModelMapping）的行为契约：映射对 Provider 与 ProviderModel 的引用约束、权重负载均衡与拖拽排序、启用禁用控制，以及 API 层对禁用 Provider 关联映射的过滤。

## Requirements

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

系统 SHALL 使用 ModelMapping.weight 字段用于路由优先级。权重由系统管理，恒为 `0..n-1` 连续整数且唯一；映射列表的行序 = weight DESC 严格确定。

#### Scenario: 更高权重先路由
- **WHEN** Model "gpt-4" 有权重为 [10, 50, 30] 的 ModelMappings
- **THEN** 路由器按权重 DESC 返回 Providers：[50, 30, 10]

#### Scenario: 默认权重为 1
- **WHEN** 管理员创建 ModelMapping 而未指定权重
- **THEN** 系统分配 weight = 0（列表末尾、最低优先级）
- **AND** 不再分配 1 或递减序列

#### Scenario: 通过拖拽排序更新权重
- **WHEN** 管理员在详情页将映射拖拽到位置 1
- **THEN** 系统设置 weight = total_mappings - 1
- **WHEN** 管理员将映射拖拽到最后位置
- **THEN** 系统设置 weight = 0

#### Scenario: 创建映射自动排到末尾
- **WHEN** 管理员创建 ModelMapping
- **THEN** 系统将已存在的映射按当前相对顺序重编号为 `1..n`
- **AND** 新映射 weight = 0，处于列表末尾、最低路由优先级

#### Scenario: 创建时显式携带权重不生效
- **WHEN** 管理员通过 API 创建映射并携带 weight 字段（任意值）
- **THEN** 创建成功，且该映射的 weight 仍为 0
- **AND** 请求不被拒绝


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

系统 SHALL 允许在详情页通过拖拽重新排序 ModelMappings，自动更新权重，并在操作完成后以服务端返回的列表为准刷新显示。

#### Scenario: 拖拽线性更新权重
- **WHEN** 管理员拖拽并放下映射以重新排序
- **THEN** 系统计算新权重：位置 1 = total - 1，位置 2 = total - 2，...，最后 = 0
- **AND** 系统调用 PUT `/models/:id/mappings/order` API
- **AND** API 在数据库中更新所有映射权重

#### Scenario: 拖拽 API 接收顺序数组
- **WHEN** 前端调用 PUT `/models/:id/mappings/order`
- **THEN** 请求体包含 `{ "order": [mapping_id_1, mapping_id_2, ...] }`
- **AND** 系统基于数组位置索引更新权重

#### Scenario: 拖拽后重新拉取服务端列表
- **WHEN** 拖拽排序保存成功
- **THEN** 前端重新请求服务端的映射列表并以其返回顺序与权重为准
- **AND** 前端不本地推算或回写权重值

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

### Requirement: 同一模型下映射权重唯一

系统 SHALL 保证同一 Model 下 ModelMapping 的权重恒为 `0..n-1` 连续整数且唯一。新增、删除、拖拽三处操作完成后均将全部映射归一化重编号；归一化不阻断操作、不返回错误。

#### Scenario: 创建时权重冲突自动顺延
- **WHEN** 管理员通过 API 直接创建映射并指定与现有映射重复的 weight
- **THEN** 创建成功，且不发生 +1 避让顺延
- **AND** 新映射 weight = 0（显式权重被忽略）

#### Scenario: 更新时权重冲突自动顺延
- **WHEN** 管理员通过 API 直接更新某映射的 weight
- **THEN** 更新成功
- **AND** 该映射的 weight 不发生变化（编辑不修改权重）
- **AND** 其他映射的权重不发生变化

#### Scenario: 连续占用时继续递增
- **WHEN** 某模型下因历史数据出现连续占用的权重
- **THEN** 任一归一化操作完成后全部映射权重构成连续序列
- **AND** 不再需要递增搜索空位

#### Scenario: 拖拽排序不产生权重冲突
- **WHEN** 管理员通过拖拽重新排序映射
- **THEN** 排序接口为全部映射分配连续且唯一的权重
- **AND** 排序成功完成

#### Scenario: 新增后归一化为连续序列
- **WHEN** 管理员创建一条新映射
- **THEN** 系统将已存在的映射按当前相对顺序重编号为 `1..n`
- **AND** 新映射 weight = 0
- **AND** 该模型下全部映射权重构成连续且唯一的序列

#### Scenario: 删除后归一化
- **WHEN** 管理员删除列表中间的某条映射
- **THEN** 系统将剩余映射重编号为 `0..n-1`
- **AND** 相对顺序保持不变
- **AND** 不留下权重空洞

#### Scenario: 存量重复权重自愈
- **WHEN** 某模型下存在重复或非连续的历史权重
- **AND** 管理员对该模型执行新增、删除或拖拽中的任一操作
- **THEN** 该模型下全部映射权重被归一化为 `0..n-1` 连续序列

#### Scenario: 相同权重下行序仍然确定
- **WHEN** 两条映射的 weight 相同（尚未归一化的存量数据）
- **THEN** 映射列表与路由查询对其返回确定且一致的先后顺序

### Requirement: 映射权重不由用户输入

系统 SHALL 将权重作为纯内部值管理：映射的新增与编辑界面不提供权重输入控件，列表不展示权重值，顺序调整的唯一入口为拖拽排序；编辑映射提交时不修改权重。

#### Scenario: 新增映射表单无权重字段
- **WHEN** 管理员打开新增映射对话框
- **THEN** 表单仅包含厂商、模型、启用状态字段
- **AND** 不显示权重输入控件

#### Scenario: 列表不展示权重值
- **WHEN** 管理员查看模型详情页的映射列表
- **THEN** 列表不包含显示权重数值的列
- **AND** 映射顺序通过行序本身表达

#### Scenario: 编辑映射不改变权重
- **WHEN** 管理员编辑映射（修改厂商指向或启用状态）并保存
- **THEN** 该映射的 weight 保持编辑前的值不变

#### Scenario: 编辑提交不携带权重字段
- **WHEN** 编辑映射的请求体不包含 weight 字段（前端表单无权重输入）
- **THEN** 服务端保持该映射权重不变

### Requirement: 统计计算排除失效模型的映射

系统 SHALL 在计算 min_context_window、min_max_output 与能力交集时，除排除禁用 Provider 的映射外，同样排除关联模型 `availability="unavailable"`（失效）的映射，使统计与真实路由能力一致。

#### Scenario: Token 统计排除失效模型映射
- **WHEN** 系统计算 min_context_window 或 min_max_output
- **AND** 某启用映射关联的 ProviderModel `availability="unavailable"`
- **THEN** 该映射从计算中排除

#### Scenario: 能力交集排除失效模型映射
- **WHEN** 系统计算能力交集
- **AND** 某启用映射关联的 ProviderModel `availability="unavailable"`
- **THEN** 该映射从交集中排除

#### Scenario: 全部映射指向失效模型时统计为空
- **WHEN** 模型的所有启用映射均关联 `availability="unavailable"` 的 ProviderModel
- **THEN** Token Summary 显示 "-"
- **AND** 能力交集显示 "-"
