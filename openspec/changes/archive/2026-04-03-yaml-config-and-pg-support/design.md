## Context

AI Gateway 是一个基于 Go + Gin + Gorm 的 AI 模型代理网关服务，当前处于活跃开发阶段。项目特点：

- **配置现状**: 所有配置通过环境变量加载（`AG_*` 前缀），配置项包括服务器、数据库、会话、认证等
- **数据库现状**: 仅支持 SQLite，通过 `gorm.io/driver/sqlite` 连接，数据文件默认为 `data.db`
- **架构特点**: 采用典型的三层架构（Handler/Service/Model），使用 Gorm ORM

约束：
- 现有环境变量命名和功能不能改变（向后兼容）
- 未来新增配置参数必须同时支持 YAML 和环境变量
- 数据库切换不能破坏现有数据和功能

## Goals / Non-Goals

**Goals:**
- 实现 YAML 配置文件加载机制，与环境变量并存
- 环境变量优先级高于 YAML，支持临时覆盖
- 支持 PostgreSQL 作为可选数据库
- 提供平滑的迁移路径（从 SQLite 到 PostgreSQL）
- 建立配置参数规范，确保未来参数遵循双重支持原则

**Non-Goals:**
- 不改变现有 API 接口和业务逻辑
- 不引入其他数据库类型（如 MySQL、MongoDB）
- 不实现配置热加载或动态配置更新
- 不改变现有的环境变量命名规范

## Decisions

### 1. YAML 配置文件位置和命名

**决策**: 使用 `config.yaml` 作为默认且固定的配置文件名，位于项目根目录（server/ 目录下）

**理由**:
- 符合 Go 项目惯例，简单直观
- 固定路径减少配置复杂度，便于查找和管理
- 不引入环境变量路径覆盖，保持配置文件位置的一致性

**替代方案**:
- 使用 `.env.yaml` - 不够直观
- 多层级配置（`/etc/`, `~/.config/`）- 过度设计
- 通过环境变量指定路径 - 增加运维复杂度，已弃用

### 2. 配置加载优先级策略

**决策**: 环境变量 > YAML 文件 > 默认值

**理由**:
- 环境变量是 Docker/K8s 环境的标准配置方式
- 便于运维人员临时覆盖配置（无需修改文件）
- YAML 文件适合静态配置（数据库连接、服务器端口等）
- 保持现有环境变量机制的兼容性

**实现方式**:
```go
// 先加载 YAML
cfg = loadYAML(configPath)

// 再用环境变量覆盖
cfg.Server.Port = getEnv("AG_SERVER_PORT", cfg.Server.Port)
cfg.Database.Type = getEnv("AG_DATABASE_TYPE", cfg.Database.Type)
```

### 3. YAML 配置结构设计

**决策**: YAML 结构与环境变量命名对应，使用嵌套结构

```yaml
debug:
  enabled: false        # 新增：调试模式开关

server:
  port: 18080
  mode: debug

database:
  type: sqlite          # 新增：sqlite 或 postgres
  path: data.db         # sqlite 专用
  host: localhost       # postgres 专用
  port: 5432            # postgres 专用
  user: postgres        # postgres 专用
  password: ""          # postgres 专用
  name: ai_gateway      # postgres 专用

session:
  secret: ""
  max_age: 86400
  secure: false
  http_only: true
  same_site: lax

auth:
  default_admin:
    username: admin
    password: admin
```

**理由**:
- 结构清晰，易于维护
- 与现有环境变量分组一致
- 便于未来扩展新配置项

### 4. PostgreSQL 连接实现

**决策**: 使用 `gorm.io/driver/postgres` 作为 PostgreSQL 驱动

**理由**:
- Gorm 官方支持的 PostgreSQL 驱动
- 与现有 SQLite 驱动接口一致（`gorm.Dialector`）
- 社区广泛使用，文档完善

**连接方式**:
```go
// SQLite
dsn := fmt.Sprintf("%s?_loc=auto", dbPath)
DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})

// PostgreSQL
dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
    host, port, user, password, dbname)
DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

### 5. 数据库切换机制

**决策**: 在 `model/db.go` 的 `InitDB` 函数中根据 `database.type` 选择驱动

**理由**:
- 最小化代码改动
- 保持统一的数据库初始化入口
- 便于测试和维护

**实现方式**:
```go
func InitDB(cfg *config.DatabaseConfig) error {
    var dialector gorm.Dialector
    
    switch cfg.Type {
    case "sqlite":
        dsn := cfg.Path + "?_loc=auto"
        dialector = sqlite.Open(dsn)
    case "postgres":
        dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
            cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
        dialector = postgres.Open(dsn)
    default:
        return fmt.Errorf("unsupported database type: %s", cfg.Type)
    }
    
    DB, err = gorm.Open(dialector, &gorm.Config{})
    // ...
}
```

### 6. 配置参数扩展规范

**决策**: 建立明确的配置参数扩展规范，强制要求双重支持

**规范内容**:
- 新增配置参数必须在 `Config` 结构体中定义字段
- YAML 结构必须包含对应配置项
- 必须实现 `getEnv()` 函数调用以支持环境变量覆盖
- 文档中必须同时说明 YAML 和环境变量两种配置方式

**理由**:
- 防止未来参数仅支持单一配置方式
- 确保配置方式的统一性和一致性
- 降低运维复杂度

### 7. 调试模式配置（debug.enabled）

**决策**: 新增 `debug.enabled` 配置参数，控制调试模式开关

**用途**:
- 开启调试模式时，启用详细日志输出
- 开启调试模式时，禁用 Gin 框架的性能优化（便于调试）
- 关闭调试模式时，使用生产环境配置（静默日志、性能优化）

**配置方式**:
- YAML: `debug.enabled: true` 或 `debug.enabled: false`
- 环境变量: `AG_DEBUG_ENABLED=true` 或 `AG_DEBUG_ENABLED=false`
- 默认值: `false`（生产环境优先）

**理由**:
- 现有 `server.mode` 参数控制 Gin 运行模式，但不够灵活
- 调试模式开关可统一控制日志级别、性能优化等行为
- 便于开发和生产环境快速切换

**实现方式**:
```go
type DebugConfig struct {
    Enabled bool
}

