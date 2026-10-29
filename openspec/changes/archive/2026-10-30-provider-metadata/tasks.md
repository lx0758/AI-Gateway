# Tasks

## 1. 后端字段与校验

- [x] 1.1 修改 `server/internal/model/db.go` Provider PO：增加 `Homepage string`、`BillingType string`（AutoMigrate 注册无需改动）；验证：`cd server && go build ./...` 通过，启动服务后 `PRAGMA table_info(providers)`（或 psql `\d providers`）可见两新列
- [x] 1.2 修改 `server/internal/handler/provider.go`：create / update / response DTO 增加 `homepage`、`billing_type` 字段与赋值映射，update 为指针语义（未传不覆盖）；`billing_type` 增加 `binding:"omitempty,oneof=subscription usage free"`；验证：创建携带两字段后响应回显一致；`billing_type="hourly"` 返回 400 且数据不变；创建不带两字段成功且为空

## 2. 前端表单与展示

- [x] 2.1 修改 `web/src/views/Providers/index.vue` 表单：新增"官方主页"输入框（URL 格式校验、可选）与"付费类型"下拉框（含"未设置"选项，枚举渲染）；同步编辑回显；验证：`cd web && make build` 通过，非法 URL 提交被表单拦截，合法数据保存后重新打开回显正确
- [x] 2.2 列表展示：名称列内嵌主页跳转图标（`homepage` 非空时显示，新窗口打开）；新增"付费类型"tag 列（枚举映射 i18n，空显示 "-"）；补充 zh/en 文案（含 `provider.billing.subscription`="套餐"/"Subscription"、`usage`="按量付费"/"Pay-as-you-go"、`free`）；验证：有主页的厂商显示图标且点击新开页面、无主页的不显示；付费类型 tag 中英文切换正确
- [x] 2.3 修改 `web/src/views/Providers/Detail.vue` 头部：展示主页链接与付费类型 tag；验证：详情页两信息正确显示，链接新窗口打开

## 3. 集成验证

- [x] 3.1 根目录 `make build` 通过；完整走查 spec 场景：新建厂商填两字段 → 列表/详情正确展示 → 编辑清空 → 图标消失、tag 变 "-"；非法枚举被 400 拒绝
- [x] 3.2 运行 `openspec validate provider-metadata --strict` 通过

## Workflow follow-up

- 实现完成后运行 `/opsx-apply` 的验证流程，或直接归档：`openspec archive provider-metadata`
