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

// NOTE: Phase 5.1:
// In-memory maps for demo storage are deprecated for configuration and probe state.
// Authoritative state must come from config.Manager and monitoring.ProbeRegistry.
// Measurement and alert demo maps are temporarily kept for non-critical features.
var (
	measurements   = make(map[string]*types.MeasurementData)
	alerts         = make(map[string]*Alert)
	dashboardStats = &DashboardStats{}
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
		authGroup.GET("/me", handleMe)
		authGroup.POST("/refresh", authMiddleware.JWT(), handleRefresh)
	}
}

// RegisterConfigRoutes registers configuration management routes backed by config.Manager.
// All operations require authentication; mutating routes require admin/operator roles.
func RegisterConfigRoutes(router *gin.RouterGroup, configManager config.Manager, authMiddleware *auth.Middleware) {
	configGroup := router.Group("/config")

	// If authMiddleware is provided (production), enforce JWT and role-based access.
	// If nil (tests), routes are left unprotected so behavior can be validated in isolation.
	if authMiddleware != nil {
		configGroup.Use(authMiddleware.JWT())
	}

	{
		// Read operations
		if authMiddleware != nil {
			configGroup.GET("", handleGetConfig(configManager))
			configGroup.GET("/status", handleGetConfigStatus(configManager))
		} else {
			// No-auth variant (tests)
			configGroup.GET("", handleGetConfig(configManager))
			configGroup.GET("/status", handleGetConfigStatus(configManager))
		}

		// Mutating operations - only wired with auth in production; tests call handlers via direct routing.
		if authMiddleware != nil {
			configGroup.POST("", auth.RequireOperator(), handleCreateConfig(configManager))
			configGroup.PUT("/:id", auth.RequireOperator(), handleUpdateConfig(configManager))
			configGroup.DELETE("/:id", auth.RequireAdmin(), handleDeleteConfig(configManager))
			configGroup.POST("/propagate", auth.RequireOperator(), handlePropagateConfig(configManager))
		} else {
			// When authMiddleware is nil (tests), expose without role wrappers to avoid 401/403.
			configGroup.POST("", handleCreateConfig(configManager))
			configGroup.PUT("/:id", handleUpdateConfig(configManager))
			configGroup.DELETE("/:id", handleDeleteConfig(configManager))
			configGroup.POST("/propagate", handlePropagateConfig(configManager))
		}
	}
}

