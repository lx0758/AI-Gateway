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
	var openAIResult, anthropicResult *protocolTestResult

	if p.OpenAIBaseURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := executeTest(&p, &pm, "openai")
			openAIResult = &result
		}()
	}

	if p.AnthropicBaseURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := executeTest(&p, &pm, "anthropic")
			anthropicResult = &result
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
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	OpenAIBaseURL    string `json:"openai_base_url"`
	AnthropicBaseURL string `json:"anthropic_base_url"`
	Enabled          bool   `json:"enabled"`
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
	model.DB.Preload("Provider").
		Where("model_id = ? AND enabled = ?", m.ID, true).
		Order("weight DESC").
		Find(&mappings)

	type testJob struct {
		mapping model.ModelMapping
		pm      model.ProviderModel
	}

	var jobs []testJob
	for _, mapping := range mappings {
		if mapping.Provider == nil || !mapping.Provider.Enabled {
			continue
		}

		var pm model.ProviderModel
		if err := model.DB.Where("provider_id = ? AND model_id = ? AND is_available = ?", mapping.ProviderID, mapping.ProviderModelName, true).First(&pm).Error; err != nil {
			continue
		}

		jobs = append(jobs, testJob{mapping: mapping, pm: pm})
	}

	results := make([]mappingTestResult, len(jobs))
	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Add(1)
		go func(idx int, j testJob) {
			defer wg.Done()

			protocolTests := []protocolTestResult{}

			if j.mapping.Provider.OpenAIBaseURL != "" {
				result := executeTest(j.mapping.Provider, &j.pm, "openai")
				protocolTests = append(protocolTests, result)
			}

			if j.mapping.Provider.AnthropicBaseURL != "" {
				result := executeTest(j.mapping.Provider, &j.pm, "anthropic")
				protocolTests = append(protocolTests, result)
			}

			results[idx] = mappingTestResult{
				MappingID: j.mapping.ID,
				Weight:    j.mapping.Weight,
				Provider: providerBasicInfo{
					ID:               j.mapping.Provider.ID,
					Name:             j.mapping.Provider.Name,
					OpenAIBaseURL:    j.mapping.Provider.OpenAIBaseURL,
					AnthropicBaseURL: j.mapping.Provider.AnthropicBaseURL,
					Enabled:          j.mapping.Provider.Enabled,
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
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	testCtx, _ := gin.CreateTestContext(w)
	testCtx.Request = httptest.NewRequest("POST", "/", bytes.NewReader(bodyBytes))
	testCtx.Request.Header.Set("Content-Type", "application/json")

	providerImpl := provider.NewAutomatedProvider(
		p.OpenAIBaseURL,
		p.AnthropicBaseURL,
		p.APIKey,
	)
	usage := &provider.Usage{}

	ctx, cancel := context.WithTimeout(testCtx.Request.Context(), 30*time.Second)
	defer cancel()
	testCtx.Request = testCtx.Request.WithContext(ctx)

	start := time.Now()
	var err error
	callMethod := "direct"

	if protocol == "openai" {
		if p.OpenAIBaseURL == "" {
			callMethod = "convert"
		}
		err = providerImpl.ExecuteOpenAIRequest(testCtx, pm, usage)
	} else {
		if p.AnthropicBaseURL == "" {
			callMethod = "convert"
		}
		err = providerImpl.ExecuteAnthropicRequest(testCtx, pm, usage)
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

	if protocol == "openai" {
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
	} else {
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
	}

	return ""
}
