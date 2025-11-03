package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mesh-net-probe/probe/internal/platform"
)

func main() {
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

	fmt.Printf("=== Cross-Platform Test Complete ===\n")
	fmt.Printf("Test completed at: %s\n", time.Now().Format(time.RFC3339))
}