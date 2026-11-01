# Spec Delta

## MODIFIED Requirements

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
