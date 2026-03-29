package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func Static(root string) gin.HandlerFunc {
	fileServer := http.FileServer(http.Dir(root))

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/v1") || path == "/health" {
			c.Next()
			return
		}

		filePath := root + path
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.Request.URL.Path = "/"
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
