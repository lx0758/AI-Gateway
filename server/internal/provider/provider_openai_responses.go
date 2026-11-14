package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
)

// ============================================================
// Responses 协议基础类型与工具（四格转换共用）
// ============================================================

// responsesUsage 是 Responses 协议 usage 结构。
// 注意：Responses 与 Anthropic 的 token 字段同名（input_tokens/output_tokens），
// 只有 Chat 使用 prompt_tokens/completion_tokens，见 protocol-fields.md §5.2。
type responsesUsage struct {
	InputTokens        int `json:"input_tokens"`
	OutputTokens       int `json:"output_tokens"`
	TotalTokens        int `json:"total_tokens"`
	InputTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"input_tokens_details"`
	OutputTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
}

func (u responsesUsage) toUsage(usage *Usage) {
	usage.CachedTokens = u.InputTokensDetails.CachedTokens
	usage.InputTokens = u.InputTokens - u.InputTokensDetails.CachedTokens
	if usage.InputTokens < 0 {
		usage.InputTokens = 0
	}
	usage.OutputTokens = u.OutputTokens
}

func (u responsesUsage) toMap() map[string]interface{} {
	m := map[string]interface{}{
		"input_tokens":  u.InputTokens,
		"output_tokens": u.OutputTokens,
		"total_tokens":  u.TotalTokens,
	}
	if u.InputTokensDetails.CachedTokens > 0 {
		m["input_tokens_details"] = map[string]interface{}{"cached_tokens": u.InputTokensDetails.CachedTokens}
	}
	if u.OutputTokensDetails.ReasoningTokens > 0 {
		m["output_tokens_details"] = map[string]interface{}{"reasoning_tokens": u.OutputTokensDetails.ReasoningTokens}
	}
	return m
}

func respString(v interface{}) string {
	s, _ := v.(string)
	return s
}

func respInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return 0
}

func respMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}

func respSlice(v interface{}) []interface{} {
	s, _ := v.([]interface{})
	return s
}

// responsesSSEWriter 统一负责 event 名、sequence_number 自增与 flush。
type responsesSSEWriter struct {
	w   io.Writer
	seq int
}

func newResponsesSSEWriter(w io.Writer) *responsesSSEWriter {
	return &responsesSSEWriter{w: w}
}

func (s *responsesSSEWriter) emit(eventType string, payload map[string]interface{}) {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["type"] = eventType
	payload["sequence_number"] = s.seq
	s.seq++

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", eventType, data)
	if flusher, ok := s.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// consumeSSE 逐行读取 SSE，交给 handle 处理；handle 返回 true 表示提前结束。
func consumeSSE(ctx context.Context, src io.Reader, handle func(line string) (bool, error)) error {
	reader := bufio.NewReader(src)
	errorCount := 0
	for {
		type readResult struct {
			line string
			err  error
		}
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
					return nil
				}
				errorCount++
				if errorCount >= 3 {
					return result.err
				}
				continue
			}
			stop, err := handle(result.line)
			if err != nil {
				return err
			}
			if stop {
				return nil
			}
		}
	}
}

func sseData(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "data:") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(line, "data:")), true
}

// ============================================================
// gw1 信封（Anthropic thinking 跨轮保真，见 protocol-fields.md §3.3）
// ============================================================

const gatewayEnvelopePrefix = "gw1:"

type gatewayEnvelope struct {
	P string `json:"p"`
	S string `json:"s"`
	T string `json:"t"`
}

func encodeGatewayEnvelope(anthropicBaseURL, signature, thinking string) string {
	payload, err := json.Marshal(gatewayEnvelope{P: anthropicBaseURL, S: signature, T: thinking})
	if err != nil {
		return ""
	}
	return gatewayEnvelopePrefix + base64.RawURLEncoding.EncodeToString(payload)
}

func decodeGatewayEnvelope(value string) (*gatewayEnvelope, bool) {
	if !strings.HasPrefix(value, gatewayEnvelopePrefix) {
		return nil, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, gatewayEnvelopePrefix))
	if err != nil {
		return nil, false
	}
	var env gatewayEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, false
	}
	if env.S == "" {
		return nil, false
	}
	return &env, true
}

// ============================================================
// 响应对象与 output item 构造
// ============================================================

func responsesObject(id, model, status string, createdAt int64, output []map[string]interface{}, usage map[string]interface{}, incomplete map[string]interface{}) map[string]interface{} {
	if output == nil {
		output = []map[string]interface{}{}
	}
	obj := map[string]interface{}{
		"id":                  id,
		"object":              "response",
		"created_at":          createdAt,
		"status":              status,
		"model":               model,
		"output":              output,
		"store":               false,
		"parallel_tool_calls": true,
		"tool_choice":         "auto",
		"tools":               []interface{}{},
	}
	if usage != nil {
		obj["usage"] = usage
	}
	if incomplete != nil {
		obj["incomplete_details"] = incomplete
	}
	return obj
}

func responsesMessageItem(id, text, status string) map[string]interface{} {
	return map[string]interface{}{
		"type":   "message",
		"id":     id,
		"status": status,
		"role":   "assistant",
		"content": []interface{}{
			map[string]interface{}{"type": "output_text", "text": text, "annotations": []interface{}{}},
		},
	}
}

func responsesReasoningItem(id, text, encrypted string) map[string]interface{} {
	item := map[string]interface{}{
		"type": "reasoning",
		"id":   id,
		"summary": []interface{}{
			map[string]interface{}{"type": "summary_text", "text": text},
		},
	}
	if encrypted != "" {
		item["encrypted_content"] = encrypted
	}
	return item
}

func responsesFunctionCallItem(id, callID, name, arguments, status string) map[string]interface{} {
	return map[string]interface{}{
		"type":      "function_call",
		"id":        id,
		"call_id":   callID,
		"name":      name,
		"arguments": arguments,
		"status":    status,
	}
}

func responsesIncompleteDetails(status string) map[string]interface{} {
	if status != "incomplete" {
		return nil
	}
	return map[string]interface{}{"reason": "max_output_tokens"}
}

// ============================================================
// ResponsesProvider：原生透传
// ============================================================

// ResponsesProvider 负责原生 Responses 方言（透传 + 作为转换目标）。
type ResponsesProvider struct {
	cfg *Config
}

func NewResponsesProvider(cfg *Config) *ResponsesProvider {
	return &ResponsesProvider{cfg: cfg}
}

func (m *ResponsesProvider) SyncModels(providerID uint) ([]model.ProviderModel, error) {
	baseURL := m.cfg.OpenAIResponsesBaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("OpenAI Responses base URL is required")
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
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := []model.ProviderModel{}
	for _, item := range result.Data {
		if item.ID == "" {
			continue
		}
		models = append(models, model.ProviderModel{
			ProviderID:    providerID,
			ModelID:       item.ID,
			DisplayName:   AnyString(item.Name, item.DisplayName),
			OwnedBy:       item.OwnedBy,
			ContextWindow: item.ContextLength,
		})
	}
	return models, nil
}

func (m *ResponsesProvider) ExecuteOpenAIResponsesRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
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
	bodyJson["store"] = false
	body, err = json.Marshal(bodyJson)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("OR2OR", "raw", body)
	req, err := http.NewRequest("POST", m.cfg.OpenAIResponsesBaseURL+"/responses", bytes.NewReader(body))
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
		recordError("OR2OR", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = m.copyResponsesStreaming(c.Request.Context(), c.Writer, resp.Body, usage)
	} else {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "application/json")
		err = m.copyResponsesResponse(c.Writer, resp.Body, usage)
	}
	return err
}

func (m *ResponsesProvider) isStreaming(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return len(resp.Header["Transfer-Encoding"]) > 0 ||
		(len(contentType) > 0 && strings.HasPrefix(contentType, "text/event-stream"))
}

// copyResponsesStreaming 逐行原样转发 SSE，并旁路解析 response.completed 提取 usage。
func (m *ResponsesProvider) copyResponsesStreaming(ctx context.Context, dst io.Writer, src io.Reader, usage *Usage) error {
	src, dst = recordStream("OR2OR", src, dst)
	return consumeSSE(ctx, src, func(line string) (bool, error) {
		if _, err := fmt.Fprint(dst, line); err != nil {
			return true, nil
		}
		if flusher, ok := dst.(http.Flusher); ok {
			flusher.Flush()
		}
		data, ok := sseData(line)
		if !ok || data == "" || data == "[DONE]" {
			return false, nil
		}
		var event struct {
			Type     string `json:"type"`
			Response struct {
				Usage responsesUsage `json:"usage"`
			} `json:"response"`
		}
		if err := json.Unmarshal([]byte(data), &event); err == nil && event.Type == "response.completed" {
			event.Response.Usage.toUsage(usage)
		}
		return false, nil
	})
}

func (m *ResponsesProvider) copyResponsesResponse(dst io.Writer, src io.Reader, usage *Usage) error {
	body, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	dst.Write(body)
	if flusher, ok := dst.(http.Flusher); ok {
		flusher.Flush()
	}

	var resp struct {
		Usage responsesUsage `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil
	}
	resp.Usage.toUsage(usage)
	return nil
}

