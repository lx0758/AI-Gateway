## Context

当前系统删除策略混乱：

```
软删除实体（有 DeletedAt）
├── User
├── Provider
├── ProviderModel ← 问题：关联表，应为硬删除
├── Model
├── ModelMapping ← 问题：关联表，应为硬删除
├── MCP
├── MCPTool ← 问题：关联表，应为硬删除
├── MCPResource ← 问题：关联表，应为硬删除
├── MCPPrompt ← 问题：关联表，应为硬删除
└── Key

硬删除实体（无 DeletedAt）
├── KeyModel
├── KeyMCPTool
├── KeyMCPResource
└── KeyMCPPrompt
```

**问题：**
1. 关联表（如 ModelMapping）使用软删除不合理，会导致数据孤岛
2. 软删除的实体上有 `OnDelete:CASCADE` 约束，但这些约束在软删除时不会触发
3. 删除策略不一致，难以维护

## Goals / Non-Goals

**Goals:**
- 统一删除策略，区分业务实体和关联表
- 业务实体（User, Provider, Model, MCP, Key）使用软删除
- 关联表和子资源使用硬删除
- 在代码层面正确实现级联删除

**Non-Goals:**
- 不改变外部 API 行为
- 不修改现有的软删除实体（User, Provider, Model, MCP, Key）

## Decisions

### Decision 1: 业务实体软删除

**选择**: User, Provider, Model, MCP, Key 保持软删除

**理由**:
- 这些是用户主动创建的业务实体
- 可能需要恢复或审计历史记录
- 软删除可以保护数据安全

### Decision 2: 关联表和子资源硬删除

**选择**: ProviderModel, ModelMapping, MCPTool, MCPResource, MCPPrompt 使用硬删除

**理由**:
- 这些是配置数据或关联关系，不需要恢复
- 硬删除简化数据管理，避免数据孤岛
- 级联删除逻辑更清晰

**替代方案**:
- 全部软删除：数据冗余，查询复杂
- 全部硬删除：业务实体无法恢复

### Decision 3: 代码层面实现级联删除

**选择**: 移除数据库层面的 `OnDelete:CASCADE`，在 Handler 中手动处理

**理由**:
- 软删除不触发数据库级联
- 代码层面可以精确控制删除顺序
- 便于添加业务逻辑（如日志记录）

**替代方案**:
- 数据库触发器：复杂度高，难以调试

### Decision 4: 删除顺序

**选择**: 先删除子记录，再删除父记录

```
Provider → ProviderModel → ModelMapping
Model → ModelMapping
MCP → MCPTool, MCPResource, MCPPrompt
Key → KeyModel, KeyMCPTool, KeyMCPResource, KeyMCPPrompt
```

**理由**:
- 避免外键约束错误
- 保证数据一致性

## Risks / Trade-offs

### Risk: 已部署环境数据丢失
→ **Mitigation**: 迁移前备份数据库

### Risk: 硬删除无法恢复
→ **Mitigation**: 关联表数据可以从上游重新同步（ProviderModel）或重新配置（ModelMapping）

### Risk: 迁移脚本执行失败
→ **Mitigation**: 在事务中执行，失败自动回滚

## Migration Plan

### 步骤1: 备份数据库

```bash
pg_dump -U <用户名> -d <数据库名> -f backup_$(date +%Y%m%d_%H%M%S).sql
```

### 步骤2: 执行迁移脚本

```sql
BEGIN;

-- 移除 ProviderModel 的 DeletedAt 列
ALTER TABLE provider_models DROP COLUMN IF EXISTS deleted_at;

-- 移除 ModelMapping 的 DeletedAt 列
ALTER TABLE model_mappings DROP COLUMN IF EXISTS deleted_at;

-- 移除 MCP 相关表的 DeletedAt 列
ALTER TABLE mcp_tools DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE mcp_resources DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE mcp_prompts DROP COLUMN IF EXISTS deleted_at;

COMMIT;
```

### 步骤3: 部署新代码

### 步骤4: 验证功能

- 验证删除 Provider 时级联删除正确
- 验证删除 ProviderModel 时级联删除正确
- 验证删除 Model 时级联删除正确
- 验证删除 MCP 时级联删除正确
- 验证删除 Key 时级联删除正确
