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
	var keys []model.APIKey
	if err := model.DB.Preload("Models").Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range keys {
		if len(keys[i].Key) > 8 {
			keys[i].Key = keys[i].Key[:8] + "****" + keys[i].Key[len(keys[i].Key)-4:]
		}
	}

	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := model.APIKey{
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
		akm := model.APIKeyModel{
			APIKeyID:   key.ID,
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

	model.DB.Where("api_key_id = ?", id).Delete(&model.APIKeyModel{})

	if err := model.DB.Delete(&model.APIKey{}, id).Error; err != nil {
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

	var key model.APIKey
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

	var existing model.APIKeyModel
	if err := model.DB.Where("api_key_id = ? AND model_alias = ?", id, req.ModelAlias).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "model already added"})
		return
	}

	akm := model.APIKeyModel{
		APIKeyID:   uint(id),
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

	if err := model.DB.Where("api_key_id = ? AND model_alias = ?", id, modelAlias).Delete(&model.APIKeyModel{}).Error; err != nil {
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

	var models []model.APIKeyModel
	if err := model.DB.Where("api_key_id = ?", id).Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}
