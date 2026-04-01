## Why

当前厂商（Provider）配置采用"单一类型+单一 BaseURL"的设计，导致支持多种 API 风格的后端需要创建多个 Provider 实例。例如，一个同时支持 OpenAI 和 Anthropic 格式的自建服务需要分别创建两个 Provider，造成模型数据重复存储、管理复杂度增加。同时，路由逻辑未考虑格式匹配优先级，可能产生不必要的格式转换开销。

## What Changes

- Provider 数据模型调整：从 `Type + BaseURL` 改为 `OpenAIBaseURL + AnthropicBaseURL`，支持一个厂商配置多种 API 风格
- **BREAKING** 删除 Provider.Type 字段，改为通过 BaseURL 是否存在判断支持的能力
- Provider 创建/更新 API：请求和响应字段相应调整，验证逻辑改为"至少一个 BaseURL 不为空"
- Provider 实现类（OpenAIProvider、AnthropicProvider）保持不变，由 Router 内部创建
- 路由逻辑优化：增加 `requestStyle` 参数，按 weight 排序后优先匹配支持同格式的 Provider，降级时才使用转换
- Router.Route() 返回 Provider 实例列表（[]Provider），支持未来负载均衡和故障转移
- 删除 Factory 层，Router 内部根据 Provider 数据和 requestStyle 直接创建实例
- Handler 调整：直接使用 Router 返回的 Provider 实例列表第一个元素
- 前端表单：删除 Type 下拉框，改为两个 BaseURL 输入框（至少一个必填）

## Capabilities

### New Capabilities

- `provider-multi-style`: 厂商支持多种 API 风格的配置能力（一个 Provider 可配置 OpenAI 和 Anthropic 的 BaseURL）

### Modified Capabilities

- `model-mapping`: 路由逻辑增加格式匹配优先级，返回结果增加 DirectCall/CallStyle 标记

## Impact

- 数据模型：`model.Provider` 结构体变更（删除 Type，新增两个 BaseURL 字段）
- 数据库：需要迁移，删除 type 列，新增 openai_base_url 和 anthropic_base_url 列
- API 接口：`/api/v1/providers` 创建/更新/响应字段变更
- 路由逻辑：`router.ModelRouter.Route()` 方法签名和返回值变更
- Provider Factory：新增 `CreateForOpenAI()` 和 `CreateForAnthropic()` 方法
- Handler：`proxy_openai.go` 和 `proxy_anthropic.go` 调用路由和创建 Provider 的逻辑调整
- 前端：厂商编辑表单 UI 和验证逻辑调整
- 向后兼容：Provider.Type 字段删除，需要数据迁移策略