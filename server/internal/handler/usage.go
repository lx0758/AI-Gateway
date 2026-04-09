package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
	"ai-gateway/internal/provider"
	"ai-gateway/internal/router"
)

type UsageHandler struct{}

type dailyStat struct {
	Date    string `json:"date"`
	Count   int64  `json:"count"`
	Success int64  `json:"success"`
}

type tokenDailyStat struct {
	Date        string `json:"date"`
	TotalTokens int64  `json:"total_tokens"`
}

func NewUsageHandler() *UsageHandler {
	return &UsageHandler{}
}

func (h *UsageHandler) Dashboard(c *gin.Context) {
	nDays := 7
	nDaysAgo := time.Now().AddDate(0, 0, -nDays).Format("2006-01-02")
	lastNDays := generateLastNDays(nDays)

	// 资产统计
	var totalProviders int64
	model.DB.Model(&model.Provider{}).Count(&totalProviders)

	var activeProviders int64
	model.DB.Model(&model.Provider{}).Where("enabled = ?", true).Count(&activeProviders)

	var totalModels int64
	model.DB.Model(&model.Model{}).Count(&totalModels)

	var activeModels int64
	model.DB.Model(&model.Model{}).Where("enabled = ?", true).Count(&activeModels)

	var totalProviderModels int64
	model.DB.Model(&model.ProviderModel{}).Count(&totalProviderModels)

	var activeProviderModels int64
	model.DB.Raw(`
		SELECT COUNT(DISTINCT pm.id)
		FROM provider_models pm
		JOIN providers p ON pm.provider_id = p.id
		WHERE pm.is_available = ? AND p.enabled = ?
	`, true, true).Scan(&activeProviderModels)

	var totalMCPs int64
	model.DB.Model(&model.MCP{}).Count(&totalMCPs)

	var activeMCPs int64
	model.DB.Model(&model.MCP{}).Where("enabled = ?", true).Count(&activeMCPs)

	var totalKeys int64
	model.DB.Model(&model.Key{}).Count(&totalKeys)

	var activeKeys int64
	model.DB.Model(&model.Key{}).Where("enabled = ?", true).Count(&activeKeys)

	// Model API 统计 (过去N天)
	var modelTotalRequests int64
	model.DB.Model(&model.ModelLog{}).
		Where("created_at >= ?", nDaysAgo).
		Count(&modelTotalRequests)

	var modelSuccessCount int64
	model.DB.Model(&model.ModelLog{}).
		Where("created_at >= ? AND status = ?", nDaysAgo, "success").
		Count(&modelSuccessCount)

	var modelTotalTokens int64
	model.DB.Model(&model.ModelLog{}).
		Where("created_at >= ?", nDaysAgo).
		Select("COALESCE(SUM(total_tokens), 0)").Scan(&modelTotalTokens)

	var modelAvgLatency float64
	model.DB.Model(&model.ModelLog{}).
		Where("created_at >= ?", nDaysAgo).
		Select("COALESCE(AVG(latency_ms), 0)").Scan(&modelAvgLatency)

	var modelDailyStats []dailyStat
	model.DB.Raw(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as count,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success
		FROM model_logs
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, nDaysAgo).Scan(&modelDailyStats)

	modelDailyStats = fillDailyStats(lastNDays, modelDailyStats,
		func(s dailyStat) string {
			if len(s.Date) > 10 {
				return s.Date[:10]
			}
			return s.Date
		},
		func(date string) dailyStat { return dailyStat{Date: date} })

	var modelTokenDailyStats []tokenDailyStat
	model.DB.Raw(`
		SELECT 
			DATE(created_at) as date,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM model_logs
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, nDaysAgo).Scan(&modelTokenDailyStats)

	modelTokenDailyStats = fillDailyStats(lastNDays, modelTokenDailyStats,
		func(s tokenDailyStat) string {
			if len(s.Date) > 10 {
				return s.Date[:10]
			}
			return s.Date
		},
		func(date string) tokenDailyStat { return tokenDailyStat{Date: date} })

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
			COALESCE(SUM(ml.total_tokens), 0) as tokens,
			COALESCE(AVG(ml.latency_ms), 0) as avg_latency
		FROM model_logs ml
		JOIN providers p ON ml.provider_id = p.id
		WHERE ml.created_at >= ?
		GROUP BY p.name
		ORDER BY count DESC
	`, nDaysAgo).Scan(&providerStats)

	var modelStats []struct {
		Model string `json:"model"`
		Count int64  `json:"count"`
	}
	model.DB.Raw(`
		SELECT model, COUNT(*) as count
		FROM model_logs
		WHERE created_at >= ?
		GROUP BY model
		ORDER BY count DESC
		LIMIT 10
	`, nDaysAgo).Scan(&modelStats)

	// MCP 资产统计
	var totalMCPTools int64
	var totalMCPResources int64
	var totalMCPPrompts int64
	var activeMCPTools int64
	var activeMCPResources int64
	var activeMCPPrompts int64
	model.DB.Table("mcp_tools").Count(&totalMCPTools)
	model.DB.Table("mcp_resources").Count(&totalMCPResources)
	model.DB.Table("mcp_prompts").Count(&totalMCPPrompts)
	model.DB.Raw(`
		SELECT COUNT(DISTINCT mt.id)
		FROM mcp_tools mt
		JOIN mcps m ON mt.mcp_id = m.id
		WHERE mt.enabled = ? AND m.enabled = ?
	`, true, true).Scan(&activeMCPTools)
	model.DB.Raw(`
		SELECT COUNT(DISTINCT mr.id)
		FROM mcp_resources mr
		JOIN mcps m ON mr.mcp_id = m.id
		WHERE mr.enabled = ? AND m.enabled = ?
	`, true, true).Scan(&activeMCPResources)
	model.DB.Raw(`
		SELECT COUNT(DISTINCT mp.id)
		FROM mcp_prompts mp
		JOIN mcps m ON mp.mcp_id = m.id
		WHERE mp.enabled = ? AND m.enabled = ?
	`, true, true).Scan(&activeMCPPrompts)

	// MCP 服务统计 (过去7天)
	var mcpTotalRequests int64
	model.DB.Model(&model.MCPLog{}).
		Where("created_at >= ?", nDaysAgo).
		Count(&mcpTotalRequests)

	var mcpSuccessCount int64
	model.DB.Model(&model.MCPLog{}).
		Where("created_at >= ? AND status = ?", nDaysAgo, "success").
		Count(&mcpSuccessCount)

	var mcpTotalSize int64
	model.DB.Model(&model.MCPLog{}).
		Where("created_at >= ?", nDaysAgo).
		Select("COALESCE(SUM(input_size + output_size), 0)").Scan(&mcpTotalSize)

	var mcpAvgLatency float64
	model.DB.Model(&model.MCPLog{}).
		Where("created_at >= ?", nDaysAgo).
		Select("COALESCE(AVG(latency_ms), 0)").Scan(&mcpAvgLatency)

	var mcpDailyStats []dailyStat
	model.DB.Raw(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as count,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success
		FROM mcp_logs
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, nDaysAgo).Scan(&mcpDailyStats)

	mcpDailyStats = fillDailyStats(lastNDays, mcpDailyStats,
		func(s dailyStat) string {
			if len(s.Date) > 10 {
				return s.Date[:10]
			}
			return s.Date
		},
		func(date string) dailyStat { return dailyStat{Date: date} })

	var mcpTypeStats []struct {
		MCPType string `json:"mcp_type"`
		Count   int64  `json:"count"`
	}
	model.DB.Raw(`
		SELECT mcp_type, COUNT(*) as count
		FROM mcp_logs
		WHERE created_at >= ?
		GROUP BY mcp_type
		ORDER BY count DESC
	`, nDaysAgo).Scan(&mcpTypeStats)

	var mcpServiceStats []struct {
		MCPName string `json:"mcp_name"`
		Count   int64  `json:"count"`
	}
	model.DB.Raw(`
		SELECT mcp_name, COUNT(*) as count
		FROM mcp_logs
		WHERE created_at >= ?
		GROUP BY mcp_name
		ORDER BY count DESC
		LIMIT 10
	`, nDaysAgo).Scan(&mcpServiceStats)

	for i := range modelDailyStats {
		if len(modelDailyStats[i].Date) > 10 {
			modelDailyStats[i].Date = modelDailyStats[i].Date[:10]
		}
	}
	for i := range modelTokenDailyStats {
		if len(modelTokenDailyStats[i].Date) > 10 {
			modelTokenDailyStats[i].Date = modelTokenDailyStats[i].Date[:10]
		}
	}
	for i := range mcpDailyStats {
		if len(mcpDailyStats[i].Date) > 10 {
			mcpDailyStats[i].Date = mcpDailyStats[i].Date[:10]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"days": nDays,
		"assets": gin.H{
			"totalProviders":       totalProviders,
			"activeProviders":      activeProviders,
			"totalModels":          totalModels,
			"activeModels":         activeModels,
			"totalProviderModels":  totalProviderModels,
			"activeProviderModels": activeProviderModels,
			"totalMCPs":            totalMCPs,
			"activeMCPs":           activeMCPs,
			"totalKeys":            totalKeys,
			"activeKeys":           activeKeys,
			"totalMCPTools":        totalMCPTools,
			"activeMCPTools":       activeMCPTools,
			"totalMCPResources":    totalMCPResources,
			"activeMCPResources":   activeMCPResources,
			"totalMCPPrompts":      totalMCPPrompts,
			"activeMCPPrompts":     activeMCPPrompts,
		},
		"modelUsage": gin.H{
			"totalRequests":   modelTotalRequests,
			"successCount":    modelSuccessCount,
			"totalTokens":     modelTotalTokens,
			"avgLatency":      modelAvgLatency,
			"dailyStats":      modelDailyStats,
			"tokenDailyStats": modelTokenDailyStats,
			"providerStats":   providerStats,
			"modelStats":      modelStats,
		},
		"mcpUsage": gin.H{
			"totalRequests": mcpTotalRequests,
			"successCount":  mcpSuccessCount,
			"totalSize":     mcpTotalSize,
			"avgLatency":    mcpAvgLatency,
			"dailyStats":    mcpDailyStats,
			"typeStats":     mcpTypeStats,
			"serviceStats":  mcpServiceStats,
		},
	})
}

