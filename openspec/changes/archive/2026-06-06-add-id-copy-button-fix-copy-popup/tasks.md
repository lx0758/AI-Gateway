## 1. 创建可复用复制基础设施

- [x] 1.1 创建 `web/src/composables/useCopyText.ts`，封装 `navigator.clipboard.writeText` + `ElMessage` 提示逻辑
- [x] 1.2 创建 `web/src/components/CopyButton.vue` 组件，接受 `text` prop，内部使用 `useCopyText`，点击时 `.stop` 阻止事件冒泡，text 为空时不渲染

## 2. 厂商模型页面 — 添加复制按钮并修复弹窗问题

- [x] 2.1 修改 `web/src/views/Providers/Detail.vue`：将 model_id 列从纯文本 `<el-table-column prop="model_id">` 改为使用 `#default` 模板，在文本后添加 `CopyButton` 组件

## 3. MCP 页面 — 添加复制按钮

- [x] 3.1 修改 `web/src/views/MCPs/index.vue`：将 name 列从纯文本改为使用 `#default` 模板，在文本后添加 `CopyButton` 组件
- [x] 3.2 修改 `web/src/views/MCPs/Detail.vue`：在工具、资源、提示词的 name 列添加 `CopyButton` 组件

## 4. 重构已有内联 copyText 实现

- [x] 4.1 替换 `web/src/views/MCPs/Detail.vue` 中的内联 `copyText` 函数和描述列复制按钮为 `CopyButton` 组件
- [x] 4.2 替换 `web/src/views/Keys/Detail.vue` 中的内联 `copyText` 函数，API Key 复制改用 `useCopyText` composable，描述复制改用 `CopyButton` 组件
- [x] 4.3 替换 `web/src/views/Keys/index.vue` 中的内联 `copyText` 函数为 `useCopyText` composable
- [x] 4.4 替换 `web/src/components/JsonViewer.vue` 中的内联 `copyText` 函数为 `useCopyText` composable
- [x] 4.5 替换 `web/src/views/ModelUsage/index.vue` 和 `web/src/views/MCPUsage/index.vue` 中的内联复制逻辑为 `useCopyText` composable

## 5. 验证

- [x] 5.1 验证厂商模型页面：点击复制按钮不触发弹窗，复制 model_id 成功
- [x] 5.2 验证 MCP 列表页面：复制服务名称成功
- [x] 5.3 验证 MCP 详情页面：复制工具/资源/提示词名称成功
- [x] 5.4 验证已重构页面：API Key 复制、描述复制、JSON 复制等功能正常
