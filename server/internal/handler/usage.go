package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
)

type UsageHandler struct{}

func NewUsageHandler() *UsageHandler {
	return &UsageHandler{}
}

func (h *UsageHandler) Stats(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	var totalRequests int64
	var successRequests int64
	var totalTokens int64
	var totalPromptTokens int64
	var totalCompletionTokens int64

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate+" 23:59:59").
		Count(&totalRequests)

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ? AND status = ?", startDate, endDate+" 23:59:59", "success").
		Count(&successRequests)

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate+" 23:59:59").
		Select("COALESCE(SUM(prompt_tokens + completion_tokens), 0)").Scan(&totalTokens)

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate+" 23:59:59").
		Select("COALESCE(SUM(prompt_tokens), 0)").Scan(&totalPromptTokens)

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate+" 23:59:59").
		Select("COALESCE(SUM(completion_tokens), 0)").Scan(&totalCompletionTokens)

	successRate := float64(0)
	if totalRequests > 0 {
		successRate = float64(successRequests) / float64(totalRequests) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"totalRequests":    totalRequests,
		"successRequests":  successRequests,
		"successRate":      successRate,
		"totalTokens":      totalTokens,
		"promptTokens":     totalPromptTokens,
		"completionTokens": totalCompletionTokens,
	})
}

func (h *UsageHandler) Logs(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -1).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	apiKeyID := c.Query("api_key_id")
	modelName := c.Query("model")

	query := model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate+" 23:59:59")

	if apiKeyID != "" {
		query = query.Where("api_key_id = ?", apiKeyID)
	}

	if modelName != "" {
		query = query.Where("model = ?", modelName)
	}

	var logs []model.UsageLog
	if err := query.Order("created_at DESC").Limit(1000).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func (h *UsageHandler) Dashboard(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	var todayRequests int64
	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ?", today).
		Count(&todayRequests)

	var activeProviders int64
	model.DB.Model(&model.Provider{}).Where("enabled = ?", true).Count(&activeProviders)

	var activeKeys int64
	model.DB.Model(&model.APIKey{}).Where("enabled = ?", true).Count(&activeKeys)

	var dailyStats []struct {
		Date    string `json:"date"`
		Count   int64  `json:"count"`
		Success int64  `json:"success"`
	}
	model.DB.Raw(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as count,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success
		FROM usage_logs
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, sevenDaysAgo).Scan(&dailyStats)

	var providerStats []struct {
		Provider string `json:"provider"`
		Count    int64  `json:"count"`
	}
	model.DB.Raw(`
		SELECT p.name as provider, COUNT(*) as count
		FROM usage_logs ul
		JOIN providers p ON ul.provider_id = p.id
		WHERE ul.created_at >= ?
		GROUP BY p.name
		ORDER BY count DESC
	`, sevenDaysAgo).Scan(&providerStats)

	var modelStats []struct {
		Model string `json:"model"`
		Count int64  `json:"count"`
	}
	model.DB.Raw(`
		SELECT model, COUNT(*) as count
		FROM usage_logs
		WHERE created_at >= ?
		GROUP BY model
		ORDER BY count DESC
		LIMIT 10
	`, sevenDaysAgo).Scan(&modelStats)

	c.JSON(http.StatusOK, gin.H{
		"todayRequests":   todayRequests,
		"activeProviders": activeProviders,
		"activeKeys":      activeKeys,
		"dailyStats":      dailyStats,
		"providerStats":   providerStats,
		"modelStats":      modelStats,
	})
}
