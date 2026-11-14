# provider-multi-style 规格

## Purpose

定义 Provider 的多 API 风格支持：OpenAI Chat Completions、Anthropic Messages 与 OpenAI Responses 三种 BaseURL 配置、API Key 跨风格共享、按调用协议创建 Provider 实例与方言分发。

## Requirements

### Requirement: Provider 支持多种 API 风格端点

系统 SHALL 允许单个 Provider 同时配置 OpenAI 风格、Anthropic 风格和 Responses 风格的 BaseURL 端点。

#### Scenario: Provider 配置了三种风格
- **WHEN** 管理员创建或更新 Provider，OpenAIChatCompletionsBaseURL="https://api.openai.com/v1"，AnthropicMessagesBaseURL="https://api.anthropic.com/v1"，OpenAIResponsesBaseURL="https://open.bigmodel.cn/api/v1"
- **THEN** 系统存储三个 BaseURL 值
- **AND** Provider 被视为支持 OpenAI、Anthropic 和 Responses 三种 API 风格

#### Scenario: Provider 配置了两种风格
- **WHEN** 管理员创建或更新 Provider，OpenAIChatCompletionsBaseURL="https://api.openai.com/v1"，AnthropicMessagesBaseURL="https://api.anthropic.com/v1"，OpenAIResponsesBaseURL=""
- **THEN** 系统存储 OpenAIChatCompletionsBaseURL、AnthropicMessagesBaseURL 和空的 OpenAIResponsesBaseURL
- **AND** Provider 被视为支持 OpenAI 和 Anthropic 两种 API 风格，不支持原生 Responses

#### Scenario: Provider 配置了单种风格
- **WHEN** 管理员创建 Provider，仅 OpenAIChatCompletionsBaseURL="https://example.com/v1"，其余两个 BaseURL 为空
- **THEN** 系统存储该值和空的其余两个 BaseURL
- **AND** Provider 被视为仅支持 OpenAI API 风格

#### Scenario: Provider 创建验证
- **WHEN** 管理员创建 Provider，OpenAIChatCompletionsBaseURL=""、AnthropicMessagesBaseURL="" 且 OpenAIResponsesBaseURL=""
- **THEN** 系统拒绝请求，返回验证错误 "至少必须提供一个 BaseURL"

#### Scenario: Provider 模型同步使用适当的端点
- **WHEN** 管理员在配置了一个以上 BaseURL 的 Provider 上触发模型同步
- **THEN** 系统按 OpenAI → Anthropic → Responses 的顺位选择可用端点
- **AND** 模型在该 Provider 下只存储一次（不重复）

### Requirement: Provider API Key 在风格间共享

系统 SHALL 使用单个 APIKey 字段用于 OpenAI、Anthropic 和 Responses 端点认证。

#### Scenario: 共享认证
- **WHEN** Provider 有 APIKey="sk-xxx" 并配置了多个 BaseURL
- **THEN** OpenAI 与 Responses 端点调用在 Authorization header 中使用 APIKey
- **AND** Anthropic 端点调用在 x-api-key header 中使用相同的 APIKey

#### Scenario: API Key 更新影响两种风格
- **WHEN** 管理员更新 Provider.APIKey
- **THEN** 后续对 OpenAI、Anthropic 和 Responses 端点的调用都使用新 Key

### Requirement: Provider 响应暴露配置的风格

系统 SHALL 在 Provider API 响应中暴露 OpenAIChatCompletionsBaseURL、AnthropicMessagesBaseURL 和 OpenAIResponsesBaseURL，指示支持的能力。

#### Scenario: 列出 Provider 显示支持的风格
- **WHEN** 客户端请求 GET /api/v1/providers
- **THEN** 每个 Provider 响应包含 openai_chat_completions_base_url、anthropic_messages_base_url 和 openai_responses_base_url 字段
- **AND** 空字符串表示该风格不被支持

#### Scenario: Provider 详情显示支持的风格
- **WHEN** 客户端请求 GET /api/v1/providers/:id
- **THEN** 响应包含 openai_chat_completions_base_url、anthropic_messages_base_url 和 openai_responses_base_url 字段
- **AND** 非空值表示支持的 API 风格

### Requirement: Provider 构造器接受 Config

系统 SHALL 允许 Provider 实现使用包含全部已配置 BaseURL 与 APIKey 的 Config 结构体构造，而不依赖 model.Provider。

#### Scenario: OpenAI Compatible Provider 构造
- **WHEN** 使用 Config{OpenAIChatCompletionsBaseURL: "https://api.openai.com/v1", APIKey: "sk-xxx"} 创建 OpenAIProvider
- **THEN** Provider 实例使用指定的端点和认证创建
- **AND** Provider 不持有对 model.Provider 的引用

