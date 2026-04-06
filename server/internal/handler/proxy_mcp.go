package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/mcp"
	"ai-gateway/internal/model"
	"ai-gateway/internal/utils"
)

type MCPProxyHandler struct{}

func NewMCPProxyHandler() *MCPProxyHandler {
	return &MCPProxyHandler{}
}

func (h *MCPProxyHandler) hasFullToolAccess(keyID uint) bool {
	var count int64
	model.DB.Model(&model.KeyMCPTool{}).Where("key_id = ?", keyID).Count(&count)
	return count == 0
}

func (h *MCPProxyHandler) hasFullResourceAccess(keyID uint) bool {
	var count int64
	model.DB.Model(&model.KeyMCPResource{}).Where("key_id = ?", keyID).Count(&count)
	return count == 0
}

func (h *MCPProxyHandler) hasFullPromptAccess(keyID uint) bool {
	var count int64
	model.DB.Model(&model.KeyMCPPrompt{}).Where("key_id = ?", keyID).Count(&count)
	return count == 0
}

func (h *MCPProxyHandler) Handle(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodPost:
		h.handlePost(c)
	case http.MethodGet:
		h.handleGet(c)
	case http.MethodDelete:
		h.handleDelete(c)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"jsonrpc": "2.0",
			"error":   gin.H{"code": mcp.ErrInvalidRequest.Code, "message": "Method not allowed"},
			"id":      nil,
		})
	}
}

func (h *MCPProxyHandler) handlePost(c *gin.Context) {
	accept := c.GetHeader("Accept")
	_, hasSSE := h.parseAcceptHeader(accept)

	contentType := c.GetHeader("Content-Type")
	if !h.isJSONContentType(contentType) {
		h.writeJSONError(c, http.StatusUnsupportedMediaType, "Unsupported Media Type: Content-Type must be application/json")
		return
	}

	var rawRequest json.RawMessage
	if err := c.ShouldBindJSON(&rawRequest); err != nil {
		if hasSSE {
			h.writeSSEError(c, nil, mcp.ErrParseError)
		} else {
			h.writeJSONError(c, http.StatusBadRequest, "Parse error: "+err.Error())
		}
		return
	}

	var batch []json.RawMessage
	isBatch := json.Unmarshal(rawRequest, &batch) == nil

	if hasSSE {
		h.handlePostSSE(c, rawRequest, isBatch)
	} else {
		h.handlePostJSON(c, rawRequest, isBatch)
	}
}

func (h *MCPProxyHandler) handlePostJSON(c *gin.Context, rawRequest json.RawMessage, isBatch bool) {
	c.Header("Content-Type", "application/json")

	if isBatch {
		requests, err := mcp.ParseJSONRPCBatch(rawRequest)
		if err != nil {
			c.JSON(http.StatusOK, []interface{}{mcp.NewErrorResponse(nil, mcp.ErrParseError)})
			return
		}

		var responses []*mcp.JSONRPCResponse
		for _, req := range requests {
			if mcp.IsNotification(&req) {
				go h.routeMethod(c, &req)
				continue
			}
			response := h.routeMethod(c, &req)
			if response != nil {
				responses = append(responses, response)
			}
		}

		if len(responses) == 0 {
			c.Status(http.StatusAccepted)
			return
		}
		c.JSON(http.StatusOK, responses)
	} else {
		req, err := mcp.ParseJSONRPCMessage(rawRequest)
		if err != nil {
			c.JSON(http.StatusOK, mcp.NewErrorResponse(nil, mcp.ErrParseError))
			return
		}

		if mcp.IsNotification(req) {
			go h.routeMethod(c, req)
			c.Status(http.StatusAccepted)
			return
		}

		response := h.routeMethod(c, req)
		c.JSON(http.StatusOK, response)
	}
}

