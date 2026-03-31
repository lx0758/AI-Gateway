package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/model"
	"ai-proxy/internal/router"
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
	var avgLatency float64

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate+" 23:59:59").
		Count(&totalRequests)

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ? AND status = ?", startDate, endDate+" 23:59:59", "success").
		Count(&successRequests)

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate+" 23:59:59").
		Select("COALESCE(SUM(total_tokens), 0)").Scan(&totalTokens)

	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate+" 23:59:59").
		Select("COALESCE(AVG(latency_ms), 0)").Scan(&avgLatency)

	successRate := float64(0)
	if totalRequests > 0 {
		successRate = float64(successRequests) / float64(totalRequests) * 100
	}

	var modelStats []struct {
		Model        string  `json:"model"`
		ActualModel  string  `json:"actual_model"`
		ProviderName string  `json:"provider_name"`
		Count        int64   `json:"count"`
		Tokens       int64   `json:"tokens"`
		AvgLatency   float64 `json:"avg_latency"`
	}
	model.DB.Raw(`
		SELECT 
			ul.model,
			COALESCE(ul.actual_model, ul.model) as actual_model,
			COALESCE(p.name, 'Unknown') as provider_name,
			COUNT(*) as count,
			COALESCE(SUM(ul.total_tokens), 0) as tokens,
			COALESCE(AVG(ul.latency_ms), 0) as avg_latency
		FROM usage_logs ul
		LEFT JOIN providers p ON ul.provider_id = p.id
		WHERE ul.created_at >= ? AND ul.created_at <= ?
		GROUP BY ul.model, ul.actual_model, p.name
		ORDER BY count DESC
		LIMIT 20
	`, startDate, endDate+" 23:59:59").Scan(&modelStats)

	c.JSON(http.StatusOK, gin.H{
		"totalRequests":   totalRequests,
		"successRequests": successRequests,
		"successRate":     successRate,
		"totalTokens":     totalTokens,
		"avgLatency":      avgLatency,
		"modelStats":      modelStats,
	})
}

type logsResponse struct {
	ID              uint      `json:"id"`
	Source          string    `json:"source"`
	KeyID           uint      `json:"key_id"`
	KeyName         string    `json:"key_name"`
	Model           string    `json:"model"`
	ProviderType    string    `json:"provider_type"`
	ProviderID      uint      `json:"provider_id"`
	ProviderName    string    `json:"provider_name"`
	ActualModelID   string    `json:"actual_model_id"`
	ActualModelName string    `json:"actual_model_name"`
	TotalTokens     int64     `json:"total_tokens"`
	LatencyMs       int64     `json:"latency_ms"`
	Status          string    `json:"status"`
	ErrorMsg        string    `json:"error_msg"`
	CreatedAt       time.Time `json:"created_at"`
}

func (h *UsageHandler) Logs(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -1).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	paramKey := c.Query("key")
	paramModel := c.Query("model")

	query := "SELECT * FROM usage_logs WHERE created_at >= ? AND created_at <= ?"
	args := []interface{}{startDate, endDate + " 23:59:59"}

	if paramKey != "" {
		query += " AND key_id = ?"
		args = append(args, paramKey)
	}

	if paramModel != "" {
		query += " AND model = ?"
		args = append(args, paramModel)
	}

	query += " ORDER BY created_at DESC LIMIT 1000"

	var logsResponse []logsResponse
	if err := model.DB.Raw(query, args...).Scan(&logsResponse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logsResponse})
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
	model.DB.Model(&model.Key{}).Where("enabled = ?", true).Count(&activeKeys)

	var totalTokens int64
	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ?", sevenDaysAgo).
		Select("COALESCE(SUM(total_tokens), 0)").Scan(&totalTokens)

	var avgLatency float64
	model.DB.Model(&model.UsageLog{}).
		Where("created_at >= ?", sevenDaysAgo).
		Select("COALESCE(AVG(latency_ms), 0)").Scan(&avgLatency)

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
		Provider   string  `json:"provider"`
		Count      int64   `json:"count"`
		Tokens     int64   `json:"tokens"`
		AvgLatency float64 `json:"avg_latency"`
	}
	model.DB.Raw(`
		SELECT 
			p.name as provider, 
			COUNT(*) as count,
			COALESCE(SUM(ul.total_tokens), 0) as tokens,
			COALESCE(AVG(ul.latency_ms), 0) as avg_latency
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
		"totalTokens":     totalTokens,
		"avgLatency":      avgLatency,
		"dailyStats":      dailyStats,
		"providerStats":   providerStats,
		"modelStats":      modelStats,
	})
}

func (h *UsageHandler) KeyStats(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	var keyStats []struct {
		KeyID      uint    `json:"key_id"`
		KeyName    string  `json:"key_name"`
		Count      int64   `json:"count"`
		Tokens     int64   `json:"tokens"`
		AvgLatency float64 `json:"avg_latency"`
	}
	model.DB.Raw(`
		SELECT 
			ul.key_id as key_id,
			COALESCE(k.name, 'Unknown') as key_name,
			COUNT(*) as count,
			COALESCE(SUM(ul.total_tokens), 0) as tokens,
			COALESCE(AVG(ul.latency_ms), 0) as avg_latency
		FROM usage_logs ul
		LEFT JOIN keys k ON ul.key_id = k.id
		WHERE ul.created_at >= ? AND ul.created_at <= ?
		GROUP BY ul.key_id, k.name
		ORDER BY count DESC
	`, startDate, endDate+" 23:59:59").Scan(&keyStats)

	c.JSON(http.StatusOK, gin.H{
		"keyStats": keyStats,
	})
}

func NewUsageLog(source string, keyID uint, keyName, modelName string, result *router.RouteResult, tokens int, latencyMs int, status string, errorMsg string) *model.UsageLog {
	actualModelName := result.ProviderModel.DisplayName
	if actualModelName == "" {
		actualModelName = result.ProviderModel.ModelID
	}
	return &model.UsageLog{
		Source:          source,
		KeyID:           keyID,
		KeyName:         keyName,
		Model:           modelName,
		ProviderType:    result.Provider.Type,
		ProviderID:      result.Provider.ID,
		ProviderName:    result.Provider.Name,
		ActualModelID:   result.ProviderModel.ModelID,
		ActualModelName: actualModelName,
		TotalTokens:     tokens,
		LatencyMs:       latencyMs,
		Status:          status,
		ErrorMsg:        errorMsg,
	}
}
