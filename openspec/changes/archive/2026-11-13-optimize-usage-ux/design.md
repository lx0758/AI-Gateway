# Design

## Context

见 proposal.md - Why。与实现相关的现状约束：

- **白名单是隐式的**：`VerifyKeyID`（`proxy.go`）与 `hasFullTool/Resource/PromptAccess`（`proxy_mcp.go`）都用"join 表记录数 == 0 ⇒ 全放行"推断权限模式；UI 用单选 [默认|仅允许] + "全部允许"按钮（删除全部记录）表达同一语义。
- **Element Plus 排序机制**：`orderBy` 对列的 `sort-method` 返回值统一乘以方向系数（`order * +reverse`），`default-sort` 走同一路径——因此第二排序因子会**自动与主列同向**。本变更有意对抗该行为（要求名称顺序恒定：字母在前+中文拼音，不随方向翻转），需在 `sort-method` 内预乘方向系数抵消，见 D3。
- **localStorage 排序偏好**：`tableSort.ts` 的 `getSortConfig(key, defaultProp)`，规格要求"恢复已保存偏好"。
- **两个 Update 接口已支持 `enabled`**（`updateProviderRequest`/`updateKeyRequest` 均为 `*bool`），行内状态开关已在用；两个 Create 接口不接受 `enabled`，硬编码 `true`。
- **capabilities 三层分布**：前端 4 处列/1 处表单 + 后端 3 个响应结构与交集计算 + 10 个规格文件；DB 字段 `provider_models.capabilities` 与同步写入（Anthropic 同步器、smart-model-sync 合并）按决策保留。
- **映射对话框**：`@change="loadProviderModels"` 只替换选项列表，`provider_model_id` 不重置，陈旧 ID 一路通过前端 `required` 校验，被后端 `provider mismatch`（`model.go:319/397`）拒绝。

## Goals / Non-Goals

**Goals:**

- 白名单执行路径收敛为"开关驱动"的两段式判定，四处同构逻辑（models/tools/resources/prompts）行为一致
- 排序逻辑收敛为 `tableSort.ts` 中一个共享助手，避免 8 个视图各自手写比较器
- 五个议题的前后端改动均落在现有代码结构内，不引入新依赖、不重构无关代码

**Non-Goals:**

- 不自动迁移存量白名单数据（管理员手动清理，见 Migration Plan）
- 不改动 MCP 协议 `capabilities`（`mcp/manager.go`）、不改动 capabilities 的同步写入链路
- 不修复规格中其余历史遗留描述（如 AliasID 引用、`is_available` 旧字段名），仅处理本次五个议题触及的部分
- 不调整 Keys 列表的计数列显示逻辑（白名单关闭时记录为 0，计数自然为 0，维持现状）

## Decisions

### D1: 白名单开关存储 —— `keys` 表 4 个独立布尔列

```
keys 表新增（GORM AutoMigrate，SQLite/Postgres 均为 ADD COLUMN）:
  models_whitelist    bool  default false
  tools_whitelist     bool  default false
  resources_whitelist bool  default false
  prompts_whitelist   bool  default false

数据不变量: flag = false ⇒ 该类型 join 表记录为空
  （运行时由"关闭开关"动作保证；存量数据由管理员手动清理）
```

**备选与取舍**：
- *单 JSON 列*：省列数，但每次校验需 unmarshal、无法 SQL 查询、每个 handler 要包解析层——4 个固定类型不值得这层抽象
- *独立关联表存开关*：过度设计
- *继续隐式推导*：无法表达"开启但 0 勾选 = 全部拒绝"这一必需状态（空记录与"关"不可区分），已被需求否定

### D2: 执行路径 —— 开关短路的两段式判定

```
flag == false ──> 直接放行（不查 join 表）
flag == true  ──> 查 join 表: 有记录=允许, 无记录=拒绝
```

- `proxy.go VerifyKeyID`：读取 `apiKey.ModelsWhitelist`，false 直接 `return nil`
- `proxy_mcp.go hasFullTool/Resource/PromptAccess`：改为返回 `!flag`；`tools/list`、`resources/list`、`prompts/list` 及 initialize 过滤、逐项调用检查全部改由 flag 驱动
- `key.go` 新增 `PUT /keys/:id/whitelist/:type`（body `{"enabled": bool}`）：置位；`enabled=false` 时在同一事务内 `DELETE` 该类型全部关联记录（清库动作放后端，前端不单独调 DELETE）
- `keyResponse`/`Get` 响应增加 4 个白名单字段；`Create` 携带 `req.Models` 时同步置 `models_whitelist=true`（维持 D1 不变量）
- "开启白名单"的确认警示由**前端**在调用 API 前弹出（`ElMessageBox.confirm`），后端不感知

### D3: 排序 —— `tableSort.ts` 共享助手 + 各状态列挂 `sort-method`