func (h *MCPProxyHandler) handlePostSSE(c *gin.Context, rawRequest json.RawMessage, isBatch bool) {
	h.writeSSEHeaders(c)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE not supported"})
		return
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

		if isBatch {
			requests, err := mcp.ParseJSONRPCBatch(rawRequest)
			if err != nil {
				h.writeSSEMessage(c, flusher, mcp.NewErrorResponse(nil, mcp.ErrParseError))
				return
			}

			notificationCount := 0
			for _, req := range requests {
				if mcp.IsNotification(&req) {
					go h.routeMethod(c, &req)
					notificationCount++
					continue
				}
				response := h.routeMethod(c, &req)
				if response != nil {
					h.writeSSEMessage(c, flusher, response)
				}
			}

			if notificationCount == len(requests) {
				c.Status(http.StatusAccepted)
			}
		} else {
			req, err := mcp.ParseJSONRPCMessage(rawRequest)
			if err != nil {
				h.writeSSEMessage(c, flusher, mcp.NewErrorResponse(nil, mcp.ErrParseError))
				return
			}

			if mcp.IsNotification(req) {
				go h.routeMethod(c, req)
				c.Status(http.StatusAccepted)
				return
			}

			response := h.routeMethod(c, req)
			h.writeSSEMessage(c, flusher, response)
		}
	}()

	<-done
}

func (h *MCPProxyHandler) handleGet(c *gin.Context) {
	accept := c.GetHeader("Accept")
	_, hasSSE := h.parseAcceptHeader(accept)

	if !hasSSE {
		h.writeJSONError(c, http.StatusNotAcceptable, "Not Acceptable: Client must accept text/event-stream")
		return
	}

	h.writeSSEHeaders(c)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE not supported"})
		return
	}

	notify := c.Writer.CloseNotify()

	c.Writer.Write([]byte(": keep-alive\n\n"))
	flusher.Flush()

	<-notify
}

func (h *MCPProxyHandler) handleDelete(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (h *MCPProxyHandler) parseAcceptHeader(accept string) (hasJSON, hasSSE bool) {
	if accept == "" {
		return true, false
	}

	types := strings.Split(accept, ",")
	for _, t := range types {
		mediaType := strings.TrimSpace(strings.Split(t, ";")[0])
		mediaType = strings.ToLower(mediaType)

		switch mediaType {
		case "*/*", "application/*", "application/json":
			hasJSON = true
		case "text/*", "text/event-stream":
			hasSSE = true
		}
	}

	return hasJSON, hasSSE
}

func (h *MCPProxyHandler) isJSONContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	types := strings.Split(contentType, ",")
	for _, t := range types {
		mediaType := strings.TrimSpace(strings.Split(t, ";")[0])
		if strings.ToLower(mediaType) == "application/json" {
			return true
		}
	}

	return false
}

func (h *MCPProxyHandler) writeSSEHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "close")
	c.Header("Access-Control-Allow-Origin", "*")
}

func (h *MCPProxyHandler) writeSSEMessage(c *gin.Context, flusher http.Flusher, response *mcp.JSONRPCResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		return
	}

	c.Writer.Write([]byte("event: message\n"))
	c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
	flusher.Flush()
}

func (h *MCPProxyHandler) writeSSEError(c *gin.Context, flusher http.Flusher, rpcErr *mcp.RPCError) {
	h.writeSSEHeaders(c)

	if flusher == nil {
		var ok bool
		flusher, ok = c.Writer.(http.Flusher)
		if !ok {
			return
		}
	}

	response := mcp.NewErrorResponse(nil, rpcErr)
	h.writeSSEMessage(c, flusher, response)
}

func (h *MCPProxyHandler) writeJSONError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"jsonrpc": "2.0",
		"error": gin.H{
			"code":    mcp.ErrInvalidRequest.Code,
			"message": message,
		},
		"id": nil,
	})
}

func (h *MCPProxyHandler) routeMethod(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return h.handleInitialize(c, req)
	case "tools/list":
		return h.handleToolsList(c, req)
	case "tools/call":
		return h.handleToolsCall(c, req)
	case "resources/list":
		return h.handleResourcesList(c, req)
	case "resources/read":
		return h.handleResourcesRead(c, req)
	case "resources/subscribe":
		return h.handleResourcesSubscribe(c, req)
	case "prompts/list":
		return h.handlePromptsList(c, req)
	case "prompts/get":
		return h.handlePromptsGet(c, req)
	case "ping":
		return h.handlePing(c, req)
	case "notifications/initialized":
		return nil
	default:
		return mcp.NewErrorResponse(req.ID, mcp.ErrMethodNotFound)
	}
}

