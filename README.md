# AI Gateway
---

**此项目是作者学习大模型接口的副产物，最早的核心功能仅仅是实现 OpenAI/Anthropic API 的相互转换。**
**目前转换功能能处理文字内容（包括 “context”、“thinking”、“tools”），多模态转换从未尝试过。**

统一的 AI 服务网关平台。聚合多个大模型厂商的 API，未来将扩展支持 MCP/ACP 等协议代理，打造多协议 AI 服务接入中心。

## 特性

- **多协议网关**: 当前支持 OpenAI/Anthropic API 代理，未来将支持 MCP 等协议
- **OpenAI 兼容 API**: 暴露标准的 `/openai/v1/chat/completions` 和 `/openai/v1/models` 接口
- **Anthropic 兼容 API**: 暴露标准的 `/anthropic/v1/messages` 和 `/anthropic/v1/models` 接口
- **Web 控制台**: Vue 3 管理界面，支持中英文、暗色模式
- **多厂商支持**: 支持多种 AI 服务厂商（OpenAI、Anthropic双协议兼容），可轻松扩展
- **格式自动转换**: OpenAI ↔ Anthropic 请求/响应格式自动转换
- ~~**智能路由**: 支持多厂商轮询、故障转移和格式匹配优化~~
- **API Key 管理**: 生成和管理网关 API Key，支持模型访问权限控制
- **用量统计**: 请求日志和用量仪表盘，实时监控服务调用

## 外观

| 仪表盘 | 厂商列表 | 厂商模型 | 模型映射 | 密钥管理 |
| --- | --- | --- | --- | --- |
| ![仪表盘](docs/仪表盘.png) | ![厂商列表](docs/厂商列表.png) | ![厂商模型](docs/厂商模型.png) | ![模型映射](docs/模型映射.png) | ![密钥管理](docs/密钥管理.png) |

| 日志统计 | 日志统计 | 日志统计 |
| --- | --- | --- |
| ![日志统计](docs/日志统计_1.png) | ![日志统计](docs/日志统计_2.png) | ![日志统计](docs/日志统计_3.png) |

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+

### 安装运行

```bash
# 克隆项目
git clone https://github.com/lx0758/AI-Gateway.git ai-gateway

# 构建
cd ai-gateway
make

# 单独构建前端
cd ai-gateway/web
make

# 单独构建后端
cd ai-gateway/server
make

# 运行
ai-gateway/server/bin/ai-gateway-server
```
> 编译产物在 `ai-gateway/server/bin/ai-gateway-server`

服务启动后访问 <http://localhost:18080>

### 默认账号

- 用户名: `admin`
- 密码: `admin`

### 配置

支持 YAML 配置文件和环境变量两种配置方式，优先级：**环境变量 > YAML 配置 > 默认值**

#### YAML 配置

在 `server/` 目录下创建 `config.yaml` 文件（参考 `config.yaml.example`）：

```yaml
debug:
  enabled: false  # 调试模式开关

server:
  port: 18080     # 服务端口
  mode: debug     # 运行模式

database:
  type: sqlite    # 数据库类型 (sqlite/postgres)
  path: data.db   # SQLite 数据库路径

session:
  secret: ""      # Session 密钥
  max_age: 86400  # Session 有效期

auth:
  default_admin:
    username: admin
    password: admin
```

#### 环境变量配置

所有变量以 `AG_` 为前缀：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `AG_DEBUG_ENABLED` | `false` | 调试模式开关 |
| `AG_SERVER_PORT` | `18080` | 服务端口 |
| `AG_SERVER_MODE` | `debug` | 运行模式 (debug/release) |
| `AG_DATABASE_TYPE` | `sqlite` | 数据库类型 (sqlite/postgres) |
| `AG_DATABASE_PATH` | `data.db` | SQLite 数据库路径 |
| `AG_DATABASE_HOST` | `localhost` | PostgreSQL 服务器地址 |
| `AG_DATABASE_PORT` | `5432` | PostgreSQL 服务器端口 |
| `AG_DATABASE_USERNAME` | `postgres` | PostgreSQL 用户名 |
| `AG_DATABASE_PASSWORD` | `""` | PostgreSQL 密码 |
| `AG_DATABASE_DBNAME` | `ai_gateway` | PostgreSQL 数据库名 |
| `AG_SESSION_SECRET` | (自动生成) | Session 密钥，未设置时自动生成 |
| `AG_SESSION_MAX_AGE` | `86400` | Session 有效期(秒) |
| `AG_SESSION_SECURE` | `false` | Cookie Secure 标志 |
| `AG_SESSION_HTTP_ONLY` | `true` | Cookie HttpOnly 标志 |
| `AG_SESSION_SAME_SITE` | `lax` | Cookie SameSite 属性 |
| `AG_ADMIN_USERNAME` | `admin` | 默认管理员用户名 |
| `AG_ADMIN_PASSWORD` | `admin` | 默认管理员密码 |

