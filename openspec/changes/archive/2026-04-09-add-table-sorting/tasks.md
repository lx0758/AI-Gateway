## 1. 创建排序工具函数

- [x] 1.1 创建 `web/src/utils/tableSort.ts` 文件
- [x] 1.2 实现 `getSortConfig` 函数（从 localStorage 读取排序配置）
- [x] 1.3 实现 `setSortConfig` 函数（保存排序配置到 localStorage）
- [x] 1.4 实现 `sortByDate` 函数（日期排序，空值排末尾）
- [x] 1.5 实现 `sortByArrayLength` 函数（数组长度排序）

## 2. Models 主列表添加排序

- [x] 2.1 在 `web/src/views/Models/index.vue` 中导入排序工具函数
- [x] 2.2 为表格添加 `:default-sort` 和 `@sort-change` 属性
- [x] 2.3 为 `name` 列添加 `sortable` 属性
- [x] 2.4 为 `mapping_count` 列添加 `sortable` 属性
- [x] 2.5 为 `min_context_window` 列添加 `sortable` 属性
- [x] 2.6 为 `enabled` 列添加 `sortable` 属性
- [x] 2.7 实现 `handleSortChange` 函数保存排序配置

## 3. Providers 主列表添加排序

- [x] 3.1 在 `web/src/views/Providers/index.vue` 中导入排序工具函数
- [x] 3.2 移除 `fetchProviders` 函数中的后端排序逻辑
- [x] 3.3 为表格添加 `:default-sort` 和 `@sort-change` 属性
- [x] 3.4 为 `name` 列添加 `sortable` 属性
- [x] 3.5 为模型数量列添加 `sortable` 属性和 `sort-method`
- [x] 3.6 为 `enabled` 列添加 `sortable` 属性
- [x] 3.7 实现 `handleSortChange` 函数保存排序配置

## 4. Keys 主列表添加排序

- [x] 4.1 在 `web/src/views/Keys/index.vue` 中导入排序工具函数
- [x] 4.2 为表格添加 `:default-sort` 和 `@sort-change` 属性
- [x] 4.3 为 `name` 列添加 `sortable` 属性
- [x] 4.4 为模型列添加 `sortable` 属性和 `sort-method`
- [x] 4.5 为 `mcp_tools_count` 列添加 `sortable` 属性
- [x] 4.6 为 `mcp_resources_count` 列添加 `sortable` 属性
- [x] 4.7 为 `mcp_prompts_count` 列添加 `sortable` 属性
- [x] 4.8 为 `enabled` 列添加 `sortable` 属性
- [x] 4.9 实现 `handleSortChange` 函数保存排序配置

## 5. MCPs 主列表添加排序

- [x] 5.1 在 `web/src/views/MCPs/index.vue` 中导入排序工具函数
- [x] 5.2 为表格添加 `:default-sort` 和 `@sort-change` 属性
- [x] 5.3 为 `name` 列添加 `sortable` 属性
- [x] 5.4 为 `type` 列添加 `sortable` 属性
- [x] 5.5 为 `tool_count` 列添加 `sortable` 属性
- [x] 5.6 为 `resource_count` 列添加 `sortable` 属性
- [x] 5.7 为 `prompt_count` 列添加 `sortable` 属性
- [x] 5.8 为 `enabled` 列添加 `sortable` 属性
- [x] 5.9 为 `last_sync_at` 列添加 `sortable` 属性和 `sort-method`
- [x] 5.10 实现 `handleSortChange` 函数保存排序配置

## 6. Models Detail 页面添加排序和禁用状态

- [x] 6.1 为映射列表添加排序功能
- [x] 6.2 模型禁用时映射开关显示禁用状态
- [x] 6.3 添加国际化文本 `modelDisabled`

## 7. Providers Detail 页面添加排序和功能

- [x] 7.1 为模型列表添加排序功能
- [x] 7.2 移除后端排序逻辑
- [x] 7.3 后端支持修改 `is_available` 字段
- [x] 7.4 前端添加 `toggleAvailable` 函数
- [x] 7.5 厂商禁用时模型开关显示禁用状态
- [x] 7.6 添加国际化文本 `providerDisabled`

## 8. Keys Detail 页面添加排序和禁用状态

- [x] 8.1 为 models 列表添加排序功能
- [x] 8.2 为 tools 列表添加排序功能
- [x] 8.3 为 resources 列表添加排序功能
- [x] 8.4 为 prompts 列表添加排序功能
- [x] 8.5 密钥禁用时 radio-group 显示禁用状态
- [x] 8.6 密钥禁用时"全部允许"按钮禁用
- [x] 8.7 添加国际化文本 `keyDisabled`

## 9. MCPs Detail 页面添加排序和禁用状态

- [x] 9.1 为 tools 列表添加排序功能
- [x] 9.2 为 resources 列表添加排序功能
- [x] 9.3 为 prompts 列表添加排序功能
- [x] 9.4 服务禁用时开关显示禁用状态
- [x] 9.5 添加国际化文本 `serviceDisabled`

## 10. 修改默认排序字段

- [x] 10.1 将默认排序字段从 `id` 改为 `name`
- [x] 10.2 更新各列表页面的默认排序配置

## 11. 测试验证

- [x] 11.1 TypeScript 类型检查通过
- [x] 11.2 Go 后端编译通过
