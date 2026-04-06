## Context

当前系统使用 "alias"（别名）术语来表示用户可调用的模型名称，这与业界通用术语不一致。"model"（模型）一词更直观，用户更容易理解。

重命名涉及三个核心结构：
- `Alias` → `Model`：用户可调用的模型名称
- `AliasMapping` → `ModelMapping`：模型到提供商模型的映射
- `UsageLog` → `ModelLog`：模型调用日志

间接依赖（引用了上述结构的其他表/结构）：
- `KeyModel.alias_id` → `KeyModel.model_id`：API Key 与模型的关联

需清理的冗余代码：
- `Mapping` PO：未使用的冗余结构体，需要删除

当前约束：
- 系统使用 GORM ORM，支持表名自定义
- 前端使用 Vue 3 + Vue Router
- 已有 API 需要保持兼容或提供迁移路径

## Goals / Non-Goals

**Goals:**
- 重命名所有相关代码结构，提高代码可读性
- 保持数据完整性，不丢失任何现有数据
- 更新所有文档和规范

**Non-Goals:**
- 不改变任何业务逻辑
- 不添加新功能
- 不优化性能

## Decisions

### 1. 数据库迁移策略

**决策**: 使用 SQL 迁移脚本重命名表，而非删除重建

**原因**: 
- 保留现有数据
- 外键关系自动保留
- 回滚更简单

**替代方案**:
- ❌ 删除旧表创建新表 - 数据丢失风险
- ❌ 在应用层做兼容映射 - 增加复杂度

### 2. API 路径兼容性

**决策**: 保留旧 API 路径作为别名，添加新路径作为主路径

**原因**:
- 避免破坏现有客户端
- 提供平滑迁移期

**替代方案**:
- ❌ 直接替换所有路径 - 破坏性变更
- ❌ 只保留旧路径 - 命名不一致

### 3. 前端路由策略

**决策**: 完全重命名路由路径，不保留旧路径

**原因**:
- 前端是内部管理界面，不存在外部客户端
- 用户通过导航菜单访问，路径变更影响小
- 保持代码一致性

## Risks / Trade-offs

**风险 1: 数据迁移失败**
→ 缓解: 迁移前备份数据库，提供回滚脚本

**风险 2: 遗漏代码引用**
→ 缓解: 使用全局搜索确保所有引用已更新，运行完整测试

**风险 3: 第三方工具依赖旧表名**
→ 缓解: 检查是否有外部依赖，文档中说明变更

## Migration Plan

### Phase 1: 数据库迁移

```sql
-- 备份数据库（手动执行）
-- 重命名表
ALTER TABLE aliases RENAME TO models;
ALTER TABLE alias_mappings RENAME TO model_mappings;
ALTER TABLE usage_logs RENAME TO model_logs;

-- 更新外键列名（保持一致性）
ALTER TABLE model_mappings RENAME COLUMN alias_id TO model_id;
ALTER TABLE key_models RENAME COLUMN alias_id TO model_id;  -- 间接依赖
```

### Phase 2: 后端代码更新

1. 更新 PO 结构体定义
2. 更新所有 handler 函数
3. 更新路由定义
4. 更新所有引用

### Phase 3: 前端代码更新

1. 重命名视图目录和文件
2. 更新路由配置
3. 更新组件代码
4. 更新国际化文件

### Phase 4: 文档更新

1. 重命名 spec 目录
2. 更新所有文档引用

### Rollback Plan

```sql
-- 回滚数据库变更
ALTER TABLE models RENAME TO aliases;
ALTER TABLE model_mappings RENAME COLUMN model_id TO alias_id;
ALTER TABLE model_mappings RENAME TO alias_mappings;
ALTER TABLE key_models RENAME COLUMN model_id TO alias_id;
ALTER TABLE model_logs RENAME TO usage_logs;
```

代码回滚: `git revert` 相关提交

## Open Questions

无需确认，所有决策已明确。
