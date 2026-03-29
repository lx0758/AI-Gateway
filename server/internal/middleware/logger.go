package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		if status >= 400 {
			log.Printf("[API] %s %s %d %v %s errors=%v",
				method, path, status, latency, clientIP, c.Errors.String())
		} else {
			log.Printf("[API] %s %s %d %v %s",
				method, path, status, latency, clientIP)
		}
	}
}
