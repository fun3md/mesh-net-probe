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
	capabilities, err := platform.GetCapabilities(platformInfo)
	if err != nil {
		t.Fatalf("Failed to get platform capabilities: %v", err)
	}

	t.Logf("=== Platform Capabilities ===")
	t.Logf("Precision: %s", capabilities.Precision)
	t.Logf("Capabilities: %v", capabilities.Capabilities)
	t.Logf("Limitations: %v", capabilities.Limitations)
	t.Logf("Custom settings: %v", capabilities.Custom)

	// Test architecture optimizations
	optimizations := platform.GetArchitectureOptimizations(platformInfo)
	if optimizations != nil {
		t.Logf("=== Architecture Optimizations ===")
		for key, value := range optimizations {
			t.Logf("%s: %v", key, value)
		}
	}

	// Test platform-specific configuration
	config := platform.GetPlatformSpecificConfig(platformInfo)
	if config != nil {
		t.Logf("=== Platform-Specific Configuration ===")
		for key, value := range config {
			t.Logf("%s: %v", key, value)
		}
	}

	// Test compatibility validation
	err = platform.ValidatePlatformCompatibility(platformInfo)
	if err != nil {
		t.Errorf("Platform validation failed: %v", err)
	} else {
		t.Logf("✅ Platform validation passed")
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
	capabilities, err := platform.GetCapabilities(platformInfo)
	if err != nil {
		log.Fatalf("Failed to get platform capabilities: %v", err)
	}

	fmt.Printf("=== Platform Capabilities ===\n")
	fmt.Printf("Precision: %s\n", capabilities.Precision)
	fmt.Printf("Capabilities: %v\n", capabilities.Capabilities)
	fmt.Printf("Limitations: %v\n", capabilities.Limitations)
	fmt.Printf("Custom settings: %v\n\n", capabilities.Custom)

	// Test architecture optimizations
	optimizations := platform.GetArchitectureOptimizations(platformInfo)
	if optimizations != nil {
		fmt.Printf("=== Architecture Optimizations ===\n")
		for key, value := range optimizations {
			fmt.Printf("%s: %v\n", key, value)
		}
		fmt.Println()
	}

	// Test platform-specific configuration
	config := platform.GetPlatformSpecificConfig(platformInfo)
	if config != nil {
		fmt.Printf("=== Platform-Specific Configuration ===\n")
		for key, value := range config {
			fmt.Printf("%s: %v\n", key, value)
		}
		fmt.Println()
	}

	// Test compatibility validation
	err = platform.ValidatePlatformCompatibility(platformInfo)
	if err != nil {
		fmt.Printf("=== Platform Compatibility ===\n")
		fmt.Printf("❌ Validation failed: %v\n\n", err)
	} else {
		fmt.Printf("=== Platform Compatibility ===\n")
		fmt.Printf("✅ Platform validation passed\n\n")
	}

	fmt.Printf("=== Cross-Platform Demo Complete ===\n")
	fmt.Printf("Test completed at: %s\n", time.Now().Format(time.RFC3339))
}