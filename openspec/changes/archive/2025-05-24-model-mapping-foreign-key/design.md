## Context

当前 `ModelMapping` 使用字符串字段 `ProviderModelName` 匹配 `ProviderModel.ModelID`：

```
ModelMapping.ProviderModelName (string)  ←→  ProviderModel.ModelID (string)
                    字符串匹配，无外键约束
```

问题：
- 修改 `ProviderModel.ModelID` 导致映射"断裂"
- 删除 `ProviderModel` 不会级联删除 `ModelMapping`，产生孤儿数据

## Goals / Non-Goals

**Goals:**
- ModelMapping 通过外键关联 ProviderModel
- ProviderModel.ModelID 可编辑
- 删除 ProviderModel 级联删除相关 ModelMapping
- 数据一致性由数据库约束保证

**Non-Goals:**
- 不修改 ProviderModel 其他字段的编辑逻辑
- 不修改路由算法
- 不修改前端 UI 布局

## Decisions

### Decision 1: 使用外键替代字符串匹配

**选择**: `ModelMapping.ProviderModelName` → `ModelMapping.ProviderModelID`

**理由**:
- 外键约束保证数据一致性
- 级联删除自动清理孤儿数据
- ProviderModel.ModelID 可独立修改，不影响关联

**替代方案**:
- 保留字符串 + 同步维护：需要手动维护多处逻辑，易遗漏
- 不做改动：问题持续存在

### Decision 2: 保留 ProviderID 字段

**选择**: 保留 `ModelMapping.ProviderID`（冗余但有用）

**理由**:
- 查询方便，无需 JOIN
- 已有索引，性能无影响
- 可快速过滤同一 Provider 的映射

### Decision 3: 删除旧字段

**选择**: 直接删除 `ProviderModelName` 字段

**理由**:
- 数据量小（几十条），迁移简单
- 避免 GORM AutoMigrate 混淆
- 新旧代码不会同时维护

### Decision 4: API 兼容性

**选择**: 请求体使用 `provider_model_id`，响应体保留 `provider_model_name`

**理由**:
- 响应兼容：前端无需立即修改显示逻辑
- 请求变化：前端需要适配，但改动明确

### Decision 5: 不添加唯一约束

**选择**: 不对 `(provider_id, provider_model_id)` 添加唯一约束

**理由**:
- 业务需求：同一个 ProviderModel 可以被多次映射到同一个 Model
- 例如：不同权重的多个映射，用于负载均衡

**替代方案**:
- 添加唯一约束：会限制业务灵活性

## Risks / Trade-offs

### Risk: 数据迁移失败导致服务不可用
→ **Mitigation**: 迁移前备份数据库，迁移脚本包含回滚步骤

### Risk: 存在孤儿数据（ProviderModelName 匹配不到 ProviderModel）
→ **Mitigation**: 迁移脚本检查孤儿数据，需手动处理后再继续

### Risk: 前端未及时适配新 API
→ **Mitigation**: 先部署后端，确保服务启动正常后再更新前端

## Migration Plan

详见 `migration-guide.md`

步骤概要：
1. 停止服务
2. 备份数据库
3. 执行数据库迁移脚本
4. 部署新代码
5. 启动服务，验证功能
6. 前端适配新 API
