package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
)

type KeyHandler struct{}

type createKeyRequest struct {
	Name      string  `json:"name" binding:"required"`
	Models    []uint  `json:"models"`
	ExpiresAt *string `json:"expires_at"`
}

type updateKeyRequest struct {
	Name      *string `json:"name"`
	Models    []uint  `json:"models"`
	ExpiresAt *string `json:"expires_at"`
	Enabled   *bool   `json:"enabled"`
}

type keyModelResponse struct {
	ID        uint   `json:"id"`
	ModelID   uint   `json:"model_id"`
	ModelName string `json:"model_name"`
}

type keyResponse struct {
	ID        uint               `json:"id"`
	Key       string             `json:"key"`
	Name      string             `json:"name"`
	Enabled   bool               `json:"enabled"`
	ExpiresAt *time.Time         `json:"expires_at"`
	CreatedAt time.Time          `json:"created_at"`
	Models    []keyModelResponse `json:"models,omitempty"`
}

type keyCreateResponse struct {
	Key    keyResponse `json:"key"`
	RawKey string      `json:"raw_key"`
}

type keyMCPToolResponse struct {
	ID       uint   `json:"id"`
	ToolID   uint   `json:"tool_id"`
	ToolName string `json:"tool_name"`
	MCPID    uint   `json:"mcp_id"`
	MCPName  string `json:"mcp_name"`
}

