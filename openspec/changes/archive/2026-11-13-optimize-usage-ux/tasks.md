# Tasks

## 1. 后端：白名单数据模型与 API

- [x] 1.1 `server/internal/model/db.go` 的 Key 结构体新增 `models_whitelist`/`tools_whitelist`/`resources_whitelist`/`prompts_whitelist` 四个布尔字段并纳入 AutoMigrate — 验证：`go build ./...` 通过，启动服务后用 `PRAGMA table_info(keys)`（SQLite）或 `\d keys`（PG）可见 4 个新列且默认 false
- [x] 1.2 `key.go` 白名单读写：`keyResponse`/`Get` 响应增加 4 个白名单字段；新增 `PUT /keys/:id/whitelist/:type` handler（body `{"enabled": bool}`，`enabled=false` 时同事务删除该类型全部关联记录）并在 `main.go` 注册路由 — 验证：curl `GET /keys/:id` 返回 4 字段；curl PUT `enabled=false` 后该 key 对应关联表记录数为 0
- [x] 1.3 `createKeyRequest` 增加 `Enabled *bool`（缺省视为 true）；`Create` 携带 `req.Models` 时同步置 `models_whitelist=true` — 验证：curl 创建带模型列表的 key 后 `models_whitelist=true`；不带 `enabled` 创建的 key 状态为启用

## 2. 后端：白名单执行路径切换

- [x] 2.1 `proxy.go VerifyKeyID` 改为开关短路两段式：`ModelsWhitelist=false` 直接放行，`true` 时查 join 表判定 — 验证：curl 代理入口，开关关时任意模型放行；开+未勾选模型返回 403 "模型不允许"；开+已勾选模型放行
- [x] 2.2 `proxy_mcp.go` 的 `hasFullTool/Resource/PromptAccess`、三类 list 过滤、initialize 过滤与逐项调用检查全部改为开关驱动 — 验证：tools/resources/prompts 分别实测四态：关=全放行、开+勾选=放行、开+未勾选=JSON-RPC -32602、开+0 勾选=全拒

## 3. 后端：Create 补 enabled 与 capabilities 响应移除

- [x] 3.1 `createProviderRequest`/`createKeyRequest` 增加 `Enabled *bool`，Create 中按 `nil→true` 缺省逻辑写入 — 验证：curl `POST /providers` 与 `POST /keys` 分别携带 `enabled=false` 与不携带，前者创建禁用对象、后者创建启用对象
- [x] 3.2 `model.go`/`key.go`/`provider_model.go` 响应结构删除 `Capabilities` 字段，删除 `calculateCapabilitiesIntersection` 及全部调用点，`createProviderModelRequest`/`updateProviderModelRequest` 删除 `Capabilities` 输入字段 — 验证：`go build ./...` 通过；curl 模型列表、映射列表、`GET /keys/:id/models` 响应 JSON 均不含 `"capabilities"` 键，且 MCP 协议 initialize 响应的 `capabilities` 不受影响
- [x] 3.3 验证同步写入链路完好：触发一次厂商模型同步，对比同步前后 `provider_models.capabilities` 列 — 验证：同步返回非空时字段被覆盖、返回空时本地值保留（smart-model-sync 合并策略未被破坏）
- [x] 3.4 在 README.md 新增升级说明：两处 BREAKING（存量白名单关联记录手动清理 SQL：`DELETE FROM key_models/key_mcp_tools/key_mcp_resources/key_mcp_prompts`；响应不再返回 `capabilities`）— 验证：README 含清理语句与响应字段移除说明

## 4. 前端：排序

- [x] 4.1 `web/src/utils/tableSort.ts` 新增 `sortByStatusThenName` 共享助手：升序启用在前、名称为次因子、非布尔状态值退化为字符串比较 — 验证：`npm run build`（vue-tsc 类型检查）通过
- [x] 4.2 各可排序状态列挂 `:sort-method`（Providers/Keys/MCPs 三个 index、MCPs/Detail×3、Keys/Detail×4、Providers/Detail 的 availability 列、Models/Detail 状态列补 `prop` 与 `sortable`）— 验证：手测各表点击状态列——启用在前、状态相同按名称、降序时名称同向降序
- [x] 4.3 Providers 默认排序改为 `getSortConfig('providers', 'enabled')`（不更换 localStorage 存储键）；`handleSortChange` 在 `order === null` 时清除偏好（取消语义按 8.3 走查反馈修订为 design D4 新方案，原编程式恢复已废弃）— 验证：清 localStorage 后打开厂商页默认"状态+名称"；第三次点击表头后箭头消失且数据保持默认顺序；已有排序偏好的老用户不受影响

