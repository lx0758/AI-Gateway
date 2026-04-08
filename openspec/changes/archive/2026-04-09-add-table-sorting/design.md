## Context

前端使用 Vue 3 + Element Plus，Element Plus 的 `el-table` 组件内置了排序功能，只需通过 `sortable` 属性启用即可。当前四个列表页面（Models、Providers、Keys、MCPs）均未启用排序功能，用户只能按数据返回的默认顺序查看。

数据来源为后端 API，数据量预期在百级到千级，前端排序性能足够。

**现有问题**：
- Providers 列表在后端按 name 排序后返回，但用户无法改变排序方式
- 其他列表无任何排序，完全依赖后端返回顺序
- 子页面（Detail 页面）的列表也未支持排序
- 厂商模型页面的 `is_available` 字段只能查看不能修改
- 父对象禁用时，子对象的状态控制不够直观

## Goals / Non-Goals

**Goals:**
- 为主列表和子页面列表添加前端排序功能
- 支持多列排序（不同列有不同排序逻辑）
- 用户排序偏好持久化（localStorage）
- 空值统一排在末尾
- 默认按名称升序排序
- 支持修改模型的可用于调用状态（is_available）
- 父对象禁用时子对象状态显示为禁用但不修改数据库

**Non-Goals:**
- 不涉及后端排序或 API 变更（is_available 除外）
- 不支持多字段组合排序（Shift+点击）
- 不支持排序状态的导出/导入

## Decisions

### Decision 1: 前端排序 vs 后端排序

**选择**: 纯前端排序

**理由**:
- 数据量小（百级到千级），前端性能足够
- 实现简单，无需修改后端 API
- Element Plus 表格已内置排序能力，开箱即用
- 响应更快，无需等待网络请求

**替代方案**: 后端排序（通过 API 参数）
- 缺点：需要修改后端 API，增加复杂度
- 适用场景：数据量大（万级以上）时考虑

### Decision 2: 排序持久化方案

**选择**: localStorage，Key 格式 `table_sort_{page_name}`

**理由**:
- 简单直接，无需后端支持
- 用户级别的偏好，不需要跨设备同步
- Key 命名清晰，方便调试

**存储结构**:
```typescript
interface SortConfig {
  prop: string        // 排序字段
  order: 'ascending' | 'descending'  // 排序方向
}
```

### Decision 3: 空值处理策略

**选择**: 空值排在末尾（无论升降序）

**理由**:
- 符合用户直觉：有值的记录更重要
- 便于快速定位有效数据

**实现**: 自定义 sort-method，在比较前检查空值

### Decision 4: 默认排序字段

**选择**: 名称升序

**理由**:
- 名称是最常用的查找依据
- 升序符合直觉：按字母顺序排列
- 便于用户快速定位

### Decision 5: 工具函数设计

**选择**: 创建 `src/utils/tableSort.ts` 提供通用函数

**理由**:
- 代码复用：多个列表共用相同的排序逻辑
- 统一维护：修改排序行为只需改一处
- 职责分离：排序逻辑与页面逻辑解耦

**导出函数**:
- `getSortConfig(key, defaultProp?)` - 从 localStorage 读取排序配置
- `setSortConfig(key, config)` - 保存排序配置到 localStorage
- `sortByDate(a, b, prop)` - 日期排序（空值排末尾）
- `sortByArrayLength(a, b, prop)` - 数组长度排序

### Decision 6: 父对象禁用时的子对象状态显示

**选择**: 显示禁用状态但不修改数据库

**理由**:
- 用户体验：清晰表达"父对象禁用则子对象不可用"
- 数据安全：不因 UI 展示而意外修改数据
- 可逆性：父对象重新启用后，子对象状态保持不变

**实现**:
```vue
<el-tooltip v-if="!parent?.enabled" :content="t('xxx.parentDisabled')">
  <el-switch :model-value="false" disabled />
</el-tooltip>
<el-switch v-else v-model="row.enabled" @change="toggleEnabled(row)" />
```

### Decision 7: 模型可用性（is_available）字段

**选择**: 支持前端切换，后端新增更新能力

**理由**:
- 业务需求：厂商可能下架某个模型，需要手动禁用
- 用户控制：用户可以根据实际情况决定是否使用某个模型
- 路由影响：is_available=false 的模型不参与路由转发

**后端改动**:
- `createProviderModelRequest` 结构体添加 `IsAvailable *bool` 字段
- `Update` 方法支持更新 `is_available` 字段

## Risks / Trade-offs

**Risk 1: localStorage 数据格式变更**
- **风险**: 未来修改 SortConfig 结构可能导致旧数据无法解析
- **缓解**: 版本号或迁移逻辑（当前数据结构简单，暂不需要）

**Risk 2: 数据量大时性能问题**
- **风险**: 数据增长到万级时，前端排序可能卡顿
- **缓解**: 监控数据量，超过阈值时切换为后端排序

**Risk 3: 不同浏览器 localStorage 限制**
- **风险**: localStorage 有 5MB 限制，大量排序配置可能占满
- **缓解**: 排序配置极小（<100 bytes），实际不可能触及限制

## Open Questions

无
