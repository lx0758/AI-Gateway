package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
)

type ModelHandler struct{}

type createModelRequest struct {
	Name    string `json:"name" binding:"required"`
	Enabled bool   `json:"enabled"`
}

type updateModelRequest struct {
	Name    *string `json:"name"`
	Enabled *bool   `json:"enabled"`
}

type createMappingRequest struct {
	ProviderID        uint   `json:"provider_id" binding:"required"`
	ProviderModelName string `json:"provider_model_name" binding:"required"`
	Weight            int    `json:"weight"`
	Enabled           bool   `json:"enabled"`
}

type updateMappingRequest struct {
	ProviderID        *uint   `json:"provider_id"`
	ProviderModelName *string `json:"provider_model_name"`
	Weight            *int    `json:"weight"`
	Enabled           *bool   `json:"enabled"`
}

type updateMappingsOrderRequest struct {
	Order []uint `json:"order" binding:"required"`
}

type modelResponse struct {
	ID               uint              `json:"id"`
	Model            string            `json:"model"`
	Enabled          bool              `json:"enabled"`
	MappingCount     int               `json:"mapping_count"`
	MinContextWindow int               `json:"min_context_window"`
	MinMaxOutput     int               `json:"min_max_output"`
	SupportsVision   bool              `json:"supports_vision"`
	SupportsTools    bool              `json:"supports_tools"`
	SupportsStream   bool              `json:"supports_stream"`
	CreatedAt        string            `json:"created_at"`
	Mappings         []mappingResponse `json:"mappings,omitempty"`
}

type mappingResponse struct {
	ID                uint                   `json:"id"`
	ProviderID        uint                   `json:"provider_id"`
	ProviderModelName string                 `json:"provider_model_name"`
	Weight            int                    `json:"weight"`
	Enabled           bool                   `json:"enabled"`
	Provider          *providerBasicResponse `json:"provider,omitempty"`
	ModelInfo         *modelInfoResponse     `json:"model_info,omitempty"`
}

type modelInfoResponse struct {
	ContextWindow  int  `json:"context_window"`
	MaxOutput      int  `json:"max_output"`
	SupportsVision bool `json:"supports_vision"`
	SupportsTools  bool `json:"supports_tools"`
	SupportsStream bool `json:"supports_stream"`
}

type providerBasicResponse struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	OpenAIBaseURL    string `json:"openai_base_url"`
	AnthropicBaseURL string `json:"anthropic_base_url"`
}

func NewModelHandler() *ModelHandler {
	return &ModelHandler{}
}

func (h *ModelHandler) List(c *gin.Context) {
	var models []model.Model
	if err := model.DB.Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]modelResponse, len(models))
	for i, a := range models {
		var mappings []model.ModelMapping
		model.DB.Preload("Provider").
			Joins("JOIN providers ON providers.id = model_mappings.provider_id AND providers.enabled = ?", true).
			Where("model_id = ?", a.ID).
			Order("weight DESC").
			Find(&mappings)

		mappingResponses := make([]mappingResponse, len(mappings))
		for j, m := range mappings {
			mappingResponses[j] = toMappingResponse(m)
		}

		enabledCount := calculateEnabledCount(mappings)
		minContext, minOutput := calculateMinTokens(mappings)
		supportsVision, supportsTools, supportsStream := calculateCapabilitiesIntersection(mappings)

		result[i] = modelResponse{
			ID:               a.ID,
			Model:            a.Name,
			Enabled:          a.Enabled,
			MappingCount:     enabledCount,
			MinContextWindow: minContext,
			MinMaxOutput:     minOutput,
			SupportsVision:   supportsVision,
			SupportsTools:    supportsTools,
			SupportsStream:   supportsStream,
			CreatedAt:        a.CreatedAt.Format("2006-01-02 15:04:05"),
			Mappings:         mappingResponses,
		}
	}

	c.JSON(http.StatusOK, gin.H{"models": result})
}

