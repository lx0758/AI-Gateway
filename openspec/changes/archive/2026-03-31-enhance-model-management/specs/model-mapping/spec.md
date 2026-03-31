## MODIFIED Requirements

### Requirement: ModelMapping association

The system SHALL associate ModelMapping with ProviderModel using model_id string instead of database primary key.

#### Scenario: Create mapping with model name
- **WHEN** admin creates a model mapping
- **THEN** system stores provider_model_name (model_id string) instead of provider_model_id

#### Scenario: Query mapping routes to model
- **WHEN** router looks up model by alias
- **THEN** system finds ProviderModel by (provider_id, model_id) instead of primary key

#### Scenario: Model ID change invalidates mapping
- **WHEN** provider sync causes model_id to change
- **THEN** mapping still references the old model_id string
- **AND** query finds no matching ProviderModel
- **AND** routing returns "model not found" error

## REMOVED Requirements

### Requirement: Foreign key association

**Reason**: Using database ID creates tight coupling; model_id string is more resilient to sync changes.

**Migration**: 
- Add provider_model_name column
- Populate from existing provider_model_id → model_id lookup
- Remove provider_model_id column