// ============================================================
// 格 R→C：Responses 客户端 → Chat 上游
// ============================================================

func (m *OpenAIProvider) ExecuteOpenAIResponsesRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	chatReq, _, err := convertResponsesRequestToChat(body, pm.ModelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	chatBody, err := json.Marshal(chatReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("OR2OC", "converted", chatBody)
	req, err := http.NewRequest("POST", m.cfg.OpenAIChatCompletionsBaseURL+"/chat/completions", bytes.NewReader(chatBody))
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
		recordError("OR2OC", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = streamChatToResponses(c.Request.Context(), resp.Body, c.Writer, pm.ModelID, usage)
	} else {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		out, err := convertChatResponseToResponses(respBody, pm.ModelID, usage)
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

func convertResponsesRequestToChat(body []byte, upstreamModel string) (map[string]interface{}, bool, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, false, err
	}

	stream := false
	if v, ok := req["stream"].(bool); ok {
		stream = v
	}

	chatReq := map[string]interface{}{"model": upstreamModel}
	messages := make([]map[string]interface{}, 0)
	if instructions := respString(req["instructions"]); instructions != "" {
		messages = append(messages, map[string]interface{}{"role": "system", "content": instructions})
	}
	messages = append(messages, convertResponsesInputToChatMessages(req["input"])...)
	chatReq["messages"] = messages

	if v, ok := req["max_output_tokens"]; ok && v != nil {
		chatReq["max_tokens"] = v
	}
	if v, ok := req["temperature"]; ok {
		chatReq["temperature"] = v
	}
	if v, ok := req["top_p"]; ok {
		chatReq["top_p"] = v
	}
	if tools := convertResponsesToolsToChat(req["tools"]); len(tools) > 0 {
		chatReq["tools"] = tools
	}
	if v, ok := req["tool_choice"]; ok {
		if tc := convertResponsesToolChoiceToChat(v); tc != nil {
			chatReq["tool_choice"] = tc
		}
	}
	if stream {
		chatReq["stream"] = true
		chatReq["stream_options"] = map[string]interface{}{"include_usage": true}
	}
	return chatReq, stream, nil
}

func convertResponsesInputToChatMessages(input interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)

	var assistantTexts []string
	var assistantReasoning []string
	var assistantToolCalls []map[string]interface{}
	flushAssistant := func() {
		if len(assistantTexts) == 0 && len(assistantReasoning) == 0 && len(assistantToolCalls) == 0 {
			return
		}
		msg := map[string]interface{}{"role": "assistant", "content": nil}
		if len(assistantTexts) > 0 {
			msg["content"] = strings.Join(assistantTexts, "\n")
		}
		if len(assistantToolCalls) > 0 {
			msg["tool_calls"] = assistantToolCalls
		}
		if len(assistantReasoning) > 0 {
			msg["reasoning_content"] = strings.Join(assistantReasoning, "")
		}
		out = append(out, msg)
		assistantTexts = nil
		assistantReasoning = nil
		assistantToolCalls = nil
	}

	switch v := input.(type) {
	case string:
		if v != "" {
			out = append(out, map[string]interface{}{"role": "user", "content": v})
		}
		return out
	case []interface{}:
		for _, raw := range v {
			item := respMap(raw)
			if item == nil {
				continue
			}
			switch respString(item["type"]) {
			case "", "message":
				role := respString(item["role"])
				if role == "developer" {
					role = "system"
				}
				if role == "assistant" {
					assistantTexts = append(assistantTexts, extractResponsesTexts(item["content"])...)
					continue
				}
				flushAssistant()
				content := convertResponsesContentToChat(item["content"])
				if content == nil {
					continue
				}
				out = append(out, map[string]interface{}{"role": role, "content": content})
			case "function_call":
				assistantToolCalls = append(assistantToolCalls, map[string]interface{}{
					"id":   respString(item["call_id"]),
					"type": "function",
					"function": map[string]interface{}{
						"name":      respString(item["name"]),
						"arguments": respString(item["arguments"]),
					},
				})
			case "function_call_output":
				flushAssistant()
				out = append(out, map[string]interface{}{
					"role":         "tool",
					"tool_call_id": respString(item["call_id"]),
					"content":      outputToText(item["output"]),
				})
			case "reasoning":
				if text := extractResponsesReasoningText(item); text != "" {
					assistantReasoning = append(assistantReasoning, text)
				}
			default:
				// 未知 item（web_search_call 等托管工具）优雅跳过
			}
		}
	}
	flushAssistant()
	return out
}

func convertResponsesContentToChat(content interface{}) interface{} {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		texts := make([]string, 0)
		parts := make([]map[string]interface{}, 0)
		hasImage := false
		for _, raw := range v {
			part := respMap(raw)
			if part == nil {
				continue
			}
			switch respString(part["type"]) {
			case "input_text", "output_text", "text":
				if text := respString(part["text"]); text != "" {
					texts = append(texts, text)
					parts = append(parts, map[string]interface{}{"type": "text", "text": text})
				}
			case "input_image", "image_url":
				url := respString(part["image_url"])
				if url == "" {
					if m := respMap(part["image_url"]); m != nil {
						url = respString(m["url"])
					}
				}
				if url != "" {
					hasImage = true
					parts = append(parts, map[string]interface{}{
						"type":      "image_url",
						"image_url": map[string]interface{}{"url": url},
					})
				}
			}
		}
		if hasImage {
			return parts
		}
		if len(texts) > 0 {
			return strings.Join(texts, "\n")
		}
	}
	return nil
}

