## Context

**项目定位：** AI Proxy 是统一的 AI 服务代理平台，当前版本聚焦于 AI 模型 API 代理（OpenAI 兼容格式），未来将扩展支持 MCP/ACP 等协议代理。

当前系统在编辑功能和数据管理上存在不足：
- 前端编辑厂商时不加载现有数据
- 模型映射不支持修改关键字段
- API Key 的 `allowed_models` 使用 JSON 存储，查询和管理不便
- 模型同步会覆盖手动配置
- 仪表盘 API 返回 snake_case 字段名与前端 camelCase 不匹配
- 刷新页面后跳转登录页，路由守卫在 fetchUser 执行前就检查了登录状态

**约束：**
- 保持现有 API 路径兼容
- 数据库迁移需要平滑过渡
- 现有数据（如有）需要迁移

## Goals / Non-Goals

**Goals:**
- 修复厂商编辑的前端实现缺陷
- 修复仪表盘字段命名不一致
- 修复刷新页面跳转登录问题
- 修复 API 类型显示格式，使用 AI SDK 标准命名
- 支持模型映射的完整编辑功能
- 使用关联表管理 API Key 模型权限
- 保护手动创建的模型不被同步覆盖

**Non-Goals:**
- 不修改数据库中其他表结构
- 不新增新的厂商类型
- 不改变现有认证授权机制

## Decisions

### D1: API Key 模型权限存储方式

**决定：** 使用关联表替代 JSON 存储

```
当前：
keys.allowed_models = '["gpt-4", "claude-3"]'  -- JSON 字符串

改进后：
api_key_models 表
├── id (PK)
├── key_id (FK → keys.id)
├── model_alias (关联到 model_mappings.alias)
└── created_at
```

**理由：**
- 查询时直接 JOIN，无需解析 JSON
- 可按模型反查哪些 Key 有权限
- 删除/添加权限是单条记录操作
- 数据库层面保证引用完整性

**备选方案：**
- 保留 JSON + GORM JSON 类型：查询仍需应用层处理，不够灵活

### D2: 手动模型同步保护策略

**决定：** 同步时跳过 `source=manual` 的模型

```go
// 同步逻辑
if pm.Source == "manual" {
    continue  // 跳过手动创建的模型
}
```

**理由：**
- 简单直接，不影响现有同步流程
- 保留手动配置的完整性
- 用户可以通过删除重建来强制更新

**备选方案：**
- 合并策略（API 值覆盖部分字段）：复杂度高，容易出错

### D3: 模型映射编辑策略

**决定：** 后端扩展 Update API，前端新增编辑对话框

**理由：**
- 与其他实体的编辑模式一致
- 用户可能配置错误需要修正

### D4: 刷新页面保持登录状态

**决定：** 路由守卫中先尝试恢复用户状态再判断登录

```
当前流程：
1. Pinia 初始化 → user = null
2. 路由守卫 → isLoggedIn = false → 跳转登录
3. App.vue onMounted → fetchUser() (未执行)

改进后流程：
1. Pinia 初始化 → user = null
2. 路由守卫 → 调用 fetchUser() 等待结果
   - 成功：继续访问
   - 失败：跳转登录
```

**理由：**
- 后端 session 仍然有效，只是前端 store 未初始化
- 路由守卫是判断登录的正确位置
- App.vue 的 fetchUser 变为冗余，可移除

**备选方案：**
- 使用 localStorage 持久化：增加复杂度，需要同步状态

### D5: API 类型显示格式

**决定：** 前端显示使用 AI SDK 标准命名格式

| 内部值 | 显示标签 |
|--------|----------|
| `openai` | `@ai-sdk/openai-compatible` |
| `anthropic` | `@ai-sdk/anthropic` |

**理由：**
- 与 Vercel AI SDK 生态保持一致，方便开发者理解
- 明确表达代理的是 SDK 调用而非直接 API 调用
- 后端 `type` 字段值保持不变，仅前端显示转换

## Risks / Trade-offs

### R1: 数据迁移

**风险：** 现有 `allowed_models` JSON 数据需要迁移到关联表

**缓解：** 提供迁移脚本，启动时自动检测并迁移

### R2: 引用完整性

**风险：** `api_key_models.model_alias` 可能引用不存在的 alias

**缓解：** 
- 创建时验证 alias 存在于 `model_mappings`
- 删除 mapping 时检查是否有 Key 引用

## Open Questions

- [ ] 数据迁移时机：启动时自动迁移 vs 手动迁移命令？
- [ ] 空权限列表的含义：拒绝所有模型 vs 允许所有模型？
