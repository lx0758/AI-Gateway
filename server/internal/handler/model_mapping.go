package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type ModelMappingHandler struct{}

type CreateModelMappingRequest struct {
	Alias             string `json:"alias" binding:"required"`
	ProviderID        uint   `json:"provider_id" binding:"required"`
	ProviderModelName string `json:"provider_model_name" binding:"required"`
	Weight            int    `json:"weight"`
}

type UpdateModelMappingRequest struct {
	Alias             *string `json:"alias"`
	ProviderID        *uint   `json:"provider_id"`
	ProviderModelName *string `json:"provider_model_name"`
	Weight            *int    `json:"weight"`
	Enabled           *bool   `json:"enabled"`
}

func NewModelMappingHandler() *ModelMappingHandler {
	return &ModelMappingHandler{}
}

func (h *ModelMappingHandler) List(c *gin.Context) {
	var mappings []model.ModelMapping
	query := model.DB.Preload("Provider")

	if alias := c.Query("alias"); alias != "" {
		query = query.Where("alias = ?", alias)
	}

	if err := query.Find(&mappings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type MappingResponse struct {
		ID                uint   `json:"id"`
		Alias             string `json:"alias"`
		ProviderID        uint   `json:"provider_id"`
		ProviderModelName string `json:"provider_model_name"`
		Enabled           bool   `json:"enabled"`
		Weight            int    `json:"weight"`
		Provider          any    `json:"provider,omitempty"`
	}

	var response []MappingResponse
	for _, m := range mappings {
		item := MappingResponse{
			ID:                m.ID,
			Alias:             m.Alias,
			ProviderID:        m.ProviderID,
			ProviderModelName: m.ProviderModelName,
			Enabled:           m.Enabled,
			Weight:            m.Weight,
		}
		if m.Provider != nil {
			item.Provider = m.Provider
		}
		response = append(response, item)
	}

	c.JSON(http.StatusOK, gin.H{"mappings": response})
}

func (h *ModelMappingHandler) Create(c *gin.Context) {
	var req CreateModelMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var provider model.Provider
	if err := model.DB.First(&provider, req.ProviderID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider not found"})
		return
	}

	var pm model.ProviderModel
	if err := model.DB.Where("provider_id = ? AND model_id = ?", req.ProviderID, req.ProviderModelName).First(&pm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider model not found"})
		return
	}

	mapping := model.ModelMapping{
		Alias:             req.Alias,
		ProviderID:        req.ProviderID,
		ProviderModelName: req.ProviderModelName,
		Enabled:           true,
		Weight:            req.Weight,
	}

	if mapping.Weight == 0 {
		mapping.Weight = 1
	}

	if err := model.DB.Create(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.Preload("Provider").First(&mapping, mapping.ID)
	c.JSON(http.StatusCreated, gin.H{"mapping": mapping})
}

func (h *ModelMappingHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var mapping model.ModelMapping
	if err := model.DB.First(&mapping, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mapping not found"})
		return
	}

	var req UpdateModelMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Alias != nil {
		updates["alias"] = *req.Alias
	}
	if req.ProviderID != nil {
		var provider model.Provider
		if err := model.DB.First(&provider, *req.ProviderID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider not found"})
			return
		}
		updates["provider_id"] = *req.ProviderID
	}
	if req.ProviderModelName != nil {
		providerID := mapping.ProviderID
		if req.ProviderID != nil {
			providerID = *req.ProviderID
		}
		var pm model.ProviderModel
		if err := model.DB.Where("provider_id = ? AND model_id = ?", providerID, *req.ProviderModelName).First(&pm).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider model not found"})
			return
		}
		updates["provider_model_name"] = *req.ProviderModelName
	}
	if req.Weight != nil {
		updates["weight"] = *req.Weight
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&mapping).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	model.DB.Preload("Provider").First(&mapping, id)
	c.JSON(http.StatusOK, gin.H{"mapping": mapping})
}

func (h *ModelMappingHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := model.DB.Delete(&model.ModelMapping{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mapping deleted"})
}
