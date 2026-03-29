package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/user/ai-model-proxy/internal/model"
)

type ProviderModelHandler struct{}

type CreateProviderModelRequest struct {
	ModelID        string  `json:"model_id" binding:"required"`
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

func NewProviderModelHandler() *ProviderModelHandler {
	return &ProviderModelHandler{}
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

	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (h *ProviderModelHandler) Create(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	var req CreateProviderModelRequest
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

	c.JSON(http.StatusCreated, gin.H{"model": pm})
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

	var req CreateProviderModelRequest
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
	c.JSON(http.StatusOK, gin.H{"model": pm})
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

	if err := model.DB.Where("id = ? AND provider_id = ?", modelID, providerID).Delete(&model.ProviderModel{}).Error; err != nil {
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

	switch provider.APIType {
	case "openai":
		h.syncOpenAIModels(c, &provider)
	case "anthropic":
		h.syncAnthropicModels(c, &provider)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider type"})
	}
}

func (h *ProviderModelHandler) syncOpenAIModels(c *gin.Context, provider *model.Provider) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request: " + err.Error()})
		return
	}
	req.Header.Set("Authorization", "Bearer "+provider.GetDecryptedAPIKey())

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch models: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("OpenAI API error: %s", string(body))})
		return
	}

	var result struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response: " + err.Error()})
		return
	}

	added := 0
	updated := 0

	for _, m := range result.Data {
		var pm model.ProviderModel
		res := model.DB.Where("provider_id = ? AND model_id = ?", provider.ID, m.ID).First(&pm)

		if res.Error != nil {
			pm = model.ProviderModel{
				ProviderID:     provider.ID,
				ModelID:        m.ID,
				DisplayName:    m.ID,
				OwnedBy:        m.OwnedBy,
				SupportsStream: true,
				IsAvailable:    true,
				Source:         "sync",
			}
			model.DB.Create(&pm)
			added++
		} else {
			model.DB.Model(&pm).Updates(map[string]interface{}{
				"owned_by":     m.OwnedBy,
				"is_available": true,
			})
			updated++
		}
	}

	now := time.Now()
	model.DB.Model(provider).Update("last_sync_at", &now)

	c.JSON(http.StatusOK, gin.H{
		"message": "OpenAI models synced",
		"added":   added,
		"updated": updated,
		"total":   len(result.Data),
	})
}

func (h *ProviderModelHandler) syncAnthropicModels(c *gin.Context, provider *model.Provider) {
	baseURL := provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request: " + err.Error()})
		return
	}
	req.Header.Set("x-api-key", provider.GetDecryptedAPIKey())
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch models: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Anthropic API error: %s", string(body))})
		return
	}

	var result struct {
		Data []struct {
			ID            string `json:"id"`
			Type          string `json:"type"`
			DisplayName   string `json:"display_name"`
			CreatedAt     string `json:"created_at"`
			MaxInputToken int    `json:"max_input_tokens"`
			MaxTokens     int    `json:"max_tokens"`
			Capabilities  struct {
				ImageInput struct {
					Supported bool `json:"supported"`
				} `json:"image_input"`
				Thinking struct {
					Supported bool `json:"supported"`
				} `json:"thinking"`
			} `json:"capabilities"`
		} `json:"data"`
		HasMore bool `json:"has_more"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response: " + err.Error()})
		return
	}

	added := 0
	updated := 0

	for _, m := range result.Data {
		var pm model.ProviderModel
		res := model.DB.Where("provider_id = ? AND model_id = ?", provider.ID, m.ID).First(&pm)

		displayName := m.DisplayName
		if displayName == "" {
			displayName = m.ID
		}

		supportsVision := m.Capabilities.ImageInput.Supported
		supportsTools := true

		if res.Error != nil {
			pm = model.ProviderModel{
				ProviderID:     provider.ID,
				ModelID:        m.ID,
				DisplayName:    displayName,
				OwnedBy:        "anthropic",
				ContextWindow:  m.MaxInputToken,
				MaxOutput:      m.MaxTokens,
				SupportsVision: supportsVision,
				SupportsTools:  supportsTools,
				SupportsStream: true,
				IsAvailable:    true,
				Source:         "sync",
			}
			model.DB.Create(&pm)
			added++
		} else {
			model.DB.Model(&pm).Updates(map[string]interface{}{
				"display_name":    displayName,
				"context_window":  m.MaxInputToken,
				"max_output":      m.MaxTokens,
				"supports_vision": supportsVision,
				"supports_tools":  supportsTools,
				"is_available":    true,
			})
			updated++
		}
	}

	now := time.Now()
	model.DB.Model(provider).Update("last_sync_at", &now)

	c.JSON(http.StatusOK, gin.H{
		"message": "Anthropic models synced",
		"added":   added,
		"updated": updated,
		"total":   len(result.Data),
	})
}
