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

type ProxyHandler struct {
	factory *manufacturer.Factory
	router  *router.ModelRouter
}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{
		factory: manufacturer.NewFactory(),
		router:  router.NewModelRouter(),
	}
}

func (h *ProxyHandler) ChatCompletions(c *gin.Context) {
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

	tokens, err := mfr.ExecuteOpenAIRequest(c, result.ProviderModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: add calling count and tokens usage to database
	log.Printf("calling OpenAI api use tokens:%d", tokens)
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
		"owned_by": "ai-proxy",
	})
}
