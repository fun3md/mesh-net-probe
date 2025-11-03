package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mesh-net-probe/probe/internal/monitoring"
	"github.com/mesh-net-probe/probe/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Service handles WebSocket connections for real-time updates
type Service struct {
	probeRegistry   *monitoring.ProbeRegistry
	monitoringMgr   *monitoring.Manager
	logger          logger.Logger
	upgrader        websocket.Upgrader
	probeConnections map[*websocket.Conn]bool
	measurementConnections map[*websocket.Conn]bool
	healthConnections map[*websocket.Conn]bool
	mu             sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeProbeUpdate    MessageType = "probe_update"
	MessageTypeMeasurement    MessageType = "measurement"
	MessageTypeHealthStatus   MessageType = "health_status"
	MessageTypeConfigChange   MessageType = "config_change"
	MessageTypeAlert          MessageType = "alert"
)

// NewService creates a new WebSocket service
func NewService(probeRegistry *monitoring.ProbeRegistry, monitoringMgr *monitoring.Manager) *Service {
	return &Service{
		probeRegistry:       probeRegistry,
		monitoringMgr:       monitoringMgr,
		logger:              logger.GetGlobalLogger(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, implement proper origin checking
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		probeConnections:     make(map[*websocket.Conn]bool),
		measurementConnections: make(map[*websocket.Conn]bool),
		healthConnections:    make(map[*websocket.Conn]bool),
	}
}

// Start starts the WebSocket service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info(ctx, "Starting WebSocket service")
	
	// Start background goroutines for broadcasting updates
	go s.broadcastProbeUpdates(ctx)
	go s.broadcastMeasurementUpdates(ctx)
	go s.broadcastHealthUpdates(ctx)
	
	return nil
}

// Stop stops the WebSocket service
func (s *Service) Stop(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Close all connections
	for conn := range s.probeConnections {
		conn.Close()
	}
	for conn := range s.measurementConnections {
		conn.Close()
	}
	for conn := range s.healthConnections {
		conn.Close()
	}
	
	s.probeConnections = make(map[*websocket.Conn]bool)
	s.measurementConnections = make(map[*websocket.Conn]bool)
	s.healthConnections = make(map[*websocket.Conn]bool)
	
	s.logger.Info(ctx, "WebSocket service stopped")
}

// HandleProbeConnections handles WebSocket connections for probe updates
func (s *Service) HandleProbeConnections() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			s.logger.Error(c.Request.Context(), err, "Failed to upgrade WebSocket connection")
			return
		}
		defer conn.Close()

		// Register connection
		s.mu.Lock()
		s.probeConnections[conn] = true
		s.mu.Unlock()

		s.logger.Info(c.Request.Context(), "WebSocket connection opened for probe updates")

		// Send initial probe list
		s.sendProbeList(conn)

		// Handle connection
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		// Unregister connection
		s.mu.Lock()
		delete(s.probeConnections, conn)
		s.mu.Unlock()

		s.logger.Info(c.Request.Context(), "WebSocket connection closed for probe updates")
	}
}

// HandleMeasurementStreaming handles WebSocket connections for measurement streaming
func (s *Service) HandleMeasurementStreaming() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			s.logger.Error(c.Request.Context(), err, "Failed to upgrade WebSocket connection")
			return
		}
		defer conn.Close()

		// Register connection
		s.mu.Lock()
		s.measurementConnections[conn] = true
		s.mu.Unlock()

		s.logger.Info(c.Request.Context(), "WebSocket connection opened for measurement streaming")

		// Handle connection
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		// Unregister connection
		s.mu.Lock()
		delete(s.measurementConnections, conn)
		s.mu.Unlock()

		s.logger.Info(c.Request.Context(), "WebSocket connection closed for measurement streaming")
	}
}

// HandleHealthUpdates handles WebSocket connections for health updates
func (s *Service) HandleHealthUpdates() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			s.logger.Error(c.Request.Context(), err, "Failed to upgrade WebSocket connection")
			return
		}
		defer conn.Close()

		// Register connection
		s.mu.Lock()
		s.healthConnections[conn] = true
		s.mu.Unlock()

		s.logger.Info(c.Request.Context(), "WebSocket connection opened for health updates")

		// Handle connection
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		// Unregister connection
		s.mu.Lock()
		delete(s.healthConnections, conn)
		s.mu.Unlock()

		s.logger.Info(c.Request.Context(), "WebSocket connection closed for health updates")
	}
}

// sendProbeList sends the current list of probes to a connection
func (s *Service) sendProbeList(conn *websocket.Conn) {
	probes := s.probeRegistry.ListProbes()
	
	msg := Message{
		Type:      MessageTypeProbeUpdate,
		Data:      probes,
		Timestamp: time.Now(),
	}
	
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send probe list: %v", err)
	}
}

