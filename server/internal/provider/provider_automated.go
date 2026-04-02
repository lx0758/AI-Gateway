package provider

import (
	"ai-proxy/internal/model"
	"fmt"

	"github.com/gin-gonic/gin"
)

type AutomatedProvider struct {
	openAIProvider    Provider
	anthropicProvider Provider
}

func NewAutomatedProvider(openAIBaseURL string, anthropicBaseUrl string, apiKey string) *AutomatedProvider {
	p := AutomatedProvider{}
	if openAIBaseURL != "" {
		p.openAIProvider = NewOpenAIProvider(&Config{
			APIKey:  apiKey,
			BaseURL: openAIBaseURL,
		})
	}
	if anthropicBaseUrl != "" {
		p.anthropicProvider = NewAnthropicProvider(&Config{
			APIKey:  apiKey,
			BaseURL: anthropicBaseUrl,
		})
	}
	return &p
}

func (p *AutomatedProvider) SyncModels(providerID uint) ([]model.ProviderModel, error) {
	if p.anthropicProvider != nil {
		if models, err := p.anthropicProvider.SyncModels(providerID); err == nil && models != nil {
			return models, err
		}
	}
	if p.openAIProvider != nil {
		if models, err := p.openAIProvider.SyncModels(providerID); err == nil && models != nil {
			return models, err
		}
	}
	return nil, fmt.Errorf("no valid models found")
}

func (p *AutomatedProvider) ExecuteOpenAIRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	finialProvider := p.openAIProvider
	if finialProvider == nil {
		finialProvider = p.anthropicProvider
	}
	return finialProvider.ExecuteOpenAIRequest(ctx, pm, usage)
}

func (p *AutomatedProvider) ExecuteAnthropicRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	finialProvider := p.anthropicProvider
	if finialProvider == nil {
		finialProvider = p.openAIProvider
	}
	return finialProvider.ExecuteAnthropicRequest(ctx, pm, usage)
}
