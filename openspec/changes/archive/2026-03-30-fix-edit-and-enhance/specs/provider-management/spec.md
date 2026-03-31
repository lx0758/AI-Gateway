## ADDED Requirements

### Requirement: Load existing provider data for editing

The system SHALL populate the edit form with existing provider data.

#### Scenario: Edit dialog shows current data
- **WHEN** admin clicks edit on an existing provider
- **THEN** system loads the provider's name, type, base_url into the edit form

#### Scenario: API key field shows placeholder
- **WHEN** admin opens edit dialog for a provider
- **THEN** API key field shows placeholder indicating existing key is stored (not the actual key)

#### Scenario: Submit partial updates
- **WHEN** admin modifies only some fields and submits
- **THEN** system updates only the changed fields, preserving others
