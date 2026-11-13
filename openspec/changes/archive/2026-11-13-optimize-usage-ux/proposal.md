# Proposal

## Why

控制台在日常使用中积累了一批体验问题：列表排序不合直觉、API 密钥的资源权限用"隐式白名单"表达导致语义晦涩（最后一项拨回默认会静默变成全放行）、能力特性字段在三层界面上展示却无人使用、映射对话框切换厂商后残留旧模型导致保存被 `provider mismatch` 拒绝、厂商/密钥编辑对话框缺少与其余列表不一致的状态控制。本次变更集中修复这五项，提升控制台的一致性与可预测性。

## What Changes

1. **列表排序优化**
   - 模型厂商页默认按"状态 + 名称"排序（仅在用户无已保存排序偏好时生效，偏好恢复行为不变）
   - 所有列表的状态列在激活排序时隐式以"名称"作为第二排序因子（同向跟随）；状态升序时启用项在前
   - Models/Detail 映射列表的状态列补上可排序能力（当前不可排序，与规格不符）
2. **API 密钥资源权限改为显式白名单**（**BREAKING**）
   - `keys` 表新增 4 个独立布尔开关：`models_whitelist`、`tools_whitelist`、`resources_whitelist`、`prompts_whitelist`，默认关
   - 开关关闭 = 全部允许，且关闭时后端清除该类型已保存的关联记录（"关即全允许"）
   - 开关开启 = 仅勾选的资源允许调用，开启前弹确认警示（开启后默认拒绝所有调用，需立即勾选）
   - 每行"默认/仅允许"单选改为二值开关；白名单关闭时整列置灰
   - 移除"全部允许"按钮（关闭白名单即等价于全部允许）
   - **BREAKING**：存量"隐式受限"语义废弃——升级后开关默认关，原受限密钥变为全放行（存量关联记录由管理员手动清理，系统不自动迁移）
3. **能力特性（capabilities）全局移除**（**BREAKING**）
   - 前端：模型列表/详情、厂商详情（列、编辑表单、详情描述）、密钥详情的能力列与表单整体移除
   - 后端：各 API 响应不再返回 `capabilities`，删除能力交集计算；`provider_models.capabilities` 数据库字段保留，同步写入逻辑保留
   - OpenSpec：`model-capabilities` 能力整体退役，其余相关规格中的 capabilities 场景一并清理
   - 注意：MCP 协议自身的 `capabilities`（服务协议能力）与本项无关，不改动
4. **映射对话框级联修复**
   - 模型映射新增/编辑对话框中切换厂商时，立即清空已选模型并重新加载该厂商的模型列表，杜绝陈旧 `provider_model_id` 触发后端 `provider mismatch` 拒绝
   - 后端 `provider mismatch` 校验保留作为 API 层防线
5. **厂商/密钥编辑对话框补状态控制**
   - 厂商与 API 密钥的新增/编辑对话框增加启用状态开关（对齐模型映射与 MCP 服务对话框的现有体验）
   - 后端 `Create` 接口补可选 `enabled` 字段（缺省仍为 `true`，向后兼容）；`Update` 已支持，零改动

## Capabilities

### New Capabilities

（无——五项均为对既有行为的修改）

### Modified Capabilities

- `table-sorting`：Providers 默认排序改为"状态+名称"（仅无偏好时）；新增"状态排序隐式名称第二因子、启用在前"要求；Models/Detail 状态列纳入可排序范围；"密钥禁用时全部允许按钮置灰"场景随按钮移除而删除
- `api-key-management`："设置 Key 权限"由隐式白名单改为显式开关语义（关=全允许并清空关联、开=仅选中允许+开启警示）；创建/编辑表单增加启用状态控制（新增 `enabled` 字段）
- `api-key-mcp-resources`：工具/资源/提示词权限配置各增加独立白名单开关（关闭清除关联、开启警示、执行与过滤按开关判定）；"全部允许"按钮移除
- `model-capabilities`：**整体退役**（全部需求移除，字段降级为仅存储）
- `model-management`：移除"能力交集正确显示"场景及能力列描述
- `manual-model-ui`：添加模型对话框字段列表移除能力相关字段
- `model-mapping`：`model_info` 不再包含 capabilities；移除详情页能力标签展示与能力交集统计场景；新增"切换厂商清空模型选择"对话框行为要求
- `model-mapping/alias-mapping`：清理 model_info 中的能力字段与详情页能力标签场景（历史遗留规格，随本项一并清理）
- `provider-model-sync`：合并保护与不伪造场景的描述中移除 `capabilities` 字段（同步写入行为本身保留）
- `provider-metadata`：厂商新增/编辑表单增加启用状态开关

## Impact

- **数据库**：`keys` 表新增 4 个布尔列（GORM AutoMigrate，SQLite/Postgres 均为 ADD COLUMN）；`provider_models.capabilities` 列保留不动；存量关联记录不自动迁移（管理员手动清理）
- **后端代码**：
  - `server/internal/model/db.go`（Key 结构体）
  - `server/internal/handler/key.go`（白名单开关 API、Create 补 enabled、ListModels 去 capabilities）
  - `server/internal/handler/proxy.go`、`proxy_mcp.go`（白名单执行路径改为开关驱动）
  - `server/internal/handler/provider.go`（Create 补 enabled）
  - `server/internal/handler/model.go`、`provider_model.go`（响应移除 capabilities、删除交集计算）
- **API**：`GET /keys/:id` 增加 4 个白名单字段；新增 `PUT /keys/:id/whitelist/<type>`；`POST /providers`、`POST /keys` 增加可选 `enabled`；模型/映射/密钥相关响应移除 `capabilities` 字段（**BREAKING**）
- **前端代码**：`web/src/views/` 下 Providers/Keys/Models/MCPs 的 index 与 Detail 共 8 个视图、`utils/tableSort.ts`（共享状态排序助手）、`locales/zh.ts`、`locales/en.ts`
- **规格**：10 个既有能力的 delta 变更，其中 `model-capabilities` 退役
