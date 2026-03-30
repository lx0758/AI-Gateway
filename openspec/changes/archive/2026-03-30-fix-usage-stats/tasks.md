# 任务列表

## 验证

- [x] **T0: stream_options.include_usage 验证**
  - [x] OpenRouter: ✅ 支持，返回完整 usage 信息
  - [x] 响应格式确认: 最后一个 data chunk 包含 usage 字段

## 后端

- [x] **T1: Transformer 接口重构**
  - [x] 定义 `StreamResult` 结构体
  - [x] 修改 `Transformer.TransformStream` 签名
  - [x] 实现 `PassThroughTransformer.TransformStream` 解析 usage
  - [x] 实现 `AnthropicTransformer.TransformStream` 解析 usage

- [x] **T2: 请求注入 stream_options**
  - [x] 定义 `StreamOptions` 结构体
  - [x] 在 `ProxyHandler.ChatCompletions` 中自动注入

- [x] **T3: 数据模型扩展**
  - [x] `UsageLog` 增加 `actual_model` 字段
  - [x] `APIKey` 增加 `used_count` 字段
  - [x] GORM 自动迁移

- [x] **T4: ProxyHandler 修改**
  - [x] 保存原始模型别名
  - [x] 修改 `handleStreamResponse` 使用新签名
  - [x] 修改 `logUsage` 签名和实现
  - [x] 增加调用次数统计

- [x] **T5: API 接口扩展**
  - [x] `GET /usage/stats` 增加模型维度统计
  - [x] `GET /usage/logs` 返回 `actual_model`
  - [x] `GET /api-keys` 返回 `used_count`

## 前端

- [x] **T6: Usage 页面增强**
  - [x] 日志表格显示 `actual_model` 列
  - [x] 增加模型维度统计视图
  - [ ] 支持按 API Key 筛选 (未实现)

- [x] **T7: API Keys 页面增强**
  - [x] 显示 `used_count` (调用次数)
  - [x] 优化 `used_quota` 显示格式

## 测试

- [x] **T8: 手动验证**
  - [x] 流式请求完整流程测试
  - [x] 统计数据验证

- [ ] **T9: 单元测试** (未实现)
  - [ ] Transformer 流式解析测试
  - [ ] 数据模型测试

## 任务依赖

```
T0 ✅
 │
 └──▶ T1 ──┬──▶ T4 ──▶ T5
           │
      T2 ──┘
           
      T3 ──────▶ T4

      T5 ──────▶ T6
      T5 ──────▶ T7
```

## 执行顺序

0. **Phase 0**: T0 验证 ✅ 已完成
1. **Phase 1**: T1 + T2 + T3 ✅ 已完成
2. **Phase 2**: T4 ✅ 已完成
3. **Phase 3**: T5 ✅ 已完成
4. **Phase 4**: T6 + T7 ✅ 已完成
5. **Phase 5**: T8 ✅ 手动验证完成, T9 单元测试待补充
