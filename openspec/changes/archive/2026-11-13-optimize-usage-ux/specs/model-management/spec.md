# Spec Delta

## REMOVED Requirements

### Requirement: Model 列表以扁平表格布局显示
**Reason**: 该要求包含"能力交集正确显示"场景，随 capabilities 全局移除而废弃，需重写为不含能力列的版本。
**Migration**: 由 ADDED 要求"Model 列表以扁平表格显示 Token 摘要"承接扁平表格布局、Token 摘要与单位格式化场景。

## ADDED Requirements

### Requirement: Model 列表以扁平表格显示 Token 摘要

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

#### Scenario: Token Summary 使用 1024-based 单位格式化
- **WHEN** 系统格式化 min_context_window 和 min_max_output
- **THEN** 值 < 1024 显示为原始数字
- **AND** 值 >= 1024 且 < 1048576 显示为 "XK" 或 "X.XK"（例如 "128K"、"153.6K"，1K = 1024）
- **AND** 值 >= 1048576 显示为 "XM" 或 "X.XM"（例如 "2M"、"1.5M"，1M = 1048576）
