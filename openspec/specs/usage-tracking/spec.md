## MODIFIED Requirements

### Requirement: 记录 API 调用使用量

系统 SHALL 将每次 API 调用记录到 `model_logs` 表，包含以下信息：
- `key_id`: 请求使用的 API Key
- `provider_id`: 处理请求的 Provider
- `model`: 用户请求的虚拟模型名称
- `actual_model`: Provider 使用的实际模型名称
- `total_tokens`: 消耗的总 Token 数（int64）
- `latency_ms`: 请求延迟（毫秒）（int64）
- `status`: "success" 或 "error"
- `error_msg`: 失败时的错误消息，成功时为空

#### Scenario: 成功的 OpenAI 请求

- **WHEN** 成功的 OpenAI 兼容 API 请求完成
- **THEN** 系统记录 model log，status="success"，total_tokens 来自响应，latency_ms 为计算值

#### Scenario: 失败的 Anthropic 请求

- **WHEN** Anthropic API 请求因错误失败
- **THEN** 系统记录 model log，status="error"，error_msg 包含错误内容，total_tokens=0

### Requirement: 查询 Provider 统计信息

系统 SHALL 提供 API 来查询按 Provider 分组的使用统计信息。

#### Scenario: 查询 Provider 统计

- **WHEN** 用户请求某个日期范围的 Provider 统计信息
- **THEN** 系统返回每个 Provider 的调用次数、总 Token 数、平均延迟

### Requirement: 查询 Key 统计信息

系统 SHALL 提供 API 来查询按 API Key 分组的使用统计信息。

#### Scenario: 查询 Key 统计

- **WHEN** 用户请求某个日期范围的 API Key 统计信息
- **THEN** 系统返回每个 Key 的调用次数、总 Token 数、平均延迟

### Requirement: 在仪表盘显示使用统计信息

系统 SHALL 在仪表盘页面显示使用统计信息，包括：
- 消耗的总 Token 数
- 平均延迟
- Provider 统计表

#### Scenario: 查看仪表盘

- **WHEN** 用户打开仪表盘页面
- **THEN** 系统显示总请求数、今日请求数、活跃 Provider、活跃 Key、总 Token 数、平均延迟

### Requirement: 在使用页面显示 Key 统计信息

系统 SHALL 在使用页面显示 API Key 统计信息。

#### Scenario: 在使用页面查看 Key 统计

- **WHEN** 用户打开使用页面
- **THEN** 系统显示一个表格，包含每个 Key 的名称、调用次数、总 Token 数、平均延迟