# 删除策略重构迁移指南 (PostgreSQL)

## 目标

统一系统删除策略：
- 业务实体（User, Provider, Model, MCP, Key）使用软删除
- 关联表和子资源（ProviderModel, ModelMapping, MCPTool 等）使用硬删除

---

## 一、变更说明

### 变更前的删除策略

| 模型 | 删除策略 | 问题 |
|------|----------|------|
| User | 软删除 | ✓ 正确 |
| Provider | 软删除 | ✓ 正确 |
| ProviderModel | 软删除 | ✗ 应为硬删除 |
| Model | 软删除 | ✓ 正确 |
| ModelMapping | 软删除 | ✗ 应为硬删除 |
| MCP | 软删除 | ✓ 正确 |
| MCPTool | 软删除 | ✗ 应为硬删除 |
| MCPResource | 软删除 | ✗ 应为硬删除 |
| MCPPrompt | 软删除 | ✗ 应为硬删除 |
| Key | 软删除 | ✓ 正确 |

### 变更后的删除策略

**软删除实体（保留 DeletedAt）：**
- User
- Provider
- Model
- MCP
- Key

**硬删除实体（移除 DeletedAt）：**
- ProviderModel
- ModelMapping
- MCPTool
- MCPResource
- MCPPrompt
- KeyModel
- KeyMCPTool
- KeyMCPResource
- KeyMCPPrompt

---

## 二、迁移步骤

### 步骤1: 停止服务

```bash
systemctl stop ai-gateway
```

### 步骤2: 备份数据库

```bash
pg_dump -U <用户名> -d <数据库名> -f backup_delete_strategy_$(date +%Y%m%d_%H%M%S).sql
```

### 步骤3: 检查软删除数据

```sql
-- 检查 ProviderModel 的软删除数据
SELECT COUNT(*) FROM provider_models WHERE deleted_at IS NOT NULL;

-- 检查 ModelMapping 的软删除数据
SELECT COUNT(*) FROM model_mappings WHERE deleted_at IS NOT NULL;

-- 检查 MCPTool 的软删除数据
SELECT COUNT(*) FROM mcp_tools WHERE deleted_at IS NOT NULL;

-- 检查 MCPResource 的软删除数据
SELECT COUNT(*) FROM mcp_resources WHERE deleted_at IS NOT NULL;

-- 检查 MCPPrompt 的软删除数据
SELECT COUNT(*) FROM mcp_prompts WHERE deleted_at IS NOT NULL;
```

> 如果有软删除的数据，需要决定是恢复还是永久删除

### 步骤4: 执行迁移脚本

```sql
BEGIN;

-- 1. 恢复或删除软删除的数据（可选，根据实际情况）
-- 如果需要恢复软删除的数据：
-- UPDATE provider_models SET deleted_at = NULL WHERE deleted_at IS NOT NULL;
-- UPDATE model_mappings SET deleted_at = NULL WHERE deleted_at IS NOT NULL;
-- UPDATE mcp_tools SET deleted_at = NULL WHERE deleted_at IS NOT NULL;
-- UPDATE mcp_resources SET deleted_at = NULL WHERE deleted_at IS NOT NULL;
-- UPDATE mcp_prompts SET deleted_at = NULL WHERE deleted_at IS NOT NULL;

-- 如果需要永久删除软删除的数据：
-- DELETE FROM provider_models WHERE deleted_at IS NOT NULL;
-- DELETE FROM model_mappings WHERE deleted_at IS NOT NULL;
-- DELETE FROM mcp_tools WHERE deleted_at IS NOT NULL;
-- DELETE FROM mcp_resources WHERE deleted_at IS NOT NULL;
-- DELETE FROM mcp_prompts WHERE deleted_at IS NOT NULL;

-- 2. 移除 DeletedAt 列
ALTER TABLE provider_models DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE model_mappings DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE mcp_tools DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE mcp_resources DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE mcp_prompts DROP COLUMN IF EXISTS deleted_at;

-- 3. 移除无效的外键级联约束（如果存在）
ALTER TABLE model_mappings DROP CONSTRAINT IF EXISTS fk_model_mappings_provider_model;
ALTER TABLE model_mappings DROP CONSTRAINT IF EXISTS fk_model_mappings_model;

-- 4. 添加外键约束（无级联）
ALTER TABLE model_mappings
ADD CONSTRAINT fk_model_mappings_provider_model
FOREIGN KEY (provider_model_id) REFERENCES provider_models(id);

ALTER TABLE model_mappings
ADD CONSTRAINT fk_model_mappings_model
FOREIGN KEY (model_id) REFERENCES models(id);

COMMIT;
```

### 步骤5: 部署新代码

```bash
# 部署新版本
systemctl start ai-gateway
```

### 步骤6: 验证功能

---

## 三、验证检查

```sql
-- 验证列已删除
SELECT column_name FROM information_schema.columns 
WHERE table_name = 'provider_models' AND column_name = 'deleted_at';
-- 应返回空

SELECT column_name FROM information_schema.columns 
WHERE table_name = 'model_mappings' AND column_name = 'deleted_at';
-- 应返回空

SELECT column_name FROM information_schema.columns 
WHERE table_name = 'mcp_tools' AND column_name = 'deleted_at';
-- 应返回空

-- 验证外键约束
SELECT constraint_name FROM information_schema.table_constraints 
WHERE table_name = 'model_mappings' AND constraint_type = 'FOREIGN KEY';
```

---

## 四、功能验证清单

- [ ] 删除 Provider 时，ProviderModel 和 ModelMapping 被级联删除
- [ ] 删除 ProviderModel 时，ModelMapping 被级联删除
- [ ] 删除 Model 时，ModelMapping 被级联删除
- [ ] 删除 MCP 时，MCPTool、MCPResource、MCPPrompt 被级联删除
- [ ] 删除 Key 时，KeyModel、KeyMCPTool、KeyMCPResource、KeyMCPPrompt 被级联删除
- [ ] 删除 User、Provider、Model、MCP、Key 后可以在数据库中查到 deleted_at 字段
- [ ] 删除 ProviderModel、ModelMapping 等后数据库中记录被永久删除

---

## 五、回滚方案

如果迁移失败，从备份恢复：

```bash
# 停止服务
systemctl stop ai-gateway

# 恢复数据库
psql -U <用户名> -d <数据库名> -f backup_delete_strategy_*.sql

# 部署旧代码
# ...

# 启动服务
systemctl start ai-gateway
```

---

## 六、注意事项

1. **软删除数据处理**：迁移前必须检查并处理已软删除的数据
2. **外键约束**：新代码使用 `Unscoped()` 硬删除，但外键约束不设置 CASCADE
3. **备份数据**：迁移前务必备份，硬删除无法恢复
4. **测试环境先行**：建议先在测试环境验证迁移脚本

---

## 七、级联删除逻辑说明

| 删除操作 | 级联删除 | 删除类型 |
|----------|----------|----------|
| 删除 Provider | ProviderModel → ModelMapping | 软删除 → 硬删除 → 硬删除 |
| 删除 ProviderModel | ModelMapping | 硬删除 → 硬删除 |
| 删除 Model | ModelMapping | 软删除 → 硬删除 |
| 删除 MCP | MCPTool, MCPResource, MCPPrompt | 软删除 → 硬删除 |
| 删除 Key | KeyModel, KeyMCPTool, KeyMCPResource, KeyMCPPrompt | 软删除 → 硬删除 |
