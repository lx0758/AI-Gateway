# MCP Proxy - Technical Design

## Context

The AI Gateway currently supports OpenAI and Anthropic API protocols for model inference. The Model Context Protocol (MCP) is an emerging standard for exposing tools, resources, and prompts to AI models. This design adds MCP proxy capabilities to the gateway, allowing clients to access multiple MCP services through a unified endpoint.

**Current State:**
- Gateway exposes `/openai/v1/*` and `/anthropic/v1/*` endpoints
- API Keys control access to models via `KeyModel` associations
- Providers are configured with base URLs and API keys
- All model routing uses the `Alias` → `AliasMapping` → `Provider` system

**Constraints:**
- Must integrate with existing API Key authentication system
- Must support both remote (HTTP/SSE) and local (stdio) MCP services
- Self-implementation without external SDK dependencies (learning objective)
- SQLite and PostgreSQL compatibility required

## Goals / Non-Goals

**Goals:**
- Implement MCP protocol endpoint (`/mcp/v1`) with JSON-RPC 2.0 support
- Support HTTP and SSE transport for clients
- Aggregate resources from multiple MCP services with namespace isolation
- Provide fine-grained access control per API Key
- Enable on-demand connection to backend MCP services
- Cache MCP resources in database to minimize connection overhead
- Maintain clean code architecture with PO/DTO separation and interface abstraction

**Non-Goals:**
- Long-lived connection pooling (use on-demand strategy)
- Process supervision for local MCP servers (no daemon/restart logic)
- MCP resource subscription/real-time updates (initial version)
- Load balancing or failover between MCP services
- MCP protocol versioning (assume latest stable version)

## Decisions

### 1. Connection Strategy: On-Demand vs Long-Lived

**Decision:** On-demand connections with database caching

**Rationale:**
- **Scalability:** Gateway may serve many clients; maintaining long connections to all MCP services would be resource-intensive
- **Simplicity:** No connection state management, health checks, or reconnection logic
- **Sufficiency:** Local stdio connections have minimal overhead; HTTP connections can use keep-alive
- **Cache-first:** Database caching means most operations (list tools/resources/prompts) don't need connections

**Timeouts:**
- Connection timeout: 5 seconds
- Request timeout: 30 seconds
- Process startup timeout: 10 seconds
- Idle timeout: 5 minutes (for local processes)

**Alternatives Considered:**
- **Long-lived connection pool:** Higher complexity, resource overhead, but lower latency for tool calls
- **Hybrid approach:** Could be added later if specific MCP services show high-frequency usage patterns

### 2. Namespace Design: Name-Based Identifiers

**Decision:** Use `{name}.{resource}` format for tools/prompts and `mcp://{name}/{uri}` for resources

**Rationale:**
- **Simplified design:** Single `name` field serves as both identifier and namespace
- **Collision avoidance:** Database unique constraint on `name` field
- **Clarity:** Clients explicitly specify which service's resource they want
- **Routing:** Gateway can parse name to route to correct MCP service
- **User-friendly:** Names (2-200 chars) are easier to remember than symbols

**Example:**
```
Tool: filesystem.read_file → routes to "filesystem" MCP → calls "read_file"
Resource: mcp://db/file:///data/config.json → routes to "db" MCP → reads "file:///data/config.json"
```

**Evolution:**
- Initially planned separate `name` and `symbol` fields
- Consolidated to single `name` field with symbol validation rules
- Reduces redundancy and simplifies API

**Alternatives Considered:**
- **Dual fields:** Name + Symbol (rejected for redundancy)
- **UUID-based:** Guaranteed unique but not human-readable
- **Full service name:** Clear but verbose (e.g., `my-filesystem-service.read_file`)

### 3. Permission Granularity: Fine-Grained Control

**Decision:** API Keys map to individual tools/resources/prompts via association tables

**Rationale:**
- **Security principle:** Least privilege access
- **Flexibility:** Different keys can access different subsets of resources
- **Consistency:** Matches existing `KeyModel` pattern

**Data Model:**
```
Key → KeyMCPTool → MCPTool
Key → KeyMCPResource → MCPResource
Key → KeyMCPPrompt → MCPPrompt
```

