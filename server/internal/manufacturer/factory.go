package manufacturer

import (
	"ai-proxy/internal/model"
)

const (
	ProviderTypeOpenAI    = "openai"
	ProviderTypeAnthropic = "anthropic"
)

type Config struct {
	ProviderName string
	ProviderType string
	BaseURL      string
	APIKey       string
}

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Create(provider *model.Provider) Manufacturer {
	cfg := &Config{
		ProviderName: provider.Name,
		ProviderType: provider.APIType,
		BaseURL:      provider.BaseURL,
		APIKey:       provider.APIKey,
	}

	switch provider.APIType {
	case ProviderTypeOpenAI:
		return NewOpenAICompatibleManufacturer(cfg)
	case ProviderTypeAnthropic:
		return NewAnthropicManufacturer(cfg)
	default:
		return nil
	}
}
