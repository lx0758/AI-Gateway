## ADDED Requirements

### Requirement: PostgreSQL 数据库类型选择

系统 SHALL 支持 PostgreSQL 作为数据库后端选项，可通过 YAML 或环境变量配置。

#### Scenario: 通过 YAML 选择 PostgreSQL
- **WHEN** YAML 文件包含 `database.type: postgres`
- **THEN** 系统 SHALL 初始化 PostgreSQL 数据库连接

#### Scenario: 通过环境变量选择 PostgreSQL
- **WHEN** 环境变量 `AG_DATABASE_TYPE` 设置为 `postgres`
- **THEN** 系统 SHALL 初始化 PostgreSQL 数据库连接

#### Scenario: 默认选择 SQLite
- **WHEN** `database.type` 未设置或设置为 `sqlite`
- **THEN** 系统 SHALL 初始化 SQLite 数据库连接作为默认行为

### Requirement: PostgreSQL 连接参数

系统 SHALL 通过 YAML 和环境变量支持 PostgreSQL 连接参数。

#### Scenario: 通过 YAML 进行 PostgreSQL 连接
- **WHEN** YAML 文件包含 `database.host`、`database.port`、`database.user`、`database.password`、`database.name`
- **THEN** 系统 SHALL 使用这些参数构建 PostgreSQL DSN

#### Scenario: 通过环境变量进行 PostgreSQL 连接
- **WHEN** 环境变量 `AG_DATABASE_HOST`、`AG_DATABASE_PORT`、`AG_DATABASE_USER`、`AG_DATABASE_PASSWORD`、`AG_DATABASE_NAME` 已设置
- **THEN** 系统 SHALL 使用这些环境变量构建 PostgreSQL DSN

#### Scenario: PostgreSQL 环境变量名称
- **WHEN** PostgreSQL 配置需要时
- **THEN** 系统 SHALL 使用环境变量名称：`AG_DATABASE_HOST`、`AG_DATABASE_PORT`、`AG_DATABASE_USER`、`AG_DATABASE_PASSWORD`、`AG_DATABASE_NAME`

### Requirement: PostgreSQL 驱动集成

系统 SHALL 使用官方 Gorm PostgreSQL 驱动（`gorm.io/driver/postgres`）进行 PostgreSQL 连接。

#### Scenario: PostgreSQL 驱动加载
- **WHEN** PostgreSQL 数据库类型被选中
- **THEN** 系统 SHALL 使用 `postgres.Open()` 作为 Gorm dialector

#### Scenario: SQLite 驱动仍可用
- **WHEN** SQLite 数据库类型被选中
- **THEN** 系统 SHALL 继续使用 `sqlite.Open()` 作为 Gorm dialector

### Requirement: PostgreSQL 连接验证

系统 SHALL 在启动时验证 PostgreSQL 连接，并在失败时提供清晰的错误消息。

#### Scenario: PostgreSQL 连接成功
- **WHEN** PostgreSQL 连接参数正确 AND PostgreSQL 服务器可访问
- **THEN** 系统 SHALL 成功初始化数据库连接 AND 启动服务

#### Scenario: PostgreSQL 连接失败
- **WHEN** PostgreSQL 服务器不可访问 OR 连接参数不正确
- **THEN** 系统 SHALL 记录清晰的错误消息 AND SHALL NOT 启动服务

### Requirement: PostgreSQL Schema 兼容性

系统 SHALL 确保所有现有数据模型与 PostgreSQL 工作相同。

#### Scenario: PostgreSQL Auto-migration 工作
- **WHEN** PostgreSQL 数据库初始化
- **THEN** 系统 SHALL 成功对所有模型（User、Provider、ProviderModel、Alias、AliasMapping、Key、KeyModel、UsageLog）执行 auto-migration

#### Scenario: PostgreSQL CRUD 操作工作
- **WHEN** PostgreSQL 数据库被使用
- **THEN** 所有 CRUD 操作 SHALL 与 SQLite 功能相同（create、read、update、delete）

#### Scenario: PostgreSQL 查询性能
- **WHEN** PostgreSQL 数据库用于查询
- **THEN** 查询性能 SHALL 可接受用于生产使用

### Requirement: 数据库类型配置文档

系统 SHALL 清晰文档化 SQLite 和 PostgreSQL 配置选项。

#### Scenario: SQLite 配置文档化
- **WHEN** 部署文档被查阅
- **THEN** SQLite 配置（database.type、database.path） SHALL 清晰文档化

#### Scenario: PostgreSQL 配置文档化
- **WHEN** 部署文档被查阅
- **THEN** PostgreSQL 配置（host、port、user、password、name） SHALL 清晰文档化

#### Scenario: 数据库类型切换文档化
- **WHEN** 部署文档被查阅
- **THEN** 切换数据库类型的说明 SHALL 提供

### Requirement: SQLite 到 PostgreSQL 迁移支持

系统 SHALL 提供指导和工具用于从 SQLite 迁移数据到 PostgreSQL。

#### Scenario: 提供迁移指南
- **WHEN** 用户想从 SQLite 迁移到 PostgreSQL
- **THEN** 系统 SHALL 提供带分步说明的迁移指南文档

#### Scenario: 迁移工具可用
- **WHEN** 用户想导出 SQLite 数据
- **THEN** 系统 SHALL 提供工具或脚本以 PostgreSQL 兼容格式导出 SQLite 数据