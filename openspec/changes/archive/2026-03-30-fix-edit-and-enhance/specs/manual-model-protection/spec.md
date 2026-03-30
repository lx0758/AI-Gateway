## ADDED Requirements

### Requirement: Protect manual models during sync

The system SHALL preserve manually created models during provider model synchronization.

#### Scenario: Sync skips manual models
- **WHEN** admin triggers model sync for a provider
- **THEN** system updates only models with source="sync", skipping source="manual" models

#### Scenario: Manual model preserved after sync
- **WHEN** sync completes for a provider with manually added models
- **THEN** manually added models retain their configuration unchanged

### Requirement: Mark model source on creation

The system SHALL track the source of each provider model.

#### Scenario: Sync creates model with source=sync
- **WHEN** sync creates a new model from provider API
- **THEN** system sets source="sync" for the model

#### Scenario: Manual creation sets source=manual
- **WHEN** admin manually adds a model
- **THEN** system sets source="manual" for the model

### Requirement: Allow manual model deletion

The system SHALL allow deleting manually created models.

#### Scenario: Delete manual model
- **WHEN** admin deletes a model with source="manual"
- **THEN** system removes the model from database
