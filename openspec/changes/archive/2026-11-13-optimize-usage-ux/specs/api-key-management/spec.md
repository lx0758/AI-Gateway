# Spec Delta

## ADDED Requirements

### Requirement: API Key 创建与编辑表单提供启用状态控制
系统 SHALL 在 API Key 的新增与编辑对话框中提供启用状态开关，其状态与列表行内状态开关联动一致。

#### Scenario: 新增对话框默认启用
- **WHEN** 管理员打开新增 API Key 对话框
- **THEN** 状态开关显示为启用且可调

#### Scenario: 编辑对话框回显当前状态
- **WHEN** 管理员打开某个禁用 Key 的编辑对话框
- **THEN** 状态开关显示为禁用
- **AND** 保存后该 Key 保持禁用状态

### Requirement: 设置 Key 模型白名单权限

系统 SHALL 通过显式的模型白名单开关限制 Key 可以访问的模型：开关关闭时允许全部模型，开启时仅允许勾选的模型。

#### Scenario: 白名单关闭时允许全部模型
- **WHEN** 管理员将模型白名单开关关闭
- **THEN** Key 可以访问所有可用模型
- **AND** 系统清除该 Key 已保存的模型关联记录

#### Scenario: 白名单开启时仅允许勾选模型
- **WHEN** 管理员开启模型白名单并勾选部分模型
- **THEN** Key 只能访问勾选的模型
- **AND** 请求未勾选的模型被拒绝

#### Scenario: 白名单开启但未勾选任何模型
- **WHEN** 管理员开启模型白名单且未勾选任何模型
- **THEN** 该 Key 的所有模型请求被拒绝
- **AND** 管理员在开启白名单前看到"将拒绝所有调用"的确认警示

#### Scenario: 主列表统计反映白名单状态
- **WHEN** 管理员在 API Key 主列表查看某 Key 各类型资源的放行统计
- **THEN** 白名单开关关闭时统计显示"不限制"
- **AND** 白名单开关开启且已勾选 N 项时统计显示"允许 N 个"
- **AND** 白名单开关开启且未勾选任何项时统计显示"允许 0 个"而非"不限制"

#### Scenario: 白名单关闭时勾选开关显示为开启且禁用
- **WHEN** 管理员打开白名单开关关闭的 Key 的模型列表
- **THEN** 每个模型的勾选开关显示为开启状态且不可操作
- **AND** 开关提示"白名单已关闭，全部允许"

#### Scenario: 创建时携带模型列表自动开启白名单
- **WHEN** 管理员创建 Key 时携带模型列表
- **THEN** 模型白名单开关为开启状态
- **AND** 仅列表中的模型被允许

### Requirement: 通过 API 管理模型白名单

系统 SHALL 提供用于管理 Key-模型白名单的 API 端点。

#### Scenario: 查看 Key 的白名单开关状态
- **WHEN** 管理员请求 GET /keys/:id
- **THEN** 响应包含 models/tools/resources/prompts 四个白名单开关字段

#### Scenario: 查看 Key 的模型勾选状态
- **WHEN** 管理员请求 GET /keys/:id/models
- **THEN** 系统返回全部可用模型及每项的勾选状态

#### Scenario: 切换模型白名单开关
- **WHEN** 管理员调用 PUT /keys/:id/whitelist/models 且 enabled=false
- **THEN** 开关关闭且该 Key 的模型关联记录被清除
- **WHEN** 管理员调用同一端点且 enabled=true
- **THEN** 开关开启且模型关联记录保持为空，等待管理员勾选

#### Scenario: 勾选与取消勾选模型
- **WHEN** 管理员通过 POST 或 DELETE /keys/:id/models/:model_id 勾选或取消勾选模型
- **THEN** 系统创建或删除对应关联记录

## MODIFIED Requirements

### Requirement: 创建 API Key

系统 SHALL 允许为客户端创建访问密钥以使用代理 API。

#### Scenario: 创建带名称的 Key
- **WHEN** 管理员创建带名称/描述的 API Key
- **THEN** 系统生成带 "sk-" 前缀的唯一密钥并存储哈希版本

#### Scenario: 设置 Key 过期时间
- **WHEN** 管理员为 Key 设置过期日期
- **THEN** 系统存储 expires_at 时间戳

#### Scenario: 创建时指定启用状态
- **WHEN** 管理员创建 Key 时携带 enabled=false
- **THEN** 创建的 Key 为禁用状态
- **AND** 该 Key 的后续代理请求被拒绝

#### Scenario: 创建时省略启用状态
- **WHEN** 管理员创建 Key 时未提供 enabled 字段
- **THEN** 创建的 Key 为启用状态

### Requirement: 在请求时验证模型访问权限

系统 SHALL 在处理 API 请求时按模型白名单开关验证模型权限。

#### Scenario: 白名单关闭时放行所有模型
- **WHEN** 客户端请求的 Key 模型白名单开关为关闭
- **THEN** 系统正常处理请求

#### Scenario: 授权访问请求的模型
- **WHEN** 客户端请求一个 Key 有权限的模型
- **THEN** 系统正常处理请求

#### Scenario: 拒绝访问受限模型
- **WHEN** 客户端请求一个 Key 白名单开启但未勾选的模型
- **THEN** 系统返回 403 错误，消息为 "模型不允许"

## REMOVED Requirements

### Requirement: 设置 Key 权限
**Reason**: 隐式白名单语义（留空=全部允许）废弃，由显式白名单开关替代，且原 AliasID 引用描述已与现实现（模型 ID 关联）不符。
**Migration**: 使用"设置 Key 模型白名单权限"要求；开关关闭即等价于原"允许所有模型"，开启并勾选即等价于原"限制特定模型"。

### Requirement: 通过 API 管理模型权限
**Reason**: 原"PUT 批量替换列表"的权限管理描述与显式白名单开关模型不符。
**Migration**: 使用"通过 API 管理模型白名单"要求；批量替换端点 PUT /keys/:id/models 仍可用于更新勾选列表，开关状态由 PUT /keys/:id/whitelist/<type> 管理。
