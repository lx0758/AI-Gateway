package handler

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/mcp"
	"ai-gateway/internal/model"
)

var mcpManager = mcp.NewMCPManager()

type MCPHandler struct{}

func NewMCPHandler() *MCPHandler {
	return &MCPHandler{}
}

type mcpCreateRequest struct {
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Target  string `json:"target"`
	Params  string `json:"params"`
	Enabled *bool  `json:"enabled"`
}

type mcpUpdateRequest struct {
	Name    *string `json:"name"`
	Type    *string `json:"type"`
	Target  *string `json:"target"`
	Params  *string `json:"params"`
	Enabled *bool   `json:"enabled"`
}

type mcpResponse struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Target       string     `json:"target,omitempty"`
	Params       string     `json:"params,omitempty"`
	Enabled      bool       `json:"enabled"`
	Capabilities string     `json:"capabilities,omitempty"`
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type mcpToolResponse struct {
	ID          uint   `json:"id"`
	MCPID       uint   `json:"mcp_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema string `json:"input_schema"`
	Enabled     bool   `json:"enabled"`
}

type mcpResourceResponse struct {
	ID          uint   `json:"id"`
	MCPID       uint   `json:"mcp_id"`
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mime_type"`
	Enabled     bool   `json:"enabled"`
}

