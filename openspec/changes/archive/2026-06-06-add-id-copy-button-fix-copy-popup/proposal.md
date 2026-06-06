## Why

厂商模型页面（Providers/Detail.vue）的 model_id 列没有复制按钮，用户选中文本复制时会触发行点击事件弹出详情对话框。MCP 页面的 ID 同样缺乏便捷复制方式。项目中已有 7 处独立实现的 `copyText` 函数，代码高度重复。

## What Changes

- 在厂商模型页面的 model_id 列添加复制按钮，并使用 `@click.stop` 阻止事件冒泡，避免点击复制按钮时触发 `@row-click` 打开详情对话框
- 在 MCP 列表页面的 name 列和 MCP 详情页面的子项名称列添加复制按钮
- 创建可复用的 `CopyButton` 组件和 `useCopyText` composable，消除项目中重复的复制逻辑

## Capabilities

### New Capabilities
- `copy-button-component`: 可复用的复制按钮组件及 useCopyText composable，封装剪贴板写入和消息提示逻辑

### Modified Capabilities
- `model-management`: 厂商模型页面的 model_id 列增加复制按钮，修复点击复制触发弹窗的问题
- `mcp-service-management`: MCP 列表及详情页面的 ID/名称列增加复制按钮

## Impact

- 前端代码：新增 `CopyButton` 组件和 `useCopyText` composable
- 前端代码：修改 `Providers/Detail.vue`、`MCPs/index.vue`、`MCPs/Detail.vue`
- 前端代码：可选重构已有 7 处 `copyText` 内联实现为统一组件调用
- 无 API 变更，无数据库变更
