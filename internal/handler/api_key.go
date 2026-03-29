package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/user/ai-model-proxy/internal/model"
)

type APIKeyHandler struct{}

type CreateAPIKeyRequest struct {
	Name          string     `json:"name"`
	AllowedModels []string   `json:"allowed_models"`
	RateLimit     int        `json:"rate_limit"`
	Quota         int64      `json:"quota"`
	ExpiresAt     *time.Time `json:"expires_at"`
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
	if err := model.DB.Find(&keys).Error; err != nil {
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

	if err := model.DB.Delete(&model.APIKey{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}