type Config struct {
    Debug    DebugConfig    // 新增
    Server   ServerConfig
    Database DatabaseConfig
    Session  SessionConfig
    Auth     AuthConfig
}

// Load() 中：
cfg.Debug.Enabled = getBool("AG_DEBUG_ENABLED", cfg.Debug.Enabled)

// 根据 debug.enabled 设置 Gin 模式和日志级别
if cfg.Debug.Enabled {
    gin.SetMode(gin.DebugMode)
    logger.Default.LogMode(logger.Info)
} else {
    gin.SetMode(gin.ReleaseMode)
    logger.Default.LogMode(logger.Silent)
}

// 设置 provider 包的 debug 开关
provider.SetDebugMode(cfg.Debug.Enabled)
```

**影响范围**:
- Gin 框架运行模式（Debug/Release）
- Gorm 日志级别（Info/Silent）
- Provider 包调试记录功能（`server/internal/provider/debug.go`）
  - 启用时会记录请求/响应 body 到 `debug/` 目录
  - 启用时会记录错误信息到 `debug/` 目录
  - 启用时会记录流式数据（原始和转换后）到 `debug/` 目录

**替代方案**:
- 直接使用 `server.mode` - 缺少日志级别控制，不够灵活
- 使用多个调试参数（debug.log, debug.performance）- 配置过于分散

## Risks / Trade-offs

### 风险：配置文件缺失导致启动失败

**风险**: 如果默认 `config.yaml` 不存在且未指定路径，可能导致启动失败

**缓解方案**:
- 不强制要求 YAML 文件存在
- 如果 YAML 文件不存在，直接使用环境变量或默认值
- 提供示例配置文件 `config.yaml.example`

### 风险：SQLite 到 PostgreSQL 数据迁移

**风险**: 已有 SQLite 数据需要迁移到 PostgreSQL，数据可能丢失

**缓解方案**:
- 提供 SQL 导出工具（SQLite → PostgreSQL）
- 迁移文档明确说明步骤
- 测试环境验证迁移方案

### 风险：PostgreSQL 连接失败

**风险**: PostgreSQL 未启动或连接参数错误导致服务无法启动

**缓解方案**:
- 启动时验证数据库连接
- 提供明确的错误日志
- 支持 SQLite 作为备用方案（开发环境）

### 权衡：配置加载复杂度增加

**权衡**: 双重配置方式增加了配置加载逻辑复杂度

**接受理由**:
- 提供更好的灵活性和可维护性
- 符合现代应用的配置管理最佳实践
- 对运行时性能无影响（仅启动时加载）

### 权衡：PostgreSQL 依赖增加

**权衡**: 引入 PostgreSQL 驱动增加了依赖包

**接受理由**:
- PostgreSQL 驱动仅在配置使用 PostgreSQL 时加载
- SQLite 驱动仍然保留（轻量场景）
- 依赖包体积可控（约 1MB）

## Migration Plan

### 部署步骤（从 SQLite 到 PostgreSQL）

1. **准备 PostgreSQL 环境**
   - 安装 PostgreSQL 服务
   - 创建数据库 `ai_gateway`
   - 创建用户并授权

2. **配置切换**
   - 创建 `config.yaml` 文件，设置 `database.type: postgres`
   - 配置 PostgreSQL 连接参数

3. **数据迁移（可选）**
   - 使用 SQLite 导出工具导出数据
   - 导入到 PostgreSQL 数据库

4. **重启服务**
   - 验证 PostgreSQL 连接成功
   - 验证业务功能正常

### 回滚策略

如果 PostgreSQL 迁移失败：
1. 修改配置文件 `database.type: sqlite`
2. 恢复 SQLite 数据文件
3. 重启服务

### 新部署方式（无迁移）

直接使用 PostgreSQL：
1. 创建 `config.yaml` 配置 PostgreSQL 参数
2. 启动服务（自动创建表结构）

## Open Questions

1. **是否需要配置文件验证功能？**（如 YAML 格式错误提示）
   - 当前暂不实现，依赖启动时错误日志
   
2. **是否需要支持多配置文件（如 `config.dev.yaml`, `config.prod.yaml`）？**
   - 当前通过环境变量 `AG_CONFIG_PATH` 指定路径即可
   - 未来可考虑扩展
   
3. **PostgreSQL 连接池配置是否需要支持？**
   - 当前暂不实现，使用 Gorm 默认配置
   - 可通过未来扩展添加 `database.pool_size` 等参数