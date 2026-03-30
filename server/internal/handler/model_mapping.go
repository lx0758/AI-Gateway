package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type ModelMappingHandler struct{}

type CreateModelMappingRequest struct {
	Alias           string `json:"alias" binding:"required"`
	ProviderID      uint   `json:"provider_id" binding:"required"`
	ProviderModelID uint   `json:"provider_model_id" binding:"required"`
	Weight          int    `json:"weight"`
}

type UpdateModelMappingRequest struct {
	Alias           *string `json:"alias"`
	ProviderID      *uint   `json:"provider_id"`
	ProviderModelID *uint   `json:"provider_model_id"`
	Weight          *int    `json:"weight"`
	Enabled         *bool   `json:"enabled"`
}

func NewModelMappingHandler() *ModelMappingHandler {
	return &ModelMappingHandler{}
}

func (h *ModelMappingHandler) List(c *gin.Context) {
	var mappings []model.ModelMapping
	query := model.DB.Preload("Provider").Preload("ProviderModel")

	if alias := c.Query("alias"); alias != "" {
		query = query.Where("alias = ?", alias)
	}

	if err := query.Find(&mappings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mappings": mappings})
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
	if err := model.DB.Where("id = ? AND provider_id = ?", req.ProviderModelID, req.ProviderID).First(&pm).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider model not found"})
		return
	}

	mapping := model.ModelMapping{
		Alias:           req.Alias,
		ProviderID:      req.ProviderID,
		ProviderModelID: req.ProviderModelID,
		Enabled:         true,
		Weight:          req.Weight,
	}

	if mapping.Weight == 0 {
		mapping.Weight = 1
	}

	if err := model.DB.Create(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.Preload("Provider").Preload("ProviderModel").First(&mapping, mapping.ID)
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
	if req.ProviderModelID != nil {
		providerID := mapping.ProviderID
		if req.ProviderID != nil {
			providerID = *req.ProviderID
		}
		var pm model.ProviderModel
		if err := model.DB.Where("id = ? AND provider_id = ?", *req.ProviderModelID, providerID).First(&pm).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider model not found"})
			return
		}
		updates["provider_model_id"] = *req.ProviderModelID
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

	model.DB.Preload("Provider").Preload("ProviderModel").First(&mapping, id)
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
