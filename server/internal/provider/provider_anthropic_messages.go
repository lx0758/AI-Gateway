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
	baseURL := m.cfg.AnthropicMessagesBaseURL
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

		// 能力只取厂商接口真实返回的字段，不硬编码 stream/tools；未返回的字段留空
		var caps []string
		if m.Capabilities.ImageInput.Supported {
			caps = append(caps, "image")
		}
		if m.Capabilities.Thinking.Supported {
			caps = append(caps, "thinking")
		}

		models = append(models, model.ProviderModel{
			ProviderID:    providerID,
			ModelID:       m.ID,
			DisplayName:   m.DisplayName,
			ContextWindow: m.MaxInputToken,
			MaxOutput:     m.MaxTokens,
			Capabilities:  strings.Join(caps, ","),
		})
	}

	return models, nil
}


func (m *AnthropicProvider) isStreaming(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return len(resp.Header["Transfer-Encoding"]) > 0 ||
		(len(contentType) > 0 && len(contentType) >= 17 && contentType[:17] == "text/event-stream")
}

func (m *AnthropicProvider) ExecuteAnthropicMessagesRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
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

	recordBody("AM2AM", "raw", body)
	req, err := http.NewRequest("POST", m.cfg.AnthropicMessagesBaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

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
		recordError("AM2AM", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = m.copyAnthropicStreaming(c.Request.Context(), c.Writer, resp.Body, usage)
	} else {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "application/json")
		err = m.copyAnthropicResponse(c.Writer, resp.Body, usage)
	}
	return err
}

func (m *AnthropicProvider) copyAnthropicStreaming(ctx context.Context, dst io.Writer, src io.Reader, usage *Usage) error {
	src, dst = recordStream("OC2AM", src, dst)
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
					rError = fmt.Errorf("Anthropic copy stream error, %v", result.err)
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
	}
	return rError
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

