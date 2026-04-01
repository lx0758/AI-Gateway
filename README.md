# AI Proxy

统一的 AI 服务代理平台。当前版本聚合多个大模型厂商的 API 为 OpenAI 兼容格式，未来将扩展支持 MCP/ACP 等协议代理。

## 特性

- **OpenAI 兼容 API**: 暴露标准的 `/openai/v1/chat/completions` 和 `/openai/v1/models` 接口
- **多厂商支持**: 支持 `@ai-sdk/openai-compatible` 和 `@ai-sdk/anthropic` 类型厂商，可轻松扩展
- **格式自动转换**: OpenAI ↔ Anthropic 请求/响应格式自动转换
- **模型别名映射**: 将模型别名映射到实际厂商模型，支持编辑修改
- **负载均衡**: 支持多厂商轮询和故障转移
- **API Key 管理**: 生成和管理代理 API Key，支持模型访问权限控制
- **手动模型保护**: 同步厂商模型时保护手动创建的模型配置
- **用量统计**: 请求日志和用量仪表盘
- **Web 控制台**: Vue 3 管理界面，支持中英文、暗色模式

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+ (仅开发前端时需要)

### 安装运行

```bash
# 克隆项目
git clone https://github.com/user/ai-model-proxy.git
cd ai-model-proxy

# 构建前端
cd web && npm install && npm run build && cd ..

# 运行服务
cd server && go run ./cmd/server

# 或构建后运行
cd server && go build -o ai-model-proxy ./cmd/server
./ai-model-proxy
```

服务启动后访问 http://localhost:18080

### 默认账号

- 用户名: `admin`
- 密码: `admin`

## 配置

使用环境变量配置，所有变量以 `AMP_` 为前缀：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `AMP_SERVER_PORT` | `18080` | 服务端口 |
| `AMP_SERVER_MODE` | `debug` | 运行模式 (debug/release) |
| `AMP_DATABASE_PATH` | `data.db` | SQLite 数据库路径 |
| `AMP_SESSION_SECRET` | (自动生成) | Session 密钥，未设置时自动生成 |
| `AMP_SESSION_MAX_AGE` | `86400` | Session 有效期(秒) |
| `AMP_SESSION_SECURE` | `false` | Cookie Secure 标志 |
| `AMP_SESSION_HTTP_ONLY` | `true` | Cookie HttpOnly 标志 |
| `AMP_SESSION_SAME_SITE` | `lax` | Cookie SameSite 属性 |
| `AMP_ADMIN_USERNAME` | `admin` | 默认管理员用户名 |
| `AMP_ADMIN_PASSWORD` | `admin` | 默认管理员密码 |

### 示例

```bash
# 使用自定义端口
AMP_SERVER_PORT=3000 ./ai-model-proxy

# 生产环境配置
AMP_SERVER_MODE=release \
AMP_SESSION_SECRET=your-secret-key \
AMP_ADMIN_PASSWORD=secure-password \
./ai-model-proxy
```

## API 接口

### OpenAI 兼容接口 (需要 API Key)

```
POST /openai/v1/chat/completions   # 聊天补全
GET  /openai/v1/models             # 模型列表
GET  /openai/v1/models/:id         # 模型详情
```

### 管理接口 (需要登录)

```
POST /api/v1/auth/login     # 登录
POST /api/v1/auth/logout    # 登出
GET  /api/v1/auth/me        # 当前用户
PUT  /api/v1/auth/password  # 修改密码

GET  /api/v1/providers      # 厂商列表
POST /api/v1/providers      # 创建厂商
PUT  /api/v1/providers/:id  # 更新厂商
DELETE /api/v1/providers/:id # 删除厂商
POST /api/v1/providers/:id/test  # 测试连接
POST /api/v1/providers/:id/sync  # 同步模型

GET  /api/v1/model-mappings # 模型映射列表
POST /api/v1/model-mappings # 创建映射
PUT  /api/v1/model-mappings/:id  # 更新映射
DELETE /api/v1/model-mappings/:id # 删除映射

GET  /api/v1/api-keys       # API Key 列表
POST /api/v1/api-keys       # 创建 API Key
PUT  /api/v1/api-keys/:id   # 更新 API Key
DELETE /api/v1/api-keys/:id # 删除 API Key
GET  /api/v1/api-keys/:id/models  # 获取 Key 允许的模型列表
POST /api/v1/api-keys/:id/models  # 添加模型权限
DELETE /api/v1/api-keys/:id/models/:model_alias # 删除模型权限

GET  /api/v1/usage/stats    # 用量统计
GET  /api/v1/usage/logs     # 用量日志
GET  /api/v1/usage/dashboard # 仪表盘数据
```

