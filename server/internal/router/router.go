package router

import (
	"ai-model-proxy/internal/model"
)

type RouteResult struct {
	Provider      *model.Provider
	ProviderModel *model.ProviderModel
	ActualModel   string
}

type ModelRouter struct{}

func NewModelRouter() *ModelRouter {
	return &ModelRouter{}
}

func (r *ModelRouter) Route(alias string) (*RouteResult, error) {
	var mappings []model.ModelMapping
	if err := model.DB.Preload("Provider").Preload("ProviderModel").
		Where("alias = ? AND enabled = ?", alias, true).
		Order("weight DESC").
		Find(&mappings).Error; err != nil {
		return nil, err
	}

	if len(mappings) == 0 {
		return nil, nil
	}

	for _, m := range mappings {
		if !m.Provider.Enabled || !m.ProviderModel.IsAvailable {
			continue
		}

		return &RouteResult{
			Provider:      m.Provider,
			ProviderModel: m.ProviderModel,
			ActualModel:   m.ProviderModel.ModelID,
		}, nil
	}

	return nil, nil
}
