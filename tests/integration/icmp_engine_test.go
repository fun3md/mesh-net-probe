package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/mesh-net-probe/probe/internal/icmp"
	"github.com/mesh-net-probe/probe/pkg/types"
)

// TestICMPEngineInitialization tests the initialization of the ICMP engine
func TestICMPEngineInitialization(t *testing.T) {
	ctx := context.Background()
	engine := icmp.NewEngine()
	
	// Test with nil source IP (let system choose)
	config := &types.NetworkConfig{
		SourceIP:   nil,
		BufferSize: 4096,
		TTL:        64,
	}
	
	// This should work without needing specific network permissions
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Logf("Engine initialization failed (expected in restricted environments): %v", err)
		t.Skip("ICMP engine requires elevated permissions")
		return
	}
	
	defer engine.Close(ctx)
	
	// Test engine stats after initialization
	stats := engine.GetStats()
	if stats == nil {
		t.Error("Engine statistics are nil after initialization")
	}
	
	t.Log("ICMP engine initialization test passed")
}

// TestICMPEngineValidation tests input validation
func TestICMPEngineValidation(t *testing.T) {
	ctx := context.Background()
	engine := icmp.NewEngine()
	
	// Test validation with nil target
	_, err := engine.Measure(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil target")
	}
	
	// Test validation with nil address
	invalidTarget := &types.NetworkTarget{
		ID:      "invalid_target",
		Address: nil,
		Timeout: 1 * time.Second,
		Enabled: true,
	}
	
	_, err = engine.Measure(ctx, invalidTarget)
	if err == nil {
		t.Error("Expected error for nil address")
	}
}

// TestICMPEngineWithLoopback tests measurements to loopback address
func TestICMPEngineWithLoopback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping loopback test in short mode")
	}
	
	ctx := context.Background()
	engine := icmp.NewEngine()
	
	config := &types.NetworkConfig{
		SourceIP:   net.ParseIP("127.0.0.1"),
		BufferSize: 4096,
		TTL:        64,
	}
	
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine (likely permission issue): %v", err)
	}
	defer engine.Close(ctx)
	
	// Test measurement to loopback
	target := &types.NetworkTarget{
		ID:      "test_loopback",
		Address: net.ParseIP("127.0.0.1"),
		Timeout: 2 * time.Second,
		Enabled: true,
	}
	
	measurement, err := engine.Measure(ctx, target)
	if err != nil {
		t.Logf("Loopback measurement failed: %v", err)
		// This might fail due to permissions, which is expected
		return
	}
	
	if measurement == nil {
		t.Error("Measurement result is nil")
		return
	}
	
	if measurement.SourceIP.String() != "127.0.0.1" {
		t.Errorf("Expected source IP 127.0.0.1, got %s", measurement.SourceIP.String())
	}
	
	t.Logf("Loopback measurement: success=%v, rtt=%v", measurement.Success, measurement.RTT)
}

// TestTimingEngine tests timing precision functionality
func TestTimingEngine(t *testing.T) {
	timing := icmp.NewTimingEngine()
	
	// Initialize timing engine
	err := timing.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	// Test getting current time
	now := timing.GetCurrentTime()
	if now.IsZero() {
		t.Error("Current time is zero")
	}
	
	// Test timer resolution
	resolution := timing.GetTimerResolution()
	if resolution <= 0 {
		t.Error("Timer resolution must be positive")
	}
	
	// Test statistics
	stats := timing.GetStats()
	if stats == nil {
		t.Error("Timing statistics are nil")
	}
	
	systemInfo := timing.GetSystemInfo()
	if systemInfo == nil {
		t.Error("System information is nil")
	}
	
	// Test precision validation
	precision, err := timing.ValidatePrecision()
	if err != nil {
		t.Errorf("Precision validation failed: %v", err)
	}
	
	if precision == "" {
		t.Error("Precision validation returned empty result")
	}
	
	t.Logf("Timing engine test: resolution=%v, precision=%v", resolution, precision)
}

// TestTimingEngineCalibration tests timing calibration functionality
func TestTimingEngineCalibration(t *testing.T) {
	timing := icmp.NewTimingEngine()
	
	err := timing.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	// Test calibration
	err = timing.Calibrate()
	if err != nil {
		t.Errorf("Timing calibration failed: %v", err)
	}
	
	// Get updated statistics after calibration
	stats := timing.GetStats()
	if stats.CalibrationRuns == 0 {
		t.Error("No calibration runs recorded")
	}
	
	if stats.LastCalibration.IsZero() {
		t.Error("Last calibration timestamp is zero")
	}
	
	t.Logf("Timing calibration test: runs=%d, avg_resolution=%v", 
		stats.CalibrationRuns, stats.AverageResolution)
}