## 使用示例

### 1. 添加厂商

```bash
# 添加只支持 OpenAI 格式的厂商
curl -X POST http://localhost:18080/api/v1/providers \
  -H "Content-Type: application/json" \
  -b "session=your-session-cookie" \
  -d '{
    "name": "OpenAI",
    "openai_base_url": "https://api.openai.com/v1",
    "api_key": "sk-xxx"
  }'

# 添加只支持 Anthropic 格式的厂商
curl -X POST http://localhost:18080/api/v1/providers \
  -H "Content-Type: application/json" \
  -b "session=your-session-cookie" \
  -d '{
    "name": "Anthropic",
    "anthropic_base_url": "https://api.anthropic.com/v1",
    "api_key": "sk-xxx"
  }'

# 添加同时支持两种格式的厂商
curl -X POST http://localhost:18080/api/v1/providers \
  -H "Content-Type: application/json" \
  -b "session=your-session-cookie" \
  -d '{
    "name": "My Service",
    "openai_base_url": "https://my-service.com/v1",
    "anthropic_base_url": "https://my-service.com/v1",
    "api_key": "sk-xxx"
  }'
```

### 2. 创建模型映射

```bash
curl -X POST http://localhost:18080/api/v1/model-mappings \
  -H "Content-Type: application/json" \
  -b "session=your-session-cookie" \
  -d '{
    "alias": "gpt-4",
    "provider_id": 1,
    "model_name": "gpt-4-turbo-preview",
    "enabled": true
  }'
```

### 3. 创建 API Key

```bash
curl -X POST http://localhost:18080/api/v1/api-keys \
  -H "Content-Type: application/json" \
  -b "session=your-session-cookie" \
  -d '{
    "name": "my-api-key",
    "models": ["gpt-4"]
  }'
```

### 4. 调用代理 API

```bash
curl http://localhost:18080/openai/v1/chat/completions \
  -H "Authorization: Bearer sk-your-proxy-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## 项目结构

```
ai-model-proxy/
├── web/                        # Vue 3 前端
│   ├── src/
│   │   ├── views/              # 页面组件
│   │   ├── stores/             # Pinia 状态
│   │   ├── locales/            # i18n 翻译
│   │   └── api/                # API 客户端
│   └── vite.config.ts
├── server/                     # Go 后端
│   ├── cmd/server/main.go      # 入口
│   ├── internal/
│   │   ├── config/             # 配置加载
│   │   ├── handler/            # HTTP 处理器
│   │   ├── middleware/         # 中间件
│   │   ├── model/              # 数据模型
│   │   ├── router/             # 模型路由
│   │   └── transformer/        # 格式转换
│   └── go.mod
└── openspec/                   # 设计文档
```

## 厂商配置说明

一个厂商可以配置一个或两个 BaseURL：
- `openai_base_url`: OpenAI 兼容格式的接口地址
- `anthropic_base_url`: Anthropic 格式的接口地址

**格式转换**：
- 当请求格式与厂商支持的格式匹配时，直接透传（无转换）
- 当请求格式与厂商支持的格式不匹配时，自动转换（OpenAI ↔ Anthropic）
- 路由优先匹配同格式的厂商，减少转换开销

**示例场景**：
- OpenAI 官方：只配置 `openai_base_url`
- Anthropic 官方：只配置 `anthropic_base_url`
- 自建服务：同时配置两个 BaseURL（如果支持两种格式）

## 开发

### 前端开发

```bash
cd web
npm install
npm run dev     # 启动开发服务器
npm run build   # 构建生产版本
```

### 后端开发

```bash
cd server
go run ./cmd/server           # 运行
go build -o ai-model-proxy ./cmd/server  # 构建
```

## 迁移指南 (v0.x → v1.0)

### API 路径变更

| 旧路径 | 新路径 |
|--------|--------|
| `/v1/chat/completions` | `/openai/v1/chat/completions` |
| `/v1/models` | `/openai/v1/models` |

### 配置方式变更

- **旧版本**: 使用 `configs/config.yaml` 文件配置
- **新版本**: 使用环境变量配置，无需配置文件

### 默认值变更

| 配置项 | 旧默认值 | 新默认值 |
|--------|----------|----------|
| 服务端口 | `8080` | `18080` |
| 数据库路径 | `./data/ai-model-proxy.db` | `data.db` |
| 管理员密码 | `admin123` | `admin` |
| Session 密钥 | 硬编码 | 自动生成 |

## License

MIT
