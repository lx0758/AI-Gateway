package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
	"ai-gateway/internal/router"
)

type ProviderHandler struct{}

type createProviderRequest struct {
	Name                          string `json:"name" binding:"required"`
	OpenAIChatCompletionsBaseURL  string `json:"openai_chat_completions_base_url"`
	AnthropicMessagesBaseURL      string `json:"anthropic_messages_base_url"`
	OpenAIResponsesBaseURL        string `json:"openai_responses_base_url"`
	APIKey                        string `json:"api_key" binding:"required"`
	Homepage                      string `json:"homepage"`
	BillingType                   string `json:"billing_type"`
	Enabled                       *bool  `json:"enabled"`
	Priority                      int    `json:"priority"`
}

type updateProviderRequest struct {
	Name                          string  `json:"name"`
	OpenAIChatCompletionsBaseURL  *string `json:"openai_chat_completions_base_url"`
	AnthropicMessagesBaseURL      *string `json:"anthropic_messages_base_url"`
	OpenAIResponsesBaseURL        *string `json:"openai_responses_base_url"`
	APIKey                        string  `json:"api_key"`
	Homepage                      *string `json:"homepage"`
	BillingType                   *string `json:"billing_type"`
	Enabled                       *bool   `json:"enabled"`
	Priority                      *int    `json:"priority"`
}

// 付费类型值域：空串 = 未设置。binding tag 的 oneof 对指针空串与未传字段无法两全，
// 故在 handler 集中校验（见 validateBillingType）
var validBillingTypes = map[string]bool{
	"":             true,
	"subscription": true,
	"usage":        true,
	"free":         true,
}

func validateBillingType(billingType string) bool {
	return validBillingTypes[billingType]
}

type providerResponse struct {
	ID                           uint                    `json:"id"`
	Name                         string                  `json:"name"`
	OpenAIChatCompletionsBaseURL string                  `json:"openai_chat_completions_base_url"`
	AnthropicMessagesBaseURL     string                  `json:"anthropic_messages_base_url"`
	OpenAIResponsesBaseURL       string                  `json:"openai_responses_base_url"`
	APIKeyMasked                 string                  `json:"api_key_masked"`
	Homepage                     string                  `json:"homepage"`
	BillingType                  string                  `json:"billing_type"`
	Enabled                      bool                    `json:"enabled"`
	Priority                     int                     `json:"priority"`
	Models                       []providerModelResponse `json:"models,omitempty"`
	CreatedAt                    string                  `json:"created_at"`
}

func NewProviderHandler() *ProviderHandler {
	return &ProviderHandler{}
}

func (h *ProviderHandler) List(c *gin.Context) {
	var providers []model.Provider
	if err := model.DB.Preload("Models").Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]providerResponse, len(providers))
	for i, p := range providers {
		models := make([]providerModelResponse, len(p.Models))
		for j, m := range p.Models {
			models[j] = toProviderModelResponse(m)
		}

		result[i] = providerResponse{
			ID:                           p.ID,
			Name:                         p.Name,
			OpenAIChatCompletionsBaseURL: p.OpenAIChatCompletionsBaseURL,
			AnthropicMessagesBaseURL:     p.AnthropicMessagesBaseURL,
			OpenAIResponsesBaseURL:       p.OpenAIResponsesBaseURL,
			APIKeyMasked:                 maskAPIKey(p.APIKey),
			Homepage:                     p.Homepage,
			BillingType:                  p.BillingType,
			Enabled:                      p.Enabled,
			Priority:                     p.Priority,
			Models:                       models,
			CreatedAt:                    p.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"providers": result})
}

func (h *ProviderHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var provider model.Provider
	if err := model.DB.Preload("Models").First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	models := make([]providerModelResponse, len(provider.Models))
	for j, m := range provider.Models {
		models[j] = toProviderModelResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"provider": providerResponse{
		ID:                           provider.ID,
		Name:                         provider.Name,
		OpenAIChatCompletionsBaseURL: provider.OpenAIChatCompletionsBaseURL,
		AnthropicMessagesBaseURL:     provider.AnthropicMessagesBaseURL,
		OpenAIResponsesBaseURL:       provider.OpenAIResponsesBaseURL,
		APIKeyMasked:                 maskAPIKey(provider.APIKey),
		Homepage:                     provider.Homepage,
		BillingType:                  provider.BillingType,
		Enabled:                      provider.Enabled,
		Priority:                     provider.Priority,
		Models:                       models,
		CreatedAt:                    provider.CreatedAt.Format("2006-01-02 15:04:05"),
	}})
}

