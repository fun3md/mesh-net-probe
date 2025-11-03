package platform

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// DetectPlatform returns comprehensive platform and architecture information
func DetectPlatform() (*types.PlatformInfo, error) {
	// Get operating system and architecture
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Get detailed OS version information
	osVersion, kernelVersion, err := getOSVersionDetails(osName)
	if err != nil {
		return nil, fmt.Errorf("failed to get OS version: %w", err)
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Detect if running in container
	isContainer := detectContainerEnvironment()

	return &types.PlatformInfo{
		OS:        osName,
		Arch:      arch,
		Version:   osVersion,
		Kernel:    kernelVersion,
		Hostname:  hostname,
		Container: isContainer,
	}, nil
}

// getOSVersionDetails returns OS version and kernel version based on platform
func getOSVersionDetails(osName string) (osVersion, kernelVersion string, err error) {
	switch osName {
	case "linux":
		return getLinuxVersionDetails()
	case "darwin":
		return getDarwinVersionDetails()
	case "windows":
		return getWindowsVersionDetails()
	default:
		return "unknown", "unknown", fmt.Errorf("unsupported operating system: %s", osName)
	}
}

// getLinuxVersionDetails retrieves Linux distribution and kernel version
func getLinuxVersionDetails() (osVersion, kernelVersion string, err error) {
	// Read /etc/os-release for distribution information
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		osVersion = parseOSReleaseField(string(data), "PRETTY_NAME")
		if osVersion == "" {
			osVersion = parseOSReleaseField(string(data), "NAME") + " " + parseOSReleaseField(string(data), "VERSION")
		}
	} else {
		osVersion = "Linux (unknown distribution)"
	}

	// Get kernel version
	if data, err := os.ReadFile("/proc/version"); err == nil {
		kernelVersion = strings.TrimSpace(string(data))
	} else {
		kernelVersion = "unknown"
	}

	return osVersion, kernelVersion, nil
}

// getDarwinVersionDetails retrieves macOS version information
func getDarwinVersionDetails() (osVersion, kernelVersion string, err error) {
	// Get macOS version using system_profiler
	if output, err := runCommand("system_profiler", "SPSoftwareDataType"); err == nil {
		osVersion = parseDarwinVersion(output)
	} else {
		osVersion = "macOS (version unknown)"
	}

	// Get Darwin kernel version
	kernelVersion = runtime.Version()
	return osVersion, kernelVersion, nil
}

// getWindowsVersionDetails retrieves Windows version information
func getWindowsVersionDetails() (osVersion, kernelVersion string, err error) {
	// Windows version detection would require Windows API calls or registry access
	// For now, provide basic information
	osVersion = "Windows (detailed version not available)"
	kernelVersion = runtime.Version()
	return osVersion, kernelVersion, nil
}

// parseOSReleaseField extracts a specific field from /etc/os-release content
func parseOSReleaseField(content, field string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, field+"=") {
			value := strings.TrimPrefix(line, field+"=")
			return strings.Trim(value, `"`)
		}
	}
	return ""
}

// parseDarwinVersion extracts version from system_profiler output
func parseDarwinVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "version") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "macOS (version parsing failed)"
}

