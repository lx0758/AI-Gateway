## Context

当前代码库中存在一个名为 `manufacturer` 的包，位于 `server/internal/manufacturer/`。该包负责：
- 定义 `Manufacturer` 接口，用于处理不同 AI 提供商的请求执行
- 实现 OpenAI 和 Anthropic 两种制造商的具体实现
- 提供工厂模式创建对应的制造商实例

该包在之前的重构中被命名为 `manufacturer`，但业界通用术语为 `provider`，且该包原本就叫 `provider`。为保持一致性和可读性，需要进行重命名。

## Goals / Non-Goals

**Goals:**
- 将 `manufacturer` 包及其内容重命名为 `provider` 相关命名
- 保持所有功能不变，仅进行命名变更
- 更新所有引用该包的代码

**Non-Goals:**
- 不改变任何功能逻辑
- 不修改 API 接口或数据结构
- 不添加新功能

## Decisions

### 1. 接口命名：`Manufacturer` → `Provider`

**选择 `Provider` 的原因：**
- 与包名 `provider` 保持一致
- 在包内部使用时，`provider.Provider` 清晰明了
- 使用时通过 `p := provider.NewFactory().Create(...)` 形式调用，不会与 `model.Provider` 混淆

### 2. 文件命名：`manufacturer_*.go` → `provider_*.go`

**选择：** 保持文件名简洁，使用 `provider_` 前缀

### 3. 工厂文件保持 `factory.go`

**选择：** `factory.go` 文件名保持不变，只更新包名和内部引用

## Risks / Trade-offs

**风险：** 导入路径变更可能导致其他分支或未提交代码出现编译错误
→ **缓解：** 这是一个小型重构，所有引用都在当前代码库中，可以一次性完成

**风险：** `provider.Provider` 与 `model.Provider` 可能造成混淆
→ **缓解：** 在代码中使用时通过包名前缀区分，如 `provider.Provider` 和 `model.Provider`
