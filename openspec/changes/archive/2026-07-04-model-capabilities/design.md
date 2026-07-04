## Context

当前 ProviderModel 的能力用三个独立 bool 字段表示（supports_vision / supports_tools / supports_stream），上下文窗口用纯数字输入，用户习惯用 `256K`/`1M` 等单位表示，保存时需要手动换算。

## Goals / Non-Goals

**Goals:**
- 模型能力改为多选 checkbox，支持 5 种能力类型（Tools / Stream / Photo / Image / Video），全不选显示 None
- 移除旧的 SupportsVision / SupportsTools / SupportsStream 三个 bool 字段，统一用 capabilities 字符串
- 上下文窗口改为 1024-based 单位显示，支持灵活文本输入（`256k` / `1M`），保存时自动解析为数字
- 迁移已有数据：从旧 bool 值反推 capabilities 字符串

**Non-Goals:**
- 不改数据库 schema 中 context_window 的类型（仍为 int）
- 不改 model 级别的路由逻辑
- 不改 provider sync API 的协议

## Decisions

### 决策 1: capabilities 用逗号分隔字符串存储

**选择:** 在 ProviderModel 表新增 `capabilities` 列（text），存储逗号分隔的能力标识，如 `"tools,stream"`

**理由:**
- 简单，不需要新建关联表
- 前端多选直接映射为逗号拼接，解析简单
- 与现有 metadata 字段风格一致（都是 text）
- 同步时 provider API 返回的能力列表可直接写入

**备选:**
- 用 JSON 数组: 更结构化但解析稍复杂，GORM 不原生支持
- 新建 ProviderModelCapability 关联表: 过度设计，5 个选项不需要

### 决策 2: 完全移除旧 bool 字段

**选择:** 移除 SupportsVision / SupportsTools / SupportsStream，用 capabilities 字符串替代

**理由:**
- 语义重复：三 bool 与新多选框描述同一组能力
- 避免数据不一致：两套字段并存会导致矛盾
- 简化维护：少维护一套字段和映射逻辑

**影响:** 需要从 capabilities 字符串推导原 bool 值的逻辑，用于向后兼容和已有数据迁移

**映射规则:**
```
capabilities = "tools,stream"        → SupportsTools=true, SupportsStream=true, SupportsVision=false
capabilities = "image,video"         → SupportsVision=true, SupportsTools=false, SupportsStream=false
capabilities = "tools,stream,image"  → 三者均为 true
capabilities = "" (空)               → 三者均为 false (等同于 "none")
```

### 决策 3: 已有数据迁移（手动处理）

**选择:** 不在代码中自动迁移，生成迁移手册由管理员手动执行。

**迁移规则:**
```
SupportsTools=true + SupportsStream=true + SupportsVision=false → capabilities = "tools,stream"
SupportsTools=true + SupportsStream=true + SupportsVision=true  → capabilities = "tools,stream,image"
仅 SupportsStream=true                                        → capabilities = "stream"
全为 false                                                      → capabilities = ""
```

**手动迁移 SQL 示例:**
```sql
UPDATE provider_models SET capabilities = CASE
  WHEN supports_tools AND supports_stream AND supports_vision THEN 'tools,stream,image'
  WHEN supports_tools AND supports_stream THEN 'tools,stream'
  WHEN supports_tools AND supports_vision THEN 'tools,image'
  WHEN supports_stream AND supports_vision THEN 'stream,image'
  WHEN supports_tools THEN 'tools'
  WHEN supports_stream THEN 'stream'
  WHEN supports_vision THEN 'image'
  ELSE ''
END;
```

### 决策 4: 前端处理全部 token 单位逻辑

**选择:** token 的解析（parseContextString）和格式化（formatToken / formatContextInput / formatContextDisplay）**全部在前端**完成，后端不做任何单位处理。

**理由:**
- 单位是展示层 concern，后端只关心 raw token count
- 减少后端与前端约定：后端接收/返回纯数字，前端自行处理
- format.ts 集中管理所有 token 相关逻辑（解析 + 格式化），单点维护

## Risks / Trade-offs

| 风险 | 缓解 |
|---|---|
| 迁移脚本失败导致数据丢失 | 迁移前备份 DB，迁移后验证记录数 |
| Sync 时 provider API 不返回能力信息 | capabilities 保持为空，不影响同步流程 |
| 前端 parseContextString 解析失败 | 前端校验提示"请输入有效格式"，不提交 |
| 1024-based 与某些 provider 的 1000-based 认知不同 | 存储层始终是 raw token count，前端统一为 1024-based 展示 |
| 内部 provider 包 (provider_openai.go / provider_anthropic.go) 仍引用旧 bool 字段 | 同步更新 provider 包的 ProviderModel struct，移除旧字段 |

## Migration Plan

1. **手动迁移已有数据**（部署前）:
    - 执行迁移 SQL 将旧 bool 值写入 capabilities 列
    - 参考上方"手动迁移 SQL 示例"

2. **DB Schema 变更** (AutoMigrate):
    - 新增 `capabilities` 列 (text)
    - 移除 `supports_vision` / `supports_tools` / `supports_stream` 三列

3. **后端代码更新**:
    - db.go: ProviderModel 移除三 bool 字段
    - provider_model.go: DTO 移除三 bool 字段，新增 capabilities
    - model.go: calculateCapabilitiesIntersection 改为从 capabilities 推导
    - provider_openai.go / provider_anthropic.go: ProviderModel struct 移除三 bool 字段
    - key.go: 相关 response struct 移除三 bool 字段

4. **前端** (全部 token 单位逻辑在此层):
    - format.ts: formatToken 改为 1024-based，新增 parseContextString / formatContextInput
    - 表单改造：context_window 文本输入，前端解析为数字提交
    - 能力多选 checkbox 组

## Open Questions

无
