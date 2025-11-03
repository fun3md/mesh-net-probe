package platform

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// DetectPlatform detects the current platform and architecture information
func DetectPlatform() (*types.PlatformInfo, error) {
	osInfo, err := getOSInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to detect OS information: %w", err)
	}

	archInfo, err := getArchitectureInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to detect architecture information: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return &types.PlatformInfo{
		OS:        osInfo.name,
		Arch:      archInfo.name,
		Version:   osInfo.version,
		Kernel:    osInfo.kernel,
		Hostname:  hostname,
		Container: isRunningInContainer(),
	}, nil
}

// OSInfo represents operating system information
type OSInfo struct {
	name    string
	version string
	kernel  string
}

// getOSInfo detects the operating system information
func getOSInfo() (*OSInfo, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxInfo()
	case "darwin":
		return getDarwinInfo()
	case "windows":
		return getWindowsInfo()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// getLinuxInfo retrieves Linux-specific system information
func getLinuxInfo() (*OSInfo, error) {
	osRelease, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/os-release: %w", err)
	}

	osInfo := &OSInfo{
		name: "linux",
	}

	lines := strings.Split(string(osRelease), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			osInfo.version = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
	}

	// Get kernel version
	if kernel, err := os.ReadFile("/proc/version"); err == nil {
		osInfo.kernel = strings.TrimSpace(string(kernel))
	}

	return osInfo, nil
}

// getDarwinInfo retrieves macOS-specific system information
func getDarwinInfo() (*OSInfo, error) {
	osInfo := &OSInfo{
		name: "darwin",
	}

	// Get macOS version using system_profiler
	// Note: In a real implementation, you would call system_profiler SPSoftwareDataType
	// For now, we'll use a basic approach
	osInfo.version = "macOS (version detection not implemented)"

	// Get kernel version
	if kernel, err := os.ReadFile("/System/Library/Kernels/kernel"); err == nil {
		osInfo.kernel = strings.TrimSpace(string(kernel))
	}

	return osInfo, nil
}

// getWindowsInfo retrieves Windows-specific system information
func getWindowsInfo() (*OSInfo, error) {
	osInfo := &OSInfo{
		name: "windows",
	}

	// In a real implementation, you would use Windows API calls
	// For now, we'll use basic runtime information
	osInfo.version = fmt.Sprintf("Windows (build %s)", runtime.GOOS)
	osInfo.kernel = "Windows Kernel"

	return osInfo, nil
}

// ArchitectureInfo represents processor architecture information
type ArchitectureInfo struct {
	name string
	bits int
}

// getArchitectureInfo detects the processor architecture
func getArchitectureInfo() (*ArchitectureInfo, error) {
	var archInfo ArchitectureInfo

	switch runtime.GOARCH {
	case "amd64":
		archInfo.name = "amd64"
		archInfo.bits = 64
	case "arm64":
		archInfo.name = "arm64"
		archInfo.bits = 64
	case "arm":
		archInfo.name = "arm"
		// Determine ARM bits
		if bits := getPointerSize(); bits != 0 {
			archInfo.bits = bits
		}
	case "386":
		archInfo.name = "386"
		archInfo.bits = 32
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	return &archInfo, nil
}

// getPointerSize determines the pointer size in bits
func getPointerSize() int {
	// This is a heuristic approach - in practice, you'd use platform-specific methods
	if strings.Contains(runtime.GOARCH, "64") {
		return 64
	}
	return 32
}

// isRunningInContainer detects if the current process is running inside a container
func isRunningInContainer() bool {
	// Check for common container environment variables
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" ||
		os.Getenv("DOCKER_HOST") != "" ||
		os.Getenv("KUBERNETES_PORT") != "" {
		return true
	}

	// Check for container-specific files
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for cgroup indicators
	if cgroup, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		if strings.Contains(string(cgroup), "docker") ||
		   strings.Contains(string(cgroup), "kubepods") ||
		   strings.Contains(string(cgroup), "container") {
			return true
		}
	}

	return false
}

