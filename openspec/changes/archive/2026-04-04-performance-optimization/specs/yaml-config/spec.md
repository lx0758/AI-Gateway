## ADDED Requirements

### Requirement: Pprof configuration section

The system SHALL define a pprof configuration section in the YAML structure.

#### Scenario: Pprof configuration in YAML
- **WHEN** YAML file contains `pprof.port` field
- **THEN** the system SHALL parse this field into `PprofConfig.Port`

#### Scenario: Pprof configuration defaults
- **WHEN** `pprof.port` is not specified
- **THEN** the system SHALL use default port 6060

### Requirement: Database pool configuration section

The system SHALL define a database pool configuration section in the YAML structure.

#### Scenario: Pool configuration in YAML
- **WHEN** YAML file contains `database.pool.max_open`, `database.pool.max_idle`, `database.pool.max_lifetime`, `database.pool.max_idle_time`
- **THEN** the system SHALL parse these fields into `DatabaseConfig.PoolConfig`

### Requirement: Debug configuration granularity

The system SHALL support granular debug configuration for individual components.

#### Scenario: Gin debug configuration in YAML
- **WHEN** YAML file contains `debug.gin` field
- **THEN** the system SHALL parse this field into `DebugConfig.Gin` AND control Gin framework debug logging

#### Scenario: GORM debug configuration in YAML
- **WHEN** YAML file contains `debug.gorm` field
- **THEN** the system SHALL parse this field into `DebugConfig.Gorm` AND control GORM SQL debug logging

#### Scenario: Provider debug configuration in YAML
- **WHEN** YAML file contains `debug.provider` field
- **THEN** the system SHALL parse this field into `DebugConfig.Provider` AND control Provider debug recording

#### Scenario: Gin debug via environment variable
- **WHEN** environment variable `AG_DEBUG_GIN` is set
- **THEN** the system SHALL use this value for Gin debug mode

#### Scenario: GORM debug via environment variable
- **WHEN** environment variable `AG_DEBUG_GORM` is set
- **THEN** the system SHALL use this value for GORM debug mode

#### Scenario: Provider debug via environment variable
- **WHEN** environment variable `AG_DEBUG_PROVIDER` is set
- **THEN** the system SHALL use this value for Provider debug recording

#### Scenario: Session secret via environment variable
- **WHEN** environment variable `AG_SERVER_SESSION_SECRET` is set
- **THEN** the system SHALL use this value for session secret

#### Scenario: Session max age via environment variable
- **WHEN** environment variable `AG_SERVER_SESSION_MAX_AGE` is set
- **THEN** the system SHALL use this value for session max age

#### Scenario: Session secure via environment variable
- **WHEN** environment variable `AG_SERVER_SESSION_SECURE` is set
- **THEN** the system SHALL use this value for session secure flag

#### Scenario: Session http only via environment variable
- **WHEN** environment variable `AG_SERVER_SESSION_HTTP_ONLY` is set
- **THEN** the system SHALL use this value for session http only flag

#### Scenario: Session same site via environment variable
- **WHEN** environment variable `AG_SERVER_SESSION_SAME_SITE` is set
- **THEN** the system SHALL use this value for session same site policy

#### Scenario: Trusted proxies via environment variable
- **WHEN** environment variable `AG_SERVER_TRUSTED_PROXIES` is set
- **THEN** the system SHALL use this value for trusted proxy IP ranges

## MODIFIED Requirements

### Requirement: YAML configuration structure

The system SHALL define a YAML configuration structure that mirrors the existing configuration groups (debug, server, database, auth, pprof).

#### Scenario: Debug configuration in YAML
- **WHEN** YAML file contains `debug.gin`, `debug.gorm`, `debug.provider` fields
- **THEN** the system SHALL parse these fields into `DebugConfig` structure

#### Scenario: Server configuration in YAML
- **WHEN** YAML file contains `server.port`, `server.trusted_proxies`, `server.session.*` fields
- **THEN** the system SHALL parse these fields into `ServerConfig` structure including `SessionConfig`

#### Scenario: Database configuration in YAML
- **WHEN** YAML file contains `database.type`, `database.path`, `database.host`, `database.pool.*` etc.
- **THEN** the system SHALL parse these fields into `DatabaseConfig` structure including `PoolConfig`

#### Scenario: Session configuration in YAML
- **WHEN** YAML file contains `server.session.secret`, `server.session.max_age`, etc.
- **THEN** the system SHALL parse these fields into `ServerConfig.SessionConfig` structure

#### Scenario: Auth configuration in YAML
- **WHEN** YAML file contains `auth.default_admin.username` and `auth.default_admin.password`
- **THEN** the system SHALL parse these fields into `AuthConfig` structure

