## Why

前端列表页面（Models、Providers、Keys、MCPs）缺少排序功能，用户无法按需组织数据，难以快速定位特定记录或了解数据分布。随着数据量增长，这个问题会变得更加明显。现在添加排序功能可以提升用户体验，让数据管理更高效。

同时，子对象的状态控制也需要优化：
- 厂商模型页面的 `is_available` 字段只能查看不能修改
- 父对象禁用时，子对象的状态开关应显示禁用状态但不改变数据库

## What Changes

### 1. 主列表页面添加排序功能

- **Models 列表**: 支持按名称、映射数量、Token 汇总、状态排序
- **Providers 列表**: 支持按名称、模型数量、状态排序（移除现有后端排序，统一前端处理）
- **Keys 列表**: 支持按名称、模型数、工具数、资源数、提示词数、状态排序
- **MCPs 列表**: 支持按名称、类型、工具数、资源数、提示词数、状态、最后同步时间排序

### 2. 子页面列表添加排序功能

- **Models Detail**: 映射列表
- **Providers Detail**: 模型列表
- **Keys Detail**: models/tools/resources/prompts 列表
- **MCPs Detail**: tools/resources/prompts 列表

### 3. 父对象禁用时子对象状态显示优化

- **Providers Detail**: 厂商禁用时，模型的 `is_available` 开关显示关闭且置灰
- **Models Detail**: 模型禁用时，映射的 `enabled` 开关显示关闭且置灰
- **Keys Detail**: 密钥禁用时，权限 radio-group 显示置灰，"全部允许"按钮禁用
- **MCPs Detail**: 服务禁用时，tools/resources/prompts 的 `enabled` 开关显示关闭且置灰

### 4. 模型可用性切换功能

- 后端 Update 接口支持修改 `is_available` 字段
- 前端厂商详情页模型列表支持切换 `is_available` 开关

### 新增功能特性

- 默认按名称升序排序
- 用户选择的排序方式自动保存到 localStorage
- 空值统一排在列表末尾
- 简单直观的点击表头排序交互

## Capabilities

### New Capabilities

- `table-sorting`: 前端列表排序功能，包括排序工具函数、localStorage 持久化、各列表页面的排序支持

### Modified Capabilities

无

## Impact

**新增文件**:
- `web/src/utils/tableSort.ts` - 排序工具函数

**修改文件**:
- `web/src/views/Models/index.vue` - 主列表排序
- `web/src/views/Models/Detail.vue` - 映射列表排序 + 父禁用状态
- `web/src/views/Providers/index.vue` - 主列表排序
- `web/src/views/Providers/Detail.vue` - 模型列表排序 + is_available 切换 + 父禁用状态
- `web/src/views/Keys/index.vue` - 主列表排序
- `web/src/views/Keys/Detail.vue` - 子列表排序 + 父禁用状态
- `web/src/views/MCPs/index.vue` - 主列表排序
- `web/src/views/MCPs/Detail.vue` - 子列表排序 + 父禁用状态
- `web/src/locales/zh.ts` - 新增国际化文本
- `web/src/locales/en.ts` - 新增国际化文本
- `server/internal/handler/provider_model.go` - Update 接口支持 is_available

**技术栈**: Element Plus 表格内置排序能力，localStorage 存储

**影响范围**: 前端排序功能，后端仅新增 is_available 字段的更新支持
