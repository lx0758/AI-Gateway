## ADDED Requirements

### Requirement: Record MCP tool call

The system SHALL record each MCP tool call to `mcp_logs` table with the following information:
- `source`: "mcp-proxy"
- `client_ips`: Client IP and forwarding chain (comma-separated)
- `key_id`: The API key ID used for the request
- `key_name`: The API key name
- `mcp_id`: The MCP service ID
- `mcp_name`: The MCP service name
- `mcp_type`: The MCP service type (remote/local)
- `call_type`: "tool"
- `call_method`: "call"
- `call_target`: The tool module name
- `input_size`: Input arguments size in bytes (int64)
- `output_size`: Output content size in bytes (int64)
- `latency_ms`: Call latency in milliseconds (int64)
- `status`: "success" or "error"
- `error_msg`: Error message if failed, empty if success
- `created_at`: Timestamp of log creation

#### Scenario: Successful tool call

- **WHEN** a successful MCP tool call is completed via tools/call method
- **THEN** system records a MCP log with call_type="tool", call_method="call", status="success", input_size from input args, output_size from response content, and latency_ms calculated

#### Scenario: Failed tool call

- **WHEN** an MCP tool call fails with an error
- **THEN** system records a MCP log with call_type="tool", call_method="call", status="error", error_msg containing the error, input_size and output_size as 0

### Requirement: Record MCP resource read

The system SHALL record each MCP resource read to `mcp_logs` table with the following information:
- Same fields as tool call, except:
- `call_type`: "resource"
- `call_method`: "read"
- `call_target`: The resource module name

#### Scenario: Successful resource read

- **WHEN** a successful MCP resource read is completed via resources/read method
- **THEN** system records a MCP log with call_type="resource", call_method="read", status="success", input_size as 0, output_size from response content, and latency_ms calculated

#### Scenario: Failed resource read

- **WHEN** an MCP resource read fails with an error
- **THEN** system records a MCP log with call_type="resource", call_method="read", status="error", error_msg containing the error, input_size and output_size as 0

### Requirement: Record MCP prompt get

The system SHALL record each MCP prompt get to `mcp_logs` table with the following information:
- Same fields as tool call, except:
- `call_type`: "prompt"
- `call_method`: "get"
- `call_target`: The prompt module name

#### Scenario: Successful prompt get

- **WHEN** a successful MCP prompt get is completed via prompts/get method
- **THEN** system records a MCP log with call_type="prompt", call_method="get", status="success", input_size from arguments, output_size from response content, and latency_ms calculated

#### Scenario: Failed prompt get

- **WHEN** an MCP prompt get fails with an error
- **THEN** system records a MCP log with call_type="prompt", call_method="get", status="error", error_msg containing the error, input_size and output_size as 0

### Requirement: Do not record metadata queries

The system SHALL NOT record MCP logs for metadata query calls including:
- initialize method
- tools/list method
- resources/list method
- prompts/list method
- ping method

#### Scenario: Metadata query not logged

- **WHEN** an MCP metadata query is performed (initialize, list methods, or ping)
- **THEN** system does NOT create a MCP log entry

### Requirement: Calculate input and output sizes

The system SHALL calculate and record input_size and output_size for each MCP call:
- `input_size`: Size in bytes of the input data (arguments for tools/call and prompts/get, 0 for resources/read)
- `output_size`: Size in bytes of the output content (response content from MCP server)

#### Scenario: Calculate input size for tool call

- **WHEN** a tool call is recorded
- **THEN** system calculates input_size as the byte size of the JSON-encoded input arguments

#### Scenario: Calculate output size for resource read

- **WHEN** a resource read is recorded
- **THEN** system calculates output_size as the byte size of the response content

### Requirement: Calculate latency

The system SHALL calculate latency_ms for each MCP call by measuring the time from call initiation to completion.

#### Scenario: Calculate latency

- **WHEN** an MCP call is initiated
- **THEN** system records the start time
- **AND** after the call completes, system calculates latency_ms as the difference between completion time and start time in milliseconds

### Requirement: Database indexes

The system SHALL create database indexes for the `mcp_logs` table on the following columns:
- `key_id`: For querying logs by API key
- `mcp_id`: For querying logs by MCP service
- `call_type`: For querying logs by call type (tool/resource/prompt)
- `created_at`: For time range queries