**Alternatives Considered:**
- **Service-level permissions:** Simpler but less flexible (all-or-nothing access)
- **Role-based:** More complex, overkill for current needs

### 4. Transport Implementation: Self-Implemented

**Decision:** Implement JSON-RPC 2.0 and transport layers from scratch

**Rationale:**
- **Learning objective:** Project goal is to understand AI protocols
- **Control:** Full visibility into protocol behavior for debugging
- **Simplicity:** MCP protocol is well-specified; no need for heavy SDKs

**Implementation:**
- `internal/mcp/protocol.go`: JSON-RPC 2.0 data structures and helpers
- `internal/mcp/client.go`: MCPClient interface definition
- `internal/mcp/client_remote.go`: HTTP/SSE client implementation
- `internal/mcp/client_local.go`: stdio process client implementation
- `internal/mcp/manager.go`: Connection and routing logic

**Protocol Version:** `2025-03-26` (latest stable)

### 5. Client Interface Design: Abstraction Layer

**Decision:** Define `MCPClient` interface with RemoteMCPClient and LocalMCPClient implementations

**Rationale:**
- **Polymorphism:** Manager can work with any client type through interface
- **Extensibility:** Easy to add new client types (e.g., WebSocket, gRPC)
- **Testing:** Can mock clients for unit tests
- **Type safety:** Compile-time checking of client methods

**Interface Methods:**
```go
type MCPClient interface {
    Initialize() (*JSONRPCResponse, error)
    ListTools() (*JSONRPCResponse, error)
    CallTool(name string, arguments map[string]interface{}) (*JSONRPCResponse, error)
    ListResources() (*JSONRPCResponse, error)
    ReadResource(uri string) (*JSONRPCResponse, error)
    ListPrompts() (*JSONRPCResponse, error)
    GetPrompt(name string, arguments map[string]interface{}) (*JSONRPCResponse, error)
    Close() error
}
```

### 6. Data Model Simplification: Target and Params

**Decision:** Consolidate 4 configuration fields into 2: `target` and `params`

**Rationale:**
- **Reduce redundancy:** Both remote and local types have similar field pairs
- **Clearer semantics:** `target` = where to connect, `params` = how to configure
- **Easier maintenance:** Single source of truth for each concern

**Field Mapping:**
| Type | Target | Params |
|------|--------|--------|
| remote | URL | Headers (YAML/JSON) |
| local | Command | EnvVars (YAML/JSON) |

**Benefits:**
- Cleaner database schema
- Simpler frontend forms (dynamic labels based on type)
- Easier to extend for new types

**Alternatives Considered:**
- **Four separate fields:** Clearer but more verbose
- **JSON blob for all config:** Less structured, harder to validate

### 7. PO/DTO Separation: Clean Architecture

**Decision:** Separate Persistence Objects (PO) from Data Transfer Objects (DTO)

**Rationale:**
- **Separation of concerns:** Database models vs API models
- **Flexibility:** Can change database schema without breaking API contracts
- **Security:** Can hide sensitive fields or add computed fields in DTOs
- **Validation:** Different validation rules for persistence vs transfer

**Implementation:**
- **PO (Persistence Objects):**
  - Located in `internal/model/db.go`
  - Named with exported names (e.g., `MCP`, `MCPTool`)
  - All have explicit `TableName()` methods
  - Contain GORM tags and database constraints

- **DTO (Data Transfer Objects):**
  - Located in handler files
  - Named with private names (e.g., `mcpCreateRequest`, `mcpResponse`)
  - Contain JSON tags and validation rules
  - Separate types for create/update/response operations

**Example:**
```go
// PO in model/db.go
type MCP struct {
    ID     uint   `gorm:"primaryKey"`
    Name   string `gorm:"uniqueIndex;size:200;not null"`
    Target string `gorm:"type:text"`
    // ...
}

// DTO in handler/mcp.go
type mcpCreateRequest struct {
    Name   string `json:"name" binding:"required"`
    Target string `json:"target"`
    // ...
}
```

### 8. Local MCP Process Lifecycle

