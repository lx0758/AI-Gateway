## ADDED Requirements

### Requirement: PostgreSQL database type selection

The system SHALL support PostgreSQL as a database backend option, configurable through YAML or environment variable.

#### Scenario: PostgreSQL selected via YAML
- **WHEN** YAML file contains `database.type: postgres`
- **THEN** the system SHALL initialize a PostgreSQL database connection

#### Scenario: PostgreSQL selected via environment variable
- **WHEN** environment variable `AG_DATABASE_TYPE` is set to `postgres`
- **THEN** the system SHALL initialize a PostgreSQL database connection

#### Scenario: SQLite selected by default
- **WHEN** `database.type` is not set or set to `sqlite`
- **THEN** the system SHALL initialize an SQLite database connection as the default behavior

### Requirement: PostgreSQL connection parameters

The system SHALL support PostgreSQL connection parameters through YAML and environment variables.

#### Scenario: PostgreSQL connection via YAML
- **WHEN** YAML file contains `database.host`, `database.port`, `database.user`, `database.password`, `database.name`
- **THEN** the system SHALL construct a PostgreSQL DSN using these parameters

#### Scenario: PostgreSQL connection via environment variables
- **WHEN** environment variables `AG_DATABASE_HOST`, `AG_DATABASE_PORT`, `AG_DATABASE_USER`, `AG_DATABASE_PASSWORD`, `AG_DATABASE_NAME` are set
- **THEN** the system SHALL construct a PostgreSQL DSN using these environment variables

#### Scenario: PostgreSQL environment variable names
- **WHEN** PostgreSQL configuration is needed
- **THEN** the system SHALL use environment variable names: `AG_DATABASE_HOST`, `AG_DATABASE_PORT`, `AG_DATABASE_USER`, `AG_DATABASE_PASSWORD`, `AG_DATABASE_NAME`

### Requirement: PostgreSQL driver integration

The system SHALL use the official Gorm PostgreSQL driver (`gorm.io/driver/postgres`) for PostgreSQL connections.

#### Scenario: PostgreSQL driver loaded
- **WHEN** PostgreSQL database type is selected
- **THEN** the system SHALL use `postgres.Open()` as the Gorm dialector

#### Scenario: SQLite driver still available
- **WHEN** SQLite database type is selected
- **THEN** the system SHALL continue to use `sqlite.Open()` as the Gorm dialector

### Requirement: PostgreSQL connection validation

The system SHALL validate PostgreSQL connection at startup and provide clear error messages on failure.

#### Scenario: PostgreSQL connection successful
- **WHEN** PostgreSQL connection parameters are correct AND PostgreSQL server is accessible
- **THEN** the system SHALL successfully initialize the database connection AND start the service

#### Scenario: PostgreSQL connection failed
- **WHEN** PostgreSQL server is not accessible OR connection parameters are incorrect
- **THEN** the system SHALL log a clear error message AND SHALL NOT start the service

### Requirement: PostgreSQL schema compatibility

The system SHALL ensure all existing data models work identically with PostgreSQL.

#### Scenario: Auto-migration works with PostgreSQL
- **WHEN** PostgreSQL database is initialized
- **THEN** the system SHALL perform auto-migration for all models (User, Provider, ProviderModel, Alias, AliasMapping, Key, KeyModel, UsageLog) successfully

#### Scenario: CRUD operations work with PostgreSQL
- **WHEN** PostgreSQL database is used
- **THEN** all CRUD operations SHALL function identically to SQLite (create, read, update, delete)

#### Scenario: Query performance with PostgreSQL
- **WHEN** PostgreSQL database is used for queries
- **THEN** query performance SHALL be acceptable for production use

### Requirement: Database type configuration documentation

The system SHALL document both SQLite and PostgreSQL configuration options clearly.

#### Scenario: SQLite configuration documented
- **WHEN** deployment documentation is consulted
- **THEN** SQLite configuration (database.type, database.path) SHALL be clearly documented

#### Scenario: PostgreSQL configuration documented
- **WHEN** deployment documentation is consulted
- **THEN** PostgreSQL configuration (host, port, user, password, name) SHALL be clearly documented

#### Scenario: Database type switch documented
- **WHEN** deployment documentation is consulted
- **THEN** instructions for switching database types SHALL be provided

### Requirement: SQLite to PostgreSQL migration support

The system SHALL provide guidance and tools for migrating data from SQLite to PostgreSQL.

#### Scenario: Migration guide provided
- **WHEN** user wants to migrate from SQLite to PostgreSQL
- **THEN** the system SHALL provide a migration guide document with step-by-step instructions

#### Scenario: Migration tool available
- **WHEN** user wants to export SQLite data
- **THEN** the system SHALL provide a tool or script to export SQLite data in PostgreSQL-compatible format