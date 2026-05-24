## 1. 数据模型修改

- [x] 1.1 移除 ProviderModel 的 DeletedAt 字段
- [x] 1.2 移除 ModelMapping 的 DeletedAt 字段
- [x] 1.3 移除 MCPTool 的 DeletedAt 字段
- [x] 1.4 移除 MCPResource 的 DeletedAt 字段
- [x] 1.5 移除 MCPPrompt 的 DeletedAt 字段
- [x] 1.6 移除无效的 OnDelete:CASCADE 约束

## 2. Handler 层修改

- [x] 2.1 修改 ProviderHandler.Delete：级联删除 ProviderModel 和 ModelMapping
- [x] 2.2 修改 ProviderModelHandler.Delete：级联删除 ModelMapping
- [x] 2.3 修改 ModelHandler.Delete：级联删除 ModelMapping
- [x] 2.4 修改 KeyHandler.Delete：级联删除关联表

## 3. 数据库迁移

- [ ] 3.1 备份数据库
- [ ] 3.2 执行迁移脚本：移除 DeletedAt 列
- [ ] 3.3 验证迁移结果

## 4. 验证测试

- [ ] 4.1 验证删除 Provider 时级联删除正确
- [ ] 4.2 验证删除 ProviderModel 时级联删除正确
- [ ] 4.3 验证删除 Model 时级联删除正确
- [ ] 4.4 验证删除 MCP 时级联删除正确
- [ ] 4.5 验证删除 Key 时级联删除正确
