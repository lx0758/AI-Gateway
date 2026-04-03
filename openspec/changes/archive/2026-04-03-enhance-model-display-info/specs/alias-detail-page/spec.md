## ADDED Requirements

### Requirement: Alias detail page for managing mappings

The system SHALL provide an independent detail page for each alias to manage its mappings with complete information display and drag-drop sorting.

#### Scenario: Access alias detail page
- **WHEN** admin clicks "Detail" button in alias list
- **THEN** system navigates to `/aliases/:id` page
- **AND** page displays alias name in header and info card
- **AND** page displays status as tag (not switch)
- **AND** page displays operations panel with "Add Mapping", "Batch Delete" buttons
- **AND** no "Edit" or "Delete" buttons for the alias itself

#### Scenario: Mappings table displays complete information
- **WHEN** admin views alias detail page
- **THEN** system displays mappings table with columns: Select, Drag Handle, Provider, API Type, Model Name, Capabilities, Context Window, Weight, Status, Actions
- **AND** Capabilities column displays before Context Window column
- **AND** Capabilities shows tags for Vision, Tools, Stream based on model attributes
- **AND** Context Window shows formatted context_window / max_output (e.g., "128K / 4K", "153.6K / 4K")
- **AND** mouse hover on Context Window displays original values

#### Scenario: Token format preserves one decimal place
- **WHEN** system formats token values for display
- **THEN** values < 1000 display as original number (e.g., "512")
- **AND** values >= 1000 display as "XK" or "X.XK" (e.g., 128000 → "128K", 153600 → "153.6K")
- **AND** values >= 1000000 display as "XM" or "X.XM" (e.g., 2000000 → "2M", 1500000 → "1.5M")

#### Scenario: Capabilities tags display with colors
- **WHEN** model has supports_vision=true
- **THEN** system displays Vision tag with type="success" (green)
- **WHEN** model has supports_tools=true
- **THEN** system displays Tools tag with type="warning" (orange)
- **WHEN** model has supports_stream=true
- **THEN** system displays Stream tag with type="primary" (blue)

### Requirement: Drag-drop sorting for mappings

The system SHALL allow admin to reorder mappings by drag-drop, automatically updating weights based on position.

#### Scenario: Drag mapping to new position
- **WHEN** admin drags a mapping row and drops it at new position
- **THEN** system recalculates weights for all mappings
- **AND** weight at position 1 = total_mappings - 1
- **AND** weight at position 2 = total_mappings - 2
- **AND** weight decreases linearly by position
- **AND** weight at last position = 0
- **AND** system calls API to update all mappings weights

#### Scenario: Drag-drop updates display immediately
- **WHEN** drag-drop completes and API returns success
- **THEN** system refreshes mappings table with new weights
- **AND** weight column displays updated values
- **AND** system shows notification "Weights updated successfully"

#### Scenario: Drag-drop preserves enabled status
- **WHEN** admin reorders mappings by drag-drop
- **THEN** enabled status of each mapping remains unchanged
- **AND** only weight values are updated

### Requirement: Batch operations on mappings

The system SHALL allow batch selection and batch deletion of mappings in alias detail page.

#### Scenario: Select multiple mappings
- **WHEN** admin clicks checkboxes in mappings table
- **THEN** system tracks selected mapping IDs
- **AND** "Batch Delete" button shows count of selected items (e.g., "Batch Delete (3)")

#### Scenario: Batch delete mappings
- **WHEN** admin clicks "Batch Delete" button with selected mappings
- **THEN** system shows confirmation dialog
- **AND** upon confirmation, system deletes all selected mappings
- **AND** system refreshes mappings table

#### Scenario: Batch delete disabled without selection
- **WHEN** no mappings are selected
- **THEN** "Batch Delete" button is disabled
- **AND** button shows "Batch Delete (0)"

### Requirement: Add and edit mapping in detail page

The system SHALL allow adding new mappings and editing existing mappings in alias detail page.

#### Scenario: Add new mapping
- **WHEN** admin clicks "Add Mapping" button
- **THEN** system shows dialog with form fields: Provider, Model, Weight, Status
- **AND** Provider dropdown is filterable and lists all providers
- **AND** Model dropdown is filterable and lists provider's models after provider selection
- **AND** Weight field defaults to auto-calculated value based on current mappings count
- **AND** Status switch defaults to enabled

#### Scenario: Edit existing mapping
- **WHEN** admin clicks "Edit" button on a mapping row
- **THEN** system shows dialog with current values pre-filled
- **AND** admin can modify Provider, Model, Weight, Status
- **AND** upon submit, system updates mapping and refreshes table

#### Scenario: Delete single mapping
- **WHEN** admin clicks "Delete" button on a mapping row
- **THEN** system shows confirmation dialog
- **AND** upon confirmation, system deletes mapping and refreshes table

### Requirement: Toggle mapping status in detail page

The system SHALL allow toggling enabled/disabled status for each mapping directly in the table.

#### Scenario: Toggle mapping status
- **WHEN** admin clicks status switch in mappings table
- **THEN** system calls API to update mapping.enabled
- **AND** switch reflects new status immediately
- **AND** no confirmation dialog is needed

#### Scenario: Disabled mapping excluded from routing
- **WHEN** mapping.enabled=false
- **THEN** router excludes this mapping from provider selection
- **AND** mapping still visible in UI with status showing "Disabled"