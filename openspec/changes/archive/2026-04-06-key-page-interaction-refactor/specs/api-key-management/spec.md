## ADDED Requirements

### Requirement: Get Key Detail Info

The system SHALL provide endpoint to retrieve single Key basic information.

#### Scenario: Get single key
- **WHEN** admin calls GET `/keys/:id`
- **THEN** system returns key basic info: id, key (masked), name, enabled, expires_at, created_at

---

### Requirement: Get Key Models with Status

The system SHALL return all system models with permission status for a Key.

#### Scenario: Get models with selected status
- **WHEN** admin calls GET `/keys/:id/models`
- **THEN** system returns array of all enabled models
- **AND** each model object includes: id, name, selected (boolean)
- **AND** selected=true if model is in KeyModel table for this key
- **AND** selected=false if model is NOT in KeyModel table

---

### Requirement: Single Model Association Operations

The system SHALL support adding and removing single model association for a Key.

#### Scenario: Add model association
- **WHEN** admin calls POST `/keys/:id/models/:model_id`
- **THEN** system creates KeyModel record if not exists
- **AND** returns 200 OK

#### Scenario: Add duplicate association
- **WHEN** admin calls POST `/keys/:id/models/:model_id` for existing association
- **THEN** system returns 200 OK without error (no duplicate created)

#### Scenario: Remove model association
- **WHEN** admin calls DELETE `/keys/:id/models/:model_id`
- **THEN** system deletes KeyModel record if exists
- **AND** returns 200 OK

#### Scenario: Remove non-existent association
- **WHEN** admin calls DELETE `/keys/:id/models/:model_id` for non-existent association
- **THEN** system returns 200 OK without error

---

### Requirement: Clear All Model Associations

The system SHALL support clearing all model associations for a Key.

#### Scenario: Clear all models
- **WHEN** admin calls DELETE `/keys/:id/models`
- **THEN** system deletes all KeyModel records for this key
- **AND** returns 200 OK with message

---

### Requirement: List Keys with MCP Counts

The system SHALL return MCP component counts in Keys list response.

#### Scenario: List with counts
- **WHEN** admin calls GET `/keys`
- **THEN** each key object includes: models_count, mcp_tools_count, mcp_resources_count, mcp_prompts_count
- **AND** counts are calculated from respective association tables

#### Scenario: Empty associations
- **WHEN** key has no associations for a component type
- **THEN** corresponding count is 0