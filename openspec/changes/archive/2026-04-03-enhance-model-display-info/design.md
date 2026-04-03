## Context

当前系统存在两个主要问题：

1. **信息展示不直观**
   - Providers Detail 页面：token 显示为原始数字（如 "128000 / 4096"），能力特性仅在详情弹窗中可见
   - Aliases 页面：映射列表缺少模型的 token 和能力信息

2. **界面风格不一致**
   - Aliases 页面使用折叠面板布局，每个 Alias 展开后嵌套显示 mappings 表格
   - Providers 页面使用扁平表格布局，点击"详情"进入独立页面
   - 用户需要在不同风格间切换，体验不统一

## Goals / Non-Goals

**Goals:**

- 统一界面风格：Aliases 改为扁平表格布局，与 Providers 保持一致
- 增强信息展示：格式化显示 token（保留一位小数），显式展示能力特性标签
- 提供完整管理：新增 Aliases 详情页面，支持拖放排序、批量操作
- 汇总显示：主列表显示 token 上下文的最小值组合
- 提升交互效率：拖放排序调整权重，批量删除减少操作成本

**Non-Goals:**

- 不修改模型数据的存储结构
- 不改变现有 API 的核心逻辑（仅扩展字段）
- 不增加新的配置或管理功能（仅优化现有功能）
- 不修改其他页面（如 Usage、Dashboard、Keys）

## Decisions

### Decision 1: Token 格式化策略

**选择：使用千(K)/百万(M)简写格式，保留一位小数**

格式化规则：
- `< 1000`: 显示原数字（如 "512"）
- `>= 1000 且 < 1000000`: 
  - 整数 K 值：显示 "XK"（如 "128000" → "128K"）
  - 非整数 K 值：显示 "X.XK"（如 "153600" → "153.6K"）
- `>= 1000000`: 
  - 整数 M 值：显示 "XM"（如 "2000000" → "2M"）
  - 非整数 M 值：显示 "X.XM"（如 "1500000" → "1.5M"）

**备选方案：**
- 原数字 + 单位：如 "128,000 tokens"（占用空间大）
- 仅整数简写：如 "128K"（不够精确）

**理由：** 保留一位小数既保证精确度，又保持简洁性。AI 模型领域常见惯例（如 OpenAI 文档），用户熟悉。

### Decision 2: 能力特性显示方式

**选择：使用标签（Tags）显示能力**

在表格中为每个模型显示能力标签：
- Vision（眼睛图标或标签）
- Tools（工具图标或标签）
- Stream（流式图标或标签）

使用 `el-tag` 组件，默认颜色：
- supports_vision: type="success"（绿色）
- supports_tools: type="warning"（橙色）
- supports_stream: type="primary"（蓝色）

**备选方案：**
- 图标列：单独列显示图标（占用额外宽度）
- 文字描述：如 "支持视觉、工具"（占用空间）
- 集成到现有列：在选择模型下拉框中显示（不利于列表浏览）

**理由：** 标签在 Element Plus 中是标准组件，视觉清晰且不占用过多空间，用户可快速扫描。

### Decision 3: API 数据扩展策略

**选择：在 AliasMapping 返回中嵌套 model_info**

GET `/aliases` API 返回结构扩展：
```json
{
  "mappings": [{
    "id": 1,
    "provider_id": 5,
    "provider_model_name": "gpt-4-turbo",
    "weight": 10,
    "enabled": true,
    "provider": { ... },
    "model_info": {
      "context_window": 128000,
      "max_output": 4096,
      "supports_vision": true,
      "supports_tools": true,
      "supports_stream": true
    }
  }]
}
```

**备选方案：**
- 前端额外请求：在选择模型时查询 `/providers/:id/models/:model_id`（性能开销）
- 扩展 provider 字段：将模型信息放在 provider 中（语义不清晰）

**理由：** 嵌套 model_info 保持语义清晰（映射 -> 模型信息），前端无需额外请求，性能最优。

### Decision 4: 前端实现位置

**选择：直接在表格列中显示**

Providers Detail 模型列表：
- 新增 "Capabilities" 列，显示能力标签
- 修改 "Context Window" 列的显示模板，格式化 token

Aliases 映射列表：
- 新增 "Context" 列，显示 token 信息
- 新增 "Capabilities" 列，显示能力标签

**备选方案：**
- Tooltip 显示：鼠标悬停显示详情（不利于快速浏览）
- 详情弹窗：点击查看详情（交互成本高）

