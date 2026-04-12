# API Key 管理规格

## 概述

本规格定义了 API Key 管理的要求，包括创建、权限控制、使用跟踪和验证。

---

## Requirements

### Requirement: 创建 API Key

系统 SHALL 允许为客户端创建访问密钥以使用代理 API。

#### Scenario: 创建带名称的 Key
- **WHEN** 管理员创建带名称/描述的 API Key
- **THEN** 系统生成带 "sk-" 前缀的唯一密钥并存储哈希版本

#### Scenario: 设置 Key 过期时间
- **WHEN** 管理员为 Key 设置过期日期
- **THEN** 系统存储 expires_at 时间戳

---

### Requirement: 设置 Key 权限

系统 SHALL 支持使用 AliasID 引用限制 Key 可以访问的模型。

#### Scenario: 允许所有模型
- **WHEN** 管理员将模型列表留空
- **THEN** Key 可以访问所有可用模型

#### Scenario: 通过 AliasID 限制特定模型
- **WHEN** 管理员指定 AliasID 值的模型列表
- **THEN** Key 只能访问对应这些 AliasIDs 的模型
- **AND** 系统验证每个 AliasID 存在于 aliases 表

#### Scenario: 创建使用无效 AliasID
- **WHEN** 管理员提供不存在的 AliasID
- **THEN** 系统返回 400 Bad Request 错误，消息为 "alias 未找到"

#### Scenario: Key 创建后 Alias 重命名
- **WHEN** Alias 在被分配给 API Key 后重命名
- **THEN** API Key 仍引用相同的 AliasID
- **AND** API Key 自动显示新的 Alias 名称

---

### Requirement: 设置 Key 配额

系统 SHALL 支持 Token 基础配额限制。

#### Scenario: 设置 Token 配额
- **WHEN** 管理员为 Key 设置配额
- **THEN** 系统跟踪使用量并在配额超限时拒绝请求

#### Scenario: 无限配额
- **WHEN** 管理员将配额留空或设为零
- **THEN** Key 有无限使用量

---

### Requirement: 设置速率限制

系统 SHALL 支持每个 Key 的请求速率限制。

#### Scenario: 设置速率限制
- **WHEN** 管理员设置 rate_limit 值
- **THEN** 系统强制每分钟最大请求数

---

### Requirement: 列出 API Keys

系统 SHALL 显示所有 API Keys 并附带掩码值。

#### Scenario: 列出 Keys
- **WHEN** 管理员查看 API Keys 页面
- **THEN** 系统显示所有 Keys，附带名称、掩码 Key、使用量和状态

---

### Requirement: 撤销 API Key

系统 SHALL 允许删除/撤销 Keys。

#### Scenario: 撤销 Key
- **WHEN** 管理员删除 API Key
- **THEN** 系统将其标记为已撤销，使用此 Key 的未来请求被拒绝

---

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

---

### Requirement: 在请求时验证模型访问权限

系统 SHALL 在处理 API 请求时验证模型权限。

#### Scenario: 授权访问请求的模型
- **WHEN** 客户端请求一个 Key 有权限的模型
- **THEN** 系统正常处理请求

#### Scenario: 拒绝访问受限模型
- **WHEN** 客户端请求一个 Key 无权限的模型
- **THEN** 系统返回 403 错误，消息为 "模型不允许"

---

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

---

### Requirement: Alias 移除时级联删除

系统 SHALL 在被引用的 Alias 删除时自动移除 KeyModel 记录。

#### Scenario: 删除被分配 Keys 的 Alias
- **WHEN** 管理员删除被分配给 API Keys 的 Alias
- **THEN** 所有引用该 AliasID 的 KeyModel 记录自动删除
- **AND** API Keys 失去访问该模型的权限

---

### Requirement: 通过 API 管理模型权限

系统 SHALL 提供用于管理 Key-Model 权限的 API 端点。

#### Scenario: 列出 Key 的允许模型
- **WHEN** 管理员请求 API Key 的权限
- **THEN** 系统返回 Key 可访问的模型 Alias 列表

#### Scenario: 更新 Key 权限
- **WHEN** 管理员更新 API Key 的权限
- **THEN** 系统用新列表替换所有现有权限