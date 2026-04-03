## MODIFIED Requirements

### Requirement: Record API call usage

The system SHALL record each API call to `usage_logs` table with the following information:
- `key_id`: The API key used for the request
- `provider_id`: The provider that handled the request
- `model`: The virtual model name (alias) requested by user
- `actual_model`: The actual model name used by the provider
- `client_ip`: The trusted original client IP address (string)
- `forwarded_chain`: The complete X-Forwarded-For header value (string)
- `total_tokens`: Total tokens consumed (int64)
- `latency_ms`: Request latency in milliseconds (int64)
- `status`: "success" or "error"
- `error_msg`: Error message if failed, empty if success

#### Scenario: Successful OpenAI request

- **WHEN** a successful OpenAI-compatible API request is completed
- **THEN** system records a usage log with status="success", total_tokens from response, latency_ms calculated, client_ip from trusted source, and forwarded_chain from X-Forwarded-For header

#### Scenario: Failed Anthropic request

- **WHEN** an Anthropic API request fails with an error
- **THEN** system records a usage log with status="error", error_msg containing the error, total_tokens=0, and client_ip/forwarded_chain captured before the error occurred

#### Scenario: Request through proxy

- **WHEN** a request passes through a trusted proxy with X-Forwarded-For header
- **THEN** system records client_ip as the original client IP (parsed from header) and forwarded_chain as the complete header value