func (h *ModelHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m model.Model
	if err := model.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var mappings []model.ModelMapping
	model.DB.Preload("Provider").
		Joins("JOIN providers ON providers.id = model_mappings.provider_id AND providers.enabled = ?", true).
		Where("model_id = ?", m.ID).
		Order("weight DESC").
		Find(&mappings)

	mappingResponses := make([]mappingResponse, len(mappings))
	for j, mm := range mappings {
		mappingResponses[j] = toMappingResponse(mm)
	}

	c.JSON(http.StatusOK, gin.H{"model": modelResponse{
		ID:        m.ID,
		Model:     m.Name,
		Enabled:   m.Enabled,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		Mappings:  mappingResponses,
	}})
}

func (h *ModelHandler) Create(c *gin.Context) {
	var req createModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := model.Model{
		Name:    req.Name,
		Enabled: true,
	}
	if !req.Enabled {
		m.Enabled = false
	}

	if err := model.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"model": modelResponse{
		ID:           m.ID,
		Model:        m.Name,
		Enabled:      m.Enabled,
		MappingCount: 0,
		CreatedAt:    m.CreatedAt.Format("2006-01-02 15:04:05"),
	}})
}

func (h *ModelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m model.Model
	if err := model.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var req updateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&m).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	model.DB.First(&m, id)

	var mappings []model.ModelMapping
	model.DB.Preload("Provider").
		Joins("JOIN providers ON providers.id = model_mappings.provider_id AND providers.enabled = ?", true).
		Where("model_id = ?", m.ID).
		Order("weight DESC").
		Find(&mappings)

	mappingResponses := make([]mappingResponse, len(mappings))
	for j, mm := range mappings {
		mappingResponses[j] = toMappingResponse(mm)
	}

	c.JSON(http.StatusOK, gin.H{"model": modelResponse{
		ID:        m.ID,
		Model:     m.Name,
		Enabled:   m.Enabled,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		Mappings:  mappingResponses,
	}})
}

func (h *ModelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := model.DB.Delete(&model.Model{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "model deleted"})
}

func (h *ModelHandler) ListMappings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m model.Model
	if err := model.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var mappings []model.ModelMapping
	model.DB.Preload("Provider").
		Joins("JOIN providers ON providers.id = model_mappings.provider_id AND providers.enabled = ?", true).
		Where("model_id = ?", m.ID).
		Order("weight DESC").
		Find(&mappings)

	result := make([]mappingResponse, len(mappings))
	for i, mm := range mappings {
		result[i] = toMappingResponse(mm)
	}

	c.JSON(http.StatusOK, gin.H{"mappings": result})
}

func (h *ModelHandler) CreateMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m model.Model
	if err := model.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var req createMappingRequest
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
		ModelID:           m.ID,
		ProviderID:        req.ProviderID,
		ProviderModelName: req.ProviderModelName,
		Weight:            req.Weight,
		Enabled:           true,
	}
	if mapping.Weight == 0 {
		mapping.Weight = 1
	}
	if !req.Enabled {
		mapping.Enabled = false
	}

	if err := model.DB.Create(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.Preload("Provider").First(&mapping, mapping.ID)

	c.JSON(http.StatusCreated, gin.H{"mapping": toMappingResponse(mapping)})
}

func (h *ModelHandler) UpdateMapping(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	mappingID, err := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	var mapping model.ModelMapping
	if err := model.DB.Where("id = ? AND model_id = ?", mappingID, modelID).First(&mapping).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mapping not found"})
		return
	}

	var req updateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
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

	model.DB.Preload("Provider").First(&mapping, mappingID)

	c.JSON(http.StatusOK, gin.H{"mapping": toMappingResponse(mapping)})
}

