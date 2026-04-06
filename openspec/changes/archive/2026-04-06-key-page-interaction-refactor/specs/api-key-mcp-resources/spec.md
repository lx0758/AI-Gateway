## ADDED Requirements

### Requirement: Get MCP Tools with Status

The system SHALL return all available MCP tools with permission status for a Key, excluding disabled tools.

#### Scenario: Get tools with selected status
- **WHEN** admin calls GET `/keys/:id/mcp-tools`
- **THEN** system returns array of all enabled tools from enabled MCPs
- **AND** each tool object includes: id, name, mcp_id, mcp_name, description, selected (boolean)
- **AND** selected=true if tool is in KeyMCPTool table for this key
- **AND** selected=false if tool is NOT in KeyMCPTool table

#### Scenario: Filter disabled MCP
- **WHEN** an MCP has enabled=false
- **THEN** tools from that MCP are NOT included in response

#### Scenario: Filter disabled tool
- **WHEN** a tool has enabled=false
- **THEN** that tool is NOT included in response

---

### Requirement: Single MCP Tool Association Operations

The system SHALL support adding and removing single MCP tool association for a Key.

#### Scenario: Add tool association
- **WHEN** admin calls POST `/keys/:id/mcp-tools/:tool_id`
- **THEN** system creates KeyMCPTool record if not exists
- **AND** returns 200 OK

#### Scenario: Add duplicate tool association
- **WHEN** admin calls POST `/keys/:id/mcp-tools/:tool_id` for existing association
- **THEN** system returns 200 OK without error

#### Scenario: Remove tool association
- **WHEN** admin calls DELETE `/keys/:id/mcp-tools/:tool_id`
- **THEN** system deletes KeyMCPTool record if exists
- **AND** returns 200 OK

#### Scenario: Remove non-existent tool association
- **WHEN** admin calls DELETE `/keys/:id/mcp-tools/:tool_id` for non-existent association
- **THEN** system returns 200 OK without error

---

### Requirement: Clear All MCP Tool Associations

The system SHALL support clearing all MCP tool associations for a Key.

#### Scenario: Clear all tools
- **WHEN** admin calls DELETE `/keys/:id/mcp-tools`
- **THEN** system deletes all KeyMCPTool records for this key
- **AND** returns 200 OK with message

---

### Requirement: Get MCP Resources with Status

The system SHALL return all available MCP resources with permission status for a Key, excluding disabled resources.

#### Scenario: Get resources with selected status
- **WHEN** admin calls GET `/keys/:id/mcp-resources`
- **THEN** system returns array of all enabled resources from enabled MCPs
- **AND** each resource object includes: id, name, uri, mcp_id, mcp_name, selected (boolean)
- **AND** selected=true if resource is in KeyMCPResource table for this key
- **AND** selected=false if resource is NOT in KeyMCPResource table

#### Scenario: Filter disabled MCP for resources
- **WHEN** an MCP has enabled=false
- **THEN** resources from that MCP are NOT included in response

#### Scenario: Filter disabled resource
- **WHEN** a resource has enabled=false
- **THEN** that resource is NOT included in response

---

### Requirement: Single MCP Resource Association Operations

The system SHALL support adding and removing single MCP resource association for a Key.

#### Scenario: Add resource association
- **WHEN** admin calls POST `/keys/:id/mcp-resources/:resource_id`
- **THEN** system creates KeyMCPResource record if not exists
- **AND** returns 200 OK

#### Scenario: Add duplicate resource association
- **WHEN** admin calls POST for existing association
- **THEN** system returns 200 OK without error

#### Scenario: Remove resource association
- **WHEN** admin calls DELETE `/keys/:id/mcp-resources/:resource_id`
- **THEN** system deletes KeyMCPResource record if exists
- **AND** returns 200 OK

---

### Requirement: Clear All MCP Resource Associations

The system SHALL support clearing all MCP resource associations for a Key.

#### Scenario: Clear all resources
- **WHEN** admin calls DELETE `/keys/:id/mcp-resources`
- **THEN** system deletes all KeyMCPResource records for this key
- **AND** returns 200 OK with message

---

### Requirement: Get MCP Prompts with Status

The system SHALL return all available MCP prompts with permission status for a Key, excluding disabled prompts.

#### Scenario: Get prompts with selected status
- **WHEN** admin calls GET `/keys/:id/mcp-prompts`
- **THEN** system returns array of all enabled prompts from enabled MCPs
- **AND** each prompt object includes: id, name, mcp_id, mcp_name, description, selected (boolean)
- **AND** selected=true if prompt is in KeyMCPPrompt table for this key
- **AND** selected=false if prompt is NOT in KeyMCPPrompt table

#### Scenario: Filter disabled MCP for prompts
- **WHEN** an MCP has enabled=false
- **THEN** prompts from that MCP are NOT included in response

#### Scenario: Filter disabled prompt
- **WHEN** a prompt has enabled=false
- **THEN** that prompt is NOT included in response

---

### Requirement: Single MCP Prompt Association Operations

The system SHALL support adding and removing single MCP prompt association for a Key.

#### Scenario: Add prompt association
- **WHEN** admin calls POST `/keys/:id/mcp-prompts/:prompt_id`
- **THEN** system creates KeyMCPPrompt record if not exists
- **AND** returns 200 OK

#### Scenario: Add duplicate prompt association
- **WHEN** admin calls POST for existing association
- **THEN** system returns 200 OK without error

#### Scenario: Remove prompt association
- **WHEN** admin calls DELETE `/keys/:id/mcp-prompts/:prompt_id`
- **THEN** system deletes KeyMCPPrompt record if exists
- **AND** returns 200 OK

---

### Requirement: Clear All MCP Prompt Associations

The system SHALL support clearing all MCP prompt associations for a Key.

#### Scenario: Clear all prompts
- **WHEN** admin calls DELETE `/keys/:id/mcp-prompts`
- **THEN** system deletes all KeyMCPPrompt records for this key
- **AND** returns 200 OK with message