func extractResponsesTexts(content interface{}) []string {
	switch v := content.(type) {
	case string:
		if v != "" {
			return []string{v}
		}
	case []interface{}:
		texts := make([]string, 0)
		for _, raw := range v {
			part := respMap(raw)
			if part == nil {
				continue
			}
			switch respString(part["type"]) {
			case "input_text", "output_text", "text":
				if text := respString(part["text"]); text != "" {
					texts = append(texts, text)
				}
			}
		}
		return texts
	}
	return nil
}

func extractResponsesReasoningText(item map[string]interface{}) string {
	// 信封存在则优先取全文（chat 上游不需要签名）
	if env, ok := decodeGatewayEnvelope(respString(item["encrypted_content"])); ok {
		return env.T
	}
	var texts []string
	for _, raw := range respSlice(item["summary"]) {
		if part := respMap(raw); part != nil {
			if text := respString(part["text"]); text != "" {
				texts = append(texts, text)
			}
		}
	}
	return strings.Join(texts, "")
}

func outputToText(output interface{}) string {
	switch v := output.(type) {
	case string:
		return v
	case []interface{}:
		texts := make([]string, 0)
		for _, raw := range v {
			if part := respMap(raw); part != nil {
				switch respString(part["type"]) {
				case "input_text", "output_text", "text":
					if text := respString(part["text"]); text != "" {
						texts = append(texts, text)
					}
				}
			}
		}
		if len(texts) > 0 {
			return strings.Join(texts, "\n")
		}
	}
	if output == nil {
		return ""
	}
	return fmt.Sprintf("%v", output)
}

func convertResponsesToolsToChat(tools interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	for _, raw := range respSlice(tools) {
		tool := respMap(raw)
		if tool == nil || respString(tool["type"]) != "function" {
			continue
		}
		fn := map[string]interface{}{"name": respString(tool["name"])}
		if desc := respString(tool["description"]); desc != "" {
			fn["description"] = desc
		}
		if params, ok := tool["parameters"]; ok {
			fn["parameters"] = params
		}
		out = append(out, map[string]interface{}{"type": "function", "function": fn})
	}
	return out
}

func convertResponsesToolChoiceToChat(choice interface{}) interface{} {
	switch v := choice.(type) {
	case string:
		return v
	case map[string]interface{}:
		if respString(v["type"]) == "function" {
			return map[string]interface{}{
				"type":     "function",
				"function": map[string]interface{}{"name": respString(v["name"])},
			}
		}
	}
	return nil
}

func convertChatResponseToResponses(body []byte, model string, usage *Usage) ([]byte, error) {
	var chat struct {
		ID      string `json:"id"`
		Choices []struct {
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		} `json:"choices"`
		Usage openAIUsage `json:"usage"`
	}
	if err := json.Unmarshal(body, &chat); err != nil {
		return nil, fmt.Errorf("OpenAI response error")
	}
	if len(chat.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI choices empty")
	}

	choice := chat.Choices[0]
	output := make([]map[string]interface{}, 0)
	if reasoning := respString(choice.Message["reasoning_content"]); reasoning != "" {
		output = append(output, responsesReasoningItem("rs_"+generateRandomID(), reasoning, ""))
	}
	if content := respString(choice.Message["content"]); content != "" {
		output = append(output, responsesMessageItem("msg_"+generateRandomID(), content, "completed"))
	}
	if toolCalls := respSlice(choice.Message["tool_calls"]); len(toolCalls) > 0 {
		for _, raw := range toolCalls {
			tc := respMap(raw)
			if tc == nil {
				continue
			}
			fn := respMap(tc["function"])
			output = append(output, responsesFunctionCallItem(
				"fc_"+generateRandomID(),
				respString(tc["id"]),
				respString(fn["name"]),
				respString(fn["arguments"]),
				"completed",
			))
		}
	}

	status := "completed"
	incomplete := map[string]interface{}(nil)
	switch choice.FinishReason {
	case "length":
		status = "incomplete"
		incomplete = map[string]interface{}{"reason": "max_output_tokens"}
	case "content_filter":
		status = "incomplete"
		incomplete = map[string]interface{}{"reason": "content_filter"}
	}

	resp := responsesObject("resp_"+generateRandomID(), model, status, time.Now().Unix(), output, chat.Usage.UsageToResponsesMap(), incomplete)
	out, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	chat.Usage.toUsage(usage)
	return out, nil
}

