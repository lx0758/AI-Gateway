package provider

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestOpenAIUsage_Total(t *testing.T) {
	tests := []struct {
		name     string
		usage    openAIUsage
		expected int
	}{
		{
			name:     "empty usage",
			usage:    openAIUsage{},
			expected: 0,
		},
		{
			name: "total_tokens only",
			usage: openAIUsage{
				TotalTokens: 100,
			},
			expected: 100,
		},
		{
			name: "with reasoning tokens",
			usage: openAIUsage{
				TotalTokens: 100,
				CompletionTokensDetails: struct {
					ReasoningTokens int `json:"reasoning_tokens"`
				}{
					ReasoningTokens: 50,
				},
			},
			expected: 150,
		},
		{
			name: "full usage",
			usage: openAIUsage{
				PromptTokens:     50,
				CompletionTokens: 30,
				TotalTokens:      80,
				CompletionTokensDetails: struct {
					ReasoningTokens int `json:"reasoning_tokens"`
				}{
					ReasoningTokens: 20,
				},
			},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.usage.total(); got != tt.expected {
				t.Errorf("openAIUsage.total() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCopyOpenAIStreaming_TokenCounting(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	tokens := m.copyOpenAIStreaming(dst, strings.NewReader(openAISSE))

	if tokens != 150 {
		t.Errorf("copyOpenAIStreaming() tokens = %v, want 150", tokens)
	}
}

func TestCopyOpenAIStreaming_WithReasoning(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150,"completion_tokens_details":{"reasoning_tokens":30}}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	tokens := m.copyOpenAIStreaming(dst, strings.NewReader(openAISSE))

	if tokens != 180 {
		t.Errorf("copyOpenAIStreaming() tokens = %v, want 180", tokens)
	}
}

func TestCopyOpenAIResponse_TokenCounting(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAIResp := map[string]interface{}{
		"id":      "chatcmpl-xxx",
		"object":  "chat.completion",
		"created": 1234567890,
		"model":   "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello!",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     100,
			"completion_tokens": 50,
			"total_tokens":      150,
		},
	}

	respBytes, _ := json.Marshal(openAIResp)
	dst := &bytes.Buffer{}
	tokens := m.copyOpenAIResponse(dst, bytes.NewReader(respBytes))

	if tokens != 150 {
		t.Errorf("copyOpenAIResponse() tokens = %v, want 150", tokens)
	}
}

func TestExtractSystemContent(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil system",
			input:    nil,
			expected: "",
		},
		{
			name:     "string system",
			input:    "You are a helpful assistant.",
			expected: "You are a helpful assistant.",
		},
		{
			name: "block system",
			input: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "You are helpful.",
				},
				map[string]interface{}{
					"type": "text",
					"text": "Be concise.",
				},
			},
			expected: "You are helpful.\nBe concise.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.extractSystemContent(tt.input); got != tt.expected {
				t.Errorf("extractSystemContent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConvertAnthropicMessageToOpenAI_Text(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	anthropicMsg := map[string]interface{}{
		"role":    "user",
		"content": "Hello!",
	}

	results := m.convertAnthropicMessageToOpenAI(anthropicMsg)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]

	if result["role"] != "user" {
		t.Errorf("Expected role 'user', got %v", result["role"])
	}

	if result["content"] != "Hello!" {
		t.Errorf("Expected content 'Hello!', got %v", result["content"])
	}
}

func TestConvertAnthropicMessageToOpenAI_Blocks(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	anthropicMsg := map[string]interface{}{
		"role": "user",
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": "What's in this image?",
			},
			map[string]interface{}{
				"type": "image",
				"source": map[string]interface{}{
					"type":       "base64",
					"media_type": "image/png",
					"data":       "iVBORw0KGgo=",
				},
			},
		},
	}

	results := m.convertAnthropicMessageToOpenAI(anthropicMsg)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]

	content, ok := result["content"].([]map[string]interface{})
	if !ok {
		t.Fatalf("Expected content to be array, got %T", result["content"])
	}

	if len(content) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(content))
	}

	if content[0]["type"] != "text" {
		t.Errorf("Expected first block to be text")
	}

	if content[1]["type"] != "image_url" {
		t.Errorf("Expected second block to be image_url")
	}
}