func (h *MCPProxyHandler) handleInitialize(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	apiKey := c.MustGet("api_key").(*model.Key)

	var tools []interface{}
	var resources []interface{}
	var prompts []interface{}

	if h.hasFullToolAccess(apiKey.ID) {
		var allTools []model.MCPTool
		model.DB.Preload("MCP").Where("enabled = ?", true).Find(&allTools)
		for _, t := range allTools {
			if t.MCP != nil && t.MCP.Enabled {
				tools = append(tools, map[string]interface{}{
					"name":        t.MCP.Name + "." + t.Name,
					"description": t.Description,
					"inputSchema": json.RawMessage(t.InputSchema),
				})
			}
		}
	} else {
		var keyTools []model.KeyMCPTool
		model.DB.Preload("Tool.MCP").Where("key_id = ?", apiKey.ID).Find(&keyTools)
		for _, kt := range keyTools {
			if kt.Tool != nil && kt.Tool.MCP != nil && kt.Tool.Enabled && kt.Tool.MCP.Enabled {
				tools = append(tools, map[string]interface{}{
					"name":        kt.Tool.MCP.Name + "." + kt.Tool.Name,
					"description": kt.Tool.Description,
					"inputSchema": json.RawMessage(kt.Tool.InputSchema),
				})
			}
		}
	}

	if h.hasFullResourceAccess(apiKey.ID) {
		var allResources []model.MCPResource
		model.DB.Preload("MCP").Where("enabled = ?", true).Find(&allResources)
		for _, r := range allResources {
			if r.MCP != nil && r.MCP.Enabled {
				resources = append(resources, map[string]interface{}{
					"uri":         "mcp://" + r.MCP.Name + "/" + r.URI,
					"name":        r.Name,
					"description": r.Description,
					"mimeType":    r.MimeType,
				})
			}
		}
	} else {
		var keyResources []model.KeyMCPResource
		model.DB.Preload("Resource.MCP").Where("key_id = ?", apiKey.ID).Find(&keyResources)
		for _, kr := range keyResources {
			if kr.Resource != nil && kr.Resource.MCP != nil && kr.Resource.Enabled && kr.Resource.MCP.Enabled {
				resources = append(resources, map[string]interface{}{
					"uri":         "mcp://" + kr.Resource.MCP.Name + "/" + kr.Resource.URI,
					"name":        kr.Resource.Name,
					"description": kr.Resource.Description,
					"mimeType":    kr.Resource.MimeType,
				})
			}
		}
	}

	if h.hasFullPromptAccess(apiKey.ID) {
		var allPrompts []model.MCPPrompt
		model.DB.Preload("MCP").Where("enabled = ?", true).Find(&allPrompts)
		for _, p := range allPrompts {
			if p.MCP != nil && p.MCP.Enabled {
				prompts = append(prompts, map[string]interface{}{
					"name":        p.MCP.Name + "." + p.Name,
					"description": p.Description,
					"arguments":   json.RawMessage(p.Arguments),
				})
			}
		}
	} else {
		var keyPrompts []model.KeyMCPPrompt
		model.DB.Preload("Prompt.MCP").Where("key_id = ?", apiKey.ID).Find(&keyPrompts)
		for _, kp := range keyPrompts {
			if kp.Prompt != nil && kp.Prompt.MCP != nil && kp.Prompt.Enabled && kp.Prompt.MCP.Enabled {
				prompts = append(prompts, map[string]interface{}{
					"name":        kp.Prompt.MCP.Name + "." + kp.Prompt.Name,
					"description": kp.Prompt.Description,
					"arguments":   json.RawMessage(kp.Prompt.Arguments),
				})
			}
		}
	}

	return &mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"protocolVersion": mcp.MCP_PROTOCOL_VERSION,
			"capabilities": map[string]interface{}{
				"tools":     map[string]interface{}{},
				"resources": map[string]interface{}{},
				"prompts":   map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "ai-gateway-mcp-proxy",
				"version": "1.0.0",
			},
		},
		ID: req.ID,
	}
}

