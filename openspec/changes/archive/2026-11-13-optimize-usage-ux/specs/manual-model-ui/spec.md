# Spec Delta

## MODIFIED Requirements

### Requirement: 手动模型管理 UI

系统 SHALL 提供用户界面用于手动管理 Provider 模型。

#### Scenario: 添加模型按钮可见
- **WHEN** 管理员查看 Provider 详情页
- **THEN** 系统在模型列表上方显示"添加模型"按钮

#### Scenario: 添加模型对话框
- **WHEN** 管理员点击"添加模型"按钮
- **THEN** 系统显示表单对话框，包含字段：model_id、display_name、context_window、max_output

#### Scenario: 创建手动模型
- **WHEN** 管理员填写表单并提交
- **THEN** 系统创建新的 Provider 模型，source="manual"
- **AND** 系统在列表中显示新模型
