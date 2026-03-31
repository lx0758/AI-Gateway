package provider

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

func (u openAIUsage) total() int {
	return u.TotalTokens
}

type OpenAICompatibleProvider struct {
	cfg *Config
}

func NewOpenAICompatibleProvider(cfg *Config) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{cfg: cfg}
}

func (m *OpenAICompatibleProvider) Name() string {
	return m.cfg.ProviderName
}

func (m *OpenAICompatibleProvider) Type() string {
	return m.cfg.ProviderType
}

func (m *OpenAICompatibleProvider) SyncModels(provider *model.Provider) ([]model.ProviderModel, error) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("OpenAI compatible base URL is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+provider.APIKey)

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
			ID      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]model.ProviderModel, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, model.ProviderModel{
			ProviderID:     provider.ID,
			ModelID:        m.ID,
			DisplayName:    m.ID,
			OwnedBy:        m.OwnedBy,
			SupportsStream: true,
			IsAvailable:    true,
			Source:         "sync",
		})
	}

	return models, nil
}

func (m *OpenAICompatibleProvider) ExecuteOpenAIRequest(c *gin.Context, pm *model.ProviderModel) (int, error) {
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

	req, err := http.NewRequest("POST", m.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
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
		return 0, fmt.Errorf("%s", string(respBody))
	}

	tokens := 0
	if m.isStreaming(resp) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		tokens = m.copyOpenAIStreaming(c.Writer, resp.Body)
	} else {
		c.Header("Content-Type", "application/json")
		tokens = m.copyOpenAIResponse(c.Writer, resp.Body)
	}
	return tokens, nil
}