**理由：** 表格列显示是最直接的方式，用户可快速扫描多个模型的信息。

### Decision 5: Aliases 界面布局改版

**选择：采用扁平表格 + 详情页面布局**

布局方案：
- 主页面（`/aliases`）：扁平表格，列：选择框、别名名称、映射数量、Token汇总、状态开关、操作按钮
- 详情页面（`/aliases/:id`）：独立页面，管理 mappings 的完整信息

**备选方案：**
- 保持折叠面板：mappings 嵌套显示（空间受限，信息展示不完整）
- 表格嵌套表格：主表格可展开显示子表格（层级复杂，不符合 Providers 风格）
- 弹窗显示 mappings：点击按钮在弹窗中管理（不利于快速浏览多个 mappings）

**理由：** 
- 与 Providers 界面风格完全一致，用户体验统一
- 详情页面空间充足，可完整展示 token、能力、权重等信息
- 扁平表格便于批量操作，符合用户习惯

### Decision 6: Token 汇总显示策略

**选择：显示有效 mappings 的最小 token 组合，带鼠标提示**

汇总逻辑：
- 只统计 `enabled=true` 的 mappings
- 取所有有效 mappings 中最小的 `context_window`
- 取所有有效 mappings 中最小的 `max_output`
- 显示格式："最小context / 最小output"（如 "8K / 4K"）
- 如果无有效 mapping，显示 "-"
- 鼠标悬停显示原始值（如 "128,000 / 4,096"）

**示例：**
Alias "gpt-4" 有 3 个有效 mappings：
- gpt-4-turbo (128K / 4K)
- gpt-4-32k (32K / 4K)  
- gpt-4 (8K / 4K)

汇总显示：**8K / 4K**

**备选方案：**
- 显示最大值：表示最大可用容量（但可能误导，实际不一定能用）
- 显示平均值：无实际意义
- 显示范围：如 "8K-128K / 4K"（占用空间，难以快速判断）

**理由：** 最小值组合代表该 alias 的保守限制，用户一眼看出最小可用 token 容量，避免配置超出限制的请求。鼠标提示提供详细信息，兼顾简洁性和完整性。

### Decision 6.1: 能力交集显示策略

**选择：显示有效 mappings 的能力交集**

交集逻辑：
- 只统计 `enabled=true` 的 mappings
- 对每项能力（Vision、Tools、Stream）：
  - 如果所有 mapping 都支持，则该项能力为 true
  - 如果有任何 mapping 不支持，则该项能力为 false
- 如果无有效 mapping，所有能力为 false
- 只显示为 true 的能力标签

**示例：**
Alias "gpt-4" 有 3 个有效 mappings：
- gpt-4-turbo (Vision✓, Tools✓, Stream✓)
- gpt-4-32k (Vision✗, Tools✓, Stream✓)
- gpt-4 (Vision✗, Tools✓, Stream✓)

交集显示：**Tools、Stream**（Vision 不显示）

**备选方案：**
- 显示并集：只要有一个 mapping 支持就显示（但可能误导，用户可能选择到不支持的）
- 显示全部：显示所有 mapping 的能力（信息量大，难以快速判断）
- 不显示：在详情页面查看（不利于快速浏览）

**理由：** 交集代表该 alias 的可靠能力保证，用户可以确信使用该 alias 时这些能力一定可用，避免因选择到不支持能力的 mapping 而导致调用失败。

### Decision 7: 拖放排序权重计算

**选择：位置线性递减策略**

权重计算规则：
- 第 1 位（最上）：权重 = mappings 总数 - 1
- 第 2 位：权重 = mappings 总数 - 2
- ...（依次递减）
- 最后一位（最下）：权重 = 0

**示例：**
3 个 mappings，拖放后：
- 第 1 位：weight = 2
- 第 2 位：weight = 1
- 第 3 位：weight = 0

**备选方案：**
- 固定步长：如 100、90、80...（但 mappings 数量不固定时不合理）
- 保持原权重：仅调整顺序，不更新权重（逻辑复杂，用户困惑）
- 手动输入权重：拖放后需用户手动填写（交互成本高）

**理由：** 
- 简单直观：位置越高权重越大
- 自动计算：拖放即生效，无需额外输入
- 线性递减：权重差异明显，便于路由选择

### Decision 8: 拖放实现技术选型

**选择：Element Plus el-table 拖放支持**

实现方案：
- 使用 Element Plus 表格的行拖拽功能
- 监听拖拽事件，获取新顺序
- 调用 API 更新权重

