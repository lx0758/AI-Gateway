## Context

AI-Proxy 已有完整的 Provider 模型管理 API，但前端 UI 缺少手动管理功能。管理员无法通过界面为不支持自动同步的厂商添加模型。

**核心问题**: ModelMapping 使用 `provider_model_id`（数据库主键）关联 ProviderModel。当厂商改名模型时，同步会创建新记录，ModelMapping 仍指向旧记录，导致路由失效。

### 现有关联方式的问题

```
同步前:
ProviderModel: ID=1, model_id="gpt-4"
ModelMapping:  provider_model_id=1

厂商改名 gpt-4 → gpt-4-turbo

同步后:
ProviderModel: ID=1, model_id="gpt-4" (孤立，不再更新)
ProviderModel: ID=2, model_id="gpt-4-turbo" (新建)
ModelMapping:  provider_model_id=1 (仍指向旧记录!)
```

### 解决方案

改用 `provider_model_name` 存储 model_id 字符串，直接关联厂商的模型标识。

```
新设计:
ModelMapping: provider_model_name="gpt-4-turbo"

同步后直接按 (provider_id, model_id) 查找，无主键依赖。
```

### 现有 API

| 端点 | 功能 | 状态 |
|------|------|------|
| `GET /api/v1/providers/:id/models` | 列出模型 | ✅ 已实现 |
| `POST /api/v1/providers/:id/models` | 创建模型 | ✅ 已实现 |
| `PUT /api/v1/providers/:id/models/:mid` | 更新模型 | ✅ 已实现 |
| `DELETE /api/v1/providers/:id/models/:mid` | 删除模型 | ⚠️ 缺少保护 |

### 数据模型

`ProviderModel` 已有 `source` 字段，值为 `sync` 或 `manual`。

## Goals / Non-Goals

**Goals:**
- **BREAKING**: ModelMapping 改用 model_id 字符串关联
- Provider 详情页增加手动添加/编辑/删除模型的 UI
- 列表显示模型来源标签
- 删除时检查 source，拒绝删除同步模型

**Non-Goals:**
- 不修改模型同步逻辑
- 不实现自动迁移 ModelMapping（改关联方式后不需要）

## Decisions

### 1. 改变关联方式 [BREAKING]

**决定**: ModelMapping 改用 `provider_model_name` 字段存储 model_id，删除 `provider_model_id` 字段。

**理由**:
- 模型 ID 变更时，ModelMapping 自动失效（查询不到对应模型）
- 不需要维护外键关系
- 简化同步逻辑，无需处理迁移

**替代方案**:
- 自动迁移 ModelMapping：需要判断新旧模型关系，复杂且易出错
- 仅警告：需要管理员手动处理，体验差

### 2. 数据迁移

**决定**: 创建迁移脚本，将现有 `provider_model_id` 转换为 `provider_model_name`。

```sql
-- 新增字段
ALTER TABLE model_mappings ADD COLUMN provider_model_name VARCHAR(128);

-- 迁移数据
UPDATE model_mappings m 
SET provider_model_name = (SELECT model_id FROM provider_models WHERE id = m.provider_model_id);

-- 删除旧字段
ALTER TABLE model_mappings DROP COLUMN provider_model_id;
```

### 3. UI 设计

**决定**: 在 Provider 详情页的模型表格上方添加"添加模型"按钮，点击弹出表单对话框。

**理由**:
- 与现有的 Models 页面（管理 ModelMapping）保持一致的交互模式
- 模型属于特定 Provider，在 Provider 详情页管理更合理

### 4. 编辑/删除权限

**决定**: 只有 `source=manual` 的模型可编辑和删除。

**理由**:
- 同步模型会被自动更新覆盖，手动编辑无意义
- 已有 `manual-model-protection` spec 定义此行为

## Risks / Trade-offs

### 1. Breaking Change

**风险**: 修改 ModelMapping 关联方式是 Breaking Change，可能影响现有部署。

**缓解**: 
- 提供数据迁移脚本
- 在 Release Notes 中明确标注

### 2. 查询性能

**风险**: 用字符串关联可能影响查询性能（相比整数外键）。

**缓解**: 
- 在 provider_model_name 上创建索引
- model_id 通常较短（<128字符），性能影响有限

### 3. 同步模型无法编辑

**风险**: 用户可能想修改同步模型的配置（如价格）。

**缓解**: 后续可考虑允许编辑同步模型的部分字段（如价格），但保持技术信息从同步源更新。
