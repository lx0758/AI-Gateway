# manual-model-protection 规格

## Purpose

定义厂商模型的删除保护策略：区分可由管理员自主删除的手动模型与受自动同步管理的模型，避免已删除的同步模型在下次同步后回归，也防止管理员误删由同步维护的模型数据。

## Requirements

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
