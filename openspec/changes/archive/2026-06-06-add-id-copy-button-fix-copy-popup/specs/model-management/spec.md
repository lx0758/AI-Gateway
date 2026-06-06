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
- **AND** Token Summary 列显示格式化显示（例如 "8K / 4K"）
- **AND** 鼠标悬停显示原始值（例如 "128,000 / 4,096"）
- **WHEN** 模型没有启用的映射
- **THEN** Token Summary 列显示 "-"

#### Scenario: 能力交集正确显示
- **WHEN** 模型有启用的映射
- **THEN** 系统为所有启用的映射计算能力交集
- **AND** supports_vision 仅在所有启用的映射支持 vision 时为 true
- **AND** supports_tools 仅在所有启用的映射支持 tools 时为 true
- **AND** supports_stream 仅在所有启用的映射支持 stream 时为 true
- **AND** Capabilities 列仅显示为 true 的能力标签
- **WHEN** 模型没有启用的映射
- **THEN** Capabilities 列显示 "-"

#### Scenario: Token Summary 使用一位小数格式化
- **WHEN** 系统格式化 min_context_window 和 min_max_output
- **THEN** 值 < 1000 显示为原始数字
- **AND** 值 >= 1000 显示为 "XK" 或 "X.XK"（例如 "128K"、"153.6K"）
- **AND** 值 >= 1000000 显示为 "XM" 或 "X.XM"（例如 "2M"、"1.5M"）

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
