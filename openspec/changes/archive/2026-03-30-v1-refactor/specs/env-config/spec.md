## ADDED Requirements

### Requirement: Environment Variable Configuration

The system SHALL support configuration via environment variables with `AMP_` prefix.

#### Scenario: Load configuration from environment variable
- **WHEN** environment variable `AMP_SERVER_PORT` is set to `3000`
- **THEN** the server SHALL start on port 3000

#### Scenario: Use default value when environment variable not set
- **WHEN** environment variable `AMP_SERVER_PORT` is not set
- **THEN** the server SHALL use default value `8080`

### Requirement: Configuration Default Values

The system SHALL provide default values for all configuration items.

| Variable | Default Value |
|----------|---------------|
| `AMP_SERVER_PORT` | `18080` |
| `AMP_SERVER_MODE` | `debug` |
| `AMP_DATABASE_PATH` | `data.db` |
| `AMP_SESSION_SECRET` | (auto-generated) |
| `AMP_SESSION_MAX_AGE` | `86400` |
| `AMP_SESSION_SECURE` | `false` |
| `AMP_SESSION_HTTP_ONLY` | `true` |
| `AMP_SESSION_SAME_SITE` | `lax` |
| `AMP_ADMIN_USERNAME` | `admin` |
| `AMP_ADMIN_PASSWORD` | `admin` |

#### Scenario: All defaults applied
- **WHEN** no environment variables are set
- **THEN** the system SHALL start with all default values

### Requirement: Session Secret Auto-Generation

The system SHALL automatically generate a session secret when `AMP_SESSION_SECRET` is not set.

#### Scenario: Generate session secret
- **WHEN** environment variable `AMP_SESSION_SECRET` is not set
- **THEN** the system SHALL generate a random 32-byte secret
- **AND** the secret SHALL be base64 encoded

#### Scenario: Use provided session secret
- **WHEN** environment variable `AMP_SESSION_SECRET` is set to `my-secret-key`
- **THEN** the system SHALL use `my-secret-key` as the session secret

### Requirement: Environment Variable Names

The system SHALL use the following environment variable names:

| Variable | Description |
|----------|-------------|
| `AMP_SERVER_PORT` | Server listen port |
| `AMP_SERVER_MODE` | Gin mode (debug/release) |
| `AMP_DATABASE_PATH` | SQLite database file path |
| `AMP_SESSION_SECRET` | Session encryption secret |
| `AMP_SESSION_MAX_AGE` | Session max age in seconds |
| `AMP_SESSION_SECURE` | Cookie secure flag |
| `AMP_SESSION_HTTP_ONLY` | Cookie httpOnly flag |
| `AMP_SESSION_SAME_SITE` | Cookie SameSite attribute |
| `AMP_ADMIN_USERNAME` | Default admin username |
| `AMP_ADMIN_PASSWORD` | Default admin password |

#### Scenario: Invalid variable name ignored
- **WHEN** environment variable `INVALID_VAR` is set
- **THEN** the system SHALL ignore it and use corresponding default

### Requirement: No YAML Configuration File

The system SHALL NOT require a YAML configuration file to start.

#### Scenario: Start without config file
- **WHEN** no `configs/config.yaml` file exists
- **THEN** the system SHALL start successfully using environment variables and defaults

### Requirement: Configuration Logging

The system SHALL log the effective configuration at startup (sensitive values masked).

#### Scenario: Log configuration on startup
- **WHEN** the server starts
- **THEN** it SHALL log all configuration values
- **AND** sensitive values (secrets, passwords) SHALL be masked as `****`
