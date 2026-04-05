package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"ai-gateway/internal/model"

	"gopkg.in/yaml.v3"
)

func parseMapString(data string) (map[string]string, error) {
	if data == "" {
		return nil, nil
	}

	result := make(map[string]string)

	if err := yaml.Unmarshal([]byte(data), &result); err == nil {
		return result, nil
	}

	if err := json.Unmarshal([]byte(data), &result); err == nil {
		return result, nil
	}

	return nil, fmt.Errorf("invalid format: must be valid JSON or YAML")
}

type MCPManager struct {
	mu      sync.RWMutex
	clients map[uint]MCPClient
}

func NewMCPManager() *MCPManager {
	return &MCPManager{
		clients: make(map[uint]MCPClient),
	}
}

func (m *MCPManager) TestClient(mcp *model.MCP) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	testClient, err := m.createClientLocked(mcp)
	if err != nil {
		return err
	}
	defer testClient.Close()

	initResp, err := testClient.Initialize()
	if err != nil {
		return err
	}

	if initResp.Error != nil {
		return fmt.Errorf("initialization failed: %s", initResp.Error.Message)
	}

	return nil
}

func (m *MCPManager) GetClient(mcp *model.MCP) (MCPClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[mcp.ID]
	if !exists {
		var err error
		client, err = m.createClientLocked(mcp)
		if err != nil {
			return nil, err
		}
		m.clients[mcp.ID] = client
	}

	if !client.IsInitialized() {
		initResp, err := client.Initialize()
		if err != nil {
			return nil, fmt.Errorf("initialize failed: %w", err)
		}
		if initResp.Error != nil {
			return nil, fmt.Errorf("initialize error: %s", initResp.Error.Message)
		}
	}

	return client, nil
}

func (m *MCPManager) CloseClient(serviceID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.clients[serviceID]; exists {
		if err := client.Close(); err != nil {
			log.Printf("[MCP Manager] failed to close client %d: %v", serviceID, err)
		}
		delete(m.clients, serviceID)
	}
}

func (m *MCPManager) SyncMCP(mcp *model.MCP) error {
	client, err := m.GetClient(mcp)
	if err != nil {
		return err
	}

	capabilities := client.GetCapabilities()
	if capabilities == nil {
		capabilities = map[string]bool{"tools": true}
	}

	capabilitiesJSON, _ := json.Marshal(capabilities)
	mcp.Capabilities = string(capabilitiesJSON)

	if capabilities["tools"] {
		if err := m.syncTools(client, mcp.ID); err != nil {
			log.Printf("[MCP Manager] failed to sync tools: %v", err)
		}
	}

	if capabilities["resources"] {
		if err := m.syncResources(client, mcp.ID); err != nil {
			log.Printf("[MCP Manager] failed to sync resources: %v", err)
		}
	}

	if capabilities["prompts"] {
		if err := m.syncPrompts(client, mcp.ID); err != nil {
			log.Printf("[MCP Manager] failed to sync prompts: %v", err)
		}
	}

	now := time.Now()
	mcp.LastSyncAt = &now
	if err := model.DB.Save(mcp).Error; err != nil {
		return fmt.Errorf("failed to save mcp: %w", err)
	}

	return nil
}

func (m *MCPManager) syncTools(client MCPClient, serviceID uint) error {
	resp, err := client.ListTools()
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("list tools error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid result type")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid tools type")
	}

	for _, toolInterface := range tools {
		toolMap, ok := toolInterface.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := toolMap["name"].(string)
		description, _ := toolMap["description"].(string)
		inputSchemaJSON, _ := json.Marshal(toolMap["inputSchema"])

		var existingTool model.MCPTool
		err := model.DB.Where("mcp_id = ? AND name = ?", serviceID, name).First(&existingTool).Error
		if err == nil {
			existingTool.Description = description
			existingTool.InputSchema = string(inputSchemaJSON)
			model.DB.Save(&existingTool)
		} else {
			newTool := model.MCPTool{
				MCPID:       serviceID,
				Name:        name,
				Description: description,
				InputSchema: string(inputSchemaJSON),
				Enabled:     true,
			}
			model.DB.Create(&newTool)
		}
	}

	return nil
}

