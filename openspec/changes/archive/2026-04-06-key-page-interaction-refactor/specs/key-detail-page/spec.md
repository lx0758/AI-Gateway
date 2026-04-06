## ADDED Requirements

### Requirement: Key Detail Page Structure

The system SHALL provide a dedicated detail page for viewing and managing individual API Key permissions.

#### Scenario: Access detail page
- **WHEN** admin clicks "Detail" button in Keys list
- **THEN** system navigates to `/keys/:id` route
- **AND** displays Key detail page with basic info card and 4 permission tabs

#### Scenario: Display basic info
- **WHEN** admin opens Key detail page
- **THEN** system displays basic info card with: name, masked key, enabled status, expires_at, created_at

#### Scenario: Permission tabs layout
- **WHEN** admin views Key detail page
- **THEN** system displays 4 tabs: Models, MCP Tools, MCP Resources, MCP Prompts

---

### Requirement: Single-radio Permission Control

The system SHALL allow administrators to control Key permissions using radio buttons with "Default" and "Allow Only" states.

#### Scenario: Radio button states
- **WHEN** admin views any permission tab
- **THEN** each component row displays radio group: ○ Default  ○ Allow Only

#### Scenario: Default state meaning
- **WHEN** radio is set to "Default"
- **THEN** component is NOT in Key's permission association table
- **AND** Key can access all components of this type (empty association = allow all)

#### Scenario: Allow Only state meaning
- **WHEN** radio is set to "Allow Only"
- **THEN** component IS in Key's permission association table
- **AND** Key can only access components in the association table

#### Scenario: Switch to Allow Only
- **WHEN** admin clicks "Allow Only" radio for a component
- **THEN** frontend calls POST `/keys/:id/<type>/<component_id>`
- **AND** backend creates association record
- **AND** radio stays selected on success

#### Scenario: Switch to Default
- **WHEN** admin clicks "Default" radio for a component
- **THEN** frontend calls DELETE `/keys/:id/<type>/<component_id>`
- **AND** backend deletes association record
- **AND** radio stays selected on success

#### Scenario: API call failure
- **WHEN** permission toggle API call fails
- **THEN** frontend reverts radio to previous state
- **AND** displays error message to user

---

### Requirement: Allow All Button

The system SHALL provide "Allow All" button for each permission tab to quickly clear all restrictions.

#### Scenario: Allow All button position
- **WHEN** admin views any permission tab
- **THEN** system displays "Allow All" button above the component table

#### Scenario: Click Allow All
- **WHEN** admin clicks "Allow All" button
- **THEN** frontend calls DELETE `/keys/:id/<type>`
- **AND** backend deletes all association records for that component type
- **AND** all radio buttons in tab switch to "Default" state
- **AND** frontend displays success message

#### Scenario: Allow All failure
- **WHEN** Allow All API call fails
- **THEN** frontend displays error message
- **AND** radio buttons remain in current state

---

### Requirement: Permission Tab Data Loading

The system SHALL load permission tab data on-demand when switching tabs.

#### Scenario: Initial tab load
- **WHEN** admin opens Key detail page
- **THEN** system loads basic info immediately
- **AND** loads first tab (Models) data automatically

#### Scenario: Switch to new tab
- **WHEN** admin clicks on unvisited tab (e.g., MCP Tools)
- **THEN** system calls GET `/keys/:id/mcp-tools`
- **AND** displays loading indicator
- **AND** renders component table with radio states when data arrives

#### Scenario: Tab data caching
- **WHEN** admin switches back to previously visited tab
- **THEN** system displays cached data without new API call
- **AND** radio states reflect current server state

---

### Requirement: Component Table Display

The system SHALL display component information with permission state in each permission tab.

#### Scenario: Models tab display
- **WHEN** admin views Models tab
- **THEN** table displays: model name, permission radio group
- **AND** each row has radio set to "Allow Only" if model is in KeyModel table
- **AND** each row has radio set to "Default" if model is NOT in KeyModel table

#### Scenario: MCP Tools tab display
- **WHEN** admin views MCP Tools tab
- **THEN** table displays: tool name (with MCP prefix), permission radio group
- **AND** disabled tools (MCP.enabled=false or Tool.enabled=false) are NOT shown

#### Scenario: MCP Resources tab display
- **WHEN** admin views MCP Resources tab
- **THEN** table displays: resource name (with MCP prefix), permission radio group
- **AND** disabled resources are NOT shown

#### Scenario: MCP Prompts tab display
- **WHEN** admin views MCP Prompts tab
- **THEN** table displays: prompt name (with MCP prefix), permission radio group
- **AND** disabled prompts are NOT shown

---

### Requirement: Navigate Back to List

The system SHALL allow admin to navigate back to Keys list from detail page.

#### Scenario: Back button
- **WHEN** admin clicks back navigation in detail page header
- **THEN** system navigates to `/keys` list page