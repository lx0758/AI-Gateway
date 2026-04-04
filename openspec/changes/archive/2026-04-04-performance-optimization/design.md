## Context

当前 AI Gateway 服务在生产环境面临以下挑战：

1. **性能诊断困难**：缺乏运行时性能分析工具，CPU/内存问题难以定位
2. **数据库连接管理不当**：默认连接池配置可能导致连接泄漏或资源耗尽
3. **调试日志过于冗余**：单一开关无法精细控制不同组件的日志输出
4. **流处理健壮性不足**：网络异常可能导致无限循环

现有配置结构：
```
Config {
  Debug { Enabled }          // 粒度过粗
  Server { Port, Mode, ... } // Mode 无实际作用
  Database { Type, Path, ... } // 缺少连接池参数
  Session { ... }            // 与 Server 关联更强
}
```

目标配置结构：
```
Config {
  Debug { Gin, Gorm, Provider }  // 精细化控制
  Pprof { Port }
  Server {
    Port, TrustedProxies
    Session { ... }              // 移入 Server 下
  }
  Database { Type, Path, Pool { ... } }
  Auth { ... }
}
```

## Goals / Non-Goals

**Goals:**
- 提供独立的 pprof 性能分析端点，便于生产诊断
- 配置化数据库连接池参数，适应不同负载场景
- 精细化调试日志控制，减少生产噪音
- 增强流处理错误恢复能力
- 清理冗余配置项

**Non-Goals:**
- 不引入新的外部依赖
- 不改变现有 API 行为
- 不实现自动性能调优

## Decisions

### D1: Pprof 端口策略

**决策**: 使用独立端口（默认 6060），而非在主服务上挂载端点

**理由**:
- 主服务端口暴露给外部，pprof 端点不应被外部访问
- 独立端口便于防火墙隔离，仅允许内网访问
- 默认 6060 是 pprof 社区惯例，降低认知负担

**替代方案**:
- 在主服务挂载 `/debug/pprof` → 拒绝：暴露安全风险
- 动态端口分配 → 拒绝：增加运维复杂度

### D2: 数据库连接池默认值

**决策**: 
```
Postgres/MySQL:
  MaxOpenConns = 20
  MaxIdleConns = 5
  ConnMaxLifetime = 1h
  ConnMaxIdleTime = 5m

SQLite:
  MaxOpenConns = 1
  MaxIdleConns = 1
  (SQLite 单连接限制)
```

**理由**:
- 20 连接适合中等负载，可配置调整
- 1h Lifetime 防止长期连接积累问题
- SQLite 单连接是 Go SQLite 驱动的最佳实践

### D3: 调试配置拆分

**决策**: 拆分为 `debug.gin`、`debug.gorm`、`debug.provider` 三个独立开关

**理由**:
- Gin、GORM、Provider 是独立组件，日志用途不同
- Gin 日志主要用于请求追踪
- GORM 日志主要用于 SQL 调试
- Provider 日志主要用于 API 请求/响应分析
- 生产环境可能只需要其中一个或两个

**配置结构**:
```yaml
debug:
  gin: false      # Gin 框架调试日志
  gorm: false     # GORM SQL 调试日志
  provider: false # Provider 请求/响应记录
```

### D4: 流错误处理阈值

**决策**: 连续错误 ≥3 次时终止流处理

**理由**:
- 单次错误可能是临时抖动，应容错
- 连续 3 次表明链路已断，继续无意义
- 记录日志便于事后分析

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| pprof 端口被外部访问 | 默认绑定 localhost，文档说明防火墙配置 |
| 连接池参数不当导致性能问题 | 提供合理默认值，文档说明调优建议 |
| 移除 server.mode 影响现有配置 | 兼容处理：忽略该字段，输出警告日志 |
| session 配置路径变更 | 文档明确迁移路径，环境变量名同步变更 |
| 环境变量名变更 | 新命名更清晰（AG_SERVER_*），文档说明迁移 |
| 流处理提前终止丢失数据 | 记录错误日志，客户端可重试 |

## Migration Plan

```
Phase 1: 配置结构变更
├── 新增 PprofConfig、PoolConfig
├── 拆分 DebugConfig (gin/gorm/provider)
├── Session 移入 ServerConfig
├── 移除 ServerConfig.Mode (兼容处理)
└── 环境变量重命名 (AG_SESSION_* → AG_SERVER_SESSION_*, AG_TRUSTED_PROXIES → AG_SERVER_TRUSTED_PROXIES)

Phase 2: 功能实现
├── pprof 服务器启动
├── 连接池初始化
├── 流错误处理增强
└── 调试日志简化

Phase 3: 文档更新
├── 配置说明文档
└── 性能调优指南
```

**Rollback**: 配置变更均为可选，旧配置文件需调整结构（session 路径、环境变量名）

## Open Questions

- 是否需要支持 pprof 认证？当前建议内网隔离，暂不实现
- 连接池监控指标是否需要暴露？可在后续版本考虑