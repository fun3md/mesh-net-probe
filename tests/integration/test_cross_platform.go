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
	engine := icmp.NewEngine()
	config := createPlatformSpecificConfig(platformInfo)
	
	err = engine.Initialize(ctx, config)
	if err != nil {
		t.Skipf("ICMP engine initialization failed on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
	}
	defer engine.Close(ctx)
	
	// Test engine statistics
	stats := engine.GetStats()
	if stats == nil {
		t.Error("Engine statistics are nil on " + platformInfo.OS + "/" + platformInfo.Arch)
	}
	
	t.Logf("Platform %s/%s: Engine stats collected successfully", platformInfo.OS, platformInfo.Arch)
}

// TestCrossPlatformTimingEngine tests timing precision across different platforms
func TestCrossPlatformTimingEngine(t *testing.T) {
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	t.Logf("Testing timing engine on: %s/%s", platformInfo.OS, platformInfo.Arch)
	
	timing := icmp.NewTimingEngine()
	
	// Initialize timing engine
	err = timing.Initialize()
	if err != nil {
		t.Errorf("Timing engine initialization failed on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	// Test timing precision
	precision, err := timing.ValidatePrecision()
	if err != nil {
		t.Errorf("Precision validation failed on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	if precision == "" {
		t.Errorf("No precision detected on %s/%s", platformInfo.OS, platformInfo.Arch)
		return
	}
	
	t.Logf("Platform %s/%s: Timing precision = %s", platformInfo.OS, platformInfo.Arch, precision)
	
	// Test timing resolution
	resolution := timing.GetTimerResolution()
	if resolution <= 0 {
		t.Errorf("Invalid timer resolution on %s/%s: %v", platformInfo.OS, platformInfo.Arch, resolution)
	}
	
	// Test system information
	systemInfo := timing.GetSystemInfo()
	if systemInfo == nil {
		t.Errorf("System info is nil on %s/%s", platformInfo.OS, platformInfo.Arch)
	} else {
		t.Logf("Platform %s/%s: Timer resolution = %v, CPU cores = %d", 
			platformInfo.OS, platformInfo.Arch, systemInfo.TimerResolution, systemInfo.CPULogicalCores)
	}
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
	
	engine := icmp.NewEngine()
	config := createPlatformSpecificConfig(platformInfo)
	
	err = engine.Initialize(ctx, config)
	if err != nil {
		t.Skipf("Cannot initialize ICMP engine on %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
	}
	defer engine.Close(ctx)
	
	// Test with loopback address for consistency
	target := &types.NetworkTarget{
		ID:      "consistency_test",
		Address: net.ParseIP("127.0.0.1"),
		Timeout: 2 * time.Second,
		Enabled: true,
	}
	
	measurements := make([]*types.MeasurementData, 5)
	for i := 0; i < 5; i++ {
		measurement, err := engine.Measure(ctx, target)
		if err != nil {
			t.Logf("Measurement %d failed on %s/%s: %v", i+1, platformInfo.OS, platformInfo.Arch, err)
			continue
		}
		
		if measurement == nil {
			t.Errorf("Measurement %d returned nil on %s/%s", i+1, platformInfo.OS, platformInfo.Arch)
			continue
		}
		
		measurements[i] = measurement
		
		// Log measurement details
		t.Logf("Measurement %d on %s/%s: success=%v, rtt=%v", 
			i+1, platformInfo.OS, platformInfo.Arch, measurement.Success, measurement.RTT)
		
		// Small delay between measurements
		time.Sleep(10 * time.Millisecond)
	}
	
	// Analyze measurement consistency
	successCount := 0
	var totalRTT time.Duration
	for _, m := range measurements {
		if m != nil && m.Success {
			successCount++
			totalRTT += m.RTT
		}
	}
	
	successRate := float64(successCount) / float64(len(measurements))
	if len(measurements) > 0 {
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
	capabilities, err := platform.GetCapabilities(platformInfo)
	if err != nil {
		t.Errorf("Failed to get capabilities for %s/%s: %v", platformInfo.OS, platformInfo.Arch, err)
		return
	}
	
	if capabilities == nil {
		t.Errorf("Capabilities is nil for %s/%s", platformInfo.OS, platformInfo.Arch)
		return
	}
	
	t.Logf("Platform %s/%s capabilities:", platformInfo.OS, platformInfo.Arch)
	t.Logf("  Precision: %s", capabilities.Precision)
	t.Logf("  Capabilities: %v", capabilities.Capabilities)
	t.Logf("  Limitations: %v", capabilities.Limitations)
	
	// Validate platform compatibility
	err = platform.ValidatePlatformCompatibility(platformInfo)
	if err != nil {
		t.Errorf("Platform %s/%s validation failed: %v", platformInfo.OS, platformInfo.Arch, err)
	} else {
		t.Logf("Platform %s/%s validation passed", platformInfo.OS, platformInfo.Arch)
	}
}

// TestPlatformSpecificConfiguration tests platform-specific configuration handling
func TestPlatformSpecificConfiguration(t *testing.T) {
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}
	
	// Get platform-specific configuration
	config := platform.GetPlatformSpecificConfig(platformInfo)
	
	t.Logf("Platform %s/%s specific configuration:", platformInfo.OS, platformInfo.Arch)
	for key, value := range config {
		t.Logf("  %s: %v", key, value)
	}
	
	// Validate expected configuration keys based on platform
	switch platformInfo.OS {
	case "linux":
		validateConfigKey(t, config, "use_raw_sockets", true)
		validateConfigKey(t, config, "buffer_size", 8192)
	case "darwin":
		validateConfigKey(t, config, "use_raw_sockets", false)
		validateConfigKey(t, config, "buffer_size", 4096)
	case "windows":
		validateConfigKey(t, config, "use_raw_sockets", false)
		validateConfigKey(t, config, "buffer_size", 2048)
	}
	
	// Validate architecture-specific settings
	switch platformInfo.Arch {
	case "arm64":
		validateConfigKey(t, config, "optimize_for_arm", true)
	case "amd64":
		validateConfigKey(t, config, "optimize_for_x86", true)
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
	config := &types.NetworkConfig{
		BufferSize: 4096,
		TTL:        64,
	}
	
	// Get platform-specific adjustments
	platformConfig := platform.GetPlatformSpecificConfig(platformInfo)
	
	// Apply platform-specific settings
	if bufferSize, ok := platformConfig["buffer_size"]; ok {
		if bs, ok := bufferSize.(int); ok {
			config.BufferSize = bs
		}
	}
	
	return config
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
	timing := icmp.NewTimingEngine()
	
	if err := timing.Initialize(); err != nil {
		b.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	// Platform information
	platformInfo, _ := platform.DetectPlatform()
	b.Logf("Benchmarking timing on %s/%s", platformInfo.OS, platformInfo.Arch)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = timing.GetCurrentTime()
	}
}

func BenchmarkCrossPlatformCalibration(b *testing.B) {
	timing := icmp.NewTimingEngine()
	
	if err := timing.Initialize(); err != nil {
		b.Fatalf("Failed to initialize timing engine: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = timing.Calibrate()
	}
}