type mcpPromptResponse struct {
	ID          uint   `json:"id"`
	MCPID       uint   `json:"mcp_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Arguments   string `json:"arguments"`
	Enabled     bool   `json:"enabled"`
}

var nameRegex = regexp.MustCompile(`^[0-9a-zA-Z_-]{2,200}$`)

func validateName(name string) bool {
	return nameRegex.MatchString(name)
}

func (h *MCPHandler) Create(c *gin.Context) {
	var req mcpCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validateName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name: must be 2-200 characters, only [0-9a-zA-Z_-] allowed"})
		return
	}

	if req.Type != "remote" && req.Type != "local" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type: must be 'remote' or 'local'"})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	m := model.MCP{
		Name:    req.Name,
		Type:    req.Type,
		Target:  req.Target,
		Params:  req.Params,
		Enabled: enabled,
	}

	err := mcpManager.TestClient(&m)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "test MCP failed: " + err.Error()})
		return
	}

	if err := model.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := mcpManager.SyncMCP(&m); err != nil {
		log.Printf("[MCP] failed to sync MCP %d: %v", m.ID, err)
	}

	c.JSON(http.StatusCreated, gin.H{"mcp": h.toResponse(&m)})
}

func (h *MCPHandler) List(c *gin.Context) {
	var mcps []model.MCP
	if err := model.DB.Find(&mcps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]mcpResponse, len(mcps))
	for i, m := range mcps {
		result[i] = h.toResponse(&m)
	}

	c.JSON(http.StatusOK, gin.H{"mcps": result})
}

func (h *MCPHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m model.MCP
	if err := model.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mcp": h.toResponse(&m)})
}

func (h *MCPHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m model.MCP
	if err := model.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp not found"})
		return
	}

	var req mcpUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	needsConnectionTest := false
	if req.Name != nil {
		if !validateName(*req.Name) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name: must be 2-200 characters, only [0-9a-zA-Z_-] allowed"})
			return
		}
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		if *req.Type != "remote" && *req.Type != "local" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type: must be 'remote' or 'local'"})
			return
		}
		updates["type"] = *req.Type
		needsConnectionTest = true
	}
	if req.Target != nil {
		updates["target"] = *req.Target
		needsConnectionTest = true
	}
	if req.Params != nil {
		updates["params"] = *req.Params
		needsConnectionTest = true
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if needsConnectionTest {
		// 使用前端传来的完整数据进行连接测试
		testMCP := model.MCP{
			Name:   m.Name,   // 默认使用数据库值
			Type:   m.Type,   // 默认使用数据库值
			Target: m.Target, // 默认使用数据库值
			Params: m.Params, // 默认使用数据库值
		}

		// 如果前端传了新值，使用新值
		if req.Name != nil {
			testMCP.Name = *req.Name
		}
		if req.Type != nil {
			testMCP.Type = *req.Type
		}
		if req.Target != nil {
			testMCP.Target = *req.Target
		}
		if req.Params != nil {
			testMCP.Params = *req.Params
		}

		err := mcpManager.TestClient(&testMCP)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "test MCP failed: " + err.Error()})
			return
		}
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&m).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	model.DB.First(&m, id)

	if needsConnectionTest {
		mcpManager.CloseClient(m.ID)
		if err := mcpManager.SyncMCP(&m); err != nil {
			log.Printf("[MCP] failed to sync MCP %d: %v", m.ID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"mcp": h.toResponse(&m)})
}

func (h *MCPHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	model.DB.Where("mcp_id = ?", id).Delete(&model.MCPTool{})
	model.DB.Where("mcp_id = ?", id).Delete(&model.MCPResource{})
	model.DB.Where("mcp_id = ?", id).Delete(&model.MCPPrompt{})

	if err := model.DB.Delete(&model.MCP{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mcp deleted"})
}

func (h *MCPHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m model.MCP
	if err := model.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp not found"})
		return
	}

	client, err := mcpManager.GetClient(&m)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	resp, err := client.Initialize()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if resp.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": resp.Error.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "connection successful",
		"capabilities": resp.Result,
	})
}

func (h *MCPHandler) Sync(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var m model.MCP
	if err := model.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mcp not found"})
		return
	}

	if err := mcpManager.SyncMCP(&m); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "sync completed successfully",
	})
}

func (h *MCPHandler) toResponse(m *model.MCP) mcpResponse {
	return mcpResponse{
		ID:           m.ID,
		Name:         m.Name,
		Type:         m.Type,
		Target:       m.Target,
		Params:       m.Params,
		Enabled:      m.Enabled,
		Capabilities: m.Capabilities,
		LastSyncAt:   m.LastSyncAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (h *MCPHandler) ListTools(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var tools []model.MCPTool
	if err := model.DB.Where("mcp_id = ?", id).Find(&tools).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]mcpToolResponse, len(tools))
	for i, t := range tools {
		result[i] = mcpToolResponse{
			ID:          t.ID,
			MCPID:       t.MCPID,
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
			Enabled:     t.Enabled,
		}
	}

	c.JSON(http.StatusOK, gin.H{"tools": result})
}

func (h *MCPHandler) ListResources(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var resources []model.MCPResource
	if err := model.DB.Where("mcp_id = ?", id).Find(&resources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]mcpResourceResponse, len(resources))
	for i, r := range resources {
		result[i] = mcpResourceResponse{
			ID:          r.ID,
			MCPID:       r.MCPID,
			URI:         r.URI,
			Name:        r.Name,
			Description: r.Description,
			MimeType:    r.MimeType,
			Enabled:     r.Enabled,
		}
	}

	c.JSON(http.StatusOK, gin.H{"resources": result})
}

func (h *MCPHandler) ListPrompts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var prompts []model.MCPPrompt
	if err := model.DB.Where("mcp_id = ?", id).Find(&prompts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]mcpPromptResponse, len(prompts))
	for i, p := range prompts {
		result[i] = mcpPromptResponse{
			ID:          p.ID,
			MCPID:       p.MCPID,
			Name:        p.Name,
			Description: p.Description,
			Arguments:   p.Arguments,
			Enabled:     p.Enabled,
		}
	}

	c.JSON(http.StatusOK, gin.H{"prompts": result})
}

func (h *MCPHandler) UpdateTool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool id"})
		return
	}

	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&model.MCPTool{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "tool updated"})
}

func (h *MCPHandler) UpdateResource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}

	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&model.MCPResource{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "resource updated"})
}

func (h *MCPHandler) UpdatePrompt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prompt id"})
		return
	}

	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&model.MCPPrompt{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "prompt updated"})
}
