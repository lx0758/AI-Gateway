# MCP Proxy

## Summary

Implement MCP (Model Context Protocol) proxy functionality in the AI Gateway, allowing the gateway to act as a unified MCP server for clients while managing multiple backend MCP services (both remote and local). The implementation includes complete backend API, frontend UI, and code structure optimizations for maintainability.

## Motivation

The AI Gateway currently supports OpenAI and Anthropic API proxying. To extend its capabilities and support the emerging MCP ecosystem, we need to add MCP protocol support. This allows:

1. **Unified MCP Access**: Clients can connect to a single MCP endpoint and access tools/resources/prompts from multiple backend MCP services
2. **Fine-grained Access Control**: API Keys can be configured with specific permissions for individual tools, resources, and prompts
3. **Service Abstraction**: Backend MCP services can be added/removed without client changes
4. **Hybrid Deployment**: Support both remote MCP services (HTTP/SSE) and local MCP servers (stdio)

## Capabilities

### New Capabilities

- `mcp-protocol-endpoint`: MCP protocol endpoint at `/mcp/v1` implementing JSON-RPC 2.0 over HTTP/SSE with API Key authentication
- `mcp-service-management`: Administrative CRUD operations for managing MCP service configurations (remote and local)
- `mcp-resource-sync`: Synchronization of tools, resources, and prompts from backend MCP services to local database cache with name-based matching to preserve enabled status
- `mcp-client-connection`: On-demand connection management for both remote (HTTP/SSE) and local (stdio) MCP services with unified interface
- `api-key-mcp-resources`: Extend API Key management to support fine-grained permissions for MCP tools, resources, and prompts
- `mcp-resource-control`: Independent enable/disable control for each tool, resource, and prompt within an MCP service

### Modified Capabilities

- `api-key-management`: Add MCP resource permission configuration to existing API Key management (only enabled resources are available for permission assignment)
- `data-model-optimization`: Consolidated MCP fields (target/params) and removed redundant Symbol field
- `code-structure-improvement`: PO/DTO separation, interface abstraction, and table name consistency

## Implementation Highlights

### Data Model Refactoring

**Before**: 4 separate fields (url, headers, command, env_vars) + Symbol field
**After**: 2 unified fields (target, params) + Name field with Symbol constraints

- **target**: Service endpoint (URL for remote, command for local)
- **params**: Configuration parameters (headers for remote, env_vars for local)
- **name**: Unique identifier with Symbol validation rules (2-200 chars, `[0-9a-zA-Z_-]`)

### Code Architecture

1. **PO/DTO Separation**:
   - PO (Persistence Objects) in `model/db.go` with `TableName()` methods
   - DTO (Data Transfer Objects) in handler files with private naming

2. **Client Interface Abstraction**:
   - `MCPClient` interface in `mcp/client.go`
   - `RemoteMCPClient` and `LocalMCPClient` implementations
   - Unified client management in `MCPManager`

3. **Naming Consistency**:
   - Backend: `MCP` model, `/api/v1/mcps` endpoints
   - Frontend: `MCPs` view folder
   - Database: `mcps` table

### MCP Protocol

- **Version**: `2025-03-26` (upgraded from `2024-11-05`)
- **Implementation**: Self-contained JSON-RPC 2.0 without external SDK dependencies
- **Transport**: Support both HTTP and SSE for clients and backend services

### Resource Control

**Fine-grained Control**: Each tool, resource, and prompt can be independently enabled/disabled

- **Data Model**: Added `enabled` field to `MCPTool`, `MCPResource`, `MCPPrompt` tables (default: `true`)
- **Sync Strategy**: Name-based matching during synchronization preserves `enabled` status
  - Tools/Prompts: Match by `name` field
  - Resources: Match by `name` field (changed from `uri` matching)
- **Permission Assignment**: Only enabled resources are available for API Key permission configuration
- **Frontend Filtering**: Automatically filters out disabled resources when assigning permissions

**Benefits**:
- Administrators can temporarily disable specific tools/resources/prompts without deleting them
- Enabled status persists across re-synchronization operations
- Clear separation between service availability and resource usability

## Impact

### Backend
- **Database**: 
  - Tables: `mcps`, `mcp_tools`, `mcp_resources`, `mcp_prompts`, `key_mcp_tools`, `key_mcp_resources`, `key_mcp_prompts`
  - Removed `symbol` field, added validation to `name` field
  - Consolidated 4 fields to 2 (`target`, `params`)
  - Added `enabled` field to `mcp_tools`, `mcp_resources`, `mcp_prompts` for independent resource control
- **API Endpoints**: 
  - `/mcp/v1` - MCP protocol endpoint (JSON-RPC 2.0)
  - `/api/v1/mcps` - MCP service management API
  - `/api/v1/keys/:id/mcp-*` - API Key MCP resource configuration
- **Code Structure**: 
  - `internal/mcp/` package with interface-based client abstraction
  - PO in `model/db.go`, DTO in handler files
  - All models have explicit `TableName()` methods

### Frontend
- **MCPs Page**: Management page for configuring MCP services with tool/resource/prompt tabs
- **API Key Page**: Extended with 3 new tabs for MCP tool/resource/prompt permissions
- **Folder Structure**: Renamed `MCPServices` to `MCPs` for consistency

### Dependencies
- No external SDK dependencies (self-implementation for learning purposes)
- JSON-RPC 2.0 implementation
- SSE (Server-Sent Events) support for real-time communication
- Process management for local MCP servers

### Performance
- On-demand connection strategy to minimize resource usage
- Database caching to avoid frequent MCP service queries
- Configurable timeouts: 5s connection, 30s request, 10s process startup, 5min idle

## Testing

- ✅ Backend compilation and build verification
- ✅ Frontend build and TypeScript type checking
- ✅ Database migration with SQLite/PostgreSQL compatibility
- ✅ Resource enable/disable control implemented
- ✅ Name-based sync matching preserves enabled status
- ✅ Frontend filtering for permission assignment
- ✅ Management UI for enable/disable resources
- ✅ API endpoints for updating resource status
- ⏳ Integration tests with real MCP servers (pending)