type keyMCPResourceResponse struct {
	ID           uint   `json:"id"`
	ResourceID   uint   `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ResourceURI  string `json:"resource_uri"`
	MCPID        uint   `json:"mcp_id"`
	MCPName      string `json:"mcp_name"`
}

type keyMCPPromptResponse struct {
	ID         uint   `json:"id"`
	PromptID   uint   `json:"prompt_id"`
	PromptName string `json:"prompt_name"`
	MCPID      uint   `json:"mcp_id"`
	MCPName    string `json:"mcp_name"`
}

func NewKeyHandler() *KeyHandler {
	return &KeyHandler{}
}

func generateKey() string {
	bytes := make([]byte, 24)
	rand.Read(bytes)
	return "sk-" + hex.EncodeToString(bytes)
}

type keyListItemResponse struct {
	ID                uint               `json:"id"`
	Key               string             `json:"key"`
	Name              string             `json:"name"`
	Enabled           bool               `json:"enabled"`
	ExpiresAt         *time.Time         `json:"expires_at"`
	CreatedAt         time.Time          `json:"created_at"`
	Models            []keyModelResponse `json:"models,omitempty"`
	MCPToolsCount     int                `json:"mcp_tools_count"`
	MCPResourcesCount int                `json:"mcp_resources_count"`
	MCPPromptsCount   int                `json:"mcp_prompts_count"`
}

func (h *KeyHandler) List(c *gin.Context) {
	var keys []model.Key
	if err := model.DB.Preload("Models.Model").Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]keyListItemResponse, len(keys))
	for i, k := range keys {
		maskedKey := k.Key
		if len(maskedKey) > 8 {
			maskedKey = maskedKey[:8] + "****" + maskedKey[len(maskedKey)-4:]
		}

		models := make([]keyModelResponse, len(k.Models))
		for j, m := range k.Models {
			modelName := ""
			if m.Model != nil {
				modelName = m.Model.Name
			}
			models[j] = keyModelResponse{
				ID:        m.ID,
				ModelID:   m.ModelID,
				ModelName: modelName,
			}
		}

		var mcpToolsCount, mcpResourcesCount, mcppromptsCount int64
		model.DB.Model(&model.KeyMCPTool{}).Where("key_id = ?", k.ID).Count(&mcpToolsCount)
		model.DB.Model(&model.KeyMCPResource{}).Where("key_id = ?", k.ID).Count(&mcpResourcesCount)
		model.DB.Model(&model.KeyMCPPrompt{}).Where("key_id = ?", k.ID).Count(&mcppromptsCount)

		result[i] = keyListItemResponse{
			ID:        k.ID,
			Key:       maskedKey,
			Name:      k.Name,
			Enabled:   k.Enabled,
			ExpiresAt: k.ExpiresAt,
			CreatedAt: k.CreatedAt,
			Models:    models,
			MCPToolsCount:     int(mcpToolsCount),
			MCPResourcesCount: int(mcpResourcesCount),
			MCPPromptsCount:   int(mcppromptsCount),
		}
	}

	c.JSON(http.StatusOK, gin.H{"keys": result})
}

func (h *KeyHandler) Create(c *gin.Context) {
	var req createKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse("2006-01-02 15:04:05", *req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	key := model.Key{
		Key:       generateKey(),
		Name:      req.Name,
		ExpiresAt: expiresAt,
		Enabled:   true,
	}

	if err := model.DB.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, modelID := range req.Models {
		var m model.Model
		if err := model.DB.First(&m, modelID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model not found: " + strconv.FormatUint(uint64(modelID), 10)})
			return
		}
		akm := model.KeyModel{
			KeyID:   key.ID,
			ModelID: modelID,
		}
		model.DB.Create(&akm)
	}

	model.DB.Preload("Models.Model").First(&key, key.ID)

	models := make([]keyModelResponse, len(key.Models))
	for j, m := range key.Models {
		modelName := ""
		if m.Model != nil {
			modelName = m.Model.Name
		}
		models[j] = keyModelResponse{
			ID:        m.ID,
			ModelID:   m.ModelID,
			ModelName: modelName,
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"key": keyResponse{
			ID:        key.ID,
			Key:       key.Key,
			Name:      key.Name,
			Enabled:   key.Enabled,
			ExpiresAt: key.ExpiresAt,
			CreatedAt: key.CreatedAt,
			Models:    models,
		},
		"raw_key": key.Key,
	})
}

func (h *KeyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	model.DB.Where("key_id = ?", id).Delete(&model.KeyModel{})

	if err := model.DB.Delete(&model.Key{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}

func (h *KeyHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var req updateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ExpiresAt != nil {
		if t, err := time.Parse("2006-01-02 15:04:05", *req.ExpiresAt); err == nil {
			updates["expires_at"] = t
		}
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&key).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.Models != nil {
		model.DB.Where("key_id = ?", key.ID).Delete(&model.KeyModel{})
		for _, modelID := range req.Models {
			var m model.Model
			if err := model.DB.First(&m, modelID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "model not found: " + strconv.FormatUint(uint64(modelID), 10)})
				return
			}
			akm := model.KeyModel{KeyID: key.ID, ModelID: modelID}
			model.DB.Create(&akm)
		}
	}

	model.DB.Preload("Models.Model").First(&key, id)

	models := make([]keyModelResponse, len(key.Models))
	for j, m := range key.Models {
		modelName := ""
		if m.Model != nil {
			modelName = m.Model.Name
		}
		models[j] = keyModelResponse{
			ID:        m.ID,
			ModelID:   m.ModelID,
			ModelName: modelName,
		}
	}

	c.JSON(http.StatusOK, gin.H{"key": keyResponse{
		ID:        key.ID,
		Key:       key.Key,
		Name:      key.Name,
		Enabled:   key.Enabled,
		ExpiresAt: key.ExpiresAt,
		CreatedAt: key.CreatedAt,
		Models:    models,
	}})
}

type modelWithStatusResponse struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	MappingCount     int    `json:"mapping_count"`
	MinContextWindow int    `json:"min_context_window"`
	MinMaxOutput     int    `json:"min_max_output"`
	SupportsVision   bool   `json:"supports_vision"`
	SupportsTools    bool   `json:"supports_tools"`
	SupportsStream   bool   `json:"supports_stream"`
	Selected         bool   `json:"selected"`
}

func (h *KeyHandler) ListModels(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var allModels []model.Model
	if err := model.DB.Preload("Mappings.Provider").Where("enabled = ?", true).Find(&allModels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var keyModelIDs []uint
	model.DB.Model(&model.KeyModel{}).Where("key_id = ?", id).Pluck("model_id", &keyModelIDs)

	keyModelMap := make(map[uint]bool)
	for _, mid := range keyModelIDs {
		keyModelMap[mid] = true
	}

	result := make([]modelWithStatusResponse, len(allModels))
	for i, m := range allModels {
		minContext, minOutput := calculateMinTokens(m.Mappings)
		supportsVision, supportsTools, supportsStream := calculateCapabilitiesIntersection(m.Mappings)

		result[i] = modelWithStatusResponse{
			ID:               m.ID,
			Name:             m.Name,
			MappingCount:     len(m.Mappings),
			MinContextWindow: minContext,
			MinMaxOutput:     minOutput,
			SupportsVision:   supportsVision,
			SupportsTools:    supportsTools,
			SupportsStream:   supportsStream,
			Selected:         keyModelMap[m.ID],
		}
	}

	c.JSON(http.StatusOK, gin.H{"models": result})
}

func (h *KeyHandler) AddModel(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	modelID, err := strconv.ParseUint(c.Param("model_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, keyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var m model.Model
	if err := model.DB.First(&m, modelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var existing model.KeyModel
	if err := model.DB.Where("key_id = ? AND model_id = ?", keyID, modelID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "association already exists"})
		return
	}

	keyModel := model.KeyModel{
		KeyID:   uint(keyID),
		ModelID: uint(modelID),
	}
	if err := model.DB.Create(&keyModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "model association added"})
}

func (h *KeyHandler) RemoveModel(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	modelID, err := strconv.ParseUint(c.Param("model_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	model.DB.Where("key_id = ? AND model_id = ?", keyID, modelID).Delete(&model.KeyModel{})

	c.JSON(http.StatusOK, gin.H{"message": "model association removed"})
}

func (h *KeyHandler) ClearModels(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	model.DB.Where("key_id = ?", keyID).Delete(&model.KeyModel{})

	c.JSON(http.StatusOK, gin.H{"message": "all model associations cleared"})
}

func (h *KeyHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	maskedKey := key.Key
	if len(maskedKey) > 8 {
		maskedKey = maskedKey[:8] + "****" + maskedKey[len(maskedKey)-4:]
	}

	c.JSON(http.StatusOK, gin.H{
		"key": keyResponse{
			ID:        key.ID,
			Key:       maskedKey,
			Name:      key.Name,
			Enabled:   key.Enabled,
			ExpiresAt: key.ExpiresAt,
			CreatedAt: key.CreatedAt,
		},
	})
}

func (h *KeyHandler) Reset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	newKey := generateKey()

	if err := model.DB.Model(&key).Update("key", newKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.Preload("Models.Model").First(&key, id)

	models := make([]keyModelResponse, len(key.Models))
	for j, m := range key.Models {
		modelName := ""
		if m.Model != nil {
			modelName = m.Model.Name
		}
		models[j] = keyModelResponse{
			ID:        m.ID,
			ModelID:   m.ModelID,
			ModelName: modelName,
		}
	}

	maskedKey := key.Key
	if len(maskedKey) > 8 {
		maskedKey = maskedKey[:8] + "****" + maskedKey[len(maskedKey)-4:]
	}

	c.JSON(http.StatusOK, gin.H{
		"key": keyResponse{
			ID:        key.ID,
			Key:       maskedKey,
			Name:      key.Name,
			Enabled:   key.Enabled,
			ExpiresAt: key.ExpiresAt,
			CreatedAt: key.CreatedAt,
			Models:    models,
		},
		"raw_key": key.Key,
	})
}

type toolWithStatusResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	MCPID       uint   `json:"mcp_id"`
	MCPName     string `json:"mcp_name"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
}

