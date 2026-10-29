# provider-metadata 规格

## Purpose

定义厂商元信息的记录与展示：官方主页（控制台入口）与付费类型（套餐 / 按量付费 / 免费）两个辅助字段的存储、校验和在控制台中的呈现方式。

## Requirements

### Requirement: Provider 可记录官方主页

系统 SHALL 允许为 Provider 记录可选的官方主页地址（`homepage`），在创建与更新时写入并在列表、详情接口中返回。

#### Scenario: 创建时携带主页
- **WHEN** 管理员创建 Provider，携带 `homepage="https://console.example.com"`
- **THEN** 系统存储该主页地址
- **AND** Provider 响应中 `homepage` 为 "https://console.example.com"

#### Scenario: 主页为可选字段
- **WHEN** 管理员创建或更新 Provider 未提供 `homepage`
- **THEN** 创建或更新成功
- **AND** `homepage` 为空字符串

#### Scenario: 更新主页
- **WHEN** 管理员更新 Provider 的 `homepage` 为新地址
- **THEN** 系统存储新地址并在响应中返回
### Requirement: 主页在控制台可点击跳转

系统 SHALL 在厂商列表的名称列显示主页跳转图标（仅当 `homepage` 非空），在厂商详情页头部显示主页链接，点击均在新窗口打开。

#### Scenario: 列表显示跳转图标
- **WHEN** 某厂商的 `homepage` 非空
- **THEN** 该厂商名称旁显示主页跳转图标
- **AND** 点击图标在新窗口打开主页地址

#### Scenario: 无主页时不显示图标
- **WHEN** 某厂商的 `homepage` 为空
- **THEN** 该厂商名称旁不显示跳转图标

#### Scenario: 详情页显示主页链接
- **WHEN** 管理员打开 `homepage` 非空的厂商详情页
- **THEN** 头部信息区显示主页链接
- **AND** 点击在新窗口打开
### Requirement: Provider 可记录付费类型

系统 SHALL 允许为 Provider 记录可选的付费类型（`billing_type`），取值限定为 `subscription`、`usage`、`free` 或空（未设置），存储英文枚举值并在展示层翻译为当前语言。

#### Scenario: 记录付费类型
- **WHEN** 管理员创建 Provider，携带 `billing_type="usage"`
- **THEN** 系统存储 "usage"
- **AND** 中文界面该厂商显示 tag "按量付费"
- **AND** 英文界面显示 "Pay-as-you-go"

#### Scenario: 拒绝非法枚举值
- **WHEN** 管理员提交 `billing_type="hourly"`
- **THEN** 请求被拒绝，返回 400 错误
- **AND** Provider 数据不发生变化

#### Scenario: 未设置付费类型
- **WHEN** Provider 的 `billing_type` 为空
- **THEN** 列表付费类型列显示 "-"
- **AND** 创建与更新均不报错

#### Scenario: 付费类型在列表以 tag 展示
- **WHEN** 厂商列表中某厂商 `billing_type="subscription"`
- **THEN** 付费类型列显示 tag "套餐"（中文）/ "Subscription"（英文）
### Requirement: 编辑表单支持元信息字段

系统 SHALL 在厂商新增 / 编辑表单中提供主页与付费类型两个可选输入：主页为 URL 格式校验的文本框，付费类型为枚举下拉框。

#### Scenario: 表单包含两个可选字段
- **WHEN** 管理员打开厂商新增或编辑表单
- **THEN** 表单包含"官方主页"输入框与"付费类型"下拉框（含未设置选项）
- **AND** 两者均为可选，不填可正常提交

#### Scenario: 主页格式校验
- **WHEN** 管理员在主页输入框填写非 URL 文本并提交
- **THEN** 表单提示格式错误
- **AND** 不发起提交请求
