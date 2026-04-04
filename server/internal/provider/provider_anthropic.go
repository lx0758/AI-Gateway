package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
)

type anthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

func (u anthropicUsage) total() int {
	return u.InputTokens + u.OutputTokens + u.CacheCreationInputTokens + u.CacheReadInputTokens
}

func (u anthropicUsage) toUsage(usage *Usage) {
	usage.CachedTokens = u.CacheReadInputTokens
	usage.InputTokens = u.InputTokens
	usage.OutputTokens = u.OutputTokens
}

type AnthropicProvider struct {
	cfg *Config
}

func NewAnthropicProvider(cfg *Config) *AnthropicProvider {
	return &AnthropicProvider{cfg: cfg}
}

func (m *AnthropicProvider) SyncModels(providerID uint) ([]model.ProviderModel, error) {
	baseURL := m.cfg.BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("Anthropic base URL is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", m.cfg.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", string(body))
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

	models := []model.ProviderModel{}
	for _, m := range result.Data {
		if m.ID == "" {
			continue
		}
		displayName := m.DisplayName
		if displayName == "" {
			displayName = m.ID
		}

		supportsVision := m.Capabilities.ImageInput.Supported
		supportsTools := true

		models = append(models, model.ProviderModel{
			ProviderID:     providerID,
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

func (m *AnthropicProvider) ExecuteOpenAIRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}

	recordBody("O2A", "raw", body)
	var openAIReq struct {
		Model     string                   `json:"model"`
		MaxTokens int                      `json:"max_tokens"`
		Messages  []map[string]interface{} `json:"messages"`
		Tools     json.RawMessage          `json:"tools,omitempty"`
		Stream    bool                     `json:"stream"`
	}
	if err := json.Unmarshal(body, &openAIReq); err != nil {
		return err
	}
	openAIReq.Model = pm.ModelID

	anthropicReq := m.convertOpenAIRequestToAnthropic(openAIReq)

	anthropicBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return err
	}

	recordBody("O2A", "converted", anthropicBody)
	req, err := http.NewRequest("POST", m.cfg.BaseURL+"/messages", bytes.NewReader(anthropicBody))
	if err != nil {
		return err
	}

	req.Header.Set("x-api-key", m.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": string(respBody)})
		recordError("O2A", resp.StatusCode, respBody)
		return fmt.Errorf("%s", string(respBody))
	}

	if m.isStreaming(resp) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		m.streamAnthropicToOpenAI(resp.Body, c.Writer, openAIReq.Model, usage)
		c.Writer.Flush()
	} else {
		anthropicRespBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		openAIResp, err := m.convertAnthropicResponseToOpenAI(anthropicRespBody, openAIReq.Model, usage)
		if err != nil {
			return err
		}
		c.Header("Content-Type", "application/json")
		c.Writer.Write(openAIResp)
	}
	return nil
}

func (m *AnthropicProvider) convertOpenAIRequestToAnthropic(openAIReq struct {
	Model     string                   `json:"model"`
	MaxTokens int                      `json:"max_tokens"`
	Messages  []map[string]interface{} `json:"messages"`
	Tools     json.RawMessage          `json:"tools,omitempty"`
	Stream    bool                     `json:"stream"`
}) map[string]interface{} {
	var systemContent string
	anthropicMessages := make([]map[string]interface{}, 0)
	var pendingToolResults []map[string]interface{}

	flushPendingToolResults := func() {
		if len(pendingToolResults) == 0 {
			return
		}
		content := make([]map[string]interface{}, 0, len(pendingToolResults))
		for _, tr := range pendingToolResults {
			content = append(content, tr)
		}
		anthropicMessages = append(anthropicMessages, map[string]interface{}{
			"role":    "user",
			"content": content,
		})
		pendingToolResults = pendingToolResults[:0]
	}

	for _, msg := range openAIReq.Messages {
		role, _ := msg["role"].(string)
		switch role {
		case "system":
			systemContent = m.extractSystemContent(msg["content"])
		case "tool":
			pendingToolResults = append(pendingToolResults, m.convertOpenAIToolResultToAnthropic(msg))
		case "assistant":
			flushPendingToolResults()
			anthropicMessages = append(anthropicMessages, m.convertOpenAIMessageToAnthropic(msg))
		default:
			flushPendingToolResults()
			anthropicMessages = append(anthropicMessages, m.convertOpenAIMessageToAnthropic(msg))
		}
	}
	flushPendingToolResults()

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

func (m *AnthropicProvider) extractSystemContent(content interface{}) string {
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

func (m *AnthropicProvider) convertOpenAIMessageToAnthropic(msg map[string]interface{}) map[string]interface{} {
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

func (m *AnthropicProvider) convertOpenAIToolResultToAnthropic(msg map[string]interface{}) map[string]interface{} {
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
		"type":        "tool_result",
		"tool_use_id": toolCallID,
		"content":     toolResultContent,
	}
}

func (m *AnthropicProvider) convertOpenAIToolToAnthropic(tool map[string]interface{}) map[string]interface{} {
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

func (m *AnthropicProvider) parseDataURL(url string) (mediaType, data string) {
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

func (m *AnthropicProvider) convertAnthropicResponseToOpenAI(anthropicResp []byte, model string, usage *Usage) ([]byte, error) {
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
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

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
			"total_tokens":      anthropic.Usage.total(),
		},
	}

	result, err := json.Marshal(openAIResp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenAI response: %w", err)
	}
	anthropic.Usage.toUsage(usage)
	return result, nil
}

func (m *AnthropicProvider) streamAnthropicToOpenAI(src io.Reader, dst io.Writer, model string, usage *Usage) {
	src, dst = recordStream("O2A", src, dst)
	reader := bufio.NewReader(src)
	messageID := fmt.Sprintf("chatcmpl-%s", m.generateID())
	tokens := 0
	created := time.Now().Unix()
	contentBuffer := ""
	toolCallsBuffer := make(map[int]map[string]interface{})
	errorCount := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			errorCount += 1
			if errorCount >= 3 {
				log.Printf("Anthropic stream error, error: %v", err)
				break
			}
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "event:") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)

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
			//if msg, ok := event.Message["usage"].(map[string]interface{}); ok {
			//	if it, ok := msg["input_tokens"].(float64); ok {
			//		inputTokens = int(it)
			//	}
			//	if it, ok := msg["output_tokens"].(float64); ok {
			//		outputTokens = int(it)
			//	}
			//	if it, ok := msg["cache_creation_input_tokens"].(float64); ok {
			//		cacheCreationInputTokens = int(it)
			//	}
			//	if it, ok := msg["cache_read_input_tokens"].(float64); ok {
			//		cacheReadInputTokens = int(it)
			//	}
			//	tokens += inputTokens + outputTokens + cacheCreationInputTokens + cacheReadInputTokens
			//}

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
					"prompt_tokens":     event.Usage.CacheCreationInputTokens + event.Usage.InputTokens,
					"completion_tokens": event.Usage.OutputTokens,
					"total_tokens":      event.Usage.total(),
					"prompt_tokens_details": map[string]interface{}{
						"cached_tokens": event.Usage.CacheReadInputTokens,
					},
					"completion_tokens_details": map[string]interface{}{
						"reasoning_tokens": event.Usage.CacheReadInputTokens,
					},
				},
			})

		case "message_stop":
			fmt.Fprint(dst, "data: [DONE]\n\n")
		}
	}

	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *AnthropicProvider) writeOpenAISSE(w io.Writer, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *AnthropicProvider) generateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 24)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

