## ADDED Requirements

### Requirement: Sync models from provider

The system SHALL fetch available models from provider API and store in database.

#### Scenario: Sync OpenAI models
- **WHEN** admin triggers sync for OpenAI provider
- **THEN** system calls GET /v1/models endpoint and stores all returned models

#### Scenario: Sync Anthropic models
- **WHEN** admin triggers sync for Anthropic provider
- **THEN** system uses built-in model list (Anthropic has no list endpoint)

#### Scenario: Record sync timestamp
- **WHEN** sync completes successfully
- **THEN** system updates provider's last_sync_at timestamp

### Requirement: Detect model changes

The system SHALL identify changes in model availability during sync.

#### Scenario: New model detected
- **WHEN** sync finds model not in database
- **THEN** system creates new provider_model record with is_available = true

#### Scenario: Model removed
- **WHEN** sync finds model in database but not in provider response
- **THEN** system sets is_available = false for that model

#### Scenario: Model metadata updated
- **WHEN** model metadata (pricing, context window) changes
- **THEN** system updates the corresponding fields

### Requirement: Manual model management

The system SHALL allow manual addition of models for providers without list API.

#### Scenario: Add model manually
- **WHEN** admin adds model with model_id and display_name
- **THEN** system creates provider_model with source = "manual"

#### Scenario: Edit model metadata
- **WHEN** admin updates model's pricing or context window
- **THEN** system saves the changes

### Requirement: List provider models

The system SHALL display all models for a provider.

#### Scenario: List all models
- **WHEN** admin views provider details
- **THEN** system shows all models with their metadata and availability status

#### Scenario: Filter by availability
- **WHEN** admin enables "show available only" filter
- **THEN** system shows only models with is_available = true

### Requirement: Store model metadata

The system SHALL store relevant metadata for each model.

#### Scenario: Store pricing info
- **WHEN** model has pricing information
- **THEN** system stores input_price and output_price per 1K tokens

#### Scenario: Store capability flags
- **WHEN** model supports specific features
- **THEN** system stores supports_vision, supports_tools, supports_stream flags

#### Scenario: Store context window
- **WHEN** model has known context window size
- **THEN** system stores context_window in tokens