**Decision:** On-demand start, no supervision, idle timeout

**Rationale:**
- **Simplicity:** No process manager complexity
- **Resource efficiency:** Processes exit when not in use
- **Failure handling:** Client receives error if process crashes; manual intervention for persistent issues

**Process Flow:**
1. Client calls tool from local MCP service
2. Gateway checks if process is running
3. If not, start process with configured command
4. Establish stdio connection
5. Forward request
6. Process exits after idle timeout (5 minutes)

**Alternatives Considered:**
- **Persistent processes:** Higher resource usage, need supervision
- **Pre-started processes:** Doesn't scale with many services

### 9. SSE Support Strategy

**Decision:** Support both HTTP and SSE transport for clients and backend services

**Rationale:**
- **Client compatibility:** Some MCP clients prefer SSE for real-time features
- **Backend compatibility:** Some MCP servers only support SSE
- **Priority:** Prefer SSE when available for better streaming support

**Implementation:**
- Client endpoint: Auto-detect SSE requests (Accept: text/event-stream)
- Backend client: Check MCP service capabilities, parse SSE responses

### 10. Database Schema Design

**Decision:** Separate tables for tools/resources/prompts with service foreign key

**Rationale:**
- **Query efficiency:** Can query resources by type without filtering
- **Type safety:** Each resource type has appropriate fields (InputSchema for tools, URI for resources)
- **Clear associations:** Permission tables reference specific resource types

**Tables:**
- `mcps`: Service configuration (name, type, target, params, enabled, capabilities)
- `mcp_tools`, `mcp_resources`, `mcp_prompts`: Cached resource definitions with `enabled` field
- `key_mcp_tools`, `key_mcp_resources`, `key_mcp_prompts`: Permission associations

**All Tables Have:**
- Explicit `TableName()` methods in Go
- Consistent naming (lowercase with underscores)
- Foreign key relationships properly defined

### 11. Resource Enable/Disable Control

**Decision:** Each tool, resource, and prompt has an independent `enabled` field

**Rationale:**
- **Granular control:** Administrators can disable specific resources without deleting them
- **Status persistence:** Enabled status survives re-synchronization
- **Flexibility:** Different environments may need different subsets of resources enabled
- **Clear intent:** Disabled resources are explicitly marked, not implicitly unavailable

**Implementation:**
- **Data Model**: Added `enabled` boolean field to `MCPTool`, `MCPResource`, `MCPPrompt` (default: `true`)
- **Sync Strategy**: Name-based matching during synchronization
  - **Tools**: Match by `mcp_id` + `name`
  - **Resources**: Match by `mcp_id` + `name` (changed from URI matching to preserve enabled status)
  - **Prompts**: Match by `mcp_id` + `name`
- **Permission Assignment**: Only enabled resources shown in API Key permission configuration
- **Frontend Filtering**: Automatically filters out disabled resources when fetching available resources

**Benefits:**
- **Non-destructive**: Disabling a resource doesn't lose configuration
- **Reversible**: Can easily re-enable resources without re-syncing
- **Selective**: Choose which resources to expose to different API Keys

**Example Scenario:**
1. MCP service provides 10 tools
2. Admin disables 3 tools that are experimental
3. Sync operation updates tool metadata but preserves disabled status
4. API Key permission configuration only shows 7 enabled tools
5. Admin can re-enable tools when ready

## Risks / Trade-offs

### Risk: MCP Service Unavailability
**Impact:** Tool calls fail, poor user experience
**Mitigation:**
- Return clear JSON-RPC errors with service name
- Admin can test connectivity before enabling service
- Database caching allows listing resources even if service is down

### Risk: Process Management Complexity
**Impact:** Local MCP servers may hang, consume resources
**Mitigation:**
- Idle timeout (5 minutes) for automatic cleanup
- Process kill on gateway shutdown
- Clear error messages to clients

### Risk: Namespace Collision
**Impact:** Users might choose duplicate names
**Mitigation:**
- Database unique constraint on `name` field
- Frontend validation before save (2-200 chars, `[0-9a-zA-Z_-]`)
- Clear error message if collision detected

