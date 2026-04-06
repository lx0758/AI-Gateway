## 1. Backend - Key Model Endpoints

- [x] 1.1 Add GET `/keys/:id` endpoint to return single Key basic info
- [x] 1.2 Modify GET `/keys/:id/models` to return all models with `selected` status
- [x] 1.3 Add POST `/keys/:id/models/:model_id` for single model association
- [x] 1.4 Add DELETE `/keys/:id/models/:model_id` for single model association removal
- [x] 1.5 Add DELETE `/keys/:id/models` for clearing all model associations

## 2. Backend - MCP Tools Endpoints

- [x] 2.1 Modify GET `/keys/:id/mcp-tools` to return all tools with `selected` status (filter disabled)
- [x] 2.2 Add POST `/keys/:id/mcp-tools/:tool_id` for single tool association
- [x] 2.3 Add DELETE `/keys/:id/mcp-tools/:tool_id` for single tool association removal
- [x] 2.4 Add DELETE `/keys/:id/mcp-tools` for clearing all tool associations

## 3. Backend - MCP Resources Endpoints

- [x] 3.1 Modify GET `/keys/:id/mcp-resources` to return all resources with `selected` status (filter disabled)
- [x] 3.2 Add POST `/keys/:id/mcp-resources/:resource_id` for single resource association
- [x] 3.3 Add DELETE `/keys/:id/mcp-resources/:resource_id` for single resource association removal
- [x] 3.4 Add DELETE `/keys/:id/mcp-resources` for clearing all resource associations

## 4. Backend - MCP Prompts Endpoints

- [x] 4.1 Modify GET `/keys/:id/mcp-prompts` to return all prompts with `selected` status (filter disabled)
- [x] 4.2 Add POST `/keys/:id/mcp-prompts/:prompt_id` for single prompt association
- [x] 4.3 Add DELETE `/keys/:id/mcp-prompts/:prompt_id` for single prompt association removal
- [x] 4.4 Add DELETE `/keys/:id/mcp-prompts` for clearing all prompt associations

## 5. Backend - Keys List Endpoint

- [x] 5.1 Modify GET `/keys` to include `mcp_tools_count`, `mcp_resources_count`, `mcp_prompts_count` in response

## 6. Backend - Router Registration

- [x] 6.1 Register all new endpoints in router

## 7. Frontend - Keys List Page Refactor

- [x] 7.1 Modify "允许模型" column to display "不限制" or "仅允许 N 个"
- [x] 7.2 Add new columns: MCP工具, MCP资源, MCP提示词 with same display logic
- [x] 7.3 Add "详情" button in action column
- [x] 7.4 Modify edit dialog to only show name input field
- [x] 7.5 Remove MCP tabs from edit dialog

## 8. Frontend - Key Detail Page Creation

- [x] 8.1 Create `web/src/views/Keys/Detail.vue` file
- [x] 8.2 Implement basic info card with el-descriptions
- [x] 8.3 Add 4 tabs: Models, MCP工具, MCP资源, MCP提示词
- [x] 8.4 Implement Models tab with radio group (○ 默认 / ○ 仅允许)
- [x] 8.5 Implement MCP Tools tab with radio group (filter disabled)
- [x] 8.6 Implement MCP Resources tab with radio group (filter disabled)
- [x] 8.7 Implement MCP Prompts tab with radio group (filter disabled)
- [x] 8.8 Add "全部允许" button above each tab table
- [x] 8.9 Implement radio toggle API calls (POST/DELETE single association)
- [x] 8.10 Implement "全部允许" button API call (DELETE all associations)
- [x] 8.11 Add back navigation to return to Keys list

## 9. Frontend - Router Configuration

- [x] 9.1 Add `/keys/:id` route pointing to Key Detail page

## 10. Testing

- [ ] 10.1 Test GET `/keys/:id/models` returns correct selected status
- [ ] 10.2 Test POST/DELETE single model association
- [ ] 10.3 Test DELETE `/keys/:id/models` clears all associations
- [ ] 10.4 Test MCP tools/resources/prompts endpoints similarly
- [ ] 10.5 Test GET `/keys` returns MCP counts correctly
- [ ] 10.6 Test frontend list displays counts correctly
- [ ] 10.7 Test detail page radio toggle functionality
- [ ] 10.8 Test "全部允许" button clears associations

**Note**: All implementation tasks (1-36) are complete and code compiles successfully. Testing tasks require running the application and manual verification.