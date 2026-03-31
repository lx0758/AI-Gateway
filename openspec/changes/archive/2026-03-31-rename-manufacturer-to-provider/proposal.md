## Why

将 `manufacturer` 包重命名为 `provider` 以与业界术语保持一致。该包原本命名为 `provider`，在之前的重构中被改名为 `manufacturer`，现在恢复为更符合行业惯例的命名。

## What Changes

- 将 `server/internal/manufacturer` 目录重命名为 `server/internal/provider`
- 将包名从 `manufacturer` 改为 `provider`
- 将接口 `Manufacturer` 重命名为 `Provider`
- 将文件名从 `manufacturer_*.go` 重命名为 `provider_*.go`
- 更新所有导入路径：`ai-proxy/internal/manufacturer` → `ai-proxy/internal/provider`
- 更新错误消息中的 "manufacturer" 引用

## Capabilities

### New Capabilities

无 - 这是纯重命名重构，不引入新功能。

### Modified Capabilities

无 - 不改变任何功能需求，仅变更内部实现命名。

## Impact

**代码变更：**
- `server/internal/manufacturer/` 目录下所有文件
- `server/internal/handler/provider_model.go` - 更新导入和使用
- `server/internal/handler/proxy_openai.go` - 更新导入和使用
- `server/internal/handler/proxy_anthropic.go` - 更新导入和使用

**受影响文件：**
- `manufacturer.go` → `provider.go`
- `manufacturer_anthropic.go` → `provider_anthropic.go`
- `manufacturer_anthropic_test.go` → `provider_anthropic_test.go`
- `manufacturer_openai_compatible.go` → `provider_openai_compatible.go`
- `manufacturer_openai_compatible_test.go` → `provider_openai_compatible_test.go`
- `factory.go` - 保持文件名，更新内容
