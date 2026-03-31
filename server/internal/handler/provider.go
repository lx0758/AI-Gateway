package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type ProviderHandler struct{}

type createProviderRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=openai anthropic"`
	BaseURL  string `json:"base_url" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
	Priority int    `json:"priority"`
}

type updateProviderRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type" binding:"omitempty,oneof=openai anthropic"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key" binding:"omitempty"`
	Enabled  *bool  `json:"enabled"`
	Priority *int   `json:"priority"`
}

type providerResponse struct {
	ID           uint                    `json:"id"`
	Name         string                  `json:"name"`
	Type      string                     `json:"type"`
	BaseURL      string                  `json:"base_url"`
	APIKeyMasked string                  `json:"api_key_masked"`
	Enabled      bool                    `json:"enabled"`
	Priority     int                     `json:"priority"`
	Models       []providerModelResponse `json:"models,omitempty"`
	CreatedAt    string                  `json:"created_at"`
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

	result := make([]providerResponse, len(providers))
	for i, p := range providers {
		models := make([]providerModelResponse, len(p.Models))
		for j, m := range p.Models {
			models[j] = toProviderModelResponse(m)
		}

		result[i] = providerResponse{
			ID:           p.ID,
			Name:         p.Name,
			Type:      p.Type,
			BaseURL:      p.BaseURL,
			APIKeyMasked: maskAPIKey(p.APIKey),
			Enabled:      p.Enabled,
			Priority:     p.Priority,
			Models:       models,
			CreatedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"providers": result})
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

	models := make([]providerModelResponse, len(provider.Models))
	for j, m := range provider.Models {
		models[j] = toProviderModelResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"provider": providerResponse{
		ID:           provider.ID,
		Name:         provider.Name,
		Type:      provider.Type,
		BaseURL:      provider.BaseURL,
		APIKeyMasked: maskAPIKey(provider.APIKey),
		Enabled:      provider.Enabled,
		Priority:     provider.Priority,
		Models:       models,
		CreatedAt:    provider.CreatedAt.Format("2006-01-02 15:04:05"),
	}})
}

func (h *ProviderHandler) Create(c *gin.Context) {
	var req createProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider := model.Provider{
		Name:     req.Name,
		Type:     req.Type,
		BaseURL:  strings.TrimSuffix(req.BaseURL, "/"),
		APIKey:   req.APIKey,
		Enabled:  true,
		Priority: req.Priority,
	}

	if err := model.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"provider": providerResponse{
		ID:           provider.ID,
		Name:         provider.Name,
		Type:      provider.Type,
		BaseURL:      provider.BaseURL,
		APIKeyMasked: maskAPIKey(provider.APIKey),
		Enabled:      provider.Enabled,
		Priority:     provider.Priority,
		CreatedAt:    provider.CreatedAt.Format("2006-01-02 15:04:05"),
	}})
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

	var req updateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
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

	model.DB.Preload("Models").First(&provider, id)

	models := make([]providerModelResponse, len(provider.Models))
	for j, m := range provider.Models {
		models[j] = toProviderModelResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"provider": providerResponse{
		ID:           provider.ID,
		Name:         provider.Name,
		Type:      provider.Type,
		BaseURL:      provider.BaseURL,
		APIKeyMasked: maskAPIKey(provider.APIKey),
		Enabled:      provider.Enabled,
		Priority:     provider.Priority,
		Models:       models,
		CreatedAt:    provider.CreatedAt.Format("2006-01-02 15:04:05"),
	}})
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

func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 4 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
