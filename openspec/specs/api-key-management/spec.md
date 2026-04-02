# API Key Management Specification

## Overview

This specification defines the requirements for API key management, including creation, permission control, usage tracking, and validation.

---

## Requirements

### Requirement: Create API key

The system SHALL allow creating access keys for clients to use the proxy API.

#### Scenario: Create key with name
- **WHEN** admin creates API key with a name/description
- **THEN** system generates unique key with "sk-" prefix and stores hashed version

#### Scenario: Set key expiration
- **WHEN** admin sets expiration date for key
- **THEN** system stores expires_at timestamp

---

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

---

### Requirement: Set key quota

The system SHALL support token-based quota limits.

#### Scenario: Set token quota
- **WHEN** admin sets quota for a key
- **THEN** system tracks usage and rejects requests when quota exceeded

#### Scenario: Unlimited quota
- **WHEN** admin leaves quota unset or zero
- **THEN** key has unlimited usage

---

### Requirement: Set rate limit

The system SHALL support request rate limiting per key.

#### Scenario: Set rate limit
- **WHEN** admin sets rate_limit value
- **THEN** system enforces maximum requests per minute

---

### Requirement: List API keys

The system SHALL display all API keys with masked values.

#### Scenario: List keys
- **WHEN** admin views API keys page
- **THEN** system shows all keys with name, masked key, usage, and status

---

### Requirement: Revoke API key

The system SHALL allow deleting/revoking keys.

#### Scenario: Revoke key
- **WHEN** admin deletes an API key
- **THEN** system marks it as revoked and future requests with this key are rejected

---

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

---

### Requirement: Validate model access on request

The system SHALL validate model permissions when processing API requests.

#### Scenario: Access granted to requested model
- **WHEN** client requests a model with a key that has permission for that model
- **THEN** system processes the request normally

#### Scenario: Access denied to restricted model
- **WHEN** client requests a model with a key that does not have permission for that model
- **THEN** system returns 403 error with "model not allowed" message

---

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

---

### Requirement: Cascade delete on alias removal

The system SHALL automatically remove KeyModel records when referenced alias is deleted.

#### Scenario: Delete alias with assigned keys
- **WHEN** admin deletes an alias that is assigned to API keys
- **THEN** all KeyModel records referencing that AliasID are automatically deleted
- **AND** API keys lose permission to access that model

---

### Requirement: Manage model permissions via API

The system SHALL provide API endpoints for managing key-model permissions.

#### Scenario: List key's allowed models
- **WHEN** admin requests permissions for an API key
- **THEN** system returns list of model aliases the key can access

#### Scenario: Update key permissions
- **WHEN** admin updates permissions for an API key
- **THEN** system replaces all existing permissions with the new list
