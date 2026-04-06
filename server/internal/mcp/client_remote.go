package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type RemoteMCPClient struct {
	url          string
	headers      map[string]string
	httpClient   *http.Client
	initialized  bool
	capabilities map[string]bool
}

func NewRemoteMCPClient(url string, headers map[string]string) *RemoteMCPClient {
	return &RemoteMCPClient{
		url:        url,
		headers:    headers,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *RemoteMCPClient) IsInitialized() bool {
	return c.initialized
}

func (c *RemoteMCPClient) Initialize() (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"protocolVersion": MCP_PROTOCOL_VERSION,
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "ai-gateway",
			"version": "1.0.0",
		},
	}

	resp, err := c.call("initialize", params, 1)
	if err != nil {
		return nil, err
	}

	if resp.Error == nil {
		c.initialized = true
		c.capabilities = parseCapabilitiesFromResult(resp.Result)
	}

	return resp, nil
}

func (c *RemoteMCPClient) GetCapabilities() map[string]bool {
	return c.capabilities
}

func (c *RemoteMCPClient) ListTools() (*JSONRPCResponse, error) {
	return c.call("tools/list", map[string]interface{}{}, 2)
}

func (c *RemoteMCPClient) CallTool(name string, arguments map[string]interface{}) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}
	return c.call("tools/call", params, 3)
}

func (c *RemoteMCPClient) ListResources() (*JSONRPCResponse, error) {
	return c.call("resources/list", map[string]interface{}{}, 4)
}

func (c *RemoteMCPClient) ReadResource(uri string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"uri": uri,
	}
	return c.call("resources/read", params, 5)
}

func (c *RemoteMCPClient) ListPrompts() (*JSONRPCResponse, error) {
	return c.call("prompts/list", map[string]interface{}{}, 6)
}

func (c *RemoteMCPClient) GetPrompt(name string, arguments map[string]interface{}) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}
	return c.call("prompts/get", params, 7)
}

func (c *RemoteMCPClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

func (c *RemoteMCPClient) call(method string, params interface{}, id interface{}) (*JSONRPCResponse, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  MustMarshalJSON(params),
		ID:      id,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	recordRemoteReq(reqBody)

	httpReq, err := http.NewRequest("POST", c.url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream, application/json")
	for key, value := range c.headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody := recordRemoteResp(resp.Body)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respBody)
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/event-stream") {
		return c.parseSSEResponse(respBody)
	}

	var jsonrpcResp JSONRPCResponse
	if err := json.NewDecoder(respBody).Decode(&jsonrpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &jsonrpcResp, nil
}

func (c *RemoteMCPClient) parseSSEResponse(body io.Reader) (*JSONRPCResponse, error) {
	scanner := bufio.NewScanner(body)
	var jsonData string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			jsonData = strings.TrimPrefix(line, "data: ")
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read SSE stream: %w", err)
	}

	if jsonData == "" {
		return nil, fmt.Errorf("no data in SSE response")
	}

	var jsonrpcResp JSONRPCResponse
	if err := json.Unmarshal([]byte(jsonData), &jsonrpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode SSE data: %w", err)
	}

	return &jsonrpcResp, nil
}
