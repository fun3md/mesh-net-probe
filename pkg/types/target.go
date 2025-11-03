package types

import (
	"time"
)

// TargetGroup represents a logical grouping of targets
type TargetGroup struct {
	ID          string       `json:"id"`           // Group identifier
	Name        string       `json:"name"`         // Human-readable name
	Description string       `json:"description"`  // Group description
	Targets     []string     `json:"targets"`      // Target IDs in this group
	Tags        []string     `json:"tags"`         // Group tags
	Labels      map[string]string `json:"labels"` // Group labels
	
	// Configuration inheritance
	InheritConfig bool       `json:"inherit_config"` // Whether targets inherit group config
	GroupConfig   *Configuration `json:"group_config"` // Shared configuration for group
	
	// Statistics
	Statistics   *GroupStatistics `json:"statistics"` // Group-level statistics
	
	CreatedAt    time.Time   `json:"created_at"`   // Creation timestamp
	UpdatedAt    time.Time   `json:"updated_at"`   // Last update timestamp
}

// GroupStatistics contains aggregated statistics for a target group
type GroupStatistics struct {
	TotalTargets      int                 `json:"total_targets"`        // Total targets in group
	ActiveTargets     int                 `json:"active_targets"`       // Currently active targets
	TotalMeasurements uint64              `json:"total_measurements"`   // Total measurements for group
	SuccessRate       float64             `json:"success_rate"`         // Group success rate
	AvgRTT            time.Duration       `json:"avg_rtt"`              // Group average RTT
	MinRTT            time.Duration       `json:"min_rtt"`              // Group minimum RTT
	MaxRTT            time.Duration       `json:"max_rtt"`              // Group maximum RTT
	PacketLoss        float64             `json:"packet_loss"`          // Group packet loss
	TargetStats       map[string]*TargetStats `json:"target_stats"`   // Individual target stats
	TrendAnalysis     *TrendAnalysis      `json:"trend_analysis"`      // Performance trend analysis
}

// TrendAnalysis contains performance trend information
type TrendAnalysis struct {
	Period           string             `json:"period"`            // Analysis period (24h, 7d, 30d)
	Direction        TrendDirection     `json:"direction"`         // Overall trend direction
	ChangePercentage float64            `json:"change_percentage"` // Percentage change
	Confidence       float64            `json:"confidence"`        // Analysis confidence (0.0-1.0)
	Anomalies        []PerformanceAnomaly `json:"anomalies"`       // Detected anomalies
	Forecasts        map[string]float64 `json:"forecasts"`         // Performance forecasts
}

// TrendDirection represents the direction of performance trends
type TrendDirection string

const (
	TrendImproving   TrendDirection = "improving"   // Performance is improving
	TrendStable      TrendDirection = "stable"      // Performance is stable
	TrendDegrading   TrendDirection = "degrading"   // Performance is degrading
	TrendVolatile    TrendDirection = "volatile"    // Performance is volatile
)

// PerformanceAnomaly represents a detected performance anomaly
type PerformanceAnomaly struct {
	Timestamp   time.Time       `json:"timestamp"`   // When anomaly was detected
	Type        AnomalyType     `json:"type"`        // Type of anomaly
	Severity    AnomalySeverity `json:"severity"`    // Severity level
	Description string          `json:"description"` // Anomaly description
	Metrics     map[string]float64 `json:"metrics"`  // Affected metrics
	AffectedTargets []string     `json:"affected_targets"` // Targets affected
}

// AnomalyType represents different types of performance anomalies
type AnomalyType string

const (
	AnomalyLatencySpike     AnomalyType = "latency_spike"     // Sudden latency increase
	AnomalyPacketLoss       AnomalyType = "packet_loss"       // Increased packet loss
	AnomalyTimeout          AnomalyType = "timeout"           // Increased timeouts
	AnomalyAvailability     AnomalyType = "availability"      // Target unavailable
	AnomalyJitter           AnomalyType = "jitter"           // Increased jitter
	AnomalyThroughput       AnomalyType = "throughput"       // Throughput degradation
	AnomalyCorrelation      AnomalyType = "correlation"      // Correlation anomaly
)

// AnomalySeverity represents the severity of a performance anomaly
type AnomalySeverity string

const (
	AnomalyInfo    AnomalySeverity = "info"     // Informational
	AnomalyWarning AnomalySeverity = "warning"  // Warning level
	AnomalyError   AnomalySeverity = "error"    // Error level
	AnomalyCritical AnomalySeverity = "critical" // Critical level
)