```typescript
// 名称比较：非中文块在前按 en 字母序，中文块在后按 zh-CN 拼音序（分块 + 显式 locale）
export function compareName(a: string, b: string): number
// 升序语义: 启用(true)在前; 状态相同时按 compareName, tableRef 用于读取当前方向
export function sortByStatusThenName(a, b, tableRef, statusProp = 'enabled', nameProp = 'name'): number
// 数据源预排序：按某列原始值升序（与 Element Plus 内置比较一致），承载默认顺序
export function sortByPropAsc(list, prop = 'name'): T[]
```

- 4 个主列表 + 详情页子表的可排序状态列统一挂 `:sort-method`；`Providers/Detail` 的 `availability` 三值字符串走同一助手（比较函数对非布尔值退化为字符串比较）
- **方向感知**：`orderBy` 把 `sort-method` 返回值整体乘方向系数，故名称次因子需**预乘当前方向**予以抵消，才能恒为固定顺序。方向在 `sortMethod` 执行瞬间直接读 `tableRef.store.states.sortOrder`——不能依赖 `sort-change` 事件（它在 `execSort` **之后**才 emit，且 `default-sort` 初始应用 `init:true` 不发事件，用事件维护会永久滞后一拍）
- **中英顺序（8.3 走查反馈 #2）**：名称比较不用裸 `localeCompare()`——不带 locale 会跟随系统 locale（zh-CN 下 ICU 把整块汉字排到拉丁字母之前），故 `compareName` 显式分块：非中文在前（`en` 字母序）、中文在后（`zh-CN` 拼音序）。`Models/Detail` 两个下拉的排序同样改用 `compareName`
- 13 个调用点均传入各自表格实例（多表格视图传 `tableRefs[key]`，单表格视图传 `tableRef`）；`Models/Detail` 表格原无 ref，本次补 `ref="tableRef"`
- `Models/Detail` 状态列补 `prop="enabled"` 与 `sortable`
- Providers 默认顺序：`getSortConfig('providers', 'enabled')`——**不更换 localStorage 存储键**，已存偏好优先，仅无偏好用户得到"状态+名称"新默认（符合"排序偏好自动保存"规格）
- 启用在前通过助手内的显式比较实现（`true` 优先于 `false`），不依赖布尔自然序

**备选**：用 Element Plus 列的 `sortBy` 多字段属性——同样可行，但 `sort-method` 方案对"启用在前"这种非自然序更可控，且与现有 `sortByArrayLength`/`sortByDate` 风格一致。

### D4: 默认顺序由数据源预排序承载，取消排序仅清偏好

**核心转变（8.3 走查反馈 #1）**：默认顺序不再依赖 `default-sort`/编程式 `sort()`，改由**取数后对数据数组预排序**（`sortByPropAsc` 或 `sortByStatusThenName`）承担：

- `getSortConfig(key, defaultProp)` 无偏好（或偏好与硬默认相同）时返回 `undefined` → 模板不绑 `default-sort` → **无排序指示器**，表格直接展示数据源顺序
- 用户主动排序 → `setSortConfig` 保存 → 下次访问绑 `default-sort` 恢复（有指示器，符合"排序偏好自动保存"规格）
- **第三次点击取消**：Element Plus 自行复位（`handleSortClick` 置 `column.order=null` 箭头消失 → `changeSortCondition` 清排序状态 → `execSort` 回数据源顺序 → 才 emit `sort-change(order:null)`），故 `handleSortChange` 的 `order===null` 分支**只需 `removeSortConfig(key)`**，不再编程式 `sort()`
- 数据刷新（`setData`）只调 `execQuery`、不发 `sort-change`，不会误清偏好；刷新后无排序状态 → 展示预排序的数据源顺序，与"取消后保持默认顺序"规格一致
- 预排序比较方式与目标列内置比较一致（`sortByPropAsc` 用原始值 `<`），故"无偏好 → 点击该列升序"不引起重排；`Providers` 预排序直接调 `sortByStatusThenName`（与状态列升序结果相同）

**废弃**：原 D4"编程式 `sort(defaultProp, defaultOrder)` 恢复默认"方案——它让厂商页（默认即状态排序）的状态箭头永远无法去掉，且与"默认=没有选择时的排序规则"语义冲突。

### D5: capabilities 移除 —— 响应层与展示层切除，存储与写入保留

| 层 | 动作 |
|---|---|
| 前端展示 | 删 4 处能力列（Models/index、Models/Detail、Providers/Detail、Keys/Detail 死列）、Providers/Detail 编辑表单 checkbox 组与详情描述、`capabilityList`/`updateCapabilityList`/`getCapabilitiesString` 辅助函数 |
| 后端响应 | `model.go`、`key.go`、`provider_model.go` 的响应结构删 `Capabilities` 字段；删除 `calculateCapabilitiesIntersection` 及全部调用点（`model.go:637`、`key.go:357` 等） |
| 后端输入 | `createProviderModelRequest`/`updateProviderModelRequest` 删 `Capabilities` 字段（新模型该列为空；Update 不再接收该字段） |
| **保留** | `db.go` 的 `provider_models.capabilities` 列、Anthropic 同步器填充、smart-model-sync 的非空合并（`updates["capabilities"]` 仅存在于同步路径，不受 DTO 删字段影响）、MCP 协议 `capabilities` |
| i18n | 删 `models.capabilities`、`provider.capabilities`、`key.allowAll` 及 `provider.tools/stream/photo/image/video` 标签词条（zh/en） |

