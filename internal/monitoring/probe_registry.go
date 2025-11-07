package monitoring

import (
	"encoding/json"
	"sync"
	"time"
)

// Probe represents a registered probe in the system
type Probe struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Version   string                 `json:"version"`
	Platform  string                 `json:"platform"`
	Arch      string                 `json:"arch"`
	IPAddress string                 `json:"ip_address"`
	Tags      []string               `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata"`

	// Status and health
	Status   ProbeStatus   `json:"status"`
	LastSeen time.Time     `json:"last_seen"`
	Health   *HealthStatus `json:"health,omitempty"`

	// Lifecycle
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Configuration application tracking (Phase 5.1: T096/T097)
	ConfigID      string    `json:"config_id,omitempty"`       // last applied configuration ID
	ConfigVersion int       `json:"config_version,omitempty"`  // last applied configuration version
	ConfigSource  string    `json:"config_source,omitempty"`   // provider/source used
	ConfigApplied time.Time `json:"config_applied_at,omitempty"` // when configuration was applied
}

// ProbeStatus represents the current status of a probe
type ProbeStatus string

const (
	ProbeStatusOnline    ProbeStatus = "online"
	ProbeStatusOffline   ProbeStatus = "offline"
	ProbeStatusDegraded  ProbeStatus = "degraded"
	ProbeStatusUnknown   ProbeStatus = "unknown"
)

// HealthStatus represents the health status of a probe
type HealthStatus struct {
	OverallStatus    string                 `json:"overall_status"`
	CPUUsage         float64                `json:"cpu_usage"`
	MemoryUsage      float64                `json:"memory_usage"`
	DiskUsage        float64                `json:"disk_usage"`
	NetworkLatency   float64                `json:"network_latency_ms"`
	ICMPSuccessRate  float64                `json:"icmp_success_rate"`
	LastHealthCheck  time.Time              `json:"last_health_check"`
	Checks           map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	Message     string        `json:"message"`
	Duration    time.Duration `json:"duration_ms"`
	LastChecked time.Time     `json:"last_checked"`
}

// ProbeRegistry manages registered probes
type ProbeRegistry struct {
	mu    sync.RWMutex
	probes map[string]*Probe
}

// NewProbeRegistry creates a new probe registry
func NewProbeRegistry() *ProbeRegistry {
	return &ProbeRegistry{
		probes: make(map[string]*Probe),
	}
}

// RegisterProbe registers a new probe
func (r *ProbeRegistry) RegisterProbe(probe *Probe) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	probe.CreatedAt = time.Now()
	probe.UpdatedAt = time.Now()
	probe.LastSeen = time.Now()
	probe.Status = ProbeStatusOnline
	
	r.probes[probe.ID] = probe
}

// UpdateProbe updates an existing probe
func (r *ProbeRegistry) UpdateProbe(probeID string, updateFunc func(*Probe) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	probe, exists := r.probes[probeID]
	if !exists {
		return nil // Ignore updates for non-existent probes
	}
	
	if err := updateFunc(probe); err != nil {
		return err
	}
	
	probe.UpdatedAt = time.Now()
	probe.LastSeen = time.Now()
	return nil
}

// UnregisterProbe removes a probe from the registry
func (r *ProbeRegistry) UnregisterProbe(probeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.probes, probeID)
}

// GetProbe returns a probe by ID
func (r *ProbeRegistry) GetProbe(probeID string) (*Probe, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	probe, exists := r.probes[probeID]
	return probe, exists
}

// ListProbes returns all registered probes
func (r *ProbeRegistry) ListProbes() []*Probe {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	probes := make([]*Probe, 0, len(r.probes))
	for _, probe := range r.probes {
		probes = append(probes, probe)
	}
	
	return probes
}

// UpdateProbeHeartbeat updates the last seen time for a probe
func (r *ProbeRegistry) UpdateProbeHeartbeat(probeID string) error {
	return r.UpdateProbe(probeID, func(probe *Probe) error {
		probe.LastSeen = time.Now()
		if probe.Status == ProbeStatusOffline {
			probe.Status = ProbeStatusOnline
		}
		return nil
	})
}

// GetProbesByStatus returns probes filtered by status
func (r *ProbeRegistry) GetProbesByStatus(status ProbeStatus) []*Probe {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*Probe
	for _, probe := range r.probes {
		if probe.Status == status {
			result = append(result, probe)
		}
	}
	
	return result
}

// UpdateProbeHealth updates the health status of a probe
func (r *ProbeRegistry) UpdateProbeHealth(probeID string, health *HealthStatus) error {
	return r.UpdateProbe(probeID, func(probe *Probe) error {
		probe.Health = health
		return nil
	})
}

// CleanupOfflineProbes removes probes that haven't been seen recently
func (r *ProbeRegistry) CleanupOfflineProbes(timeout time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	now := time.Now()
	var removedCount int
	
	for probeID, probe := range r.probes {
		if now.Sub(probe.LastSeen) > timeout && probe.Status == ProbeStatusOffline {
			delete(r.probes, probeID)
			removedCount++
		} else if now.Sub(probe.LastSeen) > timeout && probe.Status != ProbeStatusOffline {
			// Mark as offline if not seen recently but still in registry
			probe.Status = ProbeStatusOffline
		}
	}
	
	return removedCount
}

// ToJSON converts the registry to JSON
func (r *ProbeRegistry) ToJSON() (string, error) {
	probes := r.ListProbes()
	data, err := json.MarshalIndent(probes, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ProbeCount returns the total number of registered probes
func (r *ProbeRegistry) ProbeCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.probes)
}

// OnlineProbeCount returns the number of online probes
func (r *ProbeRegistry) OnlineProbeCount() int {
	return len(r.GetProbesByStatus(ProbeStatusOnline))
}

// UpdateProbeMetadata updates the metadata for a probe
func (r *ProbeRegistry) UpdateProbeMetadata(probeID string, metadata map[string]interface{}) error {
	return r.UpdateProbe(probeID, func(probe *Probe) error {
		probe.Metadata = metadata
		return nil
	})
}

// UpdateProbeConfig records configuration application metadata for a probe.
// Used by /probes/:id/config-applied endpoint for rollout tracking.
func (r *ProbeRegistry) UpdateProbeConfig(
	probeID string,
	configID string,
	configVersion int,
	configSource string,
	appliedAt time.Time,
) error {
	return r.UpdateProbe(probeID, func(probe *Probe) error {
		if probe.Metadata == nil {
			probe.Metadata = make(map[string]interface{})
		}
		probe.ConfigID = configID
		probe.ConfigVersion = configVersion
		probe.ConfigSource = configSource
		probe.ConfigApplied = appliedAt

		// Also mirror into metadata for easier querying/export if needed.
		probe.Metadata["config_id"] = configID
		probe.Metadata["config_version"] = configVersion
		probe.Metadata["config_source"] = configSource
		probe.Metadata["config_applied_at"] = appliedAt

		return nil
	})
}