// UsageToResponsesMap 将 chat usage 映射为 Responses usage（prompt/completion → input/output）。
func (u openAIUsage) UsageToResponsesMap() map[string]interface{} {
	ru := responsesUsage{
		InputTokens:  u.PromptTokens,
		OutputTokens: u.CompletionTokens,
		TotalTokens:  u.TotalTokens,
	}
	ru.InputTokensDetails.CachedTokens = u.PromptTokensDetails.CachedTokens
	ru.OutputTokensDetails.ReasoningTokens = u.CompletionTokensDetails.ReasoningTokens
	return ru.toMap()
}

// ---- 流式：Chat chunk → Responses 事件 ----

type responsesStreamItem struct {
	kind        string // message / reasoning / function_call
	id          string
	outputIndex int
	text        strings.Builder
	callID      string
	name        string
	args        strings.Builder
	encrypted   string
	closed      bool
}

func (it *responsesStreamItem) item(status string) map[string]interface{} {
	switch it.kind {
	case "message":
		return responsesMessageItem(it.id, it.text.String(), status)
	case "reasoning":
		return responsesReasoningItem(it.id, it.text.String(), it.encrypted)
	default:
		return responsesFunctionCallItem(it.id, it.callID, it.name, it.args.String(), status)
	}
}

type chatToResponsesState struct {
	writer      *responsesSSEWriter
	responseID  string
	model       string
	createdAt   int64
	started     bool
	completed   bool
	items       []*responsesStreamItem
	message     *responsesStreamItem
	reasoning   *responsesStreamItem
	toolItems   map[int]*responsesStreamItem
	usage       responsesUsage
	finish      string
	outputCount int
}

func streamChatToResponses(ctx context.Context, src io.Reader, dst io.Writer, model string, usage *Usage) error {
	state := &chatToResponsesState{
		writer:     newResponsesSSEWriter(dst),
		responseID: "resp_" + generateRandomID(),
		model:      model,
		createdAt:  time.Now().Unix(),
		toolItems:  make(map[int]*responsesStreamItem),
	}

	err := consumeSSE(ctx, src, func(line string) (bool, error) {
		data, ok := sseData(line)
		if !ok || data == "" {
			return false, nil
		}
		if data == "[DONE]" {
			state.complete(usage)
			return true, nil
		}
		state.handleChunk(data)
		return false, nil
	})
	if err != nil {
		state.fail(err.Error())
		return nil
	}
	if !state.completed {
		if state.finish != "" {
			state.complete(usage)
		} else {
			state.fail("upstream stream ended unexpectedly")
		}
	}
	return nil
}

func (s *chatToResponsesState) ensureStarted() {
	if s.started {
		return
	}
	s.started = true
	s.writer.emit("response.created", map[string]interface{}{
		"response": responsesObject(s.responseID, s.model, "in_progress", s.createdAt, nil, nil, nil),
	})
	s.writer.emit("response.in_progress", map[string]interface{}{
		"response": responsesObject(s.responseID, s.model, "in_progress", s.createdAt, nil, nil, nil),
	})
}

func (s *chatToResponsesState) newItem(kind string) *responsesStreamItem {
	it := &responsesStreamItem{kind: kind, outputIndex: s.outputCount}
	s.outputCount++
	switch kind {
	case "message":
		it.id = "msg_" + generateRandomID()
	case "reasoning":
		it.id = "rs_" + generateRandomID()
	default:
		it.id = "fc_" + generateRandomID()
	}
	s.items = append(s.items, it)
	return it
}

func (s *chatToResponsesState) handleChunk(data string) {
	var chunk struct {
		Choices []struct {
			Delta        map[string]interface{} `json:"delta"`
			FinishReason string                 `json:"finish_reason"`
		} `json:"choices"`
		Usage openAIUsage `json:"usage"`
	}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return
	}
	s.ensureStarted()

	if chunk.Usage.TotalTokens > 0 {
		s.usage = responsesUsage{
			InputTokens:  chunk.Usage.PromptTokens,
			OutputTokens: chunk.Usage.CompletionTokens,
			TotalTokens:  chunk.Usage.TotalTokens,
		}
		s.usage.InputTokensDetails.CachedTokens = chunk.Usage.PromptTokensDetails.CachedTokens
		s.usage.OutputTokensDetails.ReasoningTokens = chunk.Usage.CompletionTokensDetails.ReasoningTokens
	}

	if len(chunk.Choices) == 0 {
		return
	}
	choice := chunk.Choices[0]
	delta := choice.Delta

	reasoning := respString(delta["reasoning_content"])
	if reasoning == "" {
		reasoning = respString(delta["reasoning"])
	}
	if reasoning != "" {
		if s.reasoning == nil {
			s.reasoning = s.newItem("reasoning")
			s.writer.emit("response.output_item.added", map[string]interface{}{
				"output_index": s.reasoning.outputIndex,
				"item":         responsesReasoningItem(s.reasoning.id, "", ""),
			})
			s.writer.emit("response.reasoning_summary_part.added", map[string]interface{}{
				"item_id":       s.reasoning.id,
				"output_index":  s.reasoning.outputIndex,
				"summary_index": 0,
				"part":          map[string]interface{}{"type": "summary_text", "text": ""},
			})
		}
		s.reasoning.text.WriteString(reasoning)
		s.writer.emit("response.reasoning_summary_text.delta", map[string]interface{}{
			"item_id":       s.reasoning.id,
			"output_index":  s.reasoning.outputIndex,
			"summary_index": 0,
			"delta":         reasoning,
		})
	}

	content := respString(delta["content"])
	if content != "" {
		if s.message == nil {
			s.message = s.newItem("message")
			s.writer.emit("response.output_item.added", map[string]interface{}{
				"output_index": s.message.outputIndex,
				"item":         responsesMessageItem(s.message.id, "", "in_progress"),
			})
			s.writer.emit("response.content_part.added", map[string]interface{}{
				"item_id":       s.message.id,
				"output_index":  s.message.outputIndex,
				"content_index": 0,
				"part":          map[string]interface{}{"type": "output_text", "text": "", "annotations": []interface{}{}},
			})
		}
		s.message.text.WriteString(content)
		s.writer.emit("response.output_text.delta", map[string]interface{}{
			"item_id":       s.message.id,
			"output_index":  s.message.outputIndex,
			"content_index": 0,
			"delta":         content,
		})
	}

	for _, raw := range respSlice(delta["tool_calls"]) {
		tc := respMap(raw)
		if tc == nil {
			continue
		}
		index := respInt(tc["index"])
		it := s.toolItems[index]
		if it == nil {
			it = s.newItem("function_call")
			it.callID = respString(tc["id"])
			if fn := respMap(tc["function"]); fn != nil {
				it.name = respString(fn["name"])
			}
			s.toolItems[index] = it
			s.writer.emit("response.output_item.added", map[string]interface{}{
				"output_index": it.outputIndex,
				"item":         responsesFunctionCallItem(it.id, it.callID, it.name, "", "in_progress"),
			})
		}
		fn := respMap(tc["function"])
		if fn == nil {
			continue
		}
		if args := respString(fn["arguments"]); args != "" {
			it.args.WriteString(args)
			s.writer.emit("response.function_call_arguments.delta", map[string]interface{}{
				"item_id":      it.id,
				"output_index": it.outputIndex,
				"delta":        args,
			})
		}
	}

	if choice.FinishReason != "" {
		s.finish = choice.FinishReason
		s.closeItems()
	}
}

