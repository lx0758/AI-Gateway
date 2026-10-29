# Tasks

## 1. 后端权重管理

- [x] 1.1 修改 `server/internal/handler/model.go` `CreateMapping` 与 `createMappingRequest`：`Weight` 改为 `*int` 可选字段；未携带时查询该模型 `MIN(weight)` 取 `min-1` 插入（无现有映射取 0），不更新其他映射；显式携带时事务内检查冲突，占用则从该值起 +1 递增至空位（不阻断、不报错）；删除 `weight==0` 强改为 1 的旧逻辑；验证：`cd server && go build ./...` 通过；连续创建 3 条不带权重的映射得到递减权重（如 0、-1、-2）且新映射排最后；显式指定已被占用的权重时创建成功且落库为顺延后的空位值
- [x] 1.2 修改 `UpdateMapping`：携带 `weight`（非 nil）时排除自身检查冲突，占用则 +1 递增至空位（仅移动该映射）；未携带时不触碰权重；验证：`go build ./...` 通过，直接 API 将某映射 weight 更新为另一映射的现有权重时更新成功且顺延至空位、其他映射权重不变
- [x] 1.3 验证拖拽排序不受影响：拖拽后权重仍为连续 `0..total-1`（顶部最大）、无重复；验证：控制台拖拽两次后各映射权重互不相同且顺序与拖拽一致

## 2. 前端移除权重输入

- [x] 2.1 修改 `web/src/views/Models/Detail.vue`：映射弹窗删除权重 `el-form-item`；`mappingForm` 移除 `weight` 字段；`showMappingDialog` 删除 `weight: mapping?.weight || mappings.value.length` 默认值行；验证：`cd web && make build` 通过，新增/编辑弹窗无权重控件、提交请求体不含 weight
- [x] 2.2 验证编辑不再改权重：编辑一条拖拽排在最后（权重 0）的映射仅切换启用状态并保存，刷新后该映射仍排在最后（权重未被改为总数）

## 3. 集成验证

- [x] 3.1 根目录 `make build` 通过；完整走查 spec 场景：新增映射自动排末尾、拖拽调整顺序生效且唯一、编辑指向/状态后顺序不变、直接 API 传重复权重自动顺延成功（响应返回实际权重）
- [x] 3.2 运行 `openspec validate --strict --change mapping-order-ux` 通过

## Workflow follow-up

- 实现完成后运行 `/opsx-apply` 的验证流程，或直接归档：`openspec archive mapping-order-ux`
