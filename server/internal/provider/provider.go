package provider

import (
	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type Config struct {
	BaseURL string
	APIKey  string
}

type Provider interface {
	SyncModels(providerID uint) ([]model.ProviderModel, error)
	ExecuteOpenAIRequest(ctx *gin.Context, pm *model.ProviderModel) (int, error)
	ExecuteAnthropicRequest(ctx *gin.Context, pm *model.ProviderModel) (int, error)
}
