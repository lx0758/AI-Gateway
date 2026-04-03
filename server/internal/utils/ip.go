package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func GetClientIPInfo(c *gin.Context) string {
	clientIP := c.ClientIP()
	forwardedFor := c.GetHeader("X-Forwarded-For")

	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		cleanedIPs := []string{}
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				cleanedIPs = append(cleanedIPs, ip)
			}
		}
		if len(cleanedIPs) > 0 {
			return strings.Join(cleanedIPs, ", ")
		}
	}

	return clientIP
}
