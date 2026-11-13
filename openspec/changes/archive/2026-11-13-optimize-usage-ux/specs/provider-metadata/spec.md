# Spec Delta

## ADDED Requirements

### Requirement: 厂商新增与编辑表单提供启用状态控制

系统 SHALL 在厂商新增与编辑对话框中提供启用状态开关，其状态与列表行内状态开关联动一致。

#### Scenario: 新增表单默认启用
- **WHEN** 管理员打开新增厂商对话框
- **THEN** 状态开关显示为启用且可调

#### Scenario: 创建时省略启用状态
- **WHEN** 管理员创建厂商时未提供 enabled 字段
- **THEN** 创建的厂商为启用状态

#### Scenario: 创建时指定禁用状态
- **WHEN** 管理员创建厂商时携带 enabled=false
- **THEN** 创建的厂商为禁用状态

#### Scenario: 编辑表单回显当前状态
- **WHEN** 管理员打开某个禁用厂商的编辑对话框
- **THEN** 状态开关显示为禁用
- **AND** 保存后该厂商保持禁用状态