func (m *OpenAIProvider) ExecuteAnthropicMessagesRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("AM2OC", "raw", body)
	var anthropicReq struct {
		Model         string          `json:"model"`
		MaxTokens     int             `json:"max_tokens"`
		System        interface{}     `json:"system"`
		Messages      json.RawMessage `json:"messages"`
		Tools         json.RawMessage `json:"tools,omitempty"`
		Stream        bool            `json:"stream"`
		StreamOptions *struct{}       `json:"stream_options,omitempty"`
	}
	if err := json.Unmarshal(body, &anthropicReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	anthropicReq.Model = pm.ModelID

	var anthropicMessages []map[string]interface{}
	if err := json.Unmarshal(anthropicReq.Messages, &anthropicMessages); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	openAIMessages := make([]map[string]interface{}, 0)
	systemContent := m.extractSystemContent(anthropicReq.System)
	if systemContent != "" {
		openAIMessages = append(openAIMessages, map[string]interface{}{
			"role":    "system",
			"content": systemContent,
		})
	}
	for _, msg := range anthropicMessages {
		openAIMsgs := m.convertAnthropicMessageToOpenAI(msg)
		openAIMessages = append(openAIMessages, openAIMsgs...)
	}

	openAIReq := map[string]interface{}{
		"model":      anthropicReq.Model,
		"messages":   openAIMessages,
		"max_tokens": anthropicReq.MaxTokens,
	}
	if anthropicReq.Stream {
		openAIReq["stream"] = true
		openAIReq["stream_options"] = map[string]bool{"include_usage": true}
	}
	if anthropicReq.Tools != nil {
		var tools []interface{}
		if err := json.Unmarshal(anthropicReq.Tools, &tools); err == nil {
			openAITools := make([]map[string]interface{}, 0, len(tools))
			for _, tool := range tools {
				if t, ok := tool.(map[string]interface{}); ok {
					openAITools = append(openAITools, m.convertAnthropicToolToOpenAI(t))
				}
			}
			openAIReq["tools"] = openAITools
		}
	}

	openAIBody, err := json.Marshal(openAIReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("AM2OC", "converted", openAIBody)
	req, err := http.NewRequest("POST", m.cfg.OpenAIChatCompletionsBaseURL+"/chat/completions", bytes.NewReader(openAIBody))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

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
		recordError("AM2OC", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = m.streamOpenAIToAnthropic(c.Request.Context(), resp.Body, c.Writer, anthropicReq.Model, usage)
		c.Writer.Flush()
	} else {
		openAIRespBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		anthropicResp, err := m.convertOpenAIResponseToAnthropic(openAIRespBody, usage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		c.Status(http.StatusOK)
		c.Header("Content-Type", "application/json")
		c.Writer.Write(anthropicResp)
		err = nil
	}
	return err
}

func (m *OpenAIProvider) extractSystemContent(system interface{}) string {
	if system == nil {
		return ""
	}
	switch v := system.(type) {
	case string:
		return v
	case []interface{}:
		var texts []string
		for _, block := range v {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockType, _ := blockMap["type"].(string); blockType == "text" {
					if text, _ := blockMap["text"].(string); text != "" {
						texts = append(texts, text)
					}
				}
			}
		}
		if len(texts) > 0 {
			result := ""
			for i, t := range texts {
				if i > 0 {
					result += "\n"
				}
				result += t
			}
			return result
		}
	}
	return ""
}

func (m *OpenAIProvider) convertAnthropicMessageToOpenAI(msg map[string]interface{}) []map[string]interface{} {
	role, _ := msg["role"].(string)
	content := msg["content"]

	result := map[string]interface{}{
		"role": role,
	}

	switch v := content.(type) {
	case string:
		result["content"] = v
		return []map[string]interface{}{result}
	case []interface{}:
		textParts := make([]string, 0)
		imageParts := make([]map[string]interface{}, 0)
		toolCalls := make([]map[string]interface{}, 0)
		toolResults := make([]map[string]interface{}, 0)

		for _, block := range v {
			if blockMap, ok := block.(map[string]interface{}); ok {
				blockType, _ := blockMap["type"].(string)
				switch blockType {
				case "text":
					if text, _ := blockMap["text"].(string); text != "" {
						textParts = append(textParts, text)
					}
				case "image":
					source, _ := blockMap["source"].(map[string]interface{})
					if source != nil {
						imageURL := ""
						if mediaType, _ := source["media_type"].(string); mediaType != "" {
							if data, _ := source["data"].(string); data != "" {
								imageURL = "data:" + mediaType + ";base64," + data
							}
						} else if url, _ := source["url"].(string); url != "" {
							imageURL = url
						}
						if imageURL != "" {
							imageParts = append(imageParts, map[string]interface{}{
								"type": "image_url",
								"image_url": map[string]interface{}{
									"url": imageURL,
								},
							})
						}
					}
				case "tool_use":
					toolID, _ := blockMap["id"].(string)
					toolName, _ := blockMap["name"].(string)
					toolInput := blockMap["input"]
					inputBytes, _ := json.Marshal(toolInput)
					toolCalls = append(toolCalls, map[string]interface{}{
						"id":   toolID,
						"type": "function",
						"function": map[string]interface{}{
							"name":      toolName,
							"arguments": string(inputBytes),
						},
					})
				case "tool_result":
					toolResults = append(toolResults, m.convertAnthropicToolResultToOpenAI(blockMap))
				}
			}
		}

		if len(toolCalls) > 0 {
			result["tool_calls"] = toolCalls
			if len(textParts) > 0 {
				result["content"] = textParts[0]
				for i := 1; i < len(textParts); i++ {
					result["content"] = result["content"].(string) + "\n" + textParts[i]
				}
			} else {
				result["content"] = nil
			}
		} else if len(toolResults) > 0 && len(textParts) == 0 && len(imageParts) == 0 {
			return toolResults
		} else if len(imageParts) > 0 {
			var contentParts []map[string]interface{}
			for _, t := range textParts {
				contentParts = append(contentParts, map[string]interface{}{
					"type": "text",
					"text": t,
				})
			}
			contentParts = append(contentParts, imageParts...)
			result["content"] = contentParts
		} else if len(textParts) > 0 {
			result["content"] = textParts[0]
			for i := 1; i < len(textParts); i++ {
				result["content"] = result["content"].(string) + "\n" + textParts[i]
			}
		}

		messages := []map[string]interface{}{result}
		messages = append(messages, toolResults...)
		return messages
	default:
		result["content"] = v
		return []map[string]interface{}{result}
	}
}

func (m *OpenAIProvider) convertAnthropicToolResultToOpenAI(blockMap map[string]interface{}) map[string]interface{} {
	toolUseID, _ := blockMap["tool_use_id"].(string)
	result := map[string]interface{}{
		"role":         "tool",
		"tool_call_id": toolUseID,
	}
	content := blockMap["content"]
	switch v := content.(type) {
	case string:
		result["content"] = v
	case []interface{}:
		textParts := make([]string, 0)
		for _, block := range v {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockType, _ := blockMap["type"].(string); blockType == "text" {
					if text, _ := blockMap["text"].(string); text != "" {
						textParts = append(textParts, text)
					}
				}
			}
		}
		if len(textParts) > 0 {
			result["content"] = textParts[0]
			for i := 1; i < len(textParts); i++ {
				result["content"] = result["content"].(string) + "\n" + textParts[i]
			}
		} else {
			result["content"] = ""
		}
	default:
		result["content"] = fmt.Sprintf("%v", v)
	}
	return result
}

