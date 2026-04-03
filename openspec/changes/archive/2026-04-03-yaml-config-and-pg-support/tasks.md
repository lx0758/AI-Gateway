## 1. 依赖准备

- [x] 1.1 添加 YAML 解析库依赖（`gopkg.in/yaml.v3`）到 `server/go.mod`
- [x] 1.2 添加 PostgreSQL 驱动依赖（`gorm.io/driver/postgres`）到 `server/go.mod`
- [x] 1.3 更新依赖：运行 `go mod tidy`

## 2. 配置结构体扩展

- [x] 2.1 扩展 `DatabaseConfig` 结构体，添加 PostgreSQL 配置字段（Type, Host, Port, User, Password, Name）
- [x] 2.2 新增 `DebugConfig` 结构体，包含 `Enabled` 字段
- [x] 2.3 在 `Config` 结构体中添加 `Debug` 字段
- [x] 2.4 更新 `DSN()` 方法，支持根据数据库类型返回不同的 DSN

## 3. YAML 配置加载实现

- [x] 3.1 实现 `loadYAML(configPath string) *Config` 函数，解析 YAML 文件
- [x] 3.2 实现 YAML 文件加载逻辑（固定路径：server/config.yaml）
- [x] 3.3 修改 `Load()` 函数，实现环境变量覆盖 YAML 逻辑（环境变量 > YAML > 默认值）
- [x] 3.4 添加 YAML 解析错误处理和日志
- [x] 3.5 确保 YAML 文件不存在时使用环境变量和默认值（不报错）
- [x] 3.6 在 `Load()` 中添加 `debug.enabled` 环境变量覆盖（`AG_DEBUG_ENABLED`）

## 4. PostgreSQL 数据库连接实现

- [x] 4.1 修改 `InitDB` 函数签名，接收完整的 `DatabaseConfig` 参数
- [x] 4.2 实现数据库类型判断逻辑（switch database.type）
- [x] 4.3 实现 PostgreSQL DSN 构建逻辑（host, port, user, password, name）
- [x] 4.4 使用 `postgres.Open()` 初始化 PostgreSQL 连接
- [x] 4.5 确保 SQLite 连接逻辑保持不变（兼容性）
- [x] 4.6 添加 PostgreSQL 连接失败时的清晰错误日志

## 5. 主程序调整

- [x] 5.1 更新 `main.go`，传递完整的 `DatabaseConfig` 到 `InitDB`
- [x] 5.2 添加数据库连接验证逻辑（启动时检查连接）
- [x] 5.3 更新配置日志输出，显示数据库类型和连接参数
- [x] 5.4 根据 `debug.enabled` 设置 Gin 运行模式（Debug/Release）
- [x] 5.5 根据 `debug.enabled` 设置 Gorm 日志级别（Info/Silent）

## 6. 调试模式功能实现

- [x] 6.1 实现调试模式开关逻辑（根据 `cfg.Debug.Enabled`）
- [x] 6.2 调试模式开启时，设置 Gin 为 DebugMode
- [x] 6.3 调试模式关闭时，设置 Gin 为 ReleaseMode
- [x] 6.4 调试模式开启时，启用详细请求日志
- [x] 6.5 调试模式关闭时，禁用详细请求日志（静默模式）
- [x] 6.6 修改 `server/internal/provider/debug.go`，将 `DEBUG` 常量改为可配置变量
- [x] 6.7 在 provider 包中添加 `SetDebugMode(enabled bool)` 函数
- [x] 6.8 在 `main.go` 中调用 `provider.SetDebugMode(cfg.Debug.Enabled)` 设置 provider debug 开关
- [x] 6.9 验证 provider debug 开关控制 recordBody、recordError、recordStream 函数的行为

## 7. 配置示例和文档

- [x] 7.1 创建 `server/config.yaml.example` 文件，包含所有配置参数示例
- [x] 7.2 为 `config.yaml.example` 添加注释说明，包含 debug.enabled 说明
- [x] 7.3 更新 `README.md`，说明 YAML 配置和环境变量配置方式
- [x] 7.4 更新 `README.md`，说明 PostgreSQL 配置和迁移步骤
- [x] 7.5 更新 `README.md`，说明 debug.enabled 配置和使用方式
- [x] 7.6 创建 SQLite 到 PostgreSQL 迁移指南文档（可选独立文档或嵌入 README）

## 8. 配置参数扩展规范文档

- [x] 8.1 创建配置参数开发规范文档（确保未来参数支持 YAML + 环境变量）
- [x] 8.2 在规范中明确环境变量命名规则（`AG_*` 前缀）
- [x] 8.3 在规范中明确 YAML 结构设计原则（嵌套结构、字段命名）

## 9. 测试和验证

- [x] 9.1 测试 YAML 配置文件加载（固定路径 config.yaml）
- [x] 9.2 测试 YAML 文件不存在时的配置加载
- [x] 9.3 测试环境变量覆盖 YAML 配置
- [x] 9.4 测试 SQLite 数据库连接（默认行为）
- [ ] 9.5 测试 PostgreSQL 数据库连接（YAML 配置）
- [ ] 9.6 测试 PostgreSQL 数据库连接（环境变量配置）
- [ ] 9.7 测试 PostgreSQL 连接失败时的错误处理
- [ ] 9.8 测试 PostgreSQL 数据库的 auto-migration
- [ ] 9.9 测试 PostgreSQL 数据库的 CRUD 操作
- [x] 9.10 测试 debug.enabled 开启时的 Gin 和 Gorm 行为
- [x] 9.11 测试 debug.enabled 关闭时的 Gin 和 Gorm 行为
- [x] 9.12 测试环境变量 `AG_DEBUG_ENABLED` 覆盖 YAML 配置
- [x] 9.13 测试 debug.enabled 开启时 provider debug 记录功能（生成 debug 文件）
- [x] 9.14 测试 debug.enabled 关闭时 provider 不生成 debug 文件

**注意**: PostgreSQL 相关测试（9.5-9.9）需要在真实 PostgreSQL 环境中手动验证。

## 10. 代码质量检查

- [x] 10.1 运行 `go fmt` 格式化代码
- [x] 10.2 运行静态代码检查（如 `go vet`）
- [x] 10.3 确保 Linter 检查通过
- [x] 10.4 添加必要的代码注释