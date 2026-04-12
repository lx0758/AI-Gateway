package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
	"ai-gateway/internal/provider"
	"ai-gateway/internal/router"
	"ai-gateway/internal/utils"
)

type AnthropicProxyHandler struct {
	router *router.ModelRouter
}

func NewAnthropicProxyHandler() *AnthropicProxyHandler {
	return &AnthropicProxyHandler{
		router: router.GetRouter(),
	}
}

func (h *AnthropicProxyHandler) Messages(c *gin.Context) {
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

	result, err := h.router.Route(req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found or no available provider"})
		return
	}

	start := time.Now()
	usage := provider.Usage{}
	err = result.ProviderInstance.ExecuteAnthropicRequest(c, result.ProviderModel, &usage)
	latencyMs := time.Since(start).Milliseconds()

	status := "success"
	errorMsg := ""
	if err != nil {
		status = "error"
		errorMsg = err.Error()
		if provider.IsRateLimitError(err) {
			h.router.RecordRateLimit(result.Provider.ID, result.ProviderModel.ID)
		}
	} else {
		h.router.RecordSuccess(result.Provider.ID, result.ProviderModel.ID)
	}

	clientIPs := utils.GetClientIPInfo(c)

	modelLog := NewModelLog(
		"anthropic",
		clientIPs,
		keyID.(uint),
		keyName.(string),
		req.Model,
		result,
		result.SupportAnthropic(),
		&usage,
		int(latencyMs),
		status,
		errorMsg,
	)
	model.DB.Create(&modelLog)

	log.Println(modelLog.String())
}

func (h *AnthropicProxyHandler) ListModels(c *gin.Context) {
	var models []model.Model
	model.DB.Find(&models)

	var result []map[string]interface{}

	for _, m := range models {
		result = append(result, map[string]interface{}{
			"id":           m.Name,
			"type":         "model",
			"display_name": m.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     result,
		"has_more": false,
	})
}

func (h *AnthropicProxyHandler) GetModel(c *gin.Context) {
	modelID := c.Param("id")

	var m model.Model
	if err := model.DB.Where("name = ?", modelID).First(&m).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"type": "not_found_error", "message": "model not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           m.Name,
		"type":         "model",
		"display_name": m.Name,
	})
}
