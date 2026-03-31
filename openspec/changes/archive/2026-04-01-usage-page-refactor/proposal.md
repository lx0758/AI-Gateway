## Why

当前 Usage 页面调用 3 个独立后端接口（`/usage/stats`、`/usage/key-stats`、`/usage/logs`）各自进行数据库聚合，存在重复计算和代码冗余。日志接口限制返回 1000 条，导致大数据量时统计不准确。统计维度单一，无法满足灵活分析需求。

## What Changes

### 接口重构
- **移除** `/usage/stats` 接口（后端聚合）
- **移除** `/usage/key-stats` 接口（后端聚合）
- **重构** `/usage/logs` 接口：
  - 移除 1000 条 LIMIT 限制
  - 按前端传来的时间范围返回全量原始日志
  - 返回字段完整化（确保前端聚合所需字段都有）

### 前端聚合
- Usage 页面加载时调用单一 `/usage/logs` 接口
- 前端 JavaScript 计算所有统计指标：
  - 总请求数、成功率、总 Tokens、平均耗时
  - Key 统计（按 key_name 聚合）
  - 模型统计（按 model 聚合）
  - **新增** 接入点统计（按 source 聚合）
  - **新增** 厂商统计（按 provider_name 聚合）
  - **新增** 厂商模型统计（按 provider_name + model 聚合）

### 模型统计简化
- 原模型统计列：model | actual_model | provider_name | count | tokens | avg_latency
- 简化后列：model | count | tokens | avg_latency

### 卡片顺序重排
调整 Usage 页面卡片展示顺序：
1. 接入点统计（新增）
2. Key 统计
3. 模型统计
4. 厂商统计（新增）
5. 厂商模型统计（新增）
6. 日志

## Capabilities

### New Capabilities
- `usage-analytics`: 增强的用量分析能力，支持多维度聚合统计（接入点、厂商、厂商模型），由前端进行数据聚合

### Modified Capabilities
- `usage-tracking`: 现有的 `/usage/stats` 和 `/usage/key-stats` 接口将废弃，改为单一 `/usage/logs` 接口由前端聚合。需求场景不变，实现方式改变

## Impact

### 后端
- 删除 `/usage/stats` handler 方法
- 删除 `/usage/key-stats` handler 方法
- 修改 `/usage/logs` handler：移除 LIMIT 1000

### 前端
- 重构 `Usage/index.vue`：fetchStats/fetchKeyStats 合并为单一 fetchLogs + 前端聚合
- 新增统计计算函数：聚合接入点、厂商、厂商模型数据
- 调整卡片顺序

### 数据库
- 无变更

### API
- `GET /usage/stats` - **废弃**
- `GET /usage/key-stats` - **废弃**
- `GET /usage/logs` - **修改** 移除 LIMIT，返回全量数据
