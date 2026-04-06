## RENAMED Requirements

FROM: `Alias` → TO: `Model`
FROM: `aliases` → TO: `models`

---

## MODIFIED Requirements

### Requirement: Model is unique identifier

The system SHALL enforce uniqueness on Model.name field, ensuring each model represents a distinct model name for API calls.

#### Scenario: Create unique model
- **WHEN** admin creates model "gpt-4"
- **THEN** system creates Model record with name="gpt-4"
- **AND** subsequent creation of model "gpt-4" fails with duplicate error

#### Scenario: Model used as API model name
- **WHEN** client calls API with model="gpt-4"
- **THEN** router finds Model where name="gpt-4"
- **AND** router retrieves associated ModelMappings for routing

### Requirement: Model can be enabled or disabled

The system SHALL allow enabling/disabling Model, affecting all associated mappings' availability.

#### Scenario: Disabled model rejects routing
- **WHEN** Model.enabled=false
- **AND** client calls API with model="gpt-4"
- **THEN** router returns no providers
- **AND** handler returns "model not found" error

#### Scenario: Enabled model allows routing
- **WHEN** Model.enabled=true
- **AND** client calls API with model="gpt-4"
- **THEN** router retrieves ModelMappings
- **AND** routing proceeds normally

### Requirement: Model has zero or more ModelMappings

The system SHALL support one-to-many relationship between Model and ModelMapping.

#### Scenario: Model with multiple mappings
- **WHEN** Model "gpt-4" has 3 ModelMappings
- **THEN** router retrieves all 3 mappings for routing
- **AND** mappings are sorted by weight DESC

#### Scenario: Model with zero mappings
- **WHEN** Model "new-model" has no ModelMappings
- **THEN** router returns no providers
- **AND** handler returns "model not found" error

### Requirement: Deleting Model cascades to ModelMappings

The system SHALL cascade delete all ModelMappings when Model is deleted.

#### Scenario: Delete model removes mappings
- **WHEN** admin deletes Model "gpt-4"
- **THEN** system deletes all ModelMappings where model_id=gpt-4.id
- **AND** no orphan mappings remain

### Requirement: Model list displays in flat table layout

The system SHALL display models in a flat table layout (not collapsible panels), consistent with Providers page style.

#### Scenario: Model list displays as flat table
- **WHEN** admin views `/models` page
- **THEN** system displays table with columns: Select, Name, Mapping Count, Token Summary, Status, Actions
- **AND** each row represents one Model
- **AND** no collapsible panels are used

#### Scenario: Token summary displays minimum values with tooltip
- **WHEN** model has enabled mappings
- **THEN** system calculates min_context_window from enabled mappings
- **AND** system calculates min_max_output from enabled mappings
- **AND** Token Summary column shows formatted display (e.g., "8K / 4K")
- **AND** mouse hover displays original values (e.g., "128,000 / 4,096")
- **WHEN** model has no enabled mappings
- **THEN** Token Summary column shows "-"

#### Scenario: Capabilities intersection displays correctly
- **WHEN** model has enabled mappings
- **THEN** system calculates capability intersection for all enabled mappings
- **AND** supports_vision is true only if all enabled mappings support vision
- **AND** supports_tools is true only if all enabled mappings support tools
- **AND** supports_stream is true only if all enabled mappings support stream
- **AND** Capabilities column shows tags for true capabilities only
- **WHEN** model has no enabled mappings
- **THEN** Capabilities column shows "-"

#### Scenario: Token summary formatted with one decimal
- **WHEN** system formats min_context_window and min_max_output
- **THEN** values < 1000 display as original number
- **AND** values >= 1000 display as "XK" or "X.XK" (e.g., "128K", "153.6K")
- **AND** values >= 1000000 display as "XM" or "X.XM" (e.g., "2M", "1.5M")

### Requirement: Model list supports batch operations

The system SHALL allow batch selection and batch deletion of models.

#### Scenario: Select multiple models
- **WHEN** admin clicks checkboxes in model table
- **THEN** system tracks selected model IDs
- **AND** "Batch Delete" button shows count (e.g., "Batch Delete (3)")

#### Scenario: Batch delete models
- **WHEN** admin clicks "Batch Delete" with selected models
- **THEN** system shows confirmation dialog
- **AND** upon confirmation, system deletes all selected models and their mappings
- **AND** system refreshes model table

#### Scenario: Batch delete disabled without selection
- **WHEN** no models are selected
- **THEN** "Batch Delete" button is disabled

### Requirement: Model list provides detail page navigation

The system SHALL provide "Detail" button to navigate to model detail page for mapping management.

#### Scenario: Navigate to model detail page
- **WHEN** admin clicks "Detail" button on model row
- **THEN** system navigates to `/models/:id` page
- **AND** detail page shows complete mapping information

#### Scenario: Detail button visible for all models
- **WHEN** model table is displayed
- **THEN** every model row has "Detail" button in Actions column
- **AND** "Detail" button is always enabled regardless of mapping count
