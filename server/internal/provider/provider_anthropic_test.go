package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestAnthropicUsage_Total(t *testing.T) {
	tests := []struct {
		name     string
		usage    anthropicUsage
		expected int
	}{
		{
			name:     "empty usage",
			usage:    anthropicUsage{},
			expected: 0,
		},
		{
			name: "input only",
			usage: anthropicUsage{
				InputTokens: 100,
			},
			expected: 100,
		},
		{
			name: "output only",
			usage: anthropicUsage{
				OutputTokens: 50,
			},
			expected: 50,
		},
		{
			name: "both tokens",
			usage: anthropicUsage{
				InputTokens:  100,
				OutputTokens: 50,
			},
			expected: 150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.usage.total(); got != tt.expected {
				t.Errorf("anthropicUsage.total() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCopyAnthropicStreaming_TokenCounting(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	anthropicSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx","type":"message","role":"assistant","usage":{"input_tokens":100,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":0,"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}
`

	dst := &bytes.Buffer{}
	tokens := m.copyAnthropicStreaming(dst, strings.NewReader(anthropicSSE))

	expectedTokens := 150
	if tokens != expectedTokens {
		t.Errorf("copyAnthropicStreaming() tokens = %v, want %v", tokens, expectedTokens)
	}
}

func TestStreamAnthropicToOpenAI_TokenCounting(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	anthropicSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx","type":"message","role":"assistant","usage":{"input_tokens":100,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":0,"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}
`

	dst := &bytes.Buffer{}
	tokens := m.streamAnthropicToOpenAI(strings.NewReader(anthropicSSE), dst, "claude-3-sonnet")

	expectedTokens := 150
	if tokens != expectedTokens {
		t.Errorf("streamAnthropicToOpenAI() tokens = %v, want %v", tokens, expectedTokens)
	}
}

func TestStreamAnthropicToOpenAI_ToolCalls(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	anthropicSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx","type":"message","role":"assistant","usage":{"input_tokens":100,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_xxx","name":"get_weather","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"location\":"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"SF\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"input_tokens":0,"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}
`

	dst := &bytes.Buffer{}
	tokens := m.streamAnthropicToOpenAI(strings.NewReader(anthropicSSE), dst, "claude-3-sonnet")

	if tokens != 150 {
		t.Errorf("streamAnthropicToOpenAI() tokens = %v, want 150", tokens)
	}

	output := dst.String()

	if !strings.Contains(output, `"id":"toolu_xxx"`) {
		t.Errorf("Expected tool id 'toolu_xxx' in output, got: %s", output)
	}

	if !strings.Contains(output, `"name":"get_weather"`) {
		t.Errorf("Expected tool name 'get_weather' in output, got: %s", output)
	}

	if !strings.Contains(output, `"arguments":"{\"location\":`) {
		t.Errorf("Expected arguments in output, got: %s", output)
	}

	if !strings.Contains(output, `"finish_reason":"tool_calls"`) {
		t.Errorf("Expected finish_reason 'tool_calls' in output, got: %s", output)
	}
}

func TestStreamAnthropicToOpenAI_Thinking(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	anthropicSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx","type":"message","role":"assistant","usage":{"input_tokens":100,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Let me think about this..."}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"Hello!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":0,"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}
`

	dst := &bytes.Buffer{}
	tokens := m.streamAnthropicToOpenAI(strings.NewReader(anthropicSSE), dst, "claude-3-sonnet")

	if tokens != 150 {
		t.Errorf("streamAnthropicToOpenAI() tokens = %v, want 150", tokens)
	}

	output := dst.String()

	if !strings.Contains(output, `"reasoning_content":"Let me think about this..."`) {
		t.Errorf("Expected reasoning_content in output, got: %s", output)
	}

	if !strings.Contains(output, `"content":"Hello!"`) {
		t.Errorf("Expected content 'Hello!' in output, got: %s", output)
	}
}

func TestStreamAnthropicToOpenAI_Usage(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	anthropicSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx","type":"message","role":"assistant","usage":{"input_tokens":100,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":0,"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}
`

	dst := &bytes.Buffer{}
	m.streamAnthropicToOpenAI(strings.NewReader(anthropicSSE), dst, "claude-3-sonnet")

	output := dst.String()

	if !strings.Contains(output, `"prompt_tokens":100`) {
		t.Errorf("Expected prompt_tokens 100 in output, got: %s", output)
	}

	if !strings.Contains(output, `"completion_tokens":50`) {
		t.Errorf("Expected completion_tokens 50 in output, got: %s", output)
	}

	if !strings.Contains(output, `"total_tokens":150`) {
		t.Errorf("Expected total_tokens 150 in output, got: %s", output)
	}
}

func TestStreamAnthropicToOpenAI_UsageWithOutputTokensAtStart(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	// Test case: message_start contains output_tokens (e.g., for cached thinking)
	anthropicSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx","type":"message","role":"assistant","usage":{"input_tokens":100,"output_tokens":20}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":0,"output_tokens":30}}

