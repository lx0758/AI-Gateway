package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type AliasHandler struct{}

type createAliasRequest struct {
	Name    string `json:"name" binding:"required"`
	Enabled bool   `json:"enabled"`
}

type updateAliasRequest struct {
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

type aliasResponse struct {
	ID           uint              `json:"id"`
	Alias        string            `json:"alias"`
	Enabled      bool              `json:"enabled"`
	MappingCount int               `json:"mapping_count"`
	CreatedAt    string            `json:"created_at"`
	Mappings     []mappingResponse `json:"mappings,omitempty"`
}

type mappingResponse struct {
	ID                uint                   `json:"id"`
	ProviderID        uint                   `json:"provider_id"`
	ProviderModelName string                 `json:"provider_model_name"`
	Weight            int                    `json:"weight"`
	Enabled           bool                   `json:"enabled"`
	Provider          *providerBasicResponse `json:"provider,omitempty"`
}

type providerBasicResponse struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	OpenAIBaseURL    string `json:"openai_base_url"`
	AnthropicBaseURL string `json:"anthropic_base_url"`
}

func NewAliasHandler() *AliasHandler {
	return &AliasHandler{}
}

func (h *AliasHandler) List(c *gin.Context) {
	var aliases []model.Alias
	if err := model.DB.Preload("Mappings.Provider").Find(&aliases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]aliasResponse, len(aliases))
	for i, a := range aliases {
		mappings := make([]mappingResponse, len(a.Mappings))
		for j, m := range a.Mappings {
			mappings[j] = toMappingResponse(m)
		}

		result[i] = aliasResponse{
			ID:           a.ID,
			Alias:        a.Name,
			Enabled:      a.Enabled,
			MappingCount: len(a.Mappings),
			CreatedAt:    a.CreatedAt.Format("2006-01-02 15:04:05"),
			Mappings:     mappings,
		}
	}

	c.JSON(http.StatusOK, gin.H{"aliases": result})
}

func (h *AliasHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var alias model.Alias
	if err := model.DB.First(&alias, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alias not found"})
		return
	}

	var mappings []model.AliasMapping
	model.DB.Preload("Provider").Where("alias_id = ?", alias.ID).Order("weight DESC").Find(&mappings)

	mappingResponses := make([]mappingResponse, len(mappings))
	for j, m := range mappings {
		mappingResponses[j] = toMappingResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"alias": aliasResponse{
		ID:        alias.ID,
		Alias:     alias.Name,
		Enabled:   alias.Enabled,
		CreatedAt: alias.CreatedAt.Format("2006-01-02 15:04:05"),
		Mappings:  mappingResponses,
	}})
}

func (h *AliasHandler) Create(c *gin.Context) {
	var req createAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alias := model.Alias{
		Name:    req.Name,
		Enabled: true,
	}
	if !req.Enabled {
		alias.Enabled = false
	}

	if err := model.DB.Create(&alias).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"alias": aliasResponse{
		ID:           alias.ID,
		Alias:        alias.Name,
		Enabled:      alias.Enabled,
		MappingCount: 0,
		CreatedAt:    alias.CreatedAt.Format("2006-01-02 15:04:05"),
	}})
}

func (h *AliasHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var alias model.Alias
	if err := model.DB.First(&alias, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alias not found"})
		return
	}

	var req updateAliasRequest
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
		if err := model.DB.Model(&alias).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	model.DB.First(&alias, id)

	var mappings []model.AliasMapping
	model.DB.Preload("Provider").Where("alias_id = ?", alias.ID).Order("weight DESC").Find(&mappings)

	mappingResponses := make([]mappingResponse, len(mappings))
	for j, m := range mappings {
		mappingResponses[j] = toMappingResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"alias": aliasResponse{
		ID:        alias.ID,
		Alias:     alias.Name,
		Enabled:   alias.Enabled,
		CreatedAt: alias.CreatedAt.Format("2006-01-02 15:04:05"),
		Mappings:  mappingResponses,
	}})
}

func (h *AliasHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := model.DB.Delete(&model.Alias{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alias deleted"})
}

func (h *AliasHandler) ListMappings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var alias model.Alias
	if err := model.DB.First(&alias, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alias not found"})
		return
	}

	var mappings []model.AliasMapping
	model.DB.Preload("Provider").Where("alias_id = ?", alias.ID).Order("weight DESC").Find(&mappings)

	result := make([]mappingResponse, len(mappings))
	for i, m := range mappings {
		result[i] = toMappingResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"mappings": result})
}

func (h *AliasHandler) CreateMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var alias model.Alias
	if err := model.DB.First(&alias, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alias not found"})
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

	mapping := model.AliasMapping{
		AliasID:           alias.ID,
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

func (h *AliasHandler) UpdateMapping(c *gin.Context) {
	aliasID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alias id"})
		return
	}

	mappingID, err := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	var mapping model.AliasMapping
	if err := model.DB.Where("id = ? AND alias_id = ?", mappingID, aliasID).First(&mapping).Error; err != nil {
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

func (h *AliasHandler) DeleteMapping(c *gin.Context) {
	aliasID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alias id"})
		return
	}

	mappingID, err := strconv.ParseUint(c.Param("mid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	if err := model.DB.Where("id = ? AND alias_id = ?", mappingID, aliasID).Delete(&model.AliasMapping{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mapping deleted"})
}

func toMappingResponse(m model.AliasMapping) mappingResponse {
	var providerResp *providerBasicResponse
	if m.Provider != nil {
		providerResp = &providerBasicResponse{
			ID:               m.Provider.ID,
			Name:             m.Provider.Name,
			OpenAIBaseURL:    m.Provider.OpenAIBaseURL,
			AnthropicBaseURL: m.Provider.AnthropicBaseURL,
		}
	}
	return mappingResponse{
		ID:                m.ID,
		ProviderID:        m.ProviderID,
		ProviderModelName: m.ProviderModelName,
		Weight:            m.Weight,
		Enabled:           m.Enabled,
		Provider:          providerResp,
	}
}