func generateLastNDays(n int) []string {
	now := time.Now()
	days := make([]string, n)
	for i := 0; i < n; i++ {
		offset := -n + 1 + i
		days[i] = now.AddDate(0, 0, offset).Format("2006-01-02")
	}
	return days
}

func fillDailyStats[T any](days []string, stats []T, getDate func(T) string, createEmpty func(string) T) []T {
	statsMap := make(map[string]T)
	for _, s := range stats {
		statsMap[getDate(s)] = s
	}

	result := make([]T, len(days))
	for i, day := range days {
		if s, ok := statsMap[day]; ok {
			result[i] = s
		} else {
			result[i] = createEmpty(day)
		}
	}
	return result
}

type modelLogResponse struct {
	ID              uint      `json:"id"`
	Source          string    `json:"source"`
	ClientIPs       string    `json:"client_ips"`
	KeyID           uint      `json:"key_id"`
	KeyName         string    `json:"key_name"`
	Model           string    `json:"model"`
	ProviderID      uint      `json:"provider_id"`
	ProviderName    string    `json:"provider_name"`
	ActualModelID   string    `json:"actual_model_id"`
	ActualModelName string    `json:"actual_model_name"`
	CallMethod      string    `json:"call_method"`
	CachedTokens    int       `json:"cached_tokens"`
	InputTokens     int       `json:"input_tokens"`
	OutputTokens    int       `json:"output_tokens"`
	TotalTokens     int       `json:"total_tokens"`
	LatencyMs       int       `json:"latency_ms"`
	Status          string    `json:"status"`
	ErrorMsg        string    `json:"error_msg"`
	CreatedAt       time.Time `json:"created_at"`
}