func (h *ModelHandler) DeleteMapping(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	mappingID, err := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	if err := model.DB.Where("id = ? AND model_id = ?", mappingID, modelID).Delete(&model.ModelMapping{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mapping deleted"})
}

func (h *ModelHandler) UpdateMappingsOrder(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var m model.Model
	if err := model.DB.First(&m, modelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var req updateMappingsOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Order) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order array is empty"})
		return
	}

	totalMappings := len(req.Order)
	for i, mappingID := range req.Order {
		weight := totalMappings - 1 - i
		if err := model.DB.Model(&model.ModelMapping{}).
			Where("id = ? AND model_id = ?", mappingID, modelID).
			Update("weight", weight).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update mapping %d: %v", mappingID, err)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "mappings order updated"})
}

func toMappingResponse(m model.ModelMapping) mappingResponse {
	var providerResp *providerBasicResponse
	if m.Provider != nil {
		providerResp = &providerBasicResponse{
			ID:               m.Provider.ID,
			Name:             m.Provider.Name,
			OpenAIBaseURL:    m.Provider.OpenAIBaseURL,
			AnthropicBaseURL: m.Provider.AnthropicBaseURL,
		}
	}

	var modelInfoResp *modelInfoResponse
	var pm model.ProviderModel
	if err := model.DB.Where("provider_id = ? AND model_id = ?", m.ProviderID, m.ProviderModelName).First(&pm).Error; err == nil {
		modelInfoResp = &modelInfoResponse{
			ContextWindow:  pm.ContextWindow,
			MaxOutput:      pm.MaxOutput,
			SupportsVision: pm.SupportsVision,
			SupportsTools:  pm.SupportsTools,
			SupportsStream: pm.SupportsStream,
		}
	}

	return mappingResponse{
		ID:                m.ID,
		ProviderID:        m.ProviderID,
		ProviderModelName: m.ProviderModelName,
		Weight:            m.Weight,
		Enabled:           m.Enabled,
		Provider:          providerResp,
		ModelInfo:         modelInfoResp,
	}
}

func calculateEnabledCount(mappings []model.ModelMapping) int {
	enabledCount := 0
	for _, m := range mappings {
		if m.Enabled {
			enabledCount++
		}
	}
	return enabledCount
}

func calculateMinTokens(mappings []model.ModelMapping) (int, int) {
	minContext := 0
	minOutput := 0
	hasEnabled := false

	for _, m := range mappings {
		if !m.Enabled {
			continue
		}

		if m.Provider == nil || !m.Provider.Enabled {
			continue
		}

		hasEnabled = true
		var pm model.ProviderModel
		if err := model.DB.Where("provider_id = ? AND model_id = ?", m.ProviderID, m.ProviderModelName).First(&pm).Error; err != nil {
			continue
		}

		if pm.ContextWindow > 0 {
			if minContext == 0 || pm.ContextWindow < minContext {
				minContext = pm.ContextWindow
			}
		}
		if pm.MaxOutput > 0 {
			if minOutput == 0 || pm.MaxOutput < minOutput {
				minOutput = pm.MaxOutput
			}
		}
	}

	if !hasEnabled {
		return 0, 0
	}

	return minContext, minOutput
}

func calculateCapabilitiesIntersection(mappings []model.ModelMapping) (bool, bool, bool) {
	supportsVision := true
	supportsTools := true
	supportsStream := true
	hasEnabled := false

	for _, m := range mappings {
		if !m.Enabled {
			continue
		}

		if m.Provider == nil || !m.Provider.Enabled {
			continue
		}

		hasEnabled = true
		var pm model.ProviderModel
		if err := model.DB.Where("provider_id = ? AND model_id = ?", m.ProviderID, m.ProviderModelName).First(&pm).Error; err != nil {
			supportsVision = false
			supportsTools = false
			supportsStream = false
			continue
		}

		if !pm.SupportsVision {
			supportsVision = false
		}
		if !pm.SupportsTools {
			supportsTools = false
		}
		if !pm.SupportsStream {
			supportsStream = false
		}
	}

	if !hasEnabled {
		return false, false, false
	}

	return supportsVision, supportsTools, supportsStream
}
