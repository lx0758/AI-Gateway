package transformer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type AnthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []AnthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	Tools     []AnthropicTool    `json:"tools,omitempty"`
	Stream    bool               `json:"stream,omitempty"`
}

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type AnthropicResponse struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Role         string             `json:"role"`
	Content      []AnthropicContent `json:"content"`
	Model        string             `json:"model"`
	StopReason   string             `json:"stop_reason"`
	StopSequence string             `json:"stop_sequence"`
	Usage        AnthropicUsage     `json:"usage"`
}

type AnthropicContent struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type OpenAIToAnthropicTransformer struct{}

func NewOpenAIToAnthropicTransformer() *OpenAIToAnthropicTransformer {
	return &OpenAIToAnthropicTransformer{}
}

func (t *OpenAIToAnthropicTransformer) TransformRequest(req *OpenAIRequest) (interface{}, error) {
	anthropicReq := &AnthropicRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Messages:  []AnthropicMessage{},
		Stream:    req.Stream,
	}

	if anthropicReq.MaxTokens == 0 {
		anthropicReq.MaxTokens = 4096
	}

	var systemPrompt string
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			switch v := msg.Content.(type) {
			case string:
				systemPrompt = v
			}
		} else {
			content := t.convertContent(msg.Content)
			anthropicReq.Messages = append(anthropicReq.Messages, AnthropicMessage{
				Role:    msg.Role,
				Content: content,
			})
		}
	}
	anthropicReq.System = systemPrompt

	if len(req.Tools) > 0 {
		anthropicReq.Tools = make([]AnthropicTool, len(req.Tools))
		for i, tool := range req.Tools {
			anthropicReq.Tools[i] = AnthropicTool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: tool.Function.Parameters,
			}
		}
	}

	return anthropicReq, nil
}

func (t *OpenAIToAnthropicTransformer) convertContent(content interface{}) interface{} {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		return v
	default:
		return fmt.Sprintf("%v", content)
	}
}

func (t *OpenAIToAnthropicTransformer) TransformResponse(body []byte) (*OpenAIResponse, error) {
	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, err
	}

	var content string
	var toolCalls []ToolCall

	for _, c := range anthropicResp.Content {
		switch c.Type {
		case "text":
			content += c.Text
		case "tool_use":
			args, _ := json.Marshal(c.Input)
			toolCalls = append(toolCalls, ToolCall{
				ID:   c.ID,
				Type: "function",
				Function: FunctionCall{
					Name:      c.Name,
					Arguments: string(args),
				},
			})
		}
	}

	finishReason := t.mapStopReason(anthropicResp.StopReason)

	return &OpenAIResponse{
		ID:      anthropicResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   anthropicResp.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: &Message{
					Role:      "assistant",
					Content:   content,
					ToolCalls: toolCalls,
				},
				FinishReason: finishReason,
			},
		},
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}, nil
}

func (t *OpenAIToAnthropicTransformer) mapStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	default:
		return "stop"
	}
}

func (t *OpenAIToAnthropicTransformer) TransformStream(reader io.Reader, writer io.Writer) *StreamResult {
	result := &StreamResult{}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var messageID string
	var model string

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") && !strings.HasPrefix(line, "event: ") {
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			eventType := strings.TrimPrefix(line, "event: ")
			scanner.Scan()
			dataLine := scanner.Text()
			data := strings.TrimPrefix(dataLine, "data: ")

			switch eventType {
			case "message_start":
				var msg struct {
					Message struct {
						ID    string         `json:"id"`
						Model string         `json:"model"`
						Usage AnthropicUsage `json:"usage"`
					} `json:"message"`
				}
				if err := json.Unmarshal([]byte(data), &msg); err == nil {
					messageID = msg.Message.ID
					model = msg.Message.Model
					if result.Usage == nil {
						result.Usage = &Usage{}
					}
					result.Usage.PromptTokens = msg.Message.Usage.InputTokens
				}

			case "content_block_delta":
				var delta struct {
					Delta struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"delta"`
				}
				if err := json.Unmarshal([]byte(data), &delta); err == nil && delta.Delta.Type == "text_delta" {
					chunk := StreamChunk{
						ID:      messageID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   model,
						Choices: []Choice{
							{
								Index: 0,
								Delta: &Delta{
									Content: delta.Delta.Text,
								},
							},
						},
					}
					t.writeChunk(writer, chunk)
				}

			case "message_delta":
				var delta struct {
					Delta struct {
						StopReason string `json:"stop_reason"`
					} `json:"delta"`
					Usage struct {
						OutputTokens int `json:"output_tokens"`
					} `json:"usage"`
				}
				if err := json.Unmarshal([]byte(data), &delta); err == nil {
					if result.Usage == nil {
						result.Usage = &Usage{}
					}
					result.Usage.CompletionTokens = delta.Usage.OutputTokens

					chunk := StreamChunk{
						ID:      messageID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   model,
						Choices: []Choice{
							{
								Index:        0,
								Delta:        &Delta{},
								FinishReason: t.mapStopReason(delta.Delta.StopReason),
							},
						},
					}
					t.writeChunk(writer, chunk)
				}

			case "message_stop":
				fmt.Fprintf(writer, "data: [DONE]\n\n")
				if f, ok := writer.(interface{ Flush() }); ok {
					f.Flush()
				}
				result.Error = scanner.Err()
				return result
			}
		}
	}

	result.Error = scanner.Err()
	return result
}

func (t *OpenAIToAnthropicTransformer) writeChunk(w io.Writer, chunk StreamChunk) {
	data, _ := json.Marshal(chunk)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(interface{ Flush() }); ok {
		f.Flush()
	}
}
