## Why

当前系统所有配置参数仅支持环境变量方式，缺乏灵活的配置管理方案，部署和维护不够便捷。同时，数据库仅支持 SQLite，无法满足生产环境高并发、高可用的需求。需要引入 YAML 配置文件支持和 PostgreSQL 数据库支持，为未来扩展提供基础。

## What Changes

- 引入 YAML 配置文件支持（默认文件名 `config.yaml`，路径固定）
- 所有配置参数同时支持 YAML 文件和环境变量两种配置方式
- **环境变量优先级高于 YAML 配置**，便于临时覆盖配置
- 引入 PostgreSQL 数据库支持，与 SQLite 并存（通过配置选择）
- 新增配置参数：数据库类型选择（sqlite/postgres）
- 新增配置参数：PostgreSQL 连接参数（host, port, user, password, database）
- 新增配置参数：调试模式开关（debug.enabled）
- **未来新增配置参数必须同时支持 YAML 和环境变量**
- 现有环境变量命名保持不变（`AG_*` 前缀）

## Capabilities

### New Capabilities

- `yaml-config`: YAML 配置文件加载和管理能力，支持与环境变量并存且优先级控制
- `postgres-support`: PostgreSQL 数据库连接和操作能力，支持与 SQLite 共存

### Modified Capabilities

无

## Impact

- **代码影响**:
  - `server/internal/config/config.go`: 重构配置加载逻辑，支持 YAML + 环境变量双重来源
  - `server/internal/model/db.go`: 支持多种数据库类型初始化
  - `server/cmd/server/main.go`: 配置和数据库初始化调整
  - `server/internal/provider/debug.go`: 将硬编码的 `DEBUG` 常量改为可配置，支持动态开关
  
- **依赖新增**:
  - YAML 解析库: `gopkg.in/yaml.v3` 或类似
  - PostgreSQL 驱动: `gorm.io/driver/postgres`

- **配置文件**: 新增 `config.yaml` 示例文件和文档说明

- **部署影响**: 需要更新部署文档，说明 YAML 配置和 PostgreSQL 配置方式

- **调试功能**: debug.enabled 开关将控制 Gin/Gorm 日志级别及 provider 包的请求/响应记录功能