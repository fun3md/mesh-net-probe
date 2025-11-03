package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mesh-net-probe/probe/internal/config"
	"github.com/mesh-net-probe/probe/internal/monitoring"
	"github.com/mesh-net-probe/probe/internal/web/api"
	"github.com/mesh-net-probe/probe/internal/web/auth"
	"github.com/mesh-net-probe/probe/internal/web/websocket"
	"github.com/mesh-net-probe/probe/internal/logger"
	types "github.com/mesh-net-probe/probe/pkg/types"

	"github.com/gin-gonic/gin"
)

func main() {
	// Setup logging
	logger.InitGlobalLogger(logger.InfoLevel)
	ctx := context.Background()
	
	// Initialize configuration manager
	configManager := config.NewManager()
	
	// Add file provider
	fileProvider, err := config.NewFileProvider("./config", "config.json", 30*time.Second)
	if err != nil {
		log.Fatalf("Failed to create file provider: %v", err)
	}
	
	// Add file provider to manager with options
	configManager = config.NewManager(config.WithProvider(fileProvider))
	
	// Initialize configuration manager with default config
	defaultConfig := &types.Configuration{
		ID:   "default",
		Name: "Default Configuration",
		Version: 1,
	}
	
	err = configManager.Initialize(ctx, defaultConfig)
	if err != nil {
		log.Fatalf("Failed to initialize configuration manager: %v", err)
	}
	
	// Note: etcd and consul providers would need to be implemented separately
	// For now, using file provider as primary config source

	// Initialize monitoring services
	monitoringManager := monitoring.NewManager()
	probeRegistry := monitoring.NewProbeRegistry()

	// Initialize authentication middleware
	authMiddleware := auth.NewMiddleware()

	// Initialize WebSocket service
	websocketService := websocket.NewService(probeRegistry, monitoringManager)

	// Create Gin router
	router := gin.Default()

	// Add global middleware
	router.Use(gin.Recovery())
	router.Use(authMiddleware.CORS())
	router.Use(authMiddleware.RateLimit())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		authGroup := v1.Group("/auth")
		api.RegisterAuthRoutes(authGroup, authMiddleware)

		// Configuration routes
		configGroup := v1.Group("/config")
		configGroup.Use(authMiddleware.JWT())
		api.RegisterConfigRoutes(configGroup, configManager, authMiddleware)

		// Probe management routes
		probeGroup := v1.Group("/probes")
		probeGroup.Use(authMiddleware.JWT())
		api.RegisterProbeRoutes(probeGroup, probeRegistry, configManager, authMiddleware)

		// Measurement routes
		measurementGroup := v1.Group("/measurements")
		measurementGroup.Use(authMiddleware.JWT())
		api.RegisterMeasurementRoutes(measurementGroup, monitoringManager, authMiddleware)

		// Health and monitoring routes
		monitoringGroup := v1.Group("/monitoring")
		monitoringGroup.Use(authMiddleware.JWT())
		api.RegisterMonitoringRoutes(monitoringGroup, probeRegistry, monitoringManager, authMiddleware)
	}

	// WebSocket endpoints
	router.GET("/ws/probes", websocketService.HandleProbeConnections())
	router.GET("/ws/measurements", websocketService.HandleMeasurementStreaming())
	router.GET("/ws/health", websocketService.HandleHealthUpdates())

	// Start WebSocket service in background (don't block initialization)
	go func() {
		logger.GetGlobalLogger().Info(ctx, "Starting WebSocket service...")
		if err := websocketService.Start(ctx); err != nil {
			logger.GetGlobalLogger().Error(ctx, err, "WebSocket service failed to start")
		}
		logger.GetGlobalLogger().Info(ctx, "WebSocket service started")
	}()

	// Get port from environment or use default
	port := os.Getenv("ADMIN_WEB_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.GetGlobalLogger().Info(ctx, fmt.Sprintf("Admin web interface starting on port %s", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.GetGlobalLogger().Error(ctx, err, "Admin web interface failed to start")
		}
		logger.GetGlobalLogger().Info(ctx, "Admin web interface stopped")
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.GetGlobalLogger().Info(ctx, "Shutting down admin web interface...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.GetGlobalLogger().Error(ctx, err, "Server forced to shutdown")
	}

	// Shutdown WebSocket service
	websocketService.Stop(ctx)
	
	// Shutdown monitoring manager
	monitoringManager.Stop(ctx)

	logger.GetGlobalLogger().Info(ctx, "Admin web interface shutdown complete")
}