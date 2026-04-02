## Context

AI Proxy 项目当前处于稳定运行状态，已实现 OpenAI 兼容 API 代理、多厂商支持、负载均衡、API Key 管理等核心功能。项目采用前后端分离架构：
- **后端**: Go + Gin + Gorm，使用 SQLite 存储
- **前端**: Vue 3 + Element Plus + Pinia + vue-i18n
- **命名约定**: Go module 使用 `ai-proxy`，环境变量使用 `AMP_` 前缀

本次变更旨在将项目名称从 "AI代理/AI Proxy" 更改为 "AI网关/AI Gateway"，以更好地反映项目的定位演进——从单纯的 API 代理扩展为支持多种 AI 协议（OpenAI、Anthropic、MCP 等）的综合网关。

### 现状分析

通过代码搜索，发现以下命名使用点：
- Go module: `server/go.mod` 定义为 `ai-proxy`
- Go import: 所有 `.go` 文件使用 `ai-proxy/internal/*` 路径（31+ 处）
- Session name: `ai-proxy-session` (1 处)
- owned_by 字段: `ai-proxy` (2 处)
- 前端 i18n: zh.ts 有 "AI 代理"，en.ts 有 "AI Proxy" (各 2 处)
- README.md: 多处使用 "AI Proxy"、"ai-model-proxy" 等名称
- 环境变量: 使用 `AMP_` 前缀（需确认代码中的实际使用）

## Goals / Non-Goals

**Goals:**
- 完成所有命名标识符的全局替换，确保代码一致性
- 保持向后兼容，避免破坏性变更（环境变量前缀变更除外）
- 重新整理 README.md，突出网关定位和未来方向
- 确保所有测试通过，构建正常

**Non-Goals:**
- 不修改 `openspec/changes/archive/` 下的历史文档（保持历史记录）
- 不修改 Git 仓库名（需用户手动操作）
- 不修改 package.json 的项目名（前端项目名保持为 "web"）
- 不新增或修改功能代码

## Decisions

### 1. Go Module 命名选择

**决定**: 使用 `ai-gateway`

**理由**:
- 简洁明了，符合 Go module 命名习惯
- 与项目新定位一致
- 保持小写、无连字符（除分隔单词）的风格

**替代方案**:
- `ai-gw`: 过于简短，语义不够清晰
- `ai-service-gateway`: 过长，不符合简洁原则
- 保持 `ai-proxy`: 与新定位不符

### 2. 环境变量前缀选择

**决定**: 使用 `AG_` (AI Gateway)

**理由**:
- 极简且唯一，避免与常见系统变量冲突
- 语义清晰，便于理解配置归属
- 更简洁，减少配置时的输入负担

**替代方案**:
- `AGW_`: 三个字母，稍长
- `AI_GATEWAY_`: 过长，配置时不便
- `GW_`: 过于简短，语义不够明确
- 保持 `AMP_`: 与新定位不符

### 3. 实施策略选择

**决定**: 一次性全局替换，而非渐进式替换

**理由**:
- 本次变更不涉及功能修改，风险较低
- 全局替换可确保一致性，避免遗漏
- 单次变更便于测试和验证

**替代方案**:
- 分阶段替换（先改后端，再改前端）: 会产生中间不一致状态，增加混淆
- 只改外显名称，不改内部标识符: 会导致内外不一致，影响维护

### 4. README.md 重写策略

**决定**: 保持核心内容，更新名称，突出网关定位，简化冗余部分

**理由**:
- README 是项目门面，需准确反映定位
- 保留技术细节和配置说明，便于用户快速上手
- 强调网关的多协议能力和未来扩展方向

### 5. 兼容性处理

**决定**: 环境变量前缀变更视为 **BREAKING** 变化，需要在文档中明确说明

**理由**:
- 环境变量是用户配置的关键部分
- 前缀变更会影响所有现有配置
- 无法保持向后兼容（不支持同时接受新旧前缀）

