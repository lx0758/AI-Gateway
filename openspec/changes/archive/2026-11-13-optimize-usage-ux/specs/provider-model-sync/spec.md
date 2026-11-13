# Spec Delta

## MODIFIED Requirements

### Requirement: 模型可用性为三态状态

系统 SHALL 以三态枚举 `availability` 表达模型的可用性验证状态：`unknown`（未知，从未被同步验证）、`available`（可用，同步确认在厂商列表）、`unavailable`（失效，同步确认厂商下线）。手动新建的模型初始为 `unknown`。

#### Scenario: 手动新建模型初始为未知
- **WHEN** 管理员手动添加一个厂商模型
- **THEN** 该模型 `availability="unknown"`
- **AND** 控制台显示"未知"状态标签

#### Scenario: 同步创建的模型为可用
- **WHEN** 同步在厂商下创建新模型（出现在返回列表中）
- **THEN** 该模型 `availability="available"`
- **AND** 控制台显示"可用"状态标签

#### Scenario: 未知与可用均参与路由
- **WHEN** 某映射关联 `availability="unknown"` 或 `availability="available"` 的模型
- **THEN** 路由器正常参与选路

#### Scenario: 失效退出路由与统计
- **WHEN** 某模型 `availability="unavailable"`
- **THEN** 指向它的映射不参与路由
- **AND** 该映射在详情接口中显示为强制禁用（`forced_disabled`）
- **AND** 该模型的上下文窗口不计入聚合统计

### Requirement: 同步合并保护本地非空信息

系统 SHALL 在同步更新已存在于返回列表中的模型时逐字段判断：同步值为空字符串或零值的字段保留本地现值，仅当同步值有效时才覆盖本地值。

#### Scenario: 同步返回空字段时保留本地信息
- **WHEN** 同步返回的模型已存在于数据库
- **AND** 同步返回的 `display_name` 为空、`context_window` 为 0
- **THEN** 数据库中该模型的 `display_name`、`context_window` 保持同步前的本地值不变

#### Scenario: 同步返回有效字段时刷新本地信息
- **WHEN** 同步返回的模型已存在于数据库
- **AND** 同步返回的 `context_window=131072`
- **THEN** 数据库中该模型的 `context_window` 更新为 131072

#### Scenario: 逐字段独立判断
- **WHEN** 同步返回的模型中 `owned_by` 有效但 `max_output` 为 0
- **THEN** `owned_by` 被刷新
- **AND** `max_output` 保留本地值

## REMOVED Requirements

### Requirement: 同步结果只反映厂商真实返回的信息
**Reason**: 该要求包含"厂商未返回能力信息时不伪造"场景，随 capabilities 全局移除而废弃，要求重写为仅描述显示名不伪造。
**Migration**: 由 ADDED 要求"同步不伪造厂商未返回的显示名"承接显示名不伪造与合并保留场景。

## ADDED Requirements

### Requirement: 同步不伪造厂商未返回的显示名

系统 SHALL 使同步器只填充厂商接口实际返回的字段值，不得为未返回的字段伪造内容（如用模型 ID 冒充显示名）。

#### Scenario: 厂商未返回显示名时不伪造
- **WHEN** 厂商 `/models` 接口的条目未提供独立显示名
- **THEN** 同步结果的 `display_name` 为空
- **AND** 合并时保留本地已有的 `display_name`
