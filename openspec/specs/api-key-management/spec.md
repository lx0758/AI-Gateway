# api-key-management 规格

## Purpose

定义网关 API Key 的生命周期与权限管理：Key 的创建、配额与速率限制、请求时的认证与模型访问权限验证，以及模型权限的级联维护。

## Requirements

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

### Requirement: 设置 Key 配额

系统 SHALL 支持 Token 基础配额限制。

#### Scenario: 设置 Token 配额
- **WHEN** 管理员为 Key 设置配额
- **THEN** 系统跟踪使用量并在配额超限时拒绝请求

#### Scenario: 无限配额
- **WHEN** 管理员将配额留空或设为零
- **THEN** Key 有无限使用量

### Requirement: 设置速率限制

系统 SHALL 支持每个 Key 的请求速率限制。

#### Scenario: 设置速率限制
- **WHEN** 管理员设置 rate_limit 值
- **THEN** 系统强制每分钟最大请求数

### Requirement: 列出 API Keys

系统 SHALL 显示所有 API Keys 并附带掩码值。

#### Scenario: 列出 Keys
- **WHEN** 管理员查看 API Keys 页面
- **THEN** 系统显示所有 Keys，附带名称、掩码 Key、使用量和状态

### Requirement: 撤销 API Key

系统 SHALL 允许删除/撤销 Keys。

#### Scenario: 撤销 Key
- **WHEN** 管理员删除 API Key
- **THEN** 系统将其标记为已撤销，使用此 Key 的未来请求被拒绝

### Requirement: 在请求时验证 API Key

系统 SHALL 对每个 API 请求验证 API Key。

#### Scenario: 有效 Key 检查
- **WHEN** 请求到达时附带有效 API Key
- **THEN** 系统处理该请求

#### Scenario: 过期 Key 检查
- **WHEN** 请求到达时附带过期 Key
- **THEN** 系统返回 401 错误

#### Scenario: 配额超限检查
- **WHEN** 请求到达时 Key 已超配额
- **THEN** 系统返回 429 错误，附带配额超限消息

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

### Requirement: 响应中返回 AliasID 和名称

系统 SHALL 在 API Key 模型响应中返回 AliasID 和 Alias 名称。

#### Scenario: 列出 Keys 附带模型信息
- **WHEN** 管理员查看 API Keys 列表
- **THEN** 每个 Key 的 models 数组包含对象，带：
  - `id`: KeyModel 记录 ID
  - `alias_id`: 引用的 Alias ID
  - `alias_name`: 当前 Alias 名称

#### Scenario: 创建 Key 响应
- **WHEN** 管理员创建带模型的 API Key
- **THEN** 响应包含 models 数组，带 alias_id 和 alias_name

#### Scenario: 更新 Key 响应
- **WHEN** 管理员更新 API Key 模型
- **THEN** 响应包含更新的 models 数组，带 alias_id 和 alias_name

### Requirement: Alias 移除时级联删除

系统 SHALL 在被引用的 Alias 删除时自动移除 KeyModel 记录。

#### Scenario: 删除被分配 Keys 的 Alias
- **WHEN** 管理员删除被分配给 API Keys 的 Alias
- **THEN** 所有引用该 AliasID 的 KeyModel 记录自动删除
- **AND** API Keys 失去访问该模型的权限

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
