## Requirements

### Requirement: YAML 配置文件加载

系统 SHALL 支持从位于服务器目录的名为 `config.yaml` 的 YAML 文件加载配置参数。

#### Scenario: 默认 YAML 文件加载
- **WHEN** 服务器目录中存在 `config.yaml` 文件
- **THEN** 系统 SHALL 从 `config.yaml` 加载配置参数

#### Scenario: YAML 文件未找到
- **WHEN** 服务器目录中不存在 `config.yaml` 文件
- **THEN** 系统 SHALL 使用环境变量和默认值继续运行，不报错

### Requirement: 环境变量覆盖优先级

系统 SHALL 优先使用环境变量而非 YAML 配置值，允许在不修改文件的情况下临时覆盖配置。

#### Scenario: 环境变量覆盖 YAML 值
- **WHEN** 配置参数同时存在于 YAML 文件和环境变量
- **THEN** 系统 SHALL 使用环境变量值并忽略 YAML 值

#### Scenario: YAML 在无环境变量时提供默认值
- **WHEN** 配置参数存在于 YAML 文件但不存在于环境变量
- **THEN** 系统 SHALL 使用 YAML 文件值

#### Scenario: YAML 和环境变量均未设置时使用默认值
- **WHEN** 配置参数既不存在于 YAML 文件也不存在于环境变量
- **THEN** 系统 SHALL 使用硬编码的默认值

### Requirement: YAML 配置结构

系统 SHALL 定义一个 YAML 配置结构，映射现有的配置组（debug、server、database、auth、pprof）。

#### Scenario: YAML 中的 Debug 配置
- **WHEN** YAML 文件包含 `debug.gin`、`debug.gorm`、`debug.provider` 字段
- **THEN** 系统 SHALL 将这些字段解析为 `DebugConfig` 结构

#### Scenario: YAML 中的 Server 配置
- **WHEN** YAML 文件包含 `server.port`、`server.trusted_proxies`、`server.session.*` 字段
- **THEN** 系统 SHALL 将这些字段解析为 `ServerConfig` 结构，包括 `SessionConfig`

#### Scenario: YAML 中的 Database 配置
- **WHEN** YAML 文件包含 `database.type`、`database.path`、`database.host`、`database.pool.*` 等
- **THEN** 系统 SHALL 将这些字段解析为 `DatabaseConfig` 结构，包括 `PoolConfig`

#### Scenario: YAML 中的 Session 配置
- **WHEN** YAML 文件包含 `server.session.secret`、`server.session.max_age` 等
- **THEN** 系统 SHALL 将这些字段解析为 `ServerConfig.SessionConfig` 结构

#### Scenario: YAML 中的 Auth 配置
- **WHEN** YAML 文件包含 `auth.default_admin.username` 和 `auth.default_admin.password`
- **THEN** 系统 SHALL 将这些字段解析为 `AuthConfig` 结构

#### Scenario: YAML 中的 Pprof 配置
- **WHEN** YAML 文件包含 `pprof.port` 字段
- **THEN** 系统 SHALL 将该字段解析为 `PprofConfig.Port`

### Requirement: 配置参数扩展标准

系统 SHALL 执行一个标准，要求所有未来的配置参数必须同时支持 YAML 和环境变量配置方法。

#### Scenario: 新参数支持 YAML
- **WHEN** 向系统添加新的配置参数
- **THEN** 该参数 MUST 在 YAML 配置结构中定义

#### Scenario: 新参数支持环境变量
- **WHEN** 向系统添加新的配置参数
- **THEN** 该参数 MUST 可通过带 `AG_` 前缀的环境变量访问

#### Scenario: 新参数需文档化
- **WHEN** 向系统添加新的配置参数
- **THEN** 文档 SHALL 描述 YAML 路径和环境变量名称

### Requirement: YAML 文件格式验证

系统 SHALL 验证 YAML 文件格式，并在解析失败时提供清晰的错误消息。

#### Scenario: 无效的 YAML 语法
- **WHEN** YAML 文件包含无效语法（例如缩进错误）
- **THEN** 系统 SHALL 记录清晰的错误消息指示解析失败，并且 SHALL NOT 启动服务

#### Scenario: YAML 缺少必填字段
- **WHEN** YAML 文件有效但缺少可选配置字段
- **THEN** 系统 SHALL 对缺少的字段使用环境变量或默认值继续运行

### Requirement: 配置文件示例

系统 SHALL 提供一个示例 YAML 配置文件，展示所有支持的参数。

#### Scenario: 示例文件存在
- **WHEN** 项目部署时
- **THEN** 项目目录中 SHALL 存在 `config.yaml.example` 文件

#### Scenario: 示例文件包含所有参数
- **WHEN** 用户打开 `config.yaml.example`
- **THEN** 文件 SHALL 包含所有配置参数，附带示例值和注释

### Requirement: Debug 配置粒度

系统 SHALL 支持对单个组件的细粒度 Debug 配置。

#### Scenario: YAML 中的 Gin Debug 配置
- **WHEN** YAML 文件包含 `debug.gin` 字段
- **THEN** 系统 SHALL 将该字段解析为 `DebugConfig.Gin`，并控制 Gin 框架的 Debug 日志

#### Scenario: YAML 中的 GORM Debug 配置
- **WHEN** YAML 文件包含 `debug.gorm` 字段
- **THEN** 系统 SHALL 将该字段解析为 `DebugConfig.Gorm`，并控制 GORM SQL Debug 日志

