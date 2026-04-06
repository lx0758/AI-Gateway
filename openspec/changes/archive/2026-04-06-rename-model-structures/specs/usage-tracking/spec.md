## MODIFIED Requirements

### Requirement: Record API call usage

The system SHALL record each API call to `model_logs` table with the following information:
- `key_id`: The API key used for the request
- `provider_id`: The provider that handled the request
- `model`: The virtual model name requested by user
- `actual_model`: The actual model name used by the provider
- `total_tokens`: Total tokens consumed (int64)
- `latency_ms`: Request latency in milliseconds (int64)
- `status`: "success" or "error"
- `error_msg`: Error message if failed, empty if success

#### Scenario: Successful OpenAI request

- **WHEN** a successful OpenAI-compatible API request is completed
- **THEN** system records a model log with status="success", total_tokens from response, and latency_ms calculated

#### Scenario: Failed Anthropic request

- **WHEN** an Anthropic API request fails with an error
- **THEN** system records a model log with status="error", error_msg containing the error, and total_tokens=0

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

### Requirement: Display usage statistics on dashboard

The system SHALL display usage statistics on the dashboard page including:
- Total tokens consumed
- Average latency
- Provider statistics table

#### Scenario: View dashboard

- **WHEN** user opens the dashboard page
- **THEN** system displays total requests, today requests, active providers, active keys, total tokens, and average latency

### Requirement: Display key statistics on usage page

The system SHALL display API key statistics on the usage page.

#### Scenario: View key stats on usage page

- **WHEN** user opens the usage page
- **THEN** system displays a table with each key's name, call count, total tokens, and average latency
