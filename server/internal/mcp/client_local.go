package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type LocalMCPClient struct {
	command      string
	envVars      []string
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.Reader
	stderr       io.Reader
	mu           sync.Mutex
	lastUsed     time.Time
	idleTimer    *time.Timer
	initialized  bool
	capabilities map[string]bool
}

func NewLocalMCPClient(command string, envVars map[string]string) *LocalMCPClient {
	var envList []string
	for key, value := range envVars {
		envList = append(envList, fmt.Sprintf("%s=%s", key, value))
	}

	return &LocalMCPClient{
		command: command,
		envVars: envList,
	}
}

func (c *LocalMCPClient) IsInitialized() bool {
	return c.initialized
}

func (c *LocalMCPClient) Initialize() (*JSONRPCResponse, error) {
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

func (c *LocalMCPClient) GetCapabilities() map[string]bool {
	return c.capabilities
}

func (c *LocalMCPClient) ListTools() (*JSONRPCResponse, error) {
	return c.call("tools/list", map[string]interface{}{}, 2)
}

func (c *LocalMCPClient) CallTool(name string, arguments map[string]interface{}) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}
	return c.call("tools/call", params, 3)
}

func (c *LocalMCPClient) ListResources() (*JSONRPCResponse, error) {
	return c.call("resources/list", map[string]interface{}{}, 4)
}

func (c *LocalMCPClient) ReadResource(uri string) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"uri": uri,
	}
	return c.call("resources/read", params, 5)
}

func (c *LocalMCPClient) ListPrompts() (*JSONRPCResponse, error) {
	return c.call("prompts/list", map[string]interface{}{}, 6)
}

func (c *LocalMCPClient) GetPrompt(name string, arguments map[string]interface{}) (*JSONRPCResponse, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}
	return c.call("prompts/get", params, 7)
}

func (c *LocalMCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.idleTimer != nil {
		c.idleTimer.Stop()
	}

	c.stop()
	return nil
}

func (c *LocalMCPClient) call(method string, params interface{}, id interface{}) (*JSONRPCResponse, error) {
	if err := c.start(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastUsed = time.Now()
	c.startIdleTimer()

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

	reqBody = append(reqBody, '\n')

	if _, err := c.stdin.Write(reqBody); err != nil {
		c.stop()
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}

	scanner := bufio.NewScanner(c.stdout)
	if !scanner.Scan() {
		errMsg := "failed to read response from process"
		if err := scanner.Err(); err != nil {
			errMsg = fmt.Sprintf("failed to read response: %v", err)
		}
		c.stop()
		return nil, fmt.Errorf("%s", errMsg)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		c.stop()
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *LocalMCPClient) start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd != nil && c.cmd.Process != nil {
		return nil
	}

	parts := strings.Fields(c.command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = append(os.Environ(), c.envVars...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr

	go c.readStderr()

	c.lastUsed = time.Now()
	c.startIdleTimer()

	return nil
}

func (c *LocalMCPClient) readStderr() {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		log.Printf("[MCP Local] stderr: %s", scanner.Text())
	}
}

func (c *LocalMCPClient) startIdleTimer() {
	if c.idleTimer != nil {
		c.idleTimer.Stop()
	}

	c.idleTimer = time.AfterFunc(5*time.Minute, func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if time.Since(c.lastUsed) >= 5*time.Minute {
			c.stop()
		}
	})
}

func (c *LocalMCPClient) stop() {
	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Kill(); err != nil {
			log.Printf("[MCP Local] failed to kill process: %v", err)
		}
		c.cmd = nil
		c.stdin = nil
		c.stdout = nil
		c.stderr = nil
		c.initialized = false
	}
}
