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

func NewUsageHandler() *UsageHandler {
	return &UsageHandler{}
}

func (h *UsageHandler) Dashboard(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	var todayRequests int64
	model.DB.Model(&model.ModelLog{}).
		Where("created_at >= ?", today).
		Count(&todayRequests)

	var activeProviders int64
	model.DB.Model(&model.Provider{}).Where("enabled = ?", true).Count(&activeProviders)

	var activeKeys int64
	model.DB.Model(&model.Key{}).Where("enabled = ?", true).Count(&activeKeys)

	var totalTokens int64
	model.DB.Model(&model.ModelLog{}).
		Where("created_at >= ?", sevenDaysAgo).
		Select("COALESCE(SUM(total_tokens), 0)").Scan(&totalTokens)

	var avgLatency float64
	model.DB.Model(&model.ModelLog{}).
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
		FROM model_logs
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
		FROM model_logs ul
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
		FROM model_logs
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
