## 1. 数据模型修改

- [x] 1.1 修改 ModelMapping 结构体：删除 ProviderModelName，添加 ProviderModelID 外键
- [x] 1.2 添加 ProviderModel 关联到 ModelMapping 结构体
- [x] 1.3 添加索引：provider_id、provider_model_id（非唯一）

## 2. Handler 层修改

- [x] 2.1 修改 createMappingRequest：provider_model_name → provider_model_id
- [x] 2.2 修改 updateMappingRequest：provider_model_name → provider_model_id
- [x] 2.3 修改 CreateMapping：通过 ID 查找 ProviderModel
- [x] 2.4 修改 UpdateMapping：通过 ID 查找 ProviderModel
- [x] 2.5 修改 toMappingResponse：使用 Preload 获取 ProviderModel
- [x] 2.6 修改 calculateMinTokens：使用 Preload 获取 ProviderModel
- [x] 2.7 修改 calculateCapabilitiesIntersection：使用 Preload 获取 ProviderModel
- [x] 2.8 修改 ProviderModelHandler.Update：允许修改 ModelID 字段

## 3. Router 层修改

- [x] 3.1 修改 Route 方法：使用 ProviderModelID 查询 ProviderModel

## 4. 测试层修改

- [x] 4.1 修改 model_testing.go：使用 Preload 获取 ProviderModel

## 5. 前端修改

- [x] 5.1 修改 Models/Detail.vue：表单字段 provider_model_name → provider_model_id
- [x] 5.2 修改 Providers/Detail.vue：移除 model_id 编辑禁用

## 6. 数据迁移

- [x] 6.1 创建迁移脚本 migration-guide.md
- [ ] 6.2 备份数据库
- [ ] 6.3 执行迁移脚本
- [ ] 6.4 验证迁移结果

## 7. 验证测试

- [ ] 7.1 验证模型列表接口正常
- [ ] 7.2 验证创建映射使用 provider_model_id
- [ ] 7.3 验证更新映射使用 provider_model_id
- [ ] 7.4 验证删除 ProviderModel 级联删除 Mapping
- [ ] 7.5 验证修改 ProviderModel.ModelID 后路由正常
- [ ] 7.6 验证响应体中 provider_model_name 正确返回
