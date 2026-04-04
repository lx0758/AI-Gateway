## ADDED Requirements

### Requirement: Database connection pool configuration

The system SHALL support configurable database connection pool parameters to optimize resource usage.

#### Scenario: Connection pool with default values for PostgreSQL
- **WHEN** database type is PostgreSQL AND no pool configuration is provided
- **THEN** the system SHALL use default pool values: MaxOpenConns=20, MaxIdleConns=5, ConnMaxLifetime=1h, ConnMaxIdleTime=5m

#### Scenario: Connection pool with default values for MySQL
- **WHEN** database type is MySQL AND no pool configuration is provided
- **THEN** the system SHALL use default pool values: MaxOpenConns=20, MaxIdleConns=5, ConnMaxLifetime=1h, ConnMaxIdleTime=5m

#### Scenario: Connection pool with default values for SQLite
- **WHEN** database type is SQLite AND no pool configuration is provided
- **THEN** the system SHALL use default pool values: MaxOpenConns=1, MaxIdleConns=1 (SQLite single connection constraint)

#### Scenario: Connection pool with custom values
- **WHEN** pool configuration parameters are specified
- **THEN** the system SHALL use the specified values for connection pool settings

#### Scenario: Pool configuration via YAML
- **WHEN** YAML file contains `database.pool.max_open`, `database.pool.max_idle`, `database.pool.max_lifetime`, `database.pool.max_idle_time`
- **THEN** the system SHALL parse these fields into `DatabaseConfig.PoolConfig`

#### Scenario: Pool configuration via environment variables
- **WHEN** environment variables `AG_DATABASE_POOL_MAX_OPEN`, `AG_DATABASE_POOL_MAX_IDLE`, `AG_DATABASE_POOL_MAX_LIFETIME`, `AG_DATABASE_POOL_MAX_IDLE_TIME` are set
- **THEN** the system SHALL use these values for connection pool parameters

### Requirement: Connection pool logging

The system SHALL log connection pool configuration on startup for operational visibility.

#### Scenario: Pool configuration logged on startup
- **WHEN** database connection is initialized
- **THEN** the system SHALL log the pool configuration values (MaxOpen, MaxIdle, MaxLifetime, MaxIdleTime)

### Requirement: Connection pool constraint enforcement

The system SHALL enforce SQLite connection pool constraints.

#### Scenario: SQLite MaxOpenConns exceeds 1
- **WHEN** database type is SQLite AND configured MaxOpenConns > 1
- **THEN** the system SHALL limit MaxOpenConns to 1 AND log a warning about SQLite constraint

#### Scenario: SQLite MaxIdleConns exceeds 1
- **WHEN** database type is SQLite AND configured MaxIdleConns > 1
- **THEN** the system SHALL limit MaxIdleConns to 1 AND log a warning about SQLite constraint