## ADDED Requirements

### Requirement: Pprof server configuration

The system SHALL support a dedicated pprof server for runtime performance profiling with configurable port.

#### Scenario: Pprof server starts on default port
- **WHEN** no `pprof.port` is configured
- **THEN** the system SHALL start pprof server on port 6060

#### Scenario: Pprof server starts on configured port
- **WHEN** `pprof.port` is set to a valid port number
- **THEN** the system SHALL start pprof server on the specified port

#### Scenario: Pprof server binds to localhost
- **WHEN** pprof server starts
- **THEN** the system SHALL bind the server to localhost only (security isolation)

#### Scenario: Pprof server starts alongside main server
- **WHEN** main server starts successfully
- **THEN** the system SHALL also start pprof server in a separate goroutine

#### Scenario: Pprof server failure does not affect main server
- **WHEN** pprof server fails to start
- **THEN** the system SHALL log the error AND continue running the main server

#### Scenario: Pprof port configuration via YAML
- **WHEN** YAML file contains `pprof.port` field
- **THEN** the system SHALL parse this field into `PprofConfig.Port`

#### Scenario: Pprof port configuration via environment variable
- **WHEN** environment variable `AG_PPROF_PORT` is set
- **THEN** the system SHALL use this value for pprof server port

### Requirement: Pprof server provides standard profiling endpoints

The system SHALL expose standard Go pprof endpoints for performance analysis.

#### Scenario: CPU profiling endpoint available
- **WHEN** pprof server is running
- **THEN** the endpoint `/debug/pprof/profile` SHALL be available for CPU profiling

#### Scenario: Memory profiling endpoint available
- **WHEN** pprof server is running
- **THEN** the endpoint `/debug/pprof/heap` SHALL be available for memory profiling

#### Scenario: Goroutine profiling endpoint available
- **WHEN** pprof server is running
- **THEN** the endpoint `/debug/pprof/goroutine` SHALL be available for goroutine analysis

#### Scenario: Block profiling endpoint available
- **WHEN** pprof server is running
- **THEN** the endpoint `/debug/pprof/block` SHALL be available for blocking analysis

#### Scenario: Full profiling index available
- **WHEN** pprof server is running
- **THEN** the endpoint `/debug/pprof/` SHALL provide an index of all profiling endpoints