func (m *OpenAIProvider) convertAnthropicToolToOpenAI(tool map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        tool["name"],
			"description": tool["description"],
		},
	}
	if inputSchema, ok := tool["input_schema"].(map[string]interface{}); ok {
		result["function"].(map[string]interface{})["parameters"] = inputSchema
	}
	return result
}

func (m *OpenAIProvider) convertOpenAIResponseToAnthropic(openAIResp []byte, usage *Usage) ([]byte, error) {
	var openAI struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Created int64  `json:"created"`
		Choices []struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		} `json:"choices"`
		OpenAIUsage openAIUsage `json:"usage"`
	}
	if err := json.Unmarshal(openAIResp, &openAI); err != nil {
		return nil, fmt.Errorf("OpenAI response error")
	}

	if len(openAI.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI choices empty")
	}

	choice := openAI.Choices[0]
	content := make([]map[string]interface{}, 0)

	if msgContent := choice.Message["content"]; msgContent != nil {
		content = append(content, map[string]interface{}{
			"type": "text",
			"text": msgContent,
		})
	}

	if toolCalls, ok := choice.Message["tool_calls"].([]interface{}); ok {
		for _, tc := range toolCalls {
			if tcMap, ok := tc.(map[string]interface{}); ok {
				toolUse := map[string]interface{}{
					"type":  "tool_use",
					"id":    tcMap["id"],
					"name":  "",
					"input": map[string]interface{}{},
				}
				if fn, ok := tcMap["function"].(map[string]interface{}); ok {
					toolUse["name"] = fn["name"]
					if inputStr, _ := fn["arguments"].(string); inputStr != "" {
						var input map[string]interface{}
						if json.Unmarshal([]byte(inputStr), &input) == nil {
							toolUse["input"] = input
						}
					}
				}
				content = append(content, toolUse)
			}
		}
	}

	stopReason := "end_turn"
	switch choice.FinishReason {
	case "stop":
		stopReason = "end_turn"
	case "length":
		stopReason = "max_tokens"
	case "tool_calls":
		stopReason = "tool_use"
	}

	anthropicResp := map[string]interface{}{
		"id":            openAI.ID,
		"type":          "message",
		"role":          "assistant",
		"model":         openAI.Model,
		"content":       content,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage": map[string]interface{}{
			"input_tokens":  openAI.OpenAIUsage.PromptTokens,
			"output_tokens": openAI.OpenAIUsage.CompletionTokens,
		},
	}

	result, err := json.Marshal(anthropicResp)
	if err != nil {
		return nil, fmt.Errorf("Anthropic response serialization error")
	}
	openAI.OpenAIUsage.toUsage(usage)
	return result, nil
}
type toolCallState struct {
	id        string
	name      string
	blockIdx  int
	startSent bool
}

