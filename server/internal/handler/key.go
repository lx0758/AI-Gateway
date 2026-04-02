package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
)

type APIKeyHandler struct{}

type createAPIKeyRequest struct {
	Name      string   `json:"name" binding:"required"`
	Models    []string `json:"models"`
	ExpiresAt *string  `json:"expires_at"`
}

type updateAPIKeyRequest struct {
	Name      *string  `json:"name"`
	Models    []string `json:"models"`
	ExpiresAt *string  `json:"expires_at"`
	Enabled   *bool    `json:"enabled"`
}

type keyModelResponse struct {
	ID    uint   `json:"id"`
	Model string `json:"model"`
}

type apiKeyResponse struct {
	ID        uint               `json:"id"`
	Key       string             `json:"key"`
	Name      string             `json:"name"`
	Enabled   bool               `json:"enabled"`
	ExpiresAt *time.Time         `json:"expires_at"`
	CreatedAt time.Time          `json:"created_at"`
	Models    []keyModelResponse `json:"models,omitempty"`
}

type apiKeyCreateResponse struct {
	Key    apiKeyResponse `json:"key"`
	RawKey string         `json:"raw_key"`
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

	result := make([]apiKeyResponse, len(keys))
	for i, k := range keys {
		maskedKey := k.Key
		if len(maskedKey) > 8 {
			maskedKey = maskedKey[:8] + "****" + maskedKey[len(maskedKey)-4:]
		}

		models := make([]keyModelResponse, len(k.Models))
		for j, m := range k.Models {
			models[j] = keyModelResponse{
				ID:    m.ID,
				Model: m.Model,
			}
		}

		result[i] = apiKeyResponse{
			ID:        k.ID,
			Key:       maskedKey,
			Name:      k.Name,
			Enabled:   k.Enabled,
			ExpiresAt: k.ExpiresAt,
			CreatedAt: k.CreatedAt,
			Models:    models,
		}
	}

	c.JSON(http.StatusOK, gin.H{"keys": result})
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse("2006-01-02 15:04:05", *req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	key := model.Key{
		Key:       generateAPIKey(),
		Name:      req.Name,
		ExpiresAt: expiresAt,
		Enabled:   true,
	}

	if err := model.DB.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, alias := range req.Models {
		akm := model.KeyModel{
			KeyID: key.ID,
			Model: alias,
		}
		model.DB.Create(&akm)
	}

	model.DB.Preload("Models").First(&key, key.ID)

	models := make([]keyModelResponse, len(key.Models))
	for j, m := range key.Models {
		models[j] = keyModelResponse{
			ID:    m.ID,
			Model: m.Model,
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"key": apiKeyResponse{
			ID:        key.ID,
			Key:       key.Key,
			Name:      key.Name,
			Enabled:   key.Enabled,
			ExpiresAt: key.ExpiresAt,
			CreatedAt: key.CreatedAt,
			Models:    models,
		},
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

	var req updateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ExpiresAt != nil {
		if t, err := time.Parse("2006-01-02 15:04:05", *req.ExpiresAt); err == nil {
			updates["expires_at"] = t
		}
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

	if req.Models != nil {
		model.DB.Where("key_id = ?", key.ID).Delete(&model.KeyModel{})
		for _, alias := range req.Models {
			akm := model.KeyModel{KeyID: key.ID, Model: alias}
			model.DB.Create(&akm)
		}
	}

	model.DB.Preload("Models").First(&key, id)

	models := make([]keyModelResponse, len(key.Models))
	for j, m := range key.Models {
		models[j] = keyModelResponse{
			ID:    m.ID,
			Model: m.Model,
		}
	}

	c.JSON(http.StatusOK, gin.H{"key": apiKeyResponse{
		ID:        key.ID,
		Key:       key.Key,
		Name:      key.Name,
		Enabled:   key.Enabled,
		ExpiresAt: key.ExpiresAt,
		CreatedAt: key.CreatedAt,
		Models:    models,
	}})
}

func (h *APIKeyHandler) AddModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		ModelAlias string `binding:"required"`
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
		KeyID: uint(id),
		Model: req.ModelAlias,
	}
	if err := model.DB.Create(&akm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"model": keyModelResponse{
		ID:    akm.ID,
		Model: akm.Model,
	}})
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

	result := make([]keyModelResponse, len(models))
	for i, m := range models {
		result[i] = keyModelResponse{
			ID:    m.ID,
			Model: m.Model,
		}
	}

	c.JSON(http.StatusOK, gin.H{"models": result})
}

func (h *APIKeyHandler) Reset(c *gin.Context) {
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

	newKey := generateAPIKey()

	if err := model.DB.Model(&key).Update("key", newKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.Preload("Models").First(&key, id)

	models := make([]keyModelResponse, len(key.Models))
	for j, m := range key.Models {
		models[j] = keyModelResponse{
			ID:    m.ID,
			Model: m.Model,
		}
	}

	maskedKey := key.Key
	if len(maskedKey) > 8 {
		maskedKey = maskedKey[:8] + "****" + maskedKey[len(maskedKey)-4:]
	}

	c.JSON(http.StatusOK, gin.H{
		"key": apiKeyResponse{
			ID:        key.ID,
			Key:       maskedKey,
			Name:      key.Name,
			Enabled:   key.Enabled,
			ExpiresAt: key.ExpiresAt,
			CreatedAt: key.CreatedAt,
			Models:    models,
		},
		"raw_key": key.Key,
	})
}
