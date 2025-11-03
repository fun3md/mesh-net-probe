package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mesh-net-probe/probe/internal/config"
	"github.com/mesh-net-probe/probe/internal/monitoring"
	"github.com/mesh-net-probe/probe/internal/web/auth"
	types "github.com/mesh-net-probe/probe/pkg/types"
)

// Simple in-memory storage for demo (in production, use proper database)
var (
	configurations  = make(map[string]*types.Configuration)
	probes          = make(map[string]*types.ProbeInstance)
	measurements    = make(map[string]*types.MeasurementData)
	alerts          = make(map[string]*Alert)
	dashboardStats  = &DashboardStats{}
)

// Alert represents a monitoring alert
type Alert struct {
	ID          string    `json:"id"`           // Unique alert ID
	Title       string    `json:"title"`        // Alert title
	Description string    `json:"description"`  // Alert description
	Severity    string    `json:"severity"`     // "low", "medium", "high", "critical"
	Source      string    `json:"source"`       // Source (probe ID, system, etc.)
	Status      string    `json:"status"`       // "active", "acknowledged", "resolved"
	CreatedAt   time.Time `json:"created_at"`   // Alert creation time
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"` // Resolution time
}

// DashboardStats contains aggregated dashboard statistics
type DashboardStats struct {
	TotalProbes         int     `json:"total_probes"`
	ActiveProbes        int     `json:"active_probes"`
	TotalMeasurements   uint64  `json:"total_measurements"`
	ActiveTargets       int     `json:"active_targets"`
	AverageRTT          float64 `json:"average_rtt"`
	PacketLossRate      float64 `json:"packet_loss_rate"`
	HealthScore         float64 `json:"health_score"`
	Uptime              float64 `json:"uptime"`
	AlertsCount         int     `json:"alerts_count"`
	LastUpdated         time.Time `json:"last_updated"`
}

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(router *gin.RouterGroup, authMiddleware *auth.Middleware) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", handleLogin(authMiddleware))
		authGroup.POST("/logout", authMiddleware.JWT(), handleLogout)
		authGroup.GET("/me", authMiddleware.JWT(), handleMe)
		authGroup.POST("/refresh", authMiddleware.JWT(), handleRefresh)
	}
}

// RegisterConfigRoutes registers configuration management routes
func RegisterConfigRoutes(router *gin.RouterGroup, configManager config.Manager, authMiddleware *auth.Middleware) {
	configGroup := router.Group("/config")
	{
		configGroup.GET("", handleGetConfig(configManager))
		configGroup.POST("", handleCreateConfig(configManager))
		configGroup.PUT("/:id", handleUpdateConfig(configManager))
		configGroup.DELETE("/:id", handleDeleteConfig(configManager))
		configGroup.POST("/propagate", handlePropagateConfig(configManager))
	}
}

// RegisterProbeRoutes registers probe management routes
func RegisterProbeRoutes(router *gin.RouterGroup, probeRegistry *monitoring.ProbeRegistry, configManager config.Manager, authMiddleware *auth.Middleware) {
	probeGroup := router.Group("/probes")
	{
		probeGroup.GET("", handleListProbes(probeRegistry))
		probeGroup.GET("/:id", handleGetProbe(probeRegistry))
		probeGroup.POST("", handleRegisterProbe(probeRegistry))
		probeGroup.PUT("/:id", handleUpdateProbe(probeRegistry))
		probeGroup.DELETE("/:id", handleUnregisterProbe(probeRegistry))
		probeGroup.POST("/:id/heartbeat", handleProbeHeartbeat(probeRegistry))
		probeGroup.GET("/:id/health", handleGetProbeHealth(probeRegistry))
	}
}

