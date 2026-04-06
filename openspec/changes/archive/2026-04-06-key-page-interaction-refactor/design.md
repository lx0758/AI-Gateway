## Context

当前 Key 页面（`web/src/views/Keys/index.vue`）的编辑弹窗混合了基础信息修改和权限详情管理，用户需要在一个弹窗中切换 4 个 TAB（Models、MCP工具、MCP资源、MCP提示词）来完成权限配置，交互不够清晰。

现有业务逻辑：**"没有关联等于全部允许"**
- `KeyModel` 表为空 → Key 可访问所有模型
- `KeyMCPTool` 表为空 → Key 可访问所有 MCP 工具
- 同理 Resources 和 Prompts

现有前端使用 Transfer 组件批量选择权限，但语义不够直观。后端 `PUT /keys/:id/mcp-tools` 等接口采用整体替换模式。

## Goals / Non-Goals

**Goals:**
- 分离"编辑"和"详情管理"，提供清晰的交互层次
- 列表页只展示基础信息和权限概览
- 新增独立详情页，用单选框交互管理权限
- 后端提供单项操作接口，支持更精细的权限控制
- 提供"全部允许"按钮快速清空关联

**Non-Goals:**
- 不改变现有权限校验逻辑（"空关联=全部允许"保持不变）
- 不修改数据库表结构
- 不影响 MCP Detail 页面结构

## Decisions

### 1. 列表页展示改为"不限制"或"仅允许 N 个"

**决策**: 不再展示具体的模型名称标签，改为显示数量概览。

**理由**:
- 减少视觉干扰，列表页关注"概览"而非"详情"
- 与 MCP 详情页风格一致
- 避免标签过多时的布局问题

**实现**:
```
models.length == 0 → 显示 "不限制"
models.length > 0  → 显示 "仅允许 {N} 个"
```

### 2. 详情页用单选框而非开关

**决策**: 使用单选框（○ 默认 / ○ 仅允许）而非开关组件。

**理由**:
- 单选框语义更明确：用户需要明确选择"默认"或"仅允许"
- 开关组件有歧义：关闭可能被理解为"禁用"而非"默认允许"
- 单选框强制用户思考每个组件的权限状态

**替代方案**: 开关组件（已否决，语义有歧义）

### 3. 后端接口改为单项操作

**决策**: 新增 POST/DELETE 单项操作接口，替代现有的整体替换接口。

**新接口设计**:
```
POST   /keys/:id/models/:model_id        → 添加单个模型关联
DELETE /keys/:id/models/:model_id        → 删除单个模型关联
DELETE /keys/:id/models                  → 清空所有模型关联

POST   /keys/:id/mcp-tools/:tool_id      → 添加单个工具关联
DELETE /keys/:id/mcp-tools/:tool_id      → 删除单个工具关联
DELETE /keys/:id/mcp-tools               → 清空所有工具关联

同理 MCP Resources 和 MCP Prompts
```

**理由**:
- 单项操作语义更清晰，减少请求体大小
- 前端切换单选框时直接调用对应接口，逻辑简单
- "全部允许"按钮调用 DELETE 清空接口，语义明确

**保留**: 现有的 `PUT /keys/:id/mcp-tools` 整体替换接口保留，供将来批量操作使用。

### 4. GET 接口返回全量列表 + selected 状态

**决策**: 改造 `GET /keys/:id/mcp-tools` 等接口，返回全量组件列表（过滤禁用） + selected 状态。

**实现逻辑**:
```go
// 伪代码
func GetMCPTools(c *gin.Context) {
    keyId := c.Param("id")
    
    // 1. 获取所有可用的 MCP 工具（过滤 MCP.enabled=false 和 Tool.enabled=false）
    var allTools []MCPTool
    DB.Joins("MCP").Where("mcp.enabled = ? AND mcp_tools.enabled = ?", true, true).Find(&allTools)
    
    // 2. 获取该 Key 已关联的工具 ID
    var keyToolIds []uint
    DB.Model(&KeyMCPTool{}).Where("key_id = ?", keyId).Pluck("tool_id", &keyToolIds)
    
    // 3. 组装返回数据，添加 selected 状态
    result := make([]ToolWithStatus, len(allTools))
    for i, tool := range allTools {
        result[i] = ToolWithStatus{
            ID:          tool.ID,
            Name:        tool.Name,
            MCPID:       tool.MCPID,
            MCPName:     tool.MCP.Name,
            Description: tool.Description,
            Selected:    contains(keyToolIds, tool.ID),
        }
    }
    
    c.JSON(200, gin.H{"tools": result})
}
```

**理由**:
- 前端一次请求获取所有组件 + 选择状态，无需再调用 `/mcps` 和 `/mcps/*/tools`
- 减少请求次数，提升页面加载速度
- selected 状态由后端准备，前端逻辑更简单

### 5. 前端详情页结构参考 MCP Detail

**决策**: Key Detail 页面结构参考 `web/src/views/MCPs/Detail.vue`，但交互逻辑不同。

**差异点**:
| 页面 | 数据来源 | 开关含义 |
|------|----------|----------|
| MCP Detail | `/mcps/{id}/tools` | 控制组件是否在系统中启用 |
| Key Detail | `/keys/:id/mcp-tools` | 控制 Key 是否被允许访问该组件 |

**实现**:
- 基础信息卡片：使用 `el-descriptions` 展示名称、Key、状态、过期时间
- TAB 卡片：使用 `el-tabs` 切换 4 个组件类型
- 表格上方：每个 TAB 提供"全部允许"按钮
- 表格行：每行用单选框（○ 默认 / ○ 仅允许）

## Risks / Trade-offs

### [Risk] 列表接口性能
前端详情页需要获取全量组件列表，数据量可能较大（假设 50 模型 + 100 工具 + 50 资源 + 25 提示词）。

**Mitigation**: 
- 后端过滤禁用组件，减少返回数量
- 每个 TAB 独立加载，使用 `watch` 懒加载（切换 TAB 时才请求）

### [Risk] 单选框切换的请求频率
用户连续切换多个单选框会产生多次请求。

**Mitigation**:
- 不使用防抖，因为数据分散在 4 个 TAB，不会汇聚大量请求
- 前端切换后立即调用 API，失败时回退单选框状态并提示错误

### [Risk] 后端接口数量增加
新增 8 个单项操作接口 + 4 个清空接口，接口数量增多。

**Mitigation**:
- 语义更清晰，易于理解和使用
- 保留现有的整体替换接口，将来批量操作仍可用

## Migration Plan

本次改造不涉及数据迁移，步骤如下：

1. **后端实现新接口**
   - 改造 `GetMCPTools` 等接口返回全量列表 + selected
   - 新增单项操作接口（POST/DELETE）
   - 新增清空接口（DELETE）
   - 改造 `List` 接口返回 MCP 关联数量

2. **前端改造列表页**
   - 修改展示逻辑（不限制/仅允许 N 个）
   - 添加详情按钮
   - 修改编辑弹窗只保留名称

3. **前端新增详情页**
   - 创建 `Detail.vue`
   - 实现 4 个 TAB 的单选框交互
   - 实现"全部允许"按钮

4. **路由配置**
   - 新增 `/keys/:id` 路由

5. **测试验证**
   - 功能测试：单选框切换、全部允许按钮
   - API 测试：新接口返回正确数据

## Open Questions

暂无