func TestConvertAnthropicMessageToOpenAI_ToolUse(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	anthropicMsg := map[string]interface{}{
		"role": "assistant",
		"content": []interface{}{
			map[string]interface{}{
				"type":  "tool_use",
				"id":    "toolu_xxx",
				"name":  "get_weather",
				"input": map[string]interface{}{"location": "SF"},
			},
		},
	}

	results := m.convertAnthropicMessageToOpenAI(anthropicMsg)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]

	toolCalls, ok := result["tool_calls"].([]map[string]interface{})
	if !ok || len(toolCalls) == 0 {
		t.Fatalf("Expected tool_calls array")
	}

	if toolCalls[0]["id"] != "toolu_xxx" {
		t.Errorf("Expected tool id 'toolu_xxx', got %v", toolCalls[0]["id"])
	}

	fn := toolCalls[0]["function"].(map[string]interface{})
	if fn["name"] != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got %v", fn["name"])
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(fn["arguments"].(string)), &args); err != nil {
		t.Fatalf("Failed to parse arguments: %v", err)
	}

	if args["location"] != "SF" {
		t.Errorf("Expected location 'SF', got %v", args["location"])
	}
}

func TestConvertAnthropicToolResultToOpenAI(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	anthropicBlock := map[string]interface{}{
		"type":        "tool_result",
		"tool_use_id": "toolu_xxx",
		"content":     "Temperature: 72F",
	}

	result := m.convertAnthropicToolResultToOpenAI(anthropicBlock)

	if result["role"] != "tool" {
		t.Errorf("Expected role 'tool', got %v", result["role"])
	}

	if result["tool_call_id"] != "toolu_xxx" {
		t.Errorf("Expected tool_call_id 'toolu_xxx', got %v", result["tool_call_id"])
	}

	if result["content"] != "Temperature: 72F" {
		t.Errorf("Expected content 'Temperature: 72F', got %v", result["content"])
	}
}

func TestConvertAnthropicToolToOpenAI(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	anthropicTool := map[string]interface{}{
		"name":        "get_weather",
		"description": "Get weather info",
		"input_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"location": map[string]interface{}{"type": "string"},
			},
		},
	}

	result := m.convertAnthropicToolToOpenAI(anthropicTool)

	if result["type"] != "function" {
		t.Errorf("Expected type 'function', got %v", result["type"])
	}

	fn := result["function"].(map[string]interface{})
	if fn["name"] != "get_weather" {
		t.Errorf("Expected name 'get_weather', got %v", fn["name"])
	}

	if fn["description"] != "Get weather info" {
		t.Errorf("Expected description 'Get weather info', got %v", fn["description"])
	}

	if fn["parameters"] == nil {
		t.Errorf("Expected parameters to be set")
	}
}

func TestConvertOpenAIResponseToAnthropic(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAIResp := map[string]interface{}{
		"id":      "chatcmpl-xxx",
		"object":  "chat.completion",
		"created": 1234567890,
		"model":   "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello!",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     100,
			"completion_tokens": 50,
			"total_tokens":      150,
		},
	}

	respBytes, _ := json.Marshal(openAIResp)
	anthropicBytes, tokens, err := m.convertOpenAIResponseToAnthropic(respBytes)
	if err != nil {
		t.Fatalf("convertOpenAIResponseToAnthropic() error = %v", err)
	}

	if tokens != 150 {
		t.Errorf("Expected tokens 150, got %v", tokens)
	}

	var anthropicResp map[string]interface{}
	if err := json.Unmarshal(anthropicBytes, &anthropicResp); err != nil {
		t.Fatalf("Failed to parse Anthropic response: %v", err)
	}

	if anthropicResp["type"] != "message" {
		t.Errorf("Expected type 'message', got %v", anthropicResp["type"])
	}

	if anthropicResp["stop_reason"] != "end_turn" {
		t.Errorf("Expected stop_reason 'end_turn', got %v", anthropicResp["stop_reason"])
	}

	usage := anthropicResp["usage"].(map[string]interface{})
	if usage["input_tokens"].(float64) != 100 {
		t.Errorf("Expected input_tokens 100, got %v", usage["input_tokens"])
	}
}

