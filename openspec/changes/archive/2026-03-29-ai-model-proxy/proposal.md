## Why

企业和开发者需要同时使用多个大模型厂商的 API（OpenAI、Anthropic、Google 等），但各家 API 格式不统一，管理复杂。现有方案如 OpenRouter 虽然提供统一接口，但不支持私有化部署，无法自定义后端配置。

本项目提供一个可私有化部署的大模型 API 代理服务，统一暴露 OpenAI 兼容格式，后端可配置多个厂商，并提供 Web 控制面板进行可视化管理。

## What Changes

- **新增** 大模型 API 代理服务，统一暴露 OpenAI 兼容的 API 格式
- **新增** 格式转换层，支持 OpenAI 格式到 Anthropic 等厂商格式的双向转换
- **新增** Web 控制面板，用于管理厂商、模型映射、访问密钥和用量统计
- **新增** 用户认证系统，基于 Session 实现登录和权限控制
- **新增** 流式响应（SSE）处理和格式转换
- **新增** 多语言支持（中英文）和暗色模式

## Capabilities

### New Capabilities

- `api-proxy`: 核心 API 代理能力，接收 OpenAI 格式请求，转发到配置的后端厂商，支持流式和非流式响应
- `format-transformer`: 格式转换能力，在 OpenAI 格式和各厂商原生格式之间转换请求和响应
- `provider-management`: 厂商管理能力，支持添加、配置、启用/禁用多个后端厂商
- `model-sync`: 模型同步能力，从厂商 API 自动同步可用模型列表，检测模型变化
- `model-mapping`: 模型映射能力，将对外暴露的模型名映射到实际厂商模型，支持负载均衡
- `api-key-management`: 访问密钥管理能力，创建和管理客户端访问密钥，支持配额和权限控制
- `usage-tracking`: 用量追踪能力，记录每次 API 调用的详情，支持统计和查询
- `user-auth`: 用户认证能力，支持用户登录和会话管理
- `web-dashboard`: Web 控制面板，提供可视化的管理界面

### Modified Capabilities

（无现有能力需要修改）

## Impact

**技术栈**：
- 后端：Go + Gin + GORM + SQLite
- 前端：Vue 3 + TypeScript + Element Plus + ECharts

**新增组件**：
- API 网关层：处理 `/v1/chat/completions` 等 OpenAI 兼容接口
- 转换层：OpenAI ↔ Anthropic 格式转换
- 厂商适配器：OpenAI、Anthropic、OpenAI 兼容厂商
- Web 前端：登录、仪表盘、厂商管理、模型映射、密钥管理、用量统计等页面

**API 端点**：
- OpenAI 兼容接口：`/v1/chat/completions`, `/v1/models`
- 管理接口：`/api/v1/providers`, `/api/v1/model-mappings`, `/api/v1/api-keys`, `/api/v1/usage`, `/api/v1/auth`

**数据存储**：
- SQLite 数据库
- 核心表：providers, provider_models, model_mappings, api_keys, usage_logs, users
