# Spec Delta

## ADDED Requirements

### Requirement: 统计计算排除失效模型的映射

系统 SHALL 在计算 min_context_window、min_max_output 与能力交集时，除排除禁用 Provider 的映射外，同样排除关联模型 `availability="unavailable"`（失效）的映射，使统计与真实路由能力一致。

#### Scenario: Token 统计排除失效模型映射
- **WHEN** 系统计算 min_context_window 或 min_max_output
- **AND** 某启用映射关联的 ProviderModel `availability="unavailable"`
- **THEN** 该映射从计算中排除

#### Scenario: 能力交集排除失效模型映射
- **WHEN** 系统计算能力交集
- **AND** 某启用映射关联的 ProviderModel `availability="unavailable"`
- **THEN** 该映射从交集中排除

#### Scenario: 全部映射指向失效模型时统计为空
- **WHEN** 模型的所有启用映射均关联 `availability="unavailable"` 的 ProviderModel
- **THEN** Token Summary 显示 "-"
- **AND** 能力交集显示 "-"
