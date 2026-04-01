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

type AnthropicProxyHandler struct {
	router *router.ModelRouter
}

func NewAnthropicProxyHandler() *AnthropicProxyHandler {
	return &AnthropicProxyHandler{
		router: router.NewModelRouter(),
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
	tokens, err := result.ProviderInstance.ExecuteAnthropicRequest(c, result.ProviderModel)
	latencyMs := time.Since(start).Milliseconds()

	status := "success"
	errorMsg := ""
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	usageLog := NewUsageLog(
		"anthropic",
		keyID.(uint),
		keyName.(string),
		req.Model,
		&result,
		result.SupportAnthropic(),
		tokens,
		int(latencyMs),
		status,
		errorMsg,
	)
	model.DB.Create(&usageLog)

	log.Println(usageLog.String())
}

func (h *AnthropicProxyHandler) ListModels(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "api not implemented"})
}

func (h *AnthropicProxyHandler) GetModel(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "api not implemented"})
}