**备选方案：**
- vuedraggable 库：需要额外依赖，增加包体积
- 手动实现拖拽：开发成本高，维护复杂

**理由：** Element Plus 已内置拖放支持，无需额外依赖，与现有 UI 框架一致。

## Risks / Trade-offs

### Risk 1: API 响应体积增加

**风险：** AliasMapping 包含 model_info 后，API 响应体积略微增加（每个映射增加约 5 个字段）。

**缓解：**
- 字段数量有限（仅 5 个），对性能影响微小
- 可考虑后续优化：仅在需要时查询（如前端缓存策略）

### Risk 2: 格式化逻辑的维护

**风险：** Token 格式化逻辑需要在前端多处维护（Providers Detail、Aliases Detail）。

**缓解：**
- 提取为通用 utility 函数：`formatToken(value: number): string`
- 在 `web/src/utils/format.ts` 中统一管理格式化函数

### Risk 3: 数据不一致

**风险：** ProviderModel 数据更新后，AliasMapping 中的 model_info 可能过时。

**缓解：**
- 后端每次查询 AliasMapping 时实时关联 ProviderModel（不缓存）
- 确保数据一致性（无性能问题，关联查询开销小）

### Risk 4: 拖放排序的用户理解

**风险：** 用户可能不理解拖放后权重自动更新的逻辑。

**缓解：**
- 在拖放后显示提示："已更新权重：第1位=权重2，第2位=权重1..."
- 提供帮助文档说明排序规则

### Risk 5: 详情页面路由冲突

**风险：** 新增 `/aliases/:id` 路由可能与现有路由冲突。

**缓解：**
- 检查路由配置，确保无冲突
- 使用明确的路由命名（如 `/aliases/detail/:id` 可选）

## Migration Plan

### Phase 1: 后端 API 扩展

1. 修改 `server/internal/handler/alias.go`：
   - GET `/aliases`：为每个 alias 计算并返回 `min_context_window`、`min_max_output`
   - GET `/aliases/:id`：mappings 包含 model_info 字段
   - 新增 PUT `/aliases/:id/mappings/order` handler，处理拖放排序

2. 测试 API：
   - 验证汇总统计正确（只统计 enabled=true 的 mappings）
   - 验证 model_info 数据一致性
   - 验证拖放排序权重计算正确

### Phase 2: 前端工具函数

1. 在 `web/src/utils/format.ts` 中添加：
   - `formatToken(value: number): string`（保留一位小数）
   - `formatContextDisplay(context: number, output: number): string`

2. 单元测试：
   - 测试格式化函数的各种边界情况
   - 测试汇总逻辑

### Phase 3: Aliases 主页面改版

1. 重构 `web/src/views/Aliases/index.vue`：
   - 改为扁平表格布局
   - 添加选择框列、批量删除功能
   - 添加 Token 汇总列
   - 添加详情按钮（跳转到 `/aliases/:id`）

2. 测试主页面：
   - 验证表格显示正确
   - 验证批量删除功能
   - 验证 Token 汇总显示

### Phase 4: Aliases 详情页面新增

1. 创建 `web/src/views/Aliases/Detail.vue`：
   - 显示 Alias 基本信息
   - mappings 表格支持拖放排序
   - 显示完整信息（Token、能力、权重）
   - 支持批量删除 mappings

2. 实现拖放排序：
   - 监听拖拽事件
   - 调用 PUT `/aliases/:id/mappings/order` API
   - 更新权重显示

3. 添加路由：`web/src/router/index.ts` 新增 `/aliases/:id` 路由

4. 测试详情页面：
   - 验证拖放排序功能
   - 验证批量删除 mappings
   - 验证信息显示完整

### Phase 5: Providers Detail 页面增强

1. 修改 `web/src/views/Providers/Detail.vue`：
   - 修改 Context Window 列，使用 formatContextDisplay
   - 添加 Capabilities 列

2. 测试界面：
   - 验证 token 格式化显示（保留一位小数）
   - 验证能力标签显示

### Phase 6: 国际化支持

1. 在 `web/src/locales/*.ts` 中添加新文本：
   - 详情页面标题、按钮文本
   - Token 汇总列标题
   - 拖放提示文本

2. 测试多语言切换

### Rollback Strategy

如遇问题，可逐步回滚：
- 前端修改可快速回滚（Git revert）
- API 扩展向后兼容（新增字段不影响旧客户端）
- 数据库无变更，无需回滚
- 可先回滚详情页面，保留主页面改版
- 可完全回滚到原有折叠面板布局