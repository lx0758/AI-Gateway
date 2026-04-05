# Implementation Tasks

## 1. Database Schema ✅

- [x] 1.1 Create MCP data models in `server/internal/model/db.go`
  - `MCP`: ID, Name (unique), Type, Target, Params, Enabled, Capabilities, LastSyncAt, Timestamps
  - `MCPTool`: ID, MCPID, Name, Description, InputSchema, Timestamps
  - `MCPResource`: ID, MCPID, URI, Name, Description, MimeType, Timestamps
  - `MCPPrompt`: ID, MCPID, Name, Description, Arguments, Timestamps
  - `KeyMCPTool`: ID, KeyID, ToolID, CreatedAt
  - `KeyMCPResource`: ID, KeyID, ResourceID, CreatedAt
  - `KeyMCPPrompt`: ID, KeyID, PromptID, CreatedAt
- [x] 1.2 Add database migration in `server/internal/model/db.go` autoMigrate function
- [x] 1.3 Test migration with `make dev` or manual test
- [x] 1.4 Add `TableName()` methods to all models for explicit table naming
  - `MCP` → `mcps`
  - `MCPTool` → `mcp_tools`
  - `MCPResource` → `mcp_resources`
  - `MCPPrompt` → `mcp_prompts`
  - `KeyMCPTool` → `key_mcp_tools`
  - `KeyMCPResource` → `key_mcp_resources`
  - `KeyMCPPrompt` → `key_mcp_prompts`
- [x] 1.5 Add `TableName()` methods to existing models for consistency
  - `User` → `users`
  - `Provider` → `providers`
  - `ProviderModel` → `provider_models`
  - `Alias` → `aliases`
  - `AliasMapping` → `alias_mappings`
  - `Mapping` → `mappings`
  - `Key` → `keys`
  - `KeyModel` → `key_models`
  - `UsageLog` → `usage_logs`

## 2. MCP Protocol Implementation ✅

- [x] 2.1 Create `server/internal/mcp/protocol.go`
  - Define JSON-RPC 2.0 data structures (Request, Response, Error)
  - Implement JSON-RPC message parsing and validation
  - Handle batch requests (array of requests)
- [x] 2.2 Implement JSON-RPC error codes
  - Parse error (-32700)
  - Invalid request (-32600)
  - Method not found (-32601)
  - Invalid params (-32602)
  - Internal error (-32603)
- [x] 2.3 Add helper functions
  - `NewResponse()` for success responses
  - `NewErrorResponse()` for error responses
  - `NewNotification()` for notifications
  - `IsNotification()` to check if request is notification
  - `MustMarshalJSON()` for JSON marshaling helper

## 3. MCP Server Endpoint ✅

- [x] 3.1 Create `server/internal/handler/mcp.go` (proxy_mcp.go)
  - Implement POST `/mcp/v1` handler
  - Implement JSON-RPC method routing
  - Add API Key authentication middleware
- [x] 3.2 Implement MCP methods
  - `initialize`: Return capabilities with filtered tools/resources/prompts
  - `tools/list`: Return filtered tools with name prefixes
  - `tools/call`: Route to MCP service and forward request
  - `resources/list`: Return filtered resources with name prefixes
  - `resources/read`: Route to MCP service and forward request
  - `resources/subscribe`: Return method not found error
  - `prompts/list`: Return filtered prompts with name prefixes
  - `prompts/get`: Route to MCP service and forward request
  - `ping`: Return empty response
- [x] 3.3 Add route in `server/cmd/server/main.go`
  - Register `/mcp/v1` endpoint
  - Apply API Key middleware
- [x] 3.4 Test endpoint with curl
  - Valid JSON-RPC request
  - Invalid JSON-RPC request
  - Authentication failures
- [x] 3.5 Implement SSE support
  - Detect SSE requests (Accept header)
  - Implement SSE response writer
  - Format JSON-RPC messages as SSE events

## 4. MCP Service Management API ✅

- [x] 4.1 Create `server/internal/handler/mcp.go`
  - `Create(c *gin.Context)`: POST /api/v1/mcps
  - `List(c *gin.Context)`: GET /api/v1/mcps
  - `Get(c *gin.Context)`: GET /api/v1/mcps/:id
  - `Update(c *gin.Context)`: PUT /api/v1/mcps/:id
  - `Delete(c *gin.Context)`: DELETE /api/v1/mcps/:id
  - `TestConnection(c *gin.Context)`: POST /api/v1/mcps/:id/test
  - `Sync(c *gin.Context)`: POST /api/v1/mcps/:id/sync
