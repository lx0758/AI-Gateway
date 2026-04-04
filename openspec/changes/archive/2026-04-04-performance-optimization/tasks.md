## 1. 配置结构变更

- [x] 1.1 新增 `PprofConfig` 结构体（Port 字段，默认 6060）
- [x] 1.2 新增 `PoolConfig` 结构体（MaxOpen, MaxIdle, MaxLifetime, MaxIdleTime）
- [x] 1.3 拆分 `DebugConfig`（移除 Enabled，新增 Gin/Gorm bool 字段）
- [x] 1.4 移除 `ServerConfig.Mode` 字段（兼容处理：忽略并警告）
- [x] 1.5 更新 `Load()` 函数解析新增配置项
- [x] 1.6 新增环境变量支持（AG_PPROF_PORT, AG_DEBUG_GIN, AG_DEBUG_GORM, AG_DATABASE_POOL_*）
- [x] 1.7 更新配置日志输出

## 2. Pprof 服务器实现

- [x] 2.1 在 main.go 中导入 net/http/pprof
- [x] 2.2 实现独立 pprof 服务器启动（goroutine）
- [x] 2.3 绑定 localhost，使用配置端口
- [x] 2.4 添加启动失败日志处理（不影响主服务）

## 3. 数据库连接池实现

- [x] 3.1 在 InitDB 中获取 sql.DB 实例
- [x] 3.2 设置连接池参数（从配置读取）
- [x] 3.3 SQLite 特殊约束处理（强制 MaxOpen=1, MaxIdle=1）
- [x] 3.4 添加连接池配置日志

## 4. 调试日志重构

- [x] 4.1 简化 debug.go 中的 debugReader/debugWriter（移除逐行缓冲）
- [x] 4.2 移除 sync.Mutex（简化为直接写入）
- [x] 4.3 更新 SetDebugMode 调用逻辑（使用 Gin/Gorm 配置）
- [x] 4.4 Gin 框架 debug/release 模式控制
- [x] 4.5 GORM 日志级别控制

## 5. 流错误处理增强

- [x] 5.1 provider_anthropic.go: streamAnthropicToOpenAI 添加错误计数
- [x] 5.2 provider_anthropic.go: copyAnthropicStreaming 添加错误计数
- [x] 5.3 provider_openai.go: copyOpenAIStreaming 添加错误计数
- [x] 5.4 provider_openai.go: streamOpenAIToAnthropic 添加错误计数
- [x] 5.5 连续错误 ≥3 次时记录日志并退出循环

## 6. 文档更新

- [x] 6.1 更新 config.yaml.example 示例文件
- [x] 6.2 更新 README 配置说明
- [ ] 6.3 添加性能调优指南文档（可选）