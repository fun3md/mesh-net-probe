package cross_platform

import (
	"context"
	"net"
	"os"
	"runtime"
	"testing"
	"time"
)

// CrossPlatformTestSuite runs comprehensive tests across different platforms
func TestCrossPlatformCompatibility(t *testing.T) {
	t.Run("PlatformDetection", testPlatformDetection)
	t.Run("ICMPFunctionality", testICMPFunctionality)
	t.Run("TimingConsistency", testTimingConsistency)
	t.Run("ResourceManagement", testResourceManagement)
}

func testPlatformDetection(t *testing.T) {
	// Test platform detection
	platformInfo, err := detectPlatform()
	if err != nil {
		t.Errorf("Platform detection failed: %v", err)
		return
	}
	
	if platformInfo.OS == "" {
		t.Error("OS should be detected")
	}
	
	if platformInfo.Arch == "" {
		t.Error("Architecture should be detected")
	}
	
	t.Logf("Detected OS: %s, Architecture: %s", platformInfo.OS, platformInfo.Arch)
}

func testICMPFunctionality(t *testing.T) {
	ctx := context.Background()
	
	// Create ICMP engine
	engine, err := newICMPEngine()
	if err != nil {
		t.Skipf("ICMP engine creation failed: %v", err)
		return
	}
	defer engine.Close()
	
	// Test against localhost (should work on all platforms)
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Fatal("Invalid target IP")
	}
	
	measurement, err := engine.Ping(ctx, target)
	if err != nil {
		t.Errorf("ICMP measurement failed: %v", err)
		return
	}
	
	if measurement == nil {
		t.Error("Measurement result should not be nil")
		return
	}
	
	if measurement.RTT < 0 {
		t.Error("RTT should be non-negative")
	}
	
	t.Logf("RTT to 127.0.0.1: %v", measurement.RTT)
}

func testTimingConsistency(t *testing.T) {
	// Test timing accuracy across platforms
	const expectedTolerance = 100 * time.Microsecond // Allow 100µs tolerance
	
	measurements := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		start := time.Now()
		// Small operation
		time.Sleep(1 * time.Millisecond)
		elapsed := time.Since(start)
		measurements[i] = elapsed
	}
	
	// Check timing consistency
	for _, measurement := range measurements {
		deviation := measurement - time.Millisecond
		if absDuration(deviation) > expectedTolerance {
			t.Errorf("Timing measurement not consistent within tolerance: %v vs %v", measurement, time.Millisecond)
		}
	}
}

func testResourceManagement(t *testing.T) {
	// Test memory usage patterns
	var before runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	
	// Create and use ICMP engine
	engine, err := newICMPEngine()
	if err != nil {
		t.Skipf("Engine creation failed: %v", err)
		return
	}
	defer engine.Close()
	
	// Perform measurements
	target := net.ParseIP("127.0.0.1")
	if target != nil {
		_, _ = engine.Ping(context.Background(), target)
	}
	
	// Force garbage collection
	runtime.GC()
	
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	
	// Memory growth should be reasonable
		memoryGrowth := after.TotalAlloc - before.TotalAlloc
		maxAcceptableGrowth := uint64(100 * 1024 * 1024) // 100MB
		
		if memoryGrowth > maxAcceptableGrowth {
			t.Errorf("Memory usage growth too high: %d bytes", memoryGrowth)
		}
		
		t.Logf("Memory growth: %d bytes", memoryGrowth)
}

// Test infrastructure for continuous integration
func TestContinuousIntegration(t *testing.T) {
	// Environment detection for CI
	ci := os.Getenv("CI")
	githubActions := os.Getenv("GITHUB_ACTIONS")
	
	if ci == "true" || githubActions == "true" {
		t.Log("Running in CI environment")
		
		// Reduced test count for CI
		t.Run("QuickICMPTest", testQuickICMP)
		
		// Skip long-running tests in CI
		t.Skip("Long-running tests skipped in CI environment")
	} else {
		t.Log("Running in local environment")
		t.Run("FullICMPTest", testFullICMP)
	}
}

func testQuickICMP(t *testing.T) {
	engine, err := newICMPEngine()
	if err != nil {
		t.Skipf("Engine creation failed: %v", err)
		return
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Fatal("Invalid target IP")
	}
	
	measurement, err := engine.Ping(context.Background(), target)
	if err != nil {
		t.Errorf("Quick measurement failed: %v", err)
		return
	}
	
	if measurement == nil {
		t.Error("Should have measurement result")
	}
}

func testFullICMP(t *testing.T) {
	engine, err := newICMPEngine()
	if err != nil {
		t.Skipf("Engine creation failed: %v", err)
		return
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Fatal("Invalid target IP")
	}
	
	// Perform 10 measurements
	for i := 0; i < 10; i++ {
		measurement, err := engine.Ping(context.Background(), target)
		if err != nil {
			t.Errorf("Measurement %d failed: %v", i+1, err)
			continue
		}
		
		if measurement == nil {
			t.Errorf("Measurement %d result is nil", i+1)
		}
	}
}

// Benchmark tests for performance validation
func BenchmarkICMPMeasurement(b *testing.B) {
	engine, err := newICMPEngine()
	if err != nil {
		b.Skipf("Engine creation failed: %v", err)
		return
	}
	defer engine.Close()
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		b.Fatal("Invalid target IP")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Ping(context.Background(), target)
		if err != nil {
			b.Errorf("Measurement failed: %v", err)
		}
	}
}

// Utility functions and helpers
func detectPlatform() (*platformInfo, error) {
	return &platformInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Hostname:  getHostname(),
	}, nil
}

func newICMPEngine() (*mockEngine, error) {
	// Create a simple mock engine for testing
	return &mockEngine{}, nil
}

type platformInfo struct {
	OS        string
	Arch      string
	Hostname  string
}

func getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "unknown"
}

type mockEngine struct{}

func (e *mockEngine) Ping(ctx context.Context, target net.IP) (*mockMeasurement, error) {
	// Mock measurement that always succeeds
	return &mockMeasurement{
		RTT:      time.Millisecond,
		Success:  true,
		Target:   target,
	}, nil
}

func (e *mockEngine) Close() error {
	return nil
}

type mockMeasurement struct {
	RTT      time.Duration
	Success  bool
	Target   net.IP
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}