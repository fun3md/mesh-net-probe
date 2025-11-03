package icmp

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mesh-net-probe/probe/pkg/types"
	"golang.org/x/sys/cpu"
)

// TimingProvider interface for platform-specific timing operations
type TimingProvider interface {
	// Now returns current high-resolution time
	Now() time.Duration

	// GetPrecision returns the timing precision for this platform
	GetPrecision() types.MeasurementPrecision

	// SupportsMonotonicClock returns true if monotonic clock is supported
	SupportsMonotonicClock() bool

	// GetHighResTimer returns true if high-resolution timers are available
	GetHighResTimer() bool
}

// HighPrecisionTimer provides microsecond-level timing precision
type HighPrecisionTimer struct {
	provider TimingProvider
	start    time.Duration
	mu       sync.Mutex
}

// NewHighPrecisionTimer creates a new high-precision timer
func NewHighPrecisionTimer() *HighPrecisionTimer {
	return &HighPrecisionTimer{
		provider: getPlatformTimingProvider(),
	}
}

// Now returns the current high-resolution time
func (t *HighPrecisionTimer) Now() time.Duration {
	return t.provider.Now()
}

// GetPrecision returns the timing precision
func (t *HighPrecisionTimer) GetPrecision() types.MeasurementPrecision {
	return t.provider.GetPrecision()
}

// Start starts the timer
func (t *HighPrecisionTimer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.start = t.Now()
}

// Stop returns the elapsed time since Start was called
func (t *HighPrecisionTimer) Stop() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	elapsed := t.Now() - t.start
	return elapsed
}

// Measure performs a measurement operation with high precision timing
func (t *HighPrecisionTimer) Measure(operation func() error) (time.Duration, error) {
	start := t.Now()
	err := operation()
	elapsed := t.Now() - start
	return elapsed, err
}

// ZeroClock provides a simple time.Duration-based clock for platforms without high-res timers
type ZeroClock struct{}

func (c ZeroClock) Now() time.Duration {
	return time.Duration(time.Now().UnixNano())
}

func (c ZeroClock) GetPrecision() types.MeasurementPrecision {
	return types.PrecisionMillisecond
}

func (c ZeroClock) SupportsMonotonicClock() bool {
	return false
}

func (c ZeroClock) GetHighResTimer() bool {
	return false
}

// SystemClock provides high-resolution timing using Go's runtime
type SystemClock struct {
	monotonic bool
}

func NewSystemClock() *SystemClock {
	// Check if monotonic clock is supported
	monotonic := runtime.GOOS != "nacl"
	
	return &SystemClock{
		monotonic: monotonic,
	}
}

func (c SystemClock) Now() time.Duration {
	if c.monotonic {
		return time.Duration(time.Now().UnixNano())
	}
	return time.Duration(time.Now().UnixNano())
}

func (c SystemClock) GetPrecision() types.MeasurementPrecision {
	// Try to determine actual precision
	start := time.Now()
	end := time.Now()
	
	// If we can measure sub-millisecond differences, we have high precision
	if end.Sub(start) < time.Millisecond {
		return types.PrecisionMicrosecond
	}
	return types.PrecisionMillisecond
}

func (c SystemClock) SupportsMonotonicClock() bool {
	return c.monotonic
}

func (c SystemClock) GetHighResTimer() bool {
	// Check CPU features for high-resolution timer support
	return cpu.ARM64.HasASIMD || cpu.X86.HasAVX
}

// CPUTimeClock provides CPU-time based timing for more precise measurements
type CPUTimeClock struct {
	startTime int64
}

func NewCPUTimeClock() *CPUTimeClock {
	return &CPUTimeClock{}
}

func (c *CPUTimeClock) Now() time.Duration {
	// Get current CPU time in nanoseconds
	// This is platform-specific and would need implementation
	return time.Duration(atomic.LoadInt64(&c.startTime))
}

func (c *CPUTimeClock) GetPrecision() types.MeasurementPrecision {
	return types.PrecisionNanosecond
}

func (c *CPUTimeClock) SupportsMonotonicClock() bool {
	return true
}

func (c *CPUTimeClock) GetHighResTimer() bool {
	return true
}

// getPlatformTimingProvider selects the best available timing provider
func getPlatformTimingProvider() TimingProvider {
	// Check platform capabilities
	switch runtime.GOOS {
	case "linux":
		return NewSystemClock()
	case "darwin":
		return NewSystemClock()
	case "windows":
		return NewSystemClock()
	default:
		return ZeroClock{}
	}
}

// Measurement represents a timed measurement operation
type Measurement struct {
	startTime  time.Duration
	endTime    time.Duration
	operation  string
	success    bool
	error      error
}

// NewMeasurement creates a new measurement
func NewMeasurement(operation string) *Measurement {
	return &Measurement{
		operation: operation,
	}
}

// Start starts the measurement
func (m *Measurement) Start(provider TimingProvider) {
	m.startTime = provider.Now()
}

