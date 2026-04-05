# MCP Client Connection

## ADDED Requirements

### Requirement: On-Demand Connection Strategy
The system SHALL establish connections to MCP services only when needed (tool call, resource read, prompt get, or sync), not maintain persistent connections.

#### Scenario: First tool call
- **WHEN** client calls tool from MCP service for first time
- **THEN** system establishes connection, executes tool, and keeps connection for idle timeout period

#### Scenario: Subsequent call within timeout
- **WHEN** client makes another call to same MCP service within idle timeout
- **THEN** system reuses existing connection

#### Scenario: Call after timeout
- **WHEN** client calls tool after idle timeout has expired
- **THEN** system establishes new connection

### Requirement: Remote MCP Client (HTTP/SSE)
The system SHALL implement HTTP and SSE client for connecting to remote MCP services.

#### Scenario: HTTP connection
- **WHEN** connecting to remote MCP service via HTTP
- **THEN** system sends JSON-RPC request via POST and receives JSON-RPC response

#### Scenario: SSE connection
- **WHEN** connecting to remote MCP service that supports SSE
- **THEN** system establishes SSE connection and streams JSON-RPC messages

#### Scenario: Custom headers
- **WHEN** remote MCP service requires custom headers
- **THEN** system includes configured headers in HTTP requests

#### Scenario: Connection timeout
- **WHEN** remote service doesn't respond within connection timeout (5 seconds)
- **THEN** system returns timeout error to caller

### Requirement: Local MCP Client (stdio)
The system SHALL implement stdio client for connecting to local MCP services via process execution.

#### Scenario: Process startup
- **WHEN** connecting to local MCP service
- **THEN** system executes configured command with environment variables

#### Scenario: stdio communication
- **WHEN** communicating with local MCP process
- **THEN** system sends JSON-RPC via stdin and receives JSON-RPC via stdout

#### Scenario: Process error handling
- **WHEN** local MCP process writes to stderr
- **THEN** system logs stderr output for debugging

#### Scenario: Process startup timeout
- **WHEN** local MCP process doesn't respond within startup timeout (10 seconds)
- **THEN** system terminates process and returns error to caller

### Requirement: Process Lifecycle Management
The system SHALL manage local MCP process lifecycle with idle timeout but no supervision.

#### Scenario: Process idle timeout
- **WHEN** local MCP process is idle for 5 minutes
- **THEN** system terminates the process

#### Scenario: Process crash
- **WHEN** local MCP process crashes during request
- **THEN** system returns error to client and doesn't auto-restart

#### Scenario: Manual restart after crash
- **WHEN** client makes new request after process crashed
- **THEN** system starts new process instance

### Requirement: Request Timeout
The system SHALL enforce request timeout (30 seconds) for all MCP operations.

#### Scenario: Tool call timeout
- **WHEN** tool execution exceeds 30 seconds
- **THEN** system returns timeout error to client

#### Scenario: Resource read timeout
- **WHEN** resource read exceeds 30 seconds
- **THEN** system returns timeout error to client

### Requirement: Connection Error Propagation
The system SHALL return clear JSON-RPC errors when connection failures occur.

#### Scenario: Service unavailable
- **WHEN** MCP service is unreachable
- **THEN** system returns JSON-RPC error with code -32603 and message "MCP service unavailable: {symbol}"

#### Scenario: Authentication failure
- **WHEN** remote MCP service rejects connection due to invalid credentials
- **THEN** system returns JSON-RPC error with code -32603 and message "MCP service authentication failed"

### Requirement: JSON-RPC Protocol Handling
The system SHALL correctly implement JSON-RPC 2.0 protocol for all MCP communications.

#### Scenario: Request ID correlation
- **WHEN** system sends JSON-RPC request with ID
- **THEN** system correlates response with same ID

#### Scenario: Notification handling
- **WHEN** MCP service sends notification (no ID)
- **THEN** system processes notification without expecting response

#### Scenario: Batch request support
- **WHEN** client sends array of JSON-RPC requests
- **THEN** system processes each request and returns array of responses