- [x] 4.4 其余列表取消排序恢复默认：Keys/Models/MCPs 三个主列表、Providers/Detail、MCPs/Detail×3、Keys/Detail×4 共 11 张表在 `order === null` 时清除偏好（按 8.3 走查反馈改为 design D4 新方案：默认顺序由数据源预排序承载，Models/Detail 数据序即权重默认序无需改动）— 验证：各列表第三次点击表头后箭头消失且数据保持默认顺序

## 5. 前端：密钥白名单 UI

- [x] 5.1 `Keys/Detail.vue` 四个 tab：tab-actions 的"全部允许"按钮替换为白名单 `el-switch`；每行单选 [默认|仅允许] 改为二值开关；白名单关闭时整列禁用置灰并显示 tooltip — 验证：`npm run build` 通过；手测关闭时行控件禁用、按钮已移除
- [x] 5.2 白名单开关交互：开启前 `ElMessageBox.confirm` 弹"将拒绝所有调用"警示，确认后调 `PUT /whitelist/<type>`；关闭时调同一端点（后端清记录）后刷新全部行状态；开启后所有行呈未勾选 — 验证：手测开启出现警示、关闭后重开全部行未勾选、勾选行立即生效
- [x] 5.3 `locales/zh.ts`、`locales/en.ts` 新增开关与警示文案，删除 `key.allowAll` 词条 — 验证：中英文界面抽查无缺失翻译 key 报错

## 6. 前端：capabilities 展示移除

- [x] 6.1 删除能力相关 UI：Models/index、Models/Detail、Providers/Detail 的能力列，Keys/Detail 的死列，Providers/Detail 编辑表单 checkbox 组与详情描述，以及 `capabilityList`/`updateCapabilityList`/`getCapabilitiesString` 辅助函数 — 验证：`npm run build` 通过；四个页面无能力相关 UI 残留
- [x] 6.2 `locales/zh.ts`、`locales/en.ts` 删除 `models.capabilities`、`provider.capabilities`、`key.allowAll` 及 `provider.tools/stream/photo/image/video` 标签词条 — 验证：`grep -rn "capabilities" web/src/locales` 无残留
- [x] 6.3 展示层残留扫描：`grep -rn "capabilities\|capabilityList" web/src` 确认除 MCP 协议无关项外无展示引用 — 验证：扫描输出仅剩不相关匹配（如类型定义中的已删字段一并清除）

## 7. 前端：映射对话框与表单状态控制

- [x] 7.1 `Models/Detail.vue` 映射对话框厂商下拉 `@change` 改为 `handleProviderChange`：同步清空 `provider_model_id`、清空选项列表、再加载新厂商模型 — 验证：手测切换厂商后模型框立即清空、未重选直接保存被必填拦截、重选后保存成功且不再出现 `provider mismatch`
- [x] 7.2 `Providers/index.vue`、`Keys/index.vue` 表单增加 `enabled` 开关（编辑回填当前值、新增默认启用、提交随 body 发送）— 验证：手测新增/编辑禁用厂商与密钥后列表状态同步正确，与 MCPs 对话框行为一致

## 8. 集成验证

- [x] 8.1 构建与类型检查：`go build ./...` 与 `npm run build` 均零错误通过
- [x] 8.2 白名单四态 × 四资源类型 curl 回归（关/开+勾选/开+未勾选/开+零勾选）— 验证：结果与 `api-key-management`、`api-key-mcp-resources` delta 规格场景逐条一致
- [ ] 8.3 五议题端到端手测走查：排序默认与次因子与取消恢复、白名单 UI 全流程、capabilities 无残留、映射对话框级联、两表单状态控制 — 验证：与 10 个 delta 规格的场景清单对照无遗漏

## 9. 走查发现问题修复（8.3 反馈）

- [x] 9.1 状态列排序名称次因子方向：`sortByStatusThenName` 增加 `tableRef` 参数（第 3 位），`sortMethod` 执行瞬间直接读 `tableRef.store.states.sortOrder` 感知方向并预乘抵消，使名称恒为 A→Z（不随升降序翻转）；13 个调用点全部传入各自表格实例，`Models/Detail` 补 `ref="tableRef"` — 验证：`npm run build` 通过；升/降序手测名称均 A→Z（并入 8.3）
- [x] 9.2 密钥 models 列表删除"映射数量"列（width 改 min-width 使列表铺满），并修 `key.go ListModels` 统计口径：补 `Preload("ProviderModel")`、`MappingCount` 改用 `calculateEnabledCount`（与 `model.go` 三层过滤一致）— 验证：对照数据下 `GET /models` 与 `GET /keys/:id/models` 均返回 `mapping_count=1, min_ctx=200000, min_out=8192`（修复前 key 侧为 `2/0/0`）；规格同步修订 `table-sorting` delta（次因子恒升序 + 详情子页面可排序字段去掉映射数量）与 `design.md` D3/风险项

## 10. 第二轮走查反馈修复（8.3 反馈 #1/#2 语义澄清后）

