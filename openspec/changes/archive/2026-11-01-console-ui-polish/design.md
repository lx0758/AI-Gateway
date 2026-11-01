# Design

## Context

两处展示问题的现状：

- 厂商详情页模型列表（`web/src/views/Providers/Detail.vue` L34）：模型 ID 列 `width="200"`，单元格内 ID 后紧跟复制按钮；长 ID 被裁剪且无查看途径
- 用量页（`web/src/views/ModelUsage/index.vue`）：日志表"厂商/模型"列 L174 `width="150"`，内容为 `<el-tag>{{ provider_name }}/{{ actual_model_name }}</el-tag>`，超长直接被裁剪；"厂商模型统计"卡片 L129 的模型列无任何宽度约束与保护

同页已有可复用的 tooltip 模式：`client_ips`（L163）、`error_msg`（L200）、日志表 IP 列（L55）均使用 `el-tooltip` 包裹并以内容长度做 `:disabled` 判断。动机与范围见 proposal.md。

## Goals / Non-Goals

**Goals:**

- 长模型 ID 与长厂商/模型名可完整查看
- 复用页面内已有的 tooltip 写法，不引入新模式

**Non-Goals:**

- 不改动其他页面的列宽与展示
- 不引入虚拟滚动、自动换行等更重的方案
- 不改动任何后端接口

## Decisions

### D1: 模型 ID 列固定加宽而非自适应

`width="200"` → `width="300"`。

**替代**：`min-width` + `show-overflow-tooltip` —— 被拒：用户明确要求固定加宽；且复制按钮与 ID 同格，`show-overflow-tooltip` 的裁剪样式会与按钮布局相互干扰。固定 300px 让短 ID 行也保持一致行高与对齐。

### D2: tooltip 包裹整个单元格内容而非仅截断时触发

日志表"厂商/模型"列与统计卡片模型列均用 `el-tooltip` 包裹 tag/文本。是否以长度做 `:disabled` 由实现时按内容长度判断，沿用同页 `error_msg` 的 `:disabled="len < N"` 写法，避免短内容也弹气泡。

### D3: 统计卡片一并处理

"厂商模型统计"卡片的模型列虽无固定宽度，但列数多（模型/厂商/调用次数/Tokens/平均耗时）时同样会被挤压。与日志表一并加 tooltip，保持同页行为一致。

## Risks / Trade-offs

- **固定 300px 在窄屏下挤压其他列**：用量页与详情页均有横向滚动条兜底 → 接受：用户明确要求固定值
- **tooltip 增加悬停交互成本**：仅在长内容时触发，短内容不弹 → 影响可忽略
