package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"ai-proxy/internal/config"
	"ai-proxy/internal/handler"
	"ai-proxy/internal/middleware"
	"ai-proxy/internal/model"
	"ai-proxy/res"
)

func main() {
	cfg := config.Load()

	log.Printf("AI Model Proxy v%s", res.Version)

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
	anthropicProxyHandler := handler.NewAnthropicProxyHandler()
	usageHandler := handler.NewUsageHandler()

	openai := r.Group("/openai/v1")
	openai.Use(middleware.RequireAPIKey())
	{
		openai.POST("/chat/completions", proxyHandler.ChatCompletions)
		openai.GET("/models", proxyHandler.ListModels)
		openai.GET("/models/:id", proxyHandler.GetModel)
	}

	anthropic := r.Group("/anthropic/v1")
	anthropic.Use(middleware.RequireAPIKeyForAnthropic())
	{
		anthropic.POST("/messages", anthropicProxyHandler.Messages)
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
			protected.PUT("/api-keys/:id", apiKeyHandler.Update)
			protected.DELETE("/api-keys/:id", apiKeyHandler.Delete)
			protected.GET("/api-keys/:id/models", apiKeyHandler.ListModels)
			protected.POST("/api-keys/:id/models", apiKeyHandler.AddModel)
			protected.DELETE("/api-keys/:id/models/:model_alias", apiKeyHandler.RemoveModel)

			protected.GET("/usage/stats", usageHandler.Stats)
			protected.GET("/usage/logs", usageHandler.Logs)
			protected.GET("/usage/dashboard", usageHandler.Dashboard)
			protected.GET("/usage/key-stats", usageHandler.KeyStats)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": res.Version})
	})

	r.NoRoute(middleware.Static(res.WebFS))

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