func (s *chatToResponsesState) closeItems() {
	if s.reasoning != nil && !s.reasoning.closed {
		s.reasoning.closed = true
		s.writer.emit("response.reasoning_summary_part.done", map[string]interface{}{
			"item_id":       s.reasoning.id,
			"output_index":  s.reasoning.outputIndex,
			"summary_index": 0,
			"part":          map[string]interface{}{"type": "summary_text", "text": s.reasoning.text.String()},
		})
		s.writer.emit("response.output_item.done", map[string]interface{}{
			"output_index": s.reasoning.outputIndex,
			"item":         responsesReasoningItem(s.reasoning.id, s.reasoning.text.String(), ""),
		})
	}
	if s.message != nil && !s.message.closed {
		s.message.closed = true
		s.writer.emit("response.content_part.done", map[string]interface{}{
			"item_id":       s.message.id,
			"output_index":  s.message.outputIndex,
			"content_index": 0,
			"part":          map[string]interface{}{"type": "output_text", "text": s.message.text.String(), "annotations": []interface{}{}},
		})
		s.writer.emit("response.output_item.done", map[string]interface{}{
			"output_index": s.message.outputIndex,
			"item":         responsesMessageItem(s.message.id, s.message.text.String(), "completed"),
		})
	}
	for _, it := range s.toolItems {
		if it.closed {
			continue
		}
		it.closed = true
		s.writer.emit("response.output_item.done", map[string]interface{}{
			"output_index": it.outputIndex,
			"item":         responsesFunctionCallItem(it.id, it.callID, it.name, it.args.String(), "completed"),
		})
	}
}

func (s *chatToResponsesState) complete(usage *Usage) {
	if s.completed {
		return
	}
	s.ensureStarted()
	s.closeItems()

	status := "completed"
	incomplete := map[string]interface{}(nil)
	if s.finish == "length" {
		status = "incomplete"
		incomplete = map[string]interface{}{"reason": "max_output_tokens"}
	} else if s.finish == "content_filter" {
		status = "incomplete"
		incomplete = map[string]interface{}{"reason": "content_filter"}
	}

	output := make([]map[string]interface{}, 0, len(s.items))
	for _, it := range s.items {
		itemStatus := "completed"
		output = append(output, it.item(itemStatus))
	}
	s.writer.emit("response.completed", map[string]interface{}{
		"response": responsesObject(s.responseID, s.model, status, s.createdAt, output, s.usage.toMap(), incomplete),
	})
	s.usage.toUsage(usage)
	s.completed = true
}

func (s *chatToResponsesState) fail(message string) {
	if s.completed {
		return
	}
	s.ensureStarted()
	s.writer.emit("response.failed", map[string]interface{}{
		"response": responsesObject(s.responseID, s.model, "failed", s.createdAt, nil, nil, map[string]interface{}{"reason": message}),
	})
	s.completed = true
}

// ============================================================
// 格 R→A：Responses 客户端 → Anthropic 上游
// ============================================================

