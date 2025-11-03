package monitoring

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Measurement represents a network measurement
type Measurement struct {
	ID          string                 `json:"id"`
	ProbeID     string                 `json:"probe_id"`
	Target      string                 `json:"target"`
	Type        MeasurementType        `json:"type"`
	Status      MeasurementStatus      `json:"status"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Statistics  *MeasurementStatistics `json:"statistics,omitempty"`
}

// MeasurementType represents the type of measurement
type MeasurementType string

const (
	MeasurementTypeICMP     MeasurementType = "icmp"
	MeasurementTypeLatency  MeasurementType = "latency"
	MeasurementTypeJitter   MeasurementType = "jitter"
	MeasurementTypePacketLoss MeasurementType = "packet_loss"
)

// MeasurementStatus represents the status of a measurement
type MeasurementStatus string

const (
	MeasurementStatusSuccess MeasurementStatus = "success"
	MeasurementStatusFailed  MeasurementStatus = "failed"
	MeasurementStatusTimeout MeasurementStatus = "timeout"
)

// MeasurementStatistics represents statistical data for a measurement
type MeasurementStatistics struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Avg     float64 `json:"avg"`
	StdDev  float64 `json:"std_dev"`
	Count   int     `json:"count"`
}

// MeasurementStream represents a stream of measurements
type MeasurementStream struct {
	ProbeID  string        `json:"probe_id"`
	Target   string        `json:"target"`
	Measurements []Measurement `json:"measurements"`
	Timestamp time.Time     `json:"timestamp"`
}

// Manager manages monitoring operations
type Manager struct {
	mu           sync.RWMutex
	measurements map[string][]Measurement // key: probe_id
	activeStreams map[string]*MeasurementStream
	lastCleanup   time.Time
}

// NewManager creates a new monitoring manager
func NewManager() *Manager {
	m := &Manager{
		measurements: make(map[string][]Measurement),
		activeStreams: make(map[string]*MeasurementStream),
		lastCleanup: time.Now(),
	}
	
	// Start cleanup goroutine
	go m.cleanupOldMeasurements()
	
	return m
}

// AddMeasurement adds a new measurement
func (m *Manager) AddMeasurement(measurement *Measurement) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Add to measurements list
	m.measurements[measurement.ProbeID] = append(m.measurements[measurement.ProbeID], *measurement)
	
	// Limit memory usage by keeping only last 1000 measurements per probe
	measurements := m.measurements[measurement.ProbeID]
	if len(measurements) > 1000 {
		m.measurements[measurement.ProbeID] = measurements[len(measurements)-1000:]
	}
}

// GetMeasurements returns measurements for a probe
func (m *Manager) GetMeasurements(probeID string, limit int) []Measurement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	measurements := m.measurements[probeID]
	if limit > 0 && len(measurements) > limit {
		return measurements[len(measurements)-limit:]
	}
	
	return measurements
}

// GetMeasurementByID returns a specific measurement
func (m *Manager) GetMeasurementByID(probeID, measurementID string) (*Measurement, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, measurement := range m.measurements[probeID] {
		if measurement.ID == measurementID {
			return &measurement, true
		}
	}
	
	return nil, false
}

// UpdateMeasurement updates an existing measurement
func (m *Manager) UpdateMeasurement(probeID, measurementID string, updateFunc func(*Measurement)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for i := range m.measurements[probeID] {
		if m.measurements[probeID][i].ID == measurementID {
			updateFunc(&m.measurements[probeID][i])
			return nil
		}
	}
	
	return nil // Measurement not found
}

// StartMeasurementStream starts a measurement stream for a probe
func (m *Manager) StartMeasurementStream(probeID, target string) *MeasurementStream {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	stream := &MeasurementStream{
		ProbeID:  probeID,
		Target:   target,
		Measurements: make([]Measurement, 0),
		Timestamp: time.Now(),
	}
	
	key := probeID + ":" + target
	m.activeStreams[key] = stream
	
	return stream
}

// StopMeasurementStream stops a measurement stream
func (m *Manager) StopMeasurementStream(probeID, target string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := probeID + ":" + target
	delete(m.activeStreams, key)
}

// AddToStream adds a measurement to the active stream
func (m *Manager) AddToStream(measurement *Measurement) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := measurement.ProbeID + ":" + measurement.Target
	stream, exists := m.activeStreams[key]
	if !exists {
		return
	}
	
	stream.Measurements = append(stream.Measurements, *measurement)
	stream.Timestamp = time.Now()
	
	// Limit stream size
	if len(stream.Measurements) > 100 {
		stream.Measurements = stream.Measurements[len(stream.Measurements)-100:]
	}
}

// GetStream gets an active measurement stream
func (m *Manager) GetStream(probeID, target string) (*MeasurementStream, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	key := probeID + ":" + target
	stream, exists := m.activeStreams[key]
	
	if !exists {
		return nil, false
	}
	
	return stream, true
}

// GetAllActiveStreams returns all active measurement streams
func (m *Manager) GetAllActiveStreams() map[string]*MeasurementStream {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]*MeasurementStream)
	for key, stream := range m.activeStreams {
		result[key] = stream
	}
	
	return result
}

// GetMeasurementStatistics returns statistics for measurements
func (m *Manager) GetMeasurementStatistics(probeID, target string, duration time.Duration) (*MeasurementStatistics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	measurements := m.measurements[probeID]
	cutoff := time.Now().Add(-duration)
	
	var values []float64
	for i := len(measurements) - 1; i >= 0; i-- {
		measurement := measurements[i]
		if measurement.Timestamp.Before(cutoff) {
			break
		}
		
		if measurement.Target == target && measurement.Status == MeasurementStatusSuccess {
			values = append(values, measurement.Value)
		}
	}
	
	if len(values) == 0 {
		return &MeasurementStatistics{}, nil
	}
	
	// Calculate statistics
	stats := calculateStatistics(values)
	return &stats, nil
}

// calculateStatistics calculates min, max, avg, and std dev
func calculateStatistics(values []float64) MeasurementStatistics {
	if len(values) == 0 {
		return MeasurementStatistics{}
	}
	
	min := values[0]
	max := values[0]
	sum := 0.0
	
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	
	avg := sum / float64(len(values))
	
	// Calculate standard deviation
	var varianceSum float64
	for _, v := range values {
		diff := v - avg
		varianceSum += diff * diff
	}
	
	stdDev := 0.0
	if len(values) > 1 {
		stdDev = varianceSum / float64(len(values)-1)
		stdDev = sqrt(stdDev)
	}
	
	return MeasurementStatistics{
		Min:    min,
		Max:    max,
		Avg:    avg,
		StdDev: stdDev,
		Count:  len(values),
	}
}

// sqrt implements square root for float64
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	
	// Simple Newton-Raphson method
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// cleanupOldMeasurements removes old measurements periodically
func (m *Manager) cleanupOldMeasurements() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		m.cleanupMeasurements()
	}
}

// cleanupMeasurements removes measurements older than 24 hours
func (m *Manager) cleanupMeasurements() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	cutoff := time.Now().Add(-24 * time.Hour)
	
	for probeID, measurements := range m.measurements {
		var filtered []Measurement
		for _, measurement := range measurements {
			if measurement.Timestamp.After(cutoff) {
				filtered = append(filtered, measurement)
			}
		}
		m.measurements[probeID] = filtered
	}
	
	m.lastCleanup = time.Now()
}

// Stop stops the monitoring manager
func (m *Manager) Stop(ctx context.Context) {
	// Cleanup measurement streams
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.activeStreams = make(map[string]*MeasurementStream)
	
	// Perform final cleanup
	m.cleanupMeasurements()
}

// ToJSON converts measurements to JSON
func (m *Manager) ToJSON(probeID string) (string, error) {
	measurements := m.GetMeasurements(probeID, 0)
	data, err := json.MarshalIndent(measurements, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MeasurementCount returns the total number of measurements
func (m *Manager) MeasurementCount(probeID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.measurements[probeID])
}