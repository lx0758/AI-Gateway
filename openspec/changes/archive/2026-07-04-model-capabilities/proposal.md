## Why

用户需要更灵活地描述模型能力（支持图片、音频、视频等多种媒体类型），同时需要更直观的上下文窗口输入体验——当前使用纯数字输入，用户习惯用 `256K`、`1M` 等单位表示，保存时需要手动换算，体验差。

## What Changes

### 模型能力多选

- 在 ProviderModel 编辑表单中，将原来三个独立开关（supports_vision / supports_tools / supports_stream）改为**多选 checkbox 组**
- 选项: `Tools(工具)` / `Stream(流式)` / `Photo(照片)` / `Image(图片)` / `Video(视频)`
- 全部不选 → 显示 `None`
- 默认选中: `Tools` + `Stream`
- 后端新增 `capabilities` 字段存储多选结果（逗号分隔，如 `"tools,stream"`）
- 列表/详情展示改为标签组形式

### Context Window 前端 1024-based 处理

- 存储不变: 后端仍为 int (raw token count)
- 所有 token 单位转换（解析/格式化）**在前端完成**:
  - 显示: `formatToken` 改为 1024-based（1K = 1024, 1M = 1048576）
  - 输入: 表单文本框支持 `128k` / `1M` 等格式，前端 `parseContextString` 解析为数字后提交
  - 后端不处理任何 token 单位逻辑
- `context_window` 和 `max_output` 两个字段都支持 1024-based 灵活输入

### 展示优化

- 列表/详情中 context_window 统一以 `K`/`M` 单位显示（1024-based）
- 工具列和流式列从独立 tag 合并到能力标签组中
- 列表能力取**交集**（只有所有启用映射都支持的能力才显示）
- 编辑表单回显能力时正确勾选已存在的能力

## Capabilities

### New Capabilities
- `model-capabilities`: 模型媒体能力多选管理（后端存储 + 前端多选 UI）

### Modified Capabilities
- `model-management`: context window 输入/显示方式改造（1024-based 单位换算 + 灵活输入解析）

## Impact

| 层 | 影响 |
|---|---|
| 后端 DB | ProviderModel 表新增 `capabilities` 列 |
| 后端 Handler | DTO 新增 capabilities 字段；移除旧 bool 字段 |
| 后端 Sync | SyncModels 同步时需传递 capabilities |
| 前端 API | ProviderModel 表单提交格式变化（context_window 仍为数字，不再传文本） |
| 前端 UI | ProviderDetail.vue 表单重构；format.ts 包含全部 token 单位逻辑（parse/format） |
| 前端 i18n | 新增翻译键 |
| 现有数据 | 部署前需手动执行 SQL 迁移，将旧 bool 字段转为 capabilities 字符串 |