func (m *AnthropicProvider) isStreaming(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return len(resp.Header["Transfer-Encoding"]) > 0 ||
		(len(contentType) > 0 && len(contentType) >= 17 && contentType[:17] == "text/event-stream")
}

func (m *AnthropicProvider) ExecuteAnthropicRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}

	bodyJson := map[string]interface{}{}
	if err := json.Unmarshal(body, &bodyJson); err != nil {
		return err
	}
	bodyJson["model"] = pm.ModelID
	body, err = json.Marshal(bodyJson)
	if err != nil {
		return err
	}

	recordBody("A2A", "raw", body)
	req, err := http.NewRequest("POST", m.cfg.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("x-api-key", m.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": string(respBody)})
		recordError("A2A", resp.StatusCode, respBody)
		return fmt.Errorf("%s", string(respBody))
	}

	if m.isStreaming(resp) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		m.copyAnthropicStreaming(c.Writer, resp.Body, usage)
	} else {
		c.Header("Content-Type", "application/json")
		err := m.copyAnthropicResponse(c.Writer, resp.Body, usage)
		if err != nil {
			c.JSON(resp.StatusCode, gin.H{"error": err.Error()})
			return err
		}
	}
	return nil
}

func (m *AnthropicProvider) copyAnthropicStreaming(dst io.Writer, src io.Reader, usage *Usage) {
	src, dst = recordStream("O2A", src, dst)
	reader := bufio.NewReader(src)
	errorCount := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			errorCount += 1
			if errorCount >= 3 {
				log.Printf("Anthropic stream error, error: %v", err)
				break
			}
			continue
		}
		fmt.Fprint(dst, line)

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)

		var event struct {
			Type  string         `json:"type"`
			Usage anthropicUsage `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err == nil {
			switch event.Type {
			case "message_delta":
				event.Usage.toUsage(usage)
			}
		}
	}
	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *AnthropicProvider) copyAnthropicResponse(dst io.Writer, src io.Reader, usage *Usage) error {
	body, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	dst.Write(body)
	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}

	var resp struct {
		Usage anthropicUsage `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	resp.Usage.toUsage(usage)
	return nil
}