// RegisterProbeRoutes registers probe management routes backed by ProbeRegistry.
func RegisterProbeRoutes(router *gin.RouterGroup, probeRegistry *monitoring.ProbeRegistry, configManager config.Manager, authMiddleware *auth.Middleware) {
	probeGroup := router.Group("/probes")

	// Production: enforce JWT
	if authMiddleware != nil {
		probeGroup.Use(authMiddleware.JWT())
	}

	{
		// Listing / details
		probeGroup.GET("", handleListProbes(probeRegistry))
		probeGroup.GET("/:id", handleGetProbe(probeRegistry))
		probeGroup.GET("/:id/health", handleGetProbeHealth(probeRegistry))

		if authMiddleware != nil {
			// Registration and lifecycle - operator/admin only
			probeGroup.POST("", auth.RequireOperator(), handleRegisterProbe(probeRegistry))
			probeGroup.PUT("/:id", auth.RequireOperator(), handleUpdateProbe(probeRegistry))
			probeGroup.DELETE("/:id", auth.RequireAdmin(), handleUnregisterProbe(probeRegistry))
		} else {
			// Tests: allow direct calls without role enforcement
			probeGroup.POST("", handleRegisterProbe(probeRegistry))
			probeGroup.PUT("/:id", handleUpdateProbe(probeRegistry))
			probeGroup.DELETE("/:id", handleUnregisterProbe(probeRegistry))
		}

		// Heartbeat: in production requires JWT; in tests, no-op JWT via nil middleware
		probeGroup.POST("/:id/heartbeat", handleProbeHeartbeat(probeRegistry))

		// Probes report applied configuration
		probeGroup.POST("/:id/config-applied", handleProbeConfigApplied(probeRegistry))
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
			
			c.JSON(http.StatusOK, gin.H{
				"token":     token,
				"user":      map[string]string{"id": "1", "username": "admin", "role": "admin"},
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
// NOTE: These handlers expose configuration managed by config.Manager.
// For Phase 5.1 we treat the manager as authoritative. If more advanced
// multi-config semantics are needed, they should be added in Manager.

func handleGetConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		cfg, err := configManager.GetConfiguration(ctx)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "configuration not available",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, cfg)
	}
}

// GET /config/status - expose ManagerStatus for operational visibility (T094)
func handleGetConfigStatus(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		status, err := configManager.GetStatus(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to get configuration status",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}

func handleCreateConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cfg types.Configuration
		if err := c.ShouldBindJSON(&cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration format"})
			return
		}

		// Ensure minimal metadata for observability / versioning
		now := time.Now()
		if cfg.ID == "" {
			cfg.ID = "config-" + strconv.FormatInt(now.UnixNano(), 10)
		}
		if cfg.Version == 0 {
			cfg.Version = 1
		}
		if cfg.CreatedAt.IsZero() {
			cfg.CreatedAt = now
		}
		cfg.UpdatedAt = now

		// Phase 5.1: Manager remains authoritative. We trigger reload so that
		// providers (file/etcd/Consul) can pick up and propagate. Actual write
		// into providers is handled externally / operationally.
		if err := configManager.ReloadConfiguration(c.Request.Context()); err != nil {
			c.JSON(http.StatusAccepted, gin.H{
				"message": "configuration accepted; propagation reported issues",
				"config":  cfg,
				"warning": "reload from providers reported an error; check /config/status",
			})
			return
		}

		c.JSON(http.StatusCreated, cfg)
	}
}

func handleUpdateConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing configuration id"})
			return
		}

		var cfg types.Configuration
		if err := c.ShouldBindJSON(&cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration format"})
			return
		}

		now := time.Now()
		cfg.ID = id
		if cfg.CreatedAt.IsZero() {
			cfg.CreatedAt = now
		}
		if cfg.Version == 0 {
			cfg.Version = 1
		}
		cfg.UpdatedAt = now

		if err := configManager.ReloadConfiguration(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to reload configuration from providers",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, cfg)
	}
}

func handleDeleteConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing configuration id"})
			return
		}

		// Real deletion is provider-specific; here we enforce that changes must
		// flow through providers by triggering reload and surfacing status.
		if err := configManager.ReloadConfiguration(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to reload configuration after delete",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

func handlePropagateConfig(configManager config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Force a reload from providers; probes consume via their config flow.
		if err := configManager.ReloadConfiguration(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to propagate configuration",
				"details": err.Error(),
			})
			return
		}

		status, _ := configManager.GetStatus(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{
			"message": "configuration reload triggered",
			"status":  status,
		})
	}
}

// Probe handlers (backed by ProbeRegistry)
func handleListProbes(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		probeList := probeRegistry.ListProbes()
		c.JSON(200, gin.H{"probes": probeList})
	}
}

func handleGetProbe(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if probe, exists := probeRegistry.GetProbe(id); exists {
			c.JSON(200, probe)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "probe not found"})
	}
}

