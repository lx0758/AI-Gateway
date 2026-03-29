package middleware

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Static(fsys fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(fsys))

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/openai") || path == "/health" {
			c.Next()
			return
		}

		if _, err := fs.Stat(fsys, strings.TrimPrefix(path, "/")); err != nil {
			c.Request.URL.Path = "/"
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