func (h *KeyHandler) GetMCPTools(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var allTools []model.MCPTool
	if err := model.DB.Preload("MCP", "enabled = ?", true).
		Joins("LEFT JOIN mcps ON mcps.id = mcp_tools.mcp_id").
		Where("mcps.enabled = ? AND mcp_tools.enabled = ?", true, true).
		Find(&allTools).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var keyToolIDs []uint
	model.DB.Model(&model.KeyMCPTool{}).Where("key_id = ?", id).Pluck("tool_id", &keyToolIDs)

	keyToolMap := make(map[uint]bool)
	for _, tid := range keyToolIDs {
		keyToolMap[tid] = true
	}

	result := make([]toolWithStatusResponse, len(allTools))
	for i, t := range allTools {
		mcpName := ""
		if t.MCP != nil {
			mcpName = t.MCP.Name
		}
		result[i] = toolWithStatusResponse{
			ID:          t.ID,
			Name:        t.Name,
			MCPID:       t.MCPID,
			MCPName:     mcpName,
			Description: t.Description,
			Selected:    keyToolMap[t.ID],
		}
	}

	c.JSON(http.StatusOK, gin.H{"tools": result})
}

func (h *KeyHandler) AddMCPTool(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	toolID, err := strconv.ParseUint(c.Param("tool_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool id"})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, keyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var tool model.MCPTool
	if err := model.DB.First(&tool, toolID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	var existing model.KeyMCPTool
	if err := model.DB.Where("key_id = ? AND tool_id = ?", keyID, toolID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "association already exists"})
		return
	}

	keyTool := model.KeyMCPTool{
		KeyID:  uint(keyID),
		ToolID: uint(toolID),
	}
	if err := model.DB.Create(&keyTool).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tool association added"})
}

