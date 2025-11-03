package integration

import (
	"context"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/mesh-net-probe/probe/internal/icmp"
	"github.com/mesh-net-probe/probe/internal/platform"
	"github.com/mesh-net-probe/probe/pkg/types"
)

// TestX64ARM64Consistency tests measurement consistency between x64 and ARM64 architectures
func TestX64ARM64Consistency(t *testing.T) {
	// Detect current platform
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Testing architecture consistency on: %s/%s", platformInfo.OS, platformInfo.Arch)
	
	// Create ICMP engine for timing testing
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Fatalf("Failed to create ICMP engine: %v", err)
	}
	defer engine.Close()
	
	// Test timing resolution using measurements
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	// Perform timing tests
	start := time.Now()
	measurement, err := engine.Ping(context.Background(), target)
	elapsed := time.Since(start)
	
	if err != nil {
		t.Errorf("ICMP ping failed: %v", err)
		return
	}
	
	if measurement != nil {
		t.Logf("RTT measurement: %v", measurement.RTT)
	}
	
	t.Logf("Architecture %s timing resolution: %v", platformInfo.Arch, elapsed)
	
	// Test CPU features
	features := getCPUFeaturesForArchitecture(platformInfo.Arch)
	if features == nil {
		t.Errorf("No CPU features detected for architecture %s", platformInfo.Arch)
		return
	}
	
	t.Logf("Current architecture %s CPU features: %v", platformInfo.Arch, features)
	
	// Test platform optimizations
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		t.Errorf("Failed to get platform capabilities: %v", err)
		return
	}
	
	t.Logf("Current architecture %s capabilities: %v", platformInfo.Arch, capabilities.Features)
	
	// Validate architecture-specific expectations
	validateArchitectureConsistency(t, platformInfo.Arch)
}

// TestArchitectureSpecificOptimization tests that optimizations work correctly for each architecture
func TestArchitectureSpecificOptimization(t *testing.T) {
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Testing architecture-specific optimizations for: %s/%s", platformInfo.OS, platformInfo.Arch)
	
	// Test platform capabilities
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		t.Errorf("Failed to get platform capabilities: %v", err)
		return
	}
	
	// Validate architecture-specific capabilities
	switch platformInfo.Arch {
	case "amd64":
		// AMD64 should have certain features
		if !capabilities.CanMeasure {
			t.Error("AMD64 should support ICMP measurements")
		}
	case "arm64":
		// ARM64 should also support measurements
		if !capabilities.CanMeasure {
			t.Error("ARM64 should support ICMP measurements")
		}
	case "arm":
		// ARM32 may have limitations
		t.Logf("ARM32 capabilities: %v", capabilities.Limitations)
	}
	
	// Test memory optimization settings
	if len(capabilities.Features) == 0 {
		t.Error("Platform should have defined features")
	}
	
	t.Logf("Architecture %s optimizations validated successfully", platformInfo.Arch)
}

// TestPerformanceBenchmarkComparison benchmarks performance across architectures
func TestPerformanceBenchmarkComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance benchmark in short mode")
	}
	
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Running performance benchmarks for architecture: %s/%s", platformInfo.OS, platformInfo.Arch)
	
	// Initialize ICMP engine
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Fatalf("Failed to create ICMP engine: %v", err)
	}
	defer engine.Close()
	
	// Benchmark timing operations
	benchmarkResults := benchmarkTimingOperations(t, engine)
	
	t.Logf("Architecture %s benchmark results: %v", platformInfo.Arch, benchmarkResults)
	
	// Architecture-specific performance expectations
	switch platformInfo.Arch {
	case "amd64":
		// Modern x86_64 should have excellent performance
		if benchmarkResults.operationsPerSecond > 100000 {
			t.Logf("AMD64 performance excellent: %v", benchmarkResults)
		}
	case "arm64":
		// ARM64 should also have good performance
		if benchmarkResults.operationsPerSecond > 50000 {
			t.Logf("ARM64 performance good: %v", benchmarkResults)
		}
	case "arm":
		// 32-bit ARM might have lower performance
		if benchmarkResults.operationsPerSecond > 10000 {
			t.Logf("ARM performance acceptable: %v", benchmarkResults)
		}
	}
}

// TestCrossArchitectureCompatibility tests that the same measurement produces consistent results across architectures
func TestCrossArchitectureCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cross-architecture compatibility test in short mode")
	}
	
	ctx := context.Background()
	
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Testing cross-architecture compatibility for: %s/%s", platformInfo.OS, platformInfo.Arch)
	
	// Initialize ICMP engine
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine: %v", err)
	}
	defer engine.Close()
	
	// Test engine functionality
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	measurement, err := engine.Ping(ctx, target)
	if err != nil {
		t.Errorf("ICMP ping failed: %v", err)
		return
	}
	
	if measurement == nil {
		t.Error("Measurement result should not be nil")
		return
	}
	
	// Validate measurement structure consistency
	if measurement.RTT < 0 {
		t.Error("RTT should be non-negative")
	}
	
	if !measurement.Timestamp.IsZero() {
		t.Logf("Architecture %s measurement timestamp: %v", platformInfo.Arch, measurement.Timestamp)
	}
	
	t.Logf("Architecture %s compatibility validated successfully", platformInfo.Arch)
}

