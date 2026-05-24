# ModelMapping 外键迁移指南 (PostgreSQL)

## 目标

将 `ModelMapping.ProviderModelName` (string) 迁移为 `ProviderModelID` (uint, 外键)，实现：
1. ProviderModel.ModelID 可编辑
2. 删除 ProviderModel 时级联删除相关 Mapping
3. 数据一致性由数据库约束保证

---

## 一、变更说明

### 变更前

```
ModelMapping.ProviderModelName (string)  ←→  ProviderModel.ModelID (string)
                    字符串匹配，无外键约束
```

**问题：**
- 修改 ProviderModel.ModelID 导致映射"断裂"
- 删除 ProviderModel 不会级联删除 ModelMapping，产生孤儿数据

### 变更后

```
ModelMapping.ProviderModelID (uint)  →  ProviderModel.ID (uint)
                    外键关联，级联删除
```

**改进：**
- ProviderModel.ModelID 可随意修改
- 删除 ProviderModel 自动级联删除相关 Mapping
- 数据库约束保证数据一致性

---

## 二、影响范围

### 后端代码

| 文件 | 改动 |
|------|------|
| `internal/model/db.go` | ModelMapping 结构体：ProviderModelName → ProviderModelID (外键) |
| `internal/handler/model.go` | API 请求/响应、Preload ProviderModel |
| `internal/handler/provider_model.go` | Update 方法允许修改 ModelID |
| `internal/router/router.go` | 使用 ProviderModelID 查询 |
| `internal/handler/model_testing.go` | Preload ProviderModel |

### 前端代码

| 文件 | 改动 |
|------|------|
| `web/src/views/Models/Detail.vue` | 表单字段：provider_model_name → provider_model_id |
| `web/src/views/Providers/Detail.vue` | 移除 model_id 编辑禁用 |

### API 变化

**创建映射 POST /models/:id/mappings**

旧请求:
```json
{
  "provider_id": 1,
  "provider_model_name": "gpt-4"
}
```

新请求:
```json
{
  "provider_id": 1,
  "provider_model_id": 100
}
```

**更新映射 PUT /models/:id/mappings/:mid**

旧请求:
```json
{
  "provider_model_name": "gpt-4"
}
```

新请求:
```json
{
  "provider_model_id": 100
}
```

**响应体（保持兼容）**
```json
{
  "id": 1,
  "provider_id": 1,
  "provider_model_id": 100,
  "provider_model_name": "gpt-4",
  "weight": 1,
  "enabled": true
}
```

> `provider_model_name` 从关联的 ProviderModel.ModelID 获取，保持返回

---

## 三、迁移步骤

### 步骤1: 停止服务

```bash
systemctl stop ai-gateway
```

### 步骤2: 备份数据库

```bash
pg_dump -U <用户名> -d <数据库名> -t model_mappings -t provider_models -f backup_$(date +%Y%m%d_%H%M%S).sql
```

### 步骤3: 执行迁移脚本

```sql
BEGIN;

-- 1. 添加新字段
ALTER TABLE model_mappings ADD COLUMN provider_model_id BIGINT;

-- 2. 迁移数据
UPDATE model_mappings mm
SET provider_model_id = pm.id
FROM provider_models pm
WHERE pm.provider_id = mm.provider_id
  AND pm.model_id = mm.provider_model_name;

-- 3. 检查孤儿数据（如有需要，手动处理后继续）
-- SELECT * FROM model_mappings WHERE provider_model_id IS NULL;

-- 4. 添加 NOT NULL 约束
ALTER TABLE model_mappings ALTER COLUMN provider_model_id SET NOT NULL;

-- 5. 添加外键约束
ALTER TABLE model_mappings
ADD CONSTRAINT fk_model_mappings_provider_model
FOREIGN KEY (provider_model_id) REFERENCES provider_models(id)
ON DELETE CASCADE;

-- 6. 清理旧索引
DROP INDEX IF EXISTS idx_provider_model;
DROP INDEX IF EXISTS idx_model_mappings_provider_model;

-- 7. 添加新索引（非唯一，允许同一模型多次映射）
CREATE INDEX IF NOT EXISTS idx_model_mappings_provider_id ON model_mappings(provider_id);
CREATE INDEX IF NOT EXISTS idx_model_mappings_provider_model_id ON model_mappings(provider_model_id);

-- 8. 删除旧字段
ALTER TABLE model_mappings DROP COLUMN provider_model_name;

-- 9. 修复主键序列（重要！否则插入会报主键冲突）
SELECT setval('model_mappings_id_seq', COALESCE((SELECT MAX(id) FROM model_mappings), 0) + 1, false);

COMMIT;
```

### 步骤4: 部署新代码

```bash
# 构建并部署
make
# 或手动部署编译产物
```

### 步骤5: 启动服务

```bash
systemctl start ai-gateway
```

### 步骤6: 验证功能

- [ ] 模型列表接口正常
- [ ] 创建映射使用 provider_model_id
- [ ] 更新映射使用 provider_model_id
- [ ] 删除 ProviderModel 级联删除 Mapping
- [ ] 修改 ProviderModel.ModelID 后路由正常
- [ ] 响应体中 provider_model_name 正确返回

---

## 四、验证 SQL

```sql
-- 检查表结构
\d model_mappings

-- 检查数据
SELECT id, model_id, provider_id, provider_model_id FROM model_mappings LIMIT 10;

-- 检查外键关联
SELECT mm.id, mm.provider_model_id, pm.model_id 
FROM model_mappings mm 
LEFT JOIN provider_models pm ON mm.provider_model_id = pm.id 
LIMIT 10;

-- 检查索引
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'model_mappings';
```

---

## 五、回滚方案

如果迁移失败，从备份恢复：

```bash
# 停止服务
systemctl stop ai-gateway

# 恢复数据库
psql -U <用户名> -d <数据库名> -f backup_*.sql

# 部署旧代码
# ...

# 启动服务
systemctl start ai-gateway
```

---

## 六、注意事项

1. **执行前务必备份**
2. **孤儿数据处理**：迁移后检查是否有 NULL 值
3. **唯一索引已移除**：允许同一 ProviderModel 被多次映射到同一个 Model（业务需求）
4. **主键序列**：迁移后必须修复序列，否则插入新记录会报主键冲突
5. **前端适配**：API 请求体变化，需要同步更新前端代码