event: message_stop
data: {"type":"message_stop"}
`

	dst := &bytes.Buffer{}
	tokens := m.streamAnthropicToOpenAI(strings.NewReader(anthropicSSE), dst, "claude-3-sonnet")

	// Total: 100 (input) + 20 (initial output) + 0 (delta input) + 30 (delta output) = 150
	if tokens != 150 {
		t.Errorf("streamAnthropicToOpenAI() tokens = %v, want 150", tokens)
	}

	output := dst.String()

	if !strings.Contains(output, `"prompt_tokens":100`) {
		t.Errorf("Expected prompt_tokens 100 in output, got: %s", output)
	}

	if !strings.Contains(output, `"completion_tokens":50`) {
		t.Errorf("Expected completion_tokens 50 in output, got: %s", output)
	}
}

func TestStreamAnthropicToOpenAI_UsageWithCacheTokens(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	// Test case: message_delta contains additional input_tokens (e.g., cache read tokens)
	anthropicSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx","type":"message","role":"assistant","usage":{"input_tokens":500,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello!"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":100,"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}
`

	dst := &bytes.Buffer{}
	tokens := m.streamAnthropicToOpenAI(strings.NewReader(anthropicSSE), dst, "claude-3-sonnet")

	// Total: 500 (initial input) + 100 (delta input) + 0 (initial output) + 50 (delta output) = 650
	if tokens != 650 {
		t.Errorf("streamAnthropicToOpenAI() tokens = %v, want 650", tokens)
	}

	output := dst.String()

	if !strings.Contains(output, `"prompt_tokens":600`) {
		t.Errorf("Expected prompt_tokens 600 in output, got: %s", output)
	}

	if !strings.Contains(output, `"completion_tokens":50`) {
		t.Errorf("Expected completion_tokens 50 in output, got: %s", output)
	}
}

func TestConvertAnthropicResponseToOpenAI_WithThinking(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	anthropicResp := map[string]interface{}{
		"id":    "msg_xxx",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-3-sonnet",
		"content": []map[string]interface{}{
			{
				"type":     "thinking",
				"thinking": "Let me analyze this...",
			},
			{
				"type": "text",
				"text": "The answer is 42.",
			},
		},
		"stop_reason": "end_turn",
		"usage": map[string]interface{}{
			"input_tokens":  100,
			"output_tokens": 50,
		},
	}

	anthropicBytes, _ := json.Marshal(anthropicResp)
	openAIBytes, tokens, err := m.convertAnthropicResponseToOpenAI(anthropicBytes, "claude-3-sonnet")
	if err != nil {
		t.Fatalf("convertAnthropicResponseToOpenAI() error = %v", err)
	}

	if tokens != 150 {
		t.Errorf("Expected tokens 150, got %v", tokens)
	}

	var openAIResp map[string]interface{}
	if err := json.Unmarshal(openAIBytes, &openAIResp); err != nil {
		t.Fatalf("Failed to parse OpenAI response: %v", err)
	}

	choices := openAIResp["choices"].([]interface{})
	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})

	if message["content"] != "The answer is 42." {
		t.Errorf("Expected content 'The answer is 42.', got %v", message["content"])
	}

	if message["reasoning_content"] != "Let me analyze this..." {
		t.Errorf("Expected reasoning_content 'Let me analyze this...', got %v", message["reasoning_content"])
	}
}

func TestCopyAnthropicStreaming_WithOutputTokensAtStart(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	// Test case: message_start contains output_tokens
	anthropicSSE := `event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx","type":"message","role":"assistant","usage":{"input_tokens":100,"output_tokens":20}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":0,"output_tokens":30}}

event: message_stop
data: {"type":"message_stop"}
`

	dst := &bytes.Buffer{}
	tokens := m.copyAnthropicStreaming(dst, strings.NewReader(anthropicSSE))

	// Total: 100 (input) + 20 (initial output) + 30 (delta output) = 150
	if tokens != 150 {
		t.Errorf("copyAnthropicStreaming() tokens = %v, want 150", tokens)
	}
}

