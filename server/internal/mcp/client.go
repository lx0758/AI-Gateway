package mcp

type MCPClient interface {
	IsInitialized() bool
	Initialize() (*JSONRPCResponse, error)
	GetCapabilities() map[string]bool
	ListTools() (*JSONRPCResponse, error)
	CallTool(name string, arguments map[string]interface{}) (*JSONRPCResponse, error)
	ListResources() (*JSONRPCResponse, error)
	ReadResource(uri string) (*JSONRPCResponse, error)
	ListPrompts() (*JSONRPCResponse, error)
	GetPrompt(name string, arguments map[string]interface{}) (*JSONRPCResponse, error)
	Close() error
}

func parseCapabilitiesFromResult(result interface{}) map[string]bool {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return map[string]bool{"tools": true}
	}

	capsInterface, ok := resultMap["capabilities"]
	if !ok {
		return map[string]bool{"tools": true}
	}

	caps, ok := capsInterface.(map[string]interface{})
	if !ok {
		return map[string]bool{"tools": true}
	}

	capabilities := make(map[string]bool)
	for key := range caps {
		capabilities[key] = true
	}

	return capabilities
}
