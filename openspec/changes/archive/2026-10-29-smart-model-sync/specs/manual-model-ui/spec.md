# Spec Delta

## MODIFIED Requirements

### Requirement: 显示模型来源

系统 SHALL 在列表中以三态可用性标签（未知 / 可用 / 失效）指示每个模型的状态，替代原手动/同步来源标签。

#### Scenario: 显示来源标签
- **WHEN** 显示模型列表时
- **THEN** 每个模型显示标签，指示"未知"（unknown）、"可用"（available）或"失效"（unavailable）
- **AND** 标签颜色区分：未知为中性色、可用为成功色、失效为危险色
