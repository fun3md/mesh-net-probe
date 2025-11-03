package types

import (
	"net"
	"time"
)

// Configuration defines probe behavior including target lists, measurement intervals, reporting settings
type Configuration struct {
	ID          string            `json:"id"`           // Configuration profile ID
	Name        string            `json:"name"`         // Human-readable name
	Version     int               `json:"version"`      // Configuration version number
	CreatedAt   time.Time         `json:"created_at"`   // Creation timestamp
	UpdatedAt   time.Time         `json:"updated_at"`   // Last update timestamp
	
	// Measurement Configuration
	Targets     []NetworkTarget   `json:"targets"`      // Measurement targets
	Intervals   *MeasurementIntervals `json:"intervals"` // Measurement timing settings
	
	// Network Configuration  
	Network     *NetworkConfig    `json:"network"`      // Network and connectivity settings
	
	// Telemetry Configuration
	Telemetry   *TelemetryConfig  `json:"telemetry"`    // OpenTelemetry and logging settings
	
	// Authentication & Security
	Auth        *AuthConfig       `json:"auth"`         // Authentication configuration
	
	// Mesh Configuration
	Mesh        *MeshConfig       `json:"mesh"`         // Multi-probe coordination settings
}

// NetworkTarget represents destinations for ICMP measurement
type NetworkTarget struct {
	// Basic Identification
	ID          string        `json:"id"`           // Unique target identifier
	DisplayName string        `json:"display_name"` // Human-readable name
	Address     net.IP        `json:"address"`      // IP address for measurement
	Port        int           `json:"port"`         // Port (0 for ICMP, >0 for other protocols)
	
	// Operational Status
	Enabled     bool          `json:"enabled"`      // Whether target is active
	Priority    int           `json:"priority"`     // Measurement priority (1-10, 10 highest)
	
	// Timing Configuration
	Timeout     time.Duration `json:"timeout"`      // Individual timeout for this target
	Interval    time.Duration `json:"interval"`     // Measurement interval for this target
	Jitter      float64       `json:"jitter"`       // Random jitter factor (0.0-1.0)
	
	// Expected Performance
	ExpectedRTT time.Duration `json:"expected_rtt"` // Expected round-trip time
	MinRTT      time.Duration `json:"min_rtt"`      // Minimum acceptable RTT
	MaxRTT      time.Duration `json:"max_rtt"`      // Maximum acceptable RTT
	MaxLossPct  float64       `json:"max_loss_pct"` // Maximum acceptable packet loss percentage
	
	// Metadata
	Tags        []string      `json:"tags"`         // Target tags for organization
	Labels      map[string]string `json:"labels"`   // Key-value labels
	Region      string        `json:"region"`       // Geographic region
	Environment string        `json:"environment"`  // Environment (prod, staging, dev)
	
	// Performance History
	History     *TargetHistory `json:"history"`     // Historical performance data
	LastSeen    time.Time      `json:"last_seen"`   // Last successful measurement
	CreatedAt   time.Time      `json:"created_at"`  // Target creation timestamp
	UpdatedAt   time.Time      `json:"updated_at"`  // Last update timestamp
}

// TargetHistory contains historical performance data
type TargetHistory struct {
	MeasurementsCount uint64           `json:"measurements_count"` // Total measurements performed
	SuccessRate       float64          `json:"success_rate"`       // Historical success rate
	AvgRTT            time.Duration    `json:"avg_rtt"`            // Historical average RTT
	P95RTT            time.Duration    `json:"p95_rtt"`            // 95th percentile RTT
	P99RTT            time.Duration    `json:"p99_rtt"`            // 99th percentile RTT
	PacketLoss        float64          `json:"packet_loss"`        // Historical packet loss
	Outages           []TargetOutage   `json:"outages"`            // Historical outage periods
	Performance       map[string]float64 `json:"performance"`      // Other performance metrics
}

// TargetOutage represents a period of target unavailability
type TargetOutage struct {
	StartTime      time.Time      `json:"start_time"`      // Outage start time
	EndTime        time.Time      `json:"end_time"`        // Outage end time
	Duration       time.Duration  `json:"duration"`        // Outage duration
	Cause          string         `json:"cause"`           // Outage cause
	Severity       OutageSeverity `json:"severity"`        // Outage severity level
	Resolution     string         `json:"resolution"`      // How outage was resolved
}

// OutageSeverity represents the severity level of an outage
type OutageSeverity string

