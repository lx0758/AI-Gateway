package provider

import (
	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type Config struct {
	BaseURL string
	APIKey  string
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
	ExecuteOpenAIRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error
	ExecuteAnthropicRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error
}
