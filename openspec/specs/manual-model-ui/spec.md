## ADDED Requirements

### Requirement: 手动模型管理 UI

系统 SHALL 提供用户界面用于手动管理 Provider 模型。

#### Scenario: 添加模型按钮可见
- **WHEN** 管理员查看 Provider 详情页
- **THEN** 系统在模型列表上方显示"添加模型"按钮

#### Scenario: 添加模型对话框
- **WHEN** 管理员点击"添加模型"按钮
- **THEN** 系统显示表单对话框，包含字段：model_id、display_name、context_window、max_output、supports_vision、supports_tools、supports_stream

#### Scenario: 创建手动模型
- **WHEN** 管理员填写表单并提交
- **THEN** 系统创建新的 Provider 模型，source="manual"
- **AND** 系统在列表中显示新模型

### Requirement: 编辑手动模型

系统 SHALL 允许编辑手动创建的模型。

#### Scenario: 手动模型的编辑按钮
- **WHEN** 模型列表包含 source="manual" 的模型
- **THEN** 系统为该模型显示编辑按钮

#### Scenario: 同步模型隐藏编辑按钮
- **WHEN** 模型列表包含 source="sync" 的模型
- **THEN** 系统不显示该模型的编辑按钮

#### Scenario: 更新手动模型
- **WHEN** 管理员编辑手动模型并提交
- **THEN** 系统更新模型配置

### Requirement: 删除手动模型 UI

系统 SHALL 提供 UI 来删除手动创建的模型。

#### Scenario: 手动模型的删除按钮
- **WHEN** 模型列表包含 source="manual" 的模型
- **THEN** 系统为该模型显示删除按钮

#### Scenario: 同步模型隐藏删除按钮
- **WHEN** 模型列表包含 source="sync" 的模型
- **THEN** 系统不显示该模型的删除按钮

#### Scenario: 删除前确认
- **WHEN** 管理员点击删除按钮
- **THEN** 系统显示确认对话框

### Requirement: 显示模型来源

系统 SHALL 在列表中指示每个模型的来源。

#### Scenario: 显示来源标签
- **WHEN** 显示模型列表时
- **THEN** 每个模型显示标签，指示"手动"或"同步"