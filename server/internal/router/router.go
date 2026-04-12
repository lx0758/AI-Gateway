package router

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/provider"
)

var globalRouter = &ModelRouter{
	cooldownManager: NewCooldownManager(),
}

func GetRouter() *ModelRouter {
	return globalRouter
}

type ModelRouter struct {
	cooldownManager *CooldownManager
}

func (r *ModelRouter) Route(name string) (*RouteResult, error) {
	r.cooldownManager.ClearExpiredCooldowns()

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

	var allProviders []RouteResult
	var availableProviders []RouteResult

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
		result := RouteResult{
			Provider:         providerInfo,
			ProviderModel:    &pm,
			ProviderInstance: providerImpl,
		}
		allProviders = append(allProviders, result)

		if !r.cooldownManager.IsCooldown(providerInfo.ID, pm.ID) {
			availableProviders = append(availableProviders, result)
		}
	}

	if len(availableProviders) > 0 {
		return &availableProviders[0], nil
	}

	if len(allProviders) > 0 {
		earliest := r.cooldownManager.GetEarliestCooldownEnd(allProviders)
		if earliest != nil {
			return earliest, nil
		}
		return &allProviders[0], nil
	}

	return nil, nil
}

func (r *ModelRouter) RecordRateLimit(providerID uint, providerModelID uint) {
	r.cooldownManager.Record429(providerID, providerModelID)
}

func (r *ModelRouter) RecordSuccess(providerID uint, providerModelID uint) {
	r.cooldownManager.RecordSuccess(providerID, providerModelID)
}

func ClearCooldown(providerID uint, providerModelID uint) {
	globalRouter.cooldownManager.ClearCooldown(providerID, providerModelID)
}

func ClearAllCooldownsForProvider(providerID uint) {
	globalRouter.cooldownManager.ClearAllForProvider(providerID)
}
