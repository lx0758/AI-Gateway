package manufacturer

import (
	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type Manufacturer interface {
	Name() string
	SyncModels(provider *model.Provider) ([]model.ProviderModel, error)
	ExecuteOpenAIRequest(ctx *gin.Context, model *model.ProviderModel) (int, error)
	ExecuteAnthropicRequest(ctx *gin.Context, model *model.ProviderModel) (int, error)
}
