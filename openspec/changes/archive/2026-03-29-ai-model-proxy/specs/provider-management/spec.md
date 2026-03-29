## ADDED Requirements

### Requirement: Create provider

The system SHALL allow creating new provider configurations via API and Web UI.

#### Scenario: Create provider with required fields
- **WHEN** admin submits provider form with name, api_type, base_url, and api_key
- **THEN** system creates provider record and stores encrypted api_key

#### Scenario: Validate provider type
- **WHEN** admin selects api_type
- **THEN** system accepts only "openai" or "anthropic" values

### Requirement: List providers

The system SHALL provide a list of all configured providers.

#### Scenario: List all providers
- **WHEN** admin requests provider list
- **THEN** system returns all providers with basic info (name, type, status, model count)

#### Scenario: Mask API keys
- **WHEN** listing providers
- **THEN** system returns masked API keys (showing only last 4 characters)

### Requirement: Update provider

The system SHALL allow updating provider configurations.

#### Scenario: Update provider settings
- **WHEN** admin updates provider's base_url or other settings
- **THEN** system saves the changes and updates timestamp

#### Scenario: Update API key
- **WHEN** admin provides new API key
- **THEN** system encrypts and stores the new key

### Requirement: Delete provider

The system SHALL allow deleting providers.

#### Scenario: Delete provider without dependencies
- **WHEN** admin deletes provider with no associated model mappings
- **THEN** system removes provider and all its models from database

#### Scenario: Delete provider with mappings
- **WHEN** admin attempts to delete provider with existing model mappings
- **THEN** system shows warning and requires confirmation or mapping removal

### Requirement: Test provider connection

The system SHALL provide ability to test provider connectivity.

#### Scenario: Successful connection test
- **WHEN** admin clicks "Test Connection"
- **THEN** system sends test request to provider API and reports success

#### Scenario: Failed connection test
- **WHEN** provider API is unreachable or credentials are invalid
- **THEN** system reports error with details

### Requirement: Enable/disable provider

The system SHALL allow toggling provider enabled status.

#### Scenario: Disable provider
- **WHEN** admin disables a provider
- **THEN** system stops routing requests to this provider

#### Scenario: Re-enable provider
- **WHEN** admin enables a disabled provider
- **THEN** system resumes routing to this provider