func (m *AnthropicProvider) ExecuteOpenAIResponsesRequest(c *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	anthropicReq, _, err := convertResponsesRequestToAnthropic(body, pm.ModelID, m.cfg.AnthropicMessagesBaseURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	anthropicBody, err := json.Marshal(anthropicReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	recordBody("OR2AM", "converted", anthropicBody)
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
		recordError("OR2AM", resp.StatusCode, respBody)
		return &ProviderError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if m.isStreaming(resp) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err = streamAnthropicToResponses(c.Request.Context(), resp.Body, c.Writer, pm.ModelID, m.cfg.AnthropicMessagesBaseURL, usage)
	} else {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
		out, err := convertAnthropicResponseToResponses(respBody, pm.ModelID, m.cfg.AnthropicMessagesBaseURL, usage)
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

func convertResponsesRequestToAnthropic(body []byte, upstreamModel, anthropicBaseURL string) (map[string]interface{}, bool, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, false, err
	}

	stream := false
	if v, ok := req["stream"].(bool); ok {
		stream = v
	}

	var systemParts []string
	if instructions := respString(req["instructions"]); instructions != "" {
		systemParts = append(systemParts, instructions)
	}

	messages := make([]map[string]interface{}, 0)
	assistantBlocks := make([]map[string]interface{}, 0)
	pendingToolResults := make([]map[string]interface{}, 0)

	flushAssistant := func() {
		if len(assistantBlocks) == 0 {
			return
		}
		messages = append(messages, map[string]interface{}{"role": "assistant", "content": assistantBlocks})
		assistantBlocks = make([]map[string]interface{}, 0)
	}
	flushToolResults := func() {
		if len(pendingToolResults) == 0 {
			return
		}
		flushAssistant()
		messages = append(messages, map[string]interface{}{"role": "user", "content": pendingToolResults})
		pendingToolResults = make([]map[string]interface{}, 0)
	}

	input := req["input"]
	switch v := input.(type) {
	case string:
		if v != "" {
			messages = append(messages, map[string]interface{}{
				"role":    "user",
				"content": []map[string]interface{}{{"type": "text", "text": v}},
			})
		}
	case []interface{}:
		for _, raw := range v {
			item := respMap(raw)
			if item == nil {
				continue
			}
			switch respString(item["type"]) {
			case "", "message":
				role := respString(item["role"])
				if role == "system" || role == "developer" {
					if text := strings.Join(extractResponsesTexts(item["content"]), "\n"); text != "" {
						systemParts = append(systemParts, text)
					}
					continue
				}
				if role == "assistant" {
					flushToolResults()
					assistantBlocks = append(assistantBlocks, convertResponsesContentToAnthropicBlocks(item["content"])...)
					continue
				}
				flushToolResults()
				flushAssistant()
				messages = append(messages, map[string]interface{}{
					"role":    "user",
					"content": convertResponsesContentToAnthropicBlocks(item["content"]),
				})
			case "function_call":
				flushToolResults()
				input := map[string]interface{}{}
				if args := respString(item["arguments"]); args != "" {
					if err := json.Unmarshal([]byte(args), &input); err != nil {
						input = map[string]interface{}{}
					}
				}
				assistantBlocks = append(assistantBlocks, map[string]interface{}{
					"type":  "tool_use",
					"id":    respString(item["call_id"]),
					"name":  respString(item["name"]),
					"input": input,
				})
			case "function_call_output":
				flushAssistant()
				pendingToolResults = append(pendingToolResults, map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": respString(item["call_id"]),
					"content":     outputToText(item["output"]),
				})
			case "reasoning":
				// 仅当信封属于当前 Anthropic 上游时按原序重建 thinking 块；否则丢弃
				env, ok := decodeGatewayEnvelope(respString(item["encrypted_content"]))
				if !ok || env.P != anthropicBaseURL {
					continue
				}
				flushToolResults()
				assistantBlocks = append(assistantBlocks, map[string]interface{}{
					"type":      "thinking",
					"thinking":  env.T,
					"signature": env.S,
				})
			default:
				// 未知 item 优雅跳过
			}
		}
	}
	flushToolResults()
	flushAssistant()

	maxTokens := respInt(req["max_output_tokens"])
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	anthropicReq := map[string]interface{}{
		"model":      upstreamModel,
		"max_tokens": maxTokens,
		"messages":   messages,
		"stream":     stream,
	}
	if len(systemParts) > 0 {
		anthropicReq["system"] = strings.Join(systemParts, "\n")
	}
	if tools := convertResponsesToolsToAnthropic(req["tools"]); len(tools) > 0 {
		anthropicReq["tools"] = tools
	}
	if v, ok := req["tool_choice"]; ok {
		if tc := convertResponsesToolChoiceToAnthropic(v); tc != nil {
			anthropicReq["tool_choice"] = tc
		}
	}
	return anthropicReq, stream, nil
}

func convertResponsesContentToAnthropicBlocks(content interface{}) []map[string]interface{} {
	blocks := make([]map[string]interface{}, 0)
	switch v := content.(type) {
	case string:
		if v != "" {
			blocks = append(blocks, map[string]interface{}{"type": "text", "text": v})
		}
	case []interface{}:
		for _, raw := range v {
			part := respMap(raw)
			if part == nil {
				continue
			}
			switch respString(part["type"]) {
			case "input_text", "output_text", "text":
				if text := respString(part["text"]); text != "" {
					blocks = append(blocks, map[string]interface{}{"type": "text", "text": text})
				}
			case "input_image", "image_url":
				url := respString(part["image_url"])
				if url == "" {
					if m := respMap(part["image_url"]); m != nil {
						url = respString(m["url"])
					}
				}
				if url == "" {
					continue
				}
				if strings.HasPrefix(url, "data:") {
					mediaType, data := parseDataURLValue(url)
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
						"type":   "image",
						"source": map[string]interface{}{"type": "url", "url": url},
					})
				}
			}
		}
	}
	return blocks
}

func convertResponsesToolsToAnthropic(tools interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	for _, raw := range respSlice(tools) {
		tool := respMap(raw)
		if tool == nil || respString(tool["type"]) != "function" {
			continue
		}
		converted := map[string]interface{}{"name": respString(tool["name"])}
		if desc := respString(tool["description"]); desc != "" {
			converted["description"] = desc
		}
		if params, ok := tool["parameters"]; ok {
			converted["input_schema"] = params
		}
		out = append(out, converted)
	}
	return out
}

func convertResponsesToolChoiceToAnthropic(choice interface{}) interface{} {
	switch v := choice.(type) {
	case string:
		switch v {
		case "required":
			return map[string]interface{}{"type": "any"}
		case "none":
			return map[string]interface{}{"type": "none"}
		case "auto":
			return map[string]interface{}{"type": "auto"}
		}
	case map[string]interface{}:
		if respString(v["type"]) == "function" {
			return map[string]interface{}{"type": "tool", "name": respString(v["name"])}
		}
	}
	return nil
}

