# Design

## Context

同步入口为 `POST /api/v1/providers/:id/sync`（`server/internal/handler/provider_model.go` `Sync`，约 L216-272）。现状关键约束：

- 合并逻辑对已存在模型用 map `Updates` 无条件覆盖六字段；GORM map 更新会真实写入零值
- 同步器（`provider_openai.go` / `provider_anthropic.go`）对厂商未返回的字段伪造内容：`DisplayName = 模型ID`、`Capabilities = "stream"`（OpenAI）、`MaxOutput = 0`
- `ProviderModel` 现有 `IsAvailable bool`（default true）与 `Source string`（"sync"/"manual"）两个状态字段；`Source` 承担合并保护、失效排除、删除保护、UI 显隐四项职责，其中删除保护与编辑显隐的 spec（`manual-model-protection` / `manual-model-ui`）规划了但未实现（本 change 均不实现——三态仅承担失效判断，管理操作维持现状，见 D1）
- `is_available` 消费点：路由跳过（`router.go:49`）、映射强制禁用（`model.go:502`）、启用数统计（`model.go:531`）、批量测试跳过（`model_testing.go:262`）、仪表盘统计（`usage.go:57`）、模型列表 `available_only` 过滤（`provider_model.go:76`，前端无人使用的 dead param）
- 聚合统计 `calculateMinTokens` / `calculateCapabilitiesIntersection`（`model.go`）未检查可用性
- 前端来源标签在 `Providers/Detail.vue`（列表列 + 详情 tag，按 `source==='manual'` 渲染）；`Providers/Detail.vue` 编辑/删除按钮无显隐逻辑

动机与范围见 proposal.md。

## Goals / Non-Goals

**Goals:**

- 三态 `availability`（unknown/available/unavailable）成为模型可用性的唯一状态源，诚实表达"未验证"状态
- 移除 `source`，其四项职责由三态承接，行为不回退
- 同步不再丢失手工补充信息；下线模型自动失效并退出路由/统计
- 存量数据无损迁移

**Non-Goals:**

- 不做模型可用性的手动切换开关（失效/恢复完全由同步管理）
- 不做双协议同步来源区分（按整份列表统一判断）
- 不新增 DB 字段或表（`availability` 替换 `is_available`）

## Decisions

### D1: 三态状态机——unknown 承接原 manual 语义

```
状态定义:
  unknown      从未被同步验证 (手动新建; 存量 manual 迁移而来)
  available    最近一次非空同步列表包含它
  unavailable  曾 available, 但最近一次非空同步列表不含它

推进规则 (仅当同步列表非空):
  model_id ∈ 列表                      -> available  (空值保护合并字段)
  model_id ∉ 列表 且 当前 available     -> unavailable
  model_id ∉ 列表 且 当前 unknown      -> 保持 unknown (不参与失效)
  列表为空                              -> 不推进任何状态

消费语义:
  路由可用        = availability != unavailable   (unknown + available 均参与)
  强制禁用标记    = availability == unavailable
  聚合统计排除    = availability == unavailable
  编辑/删除       = 不限                          (维持现状: 按钮不做显隐, 操作不因可用性受限)
```

`unknown` 承接原 `manual` 的失效排除语义：不被失效误伤（手动加的自定义模型厂商列表本就可能没有）；`available`/`unavailable` 承接原 `sync` 语义（同步管理字段）。**编辑、删除等管理操作不做任何显隐与限制**（用户决策）：三态仅作控制台展示标签；被同步验证过的模型删除后由下次同步自然重建为 `available`。

**替代**：保留 source + 加第三态——被拒：双字段语义重叠（"来源"与"验证状态"各表一维但取值强相关），维护两套判断徒增分叉；单字段三态是正交收敛。

### D2: 字段改造与存量迁移

- PO：`IsAvailable bool` → `Availability string`（`gorm:"type:varchar(20);default:unknown"`），删除 `Source`
- AutoMigrate 处理列类型变化（SQLite 动态类型）；**启动时一次性迁移**（AutoMigrate 之后执行，幂等）：

```
读旧 source 列 (迁移期保留列, 读完后清理):
  source = "manual"                      -> availability = 'unknown'
  source = "sync"   且 is_available=true -> availability = 'available'
  is_available = false                   -> availability = 'unavailable'
  (无 source 的存量行按 is_available 映射)
确认迁移完成后删除 source 列
```

关键顺序：**必须先读 source 再迁移**——若直接 `true→available`，存量手动模型会被首次同步误标失效。迁移代码幂等（`availability` 已有值的行跳过）。