func (m *OpenAIProvider) streamOpenAIToAnthropic(ctx context.Context, src io.Reader, dst io.Writer, model string, usage *Usage) error {
	src, dst = recordStream("AM2OC", src, dst)
	reader := bufio.NewReader(src)
	messageID := fmt.Sprintf("msg_%s", m.generateID())
	sentMessageStart := false
	preBlockCount := 0
	inThinkingBlock := false
	inTextBlock := false
	lastUsage := openAIUsage{}
	toolCallStates := make(map[int]*toolCallState)
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
					rError = fmt.Errorf("OpenAI convert stream error, %v", result.err)
					break streamLoop
				}
				continue
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
				ID      string `json:"id"`
				Model   string `json:"model"`
				Choices []struct {
					Index        int                    `json:"index"`
					Delta        map[string]interface{} `json:"delta"`
					FinishReason string                 `json:"finish_reason"`
				} `json:"choices"`
				OpenAIUsage openAIUsage `json:"usage"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if chunk.OpenAIUsage.TotalTokens > 0 {
				lastUsage = chunk.OpenAIUsage
			}

			if !sentMessageStart {
				m.writeAnthropicSSE(dst, "message_start", map[string]interface{}{
					"type": "message_start",
					"message": map[string]interface{}{
						"id":            messageID,
						"type":          "message",
						"role":          "assistant",
						"content":       []interface{}{},
						"model":         model,
						"stop_reason":   nil,
						"stop_sequence": nil,
						"usage": map[string]interface{}{
							"input_tokens":  0,
							"output_tokens": 0,
						},
					},
				})
				sentMessageStart = true
			}

			if len(chunk.Choices) > 0 {
				choice := chunk.Choices[0]
				delta := choice.Delta

				reasoning := ""
				hasReasoning := false
				if r, ok := delta["reasoning_content"].(string); ok && r != "" {
					reasoning = r
					hasReasoning = true
				} else if r, ok := delta["reasoning"].(string); ok && r != "" {
					reasoning = r
					hasReasoning = true
				}
				if hasReasoning && reasoning != "" {
					if !inThinkingBlock {
						if inTextBlock {
							m.writeAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
								"type":  "content_block_stop",
								"index": preBlockCount,
							})
							preBlockCount++
							inTextBlock = false
						}
						m.writeAnthropicSSE(dst, "content_block_start", map[string]interface{}{
							"type":  "content_block_start",
							"index": preBlockCount,
							"content_block": map[string]interface{}{
								"type":     "thinking",
								"thinking": "",
							},
						})
						inThinkingBlock = true
					}
					m.writeAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
						"type":  "content_block_delta",
						"index": preBlockCount,
						"delta": map[string]interface{}{
							"type":     "thinking_delta",
							"thinking": reasoning,
						},
					})
				}

				content, hasContent := delta["content"].(string)
				if hasContent && content != "" {
					if inThinkingBlock {
						m.writeAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
							"type":  "content_block_stop",
							"index": preBlockCount,
						})
						preBlockCount++
						inThinkingBlock = false
					}
					if !inTextBlock {
						m.writeAnthropicSSE(dst, "content_block_start", map[string]interface{}{
							"type":  "content_block_start",
							"index": preBlockCount,
							"content_block": map[string]interface{}{
								"type": "text",
								"text": "",
							},
						})
						inTextBlock = true
					}
					m.writeAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
						"type":  "content_block_delta",
						"index": preBlockCount,
						"delta": map[string]interface{}{
							"type": "text_delta",
							"text": content,
						},
					})
				}

				if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
					if inThinkingBlock || inTextBlock {
						m.writeAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
							"type":  "content_block_stop",
							"index": preBlockCount,
						})
						preBlockCount++
						inThinkingBlock = false
						inTextBlock = false
					}

					for _, tc := range toolCalls {
						tcMap, ok := tc.(map[string]interface{})
						if !ok {
							continue
						}

						toolIndex := 0
						if idx, ok := tcMap["index"].(float64); ok {
							toolIndex = int(idx)
						}

						toolID, _ := tcMap["id"].(string)
						fn, hasFn := tcMap["function"].(map[string]interface{})
						name := ""
						args := ""
						if hasFn {
							name, _ = fn["name"].(string)
							args, _ = fn["arguments"].(string)
						}

						state, exists := toolCallStates[toolIndex]
						if !exists {
							if toolID == "" {
								continue
							}
							state = &toolCallState{
								id:        toolID,
								name:      name,
								blockIdx:  preBlockCount + toolIndex,
								startSent: false,
							}
							toolCallStates[toolIndex] = state
						}

						if !state.startSent {
							m.writeAnthropicSSE(dst, "content_block_start", map[string]interface{}{
								"type":  "content_block_start",
								"index": state.blockIdx,
								"content_block": map[string]interface{}{
									"type":  "tool_use",
									"id":    state.id,
									"name":  state.name,
									"input": json.RawMessage("{}"),
								},
							})
							state.startSent = true
						}

						if args != "" {
							m.writeAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
								"type":  "content_block_delta",
								"index": state.blockIdx,
								"delta": map[string]interface{}{
									"type":         "input_json_delta",
									"partial_json": args,
								},
							})
						}
					}
				}

				if choice.FinishReason != "" {
					if inThinkingBlock || inTextBlock {
						m.writeAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
							"type":  "content_block_stop",
							"index": preBlockCount,
						})
					}

					for _, state := range toolCallStates {
						if state.startSent {
							m.writeAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
								"type":  "content_block_stop",
								"index": state.blockIdx,
							})
						}
					}

					stopReason := "end_turn"
					switch choice.FinishReason {
					case "stop":
						stopReason = "end_turn"
					case "length":
						stopReason = "max_tokens"
					case "tool_calls":
						stopReason = "tool_use"
					}

					if chunk.OpenAIUsage.TotalTokens > 0 {
						lastUsage = chunk.OpenAIUsage
					}

					if lastUsage.TotalTokens > 0 {
						lastUsage.toUsage(usage)
						m.writeAnthropicSSE(dst, "message_delta", map[string]interface{}{
							"type": "message_delta",
							"delta": map[string]interface{}{
								"stop_reason": stopReason,
							},
							"usage": map[string]interface{}{
								"input_tokens":            lastUsage.PromptTokens - lastUsage.PromptTokensDetails.CachedTokens,
								"output_tokens":           lastUsage.CompletionTokens,
								"cache_read_input_tokens": lastUsage.PromptTokensDetails.CachedTokens,
							},
						})
						m.writeAnthropicSSE(dst, "message_stop", map[string]interface{}{
							"type": "message_stop",
						})
						break streamLoop
					}
				}
			}
		}
	}
	return rError
}

func (m *OpenAIProvider) writeAnthropicSSE(w io.Writer, eventType string, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(dataBytes))
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *OpenAIProvider) generateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 24)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// ============================================================
// 格 A→R：Anthropic 客户端 → Responses 上游（休眠格）
// ============================================================

func (m *ResponsesProvider) ExecuteAnthropicMessagesRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	responsesReq, _, err := convertAnthropicRequestToResponses(body, pm.ModelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	responsesBody, err := json.Marshal(responsesReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("AM2OR", "converted", responsesBody)
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
		recordError("AM2OR", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = streamResponsesToAnthropic(c.Request.Context(), resp.Body, c.Writer, pm.ModelID, usage)
	} else {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		out, err := convertResponsesToAnthropicResponse(respBody, pm.ModelID, usage)
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

func convertAnthropicRequestToResponses(body []byte, upstreamModel string) (map[string]interface{}, bool, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, false, err
	}

	stream := false
	if v, ok := req["stream"].(bool); ok {
		stream = v
	}

	var instructions []string
	switch system := req["system"].(type) {
	case string:
		if system != "" {
			instructions = append(instructions, system)
		}
	case []interface{}:
		for _, raw := range system {
			if block := respMap(raw); block != nil && respString(block["type"]) == "text" {
				if text := respString(block["text"]); text != "" {
					instructions = append(instructions, text)
				}
			}
		}
	}

	input := make([]map[string]interface{}, 0)
	for _, raw := range respSlice(req["messages"]) {
		msg := respMap(raw)
		if msg == nil {
			continue
		}
		role := respString(msg["role"])
		switch content := msg["content"].(type) {
		case string:
			if content == "" {
				continue
			}
			partType := "input_text"
			if role == "assistant" {
				partType = "output_text"
			}
			input = append(input, map[string]interface{}{
				"type":    "message",
				"role":    role,
				"content": []interface{}{map[string]interface{}{"type": partType, "text": content}},
			})
		case []interface{}:
			if role == "assistant" {
				for _, blockRaw := range content {
					block := respMap(blockRaw)
					if block == nil {
						continue
					}
					switch respString(block["type"]) {
					case "thinking":
						thinking := respString(block["thinking"])
						if thinking != "" {
							// 信封规则③：发往 R 上游的请求永不携带信封
							input = append(input, map[string]interface{}{
								"type":    "reasoning",
								"id":      "rs_" + generateRandomID(),
								"summary": []interface{}{map[string]interface{}{"type": "summary_text", "text": thinking}},
							})
						}
					case "text":
						if text := respString(block["text"]); text != "" {
							input = append(input, map[string]interface{}{
								"type":    "message",
								"role":    "assistant",
								"content": []interface{}{map[string]interface{}{"type": "output_text", "text": text}},
							})
						}
					case "tool_use":
						args := ""
						if inputVal, ok := block["input"]; ok {
							if b, err := json.Marshal(inputVal); err == nil {
								args = string(b)
							}
						}
						input = append(input, map[string]interface{}{
							"type":      "function_call",
							"id":        "fc_" + generateRandomID(),
							"call_id":   respString(block["id"]),
							"name":      respString(block["name"]),
							"arguments": args,
						})
					}
				}
				continue
			}

			parts := make([]interface{}, 0)
			toolOutputs := make([]map[string]interface{}, 0)
			for _, blockRaw := range content {
				block := respMap(blockRaw)
				if block == nil {
					continue
				}
				switch respString(block["type"]) {
				case "text":
					if text := respString(block["text"]); text != "" {
						parts = append(parts, map[string]interface{}{"type": "input_text", "text": text})
					}
				case "image":
					source := respMap(block["source"])
					if source == nil {
						continue
					}
					url := ""
					if mediaType := respString(source["media_type"]); mediaType != "" {
						if data := respString(source["data"]); data != "" {
							url = "data:" + mediaType + ";base64," + data
						}
					} else {
						url = respString(source["url"])
					}
					if url != "" {
						parts = append(parts, map[string]interface{}{"type": "input_image", "image_url": url})
					}
				case "tool_result":
					toolOutputs = append(toolOutputs, map[string]interface{}{
						"type":   "function_call_output",
						"call_id": respString(block["tool_use_id"]),
						"output": outputToText(block["content"]),
					})
				}
			}
			if len(parts) > 0 {
				input = append(input, map[string]interface{}{"type": "message", "role": "user", "content": parts})
			}
			input = append(input, toolOutputs...)
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
	if tools := convertAnthropicToolsToResponses(req["tools"]); len(tools) > 0 {
		responsesReq["tools"] = tools
	}
	if v, ok := req["tool_choice"]; ok {
		if tc := convertAnthropicToolChoiceToResponses(v); tc != nil {
			responsesReq["tool_choice"] = tc
		}
	}
	if stream {
		responsesReq["stream"] = true
	}
	return responsesReq, stream, nil
}

func convertAnthropicToolsToResponses(tools interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	for _, raw := range respSlice(tools) {
		tool := respMap(raw)
		if tool == nil {
			continue
		}
		flat := map[string]interface{}{"type": "function", "name": respString(tool["name"])}
		if desc := respString(tool["description"]); desc != "" {
			flat["description"] = desc
		}
		if schema, ok := tool["input_schema"]; ok {
			flat["parameters"] = schema
		}
		out = append(out, flat)
	}
	return out
}

func convertAnthropicToolChoiceToResponses(choice interface{}) interface{} {
	switch v := choice.(type) {
	case string:
		return v
	case map[string]interface{}:
		switch respString(v["type"]) {
		case "any":
			return "required"
		case "none":
			return "none"
		case "auto":
			return "auto"
		case "tool":
			return map[string]interface{}{"type": "function", "name": respString(v["name"])}
		}
	}
	return nil
}

func convertResponsesToAnthropicResponse(body []byte, model string, usage *Usage) ([]byte, error) {
	var resp struct {
		Output            []map[string]interface{} `json:"output"`
		Status            string                   `json:"status"`
		IncompleteDetails map[string]interface{}   `json:"incomplete_details"`
		Usage             responsesUsage           `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Responses response error")
	}

	content := make([]map[string]interface{}, 0)
	texts := make([]string, 0)
	hasToolUse := false
	for _, item := range resp.Output {
		switch respString(item["type"]) {
		case "message":
			for _, raw := range respSlice(item["content"]) {
				if part := respMap(raw); part != nil && respString(part["type"]) == "output_text" {
					if text := respString(part["text"]); text != "" {
						texts = append(texts, text)
					}
				}
			}
		case "function_call":
			hasToolUse = true
			input := map[string]interface{}{}
			if args := respString(item["arguments"]); args != "" {
				if err := json.Unmarshal([]byte(args), &input); err != nil {
					input = map[string]interface{}{}
				}
			}
			content = append(content, map[string]interface{}{
				"type":  "tool_use",
				"id":    respString(item["call_id"]),
				"name":  respString(item["name"]),
				"input": input,
			})
		}
	}
	if len(texts) > 0 {
		content = append([]map[string]interface{}{{"type": "text", "text": strings.Join(texts, "")}}, content...)
	}

	stopReason := "end_turn"
	if hasToolUse {
		stopReason = "tool_use"
	}
	if resp.Status == "incomplete" {
		switch respString(resp.IncompleteDetails["reason"]) {
		case "max_output_tokens":
			stopReason = "max_tokens"
		case "content_filter":
			stopReason = "refusal"
		}
	}

	anthropicResp := map[string]interface{}{
		"id":            "msg_" + generateRandomID(),
		"type":          "message",
		"role":          "assistant",
		"model":         model,
		"content":       content,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage": map[string]interface{}{
			"input_tokens":  resp.Usage.InputTokens,
			"output_tokens": resp.Usage.OutputTokens,
		},
	}
	out, err := json.Marshal(anthropicResp)
	if err != nil {
		return nil, err
	}
	au := anthropicUsage{InputTokens: resp.Usage.InputTokens, OutputTokens: resp.Usage.OutputTokens}
	au.toUsage(usage)
	return out, nil
}

