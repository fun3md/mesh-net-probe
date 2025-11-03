package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("ADMIN_WEB_PORT")
	if port == "" {
		port = "8080"
	}
	
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Create HTTP server with router
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: createRouter(),
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

// createRouter sets up the Gin router with all routes
func createRouter() *gin.Engine {
	// Create Gin router
	router := gin.Default()
	
	// Add global middleware
	router.Use(gin.Recovery())
	
	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		log.Printf("Health check requested from %s", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC(),
		})
	})
	
	// Simple API endpoints
	api := router.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) {
				log.Printf("Login requested from %s", c.ClientIP())
				c.JSON(http.StatusOK, gin.H{
					"message": "Login endpoint - implementation pending",
					"status":  "ok",
				})
			})
		}
		
		// Configuration routes (simple, no auth)
		config := api.Group("/config")
		{
			config.GET("", func(c *gin.Context) {
				log.Printf("Config requested from %s", c.ClientIP())
				c.JSON(http.StatusOK, gin.H{
					"message": "Config endpoint - implementation pending",
					"status":  "ok",
				})
			})
		}
		
		// Probe routes (simple, no auth)
		probes := api.Group("/probes")
		{
			probes.GET("", func(c *gin.Context) {
				log.Printf("Probes requested from %s", c.ClientIP())
				c.JSON(http.StatusOK, gin.H{
					"probes": []string{},
					"status": "ok",
				})
			})
		}
		
		// Measurement routes (simple, no auth)
		measurements := api.Group("/measurements")
		{
			measurements.GET("", func(c *gin.Context) {
				log.Printf("Measurements requested from %s", c.ClientIP())
				c.JSON(http.StatusOK, gin.H{
					"measurements": []string{},
					"status":       "ok",
				})
			})
		}
		
		// Monitoring routes (simple, no auth)
		monitoring := api.Group("/monitoring")
		{
			monitoring.GET("/health", func(c *gin.Context) {
				log.Printf("Monitoring health requested from %s", c.ClientIP())
				c.JSON(http.StatusOK, gin.H{
					"status": "ok",
					"health": "healthy",
				})
			})
		}
	}
	
	// WebSocket endpoints (basic implementation)
	ws := router.Group("/ws")
	{
		ws.GET("/test", func(c *gin.Context) {
			log.Printf("WebSocket test requested from %s", c.ClientIP())
			c.JSON(http.StatusOK, gin.H{
				"message": "WebSocket endpoint - implementation pending",
				"status":  "ok",
			})
		})
	}
	
	return router
}