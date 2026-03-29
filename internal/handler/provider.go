package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/user/ai-model-proxy/internal/model"
)

type ProviderHandler struct{}

type CreateProviderRequest struct {
	Name     string `json:"name" binding:"required"`
	APIType  string `json:"api_type" binding:"required,oneof=openai anthropic"`
	BaseURL  string `json:"base_url" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
	Priority int    `json:"priority"`
}

type UpdateProviderRequest struct {
	Name     string `json:"name"`
	APIType  string `json:"api_type" binding:"omitempty,oneof=openai anthropic"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Enabled  *bool  `json:"enabled"`
	Priority *int   `json:"priority"`
}

func NewProviderHandler() *ProviderHandler {
	return &ProviderHandler{}
}

func (h *ProviderHandler) List(c *gin.Context) {
	var providers []model.Provider
	if err := model.DB.Preload("Models").Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range providers {
		providers[i].APIKeyMasked = maskAPIKeyFromEncrypted(providers[i].APIKey)
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (h *ProviderHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var provider model.Provider
	if err := model.DB.Preload("Models").First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	provider.APIKeyMasked = maskAPIKeyFromEncrypted(provider.APIKey)
	c.JSON(http.StatusOK, gin.H{"provider": provider})
}

func (h *ProviderHandler) Create(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider := model.Provider{
		Name:     req.Name,
		APIType:  req.APIType,
		BaseURL:  strings.TrimSuffix(req.BaseURL, "/"),
		APIKey:   req.APIKey,
		Enabled:  true,
		Priority: req.Priority,
	}

	if err := model.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	provider.APIKeyMasked = maskAPIKeyFromEncrypted(provider.APIKey)
	c.JSON(http.StatusCreated, gin.H{"provider": provider})
}

func (h *ProviderHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var provider model.Provider
	if err := model.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.APIType != "" {
		updates["api_type"] = req.APIType
	}
	if req.BaseURL != "" {
		updates["base_url"] = strings.TrimSuffix(req.BaseURL, "/")
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	if err := model.DB.Model(&provider).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.First(&provider, id)
	provider.APIKeyMasked = maskAPIKeyFromEncrypted(provider.APIKey)
	c.JSON(http.StatusOK, gin.H{"provider": provider})
}

func (h *ProviderHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var mappingCount int64
	model.DB.Model(&model.ModelMapping{}).Where("provider_id = ?", id).Count(&mappingCount)
	if mappingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider has existing model mappings, remove them first"})
		return
	}

	if err := model.DB.Delete(&model.Provider{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "provider deleted"})
}

func (h *ProviderHandler) Test(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var provider model.Provider
	if err := model.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "connection test not implemented yet"})
}

func maskAPIKeyFromEncrypted(encryptedKey string) string {
	if encryptedKey == "" {
		return ""
	}
	decrypted := decryptAPIKey(encryptedKey)
	if len(decrypted) <= 4 {
		return "****"
	}
	return decrypted[:4] + "****" + decrypted[len(decrypted)-4:]
}

func decryptAPIKey(encryptedKey string) string {
	var p model.Provider
	p.APIKey = encryptedKey
	return p.GetDecryptedAPIKey()
}

func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
