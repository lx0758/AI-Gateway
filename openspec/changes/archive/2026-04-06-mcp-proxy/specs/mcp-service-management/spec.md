# MCP Service Management

## ADDED Requirements

### Requirement: Create MCP Service
The system SHALL allow administrators to create MCP service configurations with name, symbol, type (remote/local), and connection parameters.

#### Scenario: Create remote service
- **WHEN** administrator creates MCP service with type "remote", URL, and optional custom headers
- **THEN** system stores service configuration in database

#### Scenario: Create local service
- **WHEN** administrator creates MCP service with type "local" and command string
- **THEN** system stores service configuration in database

#### Scenario: Duplicate symbol
- **WHEN** administrator creates MCP service with symbol that already exists
- **THEN** system returns HTTP 400 error with message "symbol already exists"

### Requirement: Symbol Validation
The system SHALL validate that service symbols contain only allowed characters `[0-9a-zA-Z_-]` and have length between 2-200 characters.

#### Scenario: Valid symbol
- **WHEN** administrator creates service with symbol "filesystem"
- **THEN** system accepts the symbol

#### Scenario: Invalid symbol characters
- **WHEN** administrator creates service with symbol "file@system"
- **THEN** system returns HTTP 400 error with validation message

#### Scenario: Symbol too short
- **WHEN** administrator creates service with symbol "a"
- **THEN** system returns HTTP 400 error with validation message

### Requirement: List MCP Services
The system SHALL provide an endpoint to list all configured MCP services with their metadata.

#### Scenario: List all services
- **WHEN** administrator requests service list
- **THEN** system returns all services with name, symbol, type, enabled status, and last sync time

### Requirement: Update MCP Service
The system SHALL allow administrators to update MCP service configurations.

#### Scenario: Update service configuration
- **WHEN** administrator updates service URL or command
- **THEN** system stores updated configuration

#### Scenario: Update service symbol
- **WHEN** administrator updates service symbol to a new unique value
- **THEN** system updates symbol and all associated resources reflect new symbol

### Requirement: Delete MCP Service
The system SHALL allow administrators to delete MCP service configurations.

#### Scenario: Delete service
- **WHEN** administrator deletes MCP service
- **THEN** system removes service and all cached tools/resources/prompts

#### Scenario: Delete service with permissions
- **WHEN** administrator deletes MCP service that has API Key permissions configured
- **THEN** system removes service, cached resources, and all permission associations

### Requirement: Test MCP Service Connection
The system SHALL provide a test endpoint to verify MCP service connectivity.

#### Scenario: Test remote service success
- **WHEN** administrator tests remote MCP service that is reachable
- **THEN** system returns success with service capabilities

#### Scenario: Test local service success
- **WHEN** administrator tests local MCP service command that executes successfully
- **THEN** system returns success with service capabilities

#### Scenario: Test service failure
- **WHEN** administrator tests MCP service that is unreachable or fails
- **THEN** system returns error with failure reason

### Requirement: Sync MCP Service Resources
The system SHALL provide a sync endpoint to fetch and cache tools, resources, and prompts from an MCP service.

#### Scenario: Sync service success
- **WHEN** administrator triggers sync for MCP service
- **THEN** system connects to service, fetches all resources, and updates database cache

#### Scenario: Sync detects new tools
- **WHEN** sync discovers new tools not in database
- **THEN** system creates new MCPTool records

#### Scenario: Sync detects removed tools
- **WHEN** sync finds tools in database that no longer exist in service
- **THEN** system marks those tools as unavailable (soft delete)

### Requirement: Service Enable/Disable
The system SHALL allow administrators to enable or disable MCP services without deleting them.

#### Scenario: Disable service
- **WHEN** administrator disables MCP service
- **THEN** system marks service as disabled and it is not included in client resource lists

#### Scenario: Enable service
- **WHEN** administrator enables previously disabled MCP service
- **THEN** system marks service as enabled and it is included in client resource lists