func (h *UsageHandler) ModelLogs(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02 00:00:00"))
	endDate := c.DefaultQuery("end_date", time.Now().AddDate(0, 0, 1).Format("2006-01-02 00:00:00"))

	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", startDate, time.Local)
	if err != nil {
		startTime, err = time.ParseInLocation("2006-01-02", startDate, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", endDate, time.Local)
	if err != nil {
		endTime, err = time.ParseInLocation("2006-01-02", endDate, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
	}

	var modelLogs []model.ModelLog
	if err := model.DB.Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Order("created_at DESC").
		Find(&modelLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logsResponses := make([]modelLogResponse, len(modelLogs))
	for i, log := range modelLogs {
		logsResponses[i] = modelLogResponse{
			ID:              log.ID,
			Source:          log.Source,
			ClientIPs:       log.ClientIPs,
			KeyID:           log.KeyID,
			KeyName:         log.KeyName,
			Model:           log.Model,
			ProviderID:      log.ProviderID,
			ProviderName:    log.ProviderName,
			ActualModelID:   log.ActualModelID,
			ActualModelName: log.ActualModelName,
			CallMethod:      log.CallMethod,
			CachedTokens:    log.CachedTokens,
			InputTokens:     log.InputTokens,
			OutputTokens:    log.OutputTokens,
			TotalTokens:     log.TotalTokens,
			LatencyMs:       log.LatencyMs,
			Status:          log.Status,
			ErrorMsg:        log.ErrorMsg,
			CreatedAt:       log.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"logs": logsResponses})
}

func NewModelLog(source string, clientIPs string, keyID uint, keyName, modelName string, result *router.RouteResult, matched bool, usage *provider.Usage, latencyMs int, status string, errorMsg string) *model.ModelLog {
	actualModelName := result.ProviderModel.DisplayName
	if actualModelName == "" {
		actualModelName = result.ProviderModel.ModelID
	}
	callMethod := "direct"
	if !matched {
		callMethod = "convert"
	}
	return &model.ModelLog{
		Source:          source,
		ClientIPs:       clientIPs,
		KeyID:           keyID,
		KeyName:         keyName,
		Model:           modelName,
		ProviderID:      result.Provider.ID,
		ProviderName:    result.Provider.Name,
		ActualModelID:   result.ProviderModel.ModelID,
		ActualModelName: actualModelName,
		CallMethod:      callMethod,
		CachedTokens:    usage.CachedTokens,
		InputTokens:     usage.InputTokens,
		OutputTokens:    usage.OutputTokens,
		TotalTokens:     usage.TotalTokens(),
		LatencyMs:       latencyMs,
		Status:          status,
		ErrorMsg:        errorMsg,
	}
}

func NewMCPLog(source string, clientIPs string, keyID uint, keyName string, mcpID uint, mcpName string, mcpType string, callType string, callTarget string, callMethod string, inputSize int, outputSize int, latencyMs int, status string, errorMsg string) *model.MCPLog {
	return &model.MCPLog{
		Source:     source,
		ClientIPs:  clientIPs,
		KeyID:      keyID,
		KeyName:    keyName,
		MCPID:      mcpID,
		MCPName:    mcpName,
		MCPType:    mcpType,
		CallType:   callType,
		CallTarget: callTarget,
		CallMethod: callMethod,
		InputSize:  inputSize,
		OutputSize: outputSize,
		LatencyMs:  latencyMs,
		Status:     status,
		ErrorMsg:   errorMsg,
	}
}

type mcpLogResponse struct {
	ID         uint      `json:"id"`
	Source     string    `json:"source"`
	ClientIPs  string    `json:"client_ips"`
	KeyID      uint      `json:"key_id"`
	KeyName    string    `json:"key_name"`
	MCPID      uint      `json:"mcp_id"`
	MCPName    string    `json:"mcp_name"`
	MCPType    string    `json:"mcp_type"`
	CallType   string    `json:"call_type"`
	CallMethod string    `json:"call_method"`
	CallTarget string    `json:"call_target"`
	InputSize  int       `json:"input_size"`
	OutputSize int       `json:"output_size"`
	LatencyMs  int       `json:"latency_ms"`
	Status     string    `json:"status"`
	ErrorMsg   string    `json:"error_msg"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *UsageHandler) MCPLogs(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02 00:00:00"))
	endDate := c.DefaultQuery("end_date", time.Now().AddDate(0, 0, 1).Format("2006-01-02 00:00:00"))

	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", startDate, time.Local)
	if err != nil {
		startTime, err = time.ParseInLocation("2006-01-02", startDate, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", endDate, time.Local)
	if err != nil {
		endTime, err = time.ParseInLocation("2006-01-02", endDate, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
	}

	var mcpLogs []model.MCPLog
	if err := model.DB.Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Order("created_at DESC").
		Find(&mcpLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logsResponses := make([]mcpLogResponse, len(mcpLogs))
	for i, log := range mcpLogs {
		logsResponses[i] = mcpLogResponse{
			ID:         log.ID,
			Source:     log.Source,
			ClientIPs:  log.ClientIPs,
			KeyID:      log.KeyID,
			KeyName:    log.KeyName,
			MCPID:      log.MCPID,
			MCPName:    log.MCPName,
			MCPType:    log.MCPType,
			CallType:   log.CallType,
			CallMethod: log.CallMethod,
			CallTarget: log.CallTarget,
			InputSize:  log.InputSize,
			OutputSize: log.OutputSize,
			LatencyMs:  log.LatencyMs,
			Status:     log.Status,
			ErrorMsg:   log.ErrorMsg,
			CreatedAt:  log.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"logs": logsResponses})
}