func (m *MCPManager) syncResources(client MCPClient, serviceID uint) error {
	resp, err := client.ListResources()
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("list resources error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid result type")
	}

	resources, ok := result["resources"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid resources type")
	}

	for _, resourceInterface := range resources {
		resourceMap, ok := resourceInterface.(map[string]interface{})
		if !ok {
			continue
		}

		uri, _ := resourceMap["uri"].(string)
		name, _ := resourceMap["name"].(string)
		description, _ := resourceMap["description"].(string)
		mimeType, _ := resourceMap["mimeType"].(string)

		var existingResource model.MCPResource
		err := model.DB.Where("mcp_id = ? AND name = ?", serviceID, name).First(&existingResource).Error
		if err == nil {
			existingResource.URI = uri
			existingResource.Description = description
			existingResource.MimeType = mimeType
			model.DB.Save(&existingResource)
		} else {
			newResource := model.MCPResource{
				MCPID:       serviceID,
				URI:         uri,
				Name:        name,
				Description: description,
				MimeType:    mimeType,
				Enabled:     true,
			}
			model.DB.Create(&newResource)
		}
	}

	return nil
}

func (m *MCPManager) syncPrompts(client MCPClient, serviceID uint) error {
	resp, err := client.ListPrompts()
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("list prompts error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid result type")
	}

	prompts, ok := result["prompts"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid prompts type")
	}

	for _, promptInterface := range prompts {
		promptMap, ok := promptInterface.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := promptMap["name"].(string)
		description, _ := promptMap["description"].(string)
		argumentsJSON, _ := json.Marshal(promptMap["arguments"])

		var existingPrompt model.MCPPrompt
		err := model.DB.Where("mcp_id = ? AND name = ?", serviceID, name).First(&existingPrompt).Error
		if err == nil {
			existingPrompt.Description = description
			existingPrompt.Arguments = string(argumentsJSON)
			model.DB.Save(&existingPrompt)
		} else {
			newPrompt := model.MCPPrompt{
				MCPID:       serviceID,
				Name:        name,
				Description: description,
				Arguments:   string(argumentsJSON),
				Enabled:     true,
			}
			model.DB.Create(&newPrompt)
		}
	}

	return nil
}

func (m *MCPManager) CallTool(mcp *model.MCP, toolName string, arguments map[string]interface{}) (*JSONRPCResponse, error) {
	client, err := m.GetClient(mcp)
	if err != nil {
		return nil, err
	}

	return client.CallTool(toolName, arguments)
}

func (m *MCPManager) ReadResource(mcp *model.MCP, uri string) (*JSONRPCResponse, error) {
	client, err := m.GetClient(mcp)
	if err != nil {
		return nil, err
	}

	return client.ReadResource(uri)
}

func (m *MCPManager) GetPrompt(mcp *model.MCP, promptName string, arguments map[string]interface{}) (*JSONRPCResponse, error) {
	client, err := m.GetClient(mcp)
	if err != nil {
		return nil, err
	}

	return client.GetPrompt(promptName, arguments)
}

func (m *MCPManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for serviceID, client := range m.clients {
		if err := client.Close(); err != nil {
			log.Printf("[MCP Manager] failed to close client %d: %v", serviceID, err)
		}
	}

	m.clients = make(map[uint]MCPClient)
}

func (m *MCPManager) createClientLocked(mcp *model.MCP) (MCPClient, error) {
	var client MCPClient

	switch mcp.Type {
	case "remote":
		headers, err := parseMapString(mcp.Params)
		if err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		client = NewRemoteMCPClient(mcp.Target, headers)
	case "local":
		envVars, err := parseMapString(mcp.Params)
		if err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		client = NewLocalMCPClient(mcp.Target, envVars)
	default:
		return nil, fmt.Errorf("unknown mcp type: %s", mcp.Type)
	}

	return client, nil
}