const (
	OutageMinor    OutageSeverity = "minor"    // Minor performance degradation
	OutageModerate OutageSeverity = "moderate" // Moderate performance impact
	OutageMajor    OutageSeverity = "major"    // Significant performance impact
	OutageCritical OutageSeverity = "critical" // Complete unavailability
)

// MeasurementIntervals defines timing for measurements
type MeasurementIntervals struct {
	DefaultInterval time.Duration `json:"default_interval"` // Default interval between measurements
	MinInterval     time.Duration `json:"min_interval"`     // Minimum interval (rate limiting)
	MaxInterval     time.Duration `json:"max_interval"`     // Maximum interval (degraded mode)
	BackoffMultiplier float64    `json:"backoff_multiplier"` // Exponential backoff multiplier
}

// NetworkConfig contains network connectivity settings
type NetworkConfig struct {
	SourceIP      net.IP        `json:"source_ip"`       // Source IP address (auto-detect if empty)
	Interface     string        `json:"interface"`       // Network interface name
	TTL           int           `json:"ttl"`             // Time-to-live for ICMP packets
	BufferSize    int           `json:"buffer_size"`     // Send/receive buffer size
	BindPort      int           `json:"bind_port"`       // Local port to bind (0 for any)
	QoS           string        `json:"qos"`             // Quality of Service setting
	DSCP          int           `json:"dscp"`            // DSCP value for traffic marking
}

// TelemetryConfig defines OpenTelemetry and logging settings
type TelemetryConfig struct {
	OTLPEndpoint    string            `json:"otlp_endpoint"`     // OpenTelemetry Collector endpoint
	OTLPGateway     string            `json:"otlp_gateway"`      // OTLP/gRPC gateway URL
	EnableTracing   bool              `json:"enable_tracing"`    // Enable distributed tracing
	EnableMetrics   bool              `json:"enable_metrics"`    // Enable metrics collection
	EnableLogging   bool              `json:"enable_logging"`    // Enable structured logging
	LogLevel        string            `json:"log_level"`         // Log level (debug, info, warn, error)
	LogFormat       string            `json:"log_format"`        // Log format (json, text)
	ExportInterval  time.Duration     `json:"export_interval"`   // How often to export data
	BufferSize      int               `json:"buffer_size"`       // In-memory buffer size for offline scenarios
	CustomHeaders   map[string]string `json:"custom_headers"`    // Custom HTTP headers for export
}

// AuthConfig contains authentication and security settings
type AuthConfig struct {
	Provider       string            `json:"provider"`        // Authentication provider (etcd, consul, etc.)
	APIKey         string            `json:"api_key"`         // API key for configuration management
	Token          string            `json:"token"`           // JWT or bearer token
	CertFile       string            `json:"cert_file"`       // Client certificate file path
	KeyFile        string            `json:"key_file"`        // Client key file path
	CACertFile     string            `json:"ca_cert_file"`    // CA certificate file path
	TLSConfig      *TLSConfig        `json:"tls_config"`      // TLS configuration
	Timeout        time.Duration     `json:"timeout"`         // Authentication timeout
	RetryAttempts  int               `json:"retry_attempts"`  // Number of retry attempts
}

// TLSConfig defines TLS security settings
type TLSConfig struct {
	Enabled         bool     `json:"enabled"`          // Enable TLS
	MinVersion      string   `json:"min_version"`      // Minimum TLS version
	MaxVersion      string   `json:"max_version"`      // Maximum TLS version
	CipherSuites    []string `json:"cipher_suites"`    // Allowed cipher suites
	SkipVerify      bool     `json:"skip_verify"`      // Skip certificate verification
	ServerName      string   `json:"server_name"`      // Server name for SNI
}

// MeshConfig defines multi-probe coordination settings
type MeshConfig struct {
	Enabled         bool              `json:"enabled"`             // Enable mesh networking
	DiscoveryMethod string            `json:"discovery_method"`    // "multicast", "broadcast", "manual"
	DiscoveryPort   int               `json:"discovery_port"`      // Port for probe discovery
	HeartbeatInterval time.Duration   `json:"heartbeat_interval"`  // Interval between heartbeats
	MaxNeighbors    int               `json:"max_neighbors"`       // Maximum number of mesh neighbors
	CoordinatorID   string            `json:"coordinator_id"`      // Mesh coordinator probe ID
	SyncInterval    time.Duration     `json:"sync_interval"`       // Configuration sync interval
	AlertBroadcast  bool              `json:"alert_broadcast"`     // Broadcast alerts to mesh
	TaskDistribution bool             `json:"task_distribution"`   // Enable workload distribution
}