#### 调试模式

启用调试模式（`debug.enabled: true` 或 `AG_DEBUG_ENABLED=true`）后：
- Gin 框架运行在 DebugMode
- Gorm 输出详细日志（Info 级别）
- Provider 记录请求/响应到 `debug/` 目录

### 数据库配置

#### SQLite（默认）

无需额外配置，数据库文件自动创建在 `server/data.db`：

```yaml
database:
  type: sqlite
  path: data.db
```

#### PostgreSQL

1. 创建数据库：

```sql
CREATE DATABASE ai_gateway;
CREATE USER your_username WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE ai_gateway TO ai_gateway;
```

2. 配置连接：

**YAML 方式** (`config.yaml`):

```yaml
database:
  type: postgres
  host: localhost
  port: 5432
  username: your_username
  password: your_password
  dbname: ai_gateway
```

**环境变量方式**:

```bash
AG_DATABASE_TYPE=postgres \
AG_DATABASE_HOST=localhost \
AG_DATABASE_PORT=5432 \
AG_DATABASE_USERNAME=your_username \
AG_DATABASE_PASSWORD=your_password \
AG_DATABASE_DBNAME=ai_gateway \
./ai-gateway-server
```

#### SQLite 到 PostgreSQL 迁移

1. 导出 SQLite 数据：

```bash
sqlite3 data.db .dump > backup.sql
```

2. 创建 PostgreSQL 数据库

3. 调整 SQL 语法（如有必要）

4. 导入数据：

```bash
psql -U your_username -d ai_gateway -f backup.sql
```

5. 更新配置切换到 PostgreSQL

### 示例

```bash
# 使用自定义端口
AG_SERVER_PORT=3000 \
./ai-gateway-server

# 生产环境配置
AG_SERVER_MODE=release \
AG_SESSION_SECRET=your-secret-key \
AG_ADMIN_PASSWORD=secure-password \
./ai-gateway-server
```

## 核心设计

- **别名抽象**：用户使用统一的模型名称（别名），无需关心后端实际模型
- **多后端映射**：一个别名可映射到多个 Provider，实现负载均衡和故障转移
- **权重路由**：按 weight 值降序排列，优先选择高权重 Provider
- **双协议支持**：每个 Provider 可同时配置 OpenAI 和 Anthropic BaseURL

## 工作原理

### 请求处理流程

```mermaid
flowchart TD
    subgraph 入口["入口 (双协议支持)"]
        A1["OpenAI 入口<br/>POST /openai/v1/chat/completions<br/>Authorization: Bearer sk-xxx"]
        A2["Anthropic 入口<br/>POST /anthropic/v1/messages<br/>x-api-key: sk-xxx"]
    end

    A1 --> B
    A2 --> B

    B["API Key 认证<br/>验证 Key 有效性和模型权限"]
    --> C["模型路由<br/>Alias → AliasMapping → Provider → ProviderModel<br/>按 weight 排序"]

    --> D{"协议匹配决策"}

    D -->|"协议相同"| S1
    D -->|"协议不同"| S2

    subgraph 直通流程["直通流程 (无需转换)"]
        S1["模型替换<br/>alias → actual_model_id"]
        --> R1["请求后端 API<br/>透传请求体"]
        --> RES1["返回响应<br/>流式/非流式"]
        --> TK1["Token 统计"]
    end

    subgraph 转换流程["转换流程 (协议转换)"]
        S2["模型替换<br/>alias → actual_model_id"]
        --> TR1["请求转换<br/>OpenAI ↔ Anthropic"]
        --> R2["请求后端 API"]
        --> TR2["响应转换<br/>Anthropic ↔ OpenAI"]
        --> RES2["返回响应<br/>流式/非流式"]
        --> TK2["Token 统计"]
    end

    TK1 --> I["用量记录<br/>写入 UsageLog"]
    TK2 --> I
```

