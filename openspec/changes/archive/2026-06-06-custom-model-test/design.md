## Context

当前 Provider 详情页的模型测试功能（`TestProviderModel`）要求模型必须已存在于数据库中（通过 `provider_id` + 模型数据库 ID 查找 `ProviderModel` 记录）。用户在添加新模型之前无法验证该模型 ID 在对应 Provider 上是否可用，必须走"添加 → 测试 → 可能删除"的流程。

现有测试流程复用了 `provider.NewAutomatedProvider` + `ExecuteOpenAIRequest` / `ExecuteAnthropicRequest`，通过 `httptest` 创建测试上下文执行请求。核心参数仅依赖 `Provider`（BaseURL、APIKey）和 `ProviderModel.ModelID`。

## Goals / Non-Goals

**Goals:**
- 允许用户输入任意模型 ID 对指定 Provider 执行测试，无需预先创建模型记录
- Provider 详情页新增"自定义测试"入口，交互简洁（按钮 → 输入模型 ID → 查看结果）
- 后端复用现有测试执行逻辑（`executeTest`），最小化代码改动

**Non-Goals:**
- 不支持自定义测试消息、max_tokens 等参数（使用固定测试参数）
- 不修改现有模型测试 API 的行为
- 不在虚拟模型（Model）页面增加自定义测试功能

## Decisions

### 1. 新增独立 API 端点 vs 复用现有端点

**决定**: 新增 `POST /api/v1/providers/:id/test-custom` 端点，请求体包含 `model_id` 字段。

**理由**: 现有 `TestProviderModel` 端点的路径参数 `:mid` 是数据库 ID 而非模型 ID 字符串，修改语义会引入歧义且影响现有调用方。独立端点职责清晰，不影响现有 API。

**备选**: 在现有端点上通过 query parameter 区分 —— 放弃，因为路由参数 `:mid` 的含义（数据库 ID）已被前端使用，混用会增加理解成本。

### 2. 后端构造临时 ProviderModel 对象

**决定**: 在 handler 中构造一个临时的 `ProviderModel` 结构体，仅填充 `ModelID` 字段，传递给现有 `executeTest` 函数。

**理由**: `executeTest` 函数仅使用 `ProviderModel.ModelID` 来构建请求体中的 `model` 字段，其他字段（display_name、capabilities 等）不影响测试执行。构造临时对象可完全复用现有逻辑。

### 3. 前端交互设计

**决定**: 在 Provider 详情页的 actions 区域新增"自定义测试"按钮，点击后弹出 `ElMessageBox.prompt` 输入模型 ID，确认后调用 API 并复用现有测试结果对话框展示。

**理由**: `ElMessageBox.prompt` 是 Element Plus 内置组件，无需自定义对话框，交互最简。测试结果展示格式与现有 `testModel` 完全一致，复用 `testDialogVisible` 和 `testResults`。

## Risks / Trade-offs

- **临时 ProviderModel 缺少校验字段**: 构造的临时对象没有 display_name、capabilities 等信息，但测试结果响应中的 model 信息会显示用户输入的 model_id，这是可接受的 → 响应中 display_name 显示为空或模型 ID 本身
- **用户输入非法模型 ID**: API 请求可能返回错误 → 这是预期行为，测试结果会显示错误信息，用户可据此调整
- **未鉴权的模型 ID 注入**: 用户可输入任意字符串 → 不构成安全风险，仅用于 API 调用测试，不会写入数据库
