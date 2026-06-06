## ADDED Requirements

### Requirement: useCopyText composable
系统 SHALL 提供 `useCopyText` composable，封装剪贴板写入和消息提示逻辑。

#### Scenario: 复制成功
- **WHEN** 调用 `copy(text)` 且 `navigator.clipboard.writeText` 成功
- **THEN** 系统将文本写入剪贴板
- **AND** 显示 `ElMessage.success(t('common.copied'))` 提示

#### Scenario: 复制失败
- **WHEN** 调用 `copy(text)` 且 `navigator.clipboard.writeText` 失败
- **THEN** 系统显示 `ElMessage.error(t('common.error'))` 提示

#### Scenario: 复制空值
- **WHEN** 调用 `copy(text)` 且 text 为空或 undefined
- **THEN** 系统不执行任何操作

### Requirement: CopyButton 组件
系统 SHALL 提供 `CopyButton` 组件，以小图标按钮形式展示，点击时复制指定文本到剪贴板。

#### Scenario: 点击复制
- **WHEN** 用户点击 CopyButton
- **THEN** 系统将 `text` prop 指定的文本复制到剪贴板
- **AND** 显示成功提示

#### Scenario: 阻止事件冒泡
- **WHEN** 用户点击 CopyButton
- **THEN** click 事件不冒泡到父元素
- **AND** 不会触发父元素的 `@row-click` 或其他点击处理

#### Scenario: 空文本时隐藏
- **WHEN** `text` prop 为空或 undefined
- **THEN** CopyButton 不渲染

#### Scenario: 与行点击共存
- **WHEN** CopyButton 位于有 `@row-click` 的表格行内
- **AND** 用户点击 CopyButton
- **THEN** 仅执行复制操作，不触发行点击事件
