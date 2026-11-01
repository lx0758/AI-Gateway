# Spec Delta

## ADDED Requirements

### Requirement: 模型 ID 在列表中完整可读

系统 SHALL 在厂商详情页的模型列表中完整展示模型 ID，不因列宽固定而截断。

#### Scenario: 长模型 ID 不被截断
- **WHEN** 模型列表中某条记录的模型 ID 长度超出常规宽度
- **THEN** 该模型 ID 在列表单元格内完整显示

#### Scenario: 模型 ID 保留复制入口
- **WHEN** 管理员查看模型列表中的模型 ID
- **THEN** 模型 ID 旁仍显示复制按钮
- **AND** 复制按钮与 ID 文本不因列宽调整而重叠
