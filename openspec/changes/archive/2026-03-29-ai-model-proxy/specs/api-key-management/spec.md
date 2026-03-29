## ADDED Requirements

### Requirement: Create API key

The system SHALL allow creating access keys for clients to use the proxy API.

#### Scenario: Create key with name
- **WHEN** admin creates API key with a name/description
- **THEN** system generates unique key with "sk-" prefix and stores hashed version

#### Scenario: Set key expiration
- **WHEN** admin sets expiration date for key
- **THEN** system stores expires_at timestamp

### Requirement: Set key permissions

The system SHALL support limiting which models a key can access.

#### Scenario: Allow all models
- **WHEN** admin leaves allowed_models empty
- **THEN** key can access all available models

#### Scenario: Restrict to specific models
- **WHEN** admin specifies allowed_models list
- **THEN** key can only access those models

### Requirement: Set key quota

The system SHALL support token-based quota limits.

#### Scenario: Set token quota
- **WHEN** admin sets quota for a key
- **THEN** system tracks usage and rejects requests when quota exceeded

#### Scenario: Unlimited quota
- **WHEN** admin leaves quota unset or zero
- **THEN** key has unlimited usage

### Requirement: Set rate limit

The system SHALL support request rate limiting per key.

#### Scenario: Set rate limit
- **WHEN** admin sets rate_limit value
- **THEN** system enforces maximum requests per minute

### Requirement: List API keys

The system SHALL display all API keys with masked values.

#### Scenario: List keys
- **WHEN** admin views API keys page
- **THEN** system shows all keys with name, masked key, usage, and status

### Requirement: Revoke API key

The system SHALL allow deleting/revoking keys.

#### Scenario: Revoke key
- **WHEN** admin deletes an API key
- **THEN** system marks it as revoked and future requests with this key are rejected

### Requirement: Validate API key on request

The system SHALL validate API key for every API request.

#### Scenario: Valid key check
- **WHEN** request arrives with valid API key
- **THEN** system processes the request

#### Scenario: Expired key check
- **WHEN** request arrives with expired key
- **THEN** system returns 401 error

#### Scenario: Quota exceeded check
- **WHEN** request arrives with key that has exceeded quota
- **THEN** system returns 429 error with quota exceeded message
