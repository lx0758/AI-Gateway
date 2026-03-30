## ADDED Requirements

### Requirement: Configure model permissions for API key

The system SHALL allow configuring which models an API key can access using a relation table.

#### Scenario: Grant model access to key
- **WHEN** admin grants model "gpt-4" access to an API key
- **THEN** system creates a record in api_key_models table linking the key to the model alias

#### Scenario: Revoke model access from key
- **WHEN** admin revokes model "gpt-4" access from an API key
- **THEN** system deletes the corresponding record from api_key_models table

#### Scenario: Empty permission list allows all models
- **WHEN** API key has no records in api_key_models table
- **THEN** system allows access to all available models

### Requirement: Validate model access on request

The system SHALL validate model permissions when processing API requests.

#### Scenario: Access granted to requested model
- **WHEN** client requests model "gpt-4" with a key that has "gpt-4" permission
- **THEN** system processes the request normally

#### Scenario: Access denied to restricted model
- **WHEN** client requests model "claude-3" with a key that only has "gpt-4" permission
- **THEN** system returns 403 error with "model not allowed" message

### Requirement: Manage model permissions via API

The system SHALL provide API endpoints for managing key-model permissions.

#### Scenario: List key's allowed models
- **WHEN** admin requests permissions for an API key
- **THEN** system returns list of model aliases the key can access

#### Scenario: Update key permissions
- **WHEN** admin updates permissions for an API key
- **THEN** system replaces all existing permissions with the new list
