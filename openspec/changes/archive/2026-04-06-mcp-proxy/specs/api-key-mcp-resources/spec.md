# API Key MCP Resources

## ADDED Requirements

### Requirement: Tool Permission Configuration
The system SHALL allow administrators to configure which MCP tools each API Key can access.

#### Scenario: Grant tool access
- **WHEN** administrator grants API Key access to MCP tool
- **THEN** system creates KeyMCPTool association

#### Scenario: Revoke tool access
- **WHEN** administrator removes API Key's access to MCP tool
- **THEN** system deletes KeyMCPTool association

#### Scenario: List key tools
- **WHEN** administrator requests tools for an API Key
- **THEN** system returns all tools the key has access to

### Requirement: Resource Permission Configuration
The system SHALL allow administrators to configure which MCP resources each API Key can access.

#### Scenario: Grant resource access
- **WHEN** administrator grants API Key access to MCP resource
- **THEN** system creates KeyMCPResource association

#### Scenario: Revoke resource access
- **WHEN** administrator removes API Key's access to MCP resource
- **THEN** system deletes KeyMCPResource association

### Requirement: Prompt Permission Configuration
The system SHALL allow administrators to configure which MCP prompts each API Key can access.

#### Scenario: Grant prompt access
- **WHEN** administrator grants API Key access to MCP prompt
- **THEN** system creates KeyMCPPrompt association

#### Scenario: Revoke prompt access
- **WHEN** administrator removes API Key's access to MCP prompt
- **THEN** system deletes KeyMCPPrompt association

### Requirement: RESTful API Design
The system SHALL provide consistent RESTful endpoints for configuring MCP resource permissions.

#### Scenario: GET MCP tools for key
- **WHEN** administrator calls GET /api/v1/api-keys/:id/mcp-tools
- **THEN** system returns list of tool IDs the key can access

#### Scenario: PUT MCP tools for key
- **WHEN** administrator calls PUT /api/v1/api-keys/:id/mcp-tools with array of tool IDs
- **THEN** system replaces key's tool permissions with new list

#### Scenario: GET MCP resources for key
- **WHEN** administrator calls GET /api/v1/api-keys/:id/mcp-resources
- **THEN** system returns list of resource IDs the key can access

#### Scenario: PUT MCP resources for key
- **WHEN** administrator calls PUT /api/v1/api-keys/:id/mcp-resources with array of resource IDs
- **THEN** system replaces key's resource permissions with new list

#### Scenario: GET MCP prompts for key
- **WHEN** administrator calls GET /api/v1/api-keys/:id/mcp-prompts
- **THEN** system returns list of prompt IDs the key can access

#### Scenario: PUT MCP prompts for key
- **WHEN** administrator calls PUT /api/v1/api-keys/:id/mcp-prompts with array of prompt IDs
- **THEN** system replaces key's prompt permissions with new list

### Requirement: Permission Enforcement
The system SHALL enforce MCP resource permissions when clients call the MCP endpoint.

#### Scenario: Access granted tool
- **WHEN** client calls tool they have permission for
- **THEN** system routes request to MCP service

#### Scenario: Access denied tool
- **WHEN** client calls tool they don't have permission for
- **THEN** system returns JSON-RPC error with code -32602

#### Scenario: Access granted resource
- **WHEN** client reads resource they have permission for
- **THEN** system routes request to MCP service

#### Scenario: Access denied resource
- **WHEN** client reads resource they don't have permission for
- **THEN** system returns JSON-RPC error with code -32602

### Requirement: Initialize Response Filtering
The system SHALL filter resources in `initialize` response based on API Key permissions.

#### Scenario: Filtered tool list
- **WHEN** client calls initialize with API Key
- **THEN** system returns only tools the key has permission for, with namespace prefixes

#### Scenario: Filtered resource list
- **WHEN** client calls initialize with API Key
- **THEN** system returns only resources the key has permission for, with modified URIs

#### Scenario: Filtered prompt list
- **WHEN** client calls initialize with API Key
- **THEN** system returns only prompts the key has permission for, with namespace prefixes

### Requirement: Namespace Prefix Addition
The system SHALL add namespace prefix to all resource identifiers returned to clients.

#### Scenario: Tool name prefix
- **WHEN** system returns tool in initialize response
- **THEN** tool name is prefixed as `{symbol}.{original_name}`

#### Scenario: Resource URI prefix
- **WHEN** system returns resource in initialize response
- **THEN** resource URI is prefixed as `mcp://{symbol}/{original_uri}`

#### Scenario: Prompt name prefix
- **WHEN** system returns prompt in initialize response
- **THEN** prompt name is prefixed as `{symbol}.{original_name}`

### Requirement: Cascade Delete
The system SHALL handle permission cleanup when resources are deleted.

#### Scenario: Tool deleted from service
- **WHEN** MCP tool is removed from database (service deleted)
- **THEN** system deletes all KeyMCPTool associations for that tool

#### Scenario: API Key deleted
- **WHEN** API Key is deleted
- **THEN** system deletes all KeyMCPTool, KeyMCPResource, KeyMCPPrompt associations for that key
