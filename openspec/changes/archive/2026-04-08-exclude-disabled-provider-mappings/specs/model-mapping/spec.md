## ADDED Requirements

### Requirement: API layer excludes disabled Provider mappings

The system SHALL filter out ModelMappings associated with disabled Providers (Provider.enabled=false) in all API responses and statistics calculations.

#### Scenario: Model list excludes disabled Provider mappings
- **WHEN** admin requests GET /api/v1/models
- **AND** Model has ModelMappings where some have Provider.enabled=false
- **THEN** response mappings array only includes mappings with Provider.enabled=true
- **AND** mapping_count reflects only enabled Provider mappings
- **AND** min_context_window, min_max_output, capabilities calculated from enabled Provider mappings only

#### Scenario: Model detail excludes disabled Provider mappings
- **WHEN** admin requests GET /api/v1/models/:id
- **AND** Model has ModelMappings with Provider.enabled=false
- **THEN** response mappings array excludes those mappings
- **AND** all mapping-related statistics reflect only enabled Provider mappings

#### Scenario: Model mapping list excludes disabled Provider mappings
- **WHEN** admin requests GET /api/v1/models/:id/mappings
- **AND** Model has ModelMappings with Provider.enabled=false
- **THEN** response mappings array excludes those mappings

#### Scenario: Key model list excludes disabled Provider mappings
- **WHEN** admin requests GET /api/v1/keys/:id/models
- **AND** Model has ModelMappings with Provider.enabled=false
- **THEN** response mapping_count, min_context_window, min_max_output, and capabilities calculated from enabled Provider mappings only

#### Scenario: Statistics calculation ignores disabled Provider mappings
- **WHEN** system calculates min_context_window or min_max_output
- **AND** ModelMapping has Provider.enabled=false
- **THEN** that mapping is excluded from calculation
- **WHEN** system calculates capabilities intersection
- **AND** ModelMapping has Provider.enabled=false
- **THEN** that mapping is excluded from intersection

#### Scenario: Model update response excludes disabled Provider mappings
- **WHEN** admin updates Model via PUT /api/v1/models/:id
- **AND** Model has ModelMappings with Provider.enabled=false
- **THEN** response mappings array excludes those mappings
