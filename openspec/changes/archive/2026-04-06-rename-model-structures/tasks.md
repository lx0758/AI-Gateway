## 1. 数据库迁移

- [x] 1.1 创建数据库迁移脚本（重命名表 aliases → models）
- [x] 1.2 创建数据库迁移脚本（重命名表 alias_mappings → model_mappings）
- [x] 1.3 创建数据库迁移脚本（重命名表 usage_logs → model_logs）
- [x] 1.4 更新列名 alias_mappings.alias_id → model_mappings.model_id
- [x] 1.5 更新列名 key_models.alias_id → key_models.model_id（间接依赖）
- [x] 1.6 创建回滚脚本

## 2. 后端 PO 结构体重命名

- [x] 2.1 删除未使用的 Mapping PO（server/internal/model/db.go）
- [x] 2.2 重命名 Alias → Model（server/internal/model/db.go）
- [x] 2.3 重命名 AliasMapping → ModelMapping（server/internal/model/db.go）
- [x] 2.4 重命名 UsageLog → ModelLog（server/internal/model/db.go）
- [x] 2.5 更新 TableName 方法返回新表名
- [x] 2.6 更新 ModelMapping 结构体字段名（AliasID → ModelID）
- [x] 2.7 更新 KeyModel 结构体字段名（AliasID → ModelID）（间接依赖）

## 3. 后端 Handler 更新

- [x] 3.1 重命名 handler/alias.go → handler/model.go
- [x] 3.2 更新 model.go 中所有函数名和变量名
- [x] 3.3 更新 handler/usage.go 中的 UsageLog 引用
- [x] 3.4 更新 handler/key.go 中的 Alias/KeyModel 引用（间接依赖）
- [x] 3.5 更新 handler/proxy*.go 中的相关引用
- [x] 3.6 更新所有 handler 中的数据库查询

## 4. 后端路由更新

- [x] 4.1 更新 router/router.go 中的路由路径
- [x] 4.2 更新路由组名称和注释
- [x] 4.3 添加旧路由别名（可选）

## 5. 后端全局引用更新

- [x] 5.1 更新 main.go 中的初始化代码
- [x] 5.2 搜索并更新所有 model.Alias 引用
- [x] 5.3 搜索并更新所有 model.AliasMapping 引用
- [x] 5.4 搜索并更新所有 model.UsageLog 引用

## 6. 前端视图重命名

- [x] 6.1 重命名 web/src/views/Aliases/ → web/src/views/Models/
- [x] 6.2 更新 Models/index.vue 中的变量名和 API 调用
- [x] 6.3 更新 Models/Detail.vue 中的变量名和 API 调用
- [x] 6.4 更新组件导入路径

## 7. 前端路由更新

- [x] 7.1 更新 router/index.ts 中的路由路径
- [x] 7.2 更新路由名称和组件引用
- [x] 7.3 更新 MainLayout.vue 中的导航菜单路径

## 8. 前端国际化更新

- [x] 8.1 更新 locales/zh.ts 中的相关翻译
- [x] 8.2 更新 locales/en.ts 中的相关翻译
- [x] 8.3 更新所有页面中的 i18n 键名

## 9. 前端全局引用更新

- [x] 9.1 搜索并更新所有 "alias" 相关变量名
- [x] 9.2 搜索并更新所有 "Alias" 相关类型名
- [x] 9.3 更新 Keys/index.vue 中的模型选择器引用（间接依赖）

## 10. Spec 文档更新

- [x] 10.1 重命名 openspec/specs/model-alias/ → openspec/specs/model-management/
- [x] 10.2 重命名 openspec/specs/alias-mapping/ → openspec/specs/model-mapping/
- [x] 10.3 更新 model-management/spec.md 中的术语
- [x] 10.4 更新 model-mapping/spec.md 中的术语
- [x] 10.5 更新 usage-tracking/spec.md 中的表名引用

## 11. API 路径简化（api-keys → keys）

- [x] 11.1 更新 main.go 中的 API 路径 `/api-keys` → `/keys`
- [x] 11.2 更新前端 API 调用路径
- [x] 11.3 运行编译检查

## 12. 测试与验证

- [ ] 12.1 执行数据库迁移测试
- [ ] 12.2 验证 API 功能正常
- [ ] 12.3 验证前端页面功能正常

---

**注**: 标记"间接依赖"的任务涉及非直接重命名的结构，但因引用关系需要同步更新。
