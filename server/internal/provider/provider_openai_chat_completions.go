package provider

import (
"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
)

type openAIUsage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	TotalTokens         int `json:"total_tokens"`
	PromptTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

func (u openAIUsage) toUsage(usage *Usage) {
	usage.CachedTokens = u.PromptTokensDetails.CachedTokens
	usage.InputTokens = u.PromptTokens - u.PromptTokensDetails.CachedTokens
	usage.OutputTokens = u.CompletionTokens
}

type OpenAIProvider struct {
	cfg *Config
}

func NewOpenAIProvider(cfg *Config) *OpenAIProvider {
	return &OpenAIProvider{cfg: cfg}
}

func (m *OpenAIProvider) SyncModels(providerID uint) ([]model.ProviderModel, error) {
	baseURL := m.cfg.OpenAIChatCompletionsBaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("OpenAI base URL is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

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
			ID            string    `json:"id"`
			Name          string    `json:"name"`
			DisplayName   string    `json:"display_name"`
			Created       time.Time `json:"created"`
			CreatedAt     time.Time `json:"created_at"`
			Description   string    `json:"description"`
			ContextLength int       `json:"context_length"`
			Object        string    `json:"object"`
			OwnedBy       string    `json:"owned_by"`
			Pricing       struct {
				Completion float64 `json:"completion"`
				Prompt     float64 `json:"prompt"`
				WebSearch  float64 `json:"web_search"`
			} `json:"pricing"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := []model.ProviderModel{}
	for _, m := range result.Data {
		if m.ID == "" {
			continue
		}
		// 只填充厂商实际返回的字段，未返回的留空（display_name/max_output/capabilities），
		// 由同步合并逻辑"空则保留本地值"
		models = append(models, model.ProviderModel{
			ProviderID:    providerID,
			ModelID:       m.ID,
			DisplayName:   AnyString(m.Name, m.DisplayName),
			OwnedBy:       m.OwnedBy,
			ContextWindow: m.ContextLength,
			InputPrice:    m.Pricing.Prompt,
			OutputPrice:   m.Pricing.Completion,
		})
	}

	return models, nil
}

func AnyString(texts ...string) string {
	for _, t := range texts {
		if t != "" {
			return t
		}
	}
	return ""
}

func (m *OpenAIProvider) ExecuteOpenAIChatCompletionsRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	bodyJson := map[string]interface{}{}
	if err := json.Unmarshal(body, &bodyJson); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	bodyJson["model"] = pm.ModelID
	body, err = json.Marshal(bodyJson)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("OC2OC", "raw", body)
	req, err := http.NewRequest("POST", m.cfg.OpenAIChatCompletionsBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	req = req.WithContext(c.Request.Context())
	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.Status(resp.StatusCode)
		c.Header("Content-Type", resp.Header.Get("Content-Type"))
		c.Writer.Write(respBody)
		recordError("OC2OC", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = m.copyOpenAIStreaming(c.Request.Context(), c.Writer, resp.Body, usage)
	} else {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "application/json")
		err = m.copyOpenAIResponse(c.Writer, resp.Body, usage)
	}
	return err
}

func (m *OpenAIProvider) copyOpenAIStreaming(ctx context.Context, dst io.Writer, src io.Reader, usage *Usage) error {
	src, dst = recordStream("OC2OC", src, dst)
	reader := bufio.NewReader(src)
	var rError error
	errorCount := 0

	type readResult struct {
		line string
		err  error
	}

streamLoop:
	for {
		readCh := make(chan readResult, 1)

		go func() {
			select {
			case <-ctx.Done():
				readCh <- readResult{err: ctx.Err()}
			default:
				line, err := reader.ReadString('\n')
				readCh <- readResult{line: line, err: err}
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case result := <-readCh:
			if result.err != nil {
				if result.err == io.EOF {
					break streamLoop
				}
				errorCount++
				if errorCount >= 3 {
					rError = fmt.Errorf("OpenAI copy stream error, %v", result.err)
					break streamLoop
				}
				continue
			}

			if _, err := fmt.Fprint(dst, result.line); err != nil {
				break streamLoop
			}
			if flusher, ok := dst.(http.Flusher); ok {
				flusher.Flush()
			}

			line := strings.TrimSpace(result.line)
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if data == "[DONE]" {
				break streamLoop
			}

			var chunk struct {
				OpenAIUsage openAIUsage `json:"usage"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err == nil {
				chunk.OpenAIUsage.toUsage(usage)
			}
		}
	}
	return rError
}