// GetCapabilities returns the platform-specific capabilities for ICMP operations
func GetCapabilities(platform *types.PlatformInfo) (*types.PlatformMeta, error) {
	capabilities := []string{}
	limitations := []string{}
	precision := types.PrecisionMicrosecond // Default to microsecond

	// Common capabilities
	capabilities = append(capabilities, "icmp", "udp", "tcp")

	// Platform-specific capabilities and limitations
	switch platform.OS {
	case "linux":
		capabilities = append(capabilities, "raw_sockets", "cap_net_raw")
		if isRunningInContainer() {
			limitations = append(limitations, "containerized_environment")
			if !hasNetRawCapability() {
				limitations = append(limitations, "no_cap_net_raw")
			}
		}
		precision = getLinuxPrecision()
	case "darwin":
		capabilities = append(capabilities, "bpf")
		limitations = append(limitations, "icmp_rate_limited")
		precision = types.PrecisionMicrosecond
	case "windows":
		capabilities = append(capabilities, "winsock")
		limitations = append(limitations, "icmp_rate_limited", "no_raw_sockets")
		precision = types.PrecisionMillisecond
	}

	// Architecture-specific considerations
	switch platform.Arch {
	case "amd64":
		capabilities = append(capabilities, "sse2", "aes")
	case "arm64":
		capabilities = append(capabilities, "neon", "aes")
	}

	return &types.PlatformMeta{
		OS:           platform.OS,
		Architecture: platform.Arch,
		Precision:    precision,
		Capabilities: capabilities,
		Limitations:  limitations,
		Custom: map[string]interface{}{
			"go_version": runtime.Version(),
			"go_os":      runtime.GOOS,
			"go_arch":    runtime.GOARCH,
		},
	}, nil
}

// hasNetRawCapability checks if the current process has CAP_NET_RAW capability
func hasNetRawCapability() bool {
	// In a container environment, this check would need to be more sophisticated
	// For now, we assume containers with proper setup have the capability
	return true
}

// getLinuxPrecision determines the timing precision available on Linux
func getLinuxPrecision() types.MeasurementPrecision {
	// Check for high-resolution timer availability
	if hpet, err := os.Stat("/dev/hpet"); err == nil && !hpet.IsDir() {
		return types.PrecisionNanosecond
	}
	
	// Check for TSC availability
	if _, err := os.Stat("/sys/devices/system/clocksource/clocksource0/current_clocksource"); err == nil {
		if clocksource, err := os.ReadFile("/sys/devices/system/clocksource/clocksource0/current_clocksource"); err == nil {
			if strings.Contains(string(clocksource), "tsc") {
				return types.PrecisionNanosecond
			}
		}
	}
	
	return types.PrecisionMicrosecond
}

// ValidatePlatformCompatibility checks if the platform meets minimum requirements
func ValidatePlatformCompatibility(platform *types.PlatformInfo) error {
	// Check OS compatibility
	if !isSupportedOS(platform.OS) {
		return fmt.Errorf("unsupported operating system: %s", platform.OS)
	}

	// Check architecture compatibility
	if !isSupportedArchitecture(platform.Arch) {
		return fmt.Errorf("unsupported architecture: %s", platform.Arch)
	}

	// Check for required capabilities
	if platform.OS == "windows" {
		return fmt.Errorf("Windows platform requires additional configuration for ICMP operations")
	}

	return nil
}

// isSupportedOS checks if the operating system is supported
func isSupportedOS(os string) bool {
	supported := []string{"linux", "darwin", "windows"}
	for _, supportedOS := range supported {
		if os == supportedOS {
			return true
		}
	}
	return false
}

// isSupportedArchitecture checks if the architecture is supported
func isSupportedArchitecture(arch string) bool {
	supported := []string{"amd64", "arm64"}
	for _, supportedArch := range supported {
		if arch == supportedArch {
			return true
		}
	}
	return false
}

// GetPlatformSpecificConfig returns platform-specific configuration adjustments
func GetPlatformSpecificConfig(platform *types.PlatformInfo) map[string]interface{} {
	config := make(map[string]interface{})

	switch platform.OS {
	case "linux":
		config["use_raw_sockets"] = true
		config["icmp_idle_timeout"] = "30s"
		config["buffer_size"] = 8192
	case "darwin":
		config["use_raw_sockets"] = false
		config["icmp_idle_timeout"] = "10s"
		config["buffer_size"] = 4096
	case "windows":
		config["use_raw_sockets"] = false
		config["icmp_idle_timeout"] = "5s"
		config["buffer_size"] = 2048
	}

	// Architecture-specific adjustments
	switch platform.Arch {
	case "arm64":
		config["optimize_for_arm"] = true
	case "amd64":
		config["optimize_for_x86"] = true
	}

	return config
}