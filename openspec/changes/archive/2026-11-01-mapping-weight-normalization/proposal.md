# Proposal

## Why

权重是映射列表顺序与路由优先级的唯一依据，但当前权重值可能与实际显示顺序不符：

- 创建映射时若 API 显式携带权重且与现有值冲突，`freeWeight` 避让逻辑会 +1 递增到空位，新映射可能跳到列表最前，与管理员"新增排最后"的预期相反
- 存在重复权重时 `ORDER BY weight DESC` 在 SQLite 下行序不确定，同样权重的两条映射前后顺序会飘
- 前端拖拽保存后不重新拉取列表，而是本地重排并重算权重值；后端逐条 UPDATE 非事务，中途失败会留下半更新状态，导致显示顺序与数据库实际不一致

## What Changes

- 权重改为**恒定不变式**：任一模型的映射权重恒为 `0..n-1` 连续整数且唯一，列表行序 = `weight DESC` 严格确定
- 新增映射：**忽略传入的 `weight` 参数**，恒为 0（排列表最后、最低路由优先级，不改变现有路由语义）；同时将已存在的映射按当前相对顺序重编号 `1..n`
- 删除映射：删除后将剩余映射重编号 `0..n-1`
- 拖拽排序：重编号 `0..n-1`（现行为已符合，维持）
- 移除 `freeWeight` 避让逻辑——创建不再接受显式权重语义
- 全部映射查询补充 `id DESC` 次级排序，保证存量重复权重下行序在自愈前也不飘
- 前端映射表**移除权重列**；拖拽保存成功后重新拉取服务端列表，不再本地推算权重
- **BREAKING**（API）：创建映射请求中的 `weight` 参数被忽略（保留字段以兼容旧客户端）

**不做**：不做存量数据启动迁移——靠后续新增/拖拽操作自愈归一化。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `model-mapping`: 权重分配策略从"最小权重减一 / 显式冲突避让顺延"改为"恒定连续序列 + 三处归一化"；新增权重归一化与查询排序确定性的行为契约。

## Impact

- `server/internal/handler/model.go`：`CreateMapping`（忽略 weight + 归一化重编号）、`DeleteMapping`（删除后归一化）、删除 `freeWeight`、保留 `lockModelMappings`
- `server/internal/handler/model.go`、`server/internal/handler/key.go`、`server/internal/router/router.go`：5 处映射查询补 `Order("weight DESC, id DESC")`
- `web/src/views/Models/Detail.vue`：移除权重列；拖拽保存后改为 `fetchModel()` 重拉列表
- DB：无 schema 变更；存量重复/非连续权重由后续操作自愈
- 旧客户端传 `weight` 不报错但不生效
