# Spec Delta

## MODIFIED Requirements

### Requirement: ModelMapping 有权重用于负载均衡

系统 SHALL 使用 ModelMapping.weight 字段用于路由优先级，权重由系统管理：创建时自动分配最小权重（排到末尾），顺序调整通过拖拽完成。

#### Scenario: 更高权重先路由
- **WHEN** Model "gpt-4" 有权重为 [10, 50, 30] 的 ModelMappings
- **THEN** 路由器按权重 DESC 返回 Providers：[50, 30, 10]

#### Scenario: 默认权重为 1
- **WHEN** 管理员创建 ModelMapping 而未指定权重
- **THEN** 系统分配当前最小权重减一（排到列表末尾、最低优先级），不再默认为 1

#### Scenario: 通过拖拽排序更新权重
- **WHEN** 管理员在详情页将映射拖拽到位置 1
- **THEN** 系统设置 weight = total_mappings - 1
- **WHEN** 管理员将映射拖拽到最后位置
- **THEN** 系统设置 weight = 0

#### Scenario: 创建映射自动排到末尾
- **WHEN** 管理员创建 ModelMapping 而未指定权重
- **THEN** 系统分配当前最小权重减一（不更新其他映射）
- **AND** 新映射处于列表末尾、最低优先级

## ADDED Requirements

### Requirement: 同一模型下映射权重唯一

系统 SHALL 在后台检查同一 Model 下 ModelMapping 的权重唯一性，遇到重复时将本次操作的映射权重 +1 递增取第一个空位自动避让，不阻断操作、不返回错误。

#### Scenario: 创建时权重冲突自动顺延
- **WHEN** 管理员通过 API 直接创建映射并指定与现有映射重复的 weight
- **THEN** 创建成功
- **AND** 新映射使用从指定值起 +1 递增的第一个空闲权重

#### Scenario: 更新时权重冲突自动顺延
- **WHEN** 管理员通过 API 直接更新某映射的 weight 为同模型下另一映射的现有权重
- **THEN** 更新成功
- **AND** 该映射使用从指定值起 +1 递增的第一个空闲权重
- **AND** 其他映射的权重不发生变化

#### Scenario: 连续占用时继续递增
- **WHEN** 指定权重及其后连续多个权重均已被占用
- **THEN** 系统继续递增直至找到空位
- **AND** 操作成功完成

#### Scenario: 拖拽排序不产生权重冲突
- **WHEN** 管理员通过拖拽重新排序映射
- **THEN** 排序接口为全部映射分配连续且唯一的权重
- **AND** 排序成功完成

### Requirement: 映射权重不由用户输入

系统 SHALL 在映射的新增与编辑界面中不提供权重输入控件，顺序调整的唯一入口为拖拽排序；编辑映射提交时不修改权重。

#### Scenario: 新增映射表单无权重字段
- **WHEN** 管理员打开新增映射对话框
- **THEN** 表单仅包含厂商、模型、启用状态字段
- **AND** 不显示权重输入控件

#### Scenario: 编辑映射不改变权重
- **WHEN** 管理员编辑映射（修改厂商指向或启用状态）并保存
- **THEN** 该映射的 weight 保持编辑前的值不变

#### Scenario: 编辑提交不携带权重字段
- **WHEN** 编辑映射的请求体不包含 weight 字段（前端表单无权重输入）
- **THEN** 服务端保持该映射权重不变
