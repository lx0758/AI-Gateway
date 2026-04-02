## MODIFIED Requirements

### Requirement: Set key permissions

The system SHALL support limiting which models a key can access using AliasID references.

#### Scenario: Allow all models
- **WHEN** admin leaves models list empty
- **THEN** key can access all available models

#### Scenario: Restrict to specific models by AliasID
- **WHEN** admin specifies models list with AliasID values
- **THEN** key can only access models corresponding to those AliasIDs
- **AND** system validates each AliasID exists in aliases table

#### Scenario: Create with invalid AliasID
- **WHEN** admin provides AliasID that does not exist
- **THEN** system returns 400 Bad Request error with message "alias not found"

#### Scenario: Alias renamed after key creation
- **WHEN** alias is renamed after being assigned to an API key
- **THEN** API key still references same AliasID
- **AND** API key displays new alias name automatically

## REMOVED Requirements

### Requirement: Add single model to key
**Reason**: Simplified API design, use batch update instead
**Migration**: Use `PUT /api-keys/:id` with full models list to update key permissions

### Requirement: Remove single model from key
**Reason**: Simplified API design, use batch update instead
**Migration**: Use `PUT /api-keys/:id` with full models list to update key permissions

## ADDED Requirements

### Requirement: Return AliasID and name in response

The system SHALL return both AliasID and alias name in API key model responses.

#### Scenario: List keys with model info
- **WHEN** admin views API keys list
- **THEN** each key's models array contains objects with:
  - `id`: KeyModel record ID
  - `alias_id`: referenced Alias ID
  - `alias_name`: current Alias name

#### Scenario: Create key response
- **WHEN** admin creates API key with models
- **THEN** response includes models array with alias_id and alias_name

#### Scenario: Update key response
- **WHEN** admin updates API key models
- **THEN** response includes updated models array with alias_id and alias_name

### Requirement: Cascade delete on alias removal

The system SHALL automatically remove KeyModel records when referenced alias is deleted.

#### Scenario: Delete alias with assigned keys
- **WHEN** admin deletes an alias that is assigned to API keys
- **THEN** all KeyModel records referencing that AliasID are automatically deleted
- **AND** API keys lose permission to access that model