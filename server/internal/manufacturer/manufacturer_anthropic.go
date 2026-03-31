package manufacturer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (u anthropicUsage) total() int {
	return u.InputTokens + u.OutputTokens
}

type AnthropicManufacturer struct {
	cfg *Config
}

func NewAnthropicManufacturer(cfg *Config) *AnthropicManufacturer {
	return &AnthropicManufacturer{cfg: cfg}
}

func (m *AnthropicManufacturer) Name() string {
	return m.cfg.ProviderName
}

func (m *AnthropicManufacturer) SyncModels(provider *model.Provider) ([]model.ProviderModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("Anthropic base URL is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", provider.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error: %s", string(body))
	}

	var result struct {
		Data []struct {
			ID            string `json:"id"`
			Type          string `json:"type"`
			DisplayName   string `json:"display_name"`
			CreatedAt     string `json:"created_at"`
			MaxInputToken int    `json:"max_input_tokens"`
			MaxTokens     int    `json:"max_tokens"`
			Capabilities  struct {
				ImageInput struct {
					Supported bool `json:"supported"`
				} `json:"image_input"`
				Thinking struct {
					Supported bool `json:"supported"`
				} `json:"thinking"`
			} `json:"capabilities"`
		} `json:"data"`
		HasMore bool `json:"has_more"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]model.ProviderModel, 0, len(result.Data))
	for _, m := range result.Data {
		displayName := m.DisplayName
		if displayName == "" {
			displayName = m.ID
		}

		supportsVision := m.Capabilities.ImageInput.Supported
		supportsTools := true

		models = append(models, model.ProviderModel{
			ProviderID:     provider.ID,
			ModelID:        m.ID,
			DisplayName:    displayName,
			OwnedBy:        "anthropic",
			ContextWindow:  m.MaxInputToken,
			MaxOutput:      m.MaxTokens,
			SupportsVision: supportsVision,
			SupportsTools:  supportsTools,
			SupportsStream: true,
			IsAvailable:    true,
			Source:         "sync",
		})
	}

	return models, nil
}

func (m *AnthropicManufacturer) ExecuteOpenAIRequest(c *gin.Context, pm *model.ProviderModel) (int, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return 0, err
	}

	var openAIReq struct {
		Model     string                   `json:"model"`
		MaxTokens int                      `json:"max_tokens"`
		Messages  []map[string]interface{} `json:"messages"`
		Tools     json.RawMessage          `json:"tools,omitempty"`
		Stream    bool                     `json:"stream"`
	}
	if err := json.Unmarshal(body, &openAIReq); err != nil {
		return 0, err
	}
	openAIReq.Model = pm.ModelID

	anthropicReq := m.convertOpenAIRequestToAnthropic(openAIReq)

	anthropicBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", m.cfg.BaseURL+"/messages", bytes.NewReader(anthropicBody))
	if err != nil {
		return 0, err
	}

	req.Header.Set("x-api-key", m.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": string(respBody)})
		return 0, nil
	}

	tokens := 0
	if m.isStreaming(resp) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		tokens = m.streamAnthropicToOpenAI(resp.Body, c.Writer, openAIReq.Model)
	} else {
		anthropicRespBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}
		openAIResp, tokens, err := m.convertAnthropicResponseToOpenAI(anthropicRespBody, openAIReq.Model)
		if err != nil {
			return tokens, err
		}
		c.Header("Content-Type", "application/json")
		c.Writer.Write(openAIResp)
	}
	return tokens, nil
}

func (m *AnthropicManufacturer) convertOpenAIRequestToAnthropic(openAIReq struct {
	Model     string                   `json:"model"`
	MaxTokens int                      `json:"max_tokens"`
	Messages  []map[string]interface{} `json:"messages"`
	Tools     json.RawMessage          `json:"tools,omitempty"`
	Stream    bool                     `json:"stream"`
}) map[string]interface{} {
	var systemContent string
	anthropicMessages := make([]map[string]interface{}, 0)

	for _, msg := range openAIReq.Messages {
		role, _ := msg["role"].(string)
		switch role {
		case "system":
			systemContent = m.extractSystemContent(msg["content"])
		case "tool":
			anthropicMessages = append(anthropicMessages, m.convertOpenAIToolResultToAnthropic(msg))
		default:
			anthropicMessages = append(anthropicMessages, m.convertOpenAIMessageToAnthropic(msg))
		}
	}

	anthropicReq := map[string]interface{}{
		"model":      openAIReq.Model,
		"max_tokens": openAIReq.MaxTokens,
		"messages":   anthropicMessages,
		"stream":     openAIReq.Stream,
	}
	if systemContent != "" {
		anthropicReq["system"] = systemContent
	}
	if openAIReq.Tools != nil {
		var tools []interface{}
		if err := json.Unmarshal(openAIReq.Tools, &tools); err == nil {
			anthropicTools := make([]map[string]interface{}, 0, len(tools))
			for _, tool := range tools {
				if t, ok := tool.(map[string]interface{}); ok {
					anthropicTools = append(anthropicTools, m.convertOpenAIToolToAnthropic(t))
				}
			}
			anthropicReq["tools"] = anthropicTools
		}
	}
	return anthropicReq
}

func (m *AnthropicManufacturer) extractSystemContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var texts []string
		for _, part := range v {
			if partMap, ok := part.(map[string]interface{}); ok {
				if partType, _ := partMap["type"].(string); partType == "text" {
					if text, _ := partMap["text"].(string); text != "" {
						texts = append(texts, text)
					}
				}
			}
		}
		if len(texts) > 0 {
			return strings.Join(texts, "\n")
		}
	}
	return ""
}

func (m *AnthropicManufacturer) convertOpenAIMessageToAnthropic(msg map[string]interface{}) map[string]interface{} {
	role, _ := msg["role"].(string)
	content := msg["content"]

	result := map[string]interface{}{
		"role": role,
	}

	switch v := content.(type) {
	case string:
		result["content"] = []map[string]interface{}{
			{"type": "text", "text": v},
		}
	case []interface{}:
		blocks := make([]map[string]interface{}, 0)
		for _, part := range v {
			if partMap, ok := part.(map[string]interface{}); ok {
				partType, _ := partMap["type"].(string)
				switch partType {
				case "text":
					blocks = append(blocks, map[string]interface{}{
						"type": "text",
						"text": partMap["text"],
					})
				case "image_url":
					imageURL, _ := partMap["image_url"].(map[string]interface{})
					if imageURL != nil {
						url, _ := imageURL["url"].(string)
						if strings.HasPrefix(url, "data:") {
							mediaType, data := m.parseDataURL(url)
							blocks = append(blocks, map[string]interface{}{
								"type": "image",
								"source": map[string]interface{}{
									"type":       "base64",
									"media_type": mediaType,
									"data":       data,
								},
							})
						} else {
							blocks = append(blocks, map[string]interface{}{
								"type": "image",
								"source": map[string]interface{}{
									"type": "url",
									"url":  url,
								},
							})
						}
					}
				}
			}
		}
		result["content"] = blocks
	default:
		if v != nil {
			result["content"] = []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("%v", v)},
			}
		}
	}

	if toolCalls, ok := msg["tool_calls"].([]interface{}); ok {
		blocks, _ := result["content"].([]map[string]interface{})
		for _, tc := range toolCalls {
			if tcMap, ok := tc.(map[string]interface{}); ok {
				toolUse := map[string]interface{}{
					"type": "tool_use",
					"id":   tcMap["id"],
				}
				if fn, ok := tcMap["function"].(map[string]interface{}); ok {
					toolUse["name"] = fn["name"]
					if args, _ := fn["arguments"].(string); args != "" {
						var input map[string]interface{}
						if json.Unmarshal([]byte(args), &input) == nil {
							toolUse["input"] = input
						} else {
							toolUse["input"] = args
						}
					}
				}
				blocks = append(blocks, toolUse)
			}
		}
		result["content"] = blocks
	}

	return result
}

func (m *AnthropicManufacturer) convertOpenAIToolResultToAnthropic(msg map[string]interface{}) map[string]interface{} {
	toolCallID, _ := msg["tool_call_id"].(string)
	content := msg["content"]

	var toolResultContent string
	switch v := content.(type) {
	case string:
		toolResultContent = v
	default:
		toolResultContent = fmt.Sprintf("%v", v)
	}

	return map[string]interface{}{
		"role": "user",
		"content": []map[string]interface{}{
			{
				"type":        "tool_result",
				"tool_use_id": toolCallID,
				"content":     toolResultContent,
			},
		},
	}
}

func (m *AnthropicManufacturer) convertOpenAIToolToAnthropic(tool map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"name": tool["name"],
	}
	if desc, ok := tool["description"].(string); ok {
		result["description"] = desc
	}
	if fn, ok := tool["function"].(map[string]interface{}); ok {
		if params, ok := fn["parameters"].(map[string]interface{}); ok {
			result["input_schema"] = params
		}
		if desc, ok := fn["description"].(string); ok {
			result["description"] = desc
		}
		if name, ok := fn["name"].(string); ok {
			result["name"] = name
		}
	}
	return result
}

func (m *AnthropicManufacturer) parseDataURL(url string) (mediaType, data string) {
	if !strings.HasPrefix(url, "data:") {
		return "", ""
	}
	url = strings.TrimPrefix(url, "data:")
	parts := strings.SplitN(url, ",", 2)
	if len(parts) != 2 {
		return "", ""
	}
	mediaType = parts[0]
	if strings.Contains(mediaType, ";") {
		mediaType = strings.Split(mediaType, ";")[0]
	}
	data = parts[1]
	return mediaType, data
}

func (m *AnthropicManufacturer) convertAnthropicResponseToOpenAI(anthropicResp []byte, model string) ([]byte, int, error) {
	var anthropic struct {
		ID         string                   `json:"id"`
		Type       string                   `json:"type"`
		Role       string                   `json:"role"`
		Model      string                   `json:"model"`
		Content    []map[string]interface{} `json:"content"`
		StopReason string                   `json:"stop_reason"`
		Usage      anthropicUsage           `json:"usage"`
	}
	if err := json.Unmarshal(anthropicResp, &anthropic); err != nil {
		return nil, 0, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	tokens := anthropic.Usage.total()

	textContent := ""
	reasoningContent := ""
	var toolCalls []map[string]interface{}

	for _, block := range anthropic.Content {
		blockType, _ := block["type"].(string)
		switch blockType {
		case "text":
			if text, ok := block["text"].(string); ok {
				textContent = text
			}
		case "thinking":
			if thinking, ok := block["thinking"].(string); ok {
				reasoningContent = thinking
			}
		case "tool_use":
			var argsStr string
			if input := block["input"]; input != nil {
				if inputBytes, err := json.Marshal(input); err == nil {
					argsStr = string(inputBytes)
				}
			}
			toolCalls = append(toolCalls, map[string]interface{}{
				"id":   block["id"],
				"type": "function",
				"function": map[string]interface{}{
					"name":      block["name"],
					"arguments": argsStr,
				},
			})
		}
	}

	finishReason := "stop"
	switch anthropic.StopReason {
	case "end_turn":
		finishReason = "stop"
	case "max_tokens":
		finishReason = "length"
	case "tool_use":
		finishReason = "tool_calls"
	case "stop_sequence":
		finishReason = "stop"
	}

	message := map[string]interface{}{
		"role":    "assistant",
		"content": textContent,
	}
	if reasoningContent != "" {
		message["reasoning_content"] = reasoningContent
	}
	if len(toolCalls) > 0 {
		message["tool_calls"] = toolCalls
	}

	openAIResp := map[string]interface{}{
		"id":      anthropic.ID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index":         0,
				"message":       message,
				"finish_reason": finishReason,
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     anthropic.Usage.InputTokens,
			"completion_tokens": anthropic.Usage.OutputTokens,
			"total_tokens":      tokens,
		},
	}

	result, err := json.Marshal(openAIResp)
	if err != nil {
		return nil, tokens, fmt.Errorf("failed to marshal OpenAI response: %w", err)
	}
	return result, tokens, nil
}

func (m *AnthropicManufacturer) streamAnthropicToOpenAI(src io.Reader, dst io.Writer, model string) int {
	reader := bufio.NewReader(src)
	messageID := fmt.Sprintf("chatcmpl-%s", m.generateID())
	tokens := 0
	inputTokens := 0
	outputTokens := 0
	created := time.Now().Unix()
	contentBuffer := ""
	toolCallsBuffer := make(map[int]map[string]interface{})

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event struct {
			Type         string                 `json:"type"`
			Index        int                    `json:"index"`
			ContentBlock map[string]interface{} `json:"content_block"`
			Delta        map[string]interface{} `json:"delta"`
			Message      map[string]interface{} `json:"message"`
			Usage        anthropicUsage         `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "message_start":
			if msg, ok := event.Message["usage"].(map[string]interface{}); ok {
				if it, ok := msg["input_tokens"].(float64); ok {
					inputTokens = int(it)
				}
				if it, ok := msg["output_tokens"].(float64); ok {
					outputTokens = int(it)
				}
				tokens += inputTokens + outputTokens
			}

		case "content_block_start":
			if event.ContentBlock != nil {
				blockType, _ := event.ContentBlock["type"].(string)
				switch blockType {
				case "tool_use":
					toolID, _ := event.ContentBlock["id"].(string)
					toolName, _ := event.ContentBlock["name"].(string)
					toolCallsBuffer[event.Index] = map[string]interface{}{
						"index": event.Index,
						"id":    toolID,
						"type":  "function",
						"function": map[string]interface{}{
							"name":      toolName,
							"arguments": "",
						},
					}
					m.writeOpenAISSE(dst, map[string]interface{}{
						"id":      messageID,
						"object":  "chat.completion.chunk",
						"created": created,
						"model":   model,
						"choices": []map[string]interface{}{
							{
								"index": 0,
								"delta": map[string]interface{}{
									"tool_calls": []map[string]interface{}{
										{
											"index": event.Index,
											"id":    toolID,
											"type":  "function",
											"function": map[string]interface{}{
												"name":      toolName,
												"arguments": "",
											},
										},
									},
								},
							},
						},
					})
				}
			}

		case "content_block_delta":
			if event.Delta != nil {
				deltaType, _ := event.Delta["type"].(string)
				switch deltaType {
				case "text_delta":
					if text, ok := event.Delta["text"].(string); ok {
						contentBuffer += text
						m.writeOpenAISSE(dst, map[string]interface{}{
							"id":      messageID,
							"object":  "chat.completion.chunk",
							"created": created,
							"model":   model,
							"choices": []map[string]interface{}{
								{
									"index": 0,
									"delta": map[string]interface{}{
										"content": text,
									},
								},
							},
						})
					}
				case "thinking_delta":
					if thinking, ok := event.Delta["thinking"].(string); ok {
						m.writeOpenAISSE(dst, map[string]interface{}{
							"id":      messageID,
							"object":  "chat.completion.chunk",
							"created": created,
							"model":   model,
							"choices": []map[string]interface{}{
								{
									"index": 0,
									"delta": map[string]interface{}{
										"reasoning_content": thinking,
									},
								},
							},
						})
					}
				case "input_json_delta":
					if partialJSON, ok := event.Delta["partial_json"].(string); ok {
						if tool, exists := toolCallsBuffer[event.Index]; exists {
							fn := tool["function"].(map[string]interface{})
							fn["arguments"] = fn["arguments"].(string) + partialJSON
							m.writeOpenAISSE(dst, map[string]interface{}{
								"id":      messageID,
								"object":  "chat.completion.chunk",
								"created": created,
								"model":   model,
								"choices": []map[string]interface{}{
									{
										"index": 0,
										"delta": map[string]interface{}{
											"tool_calls": []map[string]interface{}{
												{
													"index": event.Index,
													"function": map[string]interface{}{
														"arguments": partialJSON,
													},
												},
											},
										},
									},
								},
							})
						}
					}
				}
			}

		case "content_block_stop":
			// No action needed

		case "message_delta":
			inputTokens += event.Usage.InputTokens
			outputTokens += event.Usage.OutputTokens
			tokens += event.Usage.total()
			finishReason := "stop"
			if stopReason, ok := event.Delta["stop_reason"].(string); ok {
				switch stopReason {
				case "end_turn":
					finishReason = "stop"
				case "max_tokens":
					finishReason = "length"
				case "tool_use":
					finishReason = "tool_calls"
				}
			}
			m.writeOpenAISSE(dst, map[string]interface{}{
				"id":      messageID,
				"object":  "chat.completion.chunk",
				"created": created,
				"model":   model,
				"choices": []map[string]interface{}{
					{
						"index":         0,
						"delta":         map[string]interface{}{},
						"finish_reason": finishReason,
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     inputTokens,
					"completion_tokens": outputTokens,
					"total_tokens":      tokens,
				},
			})

		case "message_stop":
			fmt.Fprint(dst, "data: [DONE]\n\n")
		}
	}

	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}
	return tokens
}