func TestConvertOpenAIResponseToAnthropic_ToolCalls(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAIResp := map[string]interface{}{
		"id":    "chatcmpl-xxx",
		"model": "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"index": 0,
				"message": map[string]interface{}{
					"role": "assistant",
					"tool_calls": []interface{}{
						map[string]interface{}{
							"id":   "call_xxx",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "get_weather",
								"arguments": `{"location":"SF"}`,
							},
						},
					},
				},
				"finish_reason": "tool_calls",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     100,
			"completion_tokens": 50,
			"total_tokens":      150,
		},
	}

	respBytes, _ := json.Marshal(openAIResp)
	anthropicBytes, _, err := m.convertOpenAIResponseToAnthropic(respBytes)
	if err != nil {
		t.Fatalf("convertOpenAIResponseToAnthropic() error = %v", err)
	}

	var anthropicResp map[string]interface{}
	if err := json.Unmarshal(anthropicBytes, &anthropicResp); err != nil {
		t.Fatalf("Failed to parse Anthropic response: %v", err)
	}

	if anthropicResp["stop_reason"] != "tool_use" {
		t.Errorf("Expected stop_reason 'tool_use', got %v", anthropicResp["stop_reason"])
	}

	content := anthropicResp["content"].([]interface{})
	toolUse := content[0].(map[string]interface{})

	if toolUse["type"] != "tool_use" {
		t.Errorf("Expected type 'tool_use', got %v", toolUse["type"])
	}

	if toolUse["id"] != "call_xxx" {
		t.Errorf("Expected id 'call_xxx', got %v", toolUse["id"])
	}

	if toolUse["name"] != "get_weather" {
		t.Errorf("Expected name 'get_weather', got %v", toolUse["name"])
	}
}

func TestStreamOpenAIToAnthropic_Text(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	tokens := m.streamOpenAIToAnthropic(strings.NewReader(openAISSE), dst, "gpt-4")

	if tokens != 150 {
		t.Errorf("streamOpenAIToAnthropic() tokens = %v, want 150", tokens)
	}

	output := dst.String()

	if !strings.Contains(output, `event: message_start`) {
		t.Errorf("Expected message_start event in output")
	}

	if !strings.Contains(output, `"type":"text_delta"`) {
		t.Errorf("Expected text_delta in output")
	}

	if !strings.Contains(output, `"stop_reason":"end_turn"`) {
		t.Errorf("Expected stop_reason 'end_turn' in output")
	}
}

func TestStreamOpenAIToAnthropic_Thinking(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"reasoning_content":"Let me think..."}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	m.streamOpenAIToAnthropic(strings.NewReader(openAISSE), dst, "gpt-4")

	output := dst.String()

	if !strings.Contains(output, `"type":"thinking"`) {
		t.Errorf("Expected thinking block in output")
	}

	if !strings.Contains(output, `"type":"thinking_delta"`) {
		t.Errorf("Expected thinking_delta in output")
	}
}

func TestStreamOpenAIToAnthropic_ToolCalls(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_xxx","type":"function","function":{"name":"get_weather","arguments":"{\"loc"}}]}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"ation\":\"SF\"}"}}]}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	m.streamOpenAIToAnthropic(strings.NewReader(openAISSE), dst, "gpt-4")

	output := dst.String()

	if !strings.Contains(output, `"type":"tool_use"`) {
		t.Errorf("Expected tool_use block in output")
	}

	if !strings.Contains(output, `"id":"call_xxx"`) {
		t.Errorf("Expected tool id 'call_xxx' in output")
	}

	if !strings.Contains(output, `"type":"input_json_delta"`) {
		t.Errorf("Expected input_json_delta in output")
	}

	if !strings.Contains(output, `"stop_reason":"tool_use"`) {
		t.Errorf("Expected stop_reason 'tool_use' in output")
	}
}

func TestCopyOpenAIStreaming_WithReasoningTokens(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	// Test case: usage contains reasoning_tokens
	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150,"completion_tokens_details":{"reasoning_tokens":30}}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	tokens := m.copyOpenAIStreaming(dst, strings.NewReader(openAISSE))

	// Total: 150 + 30 (reasoning) = 180
	if tokens != 180 {
		t.Errorf("copyOpenAIStreaming() tokens = %v, want 180", tokens)
	}
}

func TestCopyOpenAIStreaming_WithCachedTokens(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	// Test case: usage contains cached_tokens
	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150,"prompt_tokens_details":{"cached_tokens":50}}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	tokens := m.copyOpenAIStreaming(dst, strings.NewReader(openAISSE))

	// Total tokens from usage.total() = TotalTokens + ReasoningTokens = 150 + 0 = 150
	if tokens != 150 {
		t.Errorf("copyOpenAIStreaming() tokens = %v, want 150", tokens)
	}
}