func (h *KeyHandler) RemoveMCPTool(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	toolID, err := strconv.ParseUint(c.Param("tool_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool id"})
		return
	}

	model.DB.Where("key_id = ? AND tool_id = ?", keyID, toolID).Delete(&model.KeyMCPTool{})

	c.JSON(http.StatusOK, gin.H{"message": "tool association removed"})
}

func (h *KeyHandler) ClearMCPTools(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	model.DB.Where("key_id = ?", keyID).Delete(&model.KeyMCPTool{})

	c.JSON(http.StatusOK, gin.H{"message": "all tool associations cleared"})
}

func (h *KeyHandler) UpdateMCPTools(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		ToolIDs []uint `json:"tool_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	model.DB.Where("key_id = ?", id).Delete(&model.KeyMCPTool{})

	for _, toolID := range req.ToolIDs {
		var tool model.MCPTool
		if err := model.DB.First(&tool, toolID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tool not found: " + strconv.FormatUint(uint64(toolID), 10)})
			return
		}
		keyTool := model.KeyMCPTool{
			KeyID:  key.ID,
			ToolID: toolID,
		}
		model.DB.Create(&keyTool)
	}

	c.JSON(http.StatusOK, gin.H{"message": "MCP tools updated"})
}

type resourceWithStatusResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	MCPID       uint   `json:"mcp_id"`
	MCPName     string `json:"mcp_name"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	MimeType    string `json:"mime_type"`
	Selected    bool   `json:"selected"`
}

