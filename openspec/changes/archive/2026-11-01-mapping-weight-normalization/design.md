# Design

## Context

映射排序的写入点与读取点分布如下（现状）：

- 写入：`CreateMapping` 未携带 weight 时取 `MIN(weight)-1`；携带时走 `freeWeight` 顺延避让；`UpdateMappingsOrder` 写 `n-1-i`
- 读取：`GET /models/:id`、`GET /models/:id/mappings`、`GET /models`、`key.go ListModels`、`router.Route` 共 5 处，全部 `Order("weight DESC")`
- 前端 `Models/Detail.vue`：拖拽 `onEnd` 后本地重排 mappings 并本地重算 `weight = len-1-i`，不重新拉取；增/改/删已调 `fetchModel()`

三个已确认的问题源：`freeWeight` 顺延可能让新映射跳到列表最前；存量重复权重在 `ORDER BY weight DESC` 下 SQLite 行序不定；前端本地推算与后端逐条非事务 UPDATE 存在错位窗口。动机与范围见 proposal.md。

## Goals / Non-Goals

**Goals:**

- 权重成为纯内部值：恒为 `0..n-1` 连续、唯一，行序 = weight DESC 严格确定
- 新增映射恒为 weight 0（排最后、最低优先级），不改变既有路由语义
- 前端列表数据来源以后端为准，不再本地推算权重

**Non-Goals:**

- 不做存量数据启动迁移（靠后续操作自愈）
- 不改变路由选路规则（仍是 weight DESC 取第一个可用映射）
- 不移除 `weight` 字段本身（DB 与 API 响应保留，便于排查与兼容）

## Decisions

### D1: 归一化不变式——权重即位置

任一模型的映射权重恒为 `0..n-1` 连续整数，`weight DESC` 与列表行序严格等价。三处操作在**同一事务内**维护该不变式：

```
新增(事务):
  lockModelMappings(tx, modelID)
  已有映射按 weight DESC, id DESC 保序取出
  第 i 条 -> weight = n - i        (i=0 -> n, 末条 -> 1)
  新映射  -> weight = 0

删除(事务):  删完剩余按 weight DESC, id DESC 保序 -> 重编号 0..n-1
拖拽(事务):  按 order 数组 -> weight = n-1-i      (即 0..n-1)
```

**替代**：仅在新增时用 `MAX(weight)+1` —— 被拒：与"新增排最后"的路由语义冲突，且权重会无限增长、与行序脱钩。

### D2: 创建忽略 weight 参数

`createMappingRequest.Weight` 字段保留但不参与计算。`freeWeight` 整个删除（`UpdateMapping` 一并移除对 weight 的处理——编辑映射本就不该动权重）。

**理由**：归一化后"指定权重"没有落点——任何指定值都会被重编号覆盖，保留逻辑只会留下一条永不生效的分支。保留字段是为了旧客户端不报 400。

### D3: 次级排序键 `id DESC`

5 处查询统一改为 `Order("weight DESC, id DESC")`。归一化后权重唯一，该次级键不改变新数据的行序；它的作用是**存量重复权重在自愈前也保持确定且一致的顺序**（前端列表、Key 列表、真实路由三处必须看到同一个顺序）。

### D4: 前端以服务端为唯一真值

- 拖拽 `onEnd`：`PUT order` 成功后 `fetchModel()` 重新拉取，删除本地重排与 `m.weight = len-1-i` 的回写
- 增/改/删：已有 `fetchModel()`，维持
- 权重列从表格中移除；顺序只由行序表达

**替代**：乐观更新本地顺序后再静默刷新 —— 被拒：正是"显示与实际不符"的来源，用户明确要求重新拉取。

### D5: 归一化辅助函数

抽一个 `renumberMappings(tx, modelID)` 承担"按 `weight DESC, id DESC` 保序重编号 0..n-1"，新增与删除共用；`lockModelMappings` 保留用于并发防护（Postgres `FOR UPDATE`，SQLite 单写者天然串行）。

## Risks / Trade-offs

- **新增操作的写放大**：每次新增/删除要 UPDATE 全部映射。单模型映射数量通常个位数到几十，可接受；数量级上千时再评估增量方案
- **并发窗口**：Postgres 下靠 `FOR UPDATE` 串行化；SQLite 下 `maxOpen=1` 单写者天然串行。极端并发下两次归一化结果仍一致（都收敛到同一连续序列）
- **`weight` 字段成为"看起来无用"的内部值**：列表不再展示，排查需查库 → 接受：顺序本身即是表达，需要时可查 `GET /models/:id` 响应
- **存量数据不迁移**：首次操作前，重复权重下顺序由 `id DESC` 兜底确定 → 接受：自愈机制已覆盖，无需启动期风险

## Migration Plan

1. 部署新版本：无 schema 变更，无需迁移步骤
2. 每个模型的首次新增/删除/拖拽操作即完成该模型的归一化
3. 回滚 = 回退代码版本（旧代码读 weight 仍能正确排序，因新数据恒连续）
