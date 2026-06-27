# Tasks: Disabled Mappings Visible

## Implementation

- [x] 1.1 后端: `mappingResponse` struct 新增 `ForcedDisabled` 和 `DisableReason` 字段
- [x] 1.2 后端: `toMappingResponse` 函数判断强制禁用状态（Provider 禁用 > ProviderModel 不可用）
- [x] 1.3 后端: `List()` 方法移除 `JOIN providers AND providers.enabled = ? true` 过滤
- [x] 1.4 后端: `Get()` 方法移除 JOIN 过滤
- [x] 1.5 后端: `Update()` 方法移除 JOIN 过滤
- [x] 1.6 后端: `ListMappings()` 方法移除 JOIN 过滤
- [x] 1.7 前端: `Mapping` 接口新增 `forced_disabled` 和 `disable_reason` 字段
- [x] 1.8 前端: 状态列三种状态（模型禁用 → 灰色开关 / 强制禁用 → 灰色禁用开关 + 精确 tooltip / 正常 → 可切换开关）
- [x] 1.9 前端: 中文 i18n 新增强制禁用相关翻译（`forcedDisabledByProvider`, `forcedDisabledByModel`, `forcedDisabledByOther`, `providerDisabledReason`, `providerModelUnavailableReason`, `providerOrModelDisabled`）
- [x] 1.10 前端: 英文 i18n 新增对应翻译（`Provider disabled`, `Provider model disabled`, `Upstream disabled` 等）
- [x] 1.11 后端: `key.go` ListModels 方法移除 `JOIN providers AND providers.enabled = ? true` 过滤
- [x] 1.12 前端: tooltip 使用 `getDisableTooltip` 提供精确禁用原因（"厂商已禁用，此映射被强制禁用，无法启用" / "厂商模型不可用..."）

## Verification

- [ ] 2.1 启动服务，创建一个模型映射
- [ ] 2.2 禁用关联的厂商，检查模型详情页是否显示该映射并显示灰色禁用开关 + 精确 tooltip
- [ ] 2.3 启用厂商后，恢复为正常开关状态
- [ ] 2.4 将厂商模型标记为不可用（is_available=false），检查映射显示灰色禁用开关 + 精确 tooltip
- [ ] 2.5 手动禁用一个映射，确认开关仍可操作（不是强制禁用）
- [ ] 2.6 检查模型列表页的聚合统计（mapping_count 等）不受禁用映射影响
- [ ] 2.7 检查 API Key 模型列表中映射统计（mapping_count）包含所有厂商映射
- [ ] 2.8 检查英文界面翻译正确
