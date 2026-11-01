# Tasks

## 1. 后端归一化

- [x] 1.1 修改 `server/internal/handler/model.go`：新增 `renumberMappings(tx, modelID)` 辅助函数——事务内按 `weight DESC, id DESC` 保序取出该模型全部映射，重编号为 `0..n-1`；`CreateMapping` 改为忽略 `req.Weight`，事务内先 `lockModelMappings` 再将已有映射重编号 `1..n`、新映射 weight = 0；验证：`cd server && go build ./...` 通过；连续创建 3 条映射得到权重 [2,1,0] 且新映射排列表最后
- [x] 1.2 修改 `UpdateMapping`：移除对 `weight` 的处理与 `freeWeight` 调用（编辑映射不修改权重）；删除 `freeWeight` 函数；验证：`go build ./...` 通过；编辑某映射的 enabled 后其 weight 与列表顺序不变
- [x] 1.3 修改 `DeleteMapping`：删除后在事务内调用 `renumberMappings` 归一化剩余映射；验证：在 [3,2,1,0] 中删除中间一条后剩余为 [2,1,0] 且相对顺序不变
- [x] 1.4 将 5 处映射查询的 `Order("weight DESC")` 改为 `Order("weight DESC, id DESC")`：`model.go` 的 List / Get / ListMappings / Update，`key.go` 的 ListModels，`router.go` 的 Route；验证：`go build ./...` 通过；对存在重复权重的存量库，前端列表、Key 模型列表、真实路由三处返回一致顺序

## 2. 前端适配

- [x] 2.1 修改 `web/src/views/Models/Detail.vue`：移除权重列（`prop="weight"` 的 `el-table-column`）；验证：`cd web && make build` 通过，映射表不再显示权重数值列，行序仍表达顺序
- [x] 2.2 修改拖拽 `onEnd` 逻辑：`PUT /mappings/order` 成功后调用 `fetchModel()` 重新拉取列表，删除本地重排与 `m.weight = mappings.value.length - 1 - i` 的回写；失败时同样 `fetchModel()` 回滚显示；验证：拖拽后列表顺序与接口返回一致，本地不出现与数据库不符的权重值

## 3. 集成验证

- [x] 3.1 根目录 `make build` 通过；完整走查 spec 场景：新增排最后且 weight=0、显式携带 weight 被忽略且不报错、删除后归一化无空洞、拖拽后归一化且重拉列表、存量重复权重经任一操作后自愈为连续序列、相同权重下行序确定一致
- [x] 3.2 运行 `openspec validate mapping-weight-normalization --strict` 通过

## Workflow follow-up

- 实现完成后运行 `/opsx-apply` 的验证流程，或直接归档：`openspec archive mapping-weight-normalization`
