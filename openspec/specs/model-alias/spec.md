## Requirements

### Requirement: Alias is unique identifier

The system SHALL enforce uniqueness on Alias.name field, ensuring each alias represents a distinct model name for API calls.

#### Scenario: Create unique alias
- **WHEN** admin creates alias "gpt-4"
- **THEN** system creates Alias record with name="gpt-4"
- **AND** subsequent creation of alias "gpt-4" fails with duplicate error

#### Scenario: Alias used as API model name
- **WHEN** client calls API with model="gpt-4"
- **THEN** router finds Alias where name="gpt-4"
- **AND** router retrieves associated AliasMappings for routing

### Requirement: Alias can be enabled or disabled

The system SHALL allow enabling/disabling Alias, affecting all associated mappings' availability.

#### Scenario: Disabled alias rejects routing
- **WHEN** Alias.enabled=false
- **AND** client calls API with model="gpt-4"
- **THEN** router returns no providers
- **AND** handler returns "model not found" error

#### Scenario: Enabled alias allows routing
- **WHEN** Alias.enabled=true
- **AND** client calls API with model="gpt-4"
- **THEN** router retrieves AliasMappings
- **AND** routing proceeds normally

### Requirement: Alias has zero or more AliasMappings

The system SHALL support one-to-many relationship between Alias and AliasMapping.

#### Scenario: Alias with multiple mappings
- **WHEN** Alias "gpt-4" has 3 AliasMappings
- **THEN** router retrieves all 3 mappings for routing
- **AND** mappings are sorted by weight DESC

#### Scenario: Alias with zero mappings
- **WHEN** Alias "new-model" has no AliasMappings
- **THEN** router returns no providers
- **AND** handler returns "model not found" error

### Requirement: Deleting Alias cascades to AliasMappings

The system SHALL cascade delete all AliasMappings when Alias is deleted.

#### Scenario: Delete alias removes mappings
- **WHEN** admin deletes Alias "gpt-4"
- **THEN** system deletes all AliasMappings where alias_id=gpt-4.id
- **AND** no orphan mappings remain