func convertAnthropicResponseToResponses(body []byte, model, anthropicBaseURL string, usage *Usage) ([]byte, error) {
	var anthropic struct {
		ID         string                   `json:"id"`
		Content    []map[string]interface{} `json:"content"`
		StopReason string                   `json:"stop_reason"`
		Usage      anthropicUsage           `json:"usage"`
	}
	if err := json.Unmarshal(body, &anthropic); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	output := make([]map[string]interface{}, 0)
	texts := make([]string, 0)
	for _, block := range anthropic.Content {
		switch respString(block["type"]) {
		case "thinking":
			thinking := respString(block["thinking"])
			signature := respString(block["signature"])
			encrypted := encodeGatewayEnvelope(anthropicBaseURL, signature, thinking)
			output = append(output, responsesReasoningItem("rs_"+generateRandomID(), thinking, encrypted))
		case "text":
			if text := respString(block["text"]); text != "" {
				texts = append(texts, text)
			}
		case "tool_use":
			args := ""
			if input, ok := block["input"]; ok {
				if b, err := json.Marshal(input); err == nil {
					args = string(b)
				}
			}
			output = append(output, responsesFunctionCallItem(
				"fc_"+generateRandomID(),
				respString(block["id"]),
				respString(block["name"]),
				args,
				"completed",
			))
		}
	}
	if len(texts) > 0 {
		// 文本 item 按 Anthropic 内容顺序参与，但为简化放在思考之后、工具之前
		insertAt := 0
		for i, item := range output {
			if respString(item["type"]) == "reasoning" {
				insertAt = i + 1
			}
		}
		message := responsesMessageItem("msg_"+generateRandomID(), strings.Join(texts, ""), "completed")
		output = append(output[:insertAt], append([]map[string]interface{}{message}, output[insertAt:]...)...)
	}

	status, incomplete := "completed", map[string]interface{}(nil)
	switch anthropic.StopReason {
	case "max_tokens":
		status = "incomplete"
		incomplete = map[string]interface{}{"reason": "max_output_tokens"}
	case "refusal":
		status = "incomplete"
		incomplete = map[string]interface{}{"reason": "content_filter"}
	}

	ru := responsesUsage{InputTokens: anthropic.Usage.InputTokens, OutputTokens: anthropic.Usage.OutputTokens}
	ru.TotalTokens = anthropic.Usage.total()
	resp := responsesObject("resp_"+generateRandomID(), model, status, time.Now().Unix(), output, ru.toMap(), incomplete)
	out, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	anthropic.Usage.toUsage(usage)
	return out, nil
}

// ---- 流式：Anthropic 事件 → Responses 事件 ----

type anthropicBlockState struct {
	kind      string // text / thinking / tool_use
	item      *responsesStreamItem
	index     int
	signature string
}

type anthropicToResponsesState struct {
	writer        *responsesSSEWriter
	responseID    string
	model         string
	anthropicBase string
	createdAt     int64
	started       bool
	completed     bool
	items         []*responsesStreamItem
	blocks        map[int]*anthropicBlockState
	inputTokens   int
	outputTokens  int
	stopReason    string
	outputCount   int
}