func TestCopyOpenAIResponse_WithReasoningTokens(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	openAIResp := map[string]interface{}{
		"id":      "chatcmpl-xxx",
		"object":  "chat.completion",
		"created": 1234567890,
		"model":   "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello!",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     100,
			"completion_tokens": 50,
			"total_tokens":      150,
			"completion_tokens_details": map[string]interface{}{
				"reasoning_tokens": 30,
			},
		},
	}

	respBytes, _ := json.Marshal(openAIResp)
	dst := &bytes.Buffer{}
	tokens := m.copyOpenAIResponse(dst, bytes.NewReader(respBytes))

	// Total: 150 + 30 (reasoning) = 180
	if tokens != 180 {
		t.Errorf("copyOpenAIResponse() tokens = %v, want 180", tokens)
	}
}

func TestStreamOpenAIToAnthropic_WithReasoningTokens(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	// Test case: usage contains reasoning_tokens in final chunk
	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150,"completion_tokens_details":{"reasoning_tokens":30}}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	tokens := m.streamOpenAIToAnthropic(strings.NewReader(openAISSE), dst, "gpt-4")

	// Total: 150 + 30 (reasoning) = 180
	if tokens != 180 {
		t.Errorf("streamOpenAIToAnthropic() tokens = %v, want 180", tokens)
	}

	output := dst.String()

	// Verify message_delta contains output_tokens
	if !strings.Contains(output, `"output_tokens":180`) {
		t.Errorf("Expected output_tokens 180 in message_delta, got: %s", output)
	}
}

func TestStreamOpenAIToAnthropic_ThinkingAndText(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	// Test case: thinking followed by text content
	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"reasoning_content":"Let me think..."}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"The answer is 42."}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":100,"completion_tokens":50,"total_tokens":150}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	m.streamOpenAIToAnthropic(strings.NewReader(openAISSE), dst, "gpt-4")

	output := dst.String()

	// Verify thinking block is created
	if !strings.Contains(output, `"type":"thinking"`) {
		t.Errorf("Expected thinking block in output")
	}

	// Verify thinking_delta is sent
	if !strings.Contains(output, `"thinking":"Let me think..."`) {
		t.Errorf("Expected thinking content in output")
	}

	// Verify content_block_stop is called for thinking before text
	if !strings.Contains(output, `"type":"content_block_stop"`) {
		t.Errorf("Expected content_block_stop in output")
	}

	// Verify text block follows
	if !strings.Contains(output, `"type":"text"`) {
		t.Errorf("Expected text block in output")
	}

	// Verify text_delta is sent
	if !strings.Contains(output, `"text":"The answer is 42."`) {
		t.Errorf("Expected text content in output")
	}
}

func TestStreamOpenAIToAnthropic_UsageWithReasoning(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	// Test case: verify output_tokens calculation with reasoning_tokens
	openAISSE := `data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"reasoning_content":"Thinking..."}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Answer."}}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":200,"completion_tokens":80,"total_tokens":280,"completion_tokens_details":{"reasoning_tokens":20}}}

data: [DONE]
`

	dst := &bytes.Buffer{}
	tokens := m.streamOpenAIToAnthropic(strings.NewReader(openAISSE), dst, "gpt-4")

	// Total: 280 + 20 (reasoning) = 300
	if tokens != 300 {
		t.Errorf("streamOpenAIToAnthropic() tokens = %v, want 300", tokens)
	}

	output := dst.String()

	// Verify message_delta contains correct output_tokens
	if !strings.Contains(output, `"output_tokens":300`) {
		t.Errorf("Expected output_tokens 300 in message_delta, got: %s", output)
	}
}

