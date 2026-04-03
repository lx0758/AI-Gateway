## MODIFIED Requirements

### Requirement: Alias is unique identifier

The system SHALL enforce uniqueness on Alias.name field, ensuring each alias represents a distinct model name for API calls.

#### Scenario: Create unique alias
- **WHEN** admin creates alias "gpt-4"
- **THEN** system creates Alias record with name="gpt-4"
- **AND** subsequent creation of alias "gpt-4" fails with duplicate error

#### Scenario: Alias used as API model name
- **WHEN** client calls API with model="gpt-4"
- **THEN** router finds Alias where name="gpt-4"
- **AND** router retrieves associated AliasMappings for routing

### Requirement: Alias can be enabled or disabled

The system SHALL allow enabling/disabling Alias, affecting all associated mappings' availability.

#### Scenario: Disabled alias rejects routing
- **WHEN** Alias.enabled=false
- **AND** client calls API with model="gpt-4"
- **THEN** router returns no providers
- **AND** handler returns "model not found" error

#### Scenario: Enabled alias allows routing
- **WHEN** Alias.enabled=true
- **AND** client calls API with model="gpt-4"
- **THEN** router retrieves AliasMappings
- **AND** routing proceeds normally

### Requirement: Alias has zero or more AliasMappings

The system SHALL support one-to-many relationship between Alias and AliasMapping.

#### Scenario: Alias with multiple mappings
- **WHEN** Alias "gpt-4" has 3 AliasMappings
- **THEN** router retrieves all 3 mappings for routing
- **AND** mappings are sorted by weight DESC

#### Scenario: Alias with zero mappings
- **WHEN** Alias "new-model" has no AliasMappings
- **THEN** router returns no providers
- **AND** handler returns "model not found" error

### Requirement: Deleting Alias cascades to AliasMappings

The system SHALL cascade delete all AliasMappings when Alias is deleted.

#### Scenario: Delete alias removes mappings
- **WHEN** admin deletes Alias "gpt-4"
- **THEN** system deletes all AliasMappings where alias_id=gpt-4.id
- **AND** no orphan mappings remain

### Requirement: Alias list displays in flat table layout

The system SHALL display aliases in a flat table layout (not collapsible panels), consistent with Providers page style.

#### Scenario: Alias list displays as flat table
- **WHEN** admin views `/aliases` page
- **THEN** system displays table with columns: Select, Name, Mapping Count, Token Summary, Status, Actions
- **AND** each row represents one Alias
- **AND** no collapsible panels are used

#### Scenario: Token summary displays minimum values with tooltip
- **WHEN** alias has enabled mappings
- **THEN** system calculates min_context_window from enabled mappings
- **AND** system calculates min_max_output from enabled mappings
- **AND** Token Summary column shows formatted display (e.g., "8K / 4K")
- **AND** mouse hover displays original values (e.g., "128,000 / 4,096")
- **WHEN** alias has no enabled mappings
- **THEN** Token Summary column shows "-"

#### Scenario: Capabilities intersection displays correctly
- **WHEN** alias has enabled mappings
- **THEN** system calculates capability intersection for all enabled mappings
- **AND** supports_vision is true only if all enabled mappings support vision
- **AND** supports_tools is true only if all enabled mappings support tools
- **AND** supports_stream is true only if all enabled mappings support stream
- **AND** Capabilities column shows tags for true capabilities only
- **WHEN** alias has no enabled mappings
- **THEN** Capabilities column shows "-"

#### Scenario: Token summary formatted with one decimal
- **WHEN** system formats min_context_window and min_max_output
- **THEN** values < 1000 display as original number
- **AND** values >= 1000 display as "XK" or "X.XK" (e.g., "128K", "153.6K")
- **AND** values >= 1000000 display as "XM" or "X.XM" (e.g., "2M", "1.5M")

### Requirement: Alias list supports batch operations

The system SHALL allow batch selection and batch deletion of aliases.

#### Scenario: Select multiple aliases
- **WHEN** admin clicks checkboxes in alias table
- **THEN** system tracks selected alias IDs
- **AND** "Batch Delete" button shows count (e.g., "Batch Delete (3)")

#### Scenario: Batch delete aliases
- **WHEN** admin clicks "Batch Delete" with selected aliases
- **THEN** system shows confirmation dialog
- **AND** upon confirmation, system deletes all selected aliases and their mappings
- **AND** system refreshes alias table

#### Scenario: Batch delete disabled without selection
- **WHEN** no aliases are selected
- **THEN** "Batch Delete" button is disabled

### Requirement: Alias list provides detail page navigation

The system SHALL provide "Detail" button to navigate to alias detail page for mapping management.

#### Scenario: Navigate to alias detail page
- **WHEN** admin clicks "Detail" button on alias row
- **THEN** system navigates to `/aliases/:id` page
- **AND** detail page shows complete mapping information

#### Scenario: Detail button visible for all aliases
- **WHEN** alias table is displayed
- **THEN** every alias row has "Detail" button in Actions column
- **AND** "Detail" button is always enabled regardless of mapping count