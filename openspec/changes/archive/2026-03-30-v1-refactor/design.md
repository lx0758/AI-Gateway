## Context

AI Model Proxy 项目完成初始开发，进入 v1 稳定版本阶段。当前项目结构将前后端代码混合存放，配置依赖 YAML 文件，API 路径设计不够清晰。本次重构旨在解决这些问题，为后续维护和部署打下良好基础。

**当前状态：**
```
ai-model-proxy/
├── cmd/server/main.go
├── internal/
├── web/                    # 前端代码
├── configs/config.yaml     # YAML 配置
└── go.mod                  # module github.com/user/ai-model-proxy
```

**约束：**
- 保持向后兼容的管理 API (`/api/v1/*`)
- 数据库文件兼容，无需迁移
- 现有功能不变更

## Goals / Non-Goals

**Goals:**
- 清晰的项目结构：前端 `web/`，后端 `server/`
- 环境变量配置：支持 `AMP_` 前缀，YAML 值作为默认值
- 明确的 API 路径：代理 API 使用 `/openai/v1/*` 前缀
- 简洁的模块名：`ai-model-proxy`
- 更新依赖到稳定版本

**Non-Goals:**
- 不新增功能
- 不修改数据库 schema
- 不变更管理 API 路径

## Decisions

### D1: 目录结构

**决定：** 前端放入 `web/`，后端放入 `server/`

```
ai-model-proxy/
├── web/                    # Vue 3 前端
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── server/                 # Go 后端
│   ├── cmd/
│   ├── internal/
│   ├── configs/
│   └── go.mod
├── README.md
└── openspec/
```

**理由：**
- 职责分离清晰，便于独立开发
- 符合常见项目结构约定
- 便于 CI/CD 分别构建

**备选方案：**
- `frontend/` + `backend/`：命名较长，不如 `web/` + `server/` 简洁

### D2: 配置方式

**决定：** 移除 YAML 配置，改用环境变量

```go
// 默认值定义
var defaults = map[string]string{
    "AMP_SERVER_PORT":     "18080",
    "AMP_SERVER_MODE":     "debug",
    "AMP_DATABASE_PATH":   "data.db",
    "AMP_SESSION_MAX_AGE": "86400",
    "AMP_ADMIN_USERNAME":  "admin",
    "AMP_ADMIN_PASSWORD":  "admin",
}

func getEnv(key string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaults[key]
}

func getSessionSecret() string {
    if val := os.Getenv("AMP_SESSION_SECRET"); val != "" {
        return val
    }
    // 自动生成 32 字节随机密钥
    b := make([]byte, 32)
    rand.Read(b)
    return base64.StdEncoding.EncodeToString(b)
}
```

**理由：**
- 12-Factor App 最佳实践
- 容器化部署友好
- 敏感信息不落盘
- 配置来源统一

**备选方案：**
- 保留 YAML + 环境变量覆盖：增加复杂度，环境变量已足够
- 使用 Viper 库：引入额外依赖，当前需求简单

### D3: API 路径

**决定：** 代理 API 添加 `/openai` 前缀

| 功能 | 路径 |
|------|------|
| 聊天补全 | `/openai/v1/chat/completions` |
| 模型列表 | `/openai/v1/models` |
| 管理接口 | `/api/v1/*` (不变) |

**理由：**
- 明确区分代理 API 和管理 API
- 便于未来支持其他厂商（如 `/anthropic/v1/*`）
- 避免路径冲突

### D4: Go 模块名

**决定：** 使用简短名称 `ai-model-proxy`

**理由：**
- 简洁，便于 import
- 不绑定特定代码托管平台
- 符合 Go 模块命名趋势

### D5: 依赖更新

**决定：** 使用 `gin-contrib/cors` 替换自定义 CORS 中间件

**理由：**
- 社区维护，功能完善
- 支持动态配置
- 减少自定义代码

### D6: 构建系统

**决定：** 使用 Makefile 统一构建

```
Makefile (根目录)
├── make build   # 构建前后端
├── make dev     # 启动开发服务器
└── make clean   # 清理构建产物

server/Makefile
├── make build   # go build -ldflags 版本注入
├── make dev     # go run
└── make test    # go test

web/Makefile
├── make build   # npm run build → server/res/web/
└── make dev     # npm run dev
```

**理由：**
- 统一构建入口，简化 CI/CD
- 支持版本信息注入
- 开发体验一致

### D7: 静态文件嵌入

**决定：** 使用 Go embed 将前端嵌入二进制文件

```go
// server/res/res.go
var Version = "dev"  // 构建时注入

//go:embed all:web/*
var webEmbedFS embed.FS
var WebFS, _ = fs.Sub(webEmbedFS, "web")
```

**理由：**
- 单一二进制文件部署
- 无需外部静态文件
- 版本信息可追踪

**备选方案：**
- 保留外部静态文件：部署复杂，需要额外配置

## Risks / Trade-offs

### R1: 路径变更影响现有客户端

**风险：** 现有客户端使用 `/v1/*` 路径，变更后无法访问

**缓解：**
- 文档明确标注 BREAKING CHANGE
- 提供迁移指南
- 可考虑过渡期保留旧路径并返回弃用警告

### R2: 环境变量遗漏

**风险：** 部署时遗漏必要环境变量

**缓解：**
- 所有配置项有默认值
- 启动时打印配置状态（敏感值脱敏）
- README 提供完整配置清单

### R3: 目录移动影响开发习惯

**风险：** 团队成员需要适应新结构

**缓解：**
- 结构直观，学习成本低
- README 更新说明

## Migration Plan

1. **备份**：保留当前代码
2. **后端重构**：
   - 创建 `server/` 目录
   - 移动 Go 代码
   - 更新 `go.mod` 模块名
   - 替换配置加载逻辑
   - 更新 API 路由
   - 替换 CORS 中间件
3. **前端重构**：
   - 保持 `web/` 位置不变
   - 更新 `vite.config.ts` 代理路径
4. **验证**：运行测试，确认功能正常
5. **文档更新**：更新 README

## Open Questions

- [ ] 是否需要过渡期支持旧 API 路径？
- [ ] 是否需要配置验证（必填项检查）？
