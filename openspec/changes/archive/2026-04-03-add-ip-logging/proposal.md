## Why

当前系统缺少调用者IP信息记录，无法追踪请求来源、识别异常调用模式或进行安全审计。在生产环境中，IP信息对于问题诊断、滥用检测和合规审计至关重要。

## What Changes

- UsageLog数据模型新增`client_ips`字段，记录客户端IP和完整转发链（逗号分隔）
- 前端日志列表新增IP列，显示完整IP链信息
- 前端新增IP统计卡片，聚合展示所有IP的调用情况（按首个IP统计）
- Gin配置TrustedProxies防止IP伪造攻击，默认信任内网CIDR范围
- 支持通过环境变量`AG_TRUSTED_PROXIES`自定义信任代理配置

## Capabilities

### New Capabilities

- `ip-logging`: 客户端IP地址记录能力，包括原始IP获取、转发链追踪（合并存储）、安全防护配置

### Modified Capabilities

- `usage-tracking`: 用量追踪能力，UsageLog数据结构扩展（新增client_ips字段）

## Impact

### 后端影响

- `server/internal/model/db.go`: UsageLog结构体新增字段，String()方法更新
- `server/internal/config/config.go`: 新增TrustedProxies配置项
- `server/cmd/server/main.go`: Gin引擎配置TrustedProxies
- `server/internal/utils/ip.go`: 新增GetClientIPInfo工具函数，合并IP链信息
- `server/internal/handler/usage.go`: NewUsageLog()函数新增参数，logsResponse结构扩展
- `server/internal/handler/proxy_openai.go`: ChatCompletions()传递IP信息
- `server/internal/handler/proxy_anthropic.go`: Messages()传递IP信息

### 前端影响

- `web/src/views/Usage/index.vue`: LogItem接口扩展，新增IP列和IP统计卡片
- `web/src/locales/`: 新增翻译键（如果存在国际化文件）

### 数据库影响

- SQLite数据库自动迁移：UsageLog表新增client_ips列

### 部署影响

- 新增环境变量：`AG_TRUSTED_PROXIES`（可选）
- 数据库存储增长：每条日志增加约50字节（IP+转发链）