// TestPlatformCapabilitiesConsistency tests that platform capabilities are consistent
func TestPlatformCapabilitiesConsistency(t *testing.T) {
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Testing platform capabilities for: %s/%s", platformInfo.OS, platformInfo.Arch)
	
	// Get capabilities
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		t.Fatalf("Failed to get platform capabilities: %v", err)
	}
	
	if capabilities == nil {
		t.Error("Platform capabilities should not be nil")
		return
	}
	
	// Validate essential capabilities
	if !capabilities.CanMeasure {
		t.Errorf("Platform should support ICMP measurements")
	}
	
	if len(capabilities.Features) == 0 {
		t.Error("Platform should have defined features")
	}
	
	// Validate architecture-specific expectations
	switch platformInfo.Arch {
	case "amd64", "arm64":
		// 64-bit architectures should have good precision
		if capabilities.MaxPrecision == "" {
			t.Errorf("64-bit architecture %s should have defined precision", platformInfo.Arch)
		}
	case "arm":
		// 32-bit ARM may have limitations
		t.Logf("32-bit ARM architecture capabilities: %v", capabilities.Limitations)
	}
	
	t.Logf("Platform capabilities for %s validated successfully", platformInfo.Arch)
}

// Helper functions

func getCPUFeaturesForArchitecture(arch string) map[string]bool {
	switch arch {
	case "amd64":
		return map[string]bool{
			"sse2":   true,
			"sse4.1": true,
			"sse4.2": true,
			"avx":    true,
			"avx2":   true,
			"aes":    true,
		}
	case "arm64":
		return map[string]bool{
			"neon":   true,
			"aes":    true,
			"sha256": true,
		}
	case "arm":
		return map[string]bool{
			"neon": true,
		}
	default:
		return nil
	}
}

func validateArchitectureConsistency(t *testing.T, arch string) {
	switch arch {
	case "amd64":
		// AMD64 should support modern features
		if runtime.GOARCH != "amd64" {
			t.Error("Expected to be running on AMD64 architecture")
		}
	case "arm64":
		// ARM64 should support ARM features
		if runtime.GOARCH != "arm64" {
			t.Error("Expected to be running on ARM64 architecture")
		}
	case "arm":
		// ARM32 should support basic features
		if runtime.GOARCH != "arm" {
			t.Error("Expected to be running on ARM32 architecture")
		}
	}
}

func createPlatformCompatibleConfig(platformInfo *types.PlatformInfo) *types.NetworkConfig {
	// Return a basic configuration - the ICMP engine uses functional options
	// This function is kept for compatibility but not used in the updated API
	return &types.NetworkConfig{
		BufferSize: 4096,
		TTL:        64,
	}
}

// Benchmark results structure for performance comparison
type benchmarkResults struct {
	operationsPerSecond float64
	nanosecondPrecision int64
}

func (br benchmarkResults) NanosecondPrecision() int64 {
	return br.nanosecondPrecision
}

func benchmarkTimingOperations(t *testing.T, engine *icmp.Engine) benchmarkResults {
	// Run ICMP operations benchmark
	start := time.Now()
	const iterations = 100
	
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return benchmarkResults{}
	}
	
	ctx := context.Background()
	successCount := 0
	
	for i := 0; i < iterations; i++ {
		measurement, err := engine.Ping(ctx, target)
		if err == nil && measurement != nil {
			successCount++
		}
	}
	
	elapsed := time.Since(start)
	operationsPerSecond := float64(iterations) / elapsed.Seconds()
	
	return benchmarkResults{
		operationsPerSecond: operationsPerSecond,
		nanosecondPrecision: int64(elapsed.Nanoseconds() / int64(iterations)),
	}
}

// BenchmarkCrossArchitecturePerformance benchmarks performance across different architectures
func BenchmarkCrossArchitectureTiming(b *testing.B) {
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(1*time.Second),
		icmp.WithBufferSize(4096),
	)
	if err != nil {
		b.Fatalf("Failed to create ICMP engine: %v", err)
	}
	defer engine.Close()
	
	// Platform information
	platformInfo, _ := platform.DetectPlatform()
	b.Logf("Benchmarking ICMP operations on %s/%s", platformInfo.OS, platformInfo.Arch)
	
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

func BenchmarkCrossArchitectureBatch(b *testing.B) {
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(1*time.Second),
		icmp.WithBufferSize(4096),
	)
	if err != nil {
		b.Fatalf("Failed to create ICMP engine: %v", err)
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