## ADDED Requirements

### Requirement: Router prioritizes matching API style

The system SHALL prioritize Providers that support the request's API style when routing, minimizing format conversion overhead.

#### Scenario: OpenAI request returns OpenAI-capable Provider first
- **WHEN** router receives request for model alias with requestStyle="openai"
- **AND** ModelMapping lists ProviderA(weight=100, openai_base_url="xxx") and ProviderB(weight=80, anthropic_base_url="yyy")
- **THEN** router returns list where ProviderA instance appears first
- **AND** ProviderA instance is OpenAIProvider

#### Scenario: OpenAI request returns all Providers as fallback candidates
- **WHEN** router receives request for model alias with requestStyle="openai"
- **AND** ModelMapping lists ProviderA(weight=100, openai_base_url="", anthropic_base_url="yyy")
- **THEN** router returns list containing ProviderA instance
- **AND** ProviderA instance is AnthropicProvider (will convert O→A)

#### Scenario: Weight priority within same style support
- **WHEN** router receives request for model alias with requestStyle="openai"
- **AND** ModelMapping lists ProviderA(weight=50, openai_base_url="xxx") and ProviderB(weight=100, openai_base_url="yyy")
- **THEN** router returns list where ProviderB instance appears first (higher weight)
- **AND** ProviderA instance appears second
- **AND** Both instances are OpenAIProvider

#### Scenario: Anthropic request returns Anthropic-capable Provider first
- **WHEN** router receives request for model alias with requestStyle="anthropic"
- **AND** ModelMapping lists ProviderA(weight=100, anthropic_base_url="xxx") and ProviderB(weight=80, openai_base_url="yyy")
- **THEN** router returns list where ProviderA instance appears first
- **AND** ProviderA instance is AnthropicProvider

### Requirement: Router accepts request style parameter

The system SHALL accept a requestStyle parameter indicating the API format of the incoming request.

#### Scenario: Route method signature
- **WHEN** Handler calls router.Route(alias, requestStyle)
- **THEN** requestStyle parameter indicates "openai" or "anthropic"
- **AND** router uses this parameter for style matching logic

### Requirement: Router returns Provider instance list

The system SHALL return a list of Provider instances in RouteResult, sorted by format priority and weight.

#### Scenario: RouteResult structure
- **WHEN** router.Route(alias, requestStyle) completes
- **THEN** RouteResult contains Candidates field
- **AND** Candidates is a slice of Provider instances
- **AND** Instances are sorted by: matching style first (by weight DESC), then non-matching style (by weight DESC)

#### Scenario: Handler uses first instance
- **WHEN** Handler receives RouteResult with multiple Provider instances
- **THEN** Handler uses Candidates[0] for current request
- **AND** Remaining instances are available for future load balancing or failover

#### Scenario: Empty list when no Provider available
- **WHEN** router.Route(alias, requestStyle) finds no ModelMapping
- **THEN** RouteResult.Candidates is empty slice
- **AND** Handler returns "model not found" error

### Requirement: Provider instances are independent of model data

The system SHALL create Provider instances using Config struct without holding references to model.Provider.

#### Scenario: Provider instance encapsulation
- **WHEN** router creates Provider instance from model.Provider data
- **THEN** Provider instance receives Config{BaseURL, APIKey}
- **AND** Provider instance does not reference model.Provider
- **AND** Changes to model.Provider do not affect existing Provider instances