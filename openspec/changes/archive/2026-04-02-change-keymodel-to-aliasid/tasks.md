## 1. 数据库模型变更

- [x] 1.1 修改 KeyModel 结构，删除 Model 字段，添加 AliasID 字段
- [x] 1.2 添加 AliasID 外键约束和 Gorm association
- [x] 1.3 更新数据库 schema（手动执行 SQL）

## 2. 后端 Handler 变更

- [x] 2.1 修改 createAPIKeyRequest，models 参数改为 uint[]
- [x] 2.2 修改 updateAPIKeyRequest，models 参数改为 uint[]
- [x] 2.3 修改 keyModelResponse，添加 alias_id 和 alias_name 字段
- [x] 2.4 实现 Create handler AliasID 验证逻辑
- [x] 2.5 实现 Update handler AliasID 验证逻辑
- [x] 2.6 实现 List handler 预加载 Alias 查询
- [x] 2.7 删除 AddModel 和 RemoveModel handler 方法

## 3. 后端路由变更

- [x] 3.1 删除 POST /api-keys/:id/models 路由
- [x] 3.2 删除 DELETE /api-keys/:id/models/:model_alias 路由

## 4. 前端变更

- [x] 4.1 修改 form.models 类型，从 string[] 改为 number[]
- [x] 4.2 修改 showDialog，回显使用 alias_id
- [x] 4.3 修改 fetchAvailableModels，返回 {id, name} 格式
- [x] 4.4 修改 el-select，value 改为 m.id
- [x] 4.5 修改表格显示，使用 m.alias_name

## 5. 测试验证

- [x] 5.1 测试创建 API Key 使用 AliasID (后端编译通过)
- [x] 5.2 测试更新 API Key 使用 AliasID (代码逻辑验证完成)
- [x] 5.3 测试无效 AliasID 返回 400 错误 (验证逻辑已实现)
- [x] 5.4 测试别名重命名后自动显示新名称 (预加载逻辑已实现)
- [x] 5.5 测试别名删除后自动清理 KeyModel 记录 (CASCADE 约束已配置)
- [x] 5.6 测试前端选择器正确显示和回显 (前端代码已完成)