func (m *OpenAICompatibleProvider) copyOpenAIStreaming(dst io.Writer, src io.Reader) int {
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
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			OpenAIUsage openAIUsage `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err == nil {
			tokens = chunk.OpenAIUsage.total()
		}
	}

	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}

	return tokens
}

func (m *OpenAICompatibleProvider) copyOpenAIResponse(dst io.Writer, src io.Reader) int {
	body, err := io.ReadAll(src)
	if err != nil {
		return 0
	}

	dst.Write(body)
	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}

	var resp struct {
		OpenAIUsage openAIUsage `json:"usage"`
	}
	tokens := 0
	if err := json.Unmarshal(body, &resp); err == nil {
		tokens = resp.OpenAIUsage.total()
	}
	return tokens
}

func (m *OpenAICompatibleProvider) ExecuteAnthropicRequest(c *gin.Context, pm *model.ProviderModel) (int, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return 0, err
	}

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
		return 0, err
	}
	anthropicReq.Model = pm.ModelID

	var anthropicMessages []map[string]interface{}
	if err := json.Unmarshal(anthropicReq.Messages, &anthropicMessages); err != nil {
		return 0, err
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
		openAIMsg := m.convertAnthropicMessageToOpenAI(msg)
		openAIMessages = append(openAIMessages, openAIMsg)
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
		return 0, err
	}

	req, err := http.NewRequest("POST", m.cfg.BaseURL+"/chat/completions", bytes.NewReader(openAIBody))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
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
		return 0, fmt.Errorf("%s", string(respBody))
	}

	tokens := 0
	if m.isStreaming(resp) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		tokens = m.streamOpenAIToAnthropic(resp.Body, c.Writer, anthropicReq.Model)
		c.Writer.Flush()
	} else {
		openAIRespBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}
		anthropicResp, tokens, err := m.convertOpenAIResponseToAnthropic(openAIRespBody)
		if err != nil {
			return tokens, err
		}
		c.Header("Content-Type", "application/json")
		c.Writer.Write(anthropicResp)
	}
	return tokens, nil
}

func (m *OpenAICompatibleProvider) extractSystemContent(system interface{}) string {
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

func (m *OpenAICompatibleProvider) convertAnthropicMessageToOpenAI(msg map[string]interface{}) map[string]interface{} {
	role, _ := msg["role"].(string)
	content := msg["content"]

	result := map[string]interface{}{
		"role": role,
	}

	switch v := content.(type) {
	case string:
		result["content"] = v
	case []interface{}:
		textParts := make([]string, 0)
		imageParts := make([]map[string]interface{}, 0)
		toolCalls := make([]map[string]interface{}, 0)

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
					return m.convertAnthropicToolResultToOpenAI(blockMap)
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
	default:
		result["content"] = v
	}
	return result
}

func (m *OpenAICompatibleProvider) convertAnthropicToolResultToOpenAI(blockMap map[string]interface{}) map[string]interface{} {
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

func (m *OpenAICompatibleProvider) convertAnthropicToolToOpenAI(tool map[string]interface{}) map[string]interface{} {
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

func (m *OpenAICompatibleProvider) convertOpenAIResponseToAnthropic(openAIResp []byte) ([]byte, int, error) {
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
		return nil, 0, fmt.Errorf("OpenAI response error")
	}

	tokens := openAI.OpenAIUsage.total()
	if len(openAI.Choices) == 0 {
		return nil, tokens, fmt.Errorf("OpenAI choices empty")
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
		return nil, tokens, fmt.Errorf("Anthropic response serialization error")
	}
	return result, tokens, nil
}

func (m *OpenAICompatibleProvider) isStreaming(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return len(resp.Header["Transfer-Encoding"]) > 0 ||
		(len(contentType) > 0 && len(contentType) >= 17 && contentType[:17] == "text/event-stream")
}

func (m *OpenAICompatibleProvider) streamOpenAIToAnthropic(src io.Reader, dst io.Writer, model string) int {
	reader := bufio.NewReader(src)
	messageID := fmt.Sprintf("msg_%s", m.generateID())
	sentMessageStart := false
	contentBlockIndex := 0
	toolIndexToContentBlock := make(map[int]int)
	inThinkingBlock := false
	inTextBlock := false
	tokens := 0

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
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			break
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

		tokens += chunk.OpenAIUsage.total()

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

			reasoning, hasReasoning := delta["reasoning_content"].(string)
			if hasReasoning && reasoning != "" {
				if !inThinkingBlock {
					if inTextBlock {
						m.writeAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
							"type":  "content_block_stop",
							"index": contentBlockIndex,
						})
						contentBlockIndex++
					}
					m.writeAnthropicSSE(dst, "content_block_start", map[string]interface{}{
						"type":  "content_block_start",
						"index": contentBlockIndex,
						"content_block": map[string]interface{}{
							"type":     "thinking",
							"thinking": "",
						},
					})
					inThinkingBlock = true
					inTextBlock = false
				}
				m.writeAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
					"type":  "content_block_delta",
					"index": contentBlockIndex,
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
						"index": contentBlockIndex,
					})
					contentBlockIndex++
					inThinkingBlock = false
				}
				if !inTextBlock {
					m.writeAnthropicSSE(dst, "content_block_start", map[string]interface{}{
						"type":  "content_block_start",
						"index": contentBlockIndex,
						"content_block": map[string]interface{}{
							"type": "text",
							"text": "",
						},
					})
					inTextBlock = true
				}
				m.writeAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
					"type":  "content_block_delta",
					"index": contentBlockIndex,
					"delta": map[string]interface{}{
						"type": "text_delta",
						"text": content,
					},
				})
			}

			if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
				for _, tc := range toolCalls {
					tcMap, ok := tc.(map[string]interface{})
					if !ok {
						continue
					}

					toolID, _ := tcMap["id"].(string)
					toolIndex := -1
					if idx, ok := tcMap["index"].(float64); ok {
						toolIndex = int(idx)
					}

					fn, hasFn := tcMap["function"].(map[string]interface{})
					var args string
					if hasFn {
						args, _ = fn["arguments"].(string)
					}

					if toolID != "" {
						if inThinkingBlock || inTextBlock {
							m.writeAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
								"type":  "content_block_stop",
								"index": contentBlockIndex,
							})
							contentBlockIndex++
							inThinkingBlock = false
							inTextBlock = false
						}

						name := ""
						if hasFn {
							name, _ = fn["name"].(string)
						}

						toolIndexToContentBlock[toolIndex] = contentBlockIndex

						m.writeAnthropicSSE(dst, "content_block_start", map[string]interface{}{
							"type":  "content_block_start",
							"index": contentBlockIndex,
							"content_block": map[string]interface{}{
								"type":  "tool_use",
								"id":    toolID,
								"name":  name,
								"input": json.RawMessage("{}"),
							},
						})

						if args != "" {
							m.writeAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
								"type":  "content_block_delta",
								"index": contentBlockIndex,
								"delta": map[string]interface{}{
									"type":         "input_json_delta",
									"partial_json": args,
								},
							})
						}
					} else if toolIndex >= 0 {
						if cbIdx, exists := toolIndexToContentBlock[toolIndex]; exists && args != "" {
							m.writeAnthropicSSE(dst, "content_block_delta", map[string]interface{}{
								"type":  "content_block_delta",
								"index": cbIdx,
								"delta": map[string]interface{}{
									"type":         "input_json_delta",
									"partial_json": args,
								},
							})
						}
					}
				}
			}

			if choice.FinishReason != "" {
				stopReason := "end_turn"
				switch choice.FinishReason {
				case "stop":
					stopReason = "end_turn"
				case "length":
					stopReason = "max_tokens"
				case "tool_calls":
					stopReason = "tool_use"
				}

				m.writeAnthropicSSE(dst, "content_block_stop", map[string]interface{}{
					"type":  "content_block_stop",
					"index": contentBlockIndex,
				})
				m.writeAnthropicSSE(dst, "message_delta", map[string]interface{}{
					"type": "message_delta",
					"delta": map[string]interface{}{
						"stop_reason": stopReason,
					},
					"usage": map[string]interface{}{
						"input_tokens":            chunk.OpenAIUsage.PromptTokens - chunk.OpenAIUsage.PromptTokensDetails.CachedTokens,
						"output_tokens":           chunk.OpenAIUsage.CompletionTokens,
						"cache_read_input_tokens": chunk.OpenAIUsage.PromptTokensDetails.CachedTokens,
					},
				})
				m.writeAnthropicSSE(dst, "message_stop", map[string]interface{}{
					"type": "message_stop",
				})
				break
			}
		}
	}
	return tokens
}

func (m *OpenAICompatibleProvider) writeAnthropicSSE(w io.Writer, eventType string, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(dataBytes))
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *OpenAICompatibleProvider) generateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 24)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