- [x] 4.2 Add routes in `server/cmd/server/main.go`
- [x] 4.3 Implement name validation
  - Length 2-200 characters
  - Allowed characters: [0-9a-zA-Z_-]
  - Return HTTP 400 for invalid names
- [x] 4.4 Implement CRUD operations with database
  - Create, Read, Update, Delete MCP records
  - Handle cascade delete for tools/resources/prompts
- [x] 4.5 Implement resource listing endpoints
  - `ListTools(c *gin.Context)`: GET /api/v1/mcps/:id/tools
  - `ListResources(c *gin.Context)`: GET /api/v1/mcps/:id/resources
  - `ListPrompts(c *gin.Context)`: GET /api/v1/mcps/:id/prompts

## 5. MCP Client Implementation ✅

- [x] 5.1 Create `server/internal/mcp/client.go`
  - Define `MCPClient` interface
  - Define all required methods (Initialize, ListTools, CallTool, etc.)
- [x] 5.2 Create `server/internal/mcp/client_remote.go`
  - Implement `RemoteMCPClient` struct
  - Implement HTTP client for remote MCP services
  - Support custom headers from params field
  - Implement connection timeout (5s)
  - Implement request timeout (30s)
  - Parse SSE responses
- [x] 5.3 Create `server/internal/mcp/client_local.go`
  - Implement `LocalMCPClient` struct
  - Implement stdio client for local MCP services
  - Execute command from target field
  - Use params field for environment variables
  - Implement process communication via stdin/stdout
  - Implement process startup timeout (10s)
  - Implement idle timeout (5 minutes)
  - Log stderr output
- [x] 5.4 Create `server/internal/mcp/manager.go`
  - Implement connection manager (on-demand connections)
  - Use `MCPClient` interface for all operations
  - Implement request routing based on name
  - Implement resource sync logic
  - Handle connection reuse and cleanup
- [x] 5.5 Update manager to use target/params fields
  - Remote: target = URL, params = headers
  - Local: target = command, params = env_vars
- [x] 5.6 Test MCP clients
  - Verified with compilation and build
  - Tested with real MCP servers (manual)

## 6. Resource Sync Implementation ✅

- [x] 6.1 Implement sync logic in `server/internal/mcp/manager.go`
  - Connect to MCP service
  - Call `initialize` to get capabilities
  - Call `tools/list`, `resources/list`, `prompts/list`
  - Update database with discovered resources
- [x] 6.2 Implement sync endpoint in handler
  - Trigger sync from admin API
  - Return sync result (success/failure)
  - Update LastSyncAt timestamp
- [x] 6.3 Handle sync errors
  - Log errors without failing
  - Preserve cached resources on failure
- [x] 6.4 Test sync with real MCP server
  - Manual testing performed
  - Database updates verified

## 7. MCP Protocol Full Implementation ✅

- [x] 7.1 Implement `tools/call` in handler
  - Parse tool name to extract MCP name
  - Check API Key permission
  - Route to MCP service
  - Forward request with original tool name
  - Return response to client
- [x] 7.2 Implement `resources/read` in handler
  - Parse resource URI to extract MCP name
  - Check API Key permission
  - Route to MCP service
  - Forward request with original URI
  - Return response to client
- [x] 7.3 Implement `prompts/get` in handler
  - Parse prompt name to extract MCP name
  - Check API Key permission
  - Route to MCP service
  - Forward request with original prompt name
  - Return response to client
- [x] 7.4 Update `initialize` method
  - Query API Key permissions
  - Filter tools/resources/prompts by permissions
  - Add name prefixes to all identifiers
  - Return filtered list with prefixes

## 8. API Key MCP Resource Configuration ✅

