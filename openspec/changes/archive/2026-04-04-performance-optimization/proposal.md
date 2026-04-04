## Why

当前系统存在几个性能相关的问题：

1. **流处理错误处理不足**：非 EOF 错误会导致无限循环，消耗资源
2. **数据库连接池未配置**：默认配置可能导致连接泄漏或资源耗尽
3. **调试配置过于粗粒度**：单一 `debug.enabled` 无法精细控制 Gin/GORM 日志输出
4. **缺乏运行时性能分析工具**：生产环境问题难以诊断
5. **配置冗余**：`server.mode` 配置项在实际使用中无意义

这些问题在生产环境可能导致性能瓶颈和稳定性风险。

## What Changes

### 新增功能

- **Pprof 性能分析端点**：独立的 pprof 服务器，默认端口 6060，可通过 `pprof.port` 配置
- **数据库连接池配置**：新增 `MaxOpenConns`、`MaxIdleConns`、`ConnMaxLifetime`、`ConnMaxIdleTime` 配置项
- **精细调试控制**：拆分 `debug` 配置为 `debug.gin`、`debug.gorm`、`debug.provider` 三个独立开关

### 修改功能

- **流错误处理增强**：所有 Provider 流处理添加错误计数，连续错误 ≥3 次安全退出
- **调试日志简化**：移除 `debug.go` 中不必要的逐行缓冲逻辑

### 移除功能

- **BREAKING**：移除 `server.mode` 配置项（该配置无实际作用，仅部分文档提及）

## Capabilities

### New Capabilities

- `pprof-server`: 运行时性能分析服务器配置与启动
- `database-connection-pool`: 数据库连接池参数配置

### Modified Capabilities

- `yaml-config`: 配置结构变更（新增 pprof、数据库连接池、拆分 debug、session 移入 server、移除 server.mode）

## Impact

```
┌─────────────────────────────────────────────────────────────────┐
│                      影响范围                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  config/config.go      配置结构重构                              │
│  cmd/server/main.go    pprof 服务器启动                          │
│  model/db.go           连接池初始化                              │
│  provider/debug.go     调试日志简化                              │
│  provider/provider_*   流错误处理                                │
│  config.yaml           配置文件格式变更                           │
│  文档                  配置说明更新                               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**配置迁移影响**：
- 使用 `server.mode` 的配置文件需移除该字段（无功能影响）
- `session` 配置需移至 `server.session` 下
- Session 相关环境变量名变更：`AG_SESSION_*` → `AG_SERVER_SESSION_*`
- TrustedProxies 环境变量名变更：`AG_TRUSTED_PROXIES` → `AG_SERVER_TRUSTED_PROXIES`
- 新增 `pprof`、`database.pool`、拆分 `debug` 配置均为可选，有默认值