# provider-model-sync 规格

## Purpose

定义厂商模型可用性三态（未知/可用/失效）的状态机与同步的合并策略：同步如何推进模型可用性状态、保护本地非空信息、识别厂商下线的模型，以及三态对路由与统计的含义。

## Requirements

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
- **AND** 该模型的上下文窗口与能力不计入聚合统计
### Requirement: 同步推进可用性状态

系统 SHALL 在同步成功且返回列表非空时推进模型状态：出现在列表中的模型置为 `available`；此前为 `available` 但不在列表中的模型置为 `unavailable`；`unknown` 模型不在列表时保持 `unknown`（不参与失效判断）。返回列表为空时不推进任何状态。

#### Scenario: 不在同步列表中的可用模型被标记失效
- **WHEN** 同步返回模型列表 [A, B]
- **AND** 数据库中该厂商存在状态为 `available` 的模型 [A, B, C]
- **THEN** 模型 C 状态变为 `unavailable`
- **AND** 模型 A、B 保持 `available`

#### Scenario: 未知模型不参与失效判断
- **WHEN** 同步返回列表非空且不包含某 `unknown` 模型（手动添加）
- **THEN** 该模型保持 `unknown`
- **AND** 不被标记为 `unavailable`

#### Scenario: 失效模型恢复上架后自动可用
- **WHEN** 模型 C 曾被标记 `unavailable`
- **AND** 后续同步返回列表中再次包含 C
- **THEN** 模型 C 恢复 `available`

#### Scenario: 手动模型被同步接管
- **WHEN** 手动添加的模型（`unknown`）出现在同步返回列表中
- **THEN** 该模型状态推进为 `available`
- **AND** 后续同步按合并策略管理其字段与状态

#### Scenario: 空列表不改变状态
- **WHEN** 同步接口返回空模型列表
- **AND** 数据库中该厂商存在 `available` 与 `unknown` 模型
- **THEN** 所有模型保持原状态
- **AND** 同步响应中失效计数为 0
### Requirement: 同步合并保护本地非空信息

系统 SHALL 在同步更新已存在于返回列表中的模型时逐字段判断：同步值为空字符串或零值的字段保留本地现值，仅当同步值有效时才覆盖本地值。

#### Scenario: 同步返回空字段时保留本地信息
- **WHEN** 同步返回的模型已存在于数据库
- **AND** 同步返回的 `display_name` 为空、`context_window` 为 0、`capabilities` 为空
- **THEN** 数据库中该模型的 `display_name`、`context_window`、`capabilities` 保持同步前的本地值不变

#### Scenario: 同步返回有效字段时刷新本地信息
- **WHEN** 同步返回的模型已存在于数据库
- **AND** 同步返回的 `context_window=131072`、`capabilities="vision,tools"`
- **THEN** 数据库中该模型的 `context_window` 更新为 131072
- **AND** `capabilities` 更新为 "vision,tools"

#### Scenario: 逐字段独立判断
- **WHEN** 同步返回的模型中 `owned_by` 有效但 `max_output` 为 0
- **THEN** `owned_by` 被刷新
- **AND** `max_output` 保留本地值
### Requirement: 同步结果只反映厂商真实返回的信息

系统 SHALL 使同步器只填充厂商接口实际返回的字段值，不得为未返回的字段伪造内容（如用模型 ID 冒充显示名、硬编码默认能力集）。

#### Scenario: 厂商未返回显示名时不伪造
- **WHEN** 厂商 `/models` 接口的条目未提供独立显示名
- **THEN** 同步结果的 `display_name` 为空
- **AND** 合并时保留本地已有的 `display_name`

#### Scenario: 厂商未返回能力信息时不伪造
- **WHEN** 厂商 `/models` 接口的条目未提供能力信息
- **THEN** 同步结果的 `capabilities` 为空
- **AND** 合并时保留本地已有的 `capabilities`
### Requirement: 模型来源不再单独记录

系统 SHALL 移除模型的 `source` 字段（手动/同步来源标记），可用性三态承担其失效判断职责：`unknown` 模型不参与失效判断。编辑、删除等管理操作不因可用性状态受限（维持现状）。存量数据按来源与可用性映射迁移至三态。

#### Scenario: API 响应以三态替代来源
- **WHEN** 客户端请求厂商模型列表或详情接口
- **THEN** 响应包含 `availability` 字段
- **AND** 响应不再包含 `source` 与 `is_available` 字段

#### Scenario: 存量数据迁移
- **WHEN** 系统从旧版本升级
- **THEN** 原 `source="manual"` 的模型迁移为 `unknown`
- **AND** 原 `source="sync"` 且可用的模型迁移为 `available`
- **AND** 原不可用的模型迁移为 `unavailable`
### Requirement: 同步响应报告失效计数

系统 SHALL 在同步响应中返回本次被标记失效的模型数量 `deactivated`，前端在同步完成后提示管理员。

#### Scenario: 响应包含失效计数
- **WHEN** 同步完成且有 2 个模型被标记失效
- **THEN** 同步响应包含 `deactivated=2`
- **AND** 前端提示"标记 N 个失效模型"

#### Scenario: 无失效时计数为零
- **WHEN** 同步完成且无模型被标记失效
- **THEN** 同步响应包含 `deactivated=0`