func (h *MCPProxyHandler) handleToolsList(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	apiKey := c.MustGet("api_key").(*model.Key)

	var tools []interface{}

	if h.hasFullToolAccess(apiKey.ID) {
		var allTools []model.MCPTool
		model.DB.Preload("MCP").Where("enabled = ?", true).Find(&allTools)
		for _, t := range allTools {
			if t.MCP != nil && t.MCP.Enabled {
				tools = append(tools, map[string]interface{}{
					"name":        t.MCP.Name + "." + t.Name,
					"description": t.Description,
					"inputSchema": json.RawMessage(t.InputSchema),
				})
			}
		}
	} else {
		var keyTools []model.KeyMCPTool
		model.DB.Preload("Tool.MCP").Where("key_id = ?", apiKey.ID).Find(&keyTools)
		for _, kt := range keyTools {
			if kt.Tool != nil && kt.Tool.MCP != nil && kt.Tool.Enabled && kt.Tool.MCP.Enabled {
				tools = append(tools, map[string]interface{}{
					"name":        kt.Tool.MCP.Name + "." + kt.Tool.Name,
					"description": kt.Tool.Description,
					"inputSchema": json.RawMessage(kt.Tool.InputSchema),
				})
			}
		}
	}

	return mcp.NewResponse(req.ID, map[string]interface{}{
		"tools": tools,
	})
}

func (h *MCPProxyHandler) handleToolsCall(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	apiKey := c.MustGet("api_key").(*model.Key)
	startTime := time.Now()
	clientIPs := utils.GetClientIPInfo(c)

	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return mcp.NewErrorResponse(req.ID, mcp.ErrInvalidParams)
	}

	parts := strings.SplitN(params.Name, ".", 2)
	if len(parts) != 2 {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: "invalid tool name format, expected: mcp_name.tool_name",
		})
	}

	mcpName := parts[0]
	toolName := parts[1]

	var m model.MCP
	if err := model.DB.Where("name = ? AND enabled = ?", mcpName, true).First(&m).Error; err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: fmt.Sprintf("MCP not found: %s", mcpName),
		})
	}

	var tool model.MCPTool
	if err := model.DB.Where("mcp_id = ? AND name = ? AND enabled = ?", m.ID, toolName, true).First(&tool).Error; err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: fmt.Sprintf("tool not found: %s", params.Name),
		})
	}

	if !h.hasFullToolAccess(apiKey.ID) {
		var keyTool model.KeyMCPTool
		if err := model.DB.Where("key_id = ? AND tool_id = ?", apiKey.ID, tool.ID).First(&keyTool).Error; err != nil {
			return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
				Code:    mcp.ErrInvalidParams.Code,
				Message: "permission denied",
			})
		}
	}

	resp, err := mcpManager.CallTool(&m, toolName, params.Arguments)
	latencyMs := int(time.Since(startTime).Milliseconds())

	status := "success"
	errorMsg := ""
	inputSize := 0
	outputSize := 0

	if err != nil {
		status = "error"
		errorMsg = err.Error()
	} else {
		status = "success"
		if resp != nil {
			respBytes, _ := json.Marshal(resp)
			outputSize = len(respBytes)
		}
	}

	argsBytes, _ := json.Marshal(params.Arguments)
	inputSize = len(argsBytes)

	mcpLog := NewMCPLog(
		"default",
		clientIPs,
		apiKey.ID,
		apiKey.Name,
		m.ID,
		m.Name,
		m.Type,
		"tool",
		toolName,
		"call",
		inputSize,
		outputSize,
		latencyMs,
		status,
		errorMsg,
	)
	model.DB.Create(&mcpLog)
	log.Println(mcpLog.String())

	if err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInternalError.Code,
			Message: err.Error(),
		})
	}

	resp.ID = req.ID
	return resp
}