**缓解措施**:
- 在 README 中提供迁移指南
- 在 CHANGELOG 或 release note 中明确标注此 breaking change
- 提供配置迁移脚本（可选）

### 6. 编译选项优化（Alpine 容器支持）

**决定**: 添加静态编译选项，确保构建产物可在 Alpine 容器中运行

**理由**:
- Alpine 使用 musl libc，而非 glibc，默认 Go 编译可能因依赖动态库而无法运行
- SQLite 驱动 (`github.com/mattn/go-sqlite3`) 使用 CGO，需特殊处理
- 静态编译可消除依赖问题，提高部署灵活性

**实施方案**:
- 编译命令：`CGO_ENABLED=1 go build -ldflags '-linkmode external -extldflags "-static"' ./cmd/server`
- 或使用 `go build -tags netgo` 禁用 net 包的 CGO
- README 中添加 Docker/Alpine 部署说明

**替代方案**:
- 不优化编译：需要用户在 Alpine 中安装 gcompat 或使用 glibc 版本镜像
- 禁用 CGO：SQLite 需要纯 Go 实现的驱动（如 `modernc.org/sqlite`），需替换依赖

## Risks / Trade-offs

### Risk 1: Go import 路径变更可能导致 IDE/工具索引失效

**风险**: Go module 重命名后，IDE (如 GoLand, VS Code) 可能需要重新索引，现有缓存失效

**缓解措施**:
- 提醒用户在修改后清理 IDE 缓存或重新打开项目
- 使用 `go mod tidy` 清理依赖缓存

### Risk 2: 环境变量前缀变更影响现有部署

**风险**: 已部署的系统使用 `AMP_` 前缀，变更后配置失效

**缓解措施**:
- 提供配置迁移指南
- 在版本发布时明确标注 breaking change
- 建议用户提供回退方案（保留旧配置文件）

### Risk 3: 部分命名点遗漏

**风险**: 搜索可能遗漏某些命名点（如注释、测试文件等）

**缓解措施**:
- 使用多种搜索工具（grep, rg）交叉验证
- 实施后进行构建和测试，验证完整性
- 审查所有已修改文件，确保一致性

### Risk 4: Git 仓库名称未同步

**风险**: 代码名称与 Git 仓库名称不一致，造成混淆

**缓解措施**:
- 在 README 中说明 Git 仓库名称变更建议
- 提供仓库重命名步骤（GitHub 设置）

## Migration Plan

### 阶段 1: 代码修改（自动化）

1. **后端修改**:
   - 更新 `server/go.mod` 模块名
   - 执行全局替换所有 import 路径
   - 替换 Session 名称
   - 替换 owned_by 字段
   - 替换环境变量前缀（如果代码中有硬编码）

2. **前端修改**:
   - 更新 `web/src/locales/zh.ts` 和 `en.ts` 的标题
   - 搜索并替换其他可能的硬编码名称（如注释）

3. **文档修改**:
   - 重写 `README.md`
   - 更新构建命令和示例

### 阶段 2: 验证测试

1. 运行 `go mod tidy` 清理依赖
2. 运行后端构建：`go build ./cmd/server`
3. 运行前端构建：`npm run build`
4. 启动服务并验证基本功能

### 阶段 3: 部署准备

1. 准备配置迁移指南（环境变量前缀）
2. 更新 CHANGELOG 或 release note
3. 建议 Git 仓库名称变更（用户手动操作）

### Rollback Strategy

由于本次变更不涉及数据或配置文件修改，回退策略简单：
- Git revert 本次变更的所有 commit
- 用户恢复原有环境变量配置

## Open Questions

1. **是否需要提供环境变量迁移脚本？**
   - 简单的 shell script 可帮助用户批量替换配置
   - 或在 README 中提供手动迁移步骤

2. **是否需要在代码中添加兼容层？**
   - 同时接受新旧环境变量前缀（临时兼容）
   - 决定：不添加，保持简洁，明确标注 breaking change

3. **是否需要创建 CHANGELOG 或 release note？**
   - 建议创建，明确标注此 breaking change 和迁移步骤