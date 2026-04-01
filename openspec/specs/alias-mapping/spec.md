## Requirements

### Requirement: AliasMapping belongs to Alias

The system SHALL enforce foreign key relationship between AliasMapping.alias_id and Alias.id.

#### Scenario: Create mapping requires existing alias
- **WHEN** admin creates AliasMapping with alias_id=1
- **THEN** system verifies Alias.id=1 exists
- **AND** creation succeeds if alias exists
- **AND** creation fails with "alias not found" error if alias does not exist

#### Scenario: Mapping retrieved with alias
- **WHEN** router queries AliasMapping
- **THEN** system can Preload Alias relationship
- **AND** mapping.alias.name field is accessible

### Requirement: AliasMapping references Provider and ProviderModel

The system SHALL associate AliasMapping with Provider (by provider_id) and reference ProviderModel by model_id string.

#### Scenario: Mapping with valid provider
- **WHEN** admin creates AliasMapping with provider_id=5, provider_model_name="gpt-4-turbo"
- **THEN** system verifies Provider.id=5 exists
- **AND** system verifies ProviderModel(provider_id=5, model_id="gpt-4-turbo") exists
- **AND** creation succeeds

#### Scenario: Mapping with invalid provider
- **WHEN** admin creates AliasMapping with provider_id=999
- **THEN** creation fails with "provider not found" error

#### Scenario: Mapping with invalid model
- **WHEN** admin creates AliasMapping with provider_id=5, provider_model_name="nonexistent"
- **THEN** creation fails with "provider model not found" error

### Requirement: AliasMapping has weight for load balancing

The system SHALL use AliasMapping.weight field for routing priority.

#### Scenario: Higher weight routes first
- **WHEN** Alias "gpt-4" has AliasMappings with weights [10, 50, 30]
- **THEN** router returns providers sorted by weight DESC: [50, 30, 10]

#### Scenario: Default weight is 1
- **WHEN** admin creates AliasMapping without specifying weight
- **THEN** system sets weight=1 by default

### Requirement: AliasMapping can be enabled or disabled

The system SHALL allow enabling/disabling individual AliasMapping without affecting sibling mappings.

#### Scenario: Disabled mapping excluded from routing
- **WHEN** AliasMapping.enabled=false
- **THEN** router excludes this mapping from provider list

#### Scenario: Disabled mapping still visible in UI
- **WHEN** AliasMapping.enabled=false
- **THEN** admin can see mapping in management UI
- **AND** admin can re-enable mapping

### Requirement: AliasMapping supports Provider association

The system SHALL Preload Provider relationship for AliasMapping queries.

#### Scenario: Mapping includes Provider info
- **WHEN** API returns AliasMapping
- **THEN** mapping includes provider object with {id, name, openai_base_url, anthropic_base_url}
- **AND** UI can display provider name and type tags