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

func (h *KeyHandler) List(c *gin.Context) {
	var keys []model.Key
	if err := model.DB.Preload("Models.Model").Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]keyResponse, len(keys))
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

		result[i] = keyResponse{
			ID:        k.ID,
			Key:       maskedKey,
			Name:      k.Name,
			Enabled:   k.Enabled,
			ExpiresAt: k.ExpiresAt,
			CreatedAt: k.CreatedAt,
			Models:    models,
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

func (h *KeyHandler) ListModels(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var models []model.KeyModel
	if err := model.DB.Preload("Model").Where("key_id = ?", id).Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]keyModelResponse, len(models))
	for i, m := range models {
		modelName := ""
		if m.Model != nil {
			modelName = m.Model.Name
		}
		result[i] = keyModelResponse{
			ID:        m.ID,
			ModelID:   m.ModelID,
			ModelName: modelName,
		}
	}

	c.JSON(http.StatusOK, gin.H{"models": result})
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

func (h *KeyHandler) GetMCPTools(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var keyTools []model.KeyMCPTool
	if err := model.DB.Preload("Tool.MCP").Where("key_id = ?", id).Find(&keyTools).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]keyMCPToolResponse, len(keyTools))
	for i, kt := range keyTools {
		toolName := ""
		mcpID := uint(0)
		mcpName := ""
		if kt.Tool != nil {
			toolName = kt.Tool.Name
			mcpID = kt.Tool.MCPID
			if kt.Tool.MCP != nil {
				mcpName = kt.Tool.MCP.Name
			}
		}
		result[i] = keyMCPToolResponse{
			ID:       kt.ID,
			ToolID:   kt.ToolID,
			ToolName: toolName,
			MCPID:    mcpID,
			MCPName:  mcpName,
		}
	}

	c.JSON(http.StatusOK, gin.H{"tools": result})
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

func (h *KeyHandler) GetMCPResources(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var keyResources []model.KeyMCPResource
	if err := model.DB.Preload("Resource.MCP").Where("key_id = ?", id).Find(&keyResources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]keyMCPResourceResponse, len(keyResources))
	for i, kr := range keyResources {
		resourceName := ""
		resourceURI := ""
		mcpID := uint(0)
		mcpName := ""
		if kr.Resource != nil {
			resourceName = kr.Resource.Name
			resourceURI = kr.Resource.URI
			mcpID = kr.Resource.MCPID
			if kr.Resource.MCP != nil {
				mcpName = kr.Resource.MCP.Name
			}
		}
		result[i] = keyMCPResourceResponse{
			ID:           kr.ID,
			ResourceID:   kr.ResourceID,
			ResourceName: resourceName,
			ResourceURI:  resourceURI,
			MCPID:        mcpID,
			MCPName:      mcpName,
		}
	}

	c.JSON(http.StatusOK, gin.H{"resources": result})
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

func (h *KeyHandler) GetMCPPrompts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var keyPrompts []model.KeyMCPPrompt
	if err := model.DB.Preload("Prompt.MCP").Where("key_id = ?", id).Find(&keyPrompts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]keyMCPPromptResponse, len(keyPrompts))
	for i, kp := range keyPrompts {
		promptName := ""
		mcpID := uint(0)
		mcpName := ""
		if kp.Prompt != nil {
			promptName = kp.Prompt.Name
			mcpID = kp.Prompt.MCPID
			if kp.Prompt.MCP != nil {
				mcpName = kp.Prompt.MCP.Name
			}
		}
		result[i] = keyMCPPromptResponse{
			ID:         kp.ID,
			PromptID:   kp.PromptID,
			PromptName: promptName,
			MCPID:      mcpID,
			MCPName:    mcpName,
		}
	}

	c.JSON(http.StatusOK, gin.H{"prompts": result})
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
