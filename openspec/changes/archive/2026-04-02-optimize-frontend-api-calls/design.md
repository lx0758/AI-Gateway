## Context

当前系统存在前端 API 调用效率问题：

**问题1 - Alias 页面 N+1 请求**:
- `GET /aliases` 只返回 alias 列表基础信息（id, name, enabled, mapping_count）
- 前端为每个 alias 单独调用 `GET /aliases/:id` 获取 mappings
- 假设有 N 个 alias，总请求数 = 1 + N

**问题2 - Keys 页面无谓调用**:
- `Keys/index.vue` 的 `onMounted` 同时调用 `fetchKeys()` 和 `fetchAvailableModels()`
- 但 `availableModels` 仅在编辑对话框中使用
- 每次进入页面都调用，即使从不打开编辑对话框

**问题3 - Key Update Bug**:
- `key.go` 的 `Update` 函数（172-229行）接收 `updateAPIKeyRequest.Models`
- 但完全没有处理 Models 的更新逻辑
- 对比 `Create` 函数（123-129行）正确处理了 Models

## Goals / Non-Goals

**Goals:**
- 消除 Alias 页面的 N+1 API 问题
- 减少 Keys 页面的不必要 API 调用
- 修复 API Key 更新时 Models 配置不生效的 bug

**Non-Goals:**
- 不改变现有的 API 契约（除 alias list 响应结构外）
- 不添加新的 API 端点
- 不修改数据库 schema

## Decisions

### Decision 1: 修改 Alias List 接口一次性返回 Mappings

**选择**: 修改 `alias.go:List()` 预加载 mappings

**理由**:
- 最小的前端改动，只需修改一次接口调用
- 后端改动简单，在 List 查询时 Preload 即可
- 数据量可控，alias 数量通常不大

**实现**:
```go
func (h *AliasHandler) List(c *gin.Context) {
    var aliases []model.Alias
    model.DB.Preload("Mappings.Provider").Find(&aliases)
    // 转换为响应格式，包含 mappings
}
```

### Decision 2: Keys 页面延迟加载 Aliases

**选择**: 将 `fetchAvailableModels()` 移至 `showDialog()` 中调用

**理由**:
- `availableModels` 只在编辑对话框中使用
- 打开对话框时才需要数据
- 可以使用简单的缓存机制避免重复请求

**实现**:
```javascript
const modelsCache = ref<any[]>([])

function showDialog(key?: any) {
    // ...
    if (modelsCache.value.length === 0) {
        fetchAvailableModels()
    }
    // ...
}
```

### Decision 3: 修复 Key Update 的 Models 处理

**选择**: 在 Update 函数中添加删除旧 models 重新插入的逻辑

**理由**:
- 与 Create 函数保持一致的语义
- 实现简单，先删后插

**实现**:
```go
// 先删除所有关联
model.DB.Where("key_id = ?", key.ID).Delete(&model.KeyModel{})
// 再插入新的
for _, alias := range req.Models {
    akm := model.KeyModel{KeyID: key.ID, Model: alias}
    model.DB.Create(&akm)
}
```

## Risks / Trade-offs

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Alias list 响应变大 | 网络传输量增加 | alias 数量通常有限，影响可控 |
| 缓存一致性问题 | Keys 页面编辑时可能不是最新 | 对话框打开时强制刷新 |

## Open Questions

无
