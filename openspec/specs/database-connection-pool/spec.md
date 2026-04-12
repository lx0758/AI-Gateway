## ADDED Requirements

### Requirement: 数据库连接池配置

系统 SHALL 支持可配置的数据库连接池参数以优化资源使用。

#### Scenario: PostgreSQL 默认连接池值
- **WHEN** 数据库类型为 PostgreSQL 且未提供连接池配置
- **THEN** 系统 SHALL 使用默认连接池值：MaxOpenConns=20、MaxIdleConns=5、ConnMaxLifetime=1h、ConnMaxIdleTime=5m

#### Scenario: MySQL 默认连接池值
- **WHEN** 数据库类型为 MySQL 且未提供连接池配置
- **THEN** 系统 SHALL 使用默认连接池值：MaxOpenConns=20、MaxIdleConns=5、ConnMaxLifetime=1h、ConnMaxIdleTime=5m

#### Scenario: SQLite 默认连接池值
- **WHEN** 数据库类型为 SQLite 且未提供连接池配置
- **THEN** 系统 SHALL 使用默认连接池值：MaxOpenConns=1、MaxIdleConns=1（SQLite 单连接约束）

#### Scenario: 自定义连接池值
- **WHEN** 连接池配置参数被指定
- **THEN** 系统 SHALL 使用指定的值作为连接池设置

#### Scenario: 通过 YAML 配置连接池
- **WHEN** YAML 文件包含 `database.pool.max_open`、`database.pool.max_idle`、`database.pool.max_lifetime`、`database.pool.max_idle_time`
- **THEN** 系统 SHALL 将这些字段解析为 `DatabaseConfig.PoolConfig`

#### Scenario: 通过环境变量配置连接池
- **WHEN** 环境变量 `AG_DATABASE_POOL_MAX_OPEN`、`AG_DATABASE_POOL_MAX_IDLE`、`AG_DATABASE_POOL_MAX_LIFETIME`、`AG_DATABASE_POOL_MAX_IDLE_TIME` 已设置
- **THEN** 系统 SHALL 使用这些值作为连接池参数

### Requirement: 连接池日志记录

系统 SHALL 在启动时记录连接池配置以提供运维可见性。

#### Scenario: 启动时记录连接池配置
- **WHEN** 数据库连接初始化时
- **THEN** 系统 SHALL 记录连接池配置值（MaxOpen、MaxIdle、MaxLifetime、MaxIdleTime）

### Requirement: 连接池约束强制

系统 SHALL 强制 SQLite 连接池约束。

#### Scenario: SQLite MaxOpenConns 超过 1
- **WHEN** 数据库类型为 SQLite 且配置的 MaxOpenConns > 1
- **THEN** 系统 SHALL 将 MaxOpenConns 限制为 1 并记录关于 SQLite 约束的警告

#### Scenario: SQLite MaxIdleConns 超过 1
- **WHEN** 数据库类型为 SQLite 且配置的 MaxIdleConns > 1
- **THEN** 系统 SHALL 将 MaxIdleConns 限制为 1 并记录关于 SQLite 约束的警告