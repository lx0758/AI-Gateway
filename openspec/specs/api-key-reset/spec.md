# api-key-reset Specification

## Purpose
TBD - created by archiving change reset-api-key. Update Purpose after archive.
## Requirements
### Requirement: 用户可以重置 API Key
系统 SHALL 提供 API Key 重置功能，生成新的 Key 值并保留原有配置（名称、启用状态、过期时间、模型绑定）。

#### Scenario: 成功重置 API Key
- **WHEN** 用户发送 POST 请求到 /api-keys/:id/reset
- **THEN** 系统生成新的 Key 值，更新数据库记录
- **AND** 返回包含 masked key 和 raw_key 的响应
- **AND** 保留原有的 Name、Enabled、ExpiresAt、Models 配置

#### Scenario: 重置不存在的 API Key
- **WHEN** 用户发送 POST 请求到 /api-keys/:id/reset，但 ID 不存在
- **THEN** 系统返回 404 Not Found 错误

#### Scenario: 重置已删除的 API Key
- **WHEN** 用户发送 POST 请求到 /api-keys/:id/reset，但 API Key 已被软删除
- **THEN** 系统返回 404 Not Found 错误

### Requirement: 重置后旧 Key 立即失效
系统 SHALL 在重置操作完成后立即使旧 Key 值失效，所有使用旧 Key 的认证请求 SHALL 返回 401 Unauthorized。

#### Scenario: 旧 Key 认证失败
- **WHEN** API Key 重置完成
- **AND** 请求使用旧 Key 值进行认证
- **THEN** 认证中间件返回 401 Unauthorized
- **AND** 错误信息为 "invalid api key"

#### Scenario: 新 Key 认证成功
- **WHEN** API Key 重置完成
- **AND** 请求使用新 Key 值进行认证
- **THEN** 认证中间件验证成功
- **AND** 请求继续执行

### Requirement: 重置响应格式
系统 SHALL 返回与创建 API Key 相同的响应格式，包含完整的 Key 信息和 raw_key。

#### Scenario: 响应包含 masked key 和 raw_key
- **WHEN** 重置 API Key 成功
- **THEN** 响应 JSON 包含 key 对象（包含 ID、masked Key、Name、Enabled、ExpiresAt、CreatedAt、Models）
- **AND** 响应 JSON 包含 raw_key 字段（完整的未掩码 Key 值）

#### Scenario: masked key 格式
- **WHEN** 重置 API Key 成功
- **AND** 新 Key 值长度大于 8 个字符
- **THEN** 返回的 key.key 字段格式为前 8 字符 + "****" + 后 4 字符

### Requirement: 重置操作原子性
系统 SHALL 确保重置操作的原子性，数据库更新成功后才返回响应，失败时返回错误且旧 Key 保持有效。

#### Scenario: 数据库更新失败
- **WHEN** 重置操作过程中数据库更新失败
- **THEN** 系统返回 500 Internal Server Error
- **AND** 旧 Key 值保持有效

### Requirement: Key 值生成安全
系统 SHALL 使用安全随机数生成器生成新的 Key 值，格式为 "sk-" + 48 字符十六进制字符串。

#### Scenario: Key 值格式
- **WHEN** 重置 API Key 成功
- **THEN** 新 Key 值格式为 "sk-" + 48 字符十六进制字符串
- **AND** 使用 crypto/rand 生成随机数

#### Scenario: Key 值唯一性
- **WHEN** 重置 API Key 成功
- **THEN** 新 Key 值在系统中唯一（数据库唯一索引约束）
- **AND** 不会与现有其他 API Key 冲突

### Requirement: 保留模型绑定配置
系统 SHALL 在重置操作中保留 API Key 与模型的绑定关系，不删除或修改 KeyModel 关联记录。

#### Scenario: 保留模型绑定
- **WHEN** API Key 已绑定多个模型
- **AND** 用户重置该 API Key
- **THEN** 重置后响应中的 Models 列表包含原有绑定的所有模型
- **AND** 数据库中 KeyModel 记录不变