#### Scenario: Anthropic Provider 构造
- **WHEN** 使用 Config{AnthropicMessagesBaseURL: "https://api.anthropic.com/v1", APIKey: "sk-xxx"} 创建 AnthropicProvider
- **THEN** Provider 实例使用指定的端点和认证创建
- **AND** Provider 不持有对 model.Provider 的引用

### Requirement: Router 创建 Provider 实例

系统 SHALL 在 Router 中基于 Provider 数据和 requestStyle 直接创建 Provider 实例。

#### Scenario: Router 为 OpenAI 请求创建 OpenAI Provider
- **WHEN** Router.Route(alias, "openai") 遇到 OpenAIChatCompletionsBaseURL="https://api.openai.com/v1" 的 Provider
- **THEN** Router 使用携带该 Provider 全部已配置 BaseURL 与 APIKey 的 Config 创建 OpenAIProvider
- **AND** Provider 实例可以直接执行 OpenAI 请求

#### Scenario: Router 为 Responses 请求创建 Provider
- **WHEN** Router.Route(alias, "responses") 遇到配置了 OpenAIResponsesBaseURL 的 Provider
- **THEN** Router 使用携带该 Provider 全部已配置 BaseURL 与 APIKey 的 Config 创建 Provider 实例
- **AND** Provider 实例按方言分发权重执行原生透传或跨协议转换

#### Scenario: Router 为 OpenAI 请求创建 Anthropic Provider（回退）
- **WHEN** Router.Route(alias, "openai") 遇到只有 AnthropicMessagesBaseURL="https://api.anthropic.com/v1" 的 Provider
- **THEN** Router 使用携带该 Provider 全部已配置 BaseURL 与 APIKey 的 Config 创建 AnthropicProvider
- **AND** Provider 实例将 OpenAI 请求转换为 Anthropic 格式

#### Scenario: Router 返回多个 Provider 实例
- **WHEN** Router.Route(alias, "openai") 在 ModelMapping 中找到多个 Provider
- **THEN** Router 返回按优先级排序的 Provider 实例列表
- **AND** 列表中第一个实例支持 OpenAI 格式（如果任何 Provider 支持它）
- **AND** 其余实例按权重排序，用于负载均衡和故障转移

### Requirement: 数据库迁移移除 Type 字段

系统 SHALL 迁移 Provider 数据模型，移除 Type 字段并添加 OpenAIChatCompletionsBaseURL 和 AnthropicMessagesBaseURL 字段。

#### Scenario: 迁移现有的 type=openai 的 Provider
- **WHEN** 系统运行数据库迁移
- **THEN** type="openai" 的 Provider 变为 openai_chat_completions_base_url=(旧的 base_url) 且 anthropic_messages_base_url=""
- **AND** type 字段被移除

#### Scenario: 迁移现有的 type=anthropic 的 Provider
- **WHEN** 系统运行数据库迁移
- **THEN** type="anthropic" 的 Provider 变为 openai_chat_completions_base_url="" 且 anthropic_messages_base_url=(旧的 base_url)
- **AND** type 字段被移除

#### Scenario: 迁移不丢失数据
- **WHEN** 系统在现有 Provider 记录上运行数据库迁移
- **THEN** 所有 provider_model 记录保持关联到正确的 provider_id
- **AND** 所有 model_mapping 记录保持有效

### Requirement: BaseURL 字段按产品名词表命名

系统 SHALL 将三个 BaseURL 字段在数据库列、Go 结构体字段、API JSON 与前端表单中统一按外部产品名词表命名：`openai_chat_completions_base_url`、`anthropic_messages_base_url`、`openai_responses_base_url`。系统 SHALL NOT 提供新旧字段名的自动兼容读取。

#### Scenario: 存量列更名由人工执行
- **WHEN** 运维在部署新版本前手动执行数据库列重命名 DDL，随后部署使用新列名的版本
- **THEN** 系统从新列名读取到原值，功能正常
- **AND** 代码中不存在对旧列名的兼容读取逻辑

#### Scenario: 新增 Responses 列默认为空
- **WHEN** 数据库迁移新增 openai_responses_base_url 列
- **THEN** 该列默认为空字符串，存量 Provider 因不具备原生 Responses 能力而走协议转换

#### Scenario: 表单提供 Responses BaseURL 输入
- **WHEN** 管理员打开 Provider 新增或编辑表单
- **THEN** 表单提供可选的 Responses BaseURL 输入框，与其余两个 BaseURL 输入并列

#### Scenario: Responses BaseURL 前缀可与 Chat 不一致
- **WHEN** 管理员为智谱类上游配置 OpenAIChatCompletionsBaseURL="https://open.bigmodel.cn/api/paas/v4" 与 OpenAIResponsesBaseURL="https://open.bigmodel.cn/api/v1"
- **THEN** 系统分别存储两个值，Responses 请求按 OpenAIResponsesBaseURL 拼接 `/responses` 路径
