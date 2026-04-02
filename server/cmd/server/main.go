package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/config"
	"ai-gateway/internal/handler"
	"ai-gateway/internal/middleware"
	"ai-gateway/internal/model"
	"ai-gateway/res"
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
	aliasHandler := handler.NewAliasHandler()
	apiKeyHandler := handler.NewAPIKeyHandler()
	openAIProxyHandler := handler.NewOpenAIProxyHandler()
	anthropicProxyHandler := handler.NewAnthropicProxyHandler()
	usageHandler := handler.NewUsageHandler()

	openai := r.Group("/openai/v1")
	openai.Use(middleware.RequireAPIKey())
	{
		openai.POST("/chat/completions", openAIProxyHandler.ChatCompletions)
		openai.GET("/models", openAIProxyHandler.ListModels)
		openai.GET("/models/:id", openAIProxyHandler.GetModel)
	}

	anthropic := r.Group("/anthropic/v1")
	anthropic.Use(middleware.RequireAPIKeyForAnthropic())
	{
		anthropic.POST("/messages", anthropicProxyHandler.Messages)
		anthropic.GET("/models", anthropicProxyHandler.ListModels)
		anthropic.GET("/models/:id", anthropicProxyHandler.GetModel)
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

			protected.GET("/aliases", aliasHandler.List)
			protected.POST("/aliases", aliasHandler.Create)
			protected.GET("/aliases/:id", aliasHandler.Get)
			protected.PUT("/aliases/:id", aliasHandler.Update)
			protected.DELETE("/aliases/:id", aliasHandler.Delete)

			protected.GET("/aliases/:id/mappings", aliasHandler.ListMappings)
			protected.POST("/aliases/:id/mappings", aliasHandler.CreateMapping)
			protected.PUT("/aliases/:id/mappings/:mid", aliasHandler.UpdateMapping)
			protected.DELETE("/aliases/:id/mappings/:mid", aliasHandler.DeleteMapping)

			protected.GET("/api-keys", apiKeyHandler.List)
			protected.POST("/api-keys", apiKeyHandler.Create)
			protected.PUT("/api-keys/:id", apiKeyHandler.Update)
			protected.DELETE("/api-keys/:id", apiKeyHandler.Delete)
			protected.POST("/api-keys/:id/reset", apiKeyHandler.Reset)
			protected.GET("/api-keys/:id/models", apiKeyHandler.ListModels)
			protected.POST("/api-keys/:id/models", apiKeyHandler.AddModel)
			protected.DELETE("/api-keys/:id/models/:model_alias", apiKeyHandler.RemoveModel)

			protected.GET("/usage/logs", usageHandler.Logs)
			protected.GET("/usage/dashboard", usageHandler.Dashboard)
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
