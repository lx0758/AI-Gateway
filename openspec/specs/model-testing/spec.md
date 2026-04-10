## ADDED Requirements

### Requirement: Provider model testing API

The system SHALL provide an API endpoint to test a specific provider model.

```
POST /api/v1/providers/:id/models/:model_id/test
```

#### Scenario: Test provider model with OpenAI protocol

- **WHEN** provider has OpenAIBaseURL configured
- **THEN** system executes test using OpenAI protocol
- **AND** returns test result with latency, tokens, response content

#### Scenario: Test provider model with Anthropic protocol

- **WHEN** provider has AnthropicBaseURL configured
- **THEN** system executes test using Anthropic protocol
- **AND** returns test result with latency, tokens, response content

#### Scenario: Test provider model with both protocols

- **WHEN** provider has both OpenAIBaseURL and AnthropicBaseURL configured
- **THEN** system executes two tests, one for each protocol
- **AND** returns both test results

#### Scenario: Test fails with error

- **WHEN** test request fails (connection error, timeout, API error)
- **THEN** system returns success=false with error message

#### Scenario: Provider not found

- **WHEN** provider ID does not exist
- **THEN** system returns 404 error

#### Scenario: Provider model not found

- **WHEN** model ID does not exist for the provider
- **THEN** system returns 404 error

### Requirement: Virtual model testing API

The system SHALL provide an API endpoint to test a virtual model (alias) with all its mappings.

```
POST /api/v1/models/:id/test
```

#### Scenario: Test virtual model with single mapping

- **WHEN** model has one enabled mapping
- **THEN** system executes test for that mapping
- **AND** returns test result with provider and model details

#### Scenario: Test virtual model with multiple mappings

- **WHEN** model has multiple enabled mappings
- **THEN** system executes tests for each mapping in weight order
- **AND** returns all test results sorted by weight

#### Scenario: Test virtual model with disabled mappings

- **WHEN** model has disabled mappings
- **THEN** system skips disabled mappings
- **AND** only tests enabled mappings

#### Scenario: Virtual model not found

- **WHEN** model ID does not exist
- **THEN** system returns 404 error

#### Scenario: Virtual model has no enabled mappings

- **WHEN** model has no enabled mappings
- **THEN** system returns empty tests array

### Requirement: Test execution behavior

#### Scenario: Test uses fixed message

- **WHEN** test is executed
- **THEN** system sends message "Hi" to the model
- **AND** sets max_tokens to 100
- **AND** sets stream to false

#### Scenario: Test measures latency

- **WHEN** test is executed
- **THEN** system records latency in milliseconds

#### Scenario: Test extracts token usage

- **WHEN** test response includes token usage
- **THEN** system extracts input_tokens and output_tokens

#### Scenario: Test extracts response content

- **WHEN** test succeeds
- **THEN** system extracts response text content

#### Scenario: Test timeout

- **WHEN** test request exceeds 30 seconds
- **THEN** system returns timeout error

### Requirement: Test code reuses existing provider logic

#### Scenario: Test uses httptest context

- **WHEN** test is executed
- **THEN** system creates gin.Context using httptest.NewRecorder and gin.CreateTestContext
- **AND** calls existing Provider.ExecuteOpenAIRequest or ExecuteAnthropicRequest
