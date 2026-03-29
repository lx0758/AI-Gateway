package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/user/ai-model-proxy/internal/config"
	"github.com/user/ai-model-proxy/internal/handler"
	"github.com/user/ai-model-proxy/internal/middleware"
	"github.com/user/ai-model-proxy/internal/model"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	model.SetEncryptionKey(cfg.Session.Secret)

	if err := model.InitDB(cfg.Database.Path); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	if err := model.InitDefaultAdmin(cfg.Auth.DefaultAdmin.Username, cfg.Auth.DefaultAdmin.Password); err != nil {
		log.Fatalf("Failed to init default admin: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)

	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())

	r.Use(middleware.SetupSessionStore(
		cfg.Session.Secret,
		cfg.Session.MaxAge,
		cfg.Session.Secure,
		cfg.Session.HttpOnly,
		cfg.Session.SameSite,
	))

	authHandler := handler.NewAuthHandler()
	providerHandler := handler.NewProviderHandler()
	providerModelHandler := handler.NewProviderModelHandler()
	modelMappingHandler := handler.NewModelMappingHandler()
	apiKeyHandler := handler.NewAPIKeyHandler()
	proxyHandler := handler.NewProxyHandler()
	usageHandler := handler.NewUsageHandler()

	v1 := r.Group("/v1")
	v1.Use(middleware.RequireAPIKey())
	{
		v1.POST("/chat/completions", proxyHandler.ChatCompletions)
		v1.GET("/models", proxyHandler.ListModels)
		v1.GET("/models/:id", proxyHandler.GetModel)
	}

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		protected := api.Group("")
		protected.Use(middleware.RequireAuth())
		{
			protected.GET("/auth/me", authHandler.Me)
			protected.PUT("/auth/password", authHandler.ChangePassword)

			protected.GET("/providers", providerHandler.List)
			protected.GET("/providers/:id", providerHandler.Get)
			protected.POST("/providers", providerHandler.Create)
			protected.PUT("/providers/:id", providerHandler.Update)
			protected.DELETE("/providers/:id", providerHandler.Delete)
			protected.POST("/providers/:id/test", providerHandler.Test)

			protected.GET("/providers/:id/models", providerModelHandler.List)
			protected.POST("/providers/:id/models", providerModelHandler.Create)
			protected.PUT("/providers/:id/models/:mid", providerModelHandler.Update)
			protected.DELETE("/providers/:id/models/:mid", providerModelHandler.Delete)
			protected.POST("/providers/:id/sync", providerModelHandler.Sync)

			protected.GET("/model-mappings", modelMappingHandler.List)
			protected.POST("/model-mappings", modelMappingHandler.Create)
			protected.PUT("/model-mappings/:id", modelMappingHandler.Update)
			protected.DELETE("/model-mappings/:id", modelMappingHandler.Delete)

			protected.GET("/api-keys", apiKeyHandler.List)
			protected.POST("/api-keys", apiKeyHandler.Create)
			protected.DELETE("/api-keys/:id", apiKeyHandler.Delete)

			protected.GET("/usage/stats", usageHandler.Stats)
			protected.GET("/usage/logs", usageHandler.Logs)
			protected.GET("/usage/dashboard", usageHandler.Dashboard)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Use(middleware.Static("web/dist"))

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
