## MODIFIED Requirements

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
