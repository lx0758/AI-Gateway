## ADDED Requirements

### Requirement: Provider supports multiple API style endpoints

The system SHALL allow a single Provider to configure both OpenAI-style and Anthropic-style BaseURL endpoints simultaneously.

#### Scenario: Provider with both styles configured
- **WHEN** admin creates or updates a Provider with OpenAIBaseURL="https://api.openai.com/v1" and AnthropicBaseURL="https://api.anthropic.com/v1"
- **THEN** system stores both BaseURL values
- **AND** Provider is considered to support both OpenAI and Anthropic API styles

#### Scenario: Provider with single style configured
- **WHEN** admin creates a Provider with only OpenAIBaseURL="https://example.com/v1" and AnthropicBaseURL=""
- **THEN** system stores OpenAIBaseURL and empty AnthropicBaseURL
- **AND** Provider is considered to support only OpenAI API style

#### Scenario: Provider creation validation
- **WHEN** admin creates a Provider with both OpenAIBaseURL="" and AnthropicBaseURL=""
- **THEN** system rejects the request with validation error "at least one BaseURL must be provided"

#### Scenario: Provider model sync uses appropriate endpoint
- **WHEN** admin triggers model sync on a Provider with both BaseURLs configured
- **THEN** system chooses the endpoint based on Provider implementation capability
- **AND** models are stored once under this Provider (not duplicated)

### Requirement: Provider API key is shared across styles

The system SHALL use a single APIKey field for both OpenAI and Anthropic endpoint authentication.

#### Scenario: Shared authentication
- **WHEN** Provider has APIKey="sk-xxx" and both BaseURLs configured
- **THEN** OpenAI endpoint calls use APIKey in Authorization header
- **AND** Anthropic endpoint calls use the same APIKey in x-api-key header

#### Scenario: API key update affects both styles
- **WHEN** admin updates Provider.APIKey
- **THEN** subsequent calls to both OpenAI and Anthropic endpoints use the new key

### Requirement: Provider response exposes configured styles

The system SHALL expose OpenAIBaseURL and AnthropicBaseURL in Provider API responses, indicating supported capabilities.

#### Scenario: List providers shows supported styles
- **WHEN** client requests GET /api/v1/providers
- **THEN** each Provider response includes openai_base_url and anthropic_base_url fields
- **AND** empty strings indicate that style is not supported

#### Scenario: Provider detail shows supported styles
- **WHEN** client requests GET /api/v1/providers/:id
- **THEN** response includes openai_base_url and anthropic_base_url fields
- **AND** non-empty values indicate supported API styles

### Requirement: Factory creates Provider based on call style

**REMOVED** - Factory 层已删除，Provider 实例由 Router 内部创建

**Reason**: Router 已经掌握所有信息（Provider 数据 + requestStyle），无需额外的 Factory 层；合并后简化调用链，支持返回实例列表用于负载均衡。

**Migration**: Router 内部直接创建 Provider 实例，Handler 直接使用 Router 返回的实例列表。

### Requirement: Provider constructor accepts Config

The system SHALL allow Provider implementations to be constructed with a Config struct containing BaseURL and APIKey, without depending on model.Provider.

#### Scenario: OpenAI Compatible Provider construction
- **WHEN** creating OpenAIProvider with Config{BaseURL: "https://api.openai.com/v1", APIKey: "sk-xxx"}
- **THEN** Provider instance is created with the specified endpoint and authentication
- **AND** Provider does not hold reference to model.Provider

#### Scenario: Anthropic Provider construction
- **WHEN** creating AnthropicProvider with Config{BaseURL: "https://api.anthropic.com/v1", APIKey: "sk-xxx"}
- **THEN** Provider instance is created with the specified endpoint and authentication
- **AND** Provider does not hold reference to model.Provider

### Requirement: Router creates Provider instances

The system SHALL create Provider instances directly in Router based on Provider data and requestStyle.

#### Scenario: Router creates OpenAI Provider for OpenAI request
- **WHEN** Router.Route(alias, "openai") encounters Provider with OpenAIBaseURL="https://api.openai.com/v1"
- **THEN** Router creates OpenAIProvider with Config{BaseURL: OpenAIBaseURL, APIKey: APIKey}
- **AND** Provider instance can execute OpenAI requests directly

#### Scenario: Router creates Anthropic Provider for OpenAI request (fallback)
- **WHEN** Router.Route(alias, "openai") encounters Provider with only AnthropicBaseURL="https://api.anthropic.com/v1"
- **THEN** Router creates AnthropicProvider with Config{BaseURL: AnthropicBaseURL, APIKey: APIKey}
- **AND** Provider instance will convert OpenAI requests to Anthropic format

#### Scenario: Router returns multiple Provider instances
- **WHEN** Router.Route(alias, "openai") finds multiple Providers in ModelMapping
- **THEN** Router returns a list of Provider instances sorted by priority
- **AND** First instance in list supports OpenAI format (if any Provider supports it)
- **AND** Remaining instances are sorted by weight for load balancing and failover

### Requirement: Database migration removes Type field

The system SHALL migrate Provider data model to remove Type field and add OpenAIBaseURL and AnthropicBaseURL fields.

#### Scenario: Migration of existing Provider with type=openai
- **WHEN** system runs database migration
- **THEN** Provider with type="openai" becomes Provider with openai_base_url=(old base_url) and anthropic_base_url=""
- **AND** type field is removed

#### Scenario: Migration of existing Provider with type=anthropic
- **WHEN** system runs database migration
- **THEN** Provider with type="anthropic" becomes Provider with openai_base_url="" and anthropic_base_url=(old base_url)
- **AND** type field is removed

#### Scenario: Migration does not lose data
- **WHEN** system runs database migration on existing Provider records
- **THEN** all provider_model records remain associated with correct provider_id
- **AND** all model_mapping records remain valid