// RegisterMeasurementRoutes registers measurement data routes
func RegisterMeasurementRoutes(router *gin.RouterGroup, monitoringMgr *monitoring.Manager, authMiddleware *auth.Middleware) {
	measurementGroup := router.Group("/measurements")
	{
		measurementGroup.GET("", handleListMeasurements(monitoringMgr))
		measurementGroup.GET("/:id", handleGetMeasurement(monitoringMgr))
		measurementGroup.POST("", handleCreateMeasurement(monitoringMgr))
		measurementGroup.GET("/stream/:probe_id", handleStreamMeasurements(monitoringMgr))
		measurementGroup.GET("/statistics", handleGetMeasurementStatistics(monitoringMgr))
	}
}

// RegisterMonitoringRoutes registers monitoring and health routes
func RegisterMonitoringRoutes(router *gin.RouterGroup, probeRegistry *monitoring.ProbeRegistry, monitoringMgr *monitoring.Manager, authMiddleware *auth.Middleware) {
	monitoringGroup := router.Group("/monitoring")
	{
		monitoringGroup.GET("/dashboard", handleGetDashboard(probeRegistry, monitoringMgr))
		monitoringGroup.GET("/health/summary", handleGetHealthSummary(probeRegistry))
		monitoringGroup.GET("/stats", handleGetMonitoringStats(monitoringMgr))
		monitoringGroup.POST("/alerts", handleCreateAlert(monitoringMgr))
		monitoringGroup.GET("/alerts", handleListAlerts(monitoringMgr))
	}
}

// Authentication handlers
func handleLogin(authMiddleware *auth.Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginReq struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Demo authentication - in production, use proper auth system
		if loginReq.Username == "admin" && loginReq.Password == "admin" {
			// Generate demo JWT token (in production, use proper JWT)
			token := "demo-jwt-token-" + strconv.FormatInt(time.Now().Unix(), 10)
			
			c.JSON(200, gin.H{
				"token":     token,
				"user":      gin.H{"id": "1", "username": "admin", "role": "admin"},
				"expiresAt": time.Now().Add(24 * time.Hour).Unix(),
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		}
	}
}

func handleLogout(c *gin.Context) {
	// In production, invalidate JWT token here
	c.JSON(200, gin.H{"message": "Logged out successfully"})
}

func handleMe(c *gin.Context) {
	// In production, extract user from JWT token
	c.JSON(200, gin.H{
		"id":       "1",
		"username": "admin",
		"email":    "admin@mesh-probe.com",
		"role":     "admin",
		"created":  time.Now().Add(-30 * 24 * time.Hour),
	})
}

func handleRefresh(c *gin.Context) {
	// Generate new token
	token := "demo-jwt-token-" + strconv.FormatInt(time.Now().Unix(), 10)
	c.JSON(200, gin.H{
		"token":     token,
		"expiresAt": time.Now().Add(24 * time.Hour).Unix(),
	})
}

// Configuration handlers
func handleGetConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			// Return all configurations
			configList := make([]*types.Configuration, 0, len(configurations))
			for _, config := range configurations {
				configList = append(configList, config)
			}
			c.JSON(200, gin.H{"configurations": configList})
			return
		}

		// Return specific configuration
		if config, exists := configurations[id]; exists {
			c.JSON(200, config)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		}
	}
}

func handleCreateConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var config types.Configuration
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration format"})
			return
		}

		config.ID = "config-" + strconv.FormatInt(time.Now().Unix(), 10)
		config.CreatedAt = time.Now()
		config.UpdatedAt = time.Now()
		config.Version = 1

		configurations[config.ID] = &config
		c.JSON(201, config)
	}
}

func handleUpdateConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var config types.Configuration
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration format"})
			return
		}

		if existing, exists := configurations[id]; exists {
			config.ID = id
			config.CreatedAt = existing.CreatedAt
			config.UpdatedAt = time.Now()
			config.Version = existing.Version + 1
			configurations[id] = &config
			c.JSON(200, config)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		}
	}
}

func handleDeleteConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if _, exists := configurations[id]; exists {
			delete(configurations, id)
			c.JSON(http.StatusNoContent, nil)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		}
	}
}

func handlePropagateConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Demo implementation - in production, propagate to all connected probes
		c.JSON(200, gin.H{
			"message":      "Configuration propagation initiated",
			"propagatedTo": len(probes),
			"status":       "in_progress",
		})
	}
}

// Probe handlers
func handleListProbes(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		probeList := make([]*types.ProbeInstance, 0, len(probes))
		for _, probe := range probes {
			probeList = append(probeList, probe)
		}
		c.JSON(200, gin.H{"probes": probeList})
	}
}

func handleGetProbe(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if probe, exists := probes[id]; exists {
			c.JSON(200, probe)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Probe not found"})
		}
	}
}

func handleRegisterProbe(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID       string `json:"id" binding:"required"`
			Platform struct {
				OS       string `json:"os"`
				Arch     string `json:"arch"`
				Version  string `json:"version"`
				Kernel   string `json:"kernel"`
				Hostname string `json:"hostname"`
			} `json:"platform"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid probe registration format"})
			return
		}

		probe := &types.ProbeInstance{
			ID: req.ID,
			Platform: types.PlatformInfo{
				OS:       req.Platform.OS,
				Arch:     req.Platform.Arch,
				Version:  req.Platform.Version,
				Kernel:   req.Platform.Kernel,
				Hostname: req.Platform.Hostname,
				Container: true,
			},
			Status: types.ProbeStatus{
				State:         types.ProbeStateRunning,
				HealthScore:   1.0,
				ActiveTargets: 0,
				Errors:        []string{},
				Warnings:      []string{},
			},
			Metrics: &types.ProbeMetrics{
				MeasurementsTotal:   0,
				MeasurementsSuccess: 0,
				MeasurementsFailed:  0,
				ActiveConnections:   0,
				Uptime:              0,
				CPUUsage:            0.0,
				MemoryUsage:         0,
				NetworkBytesSent:    0,
				NetworkBytesRecv:    0,
				LastUpdate:          time.Now(),
			},
			StartedAt:     time.Now(),
			LastHeartbeat: time.Now(),
		}

		probes[req.ID] = probe
		c.JSON(201, gin.H{"message": "Probe registered successfully", "probe": probe})
	}
}

func handleUpdateProbe(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var probe types.ProbeInstance
		if err := c.ShouldBindJSON(&probe); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid probe format"})
			return
		}

		if existing, exists := probes[id]; exists {
			probe.ID = id
			probe.StartedAt = existing.StartedAt
			probe.LastHeartbeat = time.Now()
			probes[id] = &probe
			c.JSON(200, probe)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Probe not found"})
		}
	}
}

func handleUnregisterProbe(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if _, exists := probes[id]; exists {
			delete(probes, id)
			c.JSON(http.StatusNoContent, nil)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Probe not found"})
		}
	}
}

func handleProbeHeartbeat(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if probe, exists := probes[id]; exists {
			probe.LastHeartbeat = time.Now()
			c.JSON(200, gin.H{"message": "Heartbeat received", "timestamp": probe.LastHeartbeat})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Probe not found"})
		}
	}
}

func handleGetProbeHealth(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if probe, exists := probes[id]; exists {
			health := gin.H{
				"probe_id":     id,
				"status":       probe.Status.State,
				"health_score": probe.Status.HealthScore,
				"uptime":       time.Since(probe.StartedAt).String(),
				"last_seen":    probe.LastHeartbeat,
				"metrics":      probe.Metrics,
			}
			c.JSON(200, health)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Probe not found"})
		}
	}
}

// Measurement handlers
func handleListMeasurements(monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		measurementList := make([]*types.MeasurementData, 0, len(measurements))
		for _, measurement := range measurements {
			measurementList = append(measurementList, measurement)
		}
		c.JSON(200, gin.H{"measurements": measurementList})
	}
}

func handleGetMeasurement(monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if measurement, exists := measurements[id]; exists {
			c.JSON(200, measurement)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Measurement not found"})
		}
	}
}

func handleCreateMeasurement(monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var measurement types.MeasurementData
		if err := c.ShouldBindJSON(&measurement); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid measurement format"})
			return
		}

		measurement.ID = "measurement-" + strconv.FormatInt(time.Now().Unix(), 10)
		measurement.Timestamp = time.Now()

		measurements[measurement.ID] = &measurement
		c.JSON(201, measurement)
	}
}

func handleStreamMeasurements(monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		probeID := c.Param("probe_id")
		// Demo implementation - in production, set up WebSocket or Server-Sent Events
		c.JSON(200, gin.H{
			"message":   "Measurement streaming endpoint",
			"probe_id":  probeID,
			"status":    "streaming",
			"websocket": "/ws/measurements",
		})
	}
}

func handleGetMeasurementStatistics(monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate demo statistics
		stats := gin.H{
			"total_measurements":    uint64(len(measurements)),
			"successful_measurements": uint64(len(measurements)) * 95 / 100,
			"failed_measurements":   uint64(len(measurements)) * 5 / 100,
			"average_rtt":          "45.2ms",
			"min_rtt":              "12.1ms",
			"max_rtt":              "234.7ms",
			"packet_loss_rate":     2.3,
			"targets_monitored":    5,
		}
		c.JSON(200, stats)
	}
}

// Monitoring handlers
func handleGetDashboard(probeRegistry *monitoring.ProbeRegistry, monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Update dashboard statistics
		dashboardStats.TotalProbes = len(probes)
		dashboardStats.ActiveProbes = 0
		dashboardStats.TotalMeasurements = uint64(len(measurements))
		dashboardStats.ActiveTargets = 5 // Demo value
		dashboardStats.AverageRTT = 45.2
		dashboardStats.PacketLossRate = 2.3
		dashboardStats.HealthScore = 0.95
		dashboardStats.Uptime = 99.8
		dashboardStats.AlertsCount = len(alerts)
		dashboardStats.LastUpdated = time.Now()

		// Count active probes
		for _, probe := range probes {
			if probe.Status.State == types.ProbeStateRunning {
				dashboardStats.ActiveProbes++
			}
		}

		c.JSON(200, dashboardStats)
	}
}

func handleGetHealthSummary(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := gin.H{
			"overall_health": "healthy",
			"system_status":  "operational",
			"probes_healthy": len(probes),
			"probes_total":   len(probes),
			"alerts_active":  0,
			"last_check":     time.Now(),
		}
		c.JSON(200, health)
	}
}

func handleGetMonitoringStats(monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := gin.H{
			"measurements_per_minute": 45,
			"data_processed_mb":       1024.5,
			"network_utilization":     15.2,
			"cpu_utilization":         8.7,
			"memory_utilization":      23.1,
			"disk_utilization":        12.8,
			"network_errors":          0,
			"processing_latency":      "12ms",
		}
		c.JSON(200, stats)
	}
}

func handleCreateAlert(monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description" binding:"required"`
			Severity    string `json:"severity" binding:"required"`
			Source      string `json:"source" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert format"})
			return
		}

		alert := &Alert{
			ID:          "alert-" + strconv.FormatInt(time.Now().Unix(), 10),
			Title:       req.Title,
			Description: req.Description,
			Severity:    req.Severity,
			Source:      req.Source,
			Status:      "active",
			CreatedAt:   time.Now(),
		}

		alerts[alert.ID] = alert
		c.JSON(201, alert)
	}
}

func handleListAlerts(monitoringMgr *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		alertList := make([]*Alert, 0, len(alerts))
		for _, alert := range alerts {
			alertList = append(alertList, alert)
		}
		c.JSON(200, gin.H{"alerts": alertList})
	}
}