# Design: Disabled Mappings Visible

## Problem
Model mappings associated with disabled providers or unavailable provider models were completely hidden from the model detail page due to the `JOIN providers AND providers.enabled = ? true` filter in queries.

## Decision: Remove the JOIN filter, add forced_disabled status

Instead of keeping the `providers.enabled = true` filter, we remove it from all queries that load mappings. The `toMappingResponse` function now inspects each mapping's Provider and ProviderModel to determine if it's "force disabled":

```
┌─────────────────────────────────────────────────────────────────┐
│  toMappingResponse() logic                                      │
│                                                                 │
│  1. If Provider.Enabled == false → forced_disabled=true        │
│     disable_reason="provider_disabled"                          │
│                                                                 │
│  2. Else if ProviderModel.IsAvailable == false →               │
│     forced_disabled=true, disable_reason="provider_model_"      │
│     unavailable                                               │
│                                                                 │
│  3. Else → forced_disabled=false                               │
│                                                                 │
│  Priority: Provider disabled > ProviderModel unavailable        │
│  (if both, the provider reason wins)                            │
└─────────────────────────────────────────────────────────────────┘
```

### `key.go` — API Key 模型列表
`ListModels` 方法同样移除了 JOIN 过滤。之前虽然模型仍可见（第一层只查 `Model.enabled`），但 `mapping_count` 等统计不准确。现在统计包含所有厂商的映射。

## Backend Changes

### `mappingResponse` struct (server/internal/handler/model.go)
Two new fields:
- `ForcedDisabled bool` — whether the mapping is disabled by an external factor (provider or provider model)
- `DisableReason string` — one of `"provider_disabled"`, `"provider_model_unavailable"`, or empty

### Query changes
Removed `JOIN providers ON providers.id = model_mappings.provider_id AND providers.enabled = ? true` from 5 locations:
- `List()` — model list
- `Get()` — model detail
- `Update()` — model update response
- `ListMappings()` — mappings list API
- `key.go ListModels()` — key's model list mapping count

The Preload("Provider") and Preload("ProviderModel") remain so the association data is still fetched.

### Aggregate functions unchanged
`calculateEnabledCount`, `calculateMinTokens`, `calculateCapabilitiesIntersection` — these already skip mappings where `Provider.Enabled == false`, so their behavior is correct as-is.

## Frontend Changes

### Status column (web/src/views/Models/Detail.vue)
Three states:
1. **Model itself disabled** → gray switch, disabled, tooltip "模型已禁用"
2. **Mapping force disabled** → gray disabled switch, tooltip shows precise reason
3. **Normal** → green/gray toggle switch, user can toggle

```
┌─────────────────────────────────────────────────┐
│ Status column logic                              │
│                                                  │
│ if !modelEnabled → 灰色开关(禁用)                │
│ else if row.forced_disabled → 灰色禁用开关       │
│   + 精确原因 tooltip                              │
│ else → 开关(可切换)                               │
└─────────────────────────────────────────────────┘
```

### Tooltip precision (getDisableTooltip)
Tooltip uses `disable_reason` to show precise message:
- `provider_disabled` → "厂商已禁用，此映射被强制禁用，无法启用"
- `provider_model_unavailable` → "厂商模型不可用，此映射被强制禁用，无法启用"

### i18n keys
- `models.modelDisabled` — "模型已禁用" / "Model is disabled"
- `models.forcedDisabledByProvider` — "厂商已禁用" / "Provider disabled"
- `models.forcedDisabledByModel` — "厂商模型已禁用" / "Provider model disabled"
- `models.forcedDisabledByOther` — "上游已禁用" / "Upstream disabled"
- `models.providerDisabledReason` — "厂商已禁用，此映射被强制禁用，无法启用" / "Provider is disabled, this mapping is force disabled and cannot be enabled"
- `models.providerModelUnavailableReason` — "厂商模型不可用，此映射被强制禁用，无法启用" / "Provider model is unavailable, this mapping is force disabled and cannot be enabled"
- `models.providerOrModelDisabled` — "厂商或模型被禁用" / "Provider or model is disabled"

## Risks / Tradeoffs

- **More data in API responses** — disabled mappings are now included, increasing payload size. Negligible in practice since a model typically has <20 mappings.
- **Backward compatibility** — clients that don't know about `forced_disabled` and `disable_reason` fields will just ignore them (JSON is additive).
