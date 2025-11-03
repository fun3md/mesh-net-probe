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

// TestCrossPlatformICMPEngine tests ICMP engine functionality across different platforms
func TestCrossPlatformICMPEngine(t *testing.T) {
	ctx := context.Background()
	
	// Test current platform compatibility
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Testing on platform: %s/%s (%s)", platformInfo.OS, platformInfo.Arch, runtime.Version())
	
	// Initialize ICMP engine for current platform
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Skipf("ICMP engine initialization failed on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
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
		t.Errorf("ICMP ping failed on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	if measurement == nil {
		t.Error("Measurement result is nil on " + platformInfo.OS + "/" + platformInfo.Arch)
		return
	}
	
	t.Logf("Platform %s/%s: Engine functionality verified successfully", platformInfo.OS, platformInfo.Arch)
}

// TestCrossPlatformTimingEngine tests timing precision across different platforms
func TestCrossPlatformTimingEngine(t *testing.T) {
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Testing timing engine on: %s/%s", platformInfo.OS, platformInfo.Arch)
	
	// Get platform capabilities for timing information
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		t.Errorf("Failed to get capabilities on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	t.Logf("Platform %s/%s: Can measure = %t, Max precision = %s",
		platformInfo.OS, platformInfo.Arch, capabilities.CanMeasure, capabilities.MaxPrecision)
	
	// Get optimal timing precision
	precision, err := platform.GetOptimalTimingPrecision()
	if err != nil {
		t.Errorf("Failed to get timing precision on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	if precision == "" {
		t.Errorf("No precision detected on %s/%s", platformInfo.OS, platformInfo.Arch)
		return
	}
	
	t.Logf("Platform %s/%s: Optimal timing precision = %s", platformInfo.OS, platformInfo.Arch, precision)
	
	// Get platform statistics
	stats, err := platform.GetPlatformStats()
	if err != nil {
		t.Errorf("Failed to get platform stats on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	t.Logf("Platform %s/%s: CPU count = %d, Go routines = %d",
		platformInfo.OS, platformInfo.Arch, stats.CPUCount, stats.GoRoutines)
}

// TestCrossPlatformMeasurementConsistency tests measurement consistency across platforms
func TestCrossPlatformMeasurementConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cross-platform consistency test in short mode")
	}
	
	ctx := context.Background()
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	engine, err := icmp.NewEngine(
		icmp.WithTimeout(5*time.Second),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
	}
	defer engine.Close()
	
	// Test with loopback address for consistency
	target := net.ParseIP("127.0.0.1")
	if target == nil {
		t.Error("Invalid target IP")
		return
	}
	
	successCount := 0
	var totalRTT time.Duration
	const iterations = 5
	
	for i := 0; i < iterations; i++ {
		measurement, err := engine.Ping(ctx, target)
		if err != nil {
			t.Logf("Measurement %d failed on %s/%s: %v", i+1, platformInfo.OS, platformInfo.Arch, err)
			continue
		}
		
		if measurement == nil {
			t.Errorf("Measurement %d returned nil on %s/%s", i+1, platformInfo.OS, platformInfo.Arch)
			continue
		}
		
		if measurement.RTT >= 0 {
			successCount++
			totalRTT += measurement.RTT
		}
		
		// Log measurement details
		t.Logf("Measurement %d on %s/%s: rtt=%v",
			i+1, platformInfo.OS, platformInfo.Arch, measurement.RTT)
		
		// Small delay between measurements
		time.Sleep(10 * time.Millisecond)
	}
	
	// Analyze measurement consistency
	successRate := float64(successCount) / float64(iterations)
	if successCount > 0 {
		t.Logf("Platform %s/%s: Success rate = %.2f%%, Average RTT = %v",
			platformInfo.OS, platformInfo.Arch, successRate*100, totalRTT/time.Duration(successCount))
	}
}

// TestPlatformCapabilities tests platform-specific capabilities detection
func TestPlatformCapabilities(t *testing.T) {
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	// Test platform capabilities
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		t.Errorf("Failed to get capabilities for %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	if capabilities == nil {
		t.Errorf("Capabilities is nil for %s/%s", platformInfo.OS, platformInfo.Arch)
		return
	}
	
	t.Logf("Platform %s/%s capabilities:", platformInfo.OS, platformInfo.Arch)
	t.Logf("  Can measure: %t", capabilities.CanMeasure)
	t.Logf("  Max precision: %s", capabilities.MaxPrecision)
	t.Logf("  Features: %v", capabilities.Features)
	t.Logf("  Limitations: %v", capabilities.Limitations)
	
	// Validate privilege requirements
	needsPrivs, missingCaps, err := platform.CheckPrivilegeRequirements()
	if err != nil {
		t.Errorf("Failed to check privileges on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
	} else {
		t.Logf("Platform %s/%s: Needs privileges = %t, Missing capabilities = %v",
			platformInfo.OS, platformInfo.Arch, needsPrivs, missingCaps)
	}
}

// TestPlatformSpecificConfiguration tests platform-specific configuration handling
func TestPlatformSpecificConfiguration(t *testing.T) {
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	// Get platform capabilities (updated method name)
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		t.Errorf("Failed to get platform capabilities for %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	t.Logf("Platform %s/%s specific configuration:", platformInfo.OS, platformInfo.Arch)
	t.Logf("  Platform: %v", capabilities.Platform)
	t.Logf("  Features: %v", capabilities.Features)
	
	// Basic platform-specific expectations
	switch platformInfo.OS {
	case "linux":
		if len(capabilities.Features) == 0 {
			t.Errorf("Linux should have defined features")
		}
	case "darwin":
		if len(capabilities.Features) == 0 {
			t.Errorf("macOS should have defined features")
		}
	case "windows":
		if len(capabilities.Features) == 0 {
			t.Errorf("Windows should have defined features")
		}
	}
	
	// Architecture-specific expectations
	switch platformInfo.Arch {
	case "arm64":
		if !capabilities.CanMeasure {
			t.Errorf("ARM64 should support measurements")
		}
	case "amd64":
		if !capabilities.CanMeasure {
			t.Errorf("AMD64 should support measurements")
		}
	}
}

// TestCrossPlatformNetworkInterfaces tests network interface handling across platforms
func TestCrossPlatformNetworkInterfaces(t *testing.T) {
	// Get network interfaces
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("Failed to get network interfaces: %v", err)
	}
	
	t.Logf("Available network interfaces (%d):", len(addrs))
	for i, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok {
			t.Logf("  Interface %d: %s (local=%v)", i, ipNet.IP.String(), ipNet.IP.IsLoopback())
		} else {
			t.Logf("  Interface %d: %v", i, addr)
		}
	}
	
	// Test finding a valid source IP for measurements
	validIPs := findValidSourceIPs()
	if len(validIPs) == 0 {
		t.Error("No valid source IPs found")
	} else {
		t.Logf("Valid source IPs found: %v", validIPs)
	}
}

// TestPlatformOperatingSystemSupport tests support for different operating systems
func TestPlatformOperatingSystemSupport(t *testing.T) {
	supportedOS := []string{"linux", "darwin", "windows"}
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	// Check if current OS is supported
	isSupported := false
	for _, os := range supportedOS {
		if platformInfo.OS == os {
			isSupported = true
			break
		}
	}
	
	if !isSupported {
		t.Errorf("Unsupported operating system: %s", platformInfo.OS)
	} else {
		t.Logf("Operating system %s is supported", platformInfo.OS)
	}
}

// Helper functions

func createPlatformSpecificConfig(platformInfo *types.PlatformInfo) *types.NetworkConfig {
	// Return a basic configuration - the ICMP engine uses functional options
	// This function is kept for compatibility but not used in the updated API
	return &types.NetworkConfig{
		BufferSize: 4096,
		TTL:        64,
	}
}

func validateConfigKey(t *testing.T, config map[string]interface{}, key string, expectedValue interface{}) {
	if value, exists := config[key]; exists {
		if value != expectedValue {
			t.Errorf("Config key %s: expected %v, got %v", key, expectedValue, value)
		} else {
			t.Logf("Config key %s: %v (correct)", key, value)
		}
	} else {
		t.Errorf("Missing config key: %s", key)
	}
}

func findValidSourceIPs() []net.IP {
	var validIPs []net.IP
	addrs, _ := net.InterfaceAddrs()
	
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip := ipNet.IP.To4(); ip != nil {
				validIPs = append(validIPs, ip)
			}
		}
	}
	
	return validIPs
}

// BenchmarkCrossPlatformPerformance benchmarks performance across different platforms
func BenchmarkCrossPlatformTiming(b *testing.B) {
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
	b.Logf("Benchmarking ICMP on %s/%s", platformInfo.OS, platformInfo.Arch)
	
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

func BenchmarkCrossPlatformBatch(b *testing.B) {
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