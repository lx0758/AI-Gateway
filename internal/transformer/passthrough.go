package transformer

import (
	"encoding/json"
	"io"
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

func (t *PassThroughTransformer) TransformStream(reader io.Reader, writer io.Writer) error {
	_, err := io.Copy(writer, reader)
	return err
}
