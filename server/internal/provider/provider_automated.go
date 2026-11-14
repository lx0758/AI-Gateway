package provider

import (
	"ai-gateway/internal/model"
	"fmt"

	"github.com/gin-gonic/gin"
)

type AutomatedProvider struct {
	cfg               *Config
	openAIProvider    Provider
	anthropicProvider Provider
	responsesProvider Provider
}

func NewAutomatedProvider(openAIChatCompletionsBaseURL string, anthropicMessagesBaseURL string, openAIResponsesBaseURL string, apiKey string) *AutomatedProvider {
	cfg := &Config{
		OpenAIChatCompletionsBaseURL: openAIChatCompletionsBaseURL,
		AnthropicMessagesBaseURL:     anthropicMessagesBaseURL,
		OpenAIResponsesBaseURL:       openAIResponsesBaseURL,
		APIKey:                       apiKey,
	}
	p := AutomatedProvider{cfg: cfg}
	if cfg.OpenAIChatCompletionsBaseURL != "" {
		p.openAIProvider = NewOpenAIProvider(cfg)
	}
	if cfg.AnthropicMessagesBaseURL != "" {
		p.anthropicProvider = NewAnthropicProvider(cfg)
	}
	if cfg.OpenAIResponsesBaseURL != "" {
		p.responsesProvider = NewResponsesProvider(cfg)
	}
	return &p
}

func (p *AutomatedProvider) SyncModels(providerID uint) ([]model.ProviderModel, error) {
	if p.openAIProvider != nil {
		if models, err := p.openAIProvider.SyncModels(providerID); err == nil && models != nil {
			return models, err
		}
	}
	if p.anthropicProvider != nil {
		if models, err := p.anthropicProvider.SyncModels(providerID); err == nil && models != nil {
			return models, err
		}
	}
	if p.responsesProvider != nil {
		if models, err := p.responsesProvider.SyncModels(providerID); err == nil && models != nil {
			return models, err
		}
	}
	return nil, fmt.Errorf("no valid models found")
}

func (p *AutomatedProvider) ExecuteOpenAIChatCompletionsRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	// 原生优先，必须转换时按 Chat > Anthropic > Responses 顺位选择转换目标
	finalProvider := p.openAIProvider
	if finalProvider == nil {
		finalProvider = p.anthropicProvider
	}
	if finalProvider == nil {
		finalProvider = p.responsesProvider
	}
	if finalProvider == nil {
		return fmt.Errorf("provider has no available endpoint")
	}
	return finalProvider.ExecuteOpenAIChatCompletionsRequest(ctx, pm, usage)
}

func (p *AutomatedProvider) ExecuteAnthropicMessagesRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	finalProvider := p.anthropicProvider
	if finalProvider == nil {
		finalProvider = p.openAIProvider
	}
	if finalProvider == nil {
		finalProvider = p.responsesProvider
	}
	if finalProvider == nil {
		return fmt.Errorf("provider has no available endpoint")
	}
	return finalProvider.ExecuteAnthropicMessagesRequest(ctx, pm, usage)
}

func (p *AutomatedProvider) ExecuteOpenAIResponsesRequest(ctx *gin.Context, pm *model.ProviderModel, usage *Usage) error {
	finalProvider := p.responsesProvider
	if finalProvider == nil {
		finalProvider = p.openAIProvider
	}
	if finalProvider == nil {
		finalProvider = p.anthropicProvider
	}
	if finalProvider == nil {
		return fmt.Errorf("provider has no available endpoint")
	}
	return finalProvider.ExecuteOpenAIResponsesRequest(ctx, pm, usage)
}
