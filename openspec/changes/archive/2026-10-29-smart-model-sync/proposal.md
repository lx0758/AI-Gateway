# Proposal

## Why

厂商模型同步存在两个数据正确性问题：

1. **同步覆盖本地信息**：Sync 对已存在的同步来源模型无条件覆盖 display_name / context_window / max_output / capabilities，而同步器本身还会伪造字段（OpenAI 同步器硬编码 `DisplayName=模型ID`、`Capabilities="stream"`、`MaxOutput=0`）。管理员每次同步后手工补充的模型信息会被清空。
2. **无法识别下线模型**：同步只增不减，厂商下线的模型永远保持 `is_available=true`，路由继续把请求打向已下线的模型导致请求失败。

## What Changes

- **模型可用性状态改为三态**：`ProviderModel` 的 `is_available` 布尔字段替换为 `availability` 枚举——`unknown`（未知，从未被同步验证）/ `available`（可用，同步确认在列表）/ `unavailable`（失效，同步确认下线）；手动新建的模型默认 unknown，不再假装"可用"
- **移除 `source` 字段**：三态仅承担失效判断职责——unknown 不参与失效判断（原 manual 排除）；**编辑、删除等管理操作不做任何显隐与限制**（沿用现状）；存量数据迁移读旧 `source` 列后丢弃（manual→unknown、sync 且 true→available、false→unavailable）
- **同步接管行为**：手动添加的模型若出现在厂商同步列表中，将被同步接管（状态推进为 available，字段按空值保护合并）——已确认接受
- 同步器不再伪造字段值：OpenAI / Anthropic 同步器对厂商接口未返回的字段留空（不再用模型 ID 冒充 display_name，不再硬编码 capabilities="stream"）
- 同步合并逻辑改为逐字段判断：同步值为空/零值则**保留本地值**，有效才覆盖
- 失效识别：同步成功且返回列表**非空**时，处于 `available` 且不在返回列表中的模型标记为 `unavailable`（自动退出路由与统计）；`unknown` 模型不参与失效判断
- 自动恢复：`unavailable` 模型回归同步列表时恢复 `available`
- 空列表保护：同步返回空列表时不执行任何状态推进（防接口异常误灭全部模型）
- Sync 响应增加 `deactivated` 计数，前端同步完成后提示"标记 N 个失效模型"
- 聚合统计修复：min_context / max_output 计算与能力交集计算排除 `unavailable` 的映射，使展示与真实路由能力一致
- 控制台来源标签（Manual/Sync）替换为三态可用性标签（未知/可用/失效）；编辑/删除入口不做显隐、操作不限制

**不做**：不增加模型可用性的手动切换开关（失效/恢复完全由同步自动管理）；不做双协议同步来源区分（按整份列表统一判断）；**不做删除保护与编辑显隐**（任何模型均可编辑、删除，三态仅作展示标签）。

## Capabilities

### New Capabilities

- `provider-model-sync`: 厂商模型同步的合并策略与可用性生命周期——空值保护合并、失效识别、自动恢复、空列表保护、手动模型排除。

### Modified Capabilities

- `model-mapping`: 统计计算口径扩展——min/max token 与能力交集除排除禁用 Provider 的映射外，还需排除失效模型（`availability="unavailable"`）的映射。
- `manual-model-ui`: 来源标签的判断依据从 `source` 改为可用性状态；Manual/Sync 来源标签替换为三态可用性标签（未知/可用/失效）。编辑/删除入口维持现状，不做显隐与限制。

## Impact

- `server/internal/model/db.go`：ProviderModel PO——`IsAvailable bool` 替换为 `Availability string`，移除 `Source`；AutoMigrate + 启动时存量数据迁移（读旧 source 列）
- `server/internal/provider/provider_openai.go`、`server/internal/provider/provider_anthropic.go`：同步器字段填充去伪造
- `server/internal/handler/provider_model.go`：Sync 三态状态机、响应 DTO（`source` → `availability`）、Create/Update 适配、available_only 过滤参数清理（编辑/删除逻辑维持现状）
- `server/internal/handler/model.go`：`toMappingResponse`、`calculateEnabledCount`、`calculateMinTokens`、`calculateCapabilitiesIntersection` 的可用性判断
- `server/internal/router/router.go`、`server/internal/handler/model_testing.go`、`server/internal/handler/usage.go`：可用性判断改为 `!= "unavailable"`
- `web/src/views/Providers/Detail.vue`：来源列/详情 tag 改为三态 tag、同步完成失效计数 toast
- `web/src/views/Models/Detail.vue`：tooltip 与状态展示适配三态
- `web/src/locales/zh.ts` / `en.ts`：三态文案（未知/可用/失效）
- **BREAKING**（API）：ProviderModel 响应 `is_available`/`source` 字段移除，替换为 `availability`；`available_only` 查询参数移除
- DB：列类型 bool→text（SQLite 动态类型无碍）+ 存量数据一次性迁移；回滚需回退代码版本
