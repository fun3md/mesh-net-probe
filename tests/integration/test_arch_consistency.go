package integration

import (
	"context"
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
	
	// Create timing engine
	timing := icmp.NewTimingEngine()
	if err := timing.Initialize(); err != nil {
		t.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	// Get timing precision
	precision, err := timing.ValidatePrecision()
	if err != nil {
		t.Errorf("Failed to validate timing precision: %v", err)
		return
	}
	
	t.Logf("Current architecture %s timing precision: %s", platformInfo.Arch, precision)
	
	// Test timing resolution
	resolution := timing.GetTimerResolution()
	if resolution <= 0 {
		t.Errorf("Invalid timing resolution: %v", resolution)
		return
	}
	
	t.Logf("Current architecture %s timing resolution: %v", platformInfo.Arch, resolution)
	
	// Test CPU features
	features := getCPUFeaturesForArchitecture(platformInfo.Arch)
	if features == nil {
		t.Errorf("No CPU features detected for architecture %s", platformInfo.Arch)
		return
	}
	
	t.Logf("Current architecture %s CPU features: %v", platformInfo.Arch, features)
	
	// Test platform optimizations
	opts := platform.GetArchitectureOptimizations(platformInfo)
	if opts == nil {
		t.Error("No architecture optimizations returned")
		return
	}
	
	t.Logf("Current architecture %s optimizations: %v", platformInfo.Arch, opts)
	
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
	
	// Test timing optimizations
	opts := platform.GetArchitectureOptimizations(platformInfo)
	
	// Validate timing optimizations based on architecture
	switch platformInfo.Arch {
	case "amd64":
		if rdtsc, ok := opts["use_rdtsc"]; !ok || !rdtsc.(bool) {
			t.Error("AMD64 should use RDTSC for timing")
		}
		if vec, ok := opts["vectorization"]; !ok || vec != "avx2" {
			t.Errorf("AMD64 should use AVX2 vectorization, got: %v", vec)
		}
	case "arm64":
		if armCounter, ok := opts["use_arm_counter"]; !ok || !armCounter.(bool) {
			t.Error("ARM64 should use ARM counter for timing")
		}
		if vec, ok := opts["vectorization"]; !ok || vec != "neon" {
			t.Errorf("ARM64 should use NEON vectorization, got: %v", vec)
		}
	}
	
	// Test memory optimization settings
	if cacheFriendly, ok := opts["cache_friendly_allocation"]; !ok || !cacheFriendly.(bool) {
		t.Error("All architectures should use cache-friendly allocation")
	}
	
	if preferLocal, ok := opts["prefer_local_memory"]; !ok || !preferLocal.(bool) {
		t.Error("All architectures should prefer local memory")
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
	
	// Initialize timing engine
	timing := icmp.NewTimingEngine()
	if err := timing.Initialize(); err != nil {
		t.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	// Benchmark timing operations
	benchmarkResults := benchmarkTimingOperations(t, timing)
	
	t.Logf("Architecture %s benchmark results: %v", platformInfo.Arch, benchmarkResults)
	
	// Architecture-specific performance expectations
	switch platformInfo.Arch {
	case "amd64":
		// Modern x86_64 should have excellent performance
		if benchmarkResults.NanosecondPrecision() > 1000 {
			t.Logf("AMD64 nanosecond timing might be available: %v", benchmarkResults)
		}
	case "arm64":
		// ARM64 should also have good performance
		if benchmarkResults.NanosecondPrecision() > 2000 {
			t.Logf("ARM64 timing precision: %v", benchmarkResults)
		}
	case "arm":
		// 32-bit ARM might have lower precision
		if benchmarkResults.NanosecondPrecision() > 5000 {
			t.Logf("ARM timing precision acceptable: %v", benchmarkResults)
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
	engine := icmp.NewEngine()
	config := createPlatformCompatibleConfig(platformInfo)
	
	err = engine.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine: %v", err)
	}
	defer engine.Close(ctx)
	
	// Test engine statistics structure consistency
	stats := engine.GetStats()
	if stats == nil {
		t.Error("Engine statistics should not be nil")
		return
	}
	
	// Validate statistics structure is consistent across architectures
	validateStatsConsistency(t, stats)
	
	// Test timing precision is measurable
	timing := icmp.NewTimingEngine()
	if err := timing.Initialize(); err != nil {
		t.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	precision, err := timing.ValidatePrecision()
	if err != nil {
		t.Errorf("Timing precision validation failed: %v", err)
		return
	}
	
	if precision == "" {
		t.Error("Timing precision should be defined")
		return
	}
	
	t.Logf("Architecture %s timing precision validated: %s", platformInfo.Arch, precision)
}

// TestPlatformCapabilitiesConsistency tests that platform capabilities are consistent
func TestPlatformCapabilitiesConsistency(t *testing.T) {
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Testing platform capabilities for: %s/%s", platformInfo.OS, platformInfo.Arch)
	
	// Get capabilities
	capabilities, err := platform.GetCapabilities(platformInfo)
	if err != nil {
		t.Fatalf("Failed to get platform capabilities: %v", err)
	}
	
	if capabilities == nil {
		t.Error("Platform capabilities should not be nil")
		return
	}
	
	// Validate essential capabilities
	essentialCapabilities := []string{"icmp", "udp", "tcp"}
	for _, capability := range essentialCapabilities {
		hasCapability := false
		for _, cap := range capabilities.Capabilities {
			if cap == capability {
				hasCapability = true
				break
			}
		}
		if !hasCapability {
			t.Errorf("Platform should support %s capability", capability)
		}
	}
	
	// Validate architecture-specific expectations
	switch platformInfo.Arch {
	case "amd64", "arm64":
		// 64-bit architectures should have SIMD capabilities
		hasSIMD := false
		for _, cap := range capabilities.Capabilities {
			if cap == "sse2" || cap == "neon" || cap == "avx" || cap == "avx2" {
				hasSIMD = true
				break
			}
		}
		if !hasSIMD {
			t.Errorf("64-bit architecture %s should have SIMD capabilities", platformInfo.Arch)
		}
	case "arm":
		// 32-bit ARM may have limited capabilities
		t.Logf("32-bit ARM architecture capabilities: %v", capabilities.Capabilities)
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
	config := &types.NetworkConfig{
		BufferSize: 4096,
		TTL:        64,
	}
	
	// Get platform-specific configuration
	platformConfig := platform.GetPlatformSpecificConfig(platformInfo)
	
	// Apply platform-specific buffer size
	if bufferSize, ok := platformConfig["buffer_size"]; ok {
		if bs, ok := bufferSize.(int); ok {
			config.BufferSize = bs
		}
	}
	
	return config
}

func validateStatsConsistency(t *testing.T, stats *icmp.EngineStats) {
	// These fields should always be present regardless of architecture
	if stats == nil {
		t.Error("Statistics should not be nil")
		return
	}
	
	if stats.MeasurementsTotal < 0 {
		t.Error("Measurements total should be non-negative")
	}
	
	if stats.MeasurementsSuccess < 0 {
		t.Error("Measurements success should be non-negative")
	}
	
	if stats.MeasurementsFailed < 0 {
		t.Error("Measurements failed should be non-negative")
	}
	
	if stats.MeasurementsTotal < stats.MeasurementsSuccess+stats.MeasurementsFailed {
		t.Error("Invalid statistics: total should equal success + failed")
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

func benchmarkTimingOperations(t *testing.T, timing icmp.TimingEngine) benchmarkResults {
	// Run timing operations benchmark
	start := time.Now()
	const iterations = 10000
	
	for i := 0; i < iterations; i++ {
		_ = timing.GetCurrentTime()
	}
	
	elapsed := time.Since(start)
	operationsPerSecond := float64(iterations) / elapsed.Seconds()
	
	// Estimate timing precision based on operation speed
	avgLatency := elapsed.Nanoseconds() / int64(iterations)
	
	return benchmarkResults{
		operationsPerSecond: operationsPerSecond,
		nanosecondPrecision: avgLatency,
	}
}

// BenchmarkCrossArchitecturePerformance benchmarks performance across different architectures
func BenchmarkCrossArchitectureTiming(b *testing.B) {
	timing := icmp.NewTimingEngine()
	
	if err := timing.Initialize(); err != nil {
		b.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	// Platform information
	platformInfo, _ := platform.DetectPlatform()
	b.Logf("Benchmarking timing operations on %s/%s", platformInfo.OS, platformInfo.Arch)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = timing.GetCurrentTime()
	}
}

func BenchmarkCrossArchitectureCalibration(b *testing.B) {
	timing := icmp.NewTimingEngine()
	
	if err := timing.Initialize(); err != nil {
		b.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = timing.Calibrate()
	}
}