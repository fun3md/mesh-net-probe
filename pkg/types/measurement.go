package types

import (
	"time"
	"net"
)

// MeasurementData contains ICMP timing information, packet statistics, target information
type MeasurementData struct {
	ID           string         `json:"id"`            // Unique measurement ID
	ProbeID      string         `json:"probe_id"`      // Probe that performed the measurement
	Target       NetworkTarget  `json:"target"`        // Target that was measured
	Timestamp    time.Time      `json:"timestamp"`     // Measurement timestamp with microsecond precision
	
	// ICMP Response Data
	RequestSent  time.Time      `json:"request_sent"`  // When ICMP request was sent
	ResponseRecv time.Time      `json:"response_recv"` // When ICMP response was received
	RTT          time.Duration  `json:"rtt"`           // Round-trip time with microsecond precision
	TTL          int            `json:"ttl"`           // Time-to-live from response
	PacketSize   int            `json:"packet_size"`   // ICMP packet size in bytes
	
	// Success/Failure Information
	Success      bool           `json:"success"`       // Whether measurement was successful
	ErrorCode    ICMPErrorCode  `json:"error_code"`    // ICMP error code if failed
	ErrorMessage string         `json:"error_message"` // Human-readable error message
	
	// Network Path Information
	SourceIP     net.IP         `json:"source_ip"`     // Source IP address used
	DestIP       net.IP         `json:"dest_ip"`       // Destination IP address
	Hops         []HopInfo      `json:"hops"`          // Network path hops (if available)
	
	// Statistical Information
	Sequence     uint16         `json:"sequence"`      // ICMP sequence number
	PacketID     uint16         `json:"packet_id"`     // ICMP packet identifier
	Checksum     uint16         `json:"checksum"`      // ICMP packet checksum
	
	// Platform-specific metadata
	PlatformData PlatformMeta   `json:"platform_data"` // Platform-specific measurement metadata
	
	// Mesh coordination data
	MeshCoord    *MeshCoordInfo `json:"mesh_coord"`    // Mesh coordination information
}

// HopInfo represents a network path hop
type HopInfo struct {
	IP       net.IP        `json:"ip"`        // Hop IP address
	Hostname string        `json:"hostname"`  // Hop hostname (if resolved)
	RTT      time.Duration `json:"rtt"`       // RTT to this hop
	TTL      int           `json:"ttl"`       // TTL when reaching this hop
}

// PlatformMeta contains platform-specific measurement metadata
type PlatformMeta struct {
	OS              string            `json:"os"`               // Operating system
	Architecture    string            `json:"architecture"`     // CPU architecture
	Precision       MeasurementPrecision `json:"precision"`    // Timing precision achieved
	Capabilities    []string          `json:"capabilities"`     // Platform capabilities
	Limitations     []string          `json:"limitations"`      // Platform limitations
	Custom          map[string]interface{} `json:"custom"`      // Platform-specific custom data
}

// MeasurementPrecision indicates the timing precision achieved
type MeasurementPrecision string

const (
	PrecisionNanosecond  MeasurementPrecision = "nanosecond"
	PrecisionMicrosecond MeasurementPrecision = "microsecond"  
	PrecisionMillisecond MeasurementPrecision = "millisecond"
	PrecisionSecond      MeasurementPrecision = "second"
)

// ICMPErrorCode represents ICMP error conditions
type ICMPErrorCode uint8

const (
	ICMPErrNoError     ICMPErrorCode = 0  // No error
	ICMPErrTimeout     ICMPErrorCode = 1  // Request timed out
	ICMPErrUnreachable ICMPErrorCode = 2  // Destination unreachable
	ICMPErrTTLExceeded ICMPErrorCode = 3  // TTL exceeded in transit
	ICMPErrParameter   ICMPErrorCode = 4  // Parameter problem
	ICMPErrRedirect    ICMPErrorCode = 5  // Redirect
	ICMPErrEchoReply   ICMPErrorCode = 6  // Echo reply received instead of request
	ICMPErrMalformed   ICMPErrorCode = 7  // Malformed packet
	ICMPErrRateLimited ICMPErrorCode = 8  // Rate limited
	ICMPErrNoRoute     ICMPErrorCode = 9  // No route to host
)

// MeshCoordInfo contains mesh coordination metadata
type MeshCoordInfo struct {
	Coordinated     bool        `json:"coordinated"`     // Whether this was a coordinated measurement
	CoordinatorID   string      `json:"coordinator_id"`  // ID of mesh coordinator
	MeasurementID   string      `json:"measurement_id"`  // Coordinated measurement ID
	Participants    []string    `json:"participants"`    // Other probes participating
	StartTime       time.Time   `json:"start_time"`      // Coordinated start time
	EndTime         time.Time   `json:"end_time"`        // Coordinated end time
	SyncPrecision   time.Duration `json:"sync_precision"` // Achieved synchronization precision
}

// MeasurementBatch represents a collection of measurements
type MeasurementBatch struct {
	ID          string           `json:"id"`           // Batch identifier
	ProbeID     string           `json:"probe_id"`     // Probe that performed measurements
	StartTime   time.Time        `json:"start_time"`   // Batch start time
	EndTime     time.Time        `json:"end_time"`     // Batch end time
	Measurements []MeasurementData `json:"measurements"` // Individual measurements
	Summary     *MeasurementSummary `json:"summary"`   // Batch summary statistics
}

// MeasurementSummary contains aggregated statistics for a measurement batch
type MeasurementSummary struct {
	TotalMeasurements    uint64          `json:"total_measurements"`    // Total measurements in batch
	SuccessfulMeasurements uint64        `json:"successful_measurements"` // Successful measurements
	FailedMeasurements   uint64          `json:"failed_measurements"`   // Failed measurements
	
	// RTT Statistics
	MinRTT           time.Duration    `json:"min_rtt"`             // Minimum RTT
	MaxRTT           time.Duration    `json:"max_rtt"`             // Maximum RTT
	AvgRTT           time.Duration    `json:"avg_rtt"`             // Average RTT
	MedianRTT        time.Duration    `json:"median_rtt"`          // Median RTT
	StdDevRTT        time.Duration    `json:"stddev_rtt"`          // Standard deviation of RTT
	
	// Loss Statistics
	PacketLossPct    float64          `json:"packet_loss_pct"`     // Packet loss percentage
	
	// Target-specific statistics
	TargetStats      map[string]*TargetStats `json:"target_stats"`   // Per-target statistics
	
	// Timing statistics
	ProcessingTime   time.Duration    `json:"processing_time"`     // Time to process batch
	ExportTime       time.Duration    `json:"export_time"`         // Time to export batch
}

// TargetStats contains statistics for a specific target
type TargetStats struct {
	TargetID         string         `json:"target_id"`           // Target identifier
	Measurements     uint64         `json:"measurements"`        // Number of measurements
	SuccessRate      float64        `json:"success_rate"`        // Success rate (0.0-1.0)
	MinRTT           time.Duration  `json:"min_rtt"`             // Minimum RTT to target
	MaxRTT           time.Duration  `json:"max_rtt"`             // Maximum RTT to target
	AvgRTT           time.Duration  `json:"avg_rtt"`             // Average RTT to target
	Jitter           time.Duration  `json:"jitter"`              // RTT jitter
	Timeouts         uint64         `json:"timeouts"`            // Number of timeouts
	Unreachable      uint64         `json:"unreachable"`         // Number of unreachable responses
}