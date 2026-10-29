package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
	providerPkg "ai-gateway/internal/provider"
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
	Capabilities   string  `json:"capabilities"`
}

type providerModelResponse struct {
	ID            uint    `json:"id"`
	ProviderID    uint    `json:"provider_id"`
	ModelID       string  `json:"model_id"`
	DisplayName   string  `json:"display_name"`
	OwnedBy       string  `json:"owned_by"`
	ContextWindow int     `json:"context_window"`
	MaxOutput     int     `json:"max_output"`
	InputPrice    float64 `json:"input_price"`
	OutputPrice   float64 `json:"output_price"`
	Capabilities  string  `json:"capabilities"`
	Availability  string  `json:"availability"`
	CreatedAt     string  `json:"created_at"`
}

func NewProviderModelHandler() *ProviderModelHandler {
	return &ProviderModelHandler{}
}

func toProviderModelResponse(m model.ProviderModel) providerModelResponse {
	return providerModelResponse{
		ID:            m.ID,
		ProviderID:    m.ProviderID,
		ModelID:       m.ModelID,
		DisplayName:   m.DisplayName,
		OwnedBy:       m.OwnedBy,
		ContextWindow: m.ContextWindow,
		MaxOutput:     m.MaxOutput,
		InputPrice:    m.InputPrice,
		OutputPrice:   m.OutputPrice,
		Capabilities:  m.Capabilities,
		Availability:  m.Availability,
		CreatedAt:     m.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *ProviderModelHandler) List(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	var models []model.ProviderModel
	if err := model.DB.Where("provider_id = ?", providerID).Find(&models).Error; err != nil {
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

	ctxWindow := req.ContextWindow
	pm := model.ProviderModel{
		ProviderID:    uint(providerID),
		ModelID:       req.ModelID,
		DisplayName:   req.DisplayName,
		OwnedBy:       req.OwnedBy,
		ContextWindow: ctxWindow,
		MaxOutput:     req.MaxOutput,
		InputPrice:    req.InputPrice,
		OutputPrice:   req.OutputPrice,
		Capabilities:  req.Capabilities,
		Availability:  model.AvailabilityUnknown,
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

	ctxWindow := req.ContextWindow
	updates := map[string]interface{}{
		"display_name":    req.DisplayName,
		"context_window":  ctxWindow,
		"max_output":      req.MaxOutput,
		"input_price":     req.InputPrice,
		"output_price":    req.OutputPrice,
		"capabilities":    req.Capabilities,
	}

	if req.ModelID != "" && req.ModelID != pm.ModelID {
		var existing model.ProviderModel
		if err := model.DB.Where("provider_id = ? AND model_id = ? AND id != ?", providerID, req.ModelID, pm.ID).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model_id already exists"})
			return
		}
		updates["model_id"] = req.ModelID
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

	// 硬删除关联的 ModelMapping
	if err := model.DB.Where("provider_model_id = ?", pm.ID).Delete(&model.ModelMapping{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 硬删除 ProviderModel
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
	deactivated := 0

	// 空列表保护：接口异常返回空列表时不推进任何状态，防误灭全部模型
	if len(models) > 0 {
		syncedSet := make(map[string]bool, len(models))
		for _, pm := range models {
			syncedSet[pm.ModelID] = true

			var existing model.ProviderModel
			res := model.DB.Where("provider_id = ? AND model_id = ?", provider.ID, pm.ModelID).First(&existing)

			if res.Error != nil {
				pm.Availability = model.AvailabilityAvailable
				model.DB.Create(&pm)
				added++
			} else {
				// 逐字段空值合并：同步值为空/零值的字段不进 updates map，保留本地现值
				updates := map[string]interface{}{
					"availability": model.AvailabilityAvailable,
				}
				if pm.DisplayName != "" {
					updates["display_name"] = pm.DisplayName
				}
				if pm.OwnedBy != "" {
					updates["owned_by"] = pm.OwnedBy
				}
				if pm.ContextWindow > 0 {
					updates["context_window"] = pm.ContextWindow
				}
				if pm.MaxOutput > 0 {
					updates["max_output"] = pm.MaxOutput
				}
				if pm.Capabilities != "" {
					updates["capabilities"] = pm.Capabilities
				}
				model.DB.Model(&existing).Updates(updates)
				updated++
			}
		}

		// 失效识别：此前 available 但不在本次列表中的模型置 unavailable（unknown 不参与失效）
		var availableModels []model.ProviderModel
		if err := model.DB.
			Where("provider_id = ? AND availability = ?", provider.ID, model.AvailabilityAvailable).
			Find(&availableModels).Error; err == nil {
			var missingIDs []uint
			for _, am := range availableModels {
				if !syncedSet[am.ModelID] {
					missingIDs = append(missingIDs, am.ID)
				}
			}
			if len(missingIDs) > 0 {
				result := model.DB.Model(&model.ProviderModel{}).
					Where("id IN ?", missingIDs).
					Update("availability", model.AvailabilityUnavailable)
				deactivated = int(result.RowsAffected)
			}
		}
	}

	now := time.Now()
	model.DB.Model(&provider).Update("last_sync_at", &now)

	c.JSON(http.StatusOK, gin.H{
		"message":     fmt.Sprintf("%s models synced", provider.Name),
		"added":       added,
		"updated":     updated,
		"total":       len(models),
		"deactivated": deactivated,
	})
}
