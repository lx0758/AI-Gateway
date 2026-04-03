## 1. 后端配置

- [x] 1.1 在 config/config.go 新增 TrustedProxies 配置项，支持环境变量 AG_TRUSTED_PROXIES
- [x] 1.2 在 cmd/server/main.go 配置 Gin 的 TrustedProxies，默认信任内网 CIDR

## 2. 数据模型扩展

- [x] 2.1 在 model/db.go 的 UsageLog 结构体新增 client_ips 字段（合并存储IP链，逗号分隔）
- [x] 2.2 更新 UsageLog.String() 方法包含 IP 信息

## 3. IP获取逻辑

- [x] 3.1 创建 utils/ip.go 工具函数封装 IP 获取逻辑（GetClientIPInfo，合并IP链）
- [x] 3.2 在 handler/proxy_openai.go 的 ChatCompletions() 中获取 IP 信息
- [x] 3.3 在 handler/proxy_anthropic.go 的 Messages() 中获取 IP 信息

## 4. UsageLog创建逻辑

- [x] 4.1 在 handler/usage.go 的 NewUsageLog() 函数新增 clientIPs 参数
- [x] 4.2 在 handler/usage.go 的 logsResponse 结构新增 client_ips 字段
- [x] 4.3 在 handler/usage.go 的 Logs() 方法中返回 IP 字段

## 5. 前端类型定义

- [x] 5.1 在 web/src/views/Usage/index.vue 的 LogItem 接口新增 client_ips 字段

## 6. 前端日志列表

- [x] 6.1 在日志列表中 Key 列后新增 IPs 列（width=120）
- [x] 6.2 只显示首个 IP，完整链用 tooltip 显示（使用 InfoFilled 图标）
- [x] 6.3 无转发链时不显示图标

## 7. 前端IP统计

- [x] 7.1 新增 ipStats computed 属性聚合 IP 统计数据（按首个IP统计）
- [x] 7.2 新增 IP 统计卡片（在 Key 统计卡片后）
- [x] 7.3 实现 IP 统计表格显示（首个IP、调用次数、Tokens、平均耗时）
- [x] 7.4 只显示首个 IP，完整链用 tooltip 显示

## 8. 国际化（可选）

- [x] 8.1 如果存在 locales 文件，新增翻译键：usage.clientIp, usage.ipStats（未找到locales文件，无需添加）

## 9. 测试验证

- [ ] 9.1 测试直连场景：验证 ClientIP() 返回真实 IP
- [ ] 9.2 测试代理场景：验证 X-Forwarded-For 解析正确
- [ ] 9.3 测试安全配置：验证 TrustedProxies 防止 IP 伪造
- [ ] 9.4 测试前端展示：验证 IP 列和统计卡片显示正确
- [ ] 9.5 测试环境变量配置：验证 AG_TRUSTED_PROXIES 生效

## 10. 文档更新

- [x] 10.1 在 README.md 环境变量表新增 AG_TRUSTED_PROXIES 说明
- [x] 10.2 在 README.md 新增部署场景 IP 配置示例（直连、Nginx、Cloudflare、ALB）