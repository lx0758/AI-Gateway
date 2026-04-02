## Why

当前前端页面存在 API 调用效率问题：
1. 模型别名页面加载时产生 N+1 请求（列表接口 + N 个详情接口）
2. API 密钥页面每次进入都会调用 `/aliases` 接口获取模型列表，即使不需要
3. API 密钥的"允许模型"配置在更新时不生效（后端 bug）
4. 统计页面时间选择器默认只传日期，后端默认时间范围不合理

这些问题影响页面加载性能和用户体验，需要修复。

## What Changes

1. **优化 Alias 列表接口**: 修改 `GET /api/v1/aliases` 一次性返回所有 alias 及其 mappings，消除 N+1 请求
2. **优化 Aliases 页面 Providers 获取**: 将 `/providers` 调用延迟到打开 Mapping 对话框时
3. **优化 Keys 页面模型获取**: 将 `/aliases` 调用延迟到编辑对话框打开时，并添加缓存
4. **修复 API Key Update bug**: 修正 `PUT /api-keys/:id` 时 models 字段被忽略的问题
5. **优化统计页面时间查询**: 
   - 前端传递完整时间格式，后端解析支持 `YYYY-MM-DD HH:mm:ss` 和 `YYYY-MM-DD`
   - 默认时间范围改为当天凌晨到后一天凌晨
   - 添加 loading 效果

## Capabilities

### New Capabilities
无

### Modified Capabilities
无（本次是 bug 修复和性能优化，不涉及需求变更）

## Impact

### 后端修改
- `server/internal/handler/alias.go`: `List` 接口预加载 mappings 数据，一次性返回
- `server/internal/handler/key.go`: `Update` 接口添加 models 的更新逻辑
- `server/internal/handler/usage.go`: `Logs` 接口改用 GORM 链式调用，支持完整时间格式解析

### 前端修改
- `web/src/views/Aliases/index.vue`: 移除逐个获取详情的逻辑，providers 延迟到对话框打开时获取
- `web/src/views/Keys/index.vue`: `fetchAvailableModels()` 移至对话框打开时调用，添加缓存
- `web/src/views/Usage/index.vue`: 时间选择器传递完整时间，添加默认时间范围，添加 loading 效果

### API 变更
- `GET /api/v1/aliases` 响应增加 `mappings` 字段（一次性返回所有数据）
- `GET /usage/logs` 支持 `YYYY-MM-DD HH:mm:ss` 和 `YYYY-MM-DD` 两种时间格式