func (m *OpenAIProvider) copyOpenAIResponse(dst io.Writer, src io.Reader, usage *Usage) error {
	body, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	dst.Write(body)
	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}

	var resp struct {
		OpenAIUsage openAIUsage `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	resp.OpenAIUsage.toUsage(usage)
	return nil
}


func (m *OpenAIProvider) isStreaming(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return len(resp.Header["Transfer-Encoding"]) > 0 ||
		(len(contentType) > 0 && len(contentType) >= 17 && contentType[:17] == "text/event-stream")
}


func (m *AnthropicProvider) ExecuteOpenAIChatCompletionsRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("OC2AM", "raw", body)
	var openAIReq struct {
		Model     string                   `json:"model"`
		MaxTokens int                      `json:"max_tokens"`
		Messages  []map[string]interface{} `json:"messages"`
		Tools     json.RawMessage          `json:"tools,omitempty"`
		Stream    bool                     `json:"stream"`
	}
	if err := json.Unmarshal(body, &openAIReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	openAIReq.Model = pm.ModelID

	anthropicReq := m.convertOpenAIRequestToAnthropic(openAIReq)

	anthropicBody, err := json.Marshal(anthropicReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("OC2AM", "converted", anthropicBody)
	req, err := http.NewRequest("POST", m.cfg.AnthropicMessagesBaseURL+"/messages", bytes.NewReader(anthropicBody))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	req = req.WithContext(c.Request.Context())
	req.Header.Set("x-api-key", m.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.Status(resp.StatusCode)
		c.Header("Content-Type", resp.Header.Get("Content-Type"))
		c.Writer.Write(respBody)
		recordError("OC2AM", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = m.streamAnthropicToOpenAI(c.Request.Context(), resp.Body, c.Writer, openAIReq.Model, usage)
		c.Writer.Flush()
	} else {
		anthropicRespBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		openAIResp, err := m.convertAnthropicResponseToOpenAI(anthropicRespBody, openAIReq.Model, usage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		c.Status(http.StatusOK)
		c.Header("Content-Type", "application/json")
		c.Writer.Write(openAIResp)
		err = nil
	}
	return err
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

func (m *AnthropicProvider) streamAnthropicToOpenAI(ctx context.Context, src io.Reader, dst io.Writer, model string, usage *Usage) error {
	src, dst = recordStream("OC2AM", src, dst)
	reader := bufio.NewReader(src)
	messageID := fmt.Sprintf("chatcmpl-%s", m.generateID())
	tokens := 0
	created := time.Now().Unix()
	contentBuffer := ""
	toolCallsBuffer := make(map[int]map[string]interface{})
	var rError error
	errorCount := 0

	type readResult struct {
		line string
		err  error
	}

streamLoop:
	for {
		readCh := make(chan readResult, 1)

		go func() {
			select {
			case <-ctx.Done():
				readCh <- readResult{err: ctx.Err()}
			default:
				line, err := reader.ReadString('\n')
				readCh <- readResult{line: line, err: err}
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case result := <-readCh:
			if result.err != nil {
				if result.err == io.EOF {
					break streamLoop
				}
				errorCount++
				if errorCount >= 3 {
					rError = fmt.Errorf("Anthropic convert stream error, %v", result.err)
					break streamLoop
				}
				continue
			}

			line := strings.TrimSpace(result.line)
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
				if flusher, ok := dst.(http.Flusher); ok {
					flusher.Flush()
				}
			}
		}
	}
	return rError
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

// ============================================================
// 格 C→R：Chat 客户端 → Responses 上游（休眠格）
// ============================================================

func (m *ResponsesProvider) ExecuteOpenAIChatCompletionsRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	responsesReq, _, includeUsage, err := convertChatRequestToResponses(body, pm.ModelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	responsesBody, err := json.Marshal(responsesReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("OC2OR", "converted", responsesBody)
	req, err := http.NewRequest("POST", m.cfg.OpenAIResponsesBaseURL+"/responses", bytes.NewReader(responsesBody))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	req = req.WithContext(c.Request.Context())
	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.Status(resp.StatusCode)
		c.Header("Content-Type", resp.Header.Get("Content-Type"))
		c.Writer.Write(respBody)
		recordError("OC2OR", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = streamResponsesToChat(c.Request.Context(), resp.Body, c.Writer, pm.ModelID, includeUsage, usage)
	} else {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		out, err := convertResponsesToChatResponse(respBody, pm.ModelID, usage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		c.Status(http.StatusOK)
		c.Header("Content-Type", "application/json")
		c.Writer.Write(out)
		err = nil
	}
	return err
}

func convertChatRequestToResponses(body []byte, upstreamModel string) (map[string]interface{}, bool, bool, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, false, false, err
	}

	stream := false
	if v, ok := req["stream"].(bool); ok {
		stream = v
	}
	includeUsage := false
	if so := respMap(req["stream_options"]); so != nil {
		includeUsage, _ = so["include_usage"].(bool)
	}

	instructions := make([]string, 0)
	input := make([]map[string]interface{}, 0)

	for _, raw := range respSlice(req["messages"]) {
		msg := respMap(raw)
		if msg == nil {
			continue
		}
		role := respString(msg["role"])
		switch role {
		case "system", "developer":
			if text := outputToText(msg["content"]); text != "" {
				instructions = append(instructions, text)
			}
		case "tool":
			input = append(input, map[string]interface{}{
				"type":   "function_call_output",
				"call_id": respString(msg["tool_call_id"]),
				"output": outputToText(msg["content"]),
			})
		case "assistant":
			if reasoning := respString(msg["reasoning_content"]); reasoning == "" {
				if r := respString(msg["reasoning"]); r != "" {
					reasoning = r
				}
			} else {
				input = append(input, map[string]interface{}{
					"type":    "reasoning",
					"id":      "rs_" + generateRandomID(),
					"summary": []interface{}{map[string]interface{}{"type": "summary_text", "text": reasoning}},
				})
			}
			if text := outputToText(msg["content"]); text != "" {
				input = append(input, map[string]interface{}{
					"type":    "message",
					"role":    "assistant",
					"content": []interface{}{map[string]interface{}{"type": "output_text", "text": text}},
				})
			}
			for _, tcRaw := range respSlice(msg["tool_calls"]) {
				tc := respMap(tcRaw)
				if tc == nil {
					continue
				}
				fn := respMap(tc["function"])
				input = append(input, map[string]interface{}{
					"type":      "function_call",
					"id":        "fc_" + generateRandomID(),
					"call_id":   respString(tc["id"]),
					"name":      respString(fn["name"]),
					"arguments": respString(fn["arguments"]),
				})
			}
		default:
			parts := convertChatContentToResponses(msg["content"])
			if len(parts) > 0 {
				input = append(input, map[string]interface{}{
					"type":    "message",
					"role":    "user",
					"content": parts,
				})
			}
		}
	}

	responsesReq := map[string]interface{}{
		"model": upstreamModel,
		"input": input,
		"store": false,
	}
	if len(instructions) > 0 {
		responsesReq["instructions"] = strings.Join(instructions, "\n")
	}
	if v, ok := req["max_tokens"]; ok && v != nil {
		responsesReq["max_output_tokens"] = v
	}
	if v, ok := req["temperature"]; ok {
		responsesReq["temperature"] = v
	}
	if v, ok := req["top_p"]; ok {
		responsesReq["top_p"] = v
	}
	if tools := convertChatToolsToResponses(req["tools"]); len(tools) > 0 {
		responsesReq["tools"] = tools
	}
	if v, ok := req["tool_choice"]; ok {
		if tc := convertChatToolChoiceToResponses(v); tc != nil {
			responsesReq["tool_choice"] = tc
		}
	}
	if stream {
		responsesReq["stream"] = true
	}
	return responsesReq, stream, includeUsage, nil
}

func convertChatContentToResponses(content interface{}) []interface{} {
	parts := make([]interface{}, 0)
	switch v := content.(type) {
	case string:
		if v != "" {
			parts = append(parts, map[string]interface{}{"type": "input_text", "text": v})
		}
	case []interface{}:
		for _, raw := range v {
			part := respMap(raw)
			if part == nil {
				continue
			}
			switch respString(part["type"]) {
			case "text":
				if text := respString(part["text"]); text != "" {
					parts = append(parts, map[string]interface{}{"type": "input_text", "text": text})
				}
			case "image_url":
				url := ""
				if m := respMap(part["image_url"]); m != nil {
					url = respString(m["url"])
				}
				if url != "" {
					parts = append(parts, map[string]interface{}{"type": "input_image", "image_url": url})
				}
			}
		}
	}
	return parts
}

func convertChatToolsToResponses(tools interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	for _, raw := range respSlice(tools) {
		tool := respMap(raw)
		if tool == nil || respString(tool["type"]) != "function" {
			continue
		}
		fn := respMap(tool["function"])
		if fn == nil {
			continue
		}
		flat := map[string]interface{}{"type": "function", "name": respString(fn["name"])}
		if desc := respString(fn["description"]); desc != "" {
			flat["description"] = desc
		}
		if params, ok := fn["parameters"]; ok {
			flat["parameters"] = params
		}
		out = append(out, flat)
	}
	return out
}

func convertChatToolChoiceToResponses(choice interface{}) interface{} {
	switch v := choice.(type) {
	case string:
		return v
	case map[string]interface{}:
		if respString(v["type"]) == "function" {
			name := respString(v["name"])
			if fn := respMap(v["function"]); fn != nil {
				name = respString(fn["name"])
			}
			return map[string]interface{}{"type": "function", "name": name}
		}
	}
	return nil
}

func convertResponsesToChatResponse(body []byte, model string, usage *Usage) ([]byte, error) {
	var resp struct {
		Output            []map[string]interface{} `json:"output"`
		Status            string                   `json:"status"`
		IncompleteDetails map[string]interface{}   `json:"incomplete_details"`
		Usage             responsesUsage           `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Responses response error")
	}

	message := map[string]interface{}{"role": "assistant", "content": ""}
	texts := make([]string, 0)
	reasonings := make([]string, 0)
	toolCalls := make([]map[string]interface{}, 0)
	for _, item := range resp.Output {
		switch respString(item["type"]) {
		case "message":
			for _, raw := range respSlice(item["content"]) {
				if part := respMap(raw); part != nil {
					if respString(part["type"]) == "output_text" {
						if text := respString(part["text"]); text != "" {
							texts = append(texts, text)
						}
					}
				}
			}
		case "reasoning":
			for _, raw := range respSlice(item["summary"]) {
				if part := respMap(raw); part != nil {
					if text := respString(part["text"]); text != "" {
						reasonings = append(reasonings, text)
					}
				}
			}
		case "function_call":
			toolCalls = append(toolCalls, map[string]interface{}{
				"id":   respString(item["call_id"]),
				"type": "function",
				"function": map[string]interface{}{
					"name":      respString(item["name"]),
					"arguments": respString(item["arguments"]),
				},
			})
		}
	}
	if len(texts) > 0 {
		message["content"] = strings.Join(texts, "")
	}
	if len(reasonings) > 0 {
		message["reasoning_content"] = strings.Join(reasonings, "")
	}
	if len(toolCalls) > 0 {
		message["tool_calls"] = toolCalls
	}

	finishReason := "stop"
	if len(toolCalls) > 0 {
		finishReason = "tool_calls"
	}
	if resp.Status == "incomplete" {
		switch respString(resp.IncompleteDetails["reason"]) {
		case "max_output_tokens":
			finishReason = "length"
		case "content_filter":
			finishReason = "content_filter"
		}
	}

	chatResp := map[string]interface{}{
		"id":      "chatcmpl-" + generateRandomID(),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]interface{}{
			{"index": 0, "message": message, "finish_reason": finishReason},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     resp.Usage.InputTokens,
			"completion_tokens": resp.Usage.OutputTokens,
			"total_tokens":      resp.Usage.TotalTokens,
		},
	}
	out, err := json.Marshal(chatResp)
	if err != nil {
		return nil, err
	}
	resp.Usage.toUsage(usage)
	return out, nil
}

