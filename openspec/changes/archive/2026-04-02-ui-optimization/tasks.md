## 1. 数据模型重构

- [x] 1.1 在 `model/db.go` 中新增 `Alias` 结构体（ID, Name uniqueIndex, Enabled, Mappings hasMany）
- [x] 1.2 在 `model/db.go` 中新增 `AliasMapping` 结构体（ID, AliasID index, ProviderID, ProviderModelName, Weight, Enabled, Provider belongsTo）
- [x] 1.3 在 `model/db.go` 的 `autoMigrate()` 中注册 `Alias` 和 `AliasMapping`，删除 `ModelMapping`
- [ ] 1.4 验证数据库自动迁移创建新表

## 2. 后端 API 实现

- [x] 2.1 创建 `handler/alias.go` 实现 AliasHandler
- [x] 2.2 实现 `GET /aliases` 列表接口（含 mapping_count）
- [x] 2.3 实现 `POST /aliases` 创建接口
- [x] 2.4 实现 `GET /aliases/:id` 详情接口（含 mappings 数组）
- [x] 2.5 实现 `PUT /aliases/:id` 更新接口
- [x] 2.6 实现 `DELETE /aliases/:id` 删除接口（级联删除 mappings）
- [x] 2.7 实现 `GET /aliases/:id/mappings` 映射列表接口
- [x] 2.8 实现 `POST /aliases/:id/mappings` 添加映射接口
- [x] 2.9 实现 `PUT /aliases/:id/mappings/:mid` 更新映射接口
- [x] 2.10 实现 `DELETE /aliases/:id/mappings/:mid` 删除映射接口
- [x] 2.11 在 `main.go` 中注册新路由，删除旧 `/model-mappings` 路由
- [x] 2.12 删除 `handler/model_mapping.go`
- [ ] 2.13 验证 API 接口可用

## 3. 代理路由逻辑更新

- [x] 3.1 更新 `router/router.go` 的 `Route()` 方法，先查 Alias 再查 AliasMapping
- [x] 3.2 更新 `proxy_openai.go` 的 `ListModels()` 使用 Alias
- [x] 3.3 更新 `proxy_openai.go` 的 `GetModel()` 使用 Alias
- [x] 3.4 更新 `provider.go` 中删除厂商时的映射检查使用 AliasMapping
- [ ] 3.5 验证代理路由功能正常

## 4. 前端别名管理页面

- [x] 4.1 创建 `views/Aliases/index.vue` 别名管理页面
- [x] 4.2 使用 `el-collapse` 实现折叠卡片布局
- [x] 4.3 实现别名列表 API 调用和数据展示
- [x] 4.4 实现别名创建/编辑对话框
- [x] 4.5 实现内嵌映射表格，显示厂商名称、厂商类型、实际模型、权重
- [x] 4.6 实现厂商类型 Tag 判断（根据 provider.base_url）
- [x] 4.7 实现映射添加/编辑/删除功能
- [x] 4.8 删除旧的 `views/Models/` 目录
- [x] 4.9 更新路由配置：`/models` → `/aliases`
- [x] 4.10 更新侧边栏菜单
- [x] 4.11 更新 i18n 翻译
- [ ] 4.12 验证别名管理页面功能

## 5. 其他 UI 优化

- [x] 5.1 在 `views/Providers/index.vue` 中对 providers 按 name 排序
- [x] 5.2 在 `views/Providers/Detail.vue` 中对 models 按 model_id 排序
- [x] 5.3 在 `views/Usage/index.vue` 中添加错误信息复制按钮
- [x] 5.4 导入 Element Plus 的 CopyDocument 图标
- [ ] 5.5 验证排序和复制功能

## 6. 测试验证

- [ ] 6.1 创建测试别名和映射，验证路由功能
- [ ] 6.2 验证 API 调用使用新别名正确路由到厂商
- [ ] 6.3 验证别名禁用后路由失败
- [ ] 6.4 验证映射权重排序正确
- [ ] 6.5 验证前端别名管理页面所有功能