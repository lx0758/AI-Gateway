## ADDED Requirements

### Requirement: YAML configuration file loading

The system SHALL support loading configuration parameters from a YAML file named `config.yaml` located in the server directory.

#### Scenario: Default YAML file loading
- **WHEN** a `config.yaml` file exists in the server directory
- **THEN** the system SHALL load configuration parameters from `config.yaml`

#### Scenario: YAML file not found
- **WHEN** the `config.yaml` file does not exist in the server directory
- **THEN** the system SHALL proceed using environment variables and default values without error

### Requirement: Environment variable override priority

The system SHALL prioritize environment variables over YAML configuration values, allowing temporary configuration overrides without modifying files.

#### Scenario: Environment variable overrides YAML value
- **WHEN** a configuration parameter exists in both YAML file and environment variable
- **THEN** the system SHALL use the environment variable value and ignore the YAML value

#### Scenario: YAML provides default when no environment variable
- **WHEN** a configuration parameter exists in YAML file but not in environment variable
- **THEN** the system SHALL use the YAML file value

#### Scenario: Default value when neither YAML nor environment variable set
- **WHEN** a configuration parameter exists neither in YAML file nor in environment variable
- **THEN** the system SHALL use the hardcoded default value

### Requirement: YAML configuration structure

The system SHALL define a YAML configuration structure that mirrors the existing configuration groups (debug, server, database, session, auth).

#### Scenario: Debug configuration in YAML
- **WHEN** YAML file contains `debug.enabled` field
- **THEN** the system SHALL parse this field into `DebugConfig.Enabled`

#### Scenario: Server configuration in YAML
- **WHEN** YAML file contains `server.port` and `server.mode` fields
- **THEN** the system SHALL parse these fields into `ServerConfig.Port` and `ServerConfig.Mode`

#### Scenario: Database configuration in YAML
- **WHEN** YAML file contains `database.type`, `database.path`, `database.host`, etc.
- **THEN** the system SHALL parse these fields into `DatabaseConfig` structure

#### Scenario: Session configuration in YAML
- **WHEN** YAML file contains `session.secret`, `session.max_age`, etc.
- **THEN** the system SHALL parse these fields into `SessionConfig` structure

#### Scenario: Auth configuration in YAML
- **WHEN** YAML file contains `auth.default_admin.username` and `auth.default_admin.password`
- **THEN** the system SHALL parse these fields into `AuthConfig` structure

### Requirement: Configuration parameter extension standard

The system SHALL enforce a standard requiring all future configuration parameters to support both YAML and environment variable configuration methods.

#### Scenario: New parameter has YAML support
- **WHEN** a new configuration parameter is added to the system
- **THEN** the parameter MUST be defined in the YAML configuration structure

#### Scenario: New parameter has environment variable support
- **WHEN** a new configuration parameter is added to the system
- **THEN** the parameter MUST be accessible via an environment variable with `AG_` prefix

#### Scenario: New parameter documented
- **WHEN** a new configuration parameter is added to the system
- **THEN** documentation SHALL describe both YAML path and environment variable name

### Requirement: YAML file format validation

The system SHALL validate YAML file format and provide clear error messages when parsing fails.

#### Scenario: Invalid YAML syntax
- **WHEN** YAML file contains invalid syntax (e.g., malformed indentation)
- **THEN** the system SHALL log a clear error message indicating the parse failure AND SHALL NOT start the service

#### Scenario: Missing required YAML fields
- **WHEN** YAML file is valid but missing optional configuration fields
- **THEN** the system SHALL proceed using environment variables or default values for missing fields

### Requirement: Configuration file example

The system SHALL provide an example YAML configuration file demonstrating all supported parameters.

#### Scenario: Example file exists
- **WHEN** the project is deployed
- **THEN** a `config.yaml.example` file SHALL exist in the project directory

#### Scenario: Example file contains all parameters
- **WHEN** user opens `config.yaml.example`
- **THEN** the file SHALL contain all configuration parameters with example values and comments

### Requirement: Debug mode configuration

The system SHALL support a debug mode configuration parameter (`debug.enabled`) to control debugging behavior.

#### Scenario: Debug mode enabled via YAML
- **WHEN** YAML file contains `debug.enabled: true`
- **THEN** the system SHALL enable debug mode (verbose logging, Gin debug mode)

#### Scenario: Debug mode enabled via environment variable
- **WHEN** environment variable `AG_DEBUG_ENABLED` is set to `true`
- **THEN** the system SHALL enable debug mode

#### Scenario: Debug mode disabled by default
- **WHEN** `debug.enabled` is not set or set to `false`
- **THEN** the system SHALL run in production mode (silent logging, Gin release mode)

#### Scenario: Debug mode controls logging
- **WHEN** debug mode is enabled
- **THEN** the system SHALL set Gorm logger to `Info` level AND log detailed request information

#### Scenario: Debug mode controls Gin mode
- **WHEN** debug mode is enabled
- **THEN** the system SHALL set Gin framework to `DebugMode`

#### Scenario: Debug mode controls provider debug recording
- **WHEN** debug mode is enabled
- **THEN** the system SHALL enable provider debug recording functions (recordBody, recordError, recordStream) AND save debug logs to `debug/` directory

#### Scenario: Debug mode disables provider debug recording
- **WHEN** debug mode is disabled
- **THEN** the system SHALL disable provider debug recording functions AND NOT create debug logs

#### Scenario: Debug mode environment variable name
- **WHEN** debug mode configuration is needed
- **THEN** the environment variable name SHALL be `AG_DEBUG_ENABLED`

#### Scenario: Provider debug recording directory
- **WHEN** debug mode is enabled AND provider records debug data
- **THEN** debug files SHALL be saved in `debug/` directory with timestamp-based filenames