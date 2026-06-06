## Context

当前项目前端使用 Vue 3 + Element Plus，复制功能在 7 处独立实现 `navigator.clipboard.writeText()` + `ElMessage.success`，代码高度重复。厂商模型页面（Providers/Detail.vue）的 `el-table` 使用 `@row-click="showModelDetail"` 监听整行点击，点击 model_id 列区域会触发弹窗打开详情对话框。操作按钮通过 `@click.stop` 阻止冒泡，但 model_id 列没有此处理。

## Goals / Non-Goals

**Goals:**
- 创建可复用的 `CopyButton` 组件和 `useCopyText` composable，统一复制逻辑
- 在厂商模型页面的 model_id 列添加复制按钮，使用 `@click.stop` 阻止事件冒泡，修复点击复制触发弹窗的问题
- 在 MCP 列表页面的 name 列和 MCP 详情页面的子项名称列添加复制按钮
- 将已有 7 处内联 `copyText` 实现替换为统一组件/composable 调用

**Non-Goals:**
- 不修改后端 API
- 不修改表格行点击行为本身（保留 `@row-click` 功能）
- 不为详情对话框中的字段添加复制按钮（仅关注列表和表格中的 ID/名称列）

## Decisions

### 1. 使用 composable + 组件双层方案

**选择**: 创建 `useCopyText` composable 封装复制逻辑，创建 `CopyButton` 组件封装 UI 表现。

**理由**: composable 可在任意场景复用（如 Keys/Detail.vue 中的 API Key 复制不使用按钮），组件提供标准化的按钮样式。双层方案兼顾灵活性和一致性。

**备选**: 仅创建组件 — 部分场景需要复制但不使用按钮形式（如直接点击文本复制），不够灵活。

### 2. CopyButton 使用 `@click.stop` 阻止事件冒泡

**选择**: 在 CopyButton 组件内部对 click 事件调用 `.stop` 修饰符。

**理由**: 厂商模型页面的 `@row-click` 是问题根因，在组件层面阻止冒泡可以从根本上避免此类问题，且不影响其他场景。

### 3. CopyButton 放在 ID 文本旁边，而非单独一列

**选择**: 在 model_id 列的 `#default` 模板中，在文本后紧跟一个小图标按钮。

**理由**: 保持表格列宽不变，视觉上紧凑。用户看到 ID 时自然可以看到旁边的复制按钮。

## Risks / Trade-offs

- [CopyButton 组件内 `.stop` 可能阻止其他需要冒泡的场景] → CopyButton 只在需要时使用，且 `.stop` 仅阻止复制按钮本身的冒泡，不影响同一行其他元素的点击行为
- [重构已有 7 处 copyText 为统一调用] → 属于纯重构，功能不变，回归测试即可验证
