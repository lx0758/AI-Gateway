# provider-multi-style 规格

## 目的
待定 - 由归档变更 optimize-provider-config 创建。归档后更新 Purpose。

## Requirements
### Requirement: Provider 支持多种 API 风格端点

系统 SHALL 允许单个 Provider 同时配置 OpenAI 风格和 Anthropic 风格的 BaseURL 端点。

#### Scenario: Provider 配置了两种风格
- **WHEN** 管理员创建或更新 Provider，OpenAIBaseURL="https://api.openai.com/v1"，AnthropicBaseURL="https://api.anthropic.com/v1"
- **THEN** 系统存储两个 BaseURL 值
- **AND** Provider 被视为支持 OpenAI 和 Anthropic 两种 API 风格

#### Scenario: Provider 配置了单种风格
- **WHEN** 管理员创建 Provider，仅 OpenAIBaseURL="https://example.com/v1"，AnthropicBaseURL=""
- **THEN** 系统存储 OpenAIBaseURL 和空的 AnthropicBaseURL
- **AND** Provider 被视为仅支持 OpenAI API 风格

#### Scenario: Provider 创建验证
- **WHEN** 管理员创建 Provider，OpenAIBaseURL="" 且 AnthropicBaseURL=""
- **THEN** 系统拒绝请求，返回验证错误 "至少必须提供一个 BaseURL"

#### Scenario: Provider 模型同步使用适当的端点
- **WHEN** 管理员在配置了两个 BaseURL 的 Provider 上触发模型同步
- **THEN** 系统根据 Provider 实现能力选择端点
- **AND** 模型在该 Provider 下只存储一次（不重复）

### Requirement: Provider API Key 在风格间共享

系统 SHALL 使用单个 APIKey 字段用于 OpenAI 和 Anthropic 端点认证。

#### Scenario: 共享认证
- **WHEN** Provider 有 APIKey="sk-xxx" 并配置了两个 BaseURL
- **THEN** OpenAI 端点调用在 Authorization header 中使用 APIKey
- **AND** Anthropic 端点调用在 x-api-key header 中使用相同的 APIKey

#### Scenario: API Key 更新影响两种风格
- **WHEN** 管理员更新 Provider.APIKey
- **THEN** 后续对 OpenAI 和 Anthropic 端点的调用都使用新 Key

### Requirement: Provider 响应暴露配置的风格

系统 SHALL 在 Provider API 响应中暴露 OpenAIBaseURL 和 AnthropicBaseURL，指示支持的能力。

#### Scenario: 列出 Provider 显示支持的风格
- **WHEN** 客户端请求 GET /api/v1/providers
- **THEN** 每个 Provider 响应包含 openai_base_url 和 anthropic_base_url 字段
- **AND** 空字符串表示该风格不被支持

#### Scenario: Provider 详情显示支持的风格
- **WHEN** 客户端请求 GET /api/v1/providers/:id
- **THEN** 响应包含 openai_base_url 和 anthropic_base_url 字段
- **AND** 非空值表示支持的 API 风格

### Requirement: Factory 根据调用风格创建 Provider

**REMOVED** - Factory 层已删除，Provider 实例由 Router 内部创建

**Reason**: Router 已经掌握所有信息（Provider 数据 + requestStyle），无需额外的 Factory 层；合并后简化调用链，支持返回实例列表用于负载均衡。

**Migration**: Router 内部直接创建 Provider 实例，Handler 直接使用 Router 返回的实例列表。

### Requirement: Provider 构造器接受 Config

系统 SHALL 允许 Provider 实现使用包含 BaseURL 和 APIKey 的 Config 结构体构造，而不依赖 model.Provider。

#### Scenario: OpenAI Compatible Provider 构造
- **WHEN** 使用 Config{BaseURL: "https://api.openai.com/v1", APIKey: "sk-xxx"} 创建 OpenAIProvider
- **THEN** Provider 实例使用指定的端点和认证创建
- **AND** Provider 不持有对 model.Provider 的引用

#### Scenario: Anthropic Provider 构造
- **WHEN** 使用 Config{BaseURL: "https://api.anthropic.com/v1", APIKey: "sk-xxx"} 创建 AnthropicProvider
- **THEN** Provider 实例使用指定的端点和认证创建
- **AND** Provider 不持有对 model.Provider 的引用

### Requirement: Router 创建 Provider 实例

系统 SHALL 在 Router 中基于 Provider 数据和 requestStyle 直接创建 Provider 实例。

#### Scenario: Router 为 OpenAI 请求创建 OpenAI Provider
- **WHEN** Router.Route(alias, "openai") 遇到 OpenAIBaseURL="https://api.openai.com/v1" 的 Provider
- **THEN** Router 使用 Config{BaseURL: OpenAIBaseURL, APIKey: APIKey} 创建 OpenAIProvider
- **AND** Provider 实例可以直接执行 OpenAI 请求

#### Scenario: Router 为 OpenAI 请求创建 Anthropic Provider（回退）
- **WHEN** Router.Route(alias, "openai") 遇到只有 AnthropicBaseURL="https://api.anthropic.com/v1" 的 Provider
- **THEN** Router 使用 Config{BaseURL: AnthropicBaseURL, APIKey: APIKey} 创建 AnthropicProvider
- **AND** Provider 实例将 OpenAI 请求转换为 Anthropic 格式

#### Scenario: Router 返回多个 Provider 实例
- **WHEN** Router.Route(alias, "openai") 在 ModelMapping 中找到多个 Provider
- **THEN** Router 返回按优先级排序的 Provider 实例列表
- **AND** 列表中第一个实例支持 OpenAI 格式（如果任何 Provider 支持它）
- **AND** 其余实例按权重排序，用于负载均衡和故障转移

### Requirement: 数据库迁移移除 Type 字段

系统 SHALL 迁移 Provider 数据模型，移除 Type 字段并添加 OpenAIBaseURL 和 AnthropicBaseURL 字段。

#### Scenario: 迁移现有的 type=openai 的 Provider
- **WHEN** 系统运行数据库迁移
- **THEN** type="openai" 的 Provider 变为 openai_base_url=(旧的 base_url) 且 anthropic_base_url=""
- **AND** type 字段被移除

#### Scenario: 迁移现有的 type=anthropic 的 Provider
- **WHEN** 系统运行数据库迁移
- **THEN** type="anthropic" 的 Provider 变为 openai_base_url="" 且 anthropic_base_url=(旧的 base_url)
- **AND** type 字段被移除

#### Scenario: 迁移不丢失数据
- **WHEN** 系统在现有 Provider 记录上运行数据库迁移
- **THEN** 所有 provider_model 记录保持关联到正确的 provider_id
- **AND** 所有 model_mapping 记录保持有效