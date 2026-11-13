# Spec Delta

## REMOVED Requirements

### Requirement: 模型能力多选管理
**Reason**: 能力特性（capabilities）在控制台各列表、编辑表单与详情展示中长期无实际用途，且与 MCP 协议能力、模型可用性等概念混淆；本次变更将其从界面与 API 响应中全局移除，仅保留数据库字段用于数据积累。
**Migration**: `provider_models.capabilities` 数据库字段保留且同步写入逻辑不变，已有数据不丢失；前端不再展示、编辑该字段，API 响应不再返回该字段。管理员无需操作；如未来需要恢复展示，可基于保留的字段重新引入界面。
