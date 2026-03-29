## Why

项目进入 v1 稳定版本阶段，需要调整项目结构、配置方式和 API 路径，以提升可维护性和部署灵活性。

当前问题：
- 前后端代码混在同一目录结构中，职责不清晰
- 使用 YAML 配置文件，环境变量支持不足，容器化部署不便
- 代理 API 路径 `/v1/*` 与其他服务冲突风险高
- Go 模块名使用 `github.com/user/ai-model-proxy`，不够简洁

## What Changes

- **BREAKING** 项目结构重组：前端移至 `web/`，后端移至 `server/`
- **BREAKING** 配置方式变更：移除 YAML 配置文件，改用环境变量（`AMP_` 前缀）
- **BREAKING** 代理 API 路径调整：`/v1/*` → `/openai/v1/*`
- Go 模块名简化：`github.com/user/ai-model-proxy` → `ai-model-proxy`
- 依赖更新：使用 `gin-contrib/cors` v1.7.6 和更新版 `gin-contrib/sessions`
- **NEW** 构建系统：添加 Makefile 构建脚本，支持版本注入
- **NEW** 静态文件嵌入：使用 Go embed 将前端嵌入二进制文件
- **NEW** VS Code 调试配置
- **FIX** 国际化切换无效：修复 i18n locale 未同步更新的问题
- **FIX** 中文翻译缺失：补全 zh.ts 中文翻译内容

## Capabilities

### New Capabilities

- `env-config`: 环境变量配置系统，支持 AMP_ 前缀环境变量，提供默认值回退机制

### Modified Capabilities

<!-- 无需求变更，仅为实现层面的重构 -->

## Impact

**目录结构变更：**
```
ai-model-proxy/
├── web/                    # Vue 3 前端
│   ├── src/
│   ├── package.json
│   ├── vite.config.ts
│   └── Makefile            # 前端构建脚本
├── server/                 # Go 后端
│   ├── cmd/
│   ├── internal/
│   ├── res/                # 嵌入资源
│   │   ├── res.go          # embed 定义 + 版本信息
│   │   └── web/            # 前端构建输出 (嵌入)
│   ├── go.mod
│   └── Makefile            # 后端构建脚本
├── Makefile                # 统一构建入口
├── .vscode/                # VS Code 调试配置
│   └── launch.json
├── README.md
└── openspec/
```

**API 路径变更：**
| 原路径 | 新路径 |
|--------|--------|
| `/v1/chat/completions` | `/openai/v1/chat/completions` |
| `/v1/models` | `/openai/v1/models` |
| `/api/v1/*` | 保持不变 |

**环境变量：**
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `AMP_SERVER_PORT` | 18080 | 服务端口 |
| `AMP_SERVER_MODE` | debug | 运行模式 |
| `AMP_DATABASE_PATH` | data.db | 数据库路径 |
| `AMP_SESSION_SECRET` | (自动生成) | Session 密钥，未设置时自动生成 |
| `AMP_SESSION_MAX_AGE` | 86400 | Session 有效期 |
| `AMP_ADMIN_USERNAME` | admin | 默认管理员用户名 |
| `AMP_ADMIN_PASSWORD` | admin | 默认管理员密码 |

**依赖变更：**
- 新增 `github.com/gin-contrib/cors` v1.7.6
- 更新 `github.com/gin-contrib/sessions`