#### Scenario: Query logs by key

- **WHEN** user queries MCP logs by key_id
- **THEN** the query uses the key_id index for efficient lookup

#### Scenario: Query logs by time range

- **WHEN** user queries MCP logs by created_at time range
- **THEN** the query uses the created_at index for efficient lookup

### Requirement: Query MCP logs API

The system SHALL provide an API endpoint to query MCP logs with time range filtering.

#### Scenario: Query logs with default time range

- **WHEN** user requests GET /api/v1/usage/mcp-logs without parameters
- **THEN** system returns MCP logs for today (from 00:00:00 to 23:59:59)

#### Scenario: Query logs with custom time range

- **WHEN** user requests GET /api/v1/usage/mcp-logs with start_date and end_date parameters
- **THEN** system returns MCP logs within the specified time range
- **AND** response includes logs sorted by created_at descending

#### Scenario: Query logs response format

- **WHEN** user queries MCP logs successfully
- **THEN** response contains a "logs" array with each log including: id, source, client_ips, key_id, key_name, mcp_id, mcp_name, mcp_type, call_type, call_target, call_method, input_size, output_size, latency_ms, status, error_msg, created_at

### Requirement: Display MCP usage statistics

The system SHALL display MCP usage statistics on the MCPUsage page with the following components:

#### Scenario: Display summary statistics

- **WHEN** user opens the MCPUsage page
- **THEN** system displays summary cards showing: total requests, success rate, total data size (input + output), average latency

#### Scenario: Display statistics by source

- **WHEN** user views the MCPUsage page
- **THEN** system displays a table showing statistics grouped by source (接入点统计)
- **AND** each row shows: source, call count, input size, output size, average latency

#### Scenario: Display statistics by IP

- **WHEN** user views the MCPUsage page
- **THEN** system displays a table showing statistics grouped by client_ips (IP 统计)
- **AND** each row shows: client IP (first IP in chain), full IP chain, call count, input size, output size, average latency

#### Scenario: Display statistics by call type

- **WHEN** user views the MCPUsage page
- **THEN** system displays a table showing statistics grouped by call_type (调用类型统计)
- **AND** each row shows: call type (tool/resource/prompt), call count, input size, output size, average latency

#### Scenario: Display statistics by Key

- **WHEN** user views the MCPUsage page
- **THEN** system displays a table showing statistics grouped by key_name (Key 统计)
- **AND** each row shows: key name, call count, input size, output size, average latency

#### Scenario: Display statistics by MCP service

- **WHEN** user views the MCPUsage page
- **THEN** system displays a table showing statistics grouped by mcp_name (MCP 服务统计)
- **AND** each row shows: MCP service name, call count, input size, output size, average latency

#### Scenario: Display statistics by MCP service type

- **WHEN** user views the MCPUsage page
- **THEN** system displays a table showing statistics grouped by mcp_type (MCP 服务类型统计)
- **AND** each row shows: MCP service type (remote/local), call count, input size, output size, average latency

#### Scenario: Display statistics by MCP service and call type

- **WHEN** user views the MCPUsage page
- **THEN** system displays a table showing statistics grouped by mcp_name and call_type (MCP 服务+调用类型统计)
- **AND** each row shows: MCP service name, call type, call count, input size, output size, average latency

### Requirement: Display MCP logs detail

The system SHALL display MCP logs in a detailed table on the MCPUsage page.

#### Scenario: Display logs table

- **WHEN** user views the MCPUsage page
- **THEN** system displays a table with MCP logs
- **AND** columns include: time, source, client IP, key name, MCP service name, call type, call target, input size, output size, latency, status, error message

#### Scenario: Format data sizes

- **WHEN** displaying input_size and output_size in logs table
- **THEN** system formats bytes into human-readable format (e.g., 1024 bytes → "1 KB")

#### Scenario: Format latency

- **WHEN** displaying latency_ms in logs table
- **THEN** system formats milliseconds into human-readable format (e.g., 1500 ms → "1.5 s")

#### Scenario: Display IP chain tooltip

- **WHEN** a log has multiple IPs in client_ips field
- **THEN** system displays the first IP in the table cell
- **AND** shows a tooltip with the full IP chain on hover

#### Scenario: Display error message

- **WHEN** a log has status="error"
- **THEN** system displays the error_msg in the error column
- **AND** provides a copy button to copy the error message