## ADDED Requirements

### Requirement: Manual model management UI

The system SHALL provide a user interface for manually managing provider models.

#### Scenario: Add model button visible
- **WHEN** admin views a provider detail page
- **THEN** system displays an "Add Model" button above the model list

#### Scenario: Add model dialog
- **WHEN** admin clicks "Add Model" button
- **THEN** system displays a form dialog with fields: model_id, display_name, context_window, max_output, supports_vision, supports_tools, supports_stream

#### Scenario: Create manual model
- **WHEN** admin fills the form and submits
- **THEN** system creates a new provider model with source="manual"
- **AND** system displays the new model in the list

### Requirement: Edit manual model

The system SHALL allow editing manually created models.

#### Scenario: Edit button for manual models
- **WHEN** model list contains a model with source="manual"
- **THEN** system displays an edit button for that model

#### Scenario: Edit button hidden for sync models
- **WHEN** model list contains a model with source="sync"
- **THEN** system does NOT display an edit button for that model

#### Scenario: Update manual model
- **WHEN** admin edits a manual model and submits
- **THEN** system updates the model configuration

### Requirement: Delete manual model UI

The system SHALL provide UI to delete manually created models.

#### Scenario: Delete button for manual models
- **WHEN** model list contains a model with source="manual"
- **THEN** system displays a delete button for that model

#### Scenario: Delete button hidden for sync models
- **WHEN** model list contains a model with source="sync"
- **THEN** system does NOT display a delete button for that model

#### Scenario: Confirm before delete
- **WHEN** admin clicks delete button
- **THEN** system displays a confirmation dialog

### Requirement: Display model source

The system SHALL indicate the source of each model in the list.

#### Scenario: Source label displayed
- **WHEN** model list is displayed
- **THEN** each model shows a label indicating "Manual" or "Sync"
