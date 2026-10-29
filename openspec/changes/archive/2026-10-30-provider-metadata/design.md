# Design

## Context

Provider 实体（`server/internal/model/db.go`）现有字段：Name、OpenAIBaseURL、AnthropicBaseURL、APIKey、Enabled、Priority、Config、LastSyncAt；handler DTO 三件套在 `server/internal/handler/provider.go`（create L17 / update L25 / response L34）。厂商列表 5 列（选择 / 名称 / API 风格 / 模型数 / 状态 / 操作）。动机见 proposal.md。

## Goals / Non-Goals

**Goals:**

- 为厂商持久化两个可选元信息：官方主页、付费类型
- 主页一键直达厂商控制台；付费类型直观可见
- 零手工迁移，向后兼容

**Non-Goals:**

- 不做付费金额、周期等计费细节（仅类型标签）
- 不做按付费类型的筛选 / 统计
- 不强校验 homepage 的可达性（仅格式校验，存储宽松）

## Decisions

### D1: billing_type 存英文枚举值，展示层 i18n

值域 `subscription` / `usage` / `free`，空 = 未设置。理由：

- 与项目现有约定一致（`ProviderModel.Source` 存 `"sync"`/`"manual"`、`source` 同理）
- 存中文则英文界面无法翻译、无法规范化（"套餐"/"订阅"/"包月"歧义）；存整数（0/1/2）SQL 可读性差、排查靠猜
- 展示层用 i18n key（`provider.billing.*`）映射，中英文各自翻译

**替代**：自由文本——被拒：无法约束值域、无法保证展示一致性。

### D2: 展示位置——列表 name 列内嵌图标 + 详情页头部

- 列表：homepage 在名称列文本后放小图标（复用现有 `CopyButton` 的内嵌模式），不新增列——列表已 5 列，再加两列会拥挤；付费类型新增一列 tag（约 100px）
- 详情页：头部 `el-descriptions` 区域加主页链接与付费类型 tag

### D3: 校验策略——前端格式校验 + 后端值域校验

- `homepage`：前端 URL 格式校验（表单层）；后端不做格式强校验（宽松存储，避免厂商 URL 的怪异格式被拒）
- `billing_type`：后端 `binding:"omitempty,oneof=subscription usage free"` 拒绝非法值；前端下拉框从枚举渲染，天然不产生非法值
- 两字段均可选，空值合法

### D4: 迁移与兼容

GORM AutoMigrate 为 SQLite/PostgreSQL 自动加列（默认空字符串），无手工迁移。旧客户端不传字段时存储空值；响应新增字段对旧前端无影响。

## Risks / Trade-offs

- **枚举扩展**：将来新增付费类型（如 `free_tier`）需前后端 + 翻译同步加值 → 可接受：单点小改，oneof 校验集中一处
- **homepage 内容不受控**：存任意字符串，可能存错 → 可接受：仅辅助记录，跳转失败无实害；格式校验已拦明显错误
