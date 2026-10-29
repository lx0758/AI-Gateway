# Proposal

## Why

管理员配置厂商后，经常需要找回该厂商的官方控制台入口，目前无处记录、每次靠搜索；同时厂商的付费模式（套餐 / 按量付费 / 免费）无法记录，选型和对账时缺少直观信息。

## What Changes

- Provider 增加 `homepage` 字段（可选 URL）：厂商官方主页 / 控制台地址
- Provider 增加 `billing_type` 辅助字段：付费类型，存英文枚举值 `subscription`（套餐）/ `usage`（按量付费）/ `free`（免费），空值 = 未设置；展示层做 i18n 翻译
- 厂商列表：名称列内嵌主页跳转图标（新窗口打开）；新增"付费类型"tag 列
- 厂商详情页：头部展示主页链接与付费类型 tag
- 新增 / 编辑表单增加两个可选字段：主页输入框（URL 格式校验）、付费类型下拉框

## Capabilities

### New Capabilities

- `provider-metadata`: 厂商元信息——官方主页与付费类型的记录、校验与展示。

### Modified Capabilities

（无；`provider-multi-style` 仅覆盖 BaseURL 双协议配置，不受影响。）

## Impact

- `server/internal/model/db.go`：Provider PO 增加 `Homepage`、`BillingType`
- `server/internal/handler/provider.go`：create / update / response DTO 增加字段与值域校验
- `web/src/views/Providers/index.vue`：编辑表单、列表列
- `web/src/views/Providers/Detail.vue`：头部展示
- `web/src/locales/zh.ts` / `en.ts`：新增文案与枚举翻译
- DB：GORM AutoMigrate 自动加列（默认空字符串），无手工迁移；对现有数据与旧客户端向后兼容