### D3: 逐字段合并——"空则保留，有效才覆盖"

| 字段类型 | 有效 | 空 |
|---------|------|-----|
| 字符串（display_name / owned_by / capabilities） | 非空 | 空串 |
| 数值（context_window / max_output） | `> 0` | `= 0` |

仅有效字段写入 map `Updates`（GORM map 只写现 map 中的键，天然保留本地）。合并仅作用于**同步列表中**的模型——不在列表的 unknown 模型字段永远不被触碰。

**替代**：读-改-写整体 Save——被拒：两次 DB 往返、并发丢失更新窗口更大。

### D4: 同步器去伪造——空值从源头产生

OpenAI 同步器：`DisplayName` 不赋值为模型 ID、`Capabilities` 不硬编码 `"stream"`、`MaxOutput` 保持 0；Anthropic 同步器同样处理未真实返回的字段。D3 的"空则保留"依赖源头诚实；源头继续伪造则 D3 失效。

### D5: Sync 处理流程

```
1. 拉取同步列表 models
2. len(models) == 0 -> 跳过一切状态推进, 直接返回 (空列表保护)
3. 合并循环 (D3/D4): 已存在 -> 空值合并 + availability='available'
                     不存在 -> 创建, availability='available'
4. 失效对比: syncedSet = models 的 model_id 集合
   查询该 provider 下 availability='available' 的现有模型
   缺失者 -> 批量置 'unavailable', 计入 deactivated
   (unknown 不查询不推进; unavailable 不重复处理)
5. 更新 last_sync_at, 响应含 deactivated 计数
```

内存集合对比而非 SQL `NOT IN`：列表可达数百条，避免超长 IN 子句。

### D6: 消费点改造清单

| 位置 | 旧判断 | 新判断 |
|------|--------|--------|
| `router.go:49` | `!IsAvailable` | `Availability == "unavailable"` |
| `model.go:502`（强制禁用） | `!IsAvailable` | `== "unavailable"` |
| `model.go:531`（启用数） | `!IsAvailable` | `== "unavailable"` |
| `model.go` MinTokens/Capabilities 交集 | 无检查 | `== "unavailable"` 则跳过 |
| `model_testing.go:262` | `!IsAvailable` | `== "unavailable"` |
| `usage.go:57` | `pm.is_available = ?` | `pm.availability != 'unavailable'` |
| `provider_model.go` Create | `IsAvailable: true` | `Availability: "unknown"`（手动新建） |
| `provider_model.go:76` `available_only` | `is_available=true` | **删除该 dead param**（前端无人使用） |
| 响应 DTO | `is_available`/`source` | `availability`（BREAKING，见 proposal） |

### D7: 前端三态展示

- `Providers/Detail.vue`：来源列改三态 tag（未知 info 灰 / 可用 success 绿 / 失效 danger 红）；详情同步；编辑/删除按钮不做显隐（维持现状）
- `Models/Detail.vue`：`forced_disabled` tooltip 已按 `disable_reason` 工作，`provider_model_unavailable` 语义不变；i18n 补三态文案
- 同步完成 toast：`deactivated > 0` 时提示"标记 N 个失效模型"

## Risks / Trade-offs

- **迁移正确性**：一次性迁移读旧列映射三态，幂等可重入；迁移失败则启动报错阻断（避免半迁移状态运行）→ 缓解：迁移放在事务内
- **双协议误伤**：厂商双 URL 时不同次同步可能走不同协议、列表不同，后一次列表会把前一协议独有模型标失效 → 已接受（方案 B）：同厂商两协议列表重叠度高；空列表保护排除最大误灭风险；实际踩坑再引入协议来源字段（方案 C）作为后续 change
- **"空则保留"残留**：厂商真清空字段时本地旧值残留 → 接受：陈旧信息无害，优于被清空
- **unknown 模型长期不验证**：手动加的模型永不参与失效判断，厂商悄悄下线它不会被发现 → 接受：定位即"用户自主管理"；将来若需验证可加手动"测试"动作（已有单模型测试能力）
- **BREAKING 响应字段**：`is_available`/`source` 移除影响旧客户端脚本 → 控制台同步发布；直连 API 的脚本需适配 `availability`

## Migration Plan

1. 部署新版本：AutoMigrate 改列 → 事务内读旧 `source`/`is_available` 迁移三态 → 删除 `source` 列
2. 首次非空同步按新状态机推进（已迁移正确的存量状态不会误伤）
3. 回滚 = 回退代码版本（旧代码读不了 `availability` 列，需回滚时手工恢复列结构；单机控制台可接受）
