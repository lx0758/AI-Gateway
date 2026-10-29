# Tasks

## 1. 三态字段改造与存量迁移

- [x] 1.1 修改 `server/internal/model/db.go` ProviderModel PO：`IsAvailable bool` 替换为 `Availability string`（`gorm:"type:varchar(20);default:unknown"`），删除 `Source`；在 `AutoMigrate` 后实现一次性幂等迁移（事务内：读旧 `source`/`is_available` 列——manual→unknown、sync 且 true→available、false→unavailable——写 `availability`，完成后删除 `source` 列，已有值的行跳过）；验证：`cd server && go build ./...` 通过；用旧版本库启动一次再升级，存量数据三态映射正确（手动模型为 unknown 不被首次同步误标失效）
- [x] 1.2 修改 `server/internal/handler/provider_model.go`：响应 DTO 移除 `is_available`/`source`、增加 `availability`；Create 手动新建置 `availability="unknown"`；Update 适配；删除 `available_only` dead param；删除逻辑维持现状（不做删除保护）；验证：手动新建模型响应 `availability="unknown"`；对任意状态模型删除均可用

## 2. 同步状态机

- [x] 2.1 修改 `server/internal/provider/provider_openai.go`、`provider_anthropic.go` `SyncModels`：未返回字段留空（DisplayName 不用模型 ID 冒充、Capabilities 不硬编码 "stream"、MaxOutput 不伪造）；验证：`go build ./...` 通过
- [x] 2.2 重写 `Sync` 合并与状态推进（design D3/D5）：空列表短路；列表中模型逐字段空值合并（空字段不进 updates map）+ 置 `available`；创建的新模型置 `available`；`available` 且缺失者批量置 `unavailable` 并计数；`unknown` 不参与失效；响应增加 `deactivated`；验证：手工补全 `display_name`/`context_window`/`capabilities` 后同步字段保留；移除一模型后同步其变 `unavailable` 且映射被路由跳过、`deactivated=1`；手动加的 unknown 模型不同步不失效；空列表同步无任何状态变化

## 3. 消费点适配

- [x] 3.1 按 design D6 表修改：`router.go:49`、`model.go:502/531`、`model_testing.go:262` 的判断改为 `Availability == "unavailable"`；`usage.go:57` 统计条件改为 `pm.availability != 'unavailable'`；`model.go` 的 `calculateMinTokens`、`calculateCapabilitiesIntersection` 增加失效跳过；验证：`go build ./...` 通过；失效模型的映射不参与路由、启用数减少、仪表盘"活跃厂商模型数"减少、Token Summary/能力交集排除失效模型（全部指向失效时显示 "-"）

## 4. 前端三态展示

- [x] 4.1 修改 `web/src/views/Providers/Detail.vue`：来源列与详情 tag 改为三态 tag（未知 info/可用 success/失效 danger）；编辑/删除按钮不做显隐；同步完成 `deactivated > 0` 时 toast"标记 N 个失效模型"；验证：`cd web && make build` 通过，编辑/删除按钮对所有模型可见，三态标签颜色正确
- [x] 4.2 修改 `web/src/views/Models/Detail.vue` 适配三态（tooltip `provider_model_unavailable` 语义保持）；补充 zh/en 三态文案（未知/可用/失效、Unknown/Available/Unavailable）；验证：中英文切换下三态标签与 tooltip 正确

## 5. 集成验证

- [x] 5.1 根目录 `make build` 通过；完整走查 spec 场景：存量库升级迁移正确、补全字段同步不丢失、下线模型同步失效并退出路由/统计、恢复上架自动可用、手动 unknown 模型全程不受同步失效影响、任意状态模型均可编辑与删除
- [x] 5.2 运行 `openspec validate smart-model-sync --strict` 通过

## Workflow follow-up

- 实现完成后运行 `/opsx-apply` 的验证流程，或直接归档：`openspec archive smart-model-sync`
- 归档时 `manual-model-protection`、`manual-model-ui` 的 MODIFIED delta 将并入主 spec（两份主 spec 现为旧格式，归档前需转标准格式——既有债，见会话记录）