func (h *MCPProxyHandler) handleResourcesList(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	apiKey := c.MustGet("api_key").(*model.Key)

	var resources []interface{}

	if h.hasFullResourceAccess(apiKey.ID) {
		var allResources []model.MCPResource
		model.DB.Preload("MCP").Where("enabled = ?", true).Find(&allResources)
		for _, r := range allResources {
			if r.MCP != nil && r.MCP.Enabled {
				resources = append(resources, map[string]interface{}{
					"uri":         "mcp://" + r.MCP.Name + "/" + r.URI,
					"name":        r.Name,
					"description": r.Description,
					"mimeType":    r.MimeType,
				})
			}
		}
	} else {
		var keyResources []model.KeyMCPResource
		model.DB.Preload("Resource.MCP").Where("key_id = ?", apiKey.ID).Find(&keyResources)
		for _, kr := range keyResources {
			if kr.Resource != nil && kr.Resource.MCP != nil && kr.Resource.Enabled && kr.Resource.MCP.Enabled {
				resources = append(resources, map[string]interface{}{
					"uri":         "mcp://" + kr.Resource.MCP.Name + "/" + kr.Resource.URI,
					"name":        kr.Resource.Name,
					"description": kr.Resource.Description,
					"mimeType":    kr.Resource.MimeType,
				})
			}
		}
	}

	return mcp.NewResponse(req.ID, map[string]interface{}{
		"resources": resources,
	})
}

func (h *MCPProxyHandler) handleResourcesRead(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	apiKey := c.MustGet("api_key").(*model.Key)
	startTime := time.Now()
	clientIPs := utils.GetClientIPInfo(c)

	var params struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return mcp.NewErrorResponse(req.ID, mcp.ErrInvalidParams)
	}

	if !strings.HasPrefix(params.URI, "mcp://") {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: "invalid resource URI format, expected: mcp://mcp_name/original_uri",
		})
	}

	uriWithoutPrefix := strings.TrimPrefix(params.URI, "mcp://")
	parts := strings.SplitN(uriWithoutPrefix, "/", 2)
	if len(parts) != 2 {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: "invalid resource URI format, expected: mcp://mcp_name/original_uri",
		})
	}

	mcpName := parts[0]
	originalURI := parts[1]

	var m model.MCP
	if err := model.DB.Where("name = ? AND enabled = ?", mcpName, true).First(&m).Error; err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: fmt.Sprintf("MCP not found: %s", mcpName),
		})
	}

	var resource model.MCPResource
	if err := model.DB.Where("mcp_id = ? AND uri = ? AND enabled = ?", m.ID, originalURI, true).First(&resource).Error; err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: fmt.Sprintf("resource not found: %s", params.URI),
		})
	}

	if !h.hasFullResourceAccess(apiKey.ID) {
		var keyResource model.KeyMCPResource
		if err := model.DB.Where("key_id = ? AND resource_id = ?", apiKey.ID, resource.ID).First(&keyResource).Error; err != nil {
			return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
				Code:    mcp.ErrInvalidParams.Code,
				Message: "permission denied",
			})
		}
	}

	resp, err := mcpManager.ReadResource(&m, originalURI)
	latencyMs := int(time.Since(startTime).Milliseconds())

	status := "success"
	errorMsg := ""
	inputSize := 0
	outputSize := 0

	if err != nil {
		status = "error"
		errorMsg = err.Error()
	} else {
		status = "success"
		if resp != nil {
			respBytes, _ := json.Marshal(resp)
			outputSize = len(respBytes)
		}
	}

	mcpLog := NewMCPLog(
		"default",
		clientIPs,
		apiKey.ID,
		apiKey.Name,
		m.ID,
		m.Name,
		m.Type,
		"resource",
		originalURI,
		"read",
		inputSize,
		outputSize,
		latencyMs,
		status,
		errorMsg,
	)
	model.DB.Create(&mcpLog)
	log.Println(mcpLog.String())

	if err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInternalError.Code,
			Message: err.Error(),
		})
	}

	resp.ID = req.ID
	return resp
}

func (h *MCPProxyHandler) handleResourcesSubscribe(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	return mcp.NewErrorResponse(req.ID, mcp.ErrMethodNotFound)
}

