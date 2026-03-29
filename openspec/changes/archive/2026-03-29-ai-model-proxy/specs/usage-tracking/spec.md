## ADDED Requirements

### Requirement: Log API usage

The system SHALL record details of every API request.

#### Scenario: Log successful request
- **WHEN** API request completes successfully
- **THEN** system logs api_key_id, provider_id, model, tokens, latency, and status

#### Scenario: Log failed request
- **WHEN** API request fails
- **THEN** system logs error message and status

### Requirement: Track token usage

The system SHALL track prompt and completion tokens per request.

#### Scenario: Record token counts
- **WHEN** provider returns usage information
- **THEN** system stores prompt_tokens and completion_tokens

#### Scenario: Update key quota usage
- **WHEN** request completes
- **THEN** system updates api_key.used_quota with token count

### Requirement: Query usage statistics

The system SHALL provide aggregated usage statistics.

#### Scenario: Get overall stats
- **WHEN** admin requests usage stats
- **THEN** system returns total requests, tokens, and success rate for time period

#### Scenario: Get stats by provider
- **WHEN** admin filters by provider
- **THEN** system returns stats specific to that provider

#### Scenario: Get stats by API key
- **WHEN** admin filters by API key
- **THEN** system returns stats specific to that key

### Requirement: Query usage logs

The system SHALL provide detailed usage logs.

#### Scenario: List recent logs
- **WHEN** admin views usage logs page
- **THEN** system shows recent requests with timestamp, model, tokens, latency

#### Scenario: Filter logs by time range
- **WHEN** admin specifies date range
- **THEN** system shows logs within that range

#### Scenario: Filter logs by model
- **WHEN** admin filters by model name
- **THEN** system shows logs for that model only

### Requirement: Dashboard metrics

The system SHALL provide aggregated metrics for dashboard display.

#### Scenario: Daily request trend
- **WHEN** dashboard loads
- **THEN** system provides request counts for last 7 days

#### Scenario: Provider distribution
- **WHEN** dashboard loads
- **THEN** system provides request distribution by provider

#### Scenario: Model usage ranking
- **WHEN** dashboard loads
- **THEN** system provides top models by request count
