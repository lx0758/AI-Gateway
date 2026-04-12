package router

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/provider"
)

type RouteResult struct {
	Provider         *model.Provider
	ProviderModel    *model.ProviderModel
	ProviderInstance provider.Provider
}

func (r *RouteResult) SupportOpenAI() bool {
	return r.Provider.OpenAIBaseURL != ""
}

func (r *RouteResult) SupportAnthropic() bool {
	return r.Provider.AnthropicBaseURL != ""
}
