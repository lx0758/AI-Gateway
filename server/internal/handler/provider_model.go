package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
	providerPkg "ai-proxy/internal/provider"
)

type ProviderModelHandler struct{}

type createProviderModelRequest struct {
	ModelID        string  `json:"model_id"`
	DisplayName    string  `json:"display_name"`
	OwnedBy        string  `json:"owned_by"`
	ContextWindow  int     `json:"context_window"`
	MaxOutput      int     `json:"max_output"`
	InputPrice     float64 `json:"input_price"`
	OutputPrice    float64 `json:"output_price"`
	SupportsVision bool    `json:"supports_vision"`
	SupportsTools  bool    `json:"supports_tools"`
	SupportsStream bool    `json:"supports_stream"`
}

type providerModelResponse struct {
	ID             uint    `json:"id"`
	ProviderID     uint    `json:"provider_id"`
	ModelID        string  `json:"model_id"`
	DisplayName    string  `json:"display_name"`
	OwnedBy        string  `json:"owned_by"`
	ContextWindow  int     `json:"context_window"`
	MaxOutput      int     `json:"max_output"`
	InputPrice     float64 `json:"input_price"`
	OutputPrice    float64 `json:"output_price"`
	SupportsVision bool    `json:"supports_vision"`
	SupportsTools  bool    `json:"supports_tools"`
	SupportsStream bool    `json:"supports_stream"`
	IsAvailable    bool    `json:"is_available"`
	Source         string  `json:"source"`
	CreatedAt      string  `json:"created_at"`
}

func NewProviderModelHandler() *ProviderModelHandler {
	return &ProviderModelHandler{}
}

func toProviderModelResponse(m model.ProviderModel) providerModelResponse {
	return providerModelResponse{
		ID:             m.ID,
		ProviderID:     m.ProviderID,
		ModelID:        m.ModelID,
		DisplayName:    m.DisplayName,
		OwnedBy:        m.OwnedBy,
		ContextWindow:  m.ContextWindow,
		MaxOutput:      m.MaxOutput,
		InputPrice:     m.InputPrice,
		OutputPrice:    m.OutputPrice,
		SupportsVision: m.SupportsVision,
		SupportsTools:  m.SupportsTools,
		SupportsStream: m.SupportsStream,
		IsAvailable:    m.IsAvailable,
		Source:         m.Source,
		CreatedAt:      m.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *ProviderModelHandler) List(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	var models []model.ProviderModel
	query := model.DB.Where("provider_id = ?", providerID)

	if c.Query("available_only") == "true" {
		query = query.Where("is_available = ?", true)
	}

	if err := query.Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]providerModelResponse, len(models))
	for i, m := range models {
		result[i] = toProviderModelResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"models": result})
}

func (h *ProviderModelHandler) Create(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	var req createProviderModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pm := model.ProviderModel{
		ProviderID:     uint(providerID),
		ModelID:        req.ModelID,
		DisplayName:    req.DisplayName,
		OwnedBy:        req.OwnedBy,
		ContextWindow:  req.ContextWindow,
		MaxOutput:      req.MaxOutput,
		InputPrice:     req.InputPrice,
		OutputPrice:    req.OutputPrice,
		SupportsVision: req.SupportsVision,
		SupportsTools:  req.SupportsTools,
		SupportsStream: req.SupportsStream,
		IsAvailable:    true,
		Source:         "manual",
	}

	if err := model.DB.Create(&pm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"model": toProviderModelResponse(pm)})
}

func (h *ProviderModelHandler) Update(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	modelID, err := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var pm model.ProviderModel
	if err := model.DB.Where("id = ? AND provider_id = ?", modelID, providerID).First(&pm).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var req createProviderModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"display_name":    req.DisplayName,
		"context_window":  req.ContextWindow,
		"max_output":      req.MaxOutput,
		"input_price":     req.InputPrice,
		"output_price":    req.OutputPrice,
		"supports_vision": req.SupportsVision,
		"supports_tools":  req.SupportsTools,
		"supports_stream": req.SupportsStream,
	}

	if err := model.DB.Model(&pm).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.First(&pm, pm.ID)
	c.JSON(http.StatusOK, gin.H{"model": toProviderModelResponse(pm)})
}

func (h *ProviderModelHandler) Delete(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	modelID, err := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var pm model.ProviderModel
	if err := model.DB.Where("id = ? AND provider_id = ?", modelID, providerID).First(&pm).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	if err := model.DB.Delete(&pm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "model deleted"})
}

func (h *ProviderModelHandler) Sync(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	var provider model.Provider
	if err := model.DB.First(&provider, providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	providerImpl := providerPkg.NewAutomatedProvider(
		provider.OpenAIBaseURL,
		provider.AnthropicBaseURL,
		provider.APIKey,
	)
	models, err := providerImpl.SyncModels(provider.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	added := 0
	updated := 0

	for _, pm := range models {
		var existing model.ProviderModel
		res := model.DB.Where("provider_id = ? AND model_id = ?", provider.ID, pm.ModelID).First(&existing)

		if res.Error != nil {
			model.DB.Create(&pm)
			added++
		} else if existing.Source != "manual" {
			model.DB.Model(&existing).Updates(map[string]interface{}{
				"display_name":    pm.DisplayName,
				"owned_by":        pm.OwnedBy,
				"context_window":  pm.ContextWindow,
				"max_output":      pm.MaxOutput,
				"supports_vision": pm.SupportsVision,
				"supports_tools":  pm.SupportsTools,
				"is_available":    true,
			})
			updated++
		}
	}

	now := time.Now()
	model.DB.Model(&provider).Update("last_sync_at", &now)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s models synced", provider.Name),
		"added":   added,
		"updated": updated,
		"total":   len(models),
	})
}