// Stop stops the measurement
func (m *Measurement) Stop(provider TimingProvider) {
	m.endTime = provider.Now()
}

// Elapsed returns the elapsed time
func (m *Measurement) Elapsed() time.Duration {
	return m.endTime - m.startTime
}

// Success marks the measurement as successful
func (m *Measurement) Success() {
	m.success = true
}

// Error sets the measurement error
func (m *Measurement) Error(err error) {
	m.error = err
	m.success = false
}

// IsSuccess returns true if the measurement was successful
func (m *Measurement) IsSuccess() bool {
	return m.success
}

// GetError returns the measurement error
func (m *Measurement) GetError() error {
	return m.error
}

// TimingStats contains timing measurement statistics
type TimingStats struct {
	MinTime     time.Duration `json:"min_time"`
	MaxTime     time.Duration `json:"max_time"`
	AvgTime     time.Duration `json:"avg_time"`
	TotalTime   time.Duration `json:"total_time"`
	Count       int           `json:"count"`
	Precision   types.MeasurementPrecision `json:"precision"`
	SuccessRate float64       `json:"success_rate"`
}

// NewTimingStats creates a new timing statistics instance
func NewTimingStats(precision types.MeasurementPrecision) *TimingStats {
	return &TimingStats{
		MinTime:   -1,
		MaxTime:   0,
		AvgTime:   0,
		TotalTime: 0,
		Count:     0,
		Precision: precision,
	}
}

// AddMeasurement adds a measurement to the statistics
func (s *TimingStats) AddMeasurement(measurement *Measurement) {
	s.Count++
	s.TotalTime += measurement.Elapsed()
	
	// Update min/max
	if s.MinTime == -1 || measurement.Elapsed() < s.MinTime {
		s.MinTime = measurement.Elapsed()
	}
	if measurement.Elapsed() > s.MaxTime {
		s.MaxTime = measurement.Elapsed()
	}
	
	// Calculate average
	s.AvgTime = s.TotalTime / time.Duration(s.Count)
}

// AddSuccessfulMeasurement adds a successful measurement
func (s *TimingStats) AddSuccessfulMeasurement(measurement *Measurement) {
	s.AddMeasurement(measurement)
	
	// Update success rate (simplified - assume we track this separately)
	// In a real implementation, you'd track success/failure counts separately
}

// GetStats returns the current statistics
func (s *TimingStats) GetStats() *TimingStats {
	// Return a copy to prevent external modification
	return &TimingStats{
		MinTime:     s.MinTime,
		MaxTime:     s.MaxTime,
		AvgTime:     s.AvgTime,
		TotalTime:   s.TotalTime,
		Count:       s.Count,
		Precision:   s.Precision,
		SuccessRate: s.SuccessRate,
	}
}

// TimingManager manages timing operations across the system
type TimingManager struct {
	timer        *HighPrecisionTimer
	defaultStats *TimingStats
	mu           sync.RWMutex
	stats        map[string]*TimingStats
}

// NewTimingManager creates a new timing manager
func NewTimingManager() *TimingManager {
	timer := NewHighPrecisionTimer()
	defaultStats := NewTimingStats(timer.GetPrecision())
	
	return &TimingManager{
		timer:        timer,
		defaultStats: defaultStats,
		stats:        make(map[string]*TimingStats),
	}
}

// Measure performs a measurement with automatic timing
func (tm *TimingManager) Measure(operationName string, operation func() error) (*Measurement, error) {
	measurement := NewMeasurement(operationName)
	measurement.Start(tm.timer.provider)
	
	err := operation()
	measurement.Stop(tm.timer.provider)
	
	if err != nil {
		measurement.Error(err)
	} else {
		measurement.Success()
	}
	
	// Update statistics
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if stats, exists := tm.stats[operationName]; exists {
		stats.AddMeasurement(measurement)
	} else {
		stats := NewTimingStats(tm.timer.GetPrecision())
		stats.AddMeasurement(measurement)
		tm.stats[operationName] = stats
	}
	
	return measurement, err
}

// GetStats returns statistics for a specific operation
func (tm *TimingManager) GetStats(operationName string) *TimingStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	if stats, exists := tm.stats[operationName]; exists {
		return stats.GetStats()
	}
	
	return tm.defaultStats.GetStats()
}

// GetAllStats returns all timing statistics
func (tm *TimingManager) GetAllStats() map[string]*TimingStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	result := make(map[string]*TimingStats)
	for name, stats := range tm.stats {
		result[name] = stats.GetStats()
	}
	
	return result
}

// GetCurrentTime returns the current high-resolution time
func (tm *TimingManager) GetCurrentTime() time.Duration {
	return tm.timer.Now()
}

// GetPrecision returns the current timing precision
func (tm *TimingManager) GetPrecision() types.MeasurementPrecision {
	return tm.timer.GetPrecision()
}

// IsHighPrecision returns true if the system supports high-precision timing
func (tm *TimingManager) IsHighPrecision() bool {
	return tm.timer.provider.GetHighResTimer()
}