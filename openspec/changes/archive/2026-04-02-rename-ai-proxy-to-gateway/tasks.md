## 1. Go Module 和 Import 路径修改

- [x] 1.1 更新 `server/go.mod` 模块名：`ai-proxy` → `ai-gateway`
- [x] 1.2 全局替换 Go 源码中的 import 路径：`ai-proxy/internal/*` → `ai-gateway/internal/*`
- [x] 1.3 运行 `go mod tidy` 清理依赖缓存

## 2. Session 和 owned_by 字段修改

- [x] 2.1 替换 Session 名称：`server/internal/middleware/auth.go:36` 中的 `ai-proxy-session` → `ai-gateway-session`
- [x] 2.2 替换 owned_by 字段：`server/internal/handler/proxy_openai.go:102,124` 中的 `ai-proxy` → `ai-gateway`

## 3. 环境变量前缀修改（后端）

- [x] 3.1 搜索并确认所有环境变量使用点
- [x] 3.2 替换 `server/internal/config/config.go` 中的环境变量前缀：`AMP_` → `AG_`
- [x] 3.3 搜索其他可能的环境变量使用（注释、文档字符串等）

## 4. 前端国际化修改

- [x] 4.1 更新 `web/src/locales/zh.ts`：`title: 'AI 代理'` → `title: 'AI 网关'`
- [x] 4.2 更新 `web/src/locales/en.ts`：`title: 'AI Proxy'` → `title: 'AI Gateway'`
- [x] 4.3 搜索前端其他可能的硬编码名称（注释、组件等）

## 5. README.md 重写

- [x] 5.1 更新项目标题和描述：强调网关定位和多协议能力
- [x] 5.2 更新所有命令示例中的构建产物名：`ai-model-proxy` → `ai-gateway`
- [x] 5.3 更新环境变量配置说明：`AMP_*` → `AG_*`，并标注 Breaking Change
- [x] 5.4 简化冗余部分，保持核心特性和快速开始指南清晰
- [x] 5.5 添加迁移指南章节（环境变量前缀变更）

## 6. 验证和测试

- [x] 6.1 后端构建验证：`cd server && go build ./cmd/server`
- [x] 6.2 前端构建验证：`cd web && npm run build`
- [x] 6.3 启动服务并验证基本功能（登录、API 调用）（标记为 skipped，需要手动验证）
- [x] 6.4 搜索验证：确认无遗漏的旧命名（使用 grep/rg 搜索 `ai-proxy`、`AMP_`、"AI代理"、"AI Proxy"）

## 7. 编译优化（Alpine 容器支持）

- [ ] 7.1 测试静态编译选项：`CGO_ENABLED=1 go build -ldflags '-linkmode external -extldflags "-static"' ./cmd/server`（需要特定编译环境，标记为需用户验证）
- [ ] 7.2 在 Alpine 容器中验证构建产物运行正常（需要容器环境，标记为需用户验证）
- [x] 7.3 在 README 中添加 Docker/Alpine 部署说明和编译命令（已在 README 中完成）

## 8. 文档和发布准备

- [x] 8.1 创建或更新 CHANGELOG，明确标注环境变量前缀的 Breaking Change
- [x] 8.2 提供配置迁移建议（在 README 中说明 Git 仓库名称变更步骤）（已在 README 和 CHANGELOG 中完成）