func (h *ProviderHandler) Create(c *gin.Context) {
	var req createProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validateBillingType(req.BillingType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid billing_type: must be one of subscription/usage/free"})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	provider := model.Provider{
		Name:                         req.Name,
		OpenAIChatCompletionsBaseURL: strings.TrimSuffix(req.OpenAIChatCompletionsBaseURL, "/"),
		AnthropicMessagesBaseURL:     strings.TrimSuffix(req.AnthropicMessagesBaseURL, "/"),
		OpenAIResponsesBaseURL:       strings.TrimSuffix(req.OpenAIResponsesBaseURL, "/"),
		APIKey:                       req.APIKey,
		Homepage:                     req.Homepage,
		BillingType:                  req.BillingType,
		Enabled:                      enabled,
		Priority:                     req.Priority,
	}

	if provider.OpenAIChatCompletionsBaseURL == "" && provider.AnthropicMessagesBaseURL == "" && provider.OpenAIResponsesBaseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one base URL is required"})
		return
	}

	if err := model.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// GORM 会跳过带 default 标签字段的零值插入（落库走 DB 默认 true），禁用状态需显式回写
	if !enabled {
		model.DB.Model(&provider).Update("enabled", false)
	}

	c.JSON(http.StatusCreated, gin.H{"provider": providerResponse{
		ID:                           provider.ID,
		Name:                         provider.Name,
		OpenAIChatCompletionsBaseURL: provider.OpenAIChatCompletionsBaseURL,
		AnthropicMessagesBaseURL:     provider.AnthropicMessagesBaseURL,
		OpenAIResponsesBaseURL:       provider.OpenAIResponsesBaseURL,
		APIKeyMasked:                 maskAPIKey(provider.APIKey),
		Homepage:                     provider.Homepage,
		BillingType:                  provider.BillingType,
		Enabled:                      provider.Enabled,
		Priority:                     provider.Priority,
		CreatedAt:                    provider.CreatedAt.Format("2006-01-02 15:04:05"),
	}})
}

func (h *ProviderHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var provider model.Provider
	if err := model.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	var req updateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.BillingType != nil && !validateBillingType(*req.BillingType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid billing_type: must be one of subscription/usage/free"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}

	// 允许更新 BaseURL（包括清空）
	if req.OpenAIChatCompletionsBaseURL != nil {
		updates["openai_chat_completions_base_url"] = strings.TrimSuffix(*req.OpenAIChatCompletionsBaseURL, "/")
	}
	if req.AnthropicMessagesBaseURL != nil {
		updates["anthropic_messages_base_url"] = strings.TrimSuffix(*req.AnthropicMessagesBaseURL, "/")
	}
	if req.OpenAIResponsesBaseURL != nil {
		updates["openai_responses_base_url"] = strings.TrimSuffix(*req.OpenAIResponsesBaseURL, "/")
	}

	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	// 元信息：指针语义，未传不覆盖；显式传空串则清空（编辑表单支持清除）
	if req.Homepage != nil {
		updates["homepage"] = *req.Homepage
	}
	if req.BillingType != nil {
		updates["billing_type"] = *req.BillingType
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	if req.OpenAIChatCompletionsBaseURL != nil || req.AnthropicMessagesBaseURL != nil || req.OpenAIResponsesBaseURL != nil || req.APIKey != "" || req.Enabled != nil {
		router.ClearAllCooldownsForProvider(provider.ID)
	}

	// 验证更新后至少有一个 BaseURL
	newOpenAIChatCompletionsBaseURL := provider.OpenAIChatCompletionsBaseURL
	newAnthropicMessagesBaseURL := provider.AnthropicMessagesBaseURL
	newOpenAIResponsesBaseURL := provider.OpenAIResponsesBaseURL
	if openaiURL, ok := updates["openai_chat_completions_base_url"].(string); ok {
		newOpenAIChatCompletionsBaseURL = openaiURL
	}
	if anthropicURL, ok := updates["anthropic_messages_base_url"].(string); ok {
		newAnthropicMessagesBaseURL = anthropicURL
	}
	if responsesURL, ok := updates["openai_responses_base_url"].(string); ok {
		newOpenAIResponsesBaseURL = responsesURL
	}
	if newOpenAIChatCompletionsBaseURL == "" && newAnthropicMessagesBaseURL == "" && newOpenAIResponsesBaseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one base URL is required"})
		return
	}

	if err := model.DB.Model(&provider).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.Preload("Models").First(&provider, id)

	models := make([]providerModelResponse, len(provider.Models))
	for j, m := range provider.Models {
		models[j] = toProviderModelResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{"provider": providerResponse{
		ID:                           provider.ID,
		Name:                         provider.Name,
		OpenAIChatCompletionsBaseURL: provider.OpenAIChatCompletionsBaseURL,
		AnthropicMessagesBaseURL:     provider.AnthropicMessagesBaseURL,
		OpenAIResponsesBaseURL:       provider.OpenAIResponsesBaseURL,
		APIKeyMasked:                 maskAPIKey(provider.APIKey),
		Homepage:                     provider.Homepage,
		BillingType:                  provider.BillingType,
		Enabled:                      provider.Enabled,
		Priority:                     provider.Priority,
		Models:                       models,
		CreatedAt:                    provider.CreatedAt.Format("2006-01-02 15:04:05"),
	}})
}

func (h *ProviderHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 查询所有 ProviderModel IDs
	var providerModelIDs []uint
	model.DB.Model(&model.ProviderModel{}).Where("provider_id = ?", id).Pluck("id", &providerModelIDs)

	// 硬删除关联的 ModelMapping
	if len(providerModelIDs) > 0 {
		if err := model.DB.Where("provider_model_id IN ?", providerModelIDs).Delete(&model.ModelMapping{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 硬删除关联的 ProviderModel
	if err := model.DB.Where("provider_id = ?", id).Delete(&model.ProviderModel{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 软删除 Provider
	if err := model.DB.Delete(&model.Provider{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	router.ClearAllCooldownsForProvider(uint(id))

	c.JSON(http.StatusOK, gin.H{"message": "provider deleted"})
}

func (h *ProviderHandler) Test(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var provider model.Provider
	if err := model.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "connection test not implemented yet"})
}

func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 4 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
