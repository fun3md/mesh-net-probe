package main

import (
	"context"
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
	"github.com/mesh-net-probe/probe/internal/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Setup logging
	logger.InitGlobalLogger(logger.InfoLevel)
	
	// Get port from environment or use default
	port := os.Getenv("ADMIN_WEB_PORT")
	if port == "" {
		port = "8080"
	}
	
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Initialize system components
	configManager := initConfigManager()
	monitoringMgr := initMonitoringManager()
	probeRegistry := initProbeRegistry()
	authMiddleware := initAuthMiddleware()
	
	// Create HTTP server with router
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: createRouter(configManager, monitoringMgr, probeRegistry, authMiddleware),
	}
	
	// Start server in background
	go func() {
		log.Printf("Admin web interface starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
	
	// Wait for shutdown signal
	<-sigChan
	log.Printf("Received shutdown signal")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	log.Printf("Shutting down gracefully...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	
	log.Printf("Admin web interface shutdown complete")
}

// initialize configuration manager
func initConfigManager() config.Manager {
	// Create a simple manager without providers for demo purposes
	return config.NewManager()
}

// initialize monitoring manager
func initMonitoringManager() *monitoring.Manager {
	return monitoring.NewManager()
}

// initialize probe registry
func initProbeRegistry() *monitoring.ProbeRegistry {
	return monitoring.NewProbeRegistry()
}

// initialize authentication middleware
func initAuthMiddleware() *auth.Middleware {
	return auth.NewMiddleware()
}

// createRouter sets up the Gin router with all routes
func createRouter(configManager config.Manager, monitoringMgr *monitoring.Manager, probeRegistry *monitoring.ProbeRegistry, authMiddleware *auth.Middleware) *gin.Engine {
	// Create Gin router
	router := gin.Default()
	
	// Add global middleware
	router.Use(gin.Recovery())
	
	// CORS middleware for frontend integration
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		log.Printf("Health check requested from %s", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC(),
			"service": "admin-web",
			"version": "1.0.0",
		})
	})
	
	// Metrics endpoint for Prometheus scraping
	router.GET("/metrics", func(c *gin.Context) {
		log.Printf("Metrics requested from %s", c.ClientIP())
		// Basic metrics for Prometheus - in production, use proper metrics library
		metrics := `# HELP admin_web_requests_total Total number of requests
# TYPE admin_web_requests_total counter
admin_web_requests_total{endpoint="health"} 1
admin_web_requests_total{endpoint="metrics"} 1
admin_web_requests_total{endpoint="api_v1_auth_login"} 5

# HELP admin_web_uptime_seconds Uptime in seconds
# TYPE admin_web_uptime_seconds counter
admin_web_uptime_seconds 1234

# HELP admin_web_active_connections Active connections
# TYPE admin_web_active_connections gauge
admin_web_active_connections 3
`
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, metrics)
	})
	
	// Register all API routes
	apiRouter := router.Group("/api/v1")
	{
		// Authentication routes
		api.RegisterAuthRoutes(apiRouter, authMiddleware)
		
		// Configuration routes
		api.RegisterConfigRoutes(apiRouter, configManager, authMiddleware)
		
		// Probe routes
		api.RegisterProbeRoutes(apiRouter, probeRegistry, configManager, authMiddleware)
		
		// Measurement routes
		api.RegisterMeasurementRoutes(apiRouter, monitoringMgr, authMiddleware)
		
		// Monitoring routes
		api.RegisterMonitoringRoutes(apiRouter, probeRegistry, monitoringMgr, authMiddleware)
	}
	
	// WebSocket endpoints (basic implementation for future development)
	ws := router.Group("/ws")
	{
		ws.GET("/test", func(c *gin.Context) {
			log.Printf("WebSocket test requested from %s", c.ClientIP())
			c.JSON(http.StatusOK, gin.H{
				"message": "WebSocket endpoint - implementation pending",
				"status":  "ok",
				"endpoints": map[string]string{
					"probe_updates": "/ws/probe-updates",
					"measurements": "/ws/measurements",
					"health_status": "/ws/health-status",
				},
			})
		})
		
		// Future WebSocket implementations
		ws.GET("/probe-updates", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Real-time probe updates - future implementation",
				"protocol": "WebSocket",
			})
		})
		
		ws.GET("/measurements", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Real-time measurements - future implementation",
				"protocol": "WebSocket",
			})
		})
		
		ws.GET("/health-status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Real-time health status - future implementation",
				"protocol": "WebSocket",
			})
		})
	}
	
	return router
}