func TestConvertOpenAIResponseToAnthropic_WithReasoningContent(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	// Test case: OpenAI response with reasoning_content (some providers add this)
	openAIResp := map[string]interface{}{
		"id":    "chatcmpl-xxx",
		"model": "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"index": 0,
				"message": map[string]interface{}{
					"role":              "assistant",
					"content":           "The answer is 42.",
					"reasoning_content": "Let me analyze this...",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     100,
			"completion_tokens": 50,
			"total_tokens":      150,
		},
	}

	respBytes, _ := json.Marshal(openAIResp)
	anthropicBytes, tokens, err := m.convertOpenAIResponseToAnthropic(respBytes)
	if err != nil {
		t.Fatalf("convertOpenAIResponseToAnthropic() error = %v", err)
	}

	if tokens != 150 {
		t.Errorf("Expected tokens 150, got %v", tokens)
	}

	var anthropicResp map[string]interface{}
	if err := json.Unmarshal(anthropicBytes, &anthropicResp); err != nil {
		t.Fatalf("Failed to parse Anthropic response: %v", err)
	}

	content := anthropicResp["content"].([]interface{})

	// Should contain both thinking and text blocks
	if len(content) < 1 {
		t.Fatalf("Expected at least one content block")
	}

	// First block should be thinking (reasoning_content)
	firstBlock := content[0].(map[string]interface{})
	if firstBlock["type"] == "thinking" {
		if firstBlock["thinking"] != "Let me analyze this..." {
			t.Errorf("Expected thinking content, got %v", firstBlock["thinking"])
		}
	}
}

func TestConvertOpenAIResponseToAnthropic_EmptyContent(t *testing.T) {
	m := &OpenAIProvider{cfg: &Config{}}

	// Test case: OpenAI response with empty content (tool calls only)
	openAIResp := map[string]interface{}{
		"id":    "chatcmpl-xxx",
		"model": "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"index": 0,
				"message": map[string]interface{}{
					"role":       "assistant",
					"content":    nil,
					"tool_calls": []interface{}{},
				},
				"finish_reason": "tool_calls",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     100,
			"completion_tokens": 50,
			"total_tokens":      150,
		},
	}

	respBytes, _ := json.Marshal(openAIResp)
	anthropicBytes, _, err := m.convertOpenAIResponseToAnthropic(respBytes)
	if err != nil {
		t.Fatalf("convertOpenAIResponseToAnthropic() error = %v", err)
	}

	var anthropicResp map[string]interface{}
	if err := json.Unmarshal(anthropicBytes, &anthropicResp); err != nil {
		t.Fatalf("Failed to parse Anthropic response: %v", err)
	}

	// Should have empty content array
	content := anthropicResp["content"].([]interface{})
	if len(content) != 0 {
		t.Errorf("Expected empty content array, got %v", content)
	}
}

const openai_stream_1 = `
data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"用户"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"想"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"让我"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"执行"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"一个"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"工具"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"来"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"演示"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"。"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"我应该"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"执行"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"一个"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"简单的"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"命令"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"来"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"展示"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"工具"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"功能"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"。"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"让我"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"运行"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"一个"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"简单的"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"bash"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"命令"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"来"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"显示"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"当前"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"目录"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"的信息"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"。"}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"tool_calls":[{"id":"call_417fb00c5b2d4aada7fa6665","index":0,"type":"function","function":{"name":"bash","arguments":"{\"command\":\"echo \\\"Hello from opencode!\\\" && date\",\"description\":\"Display greeting and current date\"}"}}]}}]}

data: {"id":"20260331220108527fb37f4c4c4549","created":1774965668,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"finish_reason":"tool_calls","delta":{"role":"assistant","content":""}}],"usage":{"prompt_tokens":10984,"completion_tokens":60,"total_tokens":11044,"prompt_tokens_details":{"cached_tokens":10560}}}`

const openai_stream_2 = `
data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"用户"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"要求"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"执行"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"一个"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"工具"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"试试"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"。"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"我"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"刚刚"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"成功"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"执行"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"了"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":" bash"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":" "}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"工"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"具"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"，"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"显示了"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"问候"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"语"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"和"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"当前"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"日期"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"时间"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"。"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"执行"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"成功了"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"。"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"工具"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"执行"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"成功"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"！"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"当前"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"时间是"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":" "}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"202"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"6"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"年"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"3"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"月"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"31"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"日"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":" "}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"22"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":":"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"01"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":":"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"09"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"delta":{"role":"assistant","content":"。"}}]}

data: {"id":"202603312201111dd293dcfacb401c","created":1774965672,"object":"chat.completion.chunk","model":"glm-5","choices":[{"index":0,"finish_reason":"stop","delta":{"role":"assistant","content":""}}],"usage":{"prompt_tokens":11036,"completion_tokens":51,"total_tokens":11087,"prompt_tokens_details":{"cached_tokens":10944}}}`
