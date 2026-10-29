# usage-tracking 规格

## Purpose

定义用量跟踪与统计：API 调用记录、Provider 与 Key 维度统计查询，及仪表盘与使用页展示。

## Requirements

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
- 资产概览统计（厂商、模型、Keys、MCP服务等）

**重要**: 所有资产统计查询 SHALL 排除已软删除的记录（deleted_at IS NOT NULL）。

#### Scenario: 查看仪表盘

- **WHEN** 用户打开仪表盘页面
- **THEN** 系统显示总请求数、今日请求数、活跃 Provider、活跃 Key、总 Token 数、平均延迟
- **AND** 资产概览中"可用数量"不大于"总数量"

#### Scenario: 统计排除软删除记录

- **WHEN** 存在已软删除的 ProviderModel 记录
- **THEN** totalProviderModels 统计不包含已软删除的记录
- **AND** activeProviderModels 统计不包含已软删除的记录

#### Scenario: MCP 资产统计排除软删除记录

- **WHEN** 存在已软删除的 MCP 工具/资源/提示词记录
- **THEN** 对应的 total 和 active 统计均不包含已软删除的记录

### Requirement: 在使用页面显示 Key 统计信息

系统 SHALL 在使用页面显示 API Key 统计信息。

#### Scenario: 在使用页面查看 Key 统计

- **WHEN** 用户打开使用页面
- **THEN** 系统显示一个表格，包含每个 Key 的名称、调用次数、总 Token 数、平均延迟

### Requirement: 仪表盘资产卡片颜色一致性

系统 SHALL 使用一致的颜色方案展示资产概览卡片：
- **primary（蓝色）**：AI 模型链路相关（厂商、厂商模型、模型）
- **success（绿色）**：认证鉴权相关
- **warning（橙色）**：MCP 生态相关（MCP服务、MCP工具、MCP资源、MCP提示词）

#### Scenario: 查看资产概览卡片颜色

- **WHEN** 用户查看仪表盘资产概览
- **THEN** AI 模型链路卡片显示蓝色
- **AND** Keys 卡片显示绿色
- **AND** MCP 相关卡片显示橙色

### Requirement: 仪表盘数据格式一致性

系统 SHALL 使用一致的数值格式化方案展示统计数据：
- 数值保留 1 位小数（如 1.2K、1.5M、1.0s、1.2 KB）

#### Scenario: 查看数据格式化

- **WHEN** 用户查看仪表盘统计数据
- **THEN** 请求数、Token数、延迟、数据量等数值均显示 1 位小数