// detectContainerEnvironment checks if running inside a container
func detectContainerEnvironment() bool {
	// Check for common container indicators
	containerFiles := []string{
		"/.dockerenv",
		"/run/.containerenv",
		"/proc/1/cgroup",
	}

	for _, file := range containerFiles {
		if data, err := os.ReadFile(file); err == nil {
			content := string(data)
			if strings.Contains(strings.ToLower(content), "docker") ||
				strings.Contains(strings.ToLower(content), "container") ||
				strings.Contains(strings.ToLower(content), "kubepods") {
				return true
			}
		}
	}

	// Check environment variables
	envVars := []string{"DOCKER", "KUBERNETES", "CONTAINER"}
	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// runCommand executes a system command and returns output
func runCommand(name string, args ...string) (string, error) {
	// This would need proper command execution implementation
	// For now, return error to indicate not implemented
	return "", fmt.Errorf("command execution not implemented")
}

// ValidateCapabilities checks platform-specific capabilities and limitations
func ValidateCapabilities() (*types.PlatformCapabilities, error) {
	platformInfo, err := DetectPlatform()
	if err != nil {
		return nil, fmt.Errorf("failed to detect platform: %w", err)
	}

	capabilities := &types.PlatformCapabilities{
		Platform:  platformInfo,
		CanMeasure: true,
		MaxPrecision: types.PrecisionMicrosecond,
		Limitations: []string{},
		Features: []string{},
	}

	// Platform-specific capability checks
	switch platformInfo.OS {
	case "linux":
		if err := checkLinuxCapabilities(capabilities); err != nil {
			return nil, err
		}
	case "darwin":
		if err := checkDarwinCapabilities(capabilities); err != nil {
			return nil, err
		}
	case "windows":
		if err := checkWindowsCapabilities(capabilities); err != nil {
			return nil, err
		}
	}

	// Architecture-specific adjustments
	if err := checkArchitectureCapabilities(capabilities); err != nil {
		return nil, err
	}

	return capabilities, nil
}

// checkLinuxCapabilities validates Linux-specific capabilities
func checkLinuxCapabilities(capabilities *types.PlatformCapabilities) error {
	// Check if ICMP is available
	if !isICMPSupported() {
		capabilities.CanMeasure = false
		capabilities.Limitations = append(capabilities.Limitations, "ICMP measurement not supported")
	}

	// Check for high-resolution timer support
	if !isHighResTimerSupported() {
		capabilities.Limitations = append(capabilities.Limitations, "High-resolution timers not available")
		capabilities.MaxPrecision = types.PrecisionMillisecond
	}

	// Check for required capabilities
	capabilities.Features = append(capabilities.Features, "Standard ICMP support", "POSIX sockets")

	return nil
}

// checkDarwinCapabilities validates macOS-specific capabilities
func checkDarwinCapabilities(capabilities *types.PlatformCapabilities) error {
	// macOS generally supports ICMP but may require privileges
	capabilities.Features = append(capabilities.Features, "Darwin ICMP support", "BSD sockets")

	// Check for timing precision
	if !isHighResTimerSupported() {
		capabilities.MaxPrecision = types.PrecisionMicrosecond
		capabilities.Limitations = append(capabilities.Limitations, "Limited timing precision")
	}

	return nil
}

// checkWindowsCapabilities validates Windows-specific capabilities
func checkWindowsCapabilities(capabilities *types.PlatformCapabilities) error {
	// Windows has different ICMP handling
	capabilities.Features = append(capabilities.Features, "Windows ICMP support", "WinSock")

	// Windows timing precision may be limited
	capabilities.MaxPrecision = types.PrecisionMicrosecond
	capabilities.Limitations = append(capabilities.Limitations, "Windows timing precision varies")

	return nil
}

// checkArchitectureCapabilities validates architecture-specific capabilities
func checkArchitectureCapabilities(capabilities *types.PlatformCapabilities) error {
	switch capabilities.Platform.Arch {
	case "amd64":
		capabilities.Features = append(capabilities.Features, "x86_64 architecture", "Advanced timing instructions")
	case "arm64":
		capabilities.Features = append(capabilities.Features, "ARM64 architecture", "ARMv8 timing instructions")
	case "arm":
		capabilities.Features = append(capabilities.Features, "ARM architecture", "ARM timing instructions")
	default:
		capabilities.Limitations = append(capabilities.Limitations, fmt.Sprintf("Unknown architecture: %s", capabilities.Platform.Arch))
	}

	return nil
}

// isICMPSupported checks if ICMP is supported on this platform
func isICMPSupported() bool {
	// Basic check - actual implementation would test ICMP socket creation
	return true // Assume supported for now
}

// isHighResTimerSupported checks if high-resolution timers are available
func isHighResTimerSupported() bool {
	// Check for CLOCK_MONOTONIC or similar high-resolution timers
	return true // Assume supported for now
}

// GetOptimalTimingPrecision returns the optimal timing precision for the current platform
func GetOptimalTimingPrecision() (types.MeasurementPrecision, error) {
	capabilities, err := ValidateCapabilities()
	if err != nil {
		return types.PrecisionMillisecond, err
	}

	return capabilities.MaxPrecision, nil
}

// CheckPrivilegeRequirements determines if elevated privileges are needed
func CheckPrivilegeRequirements() (bool, []string, error) {
	platformInfo, err := DetectPlatform()
	if err != nil {
		return true, []string{"Platform detection failed"}, err
	}

	needsPrivileges := false
	missingCapabilities := []string{}

	// Check if root/admin privileges are needed
	switch platformInfo.OS {
	case "linux":
		needsPrivileges = os.Geteuid() != 0
		if needsPrivileges {
			missingCapabilities = append(missingCapabilities, "Root privileges for raw ICMP sockets")
		}
	case "darwin":
		needsPrivileges = os.Geteuid() != 0
		if needsPrivileges {
			missingCapabilities = append(missingCapabilities, "Admin privileges for raw ICMP sockets")
		}
	case "windows":
		// Windows typically doesn't require admin for basic ICMP
		needsPrivileges = false
	}

	return needsPrivileges, missingCapabilities, nil
}

// GetPlatformStats returns runtime statistics about the platform
func GetPlatformStats() (*types.PlatformStats, error) {
	// Get memory information
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := &types.PlatformStats{
		CPUCount:       runtime.NumCPU(),
		GoRoutines:     runtime.NumGoroutine(),
		MemoryAlloc:    memStats.Alloc,
		MemorySys:      memStats.Sys,
		MemoryHeapAlloc: memStats.HeapAlloc,
		MemoryHeapSys:  memStats.HeapSys,
		GCCollections:  memStats.NumGC,
		GCTime:         time.Duration(memStats.PauseTotalNs) * time.Nanosecond,
		Uptime:         time.Since(startTime),
	}

	return stats, nil
}

// Platform info for timing
var startTime = time.Now()