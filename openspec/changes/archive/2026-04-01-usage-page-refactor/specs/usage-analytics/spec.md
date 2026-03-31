## ADDED Requirements

### Requirement: Frontend aggregation of usage logs

The system SHALL provide raw usage logs via a single API endpoint, enabling frontend to aggregate and calculate various statistics.

#### Scenario: Aggregate key statistics from logs
- **WHEN** user opens the usage page
- **THEN** frontend fetches logs from `/usage/logs` API and calculates key statistics (count, total tokens, avg latency) grouped by key_name

#### Scenario: Aggregate model statistics from logs
- **WHEN** user opens the usage page
- **THEN** frontend fetches logs from `/usage/logs` API and calculates model statistics (count, total tokens, avg latency) grouped by model

#### Scenario: Aggregate source statistics from logs
- **WHEN** user opens the usage page
- **THEN** frontend fetches logs from `/usage/logs` API and calculates source statistics (count, total tokens, avg latency) grouped by source

#### Scenario: Aggregate provider statistics from logs
- **WHEN** user opens the usage page
- **THEN** frontend fetches logs from `/usage/logs` API and calculates provider statistics (count, total tokens, avg latency) grouped by provider_name

#### Scenario: Aggregate provider-model statistics from logs
- **WHEN** user opens the usage page
- **THEN** frontend fetches logs from `/usage/logs` API and calculates provider-model statistics (count, total tokens, avg latency) grouped by provider_name and model

#### Scenario: Calculate overall statistics from logs
- **WHEN** user opens the usage page
- **THEN** frontend fetches logs from `/usage/logs` API and calculates overall statistics including total requests, success rate, total tokens, and average latency
