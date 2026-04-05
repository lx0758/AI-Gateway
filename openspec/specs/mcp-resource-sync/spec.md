# MCP Resource Sync

## ADDED Requirements

### Requirement: Resource Discovery
The system SHALL connect to MCP services and discover available tools, resources, and prompts using MCP protocol methods.

#### Scenario: Discover tools
- **WHEN** system syncs with MCP service
- **THEN** system calls `tools/list` method and caches all tool definitions

#### Scenario: Discover resources
- **WHEN** system syncs with MCP service
- **THEN** system calls `resources/list` method and caches all resource definitions

#### Scenario: Discover prompts
- **WHEN** system syncs with MCP service
- **THEN** system calls `prompts/list` method and caches all prompt definitions

### Requirement: Capability Detection
The system SHALL detect MCP service capabilities during sync using the `initialize` method.

#### Scenario: Detect supported capabilities
- **WHEN** system initializes connection with MCP service
- **THEN** system stores service capabilities (tools, resources, prompts support flags)

#### Scenario: Capability mismatch
- **WHEN** service doesn't support a capability (e.g., no resources)
- **THEN** system notes capability as unsupported and skips related sync steps

### Requirement: Tool Caching
The system SHALL cache tool definitions in database with service association.

#### Scenario: Cache new tool
- **WHEN** sync discovers new tool not in database
- **THEN** system creates MCPTool record with name, description, and input schema

#### Scenario: Update existing tool
- **WHEN** sync discovers tool with same name but different schema
- **THEN** system updates existing MCPTool record

#### Scenario: Remove unavailable tool
- **WHEN** sync finds tool in database that's not in service response
- **THEN** system marks tool as unavailable instead of deleting

### Requirement: Resource Caching
The system SHALL cache resource definitions in database with service association.

#### Scenario: Cache new resource
- **WHEN** sync discovers new resource not in database
- **THEN** system creates MCPResource record with URI, name, description, and MIME type

#### Scenario: Update existing resource
- **WHEN** sync discovers resource with same URI but different metadata
- **THEN** system updates existing MCPResource record

### Requirement: Prompt Caching
The system SHALL cache prompt definitions in database with service association.

#### Scenario: Cache new prompt
- **WHEN** sync discovers new prompt not in database
- **THEN** system creates MCPPrompt record with name, description, and arguments schema

#### Scenario: Update existing prompt
- **WHEN** sync discovers prompt with same name but different arguments
- **THEN** system updates existing MCPPrompt record

### Requirement: Sync Metadata Tracking
The system SHALL track last sync time for each MCP service.

#### Scenario: Update sync timestamp
- **WHEN** sync completes successfully
- **THEN** system updates service's LastSyncAt timestamp

#### Scenario: Sync failure
- **WHEN** sync fails due to connection error
- **THEN** system logs error and preserves previous LastSyncAt value

### Requirement: Manual Sync Trigger
The system SHALL allow administrators to manually trigger sync for individual services.

#### Scenario: Manual sync via API
- **WHEN** administrator calls sync endpoint for a service
- **THEN** system performs immediate sync and returns result

#### Scenario: Sync already in progress
- **WHEN** sync is triggered while previous sync is still running
- **THEN** system returns HTTP 409 Conflict or queues request

### Requirement: Sync Error Handling
The system SHALL handle sync errors gracefully without affecting other services.

#### Scenario: One service fails
- **WHEN** sync fails for one MCP service
- **THEN** system logs error but continues to serve cached resources from other services

#### Scenario: All services fail
- **WHEN** sync fails for all services
- **THEN** system continues to serve previously cached resources
