## MODIFIED Requirements

### Requirement: Query provider statistics

The system SHALL provide an API to query usage statistics grouped by provider.

#### Scenario: Query provider stats
- **WHEN** user requests provider statistics for a date range
- **THEN** system returns each provider's call count, total tokens, and average latency

### Requirement: Query key statistics

The system SHALL provide an API to query usage statistics grouped by API key.

#### Scenario: Query key stats
- **WHEN** user requests API key statistics for a date range
- **THEN** system returns each key's call count, total tokens, and average latency

**Reason**: Replaced by frontend aggregation using raw logs from `/usage/logs` endpoint.

**Migration**: Frontend now calculates keyStats by aggregating logs from `/usage/logs` API.
