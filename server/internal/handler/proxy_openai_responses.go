package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/model"
	"ai-gateway/internal/provider"
	"ai-gateway/internal/utils"
)

// gatewayError 输出 OpenAI 风格的错误体，供拦截层与管理端点共用，
// 保证 stateless_gateway 契约在网关级拒绝处一致。
func gatewayError(c *gin.Context, status int, code, message string) {
	errorType := "invalid_request_error"
	if status >= 500 {
		errorType = "server_error"
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"message": message,
			"type":    errorType,
			"code":    code,
		},
	})
}

// responses 端点：强制无状态拦截 + store 改写 + 复用现有鉴权/路由/计费链路。
func (h *OpenAIProxyHandler) Responses(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		gatewayError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	bodyJSON := map[string]interface{}{}
	if err := json.Unmarshal(body, &bodyJSON); err != nil {
		gatewayError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if v, ok := bodyJSON["previous_response_id"].(string); ok && v != "" {
		gatewayError(c, http.StatusBadRequest, "stateless_gateway",
			"previous_response_id is not supported: this gateway is stateless. Send the full conversation input instead.")
		return
	}
	if v, exists := bodyJSON["conversation"]; exists && v != nil {
		gatewayError(c, http.StatusBadRequest, "stateless_gateway",
			"conversation is not supported: this gateway is stateless. Send the full conversation input instead.")
		return
	}
	if _, exists := bodyJSON["background"]; exists {
		gatewayError(c, http.StatusBadRequest, "stateless_gateway",
			"background is not supported: this gateway is stateless. Only synchronous requests are allowed.")
		return
	}

	// 状态只允许存在于上游或客户端，网关永不持有
	bodyJSON["store"] = false
	newBody, err := json.Marshal(bodyJSON)
	if err != nil {
		gatewayError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(newBody))

	reqModel, _ := bodyJSON["model"].(string)
	if reqModel == "" {
		gatewayError(c, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}

	keyID, _ := c.Get("key_id")
	keyName, _ := c.Get("key_name")
	if err := VerifyKeyID(keyID, reqModel); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	result, err := h.router.Route(reqModel)
	if err != nil {
		gatewayError(c, http.StatusInternalServerError, "route_error", err.Error())
		return
	}
	if result == nil {
		gatewayError(c, http.StatusNotFound, "model_not_found", "model not found or no available provider")
		return
	}

	start := time.Now()
	usage := provider.Usage{}
	err = result.ProviderInstance.ExecuteOpenAIResponsesRequest(c, result.ProviderModel, &usage)
	latencyMs := time.Since(start).Milliseconds()

	status := "success"
	errorMsg := ""
	if err != nil {
		status = "error"
		errorMsg = err.Error()
		if provider.IsRateLimitError(err) {
			h.router.RecordRateLimit(result.Provider.ID, result.ProviderModel.ID)
		}
	} else {
		h.router.RecordSuccess(result.Provider.ID, result.ProviderModel.ID)
	}

	clientIPs := utils.GetClientIPInfo(c)

	modelLog := NewModelLog(
		"responses",
		clientIPs,
		keyID.(uint),
		keyName.(string),
		reqModel,
		result,
		result.SupportResponses(),
		&usage,
		int(latencyMs),
		status,
		errorMsg,
	)
	model.DB.Create(&modelLog)
	log.Println(modelLog.String())
}

// ResponsesManagement 覆盖全部需要状态的管理端点，网关直接应答 400。
func (h *OpenAIProxyHandler) ResponsesManagement(c *gin.Context) {
	gatewayError(c, http.StatusBadRequest, "stateless_gateway",
		"This gateway does not store responses or conversations. Send the full conversation input with store: false.")
}
