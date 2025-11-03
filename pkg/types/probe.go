package types

import (
	"time"
)

// ProbeInstance represents an individual measurement agent deployed on a specific platform/architecture
type ProbeInstance struct {
	ID              string            `json:"id"`              // Unique identifier derived from container ID + hostname + timestamp
	Platform        PlatformInfo      `json:"platform"`        // Platform and architecture information
	Status          ProbeStatus       `json:"status"`          // Current operational status
	Configuration   *Configuration    `json:"configuration"`   // Current active configuration
	Metrics         *ProbeMetrics     `json:"metrics"`         // Runtime metrics and statistics
	StartedAt       time.Time         `json:"started_at"`      // Probe startup timestamp
	LastHeartbeat   time.Time         `json:"last_heartbeat"`  // Last health check timestamp
	MeshNeighbors   []string          `json:"mesh_neighbors"`  // Discovered neighboring probes in mesh
}

// PlatformInfo represents platform and architecture details
type PlatformInfo struct {
	OS        string `json:"os"`        // "linux", "darwin", "windows"
	Arch      string `json:"arch"`      // "amd64", "arm64"
	Version   string `json:"version"`   // OS version
	Kernel    string `json:"kernel"`    // Kernel version
	Hostname  string `json:"hostname"`  // Hostname
	Container bool   `json:"container"` // Running in container
}

// ProbeStatus represents the current operational state
type ProbeStatus struct {
	State         ProbeState `json:"state"`          // Current state
	HealthScore   float64    `json:"health_score"`   // 0.0-1.0 health score
	ActiveTargets int        `json:"active_targets"` // Number of targets being measured
	Errors        []string   `json:"errors"`         // Current error messages
	Warnings      []string   `json:"warnings"`       // Current warning messages
}

// ProbeState represents possible operational states
type ProbeState string

const (
	ProbeStateStarting ProbeState = "starting"
	ProbeStateRunning  ProbeState = "running"
	ProbeStateDegraded ProbeState = "degraded"
	ProbeStateError    ProbeState = "error"
	ProbeStateStopped  ProbeState = "stopped"
)

// ProbeMetrics contains runtime performance and measurement statistics
type ProbeMetrics struct {
	MeasurementsTotal    uint64        `json:"measurements_total"`     // Total measurements performed
	MeasurementsSuccess  uint64        `json:"measurements_success"`   // Successful measurements
	MeasurementsFailed   uint64        `json:"measurements_failed"`    // Failed measurements
	ActiveConnections    int           `json:"active_connections"`     // Active ICMP connections
	Uptime              time.Duration `json:"uptime"`                 // Total uptime
	CPUUsage            float64       `json:"cpu_usage"`              // CPU usage percentage
	MemoryUsage         uint64        `json:"memory_usage"`           // Memory usage in bytes
	NetworkBytesSent    uint64        `json:"network_bytes_sent"`     // Network bytes transmitted
	NetworkBytesRecv    uint64        `json:"network_bytes_recv"`     // Network bytes received
	LastUpdate          time.Time     `json:"last_update"`            // Metrics last updated timestamp
}