func TestConvertAnthropicResponseToOpenAI(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	anthropicResp := map[string]interface{}{
		"id":    "msg_xxx",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-3-sonnet",
		"content": []map[string]interface{}{
			{"type": "text", "text": "Hello, how can I help?"},
		},
		"stop_reason": "end_turn",
		"usage": map[string]interface{}{
			"input_tokens":  100,
			"output_tokens": 50,
		},
	}

	anthropicBytes, _ := json.Marshal(anthropicResp)
	openAIBytes, tokens, err := m.convertAnthropicResponseToOpenAI(anthropicBytes, "claude-3-sonnet")
	if err != nil {
		t.Fatalf("convertAnthropicResponseToOpenAI() error = %v", err)
	}

	if tokens != 150 {
		t.Errorf("convertAnthropicResponseToOpenAI() tokens = %v, want 150", tokens)
	}

	var openAIResp map[string]interface{}
	if err := json.Unmarshal(openAIBytes, &openAIResp); err != nil {
		t.Fatalf("Failed to parse OpenAI response: %v", err)
	}

	if openAIResp["object"] != "chat.completion" {
		t.Errorf("Expected object to be 'chat.completion', got %v", openAIResp["object"])
	}

	choices, ok := openAIResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Fatalf("Expected choices array with at least one element")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	if message["content"] != "Hello, how can I help?" {
		t.Errorf("Expected content 'Hello, how can I help?', got %v", message["content"])
	}
}

func TestConvertAnthropicResponseToOpenAI_ToolUse(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	anthropicResp := map[string]interface{}{
		"id":    "msg_xxx",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-3-sonnet",
		"content": []map[string]interface{}{
			{
				"type":  "tool_use",
				"id":    "toolu_xxx",
				"name":  "get_weather",
				"input": map[string]interface{}{"location": "San Francisco"},
			},
		},
		"stop_reason": "tool_use",
		"usage": map[string]interface{}{
			"input_tokens":  100,
			"output_tokens": 50,
		},
	}

	anthropicBytes, _ := json.Marshal(anthropicResp)
	openAIBytes, _, err := m.convertAnthropicResponseToOpenAI(anthropicBytes, "claude-3-sonnet")
	if err != nil {
		t.Fatalf("convertAnthropicResponseToOpenAI() error = %v", err)
	}

	var openAIResp map[string]interface{}
	if err := json.Unmarshal(openAIBytes, &openAIResp); err != nil {
		t.Fatalf("Failed to parse OpenAI response: %v", err)
	}

	choices, _ := openAIResp["choices"].([]interface{})
	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})

	toolCalls, ok := message["tool_calls"].([]interface{})
	if !ok || len(toolCalls) == 0 {
		t.Fatalf("Expected tool_calls array with at least one element")
	}

	toolCall := toolCalls[0].(map[string]interface{})
	fn := toolCall["function"].(map[string]interface{})

	if fn["name"] != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got %v", fn["name"])
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(fn["arguments"].(string)), &args); err != nil {
		t.Fatalf("Failed to parse arguments: %v", err)
	}

	if args["location"] != "San Francisco" {
		t.Errorf("Expected location 'San Francisco', got %v", args["location"])
	}

	if choice["finish_reason"] != "tool_calls" {
		t.Errorf("Expected finish_reason 'tool_calls', got %v", choice["finish_reason"])
	}
}

func TestConvertOpenAIToolResultToAnthropic(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	openAIMsg := map[string]interface{}{
		"role":         "tool",
		"tool_call_id": "call_xxx",
		"content":      "Temperature: 72F",
	}

	result := m.convertOpenAIToolResultToAnthropic(openAIMsg)

	if result["role"] != "user" {
		t.Errorf("Expected role 'user', got %v", result["role"])
	}

	content, ok := result["content"].([]map[string]interface{})
	if !ok || len(content) == 0 {
		t.Fatalf("Expected content array with at least one element")
	}

	block := content[0]
	if block["type"] != "tool_result" {
		t.Errorf("Expected type 'tool_result', got %v", block["type"])
	}

	if block["tool_use_id"] != "call_xxx" {
		t.Errorf("Expected tool_use_id 'call_xxx', got %v", block["tool_use_id"])
	}

	if block["content"] != "Temperature: 72F" {
		t.Errorf("Expected content 'Temperature: 72F', got %v", block["content"])
	}
}

func TestParseDataURL(t *testing.T) {
	m := &AnthropicProvider{cfg: &Config{}}

	tests := []struct {
		name        string
		url         string
		expectMedia string
		expectData  string
	}{
		{
			name:        "png data url",
			url:         "data:image/png;base64,iVBORw0KGgo=",
			expectMedia: "image/png",
			expectData:  "iVBORw0KGgo=",
		},
		{
			name:        "jpeg data url",
			url:         "data:image/jpeg;base64,/9j/4AAQSkZJ=",
			expectMedia: "image/jpeg",
			expectData:  "/9j/4AAQSkZJ=",
		},
		{
			name:        "webp data url",
			url:         "data:image/webp;base64,UklGRjIAAABXQ",
			expectMedia: "image/webp",
			expectData:  "UklGRjIAAABXQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mediaType, data := m.parseDataURL(tt.url)
			if mediaType != tt.expectMedia {
				t.Errorf("parseDataURL() mediaType = %v, want %v", mediaType, tt.expectMedia)
			}
			if data != tt.expectData {
				t.Errorf("parseDataURL() data = %v, want %v", data, tt.expectData)
			}
		})
	}
}

