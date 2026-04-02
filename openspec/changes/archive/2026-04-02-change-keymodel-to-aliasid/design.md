## Context

当前 KeyModel 表存储模型别名名称（字符串），例如 "gpt-4"。当别名被重命名时（如 "gpt-4" → "gpt-4-turbo"），KeyModel 表中的字符串不会自动更新，导致：
1. API Key 的模型绑定关系显示旧名称
2. 数据不一致，需要手动维护

系统使用 Gorm ORM，支持外键关联和预加载查询。Alias 表有 ID 和 Name 字段。

## Goals / Non-Goals

**Goals:**
- KeyModel 表改为存储 AliasID，建立外键关联
- 别名重命名时，API Key 自动显示新名称，保持数据一致性
- 前后端统一使用 AliasID 传递和存储
- 查询时预加载 Alias 信息，返回 AliasID 和 Alias 名称

**Non-Goals:**
- 不处理数据迁移（用户手动处理）
- 不支持过渡期兼容（前后端需要同步更新）
- 不保留被删除 AddModel/RemoveModel API

## Decisions

### 1. 数据库字段设计
**决策**: KeyModel 表从 `Model string` 改为 `AliasID uint`，添加外键约束 `OnDelete:CASCADE`。

**理由**:
- 外键约束保证数据完整性，别名删除时自动清理关联记录
- AliasID 索引优化查询性能
- Gorm association 支持预加载查询

**备选方案**: 不添加外键约束，只存储 AliasID
- 缺点：别名删除后产生孤儿记录，数据完整性无法保证

### 2. API 请求参数设计
**决策**: 创建/更新 API Key 的 `models` 参数从 `string[]`（别名名称）改为 `uint[]`（AliasID）。

**理由**:
- 前后端统一使用 ID，减少转换逻辑
- 避免名称变化导致的歧义
- ID 更稳定，不受重命名影响

**备选方案**: 后端接收名称，自动转换为 AliasID
- 缺点：需要额外查询转换，前端仍需维护名称到 ID 的映射

### 3. 响应格式设计
**决策**: KeyModel 响应包含 `alias_id` 和 `alias_name`，同时返回 ID 和名称。

**理由**:
- 前端需要 ID 用于选择器 value，需要名称用于显示 label
- 查询时预加载 Alias，一次查询获取完整信息
- 响应格式自描述，前端无需额外查询

**备选方案**: 响应只返回 AliasID，前端查询 Alias 表获取名称
- 缺点：前端需要额外查询，性能差，用户体验不佳

### 4. 删除单独操作 API
**决策**: 删除 `POST /api-keys/:id/models` 和 `DELETE /api-keys/:id/models/:model_alias` 两个端点。

**理由**:
- 前端未使用这两个 API，只用 Update API Key 批量更新
- 减少代码维护负担
- 简化 API 设计，只保留批量更新方式

**备选方案**: 保留并修改为接收 AliasID
- 缺点：维护多余代码，增加复杂度

## Risks / Trade-offs

**风险**: AliasID 不存在时创建 KeyModel 会失败
- **缓解**: 创建/更新时验证 AliasID 存在，不存在返回 400 错误

**风险**: 别名删除导致 API Key 模型绑定丢失
- **缓解**: 这是预期行为（CASCADE 删除），别名删除前应提示用户

**风险**: 前后端需要同步更新，无法兼容旧版本
- **缓解**: 用户已确认可以手动处理，不考虑兼容性

**权衡**: 不提供数据迁移脚本
- **理由**: 用户选择手动处理，简化实现

## Migration Plan

**手动迁移步骤**：
1. 停止服务
2. 执行数据库 schema 变更（删除 Model 字段，添加 AliasID 字段）
3. 手动处理现有 KeyModel 数据（将名称转换为 AliasID）
4. 更新代码，重启服务

**回滚策略**：
- 保留旧代码版本，数据库 schema 回退
- 重新导入旧数据