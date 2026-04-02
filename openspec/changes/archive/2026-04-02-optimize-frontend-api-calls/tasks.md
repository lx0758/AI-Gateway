## 1. 后端修改

- [x] 1.1 修改 `alias.go:List()` 预加载 Mappings 数据，一次性返回 alias 列表及所有 mappings
- [x] 1.2 修改 `key.go:Update()` 添加 Models 更新逻辑：先删除旧的 KeyModel 记录，再插入新的
- [x] 1.3 修改 `usage.go:Logs()` 改用 GORM 链式调用，支持完整时间格式 `YYYY-MM-DD HH:mm:ss`

## 2. 前端修改

- [x] 2.1 修改 `Aliases/index.vue:fetchAliases()` 移除逐个获取详情的逻辑，直接使用 List 返回的 mappings
- [x] 2.2 修改 `Aliases/index.vue` 将 `fetchProviders()` 移至 `showMappingDialog()` 中调用，添加缓存
- [x] 2.3 修改 `Keys/index.vue` 将 `fetchAvailableModels()` 移至 `showDialog()` 中调用，添加缓存
- [x] 2.4 修改 `Usage/index.vue` 时间选择器传递完整时间格式，设置默认时间范围（当天凌晨到后一天凌晨），添加 loading 效果
