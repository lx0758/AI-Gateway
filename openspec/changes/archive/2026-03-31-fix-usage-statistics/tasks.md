## 1. 后端 - 数据模型调整

- [x] 1.1 修改 `UsageLog` 结构：移除 `PromptTokens`/`CompletionTokens`，新增 `TotalTokens int64`，将 `LatencyMs` 改为 `int64`
- [x] 1.2 删除数据库或重新迁移（项目未上线，可直接重建表）

## 2. 后端 - 数据写入

- [x] 2.1 在 `proxy_openai.go` 中添加请求计时逻辑
- [x] 2.2 在 `proxy_openai.go` 中写入 `UsageLog` 记录（包含 apiKeyID, providerID, model, actualModel, totalTokens, latencyMs, status, errorMsg）
- [x] 2.3 在 `proxy_anthropic.go` 中添加请求计时逻辑
- [x] 2.4 在 `proxy_anthropic.go` 中写入 `UsageLog` 记录

## 3. 后端 - 统计查询优化

- [x] 3.1 修改 `usage.go` 中的 `Stats` 方法：使用 `TotalTokens` 替代分开的字段，增加平均耗时统计
- [x] 3.2 修改 `usage.go` 中的 `Dashboard` 方法：增加总 Tokens 和平均耗时统计
- [x] 3.3 新增厂商统计查询（调用次数、Tokens、平均耗时）
- [x] 3.4 新增 Key 统计查询（调用次数、Tokens、平均耗时）

## 4. 前端 - Dashboard 页面

- [x] 4.1 新增「总 Tokens」统计卡片
- [x] 4.2 新增「平均耗时」统计卡片
- [x] 4.3 新增厂商统计表格（厂商名、调用次数、Tokens、平均耗时）

## 5. 前端 - Usage 页面

- [x] 5.1 调整统计概览：移除 promptTokens，新增平均耗时
- [x] 5.2 新增 Key 统计表格
- [x] 5.3 调整日志列表：移除 prompt_tokens/completion_tokens，保留 totalTokens

## 6. 前端 - API Key 页面

- [x] 6.1 列表新增 Tokens 统计列
- [x] 6.2 列表新增平均耗时列
