package tests

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mesh-net-probe/probe/internal/platform"
)

// TestCrossPlatformFunctionality tests cross-platform functionality
func TestCrossPlatformFunctionality(t *testing.T) {
	// Test platform detection
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}

	t.Logf("=== Cross-Platform Detection Test ===")
	t.Logf("OS: %s", platformInfo.OS)
	t.Logf("Architecture: %s", platformInfo.Arch)
	t.Logf("Version: %s", platformInfo.Version)
	t.Logf("Kernel: %s", platformInfo.Kernel)
	t.Logf("Hostname: %s", platformInfo.Hostname)
	t.Logf("Container: %t", platformInfo.Container)

	// Test platform capabilities
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		t.Fatalf("Failed to get platform capabilities: %v", err)
	}

	t.Logf("=== Platform Capabilities ===")
	t.Logf("Can Measure: %t", capabilities.CanMeasure)
	t.Logf("Max Precision: %s", capabilities.MaxPrecision)
	t.Logf("Features: %v", capabilities.Features)
	t.Logf("Limitations: %v", capabilities.Limitations)

	// Test optimal timing precision
	precision, err := platform.GetOptimalTimingPrecision()
	if err != nil {
		t.Errorf("Error getting timing precision: %v", err)
	} else {
		t.Logf("=== Timing Precision ===")
		t.Logf("Optimal Precision: %s", precision)
	}

	// Test privilege requirements
	needsPrivs, missingCaps, err := platform.CheckPrivilegeRequirements()
	if err != nil {
		t.Errorf("Error checking privileges: %v", err)
	} else {
		t.Logf("=== Privilege Check ===")
		t.Logf("Needs Privileges: %t", needsPrivs)
		t.Logf("Missing Capabilities: %v", missingCaps)
	}

	// Test platform statistics
	stats, err := platform.GetPlatformStats()
	if err != nil {
		t.Errorf("Error getting platform stats: %v", err)
	} else {
		t.Logf("=== Platform Statistics ===")
		t.Logf("CPU Count: %d", stats.CPUCount)
		t.Logf("Memory Allocated: %d bytes", stats.MemoryAlloc)
		t.Logf("Go Routines: %d", stats.GoRoutines)
	}

	t.Logf("=== Cross-Platform Test Complete ===")
	t.Logf("Test completed at: %s", time.Now().Format(time.RFC3339))
}

// RunCrossPlatformDemo demonstrates cross-platform functionality
func RunCrossPlatformDemo() {
	// Test platform detection
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		log.Fatalf("Failed to detect platform: %v", err)
	}

	fmt.Printf("=== Cross-Platform Detection Test ===\n")
	fmt.Printf("OS: %s\n", platformInfo.OS)
	fmt.Printf("Architecture: %s\n", platformInfo.Arch)
	fmt.Printf("Version: %s\n", platformInfo.Version)
	fmt.Printf("Kernel: %s\n", platformInfo.Kernel)
	fmt.Printf("Hostname: %s\n", platformInfo.Hostname)
	fmt.Printf("Container: %t\n\n", platformInfo.Container)

	// Test platform capabilities
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		log.Fatalf("Failed to get platform capabilities: %v", err)
	}

	fmt.Printf("=== Platform Capabilities ===\n")
	fmt.Printf("Can Measure: %t\n", capabilities.CanMeasure)
	fmt.Printf("Max Precision: %s\n", capabilities.MaxPrecision)
	fmt.Printf("Features: %v\n", capabilities.Features)
	fmt.Printf("Limitations: %v\n", capabilities.Limitations)
	fmt.Printf("Platform: %v\n\n", capabilities.Platform)

	// Test optimal timing precision
	precision, err := platform.GetOptimalTimingPrecision()
	if err != nil {
		fmt.Printf("=== Timing Precision ===\n")
		fmt.Printf("Error getting precision: %v\n", err)
	} else {
		fmt.Printf("=== Timing Precision ===\n")
		fmt.Printf("Optimal Precision: %s\n\n", precision)
	}

	// Test privilege requirements
	needsPrivs, missingCaps, err := platform.CheckPrivilegeRequirements()
	if err != nil {
		fmt.Printf("=== Privilege Check ===\n")
		fmt.Printf("Error checking privileges: %v\n", err)
	} else {
		fmt.Printf("=== Privilege Check ===\n")
		fmt.Printf("Needs Privileges: %t\n", needsPrivs)
		fmt.Printf("Missing Capabilities: %v\n\n", missingCaps)
	}

	// Test platform statistics
	stats, err := platform.GetPlatformStats()
	if err != nil {
		fmt.Printf("=== Platform Statistics ===\n")
		fmt.Printf("Error getting stats: %v\n\n", err)
	} else {
		fmt.Printf("=== Platform Statistics ===\n")
		fmt.Printf("CPU Count: %d\n", stats.CPUCount)
		fmt.Printf("Memory Allocated: %d bytes\n", stats.MemoryAlloc)
		fmt.Printf("Go Routines: %d\n", stats.GoRoutines)
		fmt.Printf("Uptime: %v\n\n", stats.Uptime)
	}

	fmt.Printf("=== Cross-Platform Demo Complete ===\n")
	fmt.Printf("Test completed at: %s\n", time.Now().Format(time.RFC3339))
}