package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type APIKeyHandler struct{}

type CreateAPIKeyRequest struct {
	Name          string     `json:"name"`
	AllowedModels []string   `json:"allowed_models"`
	RateLimit     int        `json:"rate_limit"`
	Quota         int64      `json:"quota"`
	ExpiresAt     *time.Time `json:"expires_at"`
}

type UpdateAPIKeyRequest struct {
	Name      *string    `json:"name"`
	RateLimit *int       `json:"rate_limit"`
	Quota     *int64     `json:"quota"`
	ExpiresAt *time.Time `json:"expires_at"`
	Enabled   *bool      `json:"enabled"`
}

func NewAPIKeyHandler() *APIKeyHandler {
	return &APIKeyHandler{}
}

func generateAPIKey() string {
	bytes := make([]byte, 24)
	rand.Read(bytes)
	return "sk-" + hex.EncodeToString(bytes)
}

func (h *APIKeyHandler) List(c *gin.Context) {
	var keys []model.Key
	if err := model.DB.Preload("Models").Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type KeyWithStats struct {
		ID          uint             `json:"id"`
		Key         string           `json:"key"`
		Name        string           `json:"name"`
		Enabled     bool             `json:"enabled"`
		Quota       int64            `json:"quota"`
		UsedQuota   int64            `json:"used_quota"`
		UsedCount   int64            `json:"used_count"`
		Models      []model.KeyModel `json:"models,omitempty"`
		TotalTokens int64            `json:"total_tokens"`
		AvgLatency  float64          `json:"avg_latency"`
		ExpiresAt   *time.Time       `json:"expires_at"`
		CreatedAt   time.Time        `json:"created_at"`
	}

	result := make([]KeyWithStats, len(keys))
	for i, k := range keys {
		maskedKey := k.Key
		if len(maskedKey) > 8 {
			maskedKey = maskedKey[:8] + "****" + maskedKey[len(maskedKey)-4:]
		}

		result[i] = KeyWithStats{
			ID:        k.ID,
			Key:       maskedKey,
			Name:      k.Name,
			Enabled:   k.Enabled,
			Quota:     k.Quota,
			UsedQuota: k.UsedQuota,
			UsedCount: k.UsedCount,
			Models:    k.Models,
			ExpiresAt: k.ExpiresAt,
			CreatedAt: k.CreatedAt,
		}

		var stats struct {
			TotalTokens int64
			AvgLatency  float64
			CallCount   int64
		}
		model.DB.Model(&model.UsageLog{}).
			Where("key_id = ?", k.ID).
			Select("COALESCE(SUM(total_tokens), 0) as total_tokens, COALESCE(AVG(latency_ms), 0) as avg_latency, COUNT(*) as call_count").
			Scan(&stats)

		result[i].TotalTokens = stats.TotalTokens
		result[i].AvgLatency = stats.AvgLatency
		if stats.CallCount > 0 {
			result[i].UsedCount = stats.CallCount
		}
	}

	c.JSON(http.StatusOK, gin.H{"keys": result})
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := model.Key{
		Key:       generateAPIKey(),
		Name:      req.Name,
		RateLimit: req.RateLimit,
		Quota:     req.Quota,
		ExpiresAt: req.ExpiresAt,
		Enabled:   true,
	}

	if len(req.AllowedModels) > 0 {
		allowedBytes, _ := json.Marshal(req.AllowedModels)
		key.AllowedModels = string(allowedBytes)
	}

	if err := model.DB.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, alias := range req.AllowedModels {
		akm := model.KeyModel{
			KeyID:      key.ID,
			ModelAlias: alias,
		}
		model.DB.Create(&akm)
	}

	c.JSON(http.StatusCreated, gin.H{
		"key":     key,
		"raw_key": key.Key,
	})
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	model.DB.Where("key_id = ?", id).Delete(&model.KeyModel{})

	if err := model.DB.Delete(&model.Key{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}

func (h *APIKeyHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var key model.Key
	if err := model.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var req UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.RateLimit != nil {
		updates["rate_limit"] = *req.RateLimit
	}
	if req.Quota != nil {
		updates["quota"] = *req.Quota
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = *req.ExpiresAt
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := model.DB.Model(&key).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	model.DB.Preload("Models").First(&key, id)
	c.JSON(http.StatusOK, gin.H{"key": key})
}

func (h *APIKeyHandler) AddModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		ModelAlias string `json:"model_alias" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing model.KeyModel
	if err := model.DB.Where("key_id = ? AND model_alias = ?", id, req.ModelAlias).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "model already added"})
		return
	}

	akm := model.KeyModel{
		KeyID:      uint(id),
		ModelAlias: req.ModelAlias,
	}
	if err := model.DB.Create(&akm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"model": akm})
}

func (h *APIKeyHandler) RemoveModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	modelAlias := c.Param("model_alias")
	if modelAlias == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model_alias required"})
		return
	}

	if err := model.DB.Where("key_id = ? AND model_alias = ?", id, modelAlias).Delete(&model.KeyModel{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "model removed"})
}

func (h *APIKeyHandler) ListModels(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var models []model.KeyModel
	if err := model.DB.Where("key_id = ?", id).Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}