### 路由决策流程(暂未实现负载均衡)

```mermaid
flowchart TD
    A["输入: model_name"] --> B["查询 Alias<br/>WHERE name = model_name<br/>AND enabled = true"]

    B --> C{"Alias 存在?"}

    C -->|"No"| D["返回 404<br/>model not found"]
    C -->|"Yes"| E["查询 AliasMapping<br/>WHERE alias_id = alias.ID<br/>AND enabled = true<br/>ORDER BY weight DESC"]

    E --> F{"Mapping 列表为空?"}

    F -->|"Yes"| G["返回 404<br/>no provider available"]
    F -->|"No"| H["遍历每个 Mapping"]

    H --> I["检查 Provider.enabled<br/>查询 ProviderModel<br/>WHERE is_available = true"]

    I --> J["构建 Provider 实例<br/>OpenAI BaseURL → OpenAIProvider<br/>Anthropic BaseURL → AnthropicProvider"]

    J --> K["返回 RouteResult 列表<br/>按 weight 降序排列"]
```

## API 接口

### OpenAI 兼容接口 (需要 API Key)

```
POST /openai/v1/chat/completions   # 聊天补全
GET  /openai/v1/models             # 模型列表
GET  /openai/v1/models/:id         # 模型详情
```

### Anthropic 兼容接口 (需要 API Key)

```
POST /anthropic/v1/messages        # Anthropic Messages API
GET  /anthropic/v1/models          # 模型列表
POST /anthropic/v1/models/:id      # 模型详情
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

GET  /api/v1/aliases        # 模型别名列表
POST /api/v1/aliases        # 创建别名
PUT  /api/v1/aliases/:id    # 更新别名
DELETE /api/v1/aliases/:id  # 删除别名

GET  /api/v1/api-keys       # API Key 列表
POST /api/v1/api-keys       # 创建 API Key
PUT  /api/v1/api-keys/:id   # 更新 API Key
DELETE /api/v1/api-keys/:id # 删除 API Key
POST /api/v1/api-keys/:id/reset # 重置 API Key

GET  /api/v1/usage/stats    # 用量统计
GET  /api/v1/usage/logs     # 用量日志
GET  /api/v1/usage/dashboard # 仪表盘数据
```

## 使用示例

```bash
# OpenAI 格式调用
curl http://localhost:18080/openai/v1/chat/completions \
  -H "Authorization: Bearer sk-your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Anthropic 格式调用
curl http://localhost:18080/anthropic/v1/messages \
  -H "x-api-key: sk-your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## 项目结构

```
ai-gateway/
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
│   │   ├── provider/           # 厂商实现
│   │   ├── router/             # 模型路由
│   └── go.mod
└── openspec/                   # 设计文档
```

## Docker/Alpine 部署

### 构建静态二进制文件（支持 Alpine）

```bash
# 在 Alpine 容器中构建
docker run --rm -v "$PWD/server:/app" -w /app alpine:latest \
  sh -c "apk add --no-cache gcc musl-dev go && \
         CGO_ENABLED=1 go build -ldflags '-linkmode external -extldflags \"-static\"' -o bin/ai-gateway-server ./cmd/server/main.go"

# 或在 Linux 主机上交叉编译
cd server
CGO_ENABLED=1 go build -ldflags '-linkmode external -extldflags "-static"' -o bin/ai-gateway-server ./cmd/server/main.go
```

### Docker 部署示例

```dockerfile
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY ai-gateway-server /usr/local/bin/
EXPOSE 18080
CMD ["ai-gateway-server"]
```

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
go run ./cmd/server                                     # 运行
go build -o bin/ai-gateway-server ./cmd/server/main.go  # 构建
```

## License

MIT