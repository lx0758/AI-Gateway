# Tasks

## 1. 厂商详情页模型 ID 列

- [x] 1.1 修改 `web/src/views/Providers/Detail.vue`：模型 ID 列 `width` 从 200 调整为 300；验证：`cd web && make build` 通过；长模型 ID 完整显示，复制按钮不与 ID 文本重叠

## 2. 用量页长内容 tooltip

- [x] 2.1 修改 `web/src/views/ModelUsage/index.vue` 日志表"厂商/模型"列：用 `el-tooltip` 包裹 `el-tag`，悬浮显示完整的 `厂商/模型` 内容（沿用同页 `error_msg` 的 `:disabled` 长度判断写法，短内容不弹气泡）；验证：`cd web && make build` 通过；长内容 hover 显示完整、短内容不弹
- [x] 2.2 修改"厂商模型统计"卡片的模型列：同样用 `el-tooltip` 包裹，悬浮显示完整模型名；验证：构建通过，长模型名 hover 显示完整

## 3. 集成验证

- [x] 3.1 根目录 `make build` 通过；走查 spec 场景：长模型 ID 完整可读且保留复制入口、日志表与统计卡片长内容均可悬浮查看
- [x] 3.2 运行 `openspec validate console-ui-polish --strict` 通过

## Workflow follow-up

- 实现完成后运行 `/opsx-apply` 的验证流程，或直接归档：`openspec archive console-ui-polish`