func (m *AnthropicManufacturer) writeOpenAISSE(w io.Writer, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *AnthropicManufacturer) generateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 24)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

func (m *AnthropicManufacturer) isStreaming(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return len(resp.Header["Transfer-Encoding"]) > 0 ||
		(len(contentType) > 0 && len(contentType) >= 17 && contentType[:17] == "text/event-stream")
}

func (m *AnthropicManufacturer) ExecuteAnthropicRequest(c *gin.Context, pm *model.ProviderModel) (int, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return 0, err
	}

	bodyJson := map[string]interface{}{}
	if err := json.Unmarshal(body, &bodyJson); err != nil {
		return 0, err
	}
	bodyJson["model"] = pm.ModelID
	body, err = json.Marshal(bodyJson)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", m.cfg.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set("x-api-key", m.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": string(respBody)})
		return 0, nil
	}

	tokens := 0
	if m.isStreaming(resp) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		tokens = m.copyAnthropicStreaming(c.Writer, resp.Body)
	} else {
		c.Header("Content-Type", "application/json")
		tokens = m.copyAnthropicResponse(c.Writer, resp.Body)
	}
	return tokens, nil
}

func (m *AnthropicManufacturer) copyAnthropicResponse(dst io.Writer, src io.Reader) int {
	body, err := io.ReadAll(src)
	if err != nil {
		return 0
	}

	dst.Write(body)
	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}

	var resp struct {
		Usage anthropicUsage `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err == nil {
		return resp.Usage.total()
	}
	return 0
}

func (m *AnthropicManufacturer) copyAnthropicStreaming(dst io.Writer, src io.Reader) int {
	reader := bufio.NewReader(src)
	tokens := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		fmt.Fprintln(dst, line)

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var event struct {
			Type    string `json:"type"`
			Message struct {
				Usage anthropicUsage `json:"usage"`
			} `json:"message"`
			Usage anthropicUsage `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err == nil {
			switch event.Type {
			case "message_start":
				tokens += event.Message.Usage.total()
			case "message_delta":
				tokens += event.Usage.total()
			}
		}
	}
	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}
	return tokens
}
