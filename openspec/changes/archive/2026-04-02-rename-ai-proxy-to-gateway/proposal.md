## Why

项目定位从单纯的 AI 模型 API 代理扩展为综合 AI 网关，下一步将增加 MCP (Model Context Protocol) 协议的代理能力。"网关"这一名称更能准确反映系统的定位——作为多种 AI 服务协议的统一接入点，而非简单的代理转发。

当前名称"AI代理"/"AI Proxy"在语义上偏重于请求转发，未能充分体现网关应具备的路由、认证、限流、监控等综合能力。

## What Changes

- **项目名称**: "AI代理" → "AI网关" (中文)，"AI Proxy" → "AI Gateway" (英文)
- **Go module 名称**: `ai-proxy` → `ai-gateway`
- **构建产物名**: `ai-model-proxy` → `ai-gateway`
- **环境变量前缀**: `AMP_` (AI Model Proxy) → `AG_` (AI Gateway)
- **Session 名称**: `ai-proxy-session` → `ai-gateway-session`
- **源码 import 路径**: 所有 `ai-proxy/internal/*` → `ai-gateway/internal/*`
- **前端标题**: "AI 代理"/"AI Proxy" → "AI 网关"/"AI Gateway"
- **README.md**: 重新整理，突出网关定位，保留核心特性说明，更新所有相关名称
- **owned_by 字段**: `ai-proxy` → `ai-gateway`

## Capabilities

### New Capabilities

无新增功能能力，本次变更仅为命名调整。

### Modified Capabilities

无规格级别的需求变更，本次变更不影响功能行为，仅涉及命名和标识符的调整。

## Impact

### 代码层面

- **Go 源码**: 所有 `.go` 文件中的 import 路径（约 31+ 处）
- **Go module**: `server/go.mod` 模块声明
- **Session 配置**: `server/internal/middleware/auth.go:36`
- **Proxy handler**: `server/internal/handler/proxy_openai.go:102,124` 的 `owned_by` 字段

### 前端层面

- **国际化文件**: `web/src/locales/zh.ts` 和 `web/src/locales/en.ts` 的标题字段
- **页面标题**: 登录页、主页等所有显示项目名称的地方

### 文档层面

- **README.md**: 项目名称、描述、构建命令、环境变量说明等全部相关内容
- **历史设计文档**: `openspec/changes/archive/` 下的相关文档（保持历史记录，不做修改）

### 环境变量

- **前缀变更**: `AMP_*` → `AG_*`，影响所有配置项：
  - `AMP_SERVER_PORT` → `AG_SERVER_PORT`
  - `AMP_SERVER_MODE` → `AG_SERVER_MODE`
  - `AMP_DATABASE_PATH` → `AG_DATABASE_PATH`
  - `AMP_SESSION_*` → `AG_SESSION_*`
  - `AMP_ADMIN_*` → `AG_ADMIN_*`

### 构建产物

- **可执行文件名**: `ai-model-proxy` → `ai-gateway`
- **Git 仓库名**: 建议同步更新为 `ai-gateway`（需用户手动操作）