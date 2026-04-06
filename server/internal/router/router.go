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

type ModelRouter struct{}

func NewModelRouter() *ModelRouter {
	return &ModelRouter{}
}

func (r *ModelRouter) Route(name string) ([]RouteResult, error) {
	var m model.Model
	if err := model.DB.Where("name = ? AND enabled = ?", name, true).First(&m).Error; err != nil {
		return nil, nil
	}

	var mappings []model.ModelMapping
	if err := model.DB.Preload("Provider").
		Where("model_id = ? AND enabled = ?", m.ID, true).
		Order("weight DESC").
		Find(&mappings).Error; err != nil {
		return nil, err
	}

	if len(mappings) == 0 {
		return nil, nil
	}

	var providers []RouteResult

	for _, mapping := range mappings {
		providerInfo := mapping.Provider
		if !providerInfo.Enabled {
			continue
		}

		var pm model.ProviderModel
		if err := model.DB.Where("provider_id = ? AND model_id = ? AND is_available = ?", mapping.ProviderID, mapping.ProviderModelName, true).First(&pm).Error; err != nil {
			continue
		}

		providerImpl := provider.NewAutomatedProvider(
			providerInfo.OpenAIBaseURL,
			providerInfo.AnthropicBaseURL,
			providerInfo.APIKey,
		)
		providers = append(providers, RouteResult{
			Provider:         providerInfo,
			ProviderModel:    &pm,
			ProviderInstance: providerImpl,
		})
	}

	return providers, nil
}