#### Scenario: YAML 中的 Provider Debug 配置
- **WHEN** YAML 文件包含 `debug.provider` 字段
- **THEN** 系统 SHALL 将该字段解析为 `DebugConfig.Provider`，并控制 Provider Debug 记录

#### Scenario: 通过环境变量配置 Gin Debug
- **WHEN** 环境变量 `AG_DEBUG_GIN` 已设置
- **THEN** 系统 SHALL 使用该值作为 Gin Debug 模式

#### Scenario: 通过环境变量配置 GORM Debug
- **WHEN** 环境变量 `AG_DEBUG_GORM` 已设置
- **THEN** 系统 SHALL 使用该值作为 GORM Debug 模式

#### Scenario: 通过环境变量配置 Provider Debug
- **WHEN** 环境变量 `AG_DEBUG_PROVIDER` 已设置
- **THEN** 系统 SHALL 使用该值作为 Provider Debug 记录

#### Scenario: 通过环境变量配置 Session Secret
- **WHEN** 环境变量 `AG_SERVER_SESSION_SECRET` 已设置
- **THEN** 系统 SHALL 使用该值作为 Session Secret

#### Scenario: 通过环境变量配置 Session Max Age
- **WHEN** 环境变量 `AG_SERVER_SESSION_MAX_AGE` 已设置
- **THEN** 系统 SHALL 使用该值作为 Session Max Age

#### Scenario: 通过环境变量配置 Session Secure
- **WHEN** 环境变量 `AG_SERVER_SESSION_SECURE` 已设置
- **THEN** 系统 SHALL 使用该值作为 Session Secure 标志

#### Scenario: 通过环境变量配置 Session Http Only
- **WHEN** 环境变量 `AG_SERVER_SESSION_HTTP_ONLY` 已设置
- **THEN** 系统 SHALL 使用该值作为 Session Http Only 标志

#### Scenario: 通过环境变量配置 Session Same Site
- **WHEN** 环境变量 `AG_SERVER_SESSION_SAME_SITE` 已设置
- **THEN** 系统 SHALL 使用该值作为 Session Same Site 策略

#### Scenario: 通过环境变量配置 Trusted Proxies
- **WHEN** 环境变量 `AG_SERVER_TRUSTED_PROXIES` 已设置
- **THEN** 系统 SHALL 使用该值作为 Trusted Proxy IP 范围

### Requirement: Debug 模式配置

系统 SHALL 支持细粒度的 Debug 配置参数（`debug.gin`、`debug.gorm`、`debug.provider`）来控制各组件的调试行为。

#### Scenario: 通过 YAML 启用 Gin Debug 模式
- **WHEN** YAML 文件包含 `debug.gin: true`
- **THEN** 系统 SHALL 启用 Gin Debug 模式（Gin DebugMode）

#### Scenario: 通过环境变量启用 Gin Debug 模式
- **WHEN** 环境变量 `AG_DEBUG_GIN` 设置为 `true`
- **THEN** 系统 SHALL 启用 Gin Debug 模式

#### Scenario: 通过 YAML 启用 GORM Debug 模式
- **WHEN** YAML 文件包含 `debug.gorm: true`
- **THEN** 系统 SHALL 启用 GORM Debug 日志（Info 级别）

#### Scenario: 通过环境变量启用 GORM Debug 模式
- **WHEN** 环境变量 `AG_DEBUG_GORM` 设置为 `true`
- **THEN** 系统 SHALL 启用 GORM Debug 日志

#### Scenario: 通过 YAML 启用 Provider Debug 模式
- **WHEN** YAML 文件包含 `debug.provider: true`
- **THEN** 系统 SHALL 启用 Provider Debug 记录（保存到 'debug/' 目录）

#### Scenario: 通过环境变量启用 Provider Debug 模式
- **WHEN** 环境变量 `AG_DEBUG_PROVIDER` 设置为 `true`
- **THEN** 系统 SHALL 启用 Provider Debug 记录

#### Scenario: Debug 模式默认禁用
- **WHEN** `debug.gin`、`debug.gorm` 和 `debug.provider` 未设置或设置为 `false`
- **THEN** 系统 SHALL 以生产模式运行（Gin Release 模式、GORM Silent 日志、无 Provider 记录）

#### Scenario: Gin Debug 控制 Gin 模式
- **WHEN** debug.gin 启用时
- **THEN** 系统 SHALL 将 Gin 框架设置为 `DebugMode`

#### Scenario: GORM Debug 控制日志级别
- **WHEN** debug.gorm 启用时
- **THEN** 系统 SHALL 将 Gorm Logger 设置为 `Info` 级别

### Requirement: Server 下的 Session 配置

系统 SHALL 支持在 YAML 的 `server` 下嵌套 Session 配置。

#### Scenario: server 部分中的 Session 配置
- **WHEN** YAML 文件包含 `server.session.secret`、`server.session.max_age` 等
- **THEN** 系统 SHALL 将这些字段解析为 `ServerConfig.SessionConfig`

#### Scenario: 带 AG_SERVER_SESSION 前缀的 Session 环境变量
- **WHEN** 环境变量 `AG_SERVER_SESSION_SECRET`、`AG_SERVER_SESSION_MAX_AGE` 等已设置
- **THEN** 系统 SHALL 使用这些值作为 Session 配置