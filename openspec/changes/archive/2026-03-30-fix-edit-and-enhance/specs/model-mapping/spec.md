## ADDED Requirements

### Requirement: Update model mapping

The system SHALL allow modifying model mapping configuration after creation.

#### Scenario: Update mapping alias
- **WHEN** admin changes the alias of an existing mapping
- **THEN** system updates the alias and future requests use the new alias

#### Scenario: Update mapping provider
- **WHEN** admin changes the provider for a mapping
- **THEN** system updates the provider_id and provider_model_id accordingly

#### Scenario: Update mapping model
- **WHEN** admin changes the target model for a mapping
- **THEN** system updates the provider_model_id and routing directs to the new model
