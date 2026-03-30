package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
	"ai-proxy/internal/router"
	"ai-proxy/internal/transformer"
)

type ProxyHandler struct {
	router *router.ModelRouter
}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{
		router: router.NewModelRouter(),
	}
}

func (h *ProxyHandler) ChatCompletions(c *gin.Context) {
	var req transformer.OpenAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKeyID, _ := c.Get("api_key_id")
	if keyID, ok := apiKeyID.(uint); ok {
		var permissionCount int64
		model.DB.Model(&model.APIKeyModel{}).Where("api_key_id = ?", keyID).Count(&permissionCount)
		if permissionCount > 0 {
			var hasPermission int64
			model.DB.Model(&model.APIKeyModel{}).
				Where("api_key_id = ? AND model_alias = ?", keyID, req.Model).
				Count(&hasPermission)
			if hasPermission == 0 {
				c.JSON(http.StatusForbidden, gin.H{"error": "model not allowed for this API key"})
				return
			}
		}
	}

	result, err := h.router.Route(req.Model)
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found or no available provider"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Model = result.ActualModel

	trans := h.getTransformer(result.Provider.APIType)

	transformedReq, err := trans.TransformRequest(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reqBody, err := json.Marshal(transformedReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	endpoint := result.Provider.BaseURL
	if result.Provider.APIType == "anthropic" {
		endpoint = endpoint + "/messages"
	} else {
		endpoint = endpoint + "/chat/completions"
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if result.Provider.APIType == "anthropic" {
		httpReq.Header.Set("x-api-key", result.Provider.APIKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	} else {
		httpReq.Header.Set("Authorization", "Bearer "+result.Provider.APIKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	startTime := time.Now()

	resp, err := client.Do(httpReq)
	if err != nil {
		h.logUsage(c, result.Provider.ID, req.Model, 0, 0, int(time.Since(startTime).Milliseconds()), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if req.Stream {
		h.handleStreamResponse(c, resp, trans, result.Provider.ID, req.Model, startTime)
	} else {
		h.handleNormalResponse(c, resp, trans, result.Provider.ID, req.Model, startTime)
	}
}

func (h *ProxyHandler) handleNormalResponse(c *gin.Context, resp *http.Response, trans transformer.Transformer, providerID uint, modelName string, startTime time.Time) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logUsage(c, providerID, modelName, 0, 0, int(time.Since(startTime).Milliseconds()), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logUsage(c, providerID, modelName, 0, 0, int(time.Since(startTime).Milliseconds()), "error", string(body))
		c.JSON(resp.StatusCode, gin.H{"error": string(body)})
		return
	}

	openaiResp, err := trans.TransformResponse(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logUsage(c, providerID, modelName, openaiResp.Usage.PromptTokens, openaiResp.Usage.CompletionTokens, int(time.Since(startTime).Milliseconds()), "success", "")

	c.JSON(http.StatusOK, openaiResp)
}

func (h *ProxyHandler) handleStreamResponse(c *gin.Context, resp *http.Response, trans transformer.Transformer, providerID uint, modelName string, startTime time.Time) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	if err := trans.TransformStream(resp.Body, c.Writer); err != nil {
		h.logUsage(c, providerID, modelName, 0, 0, int(time.Since(startTime).Milliseconds()), "error", err.Error())
	}

	h.logUsage(c, providerID, modelName, 0, 0, int(time.Since(startTime).Milliseconds()), "success", "")
}

func (h *ProxyHandler) getTransformer(apiType string) transformer.Transformer {
	switch apiType {
	case "anthropic":
		return transformer.NewOpenAIToAnthropicTransformer()
	default:
		return transformer.NewPassThroughTransformer()
	}
}

func (h *ProxyHandler) logUsage(c *gin.Context, providerID uint, modelName string, promptTokens, completionTokens, latencyMs int, status, errorMsg string) {
	apiKeyID, _ := c.Get("api_key_id")

	usageLog := model.UsageLog{
		ProviderID:       providerID,
		Model:            modelName,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		LatencyMs:        latencyMs,
		Status:           status,
		ErrorMsg:         errorMsg,
	}

	if keyID, ok := apiKeyID.(uint); ok {
		usageLog.APIKeyID = keyID
	}

	model.DB.Create(&usageLog)

	if keyID, ok := apiKeyID.(uint); ok {
		model.DB.Model(&model.APIKey{}).Where("id = ?", keyID).
			UpdateColumn("used_quota", model.DB.Raw("used_quota + ?", promptTokens+completionTokens))
	}
}

func (h *ProxyHandler) ListModels(c *gin.Context) {
	var mappings []model.ModelMapping
	model.DB.Preload("ProviderModel").Find(&mappings)

	modelMap := make(map[string]bool)
	var models []map[string]interface{}

	for _, m := range mappings {
		if m.ProviderModel == nil {
			continue
		}
		if _, exists := modelMap[m.Alias]; !exists {
			modelMap[m.Alias] = true
			models = append(models, map[string]interface{}{
				"id":       m.Alias,
				"object":   "model",
				"created":  time.Now().Unix(),
				"owned_by": "ai-proxy",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}

func (h *ProxyHandler) GetModel(c *gin.Context) {
	modelID := c.Param("id")

	var mapping model.ModelMapping
	if err := model.DB.Preload("ProviderModel").Where("alias = ?", modelID).First(&mapping).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       mapping.Alias,
		"object":   "model",
		"created":  time.Now().Unix(),
		"owned_by": "ai-proxy",
	})
}