- [x] 8.1 Create permission configuration endpoints in `server/internal/handler/key.go`
  - `GetMCPTools(c *gin.Context)`: GET /api/v1/keys/:id/mcp-tools
  - `UpdateMCPTools(c *gin.Context)`: PUT /api/v1/keys/:id/mcp-tools
  - `GetMCPResources(c *gin.Context)`: GET /api/v1/keys/:id/mcp-resources
  - `UpdateMCPResources(c *gin.Context)`: PUT /api/v1/keys/:id/mcp-resources
  - `GetMCPPrompts(c *gin.Context)`: GET /api/v1/keys/:id/mcp-prompts
  - `UpdateMCPPrompts(c *gin.Context)`: PUT /api/v1/keys/:id/mcp-prompts
- [x] 8.2 Add routes in `server/cmd/server/main.go`
- [x] 8.3 Implement permission checking
  - Verify tool/resource/prompt exists
  - Verify API Key exists
  - Handle invalid IDs gracefully
- [x] 8.4 Update DTO to remove Symbol field
  - Remove `symbol` from response structs
  - Update all handler code
- [x] 8.5 Test permission APIs
  - Manual testing performed
  - Verified permission enforcement

## 9. SSE Transport Support ✅

- [x] 9.1 Add SSE support to MCP endpoint
  - Detect SSE requests (Accept header)
  - Implement SSE response writer
  - Format JSON-RPC messages as SSE events
- [x] 9.2 Add SSE client support in `client_remote.go`
  - Establish SSE connection to backend MCP services
  - Parse SSE events as JSON-RPC messages
  - Handle SSE responses
- [x] 9.3 Test SSE transport
  - Manual testing performed
  - Message format verified

## 10. Frontend: MCP Service Management ✅

- [x] 10.1 Create MCP service list page
  - Display services in table
  - Show name, type, status, lastSyncAt
  - Add create/edit/delete buttons
- [x] 10.2 Create MCP service form
  - Name input with validation (2-200 chars, [0-9a-zA-Z_-])
  - Type selector (remote/local)
  - Target input (URL for remote, command for local)
  - Params textarea (headers for remote, env_vars for local)
  - Enabled checkbox
- [x] 10.3 Create MCP service detail page with tabs
  - Tools tab: List tools with name, description, input schema
  - Resources tab: List resources with URI, name
  - Prompts tab: List prompts with name, description, arguments
  - Add sync button
  - Add test connection button
- [x] 10.4 Implement sync trigger
  - Call sync API
  - Show sync progress
  - Display sync results
- [x] 10.5 Add resource detail view
  - JsonViewer component for input schema/arguments
  - Show full JSON schema/arguments
  - Add copy JSON button
- [x] 10.6 Update forms for target/params fields
  - Dynamic labels based on type
  - Remote: URL label, Headers label
  - Local: Command label, Environment Variables label
- [x] 10.7 Add i18n translations for new fields
  - target, params translations
  - Update existing translations

## 11. Frontend: API Key MCP Configuration ✅

- [x] 11.1 Add MCP tabs to API Key management page
  - MCP Tools tab
  - MCP Resources tab
  - MCP Prompts tab
- [x] 11.2 Create resource selector component
  - Search/filter resources
  - Select/deselect resources
  - Show resource details
- [x] 11.3 Implement permission save
  - Call update API on save
  - Show success/error feedback
- [x] 11.4 Update display to use MCP name instead of symbol
  - Change from `${mcp.symbol}.${tool.name}` to `${mcp.name}.${tool.name}`
- [x] 11.5 Add i18n translations
  - MCP-related UI text
  - Error messages

## 12. Code Refactoring ✅

- [x] 12.1 Remove Symbol field from MCP model
  - Remove from database model
  - Remove from DTOs (request/response)
  - Apply Symbol validation rules to Name field
- [x] 12.2 Consolidate configuration fields
  - Remove: url, headers, command, env_vars
  - Add: target, params
  - Update all handler code
  - Update manager code
- [x] 12.3 Create MCPClient interface
  - Define interface in `mcp/client.go`
  - Update RemoteMCPClient to implement interface
  - Update LocalMCPClient to implement interface
  - Update manager to use interface instead of concrete types
- [x] 12.4 Separate PO and DTO
  - Move all PO (database models) to `model/db.go`
  - Define DTO in handler files with private naming
  - Remove `po` and `dto` package references
  - Update all imports
- [x] 12.5 Add TableName() methods
  - Add to all MCP-related models
  - Add to all existing models for consistency
  - Remove inline tableName tags where applicable
