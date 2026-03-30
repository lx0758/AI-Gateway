package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/manufacturer"
	"ai-proxy/internal/model"
	"ai-proxy/internal/router"
)

type AnthropicProxyHandler struct {
	factory *manufacturer.Factory
	router  *router.ModelRouter
}

func NewAnthropicProxyHandler() *AnthropicProxyHandler {
	return &AnthropicProxyHandler{
		factory: manufacturer.NewFactory(),
		router:  router.NewModelRouter(),
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

	mfr := h.factory.Create(result.Provider)
	if mfr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "manufacturer not found"})
		return
	}

	tokens, err := mfr.ExecuteAnthropicRequest(c, result.ProviderModel);
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: add calling count and tokens usage to database
	log.Printf("calling Anthropic api use tokens:%d", tokens)
}
