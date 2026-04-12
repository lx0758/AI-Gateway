## ADDED Requirements

### Requirement: Pprof 服务器配置

系统 SHALL 支持专用的 Pprof 服务器用于运行时性能分析，端口可配置。

#### Scenario: Pprof 服务器在默认端口启动
- **WHEN** 未配置 `pprof.port`
- **THEN** 系统 SHALL 在端口 6060 启动 Pprof 服务器

#### Scenario: Pprof 服务器在配置端口启动
- **WHEN** `pprof.port` 设置为有效端口号码
- **THEN** 系统 SHALL 在指定端口启动 Pprof 服务器

#### Scenario: Pprof 服务器绑定到 localhost
- **WHEN** Pprof 服务器启动
- **THEN** 系统 SHALL 将服务器仅绑定到 localhost（安全隔离）

#### Scenario: Pprof 服务器与主服务器同时启动
- **WHEN** 主服务器成功启动
- **THEN** 系统 SHALL 在单独的 goroutine 中同时启动 Pprof 服务器

#### Scenario: Pprof 服务器失败不影响主服务器
- **WHEN** Pprof 服务器启动失败
- **THEN** 系统 SHALL 记录错误并继续运行主服务器

#### Scenario: 通过 YAML 配置 Pprof 端口
- **WHEN** YAML 文件包含 `pprof.port` 字段
- **THEN** 系统 SHALL 将该字段解析为 `PprofConfig.Port`

#### Scenario: 通过环境变量配置 Pprof 端口
- **WHEN** 环境变量 `AG_PPROF_PORT` 已设置
- **THEN** 系统 SHALL 使用该值作为 Pprof 服务器端口

### Requirement: Pprof 服务器提供标准分析端点

系统 SHALL 暴露标准 Go Pprof 端点用于性能分析。

#### Scenario: CPU 分析端点可用
- **WHEN** Pprof 服务器运行时
- **THEN** 端点 `/debug/pprof/profile` SHALL 可用于 CPU 分析

#### Scenario: 内存分析端点可用
- **WHEN** Pprof 服务器运行时
- **THEN** 端点 `/debug/pprof/heap` SHALL 可用于内存分析

#### Scenario: Goroutine 分析端点可用
- **WHEN** Pprof 服务器运行时
- **THEN** 端点 `/debug/pprof/goroutine` SHALL 可用于 Goroutine 分析

#### Scenario: Block 分析端点可用
- **WHEN** Pprof 服务器运行时
- **THEN** 端点 `/debug/pprof/block` SHALL 可用于阻塞分析

#### Scenario: 完整分析索引可用
- **WHEN** Pprof 服务器运行时
- **THEN** 端点 `/debug/pprof/` SHALL 提供所有分析端点的索引