## 1. 修改 model.go 查询函数

- [x] 1.1 修改 List 函数（model.go:88），在查询 ModelMapping 时过滤 Provider.enabled=false
- [x] 1.2 修改 Get 函数（model.go:135），在查询 ModelMapping 时过滤 Provider.enabled=false
- [x] 1.3 修改 Update 函数（model.go:217），在查询 ModelMapping 时过滤 Provider.enabled=false
- [x] 1.4 修改 ListMappings 函数（model.go:262），在查询 ModelMapping 时过滤 Provider.enabled=false

## 2. 修改统计计算函数

- [x] 2.1 修改 calculateMinTokens 函数（model.go:485-518），增加 Provider.enabled 检查
- [x] 2.2 修改 calculateCapabilitiesIntersection 函数（model.go:520-556），增加 Provider.enabled 检查

## 3. 修改 key.go 查询函数

- [x] 3.1 修改 ListModels 函数（key.go:332），在查询 ModelMapping 时过滤 Provider.enabled=false

## 4. 测试验证

- [ ] 4.1 测试禁用 Provider 后，模型列表不再显示相关映射
- [ ] 4.2 测试禁用 Provider 后，模型详情不再显示相关映射
- [ ] 4.3 测试禁用 Provider 后，模型映射列表不再显示相关映射
- [ ] 4.4 测试禁用 Provider 后，Key 模型列表统计数据正确
- [ ] 4.5 测试启用 Provider 后，相关映射恢复显示
