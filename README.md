# AI Model Proxy

一个统一的 AI 模型代理服务，将多个大模型厂商的 API 聚合为 OpenAI 兼容格式。

## 特性

- **OpenAI 兼容 API**: 暴露标准的 `/openai/v1/chat/completions` 和 `/openai/v1/models` 接口
- **多厂商支持**: 支持 OpenAI、Anthropic 等厂商，可轻松扩展
- **格式自动转换**: OpenAI ↔ Anthropic 请求/响应格式自动转换
- **模型别名映射**: 将模型别名映射到实际厂商模型
- **负载均衡**: 支持多厂商轮询和故障转移
- **API Key 管理**: 生成和管理代理 API Key
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
DELETE /api/v1/api-keys/:id # 删除 API Key

GET  /api/v1/usage/stats    # 用量统计
GET  /api/v1/usage/logs     # 用量日志
GET  /api/v1/usage/dashboard # 仪表盘数据
```

## 使用示例

### 1. 添加厂商

```bash
curl -X POST http://localhost:18080/api/v1/providers \
  -H "Content-Type: application/json" \
  -b "session=your-session-cookie" \
  -d '{
    "name": "OpenAI",
    "api_type": "openai",
    "api_key": "sk-xxx",
    "base_url": "https://api.openai.com/v1"
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
    "name": "My API Key",
    "allowed_models": ["gpt-4"]
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

## 支持的厂商

| 厂商 | API Type | 格式转换 |
|------|----------|---------|
| OpenAI | `openai` | 直通 |
| Anthropic | `anthropic` | OpenAI ↔ Anthropic |
| 其他 OpenAI 兼容 | `openai` | 直通 |

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
