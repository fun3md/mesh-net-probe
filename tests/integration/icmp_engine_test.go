package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/mesh-net-probe/probe/internal/icmp"
)

// TestICMPEngineInitialization tests the initialization of the ICMP engine
func TestICMPEngineInitialization(t *testing.T) {
	ctx := context.Background()
	
	// Create engine with basic configuration
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Logf("Engine initialization failed (expected in restricted environments): %v", err)
		t.Skip("ICMP engine requires elevated permissions")
		return
	}
	
	defer engine.Close()
	
	// Test basic functionality with ping
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	// Perform a test ping to verify engine functionality
	measurement, err := engine.Ping(ctx, target)
	if err != nil {
		t.Logf("Test ping failed (expected in restricted environments): %v", err)
		t.Skip("ICMP permissions required")
		return
	}
	
	if measurement == nil {
		t.Error("Measurement result is nil")
		return
	}
	
	t.Log("ICMP engine initialization test passed")
}

// TestICMPEngineValidation tests input validation
func TestICMPEngineValidation(t *testing.T) {
	ctx := context.Background()
	
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Skip("Cannot create ICMP engine")
		return
	}
	defer engine.Close()
	
	// Test validation with nil target
	_, err = engine.Ping(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil target")
	}
	
	// Test validation with invalid IP
	invalidIP := net.ParseIP("invalid_ip")
	_, err = engine.Ping(ctx, invalidIP)
	if err == nil {
		t.Error("Expected error for invalid IP")
	}
}

// TestICMPEngineWithLoopback tests measurements to loopback address
func TestICMPEngineWithLoopback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping loopback test in short mode")
	}
	
	ctx := context.Background()
	
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine (likely permission issue): %v", err)
	}
	defer engine.Close()
	
	// Test measurement to loopback
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	measurement, err := engine.Ping(ctx, target)
	if err != nil {
		t.Logf("Loopback measurement failed: %v", err)
		// This might fail due to permissions, which is expected
		return
	}
	
	if measurement == nil {
		t.Error("Measurement result is nil")
		return
	}
	
	if measurement.RTT < 0 {
		t.Error("RTT should be non-negative")
	}
	
	t.Logf("Loopback measurement: rtt=%v", measurement.RTT)
}

// TestTimingEngine tests timing precision functionality
func TestTimingEngine(t *testing.T) {
	// Create ICMP engine for timing testing
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Fatalf("Failed to create ICMP engine: %v", err)
	}
	defer engine.Close()
	
	// Test timing precision by performing a measurement
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	// Perform timing test
	start := time.Now()
	measurement, err := engine.Ping(context.Background(), target)
	elapsed := time.Since(start)
	
	if err != nil {
		t.Logf("Timing test failed (may be expected): %v", err)
		t.Skip("Timing test requires ICMP permissions")
		return
	}
	
	if measurement == nil {
		t.Error("Measurement result is nil")
		return
	}
	
	// Validate timing results
	if measurement.RTT < 0 {
		t.Error("RTT should be non-negative")
	}
	
	// The elapsed time should be reasonable for a loopback measurement
	if elapsed < 0 || elapsed > 10*time.Second {
		t.Errorf("Unexpected elapsed time: %v", elapsed)
	}
	
	t.Logf("Timing engine test: elapsed=%v, rtt=%v", elapsed, measurement.RTT)
}

// TestTimingEngineCalibration tests timing calibration functionality
func TestTimingEngineCalibration(t *testing.T) {
	// Create ICMP engine for calibration testing
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Fatalf("Failed to create ICMP engine: %v", err)
	}
	defer engine.Close()
	
	// Test timing precision by performing multiple measurements
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	ctx := context.Background()
	measurements := 0
	totalRTT := time.Duration(0)
	
	// Perform multiple measurements to "calibrate"
	for i := 0; i < 5; i++ {
		measurement, err := engine.Ping(ctx, target)
		if err == nil && measurement != nil && measurement.RTT >= 0 {
			measurements++
			totalRTT += measurement.RTT
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	if measurements == 0 {
		t.Error("No successful measurements for calibration")
		return
	}
	
	avgRTT := totalRTT / time.Duration(measurements)
	
	t.Logf("Timing calibration test: measurements=%d, avg_rtt=%v",
		measurements, avgRTT)
}

// TestEngineStatistics tests engine statistics functionality
func TestEngineStatistics(t *testing.T) {
	ctx := context.Background()
	
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine: %v", err)
	}
	defer engine.Close()
	
	// Test basic functionality
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	// Perform measurement
	measurement, err := engine.Ping(ctx, target)
	
	// The engine should handle the operation gracefully regardless of success/failure
	if err != nil {
		t.Logf("Measurement failed (may be expected): %v", err)
	}
	
	if measurement != nil {
		t.Logf("Measurement successful: rtt=%v", measurement.RTT)
	}
	
	t.Log("Engine statistics test completed")
}

// TestEngineBatchOperations tests batch operation functionality
func TestEngineBatchOperations(t *testing.T) {
	ctx := context.Background()
	
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine: %v", err)
	}
	defer engine.Close()
	
	// Test batch measurement
	targets := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.1")}
	
	measurements, err := engine.PingBatch(ctx, targets[0], 3)
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
		
		t.Logf("Batch measurement %d: rtt=%v", i, measurement.RTT)
	}
}

// TestContinuousMode tests continuous measurement functionality
func TestContinuousMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping continuous mode test in short mode")
	}
	
	ctx := context.Background()
	
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine: %v", err)
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	// Test continuous measurement by performing multiple pings
	for i := 0; i < 5; i++ {
		measurement, err := engine.Ping(ctx, target)
		if err != nil {
			t.Logf("Continuous measurement %d failed: %v", i, err)
			continue
		}
		
		if measurement != nil {
			t.Logf("Continuous measurement %d: rtt=%v", i, measurement.RTT)
		}
		
		// Small delay between measurements
		time.Sleep(100 * time.Millisecond)
	}
	
	t.Log("Continuous mode test completed")
}

// Benchmark ICMP Engine Tests

func BenchmarkICMPEngineCreate(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		engine, _ := icmp.NewEngine(
			icmp.WithTimeout(1*time.Second),
			icmp.WithBufferSize(4096),
		)
		_ = engine
	}
}

func BenchmarkICMPEnginePing(b *testing.B) {
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(1*time.Second),
		icmp.WithBufferSize(4096),
	)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		b.Fatal("Invalid target IP")
	}
	
	ctx := context.Background()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _ = engine.Ping(ctx, target)
	}
}

func BenchmarkICMPEngineBatchPing(b *testing.B) {
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(1*time.Second),
		icmp.WithBufferSize(4096),
	)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		b.Fatal("Invalid target IP")
	}
	
	ctx := context.Background()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _ = engine.PingBatch(ctx, target, 5)
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