#### Scenario: Pprof configuration in YAML
- **WHEN** YAML file contains `pprof.port` field
- **THEN** the system SHALL parse this field into `PprofConfig.Port`

### Requirement: Debug mode configuration

The system SHALL support granular debug configuration parameters (`debug.gin`, `debug.gorm`, `debug.provider`) to control debugging behavior for individual components.

#### Scenario: Gin debug mode enabled via YAML
- **WHEN** YAML file contains `debug.gin: true`
- **THEN** the system SHALL enable Gin debug mode (Gin DebugMode)

#### Scenario: Gin debug mode enabled via environment variable
- **WHEN** environment variable `AG_DEBUG_GIN` is set to `true`
- **THEN** the system SHALL enable Gin debug mode

#### Scenario: GORM debug mode enabled via YAML
- **WHEN** YAML file contains `debug.gorm: true`
- **THEN** the system SHALL enable GORM debug logging (Info level)

#### Scenario: GORM debug mode enabled via environment variable
- **WHEN** environment variable `AG_DEBUG_GORM` is set to `true`
- **THEN** the system SHALL enable GORM debug logging

#### Scenario: Provider debug mode enabled via YAML
- **WHEN** YAML file contains `debug.provider: true`
- **THEN** the system SHALL enable Provider debug recording (save to 'debug/' directory)

#### Scenario: Provider debug mode enabled via environment variable
- **WHEN** environment variable `AG_DEBUG_PROVIDER` is set to `true`
- **THEN** the system SHALL enable Provider debug recording

#### Scenario: Debug mode disabled by default
- **WHEN** `debug.gin`, `debug.gorm`, and `debug.provider` are not set or set to `false`
- **THEN** the system SHALL run in production mode (Gin release mode, GORM silent logging, no provider recording)

#### Scenario: Gin debug controls Gin mode
- **WHEN** debug.gin is enabled
- **THEN** the system SHALL set Gin framework to `DebugMode`

#### Scenario: GORM debug controls logging level
- **WHEN** debug.gorm is enabled
- **THEN** the system SHALL set Gorm logger to `Info` level

### Requirement: Session configuration under server

The system SHALL support session configuration as a nested section under `server` in YAML.

#### Scenario: Session configuration in server section
- **WHEN** YAML file contains `server.session.secret`, `server.session.max_age`, etc.
- **THEN** the system SHALL parse these fields into `ServerConfig.SessionConfig`

#### Scenario: Session environment variables with AG_SERVER_SESSION prefix
- **WHEN** environment variables `AG_SERVER_SESSION_SECRET`, `AG_SERVER_SESSION_MAX_AGE`, etc. are set
- **THEN** the system SHALL use these values for session configuration

## REMOVED Requirements

### Requirement: Server mode configuration

**Reason**: The `server.mode` configuration has no functional impact. Gin mode is controlled by `debug.gin`, and `server.mode` was never used in practice, only mentioned in partial documentation.

**Migration**: Remove `server.mode` from YAML configuration files. The system will ignore this field if present and log a deprecation warning.

#### Scenario: Server mode field ignored
- **WHEN** YAML file contains `server.mode` field
- **THEN** the system SHALL ignore this field AND log a deprecation warning

### Requirement: Legacy session environment variable names

**Reason**: Session configuration moved under `server`, environment variable names updated for consistency.

**Migration**: Replace old environment variable names with new ones:
- `AG_SESSION_SECRET` → `AG_SERVER_SESSION_SECRET`
- `AG_SESSION_MAX_AGE` → `AG_SERVER_SESSION_MAX_AGE`
- `AG_SESSION_SECURE` → `AG_SERVER_SESSION_SECURE`
- `AG_SESSION_HTTP_ONLY` → `AG_SERVER_SESSION_HTTP_ONLY`
- `AG_SESSION_SAME_SITE` → `AG_SERVER_SESSION_SAME_SITE`

#### Scenario: Old session environment variable ignored
- **WHEN** environment variable `AG_SESSION_SECRET` or other `AG_SESSION_*` variables are set
- **THEN** the system SHALL NOT recognize these variables AND the new `AG_SERVER_SESSION_*` variables SHALL be used instead

### Requirement: Legacy trusted proxies environment variable name

**Reason**: TrustedProxies is a server configuration, environment variable name updated for consistency.

**Migration**: Replace `AG_TRUSTED_PROXIES` with `AG_SERVER_TRUSTED_PROXIES`

#### Scenario: Old trusted proxies environment variable ignored
- **WHEN** environment variable `AG_TRUSTED_PROXIES` is set
- **THEN** the system SHALL NOT recognize this variable AND `AG_SERVER_TRUSTED_PROXIES` SHALL be used instead