- [x] 10.1 取消排序改为"清偏好+数据源预排序承载默认顺序"（design D4 新方案，废弃编程式 `sort()` 恢复）：`getSortConfig` 无偏好/与硬默认相同时返回 `undefined`（不绑 `default-sort` → 无箭头）、新增 `removeSortConfig`、7 个视图取数后 `sortByPropAsc`/`sortByStatusThenName` 预排序、`handleSortChange` 的 `order===null` 分支只清偏好 — 验证：`vue-tsc -b` 与 `npm run build` 零错误；手测厂商页取消后箭头消失且数据仍为状态+名称（并入 8.3）
- [x] 10.2 中英名称顺序统一为"字母在前+中文按拼音"：新增 `compareName`（分块：非中文 `en` 字母序在前、中文 `zh-CN` 拼音序在后，替代裸 `localeCompare` 跟随系统 locale 的行为），`sortByStatusThenName` 次因子与 `Models/Detail` 两个下拉排序改用之 — 验证：`npm run build` 通过；中英混合列表字母在前、中文按拼音（并入 8.3）
- [x] 10.3 规格与设计同步：`table-sorting` delta（次因子措辞改"字母在前+中文拼音"并新增中英混合场景、取消排序场景改"箭头消失+清偏好+数据源默认序"并新增刷新场景、无偏好场景加"不显示排序指示器"）、`design.md` D3/D4 重写与风险项更新、`tasks.md` 4.3/4.4 描述修订

## 11. 第三轮走查反馈修复（8.3 反馈 #3：白名单开+0 勾选统计仍显示"不限制"）

- [x] 11.1 后端 `GET /keys` 主列表响应 `keyListItemResponse` 补 4 个白名单开关字段（`models/tools/resources/prompts_whitelist`，原仅详情 `keyResponse` 携带）并在 `List` 中填充 — 验证：`go build ./...` 通过；curl 确认响应含 4 字段（见 11.3）
- [x] 11.2 前端 `Keys/index.vue` 四个统计列改两分支判断（按白名单开关而非记录数）：关→"不限制"（灰色）、开→"允许 N 个"（开+0 勾选即"允许 0 个"） — 验证：`npm run build` 通过；手测白名单开+不勾选显示"允许 0 个"（并入 8.3）
- [x] 11.3 规格同步：`api-key-management` delta 新增场景"主列表统计反映白名单状态"（关=不限制、开+N=允许 N 个、开+0=允许 0 个）与"白名单关闭时勾选开关显示为开启且禁用"；`api-key-mcp-resources` delta 三类资源各新增"白名单关闭时勾选开关显示为开启且禁用"场景
- [x] 11.4 详情页 4 个 tab 的 item 勾选开关：白名单关闭时显示为**开启且禁用**（原显示关闭+禁用，误导为"未允许"） — 验证：`npm run build` 通过；手测白名单关时全部 item 开关为开启+禁用、tooltip 提示"白名单已关闭，全部允许"（并入 8.3）
- [x] 11.5 父对象禁用不再限制子对象配置（用户反馈"没必要只能开启的时候配置"，全部详情统一移除）：Keys/Detail 4 个白名单总开关去掉 `:disabled="!key?.enabled"` 与"密钥已禁用" tooltip；MCPs/Detail 3 处（工具/资源/提示词）、Models/Detail 2 处（模型禁用 + `forced_disabled`）删除"父禁用→开关置灰"分支，`getDisableTooltip` 函数删除；后端核实 `UpdateWhitelist`/`UpdateMapping`/`UpdateTool` 等配置接口均不检查父对象 enabled（前端放开不撞限制，生效仍由执行路径按父状态过滤） — 验证：`npm run build` 通过；手测禁用的厂商/模型/密钥/MCP 服务下 item 均可配置（并入 8.3）
- [x] 11.6 规格与 i18n 同步：`table-sorting` delta 将 requirement"父对象禁用时子对象状态显示为禁用"从 MODIFIED 移至 **REMOVED**（Reason/Migration 说明配置与启用解耦）；i18n 清理 7 个死键（zh/en 各 7：`keyDisabled`、`serviceDisabled`、`modelDisabled`、`providerDisabled`、`providerDisabledReason`、`providerModelUnavailableReason`、`providerOrModelDisabled`） — 验证：`openspec validate` 通过；`grep` 确认死键无残留
- [x] 11.7 详情页"已禁用"状态 tag 改红色醒目展示：Keys/Providers/MCPs/Models 四个详情页顶部状态 tag 禁用态 `type` 由 `info`（灰）改为 `danger`（红）— 验证：`npm run build` 通过；手测四个详情页禁用状态显示红色（并入 8.3）

## Workflow follow-up

- 实施与评审完成后归档变更（`/opsx-archive`）。
- 验证归档结果（含 `model-capabilities` 能力退役生效）。