func streamAnthropicToResponses(ctx context.Context, src io.Reader, dst io.Writer, model, anthropicBaseURL string, usage *Usage) error {
	state := &anthropicToResponsesState{
		writer:        newResponsesSSEWriter(dst),
		responseID:    "resp_" + generateRandomID(),
		model:         model,
		anthropicBase: anthropicBaseURL,
		createdAt:     time.Now().Unix(),
		blocks:        make(map[int]*anthropicBlockState),
	}

	err := consumeSSE(ctx, src, func(line string) (bool, error) {
		data, ok := sseData(line)
		if !ok || data == "" {
			return false, nil
		}
		var event struct {
			Type         string                 `json:"type"`
			Index        int                    `json:"index"`
			ContentBlock map[string]interface{} `json:"content_block"`
			Delta        map[string]interface{} `json:"delta"`
			Message      map[string]interface{} `json:"message"`
			Usage        anthropicUsage         `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return false, nil
		}
		state.handleEvent(event)
		if event.Type == "message_stop" {
			state.complete(usage)
			return true, nil
		}
		if event.Type == "error" {
			state.fail("anthropic upstream error")
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		state.fail(err.Error())
		return nil
	}
	if !state.completed {
		state.complete(usage)
	}
	return nil
}

func (s *anthropicToResponsesState) ensureStarted() {
	if s.started {
		return
	}
	s.started = true
	obj := responsesObject(s.responseID, s.model, "in_progress", s.createdAt, nil, nil, nil)
	s.writer.emit("response.created", map[string]interface{}{"response": obj})
	s.writer.emit("response.in_progress", map[string]interface{}{"response": obj})
}

func (s *anthropicToResponsesState) newItem(kind string) *responsesStreamItem {
	it := &responsesStreamItem{kind: kind, outputIndex: s.outputCount}
	s.outputCount++
	switch kind {
	case "message":
		it.id = "msg_" + generateRandomID()
	case "reasoning":
		it.id = "rs_" + generateRandomID()
	default:
		it.id = "fc_" + generateRandomID()
	}
	s.items = append(s.items, it)
	return it
}

func (s *anthropicToResponsesState) handleEvent(event struct {
	Type         string                 `json:"type"`
	Index        int                    `json:"index"`
	ContentBlock map[string]interface{} `json:"content_block"`
	Delta        map[string]interface{} `json:"delta"`
	Message      map[string]interface{} `json:"message"`
	Usage        anthropicUsage         `json:"usage"`
}) {
	s.ensureStarted()

	switch event.Type {
	case "message_start":
		if event.Message != nil {
			if u := respMap(event.Message["usage"]); u != nil {
				s.inputTokens = respInt(u["input_tokens"])
			}
		}
	case "content_block_start":
		blockType := respString(event.ContentBlock["type"])
		switch blockType {
		case "text":
			it := s.newItem("message")
			s.blocks[event.Index] = &anthropicBlockState{kind: "text", item: it, index: event.Index}
			s.writer.emit("response.output_item.added", map[string]interface{}{
				"output_index": it.outputIndex,
				"item":         responsesMessageItem(it.id, "", "in_progress"),
			})
			s.writer.emit("response.content_part.added", map[string]interface{}{
				"item_id":       it.id,
				"output_index":  it.outputIndex,
				"content_index": 0,
				"part":          map[string]interface{}{"type": "output_text", "text": "", "annotations": []interface{}{}},
			})
		case "thinking":
			it := s.newItem("reasoning")
			s.blocks[event.Index] = &anthropicBlockState{kind: "thinking", item: it, index: event.Index}
			s.writer.emit("response.output_item.added", map[string]interface{}{
				"output_index": it.outputIndex,
				"item":         responsesReasoningItem(it.id, "", ""),
			})
			s.writer.emit("response.reasoning_summary_part.added", map[string]interface{}{
				"item_id":       it.id,
				"output_index":  it.outputIndex,
				"summary_index": 0,
				"part":          map[string]interface{}{"type": "summary_text", "text": ""},
			})
		case "tool_use":
			it := s.newItem("function_call")
			it.callID = respString(event.ContentBlock["id"])
			it.name = respString(event.ContentBlock["name"])
			s.blocks[event.Index] = &anthropicBlockState{kind: "tool_use", item: it, index: event.Index}
			s.writer.emit("response.output_item.added", map[string]interface{}{
				"output_index": it.outputIndex,
				"item":         responsesFunctionCallItem(it.id, it.callID, it.name, "", "in_progress"),
			})
		}
	case "content_block_delta":
		block := s.blocks[event.Index]
		if block == nil {
			return
		}
		switch respString(event.Delta["type"]) {
		case "text_delta":
			text := respString(event.Delta["text"])
			block.item.text.WriteString(text)
			s.writer.emit("response.output_text.delta", map[string]interface{}{
				"item_id":       block.item.id,
				"output_index":  block.item.outputIndex,
				"content_index": 0,
				"delta":         text,
			})
		case "thinking_delta":
			thinking := respString(event.Delta["thinking"])
			block.item.text.WriteString(thinking)
			s.writer.emit("response.reasoning_summary_text.delta", map[string]interface{}{
				"item_id":       block.item.id,
				"output_index":  block.item.outputIndex,
				"summary_index": 0,
				"delta":         thinking,
			})
		case "signature_delta":
			block.signature += respString(event.Delta["signature"])
		case "input_json_delta":
			args := respString(event.Delta["partial_json"])
			block.item.args.WriteString(args)
			s.writer.emit("response.function_call_arguments.delta", map[string]interface{}{
				"item_id":      block.item.id,
				"output_index": block.item.outputIndex,
				"delta":        args,
			})
		}
	case "content_block_stop":
		s.closeBlock(event.Index)
	case "message_delta":
		if stop := respString(event.Delta["stop_reason"]); stop != "" {
			s.stopReason = stop
		}
		s.outputTokens = event.Usage.OutputTokens
		if event.Usage.InputTokens > 0 {
			s.inputTokens = event.Usage.InputTokens
		}
	}
}

func (s *anthropicToResponsesState) closeBlock(index int) {
	block := s.blocks[index]
	if block == nil || block.item.closed {
		return
	}
	block.item.closed = true
	switch block.kind {
	case "text":
		s.writer.emit("response.content_part.done", map[string]interface{}{
			"item_id":       block.item.id,
			"output_index":  block.item.outputIndex,
			"content_index": 0,
			"part":          map[string]interface{}{"type": "output_text", "text": block.item.text.String(), "annotations": []interface{}{}},
		})
		s.writer.emit("response.output_item.done", map[string]interface{}{
			"output_index": block.item.outputIndex,
			"item":         responsesMessageItem(block.item.id, block.item.text.String(), "completed"),
		})
	case "thinking":
		block.item.encrypted = encodeGatewayEnvelope(s.anthropicBase, block.signature, block.item.text.String())
		s.writer.emit("response.reasoning_summary_part.done", map[string]interface{}{
			"item_id":       block.item.id,
			"output_index":  block.item.outputIndex,
			"summary_index": 0,
			"part":          map[string]interface{}{"type": "summary_text", "text": block.item.text.String()},
		})
		s.writer.emit("response.output_item.done", map[string]interface{}{
			"output_index": block.item.outputIndex,
			"item":         responsesReasoningItem(block.item.id, block.item.text.String(), block.item.encrypted),
		})
	case "tool_use":
		s.writer.emit("response.output_item.done", map[string]interface{}{
			"output_index": block.item.outputIndex,
			"item":         responsesFunctionCallItem(block.item.id, block.item.callID, block.item.name, block.item.args.String(), "completed"),
		})
	}
}

func (s *anthropicToResponsesState) complete(usage *Usage) {
	if s.completed {
		return
	}
	s.ensureStarted()
	for index := range s.blocks {
		s.closeBlock(index)
	}

	status := "completed"
	incomplete := map[string]interface{}(nil)
	if s.stopReason == "max_tokens" {
		status = "incomplete"
		incomplete = map[string]interface{}{"reason": "max_output_tokens"}
	} else if s.stopReason == "refusal" {
		status = "incomplete"
		incomplete = map[string]interface{}{"reason": "content_filter"}
	}

	output := make([]map[string]interface{}, 0, len(s.items))
	for _, it := range s.items {
		output = append(output, it.item("completed"))
	}
	ru := responsesUsage{InputTokens: s.inputTokens, OutputTokens: s.outputTokens}
	ru.TotalTokens = s.inputTokens + s.outputTokens
	s.writer.emit("response.completed", map[string]interface{}{
		"response": responsesObject(s.responseID, s.model, status, s.createdAt, output, ru.toMap(), incomplete),
	})
	ru.toUsage(usage)
	s.completed = true
}

func (s *anthropicToResponsesState) fail(message string) {
	if s.completed {
		return
	}
	s.ensureStarted()
	s.writer.emit("response.failed", map[string]interface{}{
		"response": responsesObject(s.responseID, s.model, "failed", s.createdAt, nil, nil, map[string]interface{}{"reason": message}),
	})
	s.completed = true
}

// parseDataURLValue 与既有 AnthropicProvider.parseDataURL 同义，独立为包级函数供新格复用。
func parseDataURLValue(url string) (mediaType, data string) {
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
	return mediaType, parts[1]
}

// generateRandomID 生成跨协议 id 后缀（时间 + 原子计数，避免并发下重复）。
var randomIDCounter uint64

func generateRandomID() string {
	id := atomic.AddUint64(&randomIDCounter, 1)
	return fmt.Sprintf("%x%04x", time.Now().UnixNano(), id&0xffff)
}
