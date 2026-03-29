## ADDED Requirements

### Requirement: Transform OpenAI request to Anthropic format

The system SHALL convert OpenAI Chat Completions request format to Anthropic Messages API format when routing to Anthropic providers.

#### Scenario: Basic message transformation
- **WHEN** request contains messages with system role in messages array
- **THEN** system extracts system message to top-level `system` parameter

#### Scenario: Max tokens handling
- **WHEN** OpenAI request has no `max_tokens` parameter
- **THEN** system sets default `max_tokens` value for Anthropic API

#### Scenario: Tool definitions transformation
- **WHEN** request contains `tools` array in OpenAI format
- **THEN** system converts to Anthropic `tools` format with `input_schema`

### Requirement: Transform Anthropic response to OpenAI format

The system SHALL convert Anthropic Messages API response format to OpenAI Chat Completions format.

#### Scenario: Basic response transformation
- **WHEN** Anthropic returns response with `content` blocks
- **THEN** system converts to OpenAI `choices[].message` format

#### Scenario: Tool use transformation
- **WHEN** Anthropic returns `tool_use` content block
- **THEN** system converts to OpenAI `tool_calls` format with `arguments` as JSON string

#### Scenario: Stop reason mapping
- **WHEN** Anthropic returns `stop_reason` of "end_turn"
- **THEN** system maps to OpenAI `finish_reason` of "stop"

### Requirement: Transform streaming SSE events

The system SHALL convert streaming SSE events between formats in real-time.

#### Scenario: Anthropic to OpenAI stream transformation
- **WHEN** Anthropic sends `content_block_delta` event with text delta
- **THEN** system converts to OpenAI format `data: {"choices":[{"delta":{"content":"..."}}]}`

#### Scenario: Stream completion
- **WHEN** Anthropic sends `message_stop` event
- **THEN** system sends `data: [DONE]` to client

### Requirement: Handle OpenAI-compatible providers

The system SHALL pass through requests to OpenAI-compatible providers without transformation.

#### Scenario: Direct passthrough
- **WHEN** provider has `api_type` of "openai"
- **THEN** system forwards request as-is without format conversion