**风险确认**：`provider_model.go` Update 的空值不覆盖语义意味着前端停止发送不会清空存量数据；同步路径独立于 DTO，写入不受影响。

### D6: 映射对话框级联 —— 前端单点修复，后端不动

`Models/Detail.vue` 模板绑定改为 `@change="handleProviderChange"`：

```
handleProviderChange():
  1. mappingForm.provider_model_id = null   // 同步清空, 陈旧 ID 当场作废
  2. providerModels.value = []              // 清旧选项
  3. await loadProviderModels()             // 拉新厂商模型
```

清空后 `required` 校验自然拦截保存。后端 `provider mismatch` 校验保留作 API 层防线（已有规格场景覆盖）。

### D7: 表单状态控制 —— 对话框加开关，Create 补可选字段

- `Providers/index.vue`、`Keys/index.vue` 的 form 增加 `enabled`，`showDialog` 编辑时回填、新增时默认 `true`，提交时随 body 发送
- `createProviderRequest`/`createKeyRequest` 增加 `Enabled *bool`，Create 中 `enabled = req.Enabled == nil || *req.Enabled`（缺省 true，旧客户端零感知）
- 参照实现：MCPs 对话框的开关写法；条件渲染风格参照 `api_key` 字段的编辑/新增差异

## Risks / Trade-offs

- **[BREAKING: 存量受限密钥升级后变全放行]** → 已与需求方确认不迁移；升级说明中明确"开关默认关=全部允许"，并附存量关联记录的手动清理语句（`DELETE FROM key_models/key_mcp_tools/key_mcp_resources/key_mcp_prompts`，或经由 UI 逐密钥处理）；数据不变量靠管理员维护
- **[开启白名单瞬间全部拒绝，误伤线上调用]** → 规格已要求开启前确认弹窗（"将拒绝所有调用"）；表格常驻提示当前已选数量
- **[capabilities 从响应移除影响外部消费者]** → 该 API 为控制台内部使用，无对外版本化契约；proposal 已标注 BREAKING
- **[sort-method 依赖 orderBy 乘方向的内部行为]** → Element Plus 版本已锁定在 node_modules；方向须在 sortMethod 内直接读 `store.states.sortOrder`（`sort-change` 滞后一拍、`default-sort` init 不发事件），实现时对升/降序各做一次手测确认次因子顺序恒定（不随方向翻转）
- **[读取 store.states.sortOrder 属 Element Plus 内部结构]** → 已确认 `states` 为普通对象装 ref（`helper.mjs` 直接写 `.value`）且 `store` 随 `setup()` 返回暴露；若未来版本变动，`readOrder` 有 try/catch 兜底降级为升序（名称次因子退化为随方向翻转，不影响状态主因子）
- **[4 列 AutoMigrate 在 SQLite/PG 的兼容性]** → 两库 `ADD COLUMN` 均为幂等安全操作，GORM AutoMigrate 已在本项目用于既有表变更（`postgres-support` 规格覆盖）
- **[默认顺序依赖前端预排序，后端返回序无保证]** → SQL 无 ORDER BY 时返回序不被标准保证；预排序在每次取数后执行，展示顺序与后端返回序解耦（运行时排序架构已与需求方确认保持不变）
- **[偏好与硬默认相同时视为无偏好]** → `getSortConfig` 对"配置=该页硬默认"返回 undefined（无指示器），既兼容旧版取消逻辑写入的 fallback，也符合"默认=没有一个选择"语义；代价是用户显式选择默认排序后刷新不显示指示器（数据顺序无差异）

## Migration Plan

1. **后端先行**：`db.go` 加列 → 白名单执行路径与 API → Create 补 `enabled` → capabilities 响应移除（同一提交序列，AutoMigrate 随服务启动执行）
2. **前端跟随**：排序助手与状态列 → Keys/Detail 白名单 UI → capabilities 展示删除 → 映射对话框与两个表单
3. **升级说明**：随版本发布注明两处 BREAKING——存量白名单记录手动清理指引、capabilities 字段不再出现在响应中
4. **回滚策略**：4 个新列保留无害，可直接回滚代码版本；因 capabilities 属响应字段删除，前后端需作为同一版本整体回滚，不支持混部

## Open Questions

无——五个议题的决策已在探索阶段全部收口（含白名单迁移、开启初始态、排序方向与范围、capabilities 移除边界）。
