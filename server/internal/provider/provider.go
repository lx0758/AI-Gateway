package provider

import (
	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
)

type Config struct {
	OpenAIChatCompletionsBaseURL string
	AnthropicMessagesBaseURL     string
	OpenAIResponsesBaseURL       string
	APIKey                       string
}

type Usage struct {
	CachedTokens int
	InputTokens  int
	OutputTokens int
}

func (u Usage) TotalTokens() int {
	return u.CachedTokens + u.InputTokens + u.OutputTokens
}

type Provider interface {
	SyncModels(providerID uint) ([]model.ProviderModel, error)
	ExecuteOpenAIChatCompletionsRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error
	ExecuteAnthropicMessagesRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error
	ExecuteOpenAIResponsesRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error
}
