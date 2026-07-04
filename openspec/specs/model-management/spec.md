## RENAMED Requirements

FROM: `Alias` → TO: `Model`
FROM: `aliases` → TO: `models`

---

## MODIFIED Requirements

### Requirement: Model 是唯一标识符

系统 SHALL 对 Model.name 字段强制唯一性，确保每个模型代表 API 调用的不同模型名称。

#### Scenario: 创建唯一模型
- **WHEN** 管理员创建模型 "gpt-4"
- **THEN** 系统创建 Model 记录，name="gpt-4"
- **AND** 后续创建模型 "gpt-4" 因重复错误而失败

#### Scenario: Model 用作 API 模型名称
- **WHEN** 客户端调用 API，model="gpt-4"
- **THEN** 路由器找到 name="gpt-4" 的 Model
- **AND** 路由器检索关联的 ModelMappings 用于路由

### Requirement: Model 可以启用或禁用

系统 SHALL 允许启用/禁用 Model，影响所有关联映射的可用性。

#### Scenario: 禁用的模型拒绝路由
- **WHEN** Model.enabled=false
- **AND** 客户端调用 API，model="gpt-4"
- **THEN** 路由器不返回 Providers
- **AND** Handler 返回"模型未找到"错误

#### Scenario: 启用的模型允许路由
- **WHEN** Model.enabled=true
- **AND** 客户端调用 API，model="gpt-4"
- **THEN** 路由器检索 ModelMappings
- **AND** 路由正常进行

### Requirement: Model 有零个或多个 ModelMappings

系统 SHALL 支持 Model 和 ModelMapping 之间的一对多关系。

#### Scenario: 模型有多个映射
- **WHEN** Model "gpt-4" 有 3 个 ModelMappings
- **THEN** 路由器检索所有 3 个映射用于路由
- **AND** 映射按权重 DESC 排序

#### Scenario: 模型没有映射
- **WHEN** Model "new-model" 没有 ModelMappings
- **THEN** 路由器不返回 Providers
- **AND** Handler 返回"模型未找到"错误

### Requirement: 删除 Model 级联到 ModelMappings

系统 SHALL 在删除 Model 时级联删除所有 ModelMappings。

#### Scenario: 删除模型移除映射
- **WHEN** 管理员删除 Model "gpt-4"
- **THEN** 系统删除所有 model_id=gpt-4.id 的 ModelMappings
- **AND** 不留下孤立的映射

### Requirement: Model 列表以扁平表格布局显示

系统 SHALL 以扁平表格布局（而非折叠面板）显示模型，与 Providers 页面风格一致。

#### Scenario: Model 列表以扁平表格显示
- **WHEN** 管理员查看 `/models` 页面
- **THEN** 系统显示表格，列：Select、Name、Mapping Count、Token Summary、Status、Actions
- **AND** 每行代表一个 Model
- **AND** 不使用折叠面板

#### Scenario: Token Summary 显示最小值并附带 Tooltip
- **WHEN** 模型有启用的映射
- **THEN** 系统从启用的映射计算 min_context_window
- **AND** 系统从启用的映射计算 min_max_output
- **AND** Token Summary 列显示格式化显示（例如 "125K / 4K"，使用 1024-based 单位）
- **AND** 鼠标悬停显示原始值（例如 "128,000 / 4,096"）
- **WHEN** 模型没有启用的映射
- **THEN** Token Summary 列显示 "-"

#### Scenario: 能力交集正确显示
- **WHEN** 模型有启用的映射
- **THEN** 系统为所有启用的映射计算能力交集
- **AND** capabilities 列显示交集后的能力标签组
- **AND** 全部未选时显示 "None"
- **WHEN** 模型没有启用的映射
- **THEN** Capabilities 列显示 "-"

#### Scenario: Token Summary 使用 1024-based 单位格式化
- **WHEN** 系统格式化 min_context_window 和 min_max_output
- **THEN** 值 < 1024 显示为原始数字
- **AND** 值 >= 1024 且 < 1048576 显示为 "XK" 或 "X.XK"（例如 "128K"、"153.6K"，1K = 1024）
- **AND** 值 >= 1048576 显示为 "XM" 或 "X.XM"（例如 "2M"、"1.5M"，1M = 1048576）

