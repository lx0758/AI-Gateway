# Design

## Context

映射权重相关代码现状（动机见 proposal.md）：

- 拖拽排序：前端 `PUT /models/:id/mappings/order`，后端 `UpdateMappingsOrder`（`model.go` L437-473）逐行 `weight = total-1-位置`，产生 `0..total-1` 连续整数（顶部权重最大）
- 创建：`CreateMapping`（`model.go` L323-344）中 `weight==0` 被强制改为 1；无唯一性校验
- 更新：`UpdateMapping`（`model.go` L397-399）`req.Weight != nil` 时直接写入，无校验；`updateMappingRequest.Weight` 为 `*int`
- 前端 `Models/Detail.vue`：弹窗含 `el-input-number` 权重框；`showMappingDialog` L314 `weight: mapping?.weight || mappings.value.length`——`||` 把拖拽产生的合法权重 0 误判为假值，编辑权重 0 的映射并保存时权重被改为总数、顺序跳到最前
- 路由按 `ORDER BY weight DESC` 选路；重复权重时次序不稳定

## Goals / Non-Goals

**Goals:**

- 新增映射自动排到末尾，创建无需输入权重
- 权重字段从前端消失，顺序调整唯一入口为拖拽
- 同一模型下权重唯一性由后端保证（兜底直接 API 调用）
- 消除 `||` 零值 bug 与 `weight==0` 强改 1 的语义矛盾

**Non-Goals:**

- 不改变拖拽排序的权重公式（保持 `0..total-1`，顶部最大）
- 不做批量权重编辑、不迁移存量数据
- 不引入 DB 唯一索引（见 D2）

## Decisions

### D1: 新增排末尾——取最小权重 - 1

`CreateMapping` 未携带 weight 时：查询该模型现有映射的 `MIN(weight)`，新映射取 `min - 1`（无现有映射时取 0）。**一条查询 + 一条插入，不更新任何现有映射**。

- 允许负数：int 存储无碍、排序语义不变（更小 = 更靠后）；前端已隐藏权重字段无感知；拖拽一次即归一为 `0..total-1`
- 替代"全量顺延（现有各 +1，新取 0）"：位移方案要 UPDATE 全部现有行，写放大且并发下互相干扰；`min-1` 只写一行，效果等同（都排末尾、都唯一）
- 拖拽公式保持 0-based（`0..total-1`，顶部最大）不变

`createMappingRequest.Weight` 保持可选：显式携带 weight（如脚本指定优先级）时使用该值，冲突走 D2 自动避让；未携带（前端路径）走 `min-1` 分配。字段类型改为 `*int` 以区分"未携带"与"携带 0"。

### D2: 冲突自动避让——不加数据库约束、不阻断操作

**明确不加** `(model_id, weight)` 唯一索引：一是拖拽逐行更新的中间态必然短暂重复，索引会直接报错阻断排序；二是权重只是排序内部值，不值得为它拒绝任何操作。

后台检查与避让（应用层，事务内）：

- `Create` 携带显式 weight W：检查 W 是否被同模型其他映射占用；占用则从 W 起 +1 递增直至空位（仅移动本次新插入的这条）
- `Update` 携带 weight W（`*int` 非 nil）：排除自身后同样检查，冲突则 +1 递增至空位（仅移动被编辑的这条，其他映射不动）；未携带（`nil`）不触碰权重
- 拖拽接口：连续分配天然唯一，无需检查
- 并发：检查与写入同一事务；SQLite 单写者天然串行，PostgreSQL 下事务内 `SELECT ... FOR UPDATE` 锁定该模型现有映射行（避免并发双双选中同一空位）

替代方案：冲突返回 400 拒绝——被否决：小事阻断操作；冲突时全量重排腾位——被否决：改动别人的数据，副作用大。

### D3: 清理旧语义

- 删除 `CreateMapping` 中 `if mapping.Weight == 0 { mapping.Weight = 1 }`——0 是合法的"末尾"权重
- `UpdateMapping` 的 weight 写入路径接入 D3 校验

### D4: 前端移除权重输入

`Models/Detail.vue`：

- 映射弹窗删除 `el-input-number` 权重项；`mappingForm` 移除 `weight` 字段（提交体自然不再携带，后端按 D1/D2 走自动分配）
- `showMappingDialog` 删除 `weight: mapping?.weight || mappings.value.length` 行——`||` 零值 bug 随之消失
- 拖拽 `onEnd` 逻辑与 `PUT /mappings/order` 调用保持不变

## Risks / Trade-offs

- **显式 weight 的 API 语义变化**：携带 weight 从"原样生效（可重复）"变为"生效但冲突时自动 +1 避让（响应权重可能与请求值不同）"，未携带从"默认 1"变为"自动排末尾" → 轻微 BREAKING（proposal 已标注）；现有前端同步修改，控制台无感知
- **自动避让的隐蔽性**：直连 API 的脚本指定权重与实际落库值不一致且无报错，可能造成困惑 → 缓解：响应返回实际权重，脚本可读回确认
- **存量重复权重**：既有数据若已存在重复权重，不主动迁移；任一次拖拽或后续顺延重排即自愈
- **顺延后权重上界**：频繁新增使权重持续 +1（int 范围内无实际风险）；拖拽一次即归零为 `0..total-1`

## Migration Plan

无 DB 迁移。部署顺序：前后端同一二进制内嵌发布，无版本错配窗口。回滚 = 回退代码；旧前端对新后端（权重自动分配）兼容——旧前端仍会传 weight，命中 D2 显式路径。
