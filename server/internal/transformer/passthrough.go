package transformer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type PassThroughTransformer struct{}

func NewPassThroughTransformer() *PassThroughTransformer {
	return &PassThroughTransformer{}
}

func (t *PassThroughTransformer) TransformRequest(req *OpenAIRequest) (interface{}, error) {
	return req, nil
}

func (t *PassThroughTransformer) TransformResponse(body []byte) (*OpenAIResponse, error) {
	var resp OpenAIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *PassThroughTransformer) TransformStream(reader io.Reader, writer io.Writer) *StreamResult {
	result := &StreamResult{}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(writer, line)
		if f, ok := writer.(interface{ Flush() }); ok {
			f.Flush()
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if usage, ok := chunk["usage"].(map[string]interface{}); ok {
				result.Usage = &Usage{}
				if pt, ok := usage["prompt_tokens"].(float64); ok {
					result.Usage.PromptTokens = int(pt)
				}
				if ct, ok := usage["completion_tokens"].(float64); ok {
					result.Usage.CompletionTokens = int(ct)
				}
				if tt, ok := usage["total_tokens"].(float64); ok {
					result.Usage.TotalTokens = int(tt)
				}
			}
		}
	}

	result.Error = scanner.Err()
	return result
}