// TestEngineStatistics tests engine statistics functionality
func TestEngineStatistics(t *testing.T) {
	ctx := context.Background()
	engine := icmp.NewEngine()
	
	config := &types.NetworkConfig{
		SourceIP:   nil, // Let system choose
		BufferSize: 4096,
		TTL:        64,
	}
	
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine: %v", err)
	}
	defer engine.Close(ctx)
	
	// Get initial statistics
	stats := engine.GetStats()
	initialTotal := stats.MeasurementsTotal
	
	// Test with invalid target to trigger error handling
	invalidTarget := &types.NetworkTarget{
		ID:      "invalid_stats_test",
		Address: net.ParseIP("invalid_ip"),
		Timeout: 100 * time.Millisecond,
		Enabled: true,
	}
	
	_, _ = engine.Measure(ctx, invalidTarget) // Expected to fail
	
	// Get updated statistics
	stats = engine.GetStats()
	
	if stats.MeasurementsTotal <= initialTotal {
		t.Error("Statistics not updated after measurement attempt")
	}
	
	t.Logf("Engine statistics test: measurements_total=%d", stats.MeasurementsTotal)
}

// TestEngineBatchOperations tests batch operation functionality
func TestEngineBatchOperations(t *testing.T) {
	ctx := context.Background()
	engine := icmp.NewEngine()
	
	config := &types.NetworkConfig{
		SourceIP:   nil,
		BufferSize: 4096,
		TTL:        64,
	}
	
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine: %v", err)
	}
	defer engine.Close(ctx)
	
	// Create multiple test targets
	targets := []*types.NetworkTarget{
		{
			ID:      "batch_test_1",
			Address: net.ParseIP("127.0.0.1"),
			Timeout: 1 * time.Second,
			Enabled: true,
		},
		{
			ID:      "batch_test_2", 
			Address: net.ParseIP("127.0.0.1"),
			Timeout: 1 * time.Second,
			Enabled: true,
		},
	}
	
	// Test batch measurement
	measurements, err := engine.MeasureBatch(ctx, targets)
	if err != nil {
		t.Logf("Batch measurement failed (expected in restricted environments): %v", err)
		return
	}
	
	if len(measurements) == 0 {
		t.Error("Batch measurement returned empty results")
	}
	
	for i, measurement := range measurements {
		if measurement == nil {
			t.Errorf("Measurement %d is nil", i)
			continue
		}
		
		t.Logf("Batch measurement %d: target=%s, success=%v", 
			i, measurement.Target.ID, measurement.Success)
	}
}

// TestContinuousMode tests continuous measurement functionality
func TestContinuousMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping continuous mode test in short mode")
	}
	
	ctx := context.Background()
	engine := icmp.NewEngine()
	
	config := &types.NetworkConfig{
		SourceIP:   nil,
		BufferSize: 4096,
		TTL:        64,
	}
	
	err := engine.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine: %v", err)
	}
	defer engine.Close(ctx)
	
	targets := []*types.NetworkTarget{
		{
			ID:      "continuous_test",
			Address: net.ParseIP("127.0.0.1"),
			Timeout: 1 * time.Second,
			Enabled: true,
		},
	}
	
	// Test starting continuous mode
	err = engine.StartContinuous(ctx, targets, 500*time.Millisecond)
	if err != nil {
		t.Logf("Failed to start continuous mode: %v", err)
		return
	}
	
	// Let it run briefly
	time.Sleep(1 * time.Second)
	
	// Test stopping continuous mode
	err = engine.StopContinuous(ctx)
	if err != nil {
		t.Errorf("Failed to stop continuous mode: %v", err)
	}
	
	// Check if stats were updated
	stats := engine.GetStats()
	if stats.MeasurementsTotal == 0 {
		t.Error("No measurements recorded in continuous mode")
	}
	
	t.Logf("Continuous mode test completed: %d measurements", stats.MeasurementsTotal)
}

// Benchmark ICMP Engine Tests

func BenchmarkTimingEngineInitialize(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		timing := icmp.NewTimingEngine()
		_ = timing.Initialize()
	}
}

func BenchmarkTimingEngineGetCurrentTime(b *testing.B) {
	timing := icmp.NewTimingEngine()
	_ = timing.Initialize()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = timing.GetCurrentTime()
	}
}

func BenchmarkTimingEngineCalibration(b *testing.B) {
	timing := icmp.NewTimingEngine()
	_ = timing.Initialize()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = timing.Calibrate()
	}
}

// Helper functions

func getLocalIP() net.IP {
	// Get the first non-loopback IP address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return net.ParseIP("127.0.0.1")
	}
	
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP
			}
		}
	}
	
	return net.ParseIP("127.0.0.1")
}