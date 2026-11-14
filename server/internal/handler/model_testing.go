package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
	"ai-gateway/internal/provider"
)

type ModelTestHandler struct{}

func NewModelTestHandler() *ModelTestHandler {
	return &ModelTestHandler{}
}

type protocolTestResult struct {
	Protocol     string `json:"protocol"`
	Success      bool   `json:"success"`
	CallMethod   string `json:"call_method"`
	LatencyMs    int64  `json:"latency_ms"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Response     string `json:"response"`
	Error        string `json:"error"`
}

type providerModelTestResponse struct {
	Provider struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"provider"`
	Model struct {
		ModelID     string `json:"model_id"`
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Tests []protocolTestResult `json:"tests"`
}

type customModelTestRequest struct {
	ModelID string `json:"model_id" binding:"required"`
}

type customModelTestResponse struct {
	Provider struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"provider"`
	Model struct {
		ModelID     string `json:"model_id"`
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Tests []protocolTestResult `json:"tests"`
}

func (h *ModelTestHandler) TestCustomModel(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	var req customModelTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model_id is required"})
		return
	}

	var p model.Provider
	if err := model.DB.First(&p, providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	pm := &model.ProviderModel{
		ModelID: req.ModelID,
	}

	var wg sync.WaitGroup
	var openAIResult, anthropicResult, responsesResult *protocolTestResult

	if p.OpenAIChatCompletionsBaseURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := executeTest(&p, pm, "openai")
			openAIResult = &result
		}()
	}

	if p.AnthropicMessagesBaseURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := executeTest(&p, pm, "anthropic")
			anthropicResult = &result
		}()
	}

	if p.OpenAIResponsesBaseURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := executeTest(&p, pm, "responses")
			responsesResult = &result
		}()
	}

	wg.Wait()

	tests := []protocolTestResult{}
	if openAIResult != nil {
		tests = append(tests, *openAIResult)
	}
	if anthropicResult != nil {
		tests = append(tests, *anthropicResult)
	}
	if responsesResult != nil {
		tests = append(tests, *responsesResult)
	}

	resp := customModelTestResponse{
		Provider: struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		}{ID: p.ID, Name: p.Name},
		Model: struct {
			ModelID     string `json:"model_id"`
			DisplayName string `json:"display_name"`
		}{ModelID: req.ModelID},
		Tests: tests,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ModelTestHandler) TestProviderModel(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	modelDBID, err := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var p model.Provider
	if err := model.DB.First(&p, providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	var pm model.ProviderModel
	if err := model.DB.Where("id = ? AND provider_id = ?", modelDBID, providerID).First(&pm).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider model not found"})
		return
	}

	var wg sync.WaitGroup
	var openAIResult, anthropicResult, responsesResult *protocolTestResult

	if p.OpenAIChatCompletionsBaseURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := executeTest(&p, &pm, "openai")
			openAIResult = &result
		}()
	}

	if p.AnthropicMessagesBaseURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := executeTest(&p, &pm, "anthropic")
			anthropicResult = &result
		}()
	}

	if p.OpenAIResponsesBaseURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := executeTest(&p, &pm, "responses")
			responsesResult = &result
		}()
	}

	wg.Wait()

	tests := []protocolTestResult{}
	if openAIResult != nil {
		tests = append(tests, *openAIResult)
	}
	if anthropicResult != nil {
		tests = append(tests, *anthropicResult)
	}
	if responsesResult != nil {
		tests = append(tests, *responsesResult)
	}

	resp := providerModelTestResponse{
		Provider: struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		}{ID: p.ID, Name: p.Name},
		Model: struct {
			ModelID     string `json:"model_id"`
			DisplayName string `json:"display_name"`
		}{ModelID: pm.ModelID, DisplayName: pm.DisplayName},
		Tests: tests,
	}

	c.JSON(http.StatusOK, resp)
}

type mappingTestResult struct {
	MappingID     uint                 `json:"mapping_id"`
	Weight        int                  `json:"weight"`
	Provider      providerBasicInfo    `json:"provider"`
	ProviderModel modelBasicInfo       `json:"provider_model"`
	ProtocolTests []protocolTestResult `json:"protocol_tests"`
}

type providerBasicInfo struct {
	ID                           uint   `json:"id"`
	Name                         string `json:"name"`
	OpenAIChatCompletionsBaseURL string `json:"openai_chat_completions_base_url"`
	AnthropicMessagesBaseURL     string `json:"anthropic_messages_base_url"`
	OpenAIResponsesBaseURL       string `json:"openai_responses_base_url"`
	Enabled                      bool   `json:"enabled"`
}

type modelBasicInfo struct {
	ModelID     string `json:"model_id"`
	DisplayName string `json:"display_name"`
}

type virtualModelTestResponse struct {
	ModelName string              `json:"model_name"`
	Tests     []mappingTestResult `json:"tests"`
}

func (h *ModelTestHandler) TestModel(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var m model.Model
	if err := model.DB.First(&m, modelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var mappings []model.ModelMapping
	model.DB.Preload("Provider").Preload("ProviderModel").
		Where("model_id = ? AND enabled = ?", m.ID, true).
		Order("weight DESC").
		Find(&mappings)

	type testJob struct {
		mapping model.ModelMapping
		pm      *model.ProviderModel
	}

	var jobs []testJob
	for _, mapping := range mappings {
		if mapping.Provider == nil || !mapping.Provider.Enabled {
			continue
		}

		if mapping.ProviderModel == nil || mapping.ProviderModel.Availability == model.AvailabilityUnavailable {
			continue
		}

		jobs = append(jobs, testJob{mapping: mapping, pm: mapping.ProviderModel})
	}

	results := make([]mappingTestResult, len(jobs))
	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Add(1)
		go func(idx int, j testJob) {
			defer wg.Done()

			protocolTests := []protocolTestResult{}

			if j.mapping.Provider.OpenAIChatCompletionsBaseURL != "" {
				result := executeTest(j.mapping.Provider, j.pm, "openai")
				protocolTests = append(protocolTests, result)
			}

			if j.mapping.Provider.AnthropicMessagesBaseURL != "" {
				result := executeTest(j.mapping.Provider, j.pm, "anthropic")
				protocolTests = append(protocolTests, result)
			}

			if j.mapping.Provider.OpenAIResponsesBaseURL != "" {
				result := executeTest(j.mapping.Provider, j.pm, "responses")
				protocolTests = append(protocolTests, result)
			}

			results[idx] = mappingTestResult{
				MappingID: j.mapping.ID,
				Weight:    j.mapping.Weight,
				Provider: providerBasicInfo{
					ID:                           j.mapping.Provider.ID,
					Name:                         j.mapping.Provider.Name,
					OpenAIChatCompletionsBaseURL: j.mapping.Provider.OpenAIChatCompletionsBaseURL,
					AnthropicMessagesBaseURL:     j.mapping.Provider.AnthropicMessagesBaseURL,
					OpenAIResponsesBaseURL:       j.mapping.Provider.OpenAIResponsesBaseURL,
					Enabled:                      j.mapping.Provider.Enabled,
				},
				ProviderModel: modelBasicInfo{
					ModelID:     j.pm.ModelID,
					DisplayName: j.pm.DisplayName,
				},
				ProtocolTests: protocolTests,
			}
		}(i, job)
	}

	wg.Wait()

	c.JSON(http.StatusOK, virtualModelTestResponse{
		ModelName: m.Name,
		Tests:     results,
	})
}

func executeTest(p *model.Provider, pm *model.ProviderModel, protocol string) protocolTestResult {
	body := map[string]interface{}{
		"model":      pm.ModelID,
		"messages":   []map[string]string{{"role": "user", "content": "简短介绍一下自己。"}},
		"max_tokens": 100,
		"stream":     false,
	}
	if protocol == "responses" {
		body = map[string]interface{}{
			"model":             pm.ModelID,
			"input":             "简短介绍一下自己。",
			"max_output_tokens": 100,
			"stream":            false,
		}
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	testCtx, _ := gin.CreateTestContext(w)
	testCtx.Request = httptest.NewRequest("POST", "/", bytes.NewReader(bodyBytes))
	testCtx.Request.Header.Set("Content-Type", "application/json")

	providerImpl := provider.NewAutomatedProvider(
		p.OpenAIChatCompletionsBaseURL,
		p.AnthropicMessagesBaseURL,
		p.OpenAIResponsesBaseURL,
		p.APIKey,
	)
	usage := &provider.Usage{}

	ctx, cancel := context.WithTimeout(testCtx.Request.Context(), 30*time.Second)
	defer cancel()
	testCtx.Request = testCtx.Request.WithContext(ctx)

	start := time.Now()
	var err error
	callMethod := "direct"

	switch protocol {
	case "openai":
		if p.OpenAIChatCompletionsBaseURL == "" {
			callMethod = "convert"
		}
		err = providerImpl.ExecuteOpenAIChatCompletionsRequest(testCtx, pm, usage)
	case "anthropic":
		if p.AnthropicMessagesBaseURL == "" {
			callMethod = "convert"
		}
		err = providerImpl.ExecuteAnthropicMessagesRequest(testCtx, pm, usage)
	case "responses":
		if p.OpenAIResponsesBaseURL == "" {
			callMethod = "convert"
		}
		err = providerImpl.ExecuteOpenAIResponsesRequest(testCtx, pm, usage)
	}

	latencyMs := time.Since(start).Milliseconds()

	respBody, _ := io.ReadAll(w.Body)
	response := extractResponseContent(respBody, protocol)

	errorMsg := ""
	success := err == nil && w.Code == 200

	if err != nil {
		errorMsg = err.Error()
	} else if w.Code != 200 {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
			ErrorMsg string `json:"error"`
		}
		json.Unmarshal(respBody, &errResp)
		if errResp.Error.Message != "" {
			errorMsg = errResp.Error.Message
		} else if errResp.ErrorMsg != "" {
			errorMsg = errResp.ErrorMsg
		} else {
			errorMsg = "HTTP " + strconv.Itoa(w.Code)
		}
	}

	return protocolTestResult{
		Protocol:     protocol,
		Success:      success,
		CallMethod:   callMethod,
		LatencyMs:    latencyMs,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		Response:     response,
		Error:        errorMsg,
	}
}

func extractResponseContent(body []byte, protocol string) string {
	if len(body) == 0 {
		return ""
	}

	switch protocol {
	case "openai":
		var resp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return ""
		}
		if len(resp.Choices) > 0 {
			return strings.TrimSpace(resp.Choices[0].Message.Content)
		}
	case "anthropic":
		var resp struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return ""
		}
		for _, c := range resp.Content {
			if c.Type == "text" && c.Text != "" {
				return strings.TrimSpace(c.Text)
			}
		}
	case "responses":
		var resp struct {
			Output []struct {
				Type    string `json:"type"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"output"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return ""
		}
		for _, item := range resp.Output {
			if item.Type != "message" {
				continue
			}
			for _, c := range item.Content {
				if c.Type == "output_text" && c.Text != "" {
					return strings.TrimSpace(c.Text)
				}
			}
		}
	}

	return ""
}