func (h *KeyHandler) GetMCPResources(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var allResources []model.MCPResource
	if err := model.DB.Preload("MCP", "enabled = ?", true).
		Joins("LEFT JOIN mcps ON mcps.id = mcp_resources.mcp_id").
		Where("mcps.enabled = ? AND mcp_resources.enabled = ?", true, true).
		Find(&allResources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var keyResourceIDs []uint
	model.DB.Model(&model.KeyMCPResource{}).Where("key_id = ?", id).Pluck("resource_id", &keyResourceIDs)

	keyResourceMap := make(map[uint]bool)
	for _, rid := range keyResourceIDs {
		keyResourceMap[rid] = true
	}

	result := make([]resourceWithStatusResponse, len(allResources))
	for i, r := range allResources {
		mcpName := ""
		if r.MCP != nil {
			mcpName = r.MCP.Name
		}
		result[i] = resourceWithStatusResponse{
			ID:          r.ID,
			Name:        r.Name,
			MCPID:       r.MCPID,
			MCPName:     mcpName,
			Description: r.Description,
			URI:         r.URI,
			MimeType:    r.MimeType,
			Selected:    keyResourceMap[r.ID],
		}
	}

	c.JSON(http.StatusOK, gin.H{"resources": result})
}

func (h *KeyHandler) AddMCPResource(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	resourceID, err := strconv.ParseUint(c.Param("resource_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, keyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var resource model.MCPResource
	if err := model.DB.First(&resource, resourceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		return
	}

	var existing model.KeyMCPResource
	if err := model.DB.Where("key_id = ? AND resource_id = ?", keyID, resourceID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "association already exists"})
		return
	}

	keyResource := model.KeyMCPResource{
		KeyID:      uint(keyID),
		ResourceID: uint(resourceID),
	}
	if err := model.DB.Create(&keyResource).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resource association added"})
}

func (h *KeyHandler) RemoveMCPResource(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	resourceID, err := strconv.ParseUint(c.Param("resource_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}

	model.DB.Where("key_id = ? AND resource_id = ?", keyID, resourceID).Delete(&model.KeyMCPResource{})

	c.JSON(http.StatusOK, gin.H{"message": "resource association removed"})
}

func (h *KeyHandler) ClearMCPResources(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	model.DB.Where("key_id = ?", keyID).Delete(&model.KeyMCPResource{})

	c.JSON(http.StatusOK, gin.H{"message": "all resource associations cleared"})
}

func (h *KeyHandler) UpdateMCPResources(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		ResourceIDs []uint `json:"resource_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	model.DB.Where("key_id = ?", id).Delete(&model.KeyMCPResource{})

	for _, resourceID := range req.ResourceIDs {
		var resource model.MCPResource
		if err := model.DB.First(&resource, resourceID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource not found: " + strconv.FormatUint(uint64(resourceID), 10)})
			return
		}
		keyResource := model.KeyMCPResource{
			KeyID:      key.ID,
			ResourceID: resourceID,
		}
		model.DB.Create(&keyResource)
	}

	c.JSON(http.StatusOK, gin.H{"message": "MCP resources updated"})
}

type promptWithStatusResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	MCPID       uint   `json:"mcp_id"`
	MCPName     string `json:"mcp_name"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
}

func (h *KeyHandler) GetMCPPrompts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var allPrompts []model.MCPPrompt
	if err := model.DB.Preload("MCP", "enabled = ?", true).
		Joins("LEFT JOIN mcps ON mcps.id = mcp_prompts.mcp_id").
		Where("mcps.enabled = ? AND mcp_prompts.enabled = ?", true, true).
		Find(&allPrompts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var keyPromptIDs []uint
	model.DB.Model(&model.KeyMCPPrompt{}).Where("key_id = ?", id).Pluck("prompt_id", &keyPromptIDs)

	keyPromptMap := make(map[uint]bool)
	for _, pid := range keyPromptIDs {
		keyPromptMap[pid] = true
	}

	result := make([]promptWithStatusResponse, len(allPrompts))
	for i, p := range allPrompts {
		mcpName := ""
		if p.MCP != nil {
			mcpName = p.MCP.Name
		}
		result[i] = promptWithStatusResponse{
			ID:          p.ID,
			Name:        p.Name,
			MCPID:       p.MCPID,
			MCPName:     mcpName,
			Description: p.Description,
			Selected:    keyPromptMap[p.ID],
		}
	}

	c.JSON(http.StatusOK, gin.H{"prompts": result})
}

func (h *KeyHandler) AddMCPPrompt(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	promptID, err := strconv.ParseUint(c.Param("prompt_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prompt id"})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, keyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var prompt model.MCPPrompt
	if err := model.DB.First(&prompt, promptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "prompt not found"})
		return
	}

	var existing model.KeyMCPPrompt
	if err := model.DB.Where("key_id = ? AND prompt_id = ?", keyID, promptID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "association already exists"})
		return
	}

	keyPrompt := model.KeyMCPPrompt{
		KeyID:    uint(keyID),
		PromptID: uint(promptID),
	}
	if err := model.DB.Create(&keyPrompt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prompt association added"})
}

func (h *KeyHandler) RemoveMCPPrompt(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	promptID, err := strconv.ParseUint(c.Param("prompt_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prompt id"})
		return
	}

	model.DB.Where("key_id = ? AND prompt_id = ?", keyID, promptID).Delete(&model.KeyMCPPrompt{})

	c.JSON(http.StatusOK, gin.H{"message": "prompt association removed"})
}

func (h *KeyHandler) ClearMCPPrompts(c *gin.Context) {
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	model.DB.Where("key_id = ?", keyID).Delete(&model.KeyMCPPrompt{})

	c.JSON(http.StatusOK, gin.H{"message": "all prompt associations cleared"})
}

func (h *KeyHandler) UpdateMCPPrompts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		PromptIDs []uint `json:"prompt_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	model.DB.Where("key_id = ?", id).Delete(&model.KeyMCPPrompt{})

	for _, promptID := range req.PromptIDs {
		var prompt model.MCPPrompt
		if err := model.DB.First(&prompt, promptID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "prompt not found: " + strconv.FormatUint(uint64(promptID), 10)})
			return
		}
		keyPrompt := model.KeyMCPPrompt{
			KeyID:    key.ID,
			PromptID: promptID,
		}
		model.DB.Create(&keyPrompt)
	}

	c.JSON(http.StatusOK, gin.H{"message": "MCP prompts updated"})
}