// broadcastProbeUpdates broadcasts probe updates to all connected clients
func (s *Service) broadcastProbeUpdates(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendProbeUpdateToAll()
		}
	}
}

// sendProbeUpdateToAll sends probe updates to all connected clients
func (s *Service) sendProbeUpdateToAll() {
	probes := s.probeRegistry.ListProbes()
	
	msg := Message{
		Type:      MessageTypeProbeUpdate,
		Data:      probes,
		Timestamp: time.Now(),
	}
	
	s.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(s.probeConnections))
	for conn := range s.probeConnections {
		connections = append(connections, conn)
	}
	s.mu.RUnlock()
	
	for _, conn := range connections {
		if err := conn.WriteJSON(msg); err != nil {
			// Connection may be closed, remove it
			s.mu.Lock()
			delete(s.probeConnections, conn)
			s.mu.Unlock()
		}
	}
}

// broadcastMeasurementUpdates broadcasts measurement updates to all connected clients
func (s *Service) broadcastMeasurementUpdates(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendMeasurementUpdatesToAll()
		}
	}
}

// sendMeasurementUpdatesToAll sends measurement updates to all connected clients
func (s *Service) sendMeasurementUpdatesToAll() {
	activeStreams := s.monitoringMgr.GetAllActiveStreams()
	
	for _, stream := range activeStreams {
		msg := Message{
			Type:      MessageTypeMeasurement,
			Data:      stream,
			Timestamp: time.Now(),
		}
		
		s.mu.RLock()
		connections := make([]*websocket.Conn, 0, len(s.measurementConnections))
		for conn := range s.measurementConnections {
			connections = append(connections, conn)
		}
		s.mu.RUnlock()
		
		for _, conn := range connections {
			if err := conn.WriteJSON(msg); err != nil {
				// Connection may be closed, remove it
				s.mu.Lock()
				delete(s.measurementConnections, conn)
				s.mu.Unlock()
			}
		}
	}
}

// broadcastHealthUpdates broadcasts health updates to all connected clients
func (s *Service) broadcastHealthUpdates(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendHealthUpdatesToAll()
		}
	}
}

// sendHealthUpdatesToAll sends health updates to all connected clients
func (s *Service) sendHealthUpdatesToAll() {
	probes := s.probeRegistry.ListProbes()
	
	// Create health summary
	healthSummary := map[string]interface{}{
		"total_probes":    len(probes),
		"online_probes":   len(s.probeRegistry.GetProbesByStatus(monitoring.ProbeStatusOnline)),
		"offline_probes":  len(s.probeRegistry.GetProbesByStatus(monitoring.ProbeStatusOffline)),
		"degraded_probes": len(s.probeRegistry.GetProbesByStatus(monitoring.ProbeStatusDegraded)),
		"timestamp":       time.Now(),
	}
	
	msg := Message{
		Type:      MessageTypeHealthStatus,
		Data:      healthSummary,
		Timestamp: time.Now(),
	}
	
	s.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(s.healthConnections))
	for conn := range s.healthConnections {
		connections = append(connections, conn)
	}
	s.mu.RUnlock()
	
	for _, conn := range connections {
		if err := conn.WriteJSON(msg); err != nil {
			// Connection may be closed, remove it
			s.mu.Lock()
			delete(s.healthConnections, conn)
			s.mu.Unlock()
		}
	}
}

// SendAlert sends an alert to all connected clients
func (s *Service) SendAlert(alertType, message string, details map[string]interface{}) {
	msg := Message{
		Type:      MessageTypeAlert,
		Data: map[string]interface{}{
			"alert_type": alertType,
			"message":    message,
			"details":    details,
		},
		Timestamp: time.Now(),
	}
	
	s.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(s.probeConnections))
	for conn := range s.probeConnections {
		connections = append(connections, conn)
	}
	for conn := range s.measurementConnections {
		connections = append(connections, conn)
	}
	for conn := range s.healthConnections {
		connections = append(connections, conn)
	}
	s.mu.RUnlock()
	
	for _, conn := range connections {
		if err := conn.WriteJSON(msg); err != nil {
			// Remove closed connections
			s.mu.Lock()
			delete(s.probeConnections, conn)
			delete(s.measurementConnections, conn)
			delete(s.healthConnections, conn)
			s.mu.Unlock()
		}
	}
}

// ConnectionCount returns the number of active WebSocket connections
func (s *Service) ConnectionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return len(s.probeConnections) + len(s.measurementConnections) + len(s.healthConnections)
}