### Requirement: Model 列表支持批量操作

系统 SHALL 允许批量选择和批量删除模型。

#### Scenario: 选择多个模型
- **WHEN** 管理员在模型表中点击复选框
- **THEN** 系统跟踪选中的模型 ID
- **AND** "批量删除"按钮显示计数（例如 "批量删除 (3)"）

#### Scenario: 批量删除模型
- **WHEN** 管理员在选中模型后点击"批量删除"
- **THEN** 系统显示确认对话框
- **AND** 确认后，系统删除所有选中的模型及其映射
- **AND** 系统刷新模型表

#### Scenario: 无选择时批量删除禁用
- **WHEN** 未选中模型
- **THEN** "批量删除"按钮禁用

### Requirement: Model 列表提供详情页导航

系统 SHALL 提供"详情"按钮导航到模型详情页用于映射管理。

#### Scenario: 导航到模型详情页
- **WHEN** 管理员在模型行点击"详情"按钮
- **THEN** 系统导航到 `/models/:id` 页面
- **AND** 详情页显示完整映射信息

#### Scenario: 所有模型的详情按钮可见
- **WHEN** 显示模型表时
- **THEN** 每个模型行在 Actions 列有"详情"按钮
- **AND** "详情"按钮始终启用，无论映射数量

### Requirement: 厂商模型编辑表单上下文窗口支持灵活输入

系统 SHALL 允许管理员在编辑 ProviderModel 时使用灵活格式输入上下文窗口大小，保存时自动转换为数字。

#### Scenario: 输入纯数字
- **WHEN** 管理员输入 `128000`
- **THEN** 系统保存 context_window = 128000

#### Scenario: 输入带 K/k 后缀
- **WHEN** 管理员输入 `128k` 或 `128K`
- **THEN** 系统保存 context_window = 131072（128 × 1024）

#### Scenario: 输入带 M/m 后缀
- **WHEN** 管理员输入 `1m` 或 `1M`
- **THEN** 系统保存 context_window = 1048576（1 × 1024 × 1024）

#### Scenario: 输入带 B/b 后缀
- **WHEN** 管理员输入 `1b` 或 `1B`
- **THEN** 系统保存 context_window = 1073741824（1 × 1024 × 1024 × 1024）

#### Scenario: 输入无效格式提示错误
- **WHEN** 管理员输入 `abc` 或 `123xyz`
- **THEN** 系统提示"请输入有效格式（数字、或带 K/M/B 单位）"
- **AND** 不提交表单

#### Scenario: 编辑时回显格式化值
- **WHEN** 管理员打开已有记录的编辑表单
- **AND** 已有记录的 context_window = 131072
- **THEN** 输入框显示 `128K`

#### Scenario: 列表和详情展示使用 1024-based 单位
- **WHEN** 管理员查看模型列表或详情
- **AND** context_window = 131072
- **THEN** 显示 `128K`（1024-based）

### Requirement: 厂商模型页面的 model_id 列支持复制

系统 SHALL 在厂商模型页面（Providers/Detail.vue）的 model_id 列提供复制按钮，且点击复制按钮不会触发行点击事件。

#### Scenario: model_id 列显示复制按钮
- **WHEN** 管理员查看厂商模型列表
- **THEN** model_id 列在每个模型 ID 文本旁显示一个复制图标按钮

#### Scenario: 点击复制按钮复制 model_id
- **WHEN** 管理员点击 model_id 旁边的复制按钮
- **THEN** 系统将 model_id 文本复制到剪贴板
- **AND** 显示成功提示
- **AND** 不触发详情对话框弹出

#### Scenario: 点击 model_id 文本仍可触发行点击
- **WHEN** 管理员点击 model_id 文本区域（非复制按钮）
- **THEN** 系统正常触发 `showModelDetail` 打开详情对话框