### Risk: Performance Overhead
**Impact:** On-demand connections add latency to tool calls
**Mitigation:**
- HTTP keep-alive for remote services
- Fast stdio for local services (minimal overhead)
- Most operations use database cache (no connection needed)

### Trade-off: No Real-Time Updates
**Consequence:** Resource changes require manual sync or scheduled sync
**Mitigation:**
- Admin can trigger manual sync
- Future: Add scheduled sync or MCP subscription support

### Trade-off: No Process Supervision
**Consequence:** Crashed local servers need manual restart
**Mitigation:**
- Automatic restart on next request
- Log errors for debugging
- Future: Add optional supervision for critical services

## Migration Plan

### Phase 1: Infrastructure ✅
1. Database migration: Create MCP tables
2. Basic `/mcp/v1` endpoint with JSON-RPC 2.0 handler
3. Implement `initialize`, `tools/list`, `resources/list`, `prompts/list` returning empty arrays
4. Add API Key authentication middleware

### Phase 2: Service Management ✅
1. Admin API: CRUD for MCP services
2. Frontend: MCP service list and configuration forms
3. Support remote and local configuration fields
4. Test connection functionality

### Phase 3: MCP Client & Sync ✅
1. Implement MCP client (remote HTTP/SSE, local stdio)
2. Resource sync logic (tools/resources/prompts to database)
3. Admin sync trigger endpoint
4. Frontend: Display synced resources in tabs

### Phase 4: Permission Configuration ✅
1. Permission association tables
2. API endpoints for configuring Key permissions
3. Frontend: Add MCP tabs to API Key management page

### Phase 5: Resource Aggregation & Routing ✅
1. Filter resources by Key permissions in `initialize` response
2. Add namespace prefixes to resource identifiers
3. Implement tool call routing and forwarding
4. Implement resource read routing
5. Implement prompt get routing
6. Error handling and logging

### Phase 6: Code Refactoring ✅
1. Remove `symbol` field, apply constraints to `name` field
2. Consolidate configuration fields to `target` and `params`
3. Create `MCPClient` interface for client abstraction
4. Separate PO and DTO with proper naming conventions
5. Add `TableName()` methods to all models
6. Rename frontend folder from `MCPServices` to `MCPs`

### Rollback Strategy
- MCP tables are independent; can be dropped without affecting existing functionality
- `/mcp/v1` endpoint can be disabled via config or removed from router
- No changes to existing model/provider/alias system

## Open Questions

### ✅ 1. Batch Request Support
**Question:** Should `/mcp/v1` support JSON-RPC batch requests (array of requests)?
**Decision:** Yes, implemented for completeness
**Impact:** More complex request handling, but better protocol compliance

### ✅ 2. Resource Subscription
**Question:** Implement `resources/subscribe` for real-time updates?
**Decision:** Defer to future iteration; start with sync-based updates
**Impact:** More complexity, but better real-time behavior

### ✅ 3. Name Validation
**Question:** What characters to allow in names?
**Decision:** `[0-9a-zA-Z_-]`, length 2-200 characters
**Impact:** Simple validation, covers most use cases

### ✅ 4. Timeout Configuration
**Question:** Make connection/request timeouts configurable?
**Decision:** Start with hardcoded values (5s connect, 30s request, 10s process start, 5min idle)
**Impact:** Can add config later if needed

### ✅ 5. Sync Conflict Resolution
**Question:** How to handle deleted resources during sync? Delete or mark unavailable?
**Decision:** Update existing records, create new records (no deletion logic in current implementation)
**Impact:** Simple implementation, can add soft delete later

### ✅ 6. Symbol vs Name
**Question:** Should we keep separate Symbol field or consolidate to Name?
**Decision:** Consolidate to single Name field with Symbol validation rules
**Impact:** Simpler data model, reduced redundancy

### ✅ 7. Field Consolidation
**Question:** Should we keep 4 separate fields (url, headers, command, env_vars) or consolidate?
**Decision:** Consolidate to 2 fields: target and params
**Impact:** Cleaner schema, easier to maintain, dynamic frontend forms
