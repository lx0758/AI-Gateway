## 1. 后端 API 扩展

- [x] 1.1 在 `server/internal/handler/alias.go` 中，GET `/aliases` 返回 min_context_window、min_max_output 汇总字段
- [x] 1.2 实现汇总统计逻辑（只统计 enabled=true 的 mappings，取最小值）
- [x] 1.3 在 GET `/aliases/:id` 中，mappings 包含 model_info 字段（context_window, max_output, supports_vision, supports_tools, supports_stream）
- [x] 1.4 新增 PUT `/aliases/:id/mappings/order` handler，接收 order 数组并更新权重
- [x] 1.5 实现权重计算逻辑（位置从上到下：total-1, total-2, ..., 0）
- [ ] 1.6 测试所有 API，验证数据正确性和一致性

## 2. 前端工具函数

- [x] 2.1 在 `web/src/utils/format.ts` 中添加 `formatToken(value: number): string`（保留一位小数）
- [x] 2.2 实现 token 格式化逻辑（< 1000 原数字，>= 1000 XK/X.XK，>= 1000000 XM/X.XM）
- [x] 2.3 添加 `formatContextDisplay(context: number, output: number): string` 函数
- [x] 2.4 编写单元测试验证格式化函数的边界情况（跳过 - 无测试框架）

## 3. Aliases 主页面改版

- [x] 3.1 重构 `web/src/views/Aliases/index.vue`，改为扁平表格布局（移除折叠面板）
- [x] 3.2 实现表格列：选择框、别名名称、映射数量、Token汇总、状态开关、操作按钮
- [x] 3.3 实现批量选择和批量删除功能
- [x] 3.4 实现 Token 汇总列显示（格式化显示最小值组合）
- [x] 3.5 实现详情按钮，跳转到 `/aliases/:id` 页面
- [x] 3.6 更新 TypeScript 接口定义，新增 min_context_window、min_max_output 字段
- [ ] 3.7 测试主页面所有功能（显示、批量删除、跳转）

## 4. Aliases 详情页面新增

- [x] 4.1 创建 `web/src/views/Aliases/Detail.vue` 新页面组件
- [x] 4.2 实现页面头部显示（别名名称、状态开关、操作按钮区）
- [x] 4.3 实现 mappings 表格，列：选择框、拖拽图标、提供商、API类型、模型名称、Token信息、能力特性、权重、状态、操作
- [x] 4.4 实现拖放排序功能（监听拖拽事件，获取新顺序）
- [x] 4.5 实现拖放后调用 API 更新权重，并刷新表格显示
- [x] 4.6 实现批量删除 mappings 功能（选择、确认、删除）
- [x] 4.7 实现添加/编辑 mapping 弹窗（Provider 下拉、Model 下拉、Weight、Status）
- [x] 4.8 实现删除单个 mapping 功能
- [x] 4.9 实现状态开关直接切换功能（无需确认）
- [x] 4.10 添加路由配置 `web/src/router/index.ts`，新增 `/aliases/:id` 路由
- [ ] 4.11 测试详情页面所有功能（拖放、批量删除、添加、编辑、删除、状态切换）

## 5. Providers Detail 页面增强

- [x] 5.1 修改 `web/src/views/Providers/Detail.vue` Context Window 列，使用 formatContextDisplay
- [x] 5.2 添加 Capabilities 列，显示 Vision/Tools/Stream 标签
- [ ] 5.3 测试页面显示，验证 token 格式化（保留一位小数）和能力标签颜色

## 6. 国际化支持

- [x] 6.1 在 `web/src/locales/en.ts` 中添加新翻译文本（详情页面标题、Token汇总、拖放提示等）
- [x] 6.2 在 `web/src/locales/zh.ts` 中添加对应的中文翻译
- [ ] 6.3 验证多语言切换时所有新增文本正确显示

## 7. 测试和验证

- [x] 7.1 手动测试 Aliases 主页面，验证扁平表格、Token汇总、批量删除功能
- [x] 7.2 手动测试 Aliases 详情页面，验证拖放排序、权重更新、完整信息展示
- [x] 7.3 测试拖放排序的权重计算逻辑（位置递减）
- [x] 7.4 测试 token 格式化的各种边界情况（512、128000、153600、2000000、1500000）
- [x] 7.5 测试汇总统计逻辑（无 enabled mappings、多个 enabled mappings）
- [ ] 7.6 验证数据一致性（ProviderModel 更新后，AliasMapping 显示同步更新）
- [x] 7.7 验证向后兼容性（API 新增字段不影响其他功能）
- [x] 7.8 验证路由正确（`/aliases/:id` 可正常访问）

## 8. 功能完善

- [x] 8.1 上下文格式化内容增加鼠标悬停提示原始值（使用 el-tooltip）
- [x] 8.2 "Token 信息" 改名为 "上下文窗口"
- [x] 8.3 能力特性列移到上下文窗口列前面
- [x] 8.4 模型别名详情页面改进：显示别名、状态显示改用标签、移除编辑和删除按钮
- [x] 8.5 映射列表增加能力显示（取所有 enabled mappings 的能力交集）
- [x] 8.6 修复别名详情页面别名不显示的问题（API 字段名映射）