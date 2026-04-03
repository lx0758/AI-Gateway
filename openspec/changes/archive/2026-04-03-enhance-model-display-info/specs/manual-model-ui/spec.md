## MODIFIED Requirements

### Requirement: Manual model management UI

The system SHALL provide a user interface for manually managing provider models with enhanced display of model capabilities and token information.

#### Scenario: Model list displays formatted token information with tooltip
- **WHEN** admin views a provider detail page model list
- **THEN** system displays context_window and max_output in formatted notation
- **AND** values >= 1000 are displayed as "XK" notation (e.g., 128000 → "128K")
- **AND** values >= 1000000 are displayed as "XM" notation (e.g., 2000000 → "2M")
- **AND** values < 1000 are displayed as original number
- **AND** mouse hover displays original values (e.g., "128,000 / 4,096")

#### Scenario: Model list displays capability tags before context window
- **WHEN** admin views a provider detail page model list
- **THEN** system displays a "Capabilities" column before "Context Window" column
- **AND** each model shows tags for supported capabilities (Vision, Tools, Stream)
- **AND** models with supports_vision=true display a Vision tag (type="success")
- **AND** models with supports_tools=true display a Tools tag (type="warning")
- **AND** models with supports_stream=true display a Stream tag (type="primary")

#### Scenario: Add model button visible
- **WHEN** admin views a provider detail page
- **THEN** system displays an "Add Model" button above the model list

#### Scenario: Add model dialog
- **WHEN** admin clicks "Add Model" button
- **THEN** system displays a form dialog with fields: model_id, display_name, context_window, max_output, supports_vision, supports_tools, supports_stream

#### Scenario: Create manual model
- **WHEN** admin fills the form and submits
- **THEN** system creates a new provider model with source="manual"
- **AND** system displays the new model in the list with formatted token and capability tags

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
- **AND** system displays updated token and capability information in the list

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