func (h *MCPProxyHandler) handlePromptsList(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	apiKey := c.MustGet("api_key").(*model.Key)

	var prompts []interface{}

	if h.hasFullPromptAccess(apiKey.ID) {
		var allPrompts []model.MCPPrompt
		model.DB.Preload("MCP").Where("enabled = ?", true).Find(&allPrompts)
		for _, p := range allPrompts {
			if p.MCP != nil && p.MCP.Enabled {
				prompts = append(prompts, map[string]interface{}{
					"name":        p.MCP.Name + "." + p.Name,
					"description": p.Description,
					"arguments":   json.RawMessage(p.Arguments),
				})
			}
		}
	} else {
		var keyPrompts []model.KeyMCPPrompt
		model.DB.Preload("Prompt.MCP").Where("key_id = ?", apiKey.ID).Find(&keyPrompts)
		for _, kp := range keyPrompts {
			if kp.Prompt != nil && kp.Prompt.MCP != nil && kp.Prompt.Enabled && kp.Prompt.MCP.Enabled {
				prompts = append(prompts, map[string]interface{}{
					"name":        kp.Prompt.MCP.Name + "." + kp.Prompt.Name,
					"description": kp.Prompt.Description,
					"arguments":   json.RawMessage(kp.Prompt.Arguments),
				})
			}
		}
	}

	return mcp.NewResponse(req.ID, map[string]interface{}{
		"prompts": prompts,
	})
}

func (h *MCPProxyHandler) handlePromptsGet(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	apiKey := c.MustGet("api_key").(*model.Key)
	startTime := time.Now()
	clientIPs := utils.GetClientIPInfo(c)

	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return mcp.NewErrorResponse(req.ID, mcp.ErrInvalidParams)
	}

	parts := strings.SplitN(params.Name, ".", 2)
	if len(parts) != 2 {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: "invalid prompt name format, expected: mcp_name.prompt_name",
		})
	}

	mcpName := parts[0]
	promptName := parts[1]

	var m model.MCP
	if err := model.DB.Where("name = ? AND enabled = ?", mcpName, true).First(&m).Error; err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: fmt.Sprintf("MCP not found: %s", mcpName),
		})
	}

	var prompt model.MCPPrompt
	if err := model.DB.Where("mcp_id = ? AND name = ? AND enabled = ?", m.ID, promptName, true).First(&prompt).Error; err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInvalidParams.Code,
			Message: fmt.Sprintf("prompt not found: %s", params.Name),
		})
	}

	if !h.hasFullPromptAccess(apiKey.ID) {
		var keyPrompt model.KeyMCPPrompt
		if err := model.DB.Where("key_id = ? AND prompt_id = ?", apiKey.ID, prompt.ID).First(&keyPrompt).Error; err != nil {
			return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
				Code:    mcp.ErrInvalidParams.Code,
				Message: "permission denied",
			})
		}
	}

	resp, err := mcpManager.GetPrompt(&m, promptName, params.Arguments)
	latencyMs := int(time.Since(startTime).Milliseconds())

	status := "success"
	errorMsg := ""
	inputSize := 0
	outputSize := 0

	if err != nil {
		status = "error"
		errorMsg = err.Error()
	} else {
		status = "success"
		if resp != nil {
			respBytes, _ := json.Marshal(resp)
			outputSize = len(respBytes)
		}
	}

	argsBytes, _ := json.Marshal(params.Arguments)
	inputSize = len(argsBytes)

	mcpLog := NewMCPLog(
		"default",
		clientIPs,
		apiKey.ID,
		apiKey.Name,
		m.ID,
		m.Name,
		m.Type,
		"prompt",
		promptName,
		"get",
		inputSize,
		outputSize,
		latencyMs,
		status,
		errorMsg,
	)
	model.DB.Create(&mcpLog)
	log.Println(mcpLog.String())

	if err != nil {
		return mcp.NewErrorResponse(req.ID, &mcp.RPCError{
			Code:    mcp.ErrInternalError.Code,
			Message: err.Error(),
		})
	}

	resp.ID = req.ID
	return resp
}

func (h *MCPProxyHandler) handlePing(c *gin.Context, req *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	return mcp.NewResponse(req.ID, map[string]interface{}{})
}
