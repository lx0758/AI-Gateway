package mcp

import (
	"encoding/json"
)

const MCP_PROTOCOL_VERSION = "2025-03-26"

var (
	ErrParseError     = &RPCError{Code: -32700, Message: "Parse error"}
	ErrInvalidRequest = &RPCError{Code: -32600, Message: "Invalid request"}
	ErrMethodNotFound = &RPCError{Code: -32601, Message: "Method not found"}
	ErrInvalidParams  = &RPCError{Code: -32602, Message: "Invalid params"}
	ErrInternalError  = &RPCError{Code: -32603, Message: "Internal error"}
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return e.Message
}

func ParseJSONRPCMessage(data []byte) (*JSONRPCRequest, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	if req.JSONRPC != "2.0" {
		return nil, ErrParseError
	}

	if req.Method == "" {
		return nil, ErrInvalidRequest
	}

	return &req, nil
}

func ParseJSONRPCBatch(data []byte) ([]JSONRPCRequest, error) {
	var batch []json.RawMessage
	if err := json.Unmarshal(data, &batch); err != nil {
		return nil, ErrParseError
	}

	var requests []JSONRPCRequest
	for _, item := range batch {
		var req JSONRPCRequest
		if err := json.Unmarshal(item, &req); err != nil {
			return nil, ErrParseError
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func NewResponse(id interface{}, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

func NewErrorResponse(id interface{}, err *RPCError) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Error:   err,
		ID:      id,
	}
}

func NewNotification(method string, params interface{}) *JSONRPCNotification {
	return &JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

func IsNotification(req *JSONRPCRequest) bool {
	return req.ID == nil
}

func MustMarshalJSON(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
