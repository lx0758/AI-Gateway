## 1. 后端实现

- [x] 1.1 在 server/internal/handler/key.go 中新增 Reset 方法
- [x] 1.2 在 server/cmd/server/main.go 中添加路由 POST /api-keys/:id/reset
- [x] 1.3 实现数据库更新逻辑，生成新 Key 值并保留其他字段
- [x] 1.4 实现响应格式，返回 masked key 和 raw_key

## 2. 测试验证

- [x] 2.1 编写单元测试验证 Reset handler 的正确性
- [x] 2.2 测试重置不存在或已删除的 API Key（404 场景）
- [x] 2.3 测试重置后旧 Key 认证失败（401 场景）
- [x] 2.4 测试重置后新 Key 认证成功
- [x] 2.5 测试保留模型绑定配置
- [x] 2.6 测试 Key 值唯一性和格式

## 3. 文档更新

- [x] 3.1 更新 API 文档，说明新增的 reset 端点
- [x] 3.2 在 proposal 或 README 中添加使用示例

## 4. 前端支持（可选）

- [x] 4.1 在前端 API Key 管理界面添加重置按钮
- [x] 4.2 实现重置确认对话框，提示用户旧 Key 将失效
- [x] 4.3 重置成功后显示新的 raw_key，提示用户立即保存