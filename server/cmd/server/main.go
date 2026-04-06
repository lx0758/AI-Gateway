package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"

	"ai-gateway/internal/config"
	"ai-gateway/internal/handler"
	"ai-gateway/internal/mcp"
	"ai-gateway/internal/middleware"
	"ai-gateway/internal/model"
	"ai-gateway/internal/provider"
	"ai-gateway/res"
)

func main() {
	cfg := config.Load()

	log.Printf("AI Gateway %s", res.Version)

	if err := model.InitDB(
		cfg.Database.Type,
		cfg.Database.Path,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Pool.MaxOpen,
		cfg.Database.Pool.MaxIdle,
		cfg.Database.Pool.MaxLifetime,
		cfg.Database.Pool.MaxIdleTime,
		cfg.Debug.Gorm,
	); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	if err := model.InitDefaultAdmin(cfg.Auth.DefaultAdmin.Username, cfg.Auth.DefaultAdmin.Password); err != nil {
		log.Fatalf("Failed to init default admin: %v", err)
	}

	provider.SetDebugMode(cfg.Debug.Provider)
	mcp.SetDebugMode(cfg.Debug.MCP)

	if cfg.Debug.Gin {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	if err := r.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		log.Printf("Warning: Failed to set trusted proxies: %v, using default configuration", err)
	}

	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())

	r.Use(middleware.SetupSessionStore(
		cfg.Server.Session.Secret,
		cfg.Server.Session.MaxAge,
		cfg.Server.Session.Secure,
		cfg.Server.Session.HttpOnly,
		cfg.Server.Session.SameSite,
	))

	authHandler := handler.NewAuthHandler()
	providerHandler := handler.NewProviderHandler()
	providerModelHandler := handler.NewProviderModelHandler()
	aliasHandler := handler.NewAliasHandler()
	apiKeyHandler := handler.NewAPIKeyHandler()
	openAIProxyHandler := handler.NewOpenAIProxyHandler()
	anthropicProxyHandler := handler.NewAnthropicProxyHandler()
	usageHandler := handler.NewUsageHandler()
	mcpProxyHandler := handler.NewMCPProxyHandler()
	mcpHandler := handler.NewMCPHandler()

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

	mcp := r.Group("/mcp/v1")
	mcp.Use(middleware.RequireAPIKey())
	{
		mcp.GET("", mcpProxyHandler.Handle)
		mcp.POST("", mcpProxyHandler.Handle)
		mcp.DELETE("", mcpProxyHandler.Handle)
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
			protected.PUT("/aliases/:id/mappings/order", aliasHandler.UpdateMappingsOrder)
			protected.PUT("/aliases/:id/mappings/:mid", aliasHandler.UpdateMapping)
			protected.DELETE("/aliases/:id/mappings/:mid", aliasHandler.DeleteMapping)

			protected.GET("/api-keys", apiKeyHandler.List)
			protected.POST("/api-keys", apiKeyHandler.Create)
			protected.PUT("/api-keys/:id", apiKeyHandler.Update)
			protected.DELETE("/api-keys/:id", apiKeyHandler.Delete)
			protected.POST("/api-keys/:id/reset", apiKeyHandler.Reset)
			protected.GET("/api-keys/:id/models", apiKeyHandler.ListModels)
			protected.GET("/api-keys/:id/mcp-tools", apiKeyHandler.GetMCPTools)
			protected.PUT("/api-keys/:id/mcp-tools", apiKeyHandler.UpdateMCPTools)
			protected.GET("/api-keys/:id/mcp-resources", apiKeyHandler.GetMCPResources)
			protected.PUT("/api-keys/:id/mcp-resources", apiKeyHandler.UpdateMCPResources)
			protected.GET("/api-keys/:id/mcp-prompts", apiKeyHandler.GetMCPPrompts)
			protected.PUT("/api-keys/:id/mcp-prompts", apiKeyHandler.UpdateMCPPrompts)

			protected.GET("/usage/logs", usageHandler.Logs)
			protected.GET("/usage/dashboard", usageHandler.Dashboard)

			protected.GET("/mcps", mcpHandler.List)
			protected.POST("/mcps", mcpHandler.Create)
			protected.GET("/mcps/:id", mcpHandler.Get)
			protected.PUT("/mcps/:id", mcpHandler.Update)
			protected.DELETE("/mcps/:id", mcpHandler.Delete)
			protected.POST("/mcps/:id/test", mcpHandler.TestConnection)
			protected.POST("/mcps/:id/sync", mcpHandler.Sync)
			protected.GET("/mcps/:id/tools", mcpHandler.ListTools)
			protected.PUT("/mcps/tools/:id", mcpHandler.UpdateTool)
			protected.GET("/mcps/:id/resources", mcpHandler.ListResources)
			protected.PUT("/mcps/resources/:id", mcpHandler.UpdateResource)
			protected.GET("/mcps/:id/prompts", mcpHandler.ListPrompts)
			protected.PUT("/mcps/prompts/:id", mcpHandler.UpdatePrompt)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": res.Version})
	})

	r.NoRoute(middleware.Static(res.WebFS))

	go func() {
		pprofAddr := fmt.Sprintf("localhost:%d", cfg.Pprof.Port)
		log.Printf("[Pprof] Performance profiling server starting on http://%s/debug/pprof/", pprofAddr)
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			log.Printf("[Pprof] Failed to start pprof server: %v", err)
		}
	}()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
