## MODIFIED Requirements

### Requirement: 允许删除手动模型

系统 SHALL 允许删除手动创建的模型，并拒绝删除同步的模型。

#### Scenario: 删除手动模型
- **WHEN** 管理员删除 source="manual" 的模型
- **THEN** 系统从数据库移除该模型

#### Scenario: 拒绝删除同步模型
- **WHEN** 管理员尝试删除 source="sync" 的模型
- **THEN** 系统返回状态 400 的错误
- **AND** 系统 NOT 从数据库移除该模型

#### Scenario: 同步模型删除的错误消息
- **WHEN** 管理员尝试删除 source="sync" 的模型
- **THEN** 系统返回错误消息 "无法删除同步的模型，只有手动添加的模型才能删除"