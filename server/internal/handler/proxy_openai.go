package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
	"ai-proxy/internal/router"
)

type OpenAIProxyHandler struct {
	router *router.ModelRouter
}

func NewOpenAIProxyHandler() *OpenAIProxyHandler {
	return &OpenAIProxyHandler{
		router: router.NewModelRouter(),
	}
}

func (h *OpenAIProxyHandler) ChatCompletions(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keyID, _ := c.Get("key_id")
	keyName, _ := c.Get("key_name")
	if err := VerifyKeyID(keyID, req.Model); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	results, err := h.router.Route(req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found or no available provider"})
		return
	}

	result := results[0]

	start := time.Now()
	tokens, err := result.ProviderInstance.ExecuteOpenAIRequest(c, result.ProviderModel)
	latencyMs := time.Since(start).Milliseconds()

	status := "success"
	errorMsg := ""
	if err != nil {
		status = "error"
		errorMsg = err.Error()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	usageLog := NewUsageLog(
		"openai",
		keyID.(uint),
		keyName.(string),
		req.Model,
		&result,
		result.SupportOpenAI(),
		tokens,
		int(latencyMs),
		status,
		errorMsg,
	)
	model.DB.Create(&usageLog)
	log.Println(usageLog.String())
}

func (h *OpenAIProxyHandler) ListModels(c *gin.Context) {
	var aliases []model.Alias
	model.DB.Find(&aliases)

	var models []map[string]interface{}

	for _, a := range aliases {
		models = append(models, map[string]interface{}{
			"id":       a.Name,
			"object":   "model",
			"owned_by": "ai-proxy",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}

func (h *OpenAIProxyHandler) GetModel(c *gin.Context) {
	modelID := c.Param("id")

	var alias model.Alias
	if err := model.DB.Where("name = ?", modelID).First(&alias).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       alias.Name,
		"object":   "model",
		"owned_by": "ai-proxy",
	})
}