func emitChatSSE(w io.Writer, data interface{}) {
	payload, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", payload)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func streamResponsesToChat(ctx context.Context, src io.Reader, dst io.Writer, model string, includeUsage bool, usage *Usage) error {
	responseID := "chatcmpl-" + generateRandomID()
	created := time.Now().Unix()
	toolIndexByOutput := make(map[int]int)
	nextToolIndex := 0
	hasTool := false
	completed := false
	var lastUsage responsesUsage
	finishReason := "stop"

	emitChunk := func(delta map[string]interface{}, finish interface{}) {
		chunk := map[string]interface{}{
			"id":      responseID,
			"object":  "chat.completion.chunk",
			"created": created,
			"model":   model,
			"choices": []map[string]interface{}{
				{"index": 0, "delta": delta, "finish_reason": finish},
			},
		}
		emitChatSSE(dst, chunk)
	}
	emitChunk(map[string]interface{}{"role": "assistant"}, nil)

	err := consumeSSE(ctx, src, func(line string) (bool, error) {
		data, ok := sseData(line)
		if !ok || data == "" {
			return false, nil
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return false, nil
		}
		eventType := respString(event["type"])
		switch eventType {
		case "response.output_text.delta":
			emitChunk(map[string]interface{}{"content": respString(event["delta"])}, nil)
		case "response.reasoning_summary_text.delta":
			emitChunk(map[string]interface{}{"reasoning_content": respString(event["delta"])}, nil)
		case "response.output_item.added":
			item := respMap(event["item"])
			if item != nil && respString(item["type"]) == "function_call" {
				outputIndex := respInt(event["output_index"])
				index := nextToolIndex
				nextToolIndex++
				toolIndexByOutput[outputIndex] = index
				hasTool = true
				emitChunk(map[string]interface{}{
					"tool_calls": []map[string]interface{}{
						{
							"index": index,
							"id":    respString(item["call_id"]),
							"type":  "function",
							"function": map[string]interface{}{
								"name":      respString(item["name"]),
								"arguments": "",
							},
						},
					},
				}, nil)
			}
		case "response.function_call_arguments.delta":
			outputIndex := respInt(event["output_index"])
			index, ok := toolIndexByOutput[outputIndex]
			if !ok {
				index = nextToolIndex
				nextToolIndex++
				toolIndexByOutput[outputIndex] = index
			}
			emitChunk(map[string]interface{}{
				"tool_calls": []map[string]interface{}{
					{"index": index, "function": map[string]interface{}{"arguments": respString(event["delta"])}},
				},
			}, nil)
		case "response.completed":
			completed = true
			response := respMap(event["response"])
			if response != nil {
				u := respMap(response["usage"])
				lastUsage = responsesUsage{
					InputTokens:  respInt(u["input_tokens"]),
					OutputTokens: respInt(u["output_tokens"]),
					TotalTokens:  respInt(u["total_tokens"]),
				}
				if d := respMap(u["input_tokens_details"]); d != nil {
					lastUsage.InputTokensDetails.CachedTokens = respInt(d["cached_tokens"])
				}
				if respString(response["status"]) == "incomplete" {
					if d := respMap(response["incomplete_details"]); d != nil {
						switch respString(d["reason"]) {
						case "max_output_tokens":
							finishReason = "length"
						case "content_filter":
							finishReason = "content_filter"
						}
					}
				}
			}
			if hasTool && finishReason == "stop" {
				finishReason = "tool_calls"
			}
			chunk := map[string]interface{}{
				"id":      responseID,
				"object":  "chat.completion.chunk",
				"created": created,
				"model":   model,
				"choices": []map[string]interface{}{
					{"index": 0, "delta": map[string]interface{}{}, "finish_reason": finishReason},
				},
			}
			if includeUsage {
				chunk["usage"] = map[string]interface{}{
					"prompt_tokens":     lastUsage.InputTokens,
					"completion_tokens": lastUsage.OutputTokens,
					"total_tokens":      lastUsage.TotalTokens,
				}
			}
			emitChatSSE(dst, chunk)
			fmt.Fprint(dst, "data: [DONE]\n\n")
			if flusher, ok := dst.(http.Flusher); ok {
				flusher.Flush()
			}
			lastUsage.toUsage(usage)
			return true, nil
		case "response.failed", "error":
			completed = true
			emitChunk(map[string]interface{}{}, finishReason)
			fmt.Fprint(dst, "data: [DONE]\n\n")
			if flusher, ok := dst.(http.Flusher); ok {
				flusher.Flush()
			}
			return true, nil
		}
		return false, nil
	})
	if err != nil && !completed {
		emitChunk(map[string]interface{}{}, "stop")
		fmt.Fprint(dst, "data: [DONE]\n\n")
	}
	return nil
}
