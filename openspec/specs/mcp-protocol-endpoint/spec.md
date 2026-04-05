# MCP Protocol Endpoint

## ADDED Requirements

### Requirement: JSON-RPC 2.0 Endpoint
The system SHALL provide a JSON-RPC 2.0 endpoint at `/mcp/v1` that accepts POST requests with JSON-RPC 2.0 formatted messages.

#### Scenario: Valid JSON-RPC request
- **WHEN** client sends POST request to `/mcp/v1` with valid JSON-RPC 2.0 message
- **THEN** system returns JSON-RPC 2.0 response with result or error

#### Scenario: Invalid JSON-RPC request
- **WHEN** client sends malformed JSON-RPC request
- **THEN** system returns JSON-RPC error response with code -32700 (Parse error)

### Requirement: API Key Authentication
The system SHALL authenticate MCP requests using the existing API Key system. Clients MUST provide API Key via `Authorization: Bearer sk-xxx` header.

#### Scenario: Valid API Key
- **WHEN** client provides valid API Key in Authorization header
- **THEN** system processes the MCP request

#### Scenario: Missing API Key
- **WHEN** client does not provide API Key
- **THEN** system returns HTTP 401 Unauthorized

#### Scenario: Invalid API Key
- **WHEN** client provides invalid or disabled API Key
- **THEN** system returns HTTP 401 Unauthorized

### Requirement: Initialize Method
The system SHALL implement the `initialize` method that returns server capabilities and available resources based on the API Key's permissions.

#### Scenario: Initialize with permissions
- **WHEN** client calls `initialize` with valid API Key
- **THEN** system returns capabilities object with tools, resources, and prompts the API Key has access to

#### Scenario: Initialize with expired API Key
- **WHEN** client calls `initialize` with expired API Key
- **THEN** system returns HTTP 401 Unauthorized

### Requirement: Tools List Method
The system SHALL implement the `tools/list` method that returns all tools the API Key has permission to use, with namespace-prefixed names.

#### Scenario: List available tools
- **WHEN** client calls `tools/list` with valid API Key
- **THEN** system returns array of tools with names prefixed as `{symbol}.{tool_name}`

#### Scenario: List tools with no permissions
- **WHEN** client calls `tools/list` with API Key that has no tool permissions
- **THEN** system returns empty array

### Requirement: Tools Call Method
The system SHALL implement the `tools/call` method that routes tool execution to the appropriate MCP service.

#### Scenario: Call tool with valid permissions
- **WHEN** client calls `tools/call` with tool name `{symbol}.{tool_name}` and valid arguments
- **THEN** system routes request to the MCP service identified by symbol and returns result

#### Scenario: Call tool without permission
- **WHEN** client calls `tools/call` with tool name they don't have permission for
- **THEN** system returns JSON-RPC error with code -32602 (Invalid params)

#### Scenario: Call non-existent tool
- **WHEN** client calls `tools/call` with tool name that doesn't exist
- **THEN** system returns JSON-RPC error with code -32602 (Invalid params)

### Requirement: Resources List Method
The system SHALL implement the `resources/list` method that returns all resources the API Key has permission to access.

#### Scenario: List available resources
- **WHEN** client calls `resources/list` with valid API Key
- **THEN** system returns array of resources with URIs prefixed as `mcp://{symbol}/{original_uri}`

### Requirement: Resources Read Method
The system SHALL implement the `resources/read` method that reads resource content from the appropriate MCP service.

#### Scenario: Read resource with valid permissions
- **WHEN** client calls `resources/read` with URI `mcp://{symbol}/{original_uri}`
- **THEN** system routes request to MCP service and returns resource content

### Requirement: Prompts List Method
The system SHALL implement the `prompts/list` method that returns all prompts the API Key has permission to use.

#### Scenario: List available prompts
- **WHEN** client calls `prompts/list` with valid API Key
- **THEN** system returns array of prompts with names prefixed as `{symbol}.{prompt_name}`

### Requirement: Prompts Get Method
The system SHALL implement the `prompts/get` method that retrieves prompt template from the appropriate MCP service.

#### Scenario: Get prompt with valid permissions
- **WHEN** client calls `prompts/get` with prompt name `{symbol}.{prompt_name}`
- **THEN** system routes request to MCP service and returns prompt template

### Requirement: SSE Transport Support
The system SHALL support Server-Sent Events (SSE) transport for clients that prefer real-time streaming.

#### Scenario: SSE connection request
- **WHEN** client sends GET request to `/mcp/v1` with `Accept: text/event-stream` header
- **THEN** system establishes SSE connection for bidirectional communication

#### Scenario: SSE message format
- **WHEN** system sends JSON-RPC message over SSE
- **THEN** message is formatted as `data: {json-rpc-message}\n\n`

### Requirement: Error Handling
The system SHALL return standard JSON-RPC 2.0 error codes for all error conditions.

#### Scenario: Service unavailable
- **WHEN** MCP service is unavailable during tool call
- **THEN** system returns JSON-RPC error with code -32603 (Internal error) and descriptive message

#### Scenario: Timeout
- **WHEN** MCP service does not respond within timeout period
- **THEN** system returns JSON-RPC error with code -32603 (Internal error) and timeout message