- [x] 12.6 Rename frontend folder
  - Rename `views/MCPServices` to `views/MCPs`
  - Update router imports
- [x] 12.7 Update i18n files
  - Remove symbol-related translations
  - Add name-related translations (namePlaceholder, nameInvalid)
  - Update for target/params fields
- [x] 12.8 Build verification
  - Backend: `go build` successful
  - Frontend: `npm run build` successful
  - No TypeScript errors
  - No LSP errors

## 13. Integration Testing ✅

- [x] 13.1 Test end-to-end MCP flow
  - Create MCP service
  - Sync resources
  - Configure API Key permissions
  - Call initialize with API Key
  - Call tool with API Key
  - Verify name prefix
- [x] 13.2 Test permission enforcement
  - Try accessing tool without permission
  - Try accessing resource without permission
  - Verify error responses
- [x] 13.3 Test connection failures
  - Unavailable remote service
  - Crashing local process
  - Timeout scenarios
  - Verify error handling
- [x] 13.4 Test with real MCP servers
  - Manual testing with filesystem MCP server
  - Verify tool execution
  - Verify resource reading

## 14. Resource Control ✅

- [x] 14.1 Add `enabled` field to MCP data models
  - `MCPTool`: Add `Enabled bool` field with default `true`
  - `MCPResource`: Add `Enabled bool` field with default `true`
  - `MCPPrompt`: Add `Enabled bool` field with default `true`
- [x] 14.2 Update sync logic for name-based matching
  - Tools: Match by `mcp_id` + `name` (already implemented)
  - Resources: Changed from URI matching to `name` matching
  - Prompts: Match by `mcp_id` + `name` (already implemented)
- [x] 14.3 Ensure new resources are created with `enabled=true`
  - Updated `syncTools` to set `Enabled: true`
  - Updated `syncResources` to set `Enabled: true`
  - Updated `syncPrompts` to set `Enabled: true`
- [x] 14.4 Update response DTOs to include `enabled` field
  - `mcpToolResponse`: Added `Enabled bool` field
  - `mcpResourceResponse`: Added `Enabled bool` field
  - `mcpPromptResponse`: Added `Enabled bool` field
- [x] 14.5 Update List APIs to return `enabled` status
  - `ListTools`: Returns tools with `enabled` field
  - `ListResources`: Returns resources with `enabled` field
  - `ListPrompts`: Returns prompts with `enabled` field
- [x] 14.6 Frontend filtering for permission assignment
  - Updated `fetchAvailableMCPTools` to filter `enabled=true`
  - Updated `fetchAvailableMCPResources` to filter `enabled=true`
  - Updated `fetchAvailableMCPPrompts` to filter `enabled=true`
- [x] 14.7 Build verification
  - Backend: `go build` successful
  - Frontend: `npm run build` successful
- [x] 14.8 Add management API for updating resource status
  - `UpdateTool`: PUT `/mcps/tools/:id` - Update tool enabled status
  - `UpdateResource`: PUT `/mcps/resources/:id` - Update resource enabled status
  - `UpdatePrompt`: PUT `/mcps/prompts/:id` - Update prompt enabled status
- [x] 14.9 Register API routes in `main.go`
  - Added PUT routes for tools, resources, and prompts
- [x] 14.10 Add enable/disable UI in MCP detail page
  - Added status column with switch for tools table
  - Added status column with switch for resources table
  - Added status column with switch for prompts table
- [x] 14.11 Implement toggle functions in frontend
  - `toggleToolEnabled`: Update tool status via API
  - `toggleResourceEnabled`: Update resource status via API
  - `togglePromptEnabled`: Update prompt status via API
- [x] 14.12 Build verification with UI
  - Backend: `go build` successful
  - Frontend: `npm run build` successful

## 15. Documentation ✅

- [x] 14.1 Update proposal.md
  - Reflect all implemented features
  - Document data model changes
  - Update capabilities list
- [x] 14.2 Update design.md
  - Document all decisions made
  - Add new decisions (Symbol removal, field consolidation, interface abstraction)
  - Update trade-offs and risks
- [x] 14.3 Update tasks.md
  - Mark all completed tasks
  - Add refactoring tasks
  - Document testing status
- [x] 14.4 Update README.md (if needed)
  - Add MCP proxy feature description
  - Document new field structure (target/params)
  - Add usage examples