func emitAnthropicSSE(w io.Writer, eventType string, data interface{}) {
	payload, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, payload)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func streamResponsesToAnthropic(ctx context.Context, src io.Reader, dst io.Writer, model string, usage *Usage) error {
	messageID := "msg_" + generateRandomID()
	started := false
	completed := false
	blockIndex := 0
	blockByOutput := make(map[int]int)
	kindByOutput := make(map[int]string)
	hasToolUse := false
	stopReason := "end_turn"
	outputTokens := 0
	inputTokens := 0
	openBlocks := make(map[int]string)

	ensureStarted := func() {
		if started {
			return
		}
		started = true
		emitAnthropicSSE(dst, "message_start", map[string]interface{}{
			"type": "message_start",
			"message": map[string]interface{}{
				"id":            messageID,
				"type":          "message",
				"role":          "assistant",
				"content":       []interface{}{},
				"model":         model,
				"stop_reason":   nil,
				"stop_sequence": nil,
				"usage": map[string]interface{}{
					"input_tokens":  0,
					"output_tokens": 0,
				},
			},
		})
	}
	closeBlock := func(outputIndex int) {
		if index, ok := blockByOutput[outputIndex]; ok {
			if _, open := openBlocks[index]; open {
				emitAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
					"type":  "content_block_stop",
					"index": index,
				})
				delete(openBlocks, index)
			}
		}
	}

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
		ensureStarted()

		switch eventType {
		case "response.output_item.added":
			item := respMap(event["item"])
			if item == nil {
				return false, nil
			}
			outputIndex := respInt(event["output_index"])
			switch respString(item["type"]) {
			case "message":
				index := blockIndex
				blockIndex++
				blockByOutput[outputIndex] = index
				kindByOutput[outputIndex] = "text"
				openBlocks[index] = "text"
				emitAnthropicSSE(dst, "content_block_start", map[string]interface{}{
					"type":  "content_block_start",
					"index": index,
					"content_block": map[string]interface{}{
						"type": "text",
						"text": "",
					},
				})
			case "function_call":
				hasToolUse = true
				index := blockIndex
				blockIndex++
				blockByOutput[outputIndex] = index
				kindByOutput[outputIndex] = "tool_use"
				openBlocks[index] = "tool_use"
				emitAnthropicSSE(dst, "content_block_start", map[string]interface{}{
					"type":  "content_block_start",
					"index": index,
					"content_block": map[string]interface{}{
						"type":  "tool_use",
						"id":    respString(item["call_id"]),
						"name":  respString(item["name"]),
						"input": json.RawMessage("{}"),
					},
				})
			}
		case "response.output_text.delta":
			outputIndex := respInt(event["output_index"])
			index, ok := blockByOutput[outputIndex]
			if !ok {
				return false, nil
			}
			emitAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
				"type":  "content_block_delta",
				"index": index,
				"delta": map[string]interface{}{
					"type": "text_delta",
					"text": respString(event["delta"]),
				},
			})
		case "response.function_call_arguments.delta":
			outputIndex := respInt(event["output_index"])
			index, ok := blockByOutput[outputIndex]
			if !ok {
				return false, nil
			}
			emitAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
				"type":  "content_block_delta",
				"index": index,
				"delta": map[string]interface{}{
					"type":         "input_json_delta",
					"partial_json": respString(event["delta"]),
				},
			})
		case "response.output_item.done":
			closeBlock(respInt(event["output_index"]))
		case "response.completed":
			completed = true
			response := respMap(event["response"])
			if response != nil {
				if u := respMap(response["usage"]); u != nil {
					inputTokens = respInt(u["input_tokens"])
					outputTokens = respInt(u["output_tokens"])
				}
				if respString(response["status"]) == "incomplete" {
					if d := respMap(response["incomplete_details"]); d != nil {
						switch respString(d["reason"]) {
						case "max_output_tokens":
							stopReason = "max_tokens"
						case "content_filter":
							stopReason = "refusal"
						}
					}
				}
			}
			if hasToolUse && stopReason == "end_turn" {
				stopReason = "tool_use"
			}
			for index := range openBlocks {
				emitAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
					"type":  "content_block_stop",
					"index": index,
				})
			}
			emitAnthropicSSE(dst, "message_delta", map[string]interface{}{
				"type": "message_delta",
				"delta": map[string]interface{}{
					"stop_reason": stopReason,
				},
				"usage": map[string]interface{}{
					"input_tokens":  inputTokens,
					"output_tokens": outputTokens,
				},
			})
			emitAnthropicSSE(dst, "message_stop", map[string]interface{}{"type": "message_stop"})
			au := anthropicUsage{InputTokens: inputTokens, OutputTokens: outputTokens}
			au.toUsage(usage)
			return true, nil
		case "response.failed", "error":
			completed = true
			emitAnthropicSSE(dst, "error", map[string]interface{}{
				"type": "error",
				"error": map[string]interface{}{
					"type":    "api_error",
					"message": "responses upstream error",
				},
			})
			return true, nil
		}
		return false, nil
	})
	if err != nil && !completed {
		ensureStarted()
		emitAnthropicSSE(dst, "error", map[string]interface{}{
			"type":  "error",
			"error": map[string]interface{}{"type": "api_error", "message": err.Error()},
		})
	}
	return nil
}
