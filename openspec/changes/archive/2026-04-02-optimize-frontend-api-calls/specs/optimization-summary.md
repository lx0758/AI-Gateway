## Summary

本次变更不涉及新增或修改 capabilities，全部为 bug 修复和性能优化：

- **Bug Fix**: 修复 `PUT /api-keys/:id` 更新时不保存 models 配置的问题
- **Performance**: 优化 `GET /aliases` 接口，消除前端 N+1 请求问题
- **Performance**: 优化 Keys 页面，将 aliases 数据获取延迟到需要时

详见 `design.md`
