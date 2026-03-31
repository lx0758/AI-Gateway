## 1. 数据库迁移

- [x] 1.1 创建迁移脚本：添加 provider_model_name 列
- [x] 1.2 迁移现有数据：provider_model_id → provider_model_name
- [x] 1.3 删除 provider_model_id 列和外键约束
- [x] 1.4 更新 model/models.go 中 ModelMapping 结构体

## 2. 后端修改 - ModelMapping

- [x] 2.1 修改 handler/model_mapping.go：Create 使用 provider_model_name
- [x] 2.2 修改 handler/model_mapping.go：Update 使用 provider_model_name
- [x] 2.3 修改 handler/model_mapping.go：List 返回 provider_model_name
- [x] 2.4 修改 router/model_router.go：按 (provider_id, model_id) 查询

## 3. 后端修复 - 删除保护

- [x] 3.1 修改 handler/provider_model.go 的 Delete 方法
- [x] 3.2 移除 source 检查，允许删除所有模型
- [ ] 3.3 添加单元测试验证删除保护逻辑

## 4. 前端 UI - 模型列表增强

- [x] 4.1 在模型表格中添加"来源"列，显示 Manual/Sync 标签
- [x] 4.2 所有模型都显示编辑/删除按钮

## 5. 前端 UI - 添加模型对话框

- [x] 5.1 在 Provider 详情页添加"添加模型"按钮
- [x] 5.2 创建添加模型对话框组件
- [x] 5.3 实现表单字段：model_id, display_name, context_window, max_output, supports_vision, supports_tools, supports_stream
- [x] 5.4 调用 POST /api/v1/providers/:id/models API

## 6. 前端 UI - 编辑模型对话框

- [x] 6.1 点击编辑按钮时打开对话框并加载模型数据
- [x] 6.2 实现编辑表单（字段与添加相同）
- [x] 6.3 调用 PUT /api/v1/providers/:id/models/:mid API

## 7. 前端 UI - 删除模型

- [x] 7.1 点击删除按钮时显示确认对话框
- [x] 7.2 确认后调用 DELETE /api/v1/providers/:id/models/:mid API
- [x] 7.3 处理删除失败的情况

## 8. 前端 UI - ModelMapping 表单修改

- [x] 8.1 修改 Models/index.vue：选择模型时存储 model_id 而非 ID
- [x] 8.2 修改 API 调用：使用 provider_model_name 字段

## 9. 前端 UI - 其他增强

- [x] 9.1 厂商列表页添加启用/禁用开关
- [x] 9.2 API 密钥列表页添加编辑功能
- [x] 9.3 点击模型行显示详情对话框
- [x] 9.4 Provider 模型列表添加 Loading 状态

## 10. 前端 UI - 批量删除

- [x] 10.1 厂商列表页添加多选批量删除
- [x] 10.2 Provider 模型列表添加多选批量删除
- [x] 10.3 模型映射列表页添加多选批量删除
- [x] 10.4 API 密钥列表页添加多选批量删除
- [x] 10.5 添加中英文翻译 (batchDelete)

## 11. 国际化 (i18n)

- [x] 11.1 添加英文翻译 (en.ts)
- [x] 11.2 添加中文翻译 (zh.ts)

## 12. 测试

- [ ] 12.1 测试手动添加模型
- [ ] 12.2 测试编辑手动模型
- [ ] 12.3 测试删除模型
- [ ] 12.4 测试点击模型查看详情
- [ ] 12.5 测试厂商启用/禁用
- [ ] 12.6 测试 API 密钥编辑
- [ ] 12.7 测试 ModelMapping 创建/查询使用新关联方式
- [ ] 12.8 测试各列表页批量删除功能