type mockFlusher struct {
	io.Writer
}

func (m *mockFlusher) Flush() {}

const anthropic_stream_1 = `
event: message_start
data: {"type": "message_start", "message": {"id": "msg_202603312157426c43542335a446c4", "type": "message", "role": "assistant", "model": "glm-5", "content": [], "stop_reason": null, "stop_sequence": null, "usage": {"input_tokens": 0, "output_tokens": 0}}}

event: ping
data: {"type": "ping"}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": ""}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "好的"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "，"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "我来"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "执行"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "一个"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "简单的"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "命令"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "："}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 0}

event: content_block_start
data: {"type": "content_block_start", "index": 1, "content_block": {"type": "tool_use", "id": "call_9cd06da089d2481291d99317", "name": "bash", "input": {}}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 1, "delta": {"type": "input_json_delta", "partial_json": "{\"command\":\"echo \\\"Hello, World!\\\"\",\"description\":\"Print Hello World message\"}"}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 1}

event: message_delta
data: {"type": "message_delta", "delta": {"stop_reason": "tool_use", "stop_sequence": null}, "usage": {"input_tokens": 8167, "output_tokens": 32, "cache_read_input_tokens": 2816, "server_tool_use": {"web_search_requests": 0}, "service_tier": "standard"}}

event: message_stop
data: {"type": "message_stop"}

`

const anthropic_stream_2 = `
event: message_start
data: {"type": "message_start", "message": {"id": "msg_20260331215745f372afac73dd4e9a", "type": "message", "role": "assistant", "model": "glm-5", "content": [], "stop_reason": null, "stop_sequence": null, "usage": {"input_tokens": 0, "output_tokens": 0}}}

event: ping
data: {"type": "ping"}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": ""}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "执行"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "成功"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "！"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "输"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "出了"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": " \""}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Hello"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": ","}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": " World"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "!\""}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "。"}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 0}

event: message_delta
data: {"type": "message_delta", "delta": {"stop_reason": "end_turn", "stop_sequence": null}, "usage": {"input_tokens": 79, "output_tokens": 12, "cache_read_input_tokens": 10944, "server_tool_use": {"web_search_requests": 0}, "service_tier": "standard"}}

event: message_stop
data: {"type": "message_stop"}

`

const anthropic_stream_3 = `
event: message_start
data: {"type":"message_start","message":{"id":"061b09e000a972a9716f209001a3a452","type":"message","role":"assistant","content":[],"model":"MiniMax-M2.5","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":2707,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"service_tier":"standard"}}

event: ping
data: {"type":"ping"}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"用户要求我执行一个工具试试。我可以尝试使用一些简单的工具来展示功能。让我先检查一下工作目录中有什么文件，然后尝试使用工具。\n\n让我先列出当前目录的文件。\n"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"44adca74917efbeb5fa2f28646e270693c1bebe4867747843b0a1330c3527642"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"call_function_sy2faeqracl4_1","name":"glob","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"pattern\": \"*\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"input_tokens":9195,"output_tokens":64,"cache_creation_input_tokens":0,"cache_read_input_tokens":1792}}

event: message_stop
data: {"type":"message_stop"}

`

const anthropic_stream_4 = `
event: message_start
data: {"type":"message_start","message":{"id":"061b09e53b545c6c28cc8f0fd9e89dbe","type":"message","role":"assistant","content":[],"model":"MiniMax-M2.5","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":2731,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"service_tier":"standard"}}

event: ping
data: {"type":"ping"}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"用户想让我执行一个工具试试。我刚才已经执行了glob工具来列出当前目录的文件。这是一个测试，看看工具是否能正常工作。\n\n结果显示当前目录下有一个文件：opencode.exe。\n"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"1fd3434c085735fd2039eadc7dd5dcd1f2987c5f7ec45038a9fc4bd3cebdf8c2"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"\n\n已执行工具测试。当前目录下有一个文件 `opencode.exe`。"}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":94,"output_tokens":52,"cache_creation_input_tokens":0,"cache_read_input_tokens":10944}}

event: message_stop
data: {"type":"message_stop"}

`