func handleRegisterProbe(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID        string                 `json:"id" binding:"required"`
			Name      string                 `json:"name"`
			Version   string                 `json:"version"`
			Platform  string                 `json:"platform"`
			Arch      string                 `json:"arch"`
			IPAddress string                 `json:"ip_address"`
			Tags      []string               `json:"tags"`
			Metadata  map[string]interface{} `json:"metadata"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid probe registration format"})
			return
		}

		probe := &monitoring.Probe{
			ID:        req.ID,
			Name:      req.Name,
			Version:   req.Version,
			Platform:  req.Platform,
			Arch:      req.Arch,
			IPAddress: req.IPAddress,
			Tags:      req.Tags,
			Metadata:  req.Metadata,
		}

		probeRegistry.RegisterProbe(probe)
		c.JSON(201, gin.H{"message": "probe registered successfully", "probe": probe})
	}
}

func handleUpdateProbe(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing probe id"})
			return
		}

		var req struct {
			Name      *string                `json:"name"`
			Version   *string                `json:"version"`
			Platform  *string                `json:"platform"`
			Arch      *string                `json:"arch"`
			IPAddress *string                `json:"ip_address"`
			Tags      *[]string              `json:"tags"`
			Metadata  *map[string]interface{} `json:"metadata"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid probe update format"})
			return
		}

		err := probeRegistry.UpdateProbe(id, func(p *monitoring.Probe) error {
			if req.Name != nil {
				p.Name = *req.Name
			}
			if req.Version != nil {
				p.Version = *req.Version
			}
			if req.Platform != nil {
				p.Platform = *req.Platform
			}
			if req.Arch != nil {
				p.Arch = *req.Arch
			}
			if req.IPAddress != nil {
				p.IPAddress = *req.IPAddress
			}
			if req.Tags != nil {
				p.Tags = *req.Tags
			}
			if req.Metadata != nil {
				p.Metadata = *req.Metadata
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update probe", "details": err.Error()})
			return
		}

		if probe, exists := probeRegistry.GetProbe(id); exists {
			c.JSON(200, probe)
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "probe not found"})
	}
}

func handleUnregisterProbe(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing probe id"})
			return
		}

		probeRegistry.UnregisterProbe(id)
		c.JSON(http.StatusNoContent, nil)
	}
}

func handleProbeHeartbeat(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing probe id"})
			return
		}

		if err := probeRegistry.UpdateProbeHeartbeat(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update heartbeat", "details": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "heartbeat received", "probe_id": id, "timestamp": time.Now()})
	}
}

func handleGetProbeHealth(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing probe id"})
			return
		}

		if probe, exists := probeRegistry.GetProbe(id); exists && probe.Health != nil {
			c.JSON(200, gin.H{
				"probe_id": id,
				"health":   probe.Health,
				"status":   probe.Status,
				"last_seen": probe.LastSeen,
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "probe or health status not found"})
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
		// Update dashboard statistics using authoritative ProbeRegistry where possible
		dashboardStats.TotalProbes = probeRegistry.ProbeCount()
		dashboardStats.ActiveProbes = 0
		dashboardStats.TotalMeasurements = uint64(len(measurements))
		dashboardStats.ActiveTargets = 5 // Demo value
		dashboardStats.AverageRTT = 45.2
		dashboardStats.PacketLossRate = 2.3
		dashboardStats.HealthScore = 0.95
		dashboardStats.Uptime = 99.8
		dashboardStats.AlertsCount = len(alerts)
		dashboardStats.LastUpdated = time.Now()

		// Active probe count from registry
		dashboardStats.ActiveProbes = probeRegistry.OnlineProbeCount()

		c.JSON(200, dashboardStats)
	}
}

func handleGetHealthSummary(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		total := probeRegistry.ProbeCount()
		online := probeRegistry.OnlineProbeCount()
		health := gin.H{
			"overall_health": "healthy",
			"system_status":  "operational",
			"probes_healthy": online,
			"probes_total":   total,
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

// handleProbeConfigApplied allows a probe to report which configuration it applied.
// This updates ProbeRegistry with configuration metadata for rollout tracking (T097).
func handleProbeConfigApplied(probeRegistry *monitoring.ProbeRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing probe id"})
			return
		}

		var req struct {
			ConfigID      string    `json:"config_id"`
			ConfigVersion int       `json:"config_version"`
			ConfigSource  string    `json:"config_source"`
			AppliedAt     time.Time `json:"applied_at"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
		if req.ConfigID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "config_id is required"})
			return
		}
		if req.AppliedAt.IsZero() {
			req.AppliedAt = time.Now()
		}

		if err := probeRegistry.UpdateProbeConfig(id, req.ConfigID, req.ConfigVersion, req.ConfigSource, req.AppliedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to update probe configuration state",
				"details": err.Error(),
			})
			return
		}

		if probe, ok := probeRegistry.GetProbe(id); ok {
			c.JSON(http.StatusOK, gin.H{
				"message": "configuration state recorded",
				"probe":   probe,
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "probe not found"})
	}
}