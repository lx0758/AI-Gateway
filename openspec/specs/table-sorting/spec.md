## ADDED Requirements

### Requirement: 用户可以按任意可排序列对列表进行排序

系统 SHALL 为每个列表页面提供多列排序功能，用户可以通过点击表头来改变排序方式。

支持的主列表和可排序字段：
- **Models**: 名称、映射数量、Token 汇总、状态
- **Providers**: 名称、模型数量、状态
- **Keys**: 名称、模型数、工具数、资源数、提示词数、状态
- **MCPs**: 名称、类型、工具数、资源数、提示词数、状态、最后同步时间

支持的子页面列表和可排序字段：
- **Models Detail (映射列表)**: 厂商名称、实际模型、上下文窗口、权重、状态
- **Providers Detail (模型列表)**: 模型ID、名称、上下文窗口、价格、来源、状态
- **Keys Detail (models)**: 名称、映射数量、上下文窗口、状态
- **Keys Detail (tools)**: 名称、描述、状态
- **Keys Detail (resources)**: 名称、描述、URI、MIME Type、状态
- **Keys Detail (prompts)**: 名称、描述、状态
- **MCPs Detail (tools)**: 名称、描述、状态
- **MCPs Detail (resources)**: 名称、描述、URI、MIME Type、状态
- **MCPs Detail (prompts)**: 名称、描述、状态

#### Scenario: 用户点击表头排序
- **WHEN** 用户点击可排序的表头
- **THEN** 系统按该列进行升序排序，并显示排序指示器

#### Scenario: 用户再次点击表头切换排序方向
- **WHEN** 用户在已排序的列表上再次点击同一表头
- **THEN** 系统切换为降序排序

#### Scenario: 用户第三次点击表头取消排序
- **WHEN** 用户在降序排序的列表上再次点击同一表头
- **THEN** 系统恢复默认排序（按名称升序）

### Requirement: 排序偏好自动保存

系统 SHALL 自动将用户选择的排序方式保存到浏览器 localStorage，并在下次访问时自动恢复。

#### Scenario: 用户选择排序后自动保存
- **WHEN** 用户选择某个排序方式
- **THEN** 系统立即保存排序配置到 localStorage

#### Scenario: 用户再次访问页面自动恢复排序
- **WHEN** 用户重新打开列表页面
- **THEN** 系统从 localStorage 读取并应用之前保存的排序配置

#### Scenario: 无保存的排序配置时使用默认排序
- **WHEN** localStorage 中没有该页面的排序配置
- **THEN** 系统使用默认排序（按名称升序）

### Requirement: 空值排在列表末尾

系统 SHALL 将空值（null、undefined、空字符串）统一排在列表末尾，无论当前是升序还是降序排序。

#### Scenario: 升序排序时空值排在末尾
- **WHEN** 用户按某列升序排序
- **AND** 该列存在空值记录
- **THEN** 空值记录显示在列表末尾

#### Scenario: 降序排序时空值排在末尾
- **WHEN** 用户按某列降序排序
- **AND** 该列存在空值记录
- **THEN** 空值记录显示在列表末尾

### Requirement: 日期时间字段支持正确的排序

系统 SHALL 对日期时间字段进行正确的排序，而非简单的字符串排序。

#### Scenario: 按最后同步时间排序
- **WHEN** 用户在 MCPs 列表点击"最后同步"表头
- **THEN** 系统按日期时间先后排序，最近的时间排在前面（降序时）或后面（升序时）

#### Scenario: 无同步时间的记录排在末尾
- **WHEN** MCP 服务从未同步过（last_sync_at 为空）
- **THEN** 该记录在排序时始终显示在列表末尾

### Requirement: 用户可以修改模型的可用于调用状态

系统 SHALL 允许用户在厂商详情页修改模型的 `is_available` 状态。

#### Scenario: 用户切换模型可用性
- **WHEN** 用户在厂商详情页点击模型的状态开关
- **THEN** 系统更新模型的 is_available 字段并保存到数据库

#### Scenario: is_available 为 false 的模型不参与路由
- **WHEN** 模型的 is_available 被设置为 false
- **THEN** 该模型不会参与请求路由转发

### Requirement: 父对象禁用时子对象状态显示为禁用

系统 SHALL 在父对象禁用时，将子对象的状态开关/选择器显示为禁用状态，但不修改数据库中的实际状态。

#### Scenario: 厂商禁用时模型开关显示禁用
- **WHEN** 厂商的 enabled 为 false
- **THEN** 模型列表的状态开关显示为关闭且置灰
- **AND** hover 显示提示"厂商已禁用"
- **AND** 不修改数据库中模型的 is_available 值

#### Scenario: 模型禁用时映射开关显示禁用
- **WHEN** 模型的 enabled 为 false
- **THEN** 映射列表的状态开关显示为关闭且置灰
- **AND** hover 显示提示"模型已禁用"
- **AND** 不修改数据库中映射的 enabled 值

#### Scenario: 密钥禁用时权限选择器显示禁用
- **WHEN** 密钥的 enabled 为 false
- **THEN** 权限的 radio-group 显示为禁用状态
- **AND** "全部允许"按钮显示为禁用状态
- **AND** hover 显示提示"密钥已禁用"
- **AND** 不修改数据库中的权限配置

#### Scenario: MCP 服务禁用时工具/资源/提示词开关显示禁用
- **WHEN** MCP 服务的 enabled 为 false
- **THEN** tools/resources/prompts 列表的状态开关显示为关闭且置灰
- **AND** hover 显示提示"服务已禁用"
- **AND** 不修改数据库中各项的 enabled 值
