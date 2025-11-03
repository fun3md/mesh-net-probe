package config

import (
	"fmt"
	"runtime"
	"time"

	"github.com/mesh-net-probe/probe/internal/platform"
	"github.com/mesh-net-probe/probe/pkg/types"
)

// PlatformConfig provides platform-specific configuration handling
type PlatformConfig struct {
	// Platform-specific settings
	LinuxConfig   *LinuxConfig   `json:"linux_config"`
	MacOSConfig   *MacOSConfig   `json:"macos_config"`
	WindowsConfig *WindowsConfig `json:"windows_config"`
	
	// Cross-platform timing settings
	Timing *PlatformTimingConfig `json:"timing"`
	
	// Cross-platform security settings
	Security *PlatformSecurityConfig `json:"security"`
}

// LinuxConfig contains Linux-specific configuration
type LinuxConfig struct {
	// Network settings
	BindInterface string `json:"bind_interface"` // Default network interface
	UseRawSockets bool   `json:"use_raw_sockets"` // Use raw ICMP sockets (requires root)
	
	// Timing settings
	ClockSource string `json:"clock_source"` // "monotonic", "hpet", "tsc"
	
	// Privileged operations
	RequireRoot      bool     `json:"require_root"`      // ICMP operations require root
	CapNetAdmin      bool     `json:"cap_net_admin"`     // Need CAP_NET_ADMIN capability
	AllowedUsers     []string `json:"allowed_users"`     // Users allowed to run ICMP
}

// MacOSConfig contains macOS-specific configuration
type MacOSConfig struct {
	// Network settings
	BindInterface string `json:"bind_interface"` // Default network interface
	UseICMP       bool   `json:"use_icmp"`       // Use ICMP sockets (requires elevated privileges)
	
	// Timing settings
	TimingMode string `json:"timing_mode"` // "high_resolution", "monotonic"
	
	// Security settings
	RequireAdmin bool     `json:"require_admin"` // Operations require admin privileges
	AllowedUsers []string `json:"allowed_users"` // Users allowed to run ICMP
}

// WindowsConfig contains Windows-specific configuration
type WindowsConfig struct {
	// Network settings
	BindInterface string `json:"bind_interface"` // Default network interface
	UseWSA        bool   `json:"use_wsa"`        // Use Windows Sockets API
	
	// Timing settings
	QueryPerformanceCounter bool `json:"query_performance_counter"` // Use high-resolution timers
	PerformanceFrequency    int64 `json:"performance_frequency"`      // Timer frequency
	
	// Security settings
	RequireAdmin bool `json:"require_admin"` // Operations require admin privileges (for raw sockets)
}

// PlatformTimingConfig contains cross-platform timing settings
type PlatformTimingConfig struct {
	// High-resolution timing
	EnableHRTimer     bool `json:"enable_hr_timer"`      // Use high-resolution timers when available
	HRTimerFallbackUS int  `json:"hr_timer_fallback_us"` // Fallback to microseconds if HR unavailable
	
	// Timing precision targets
	TargetPrecision   string `json:"target_precision"`    // "nanosecond", "microsecond", "millisecond"
	ClockSyncEnabled  bool   `json:"clock_sync_enabled"`  // Enable clock synchronization
	ClockSyncInterval int    `json:"clock_sync_interval"` // Sync interval in seconds
	
	// Performance tuning
	OptimizationLevel string `json:"optimization_level"` // "none", "basic", "aggressive"
	CacheFriendly     bool   `json:"cache_friendly"`      // Optimize for cache performance
}

// PlatformSecurityConfig contains cross-platform security settings
type PlatformSecurityConfig struct {
	// Privilege requirements
	ICMPPrivilegeRequired bool `json:"icmp_privilege_required"` // ICMP requires elevated privileges
	MinimumPrivilegeLevel string `json:"minimum_privilege_level"` // "none", "user", "elevated", "root"
	
	// Capability checking
	Capabilities []string `json:"capabilities"` // Required capabilities
	
	// Security hardening
	EnableASLR bool `json:"enable_aslr"` // Address Space Layout Randomization
	EnableDEP  bool `json:"enable_dep"`  // Data Execution Prevention
}

// GetPlatformConfig returns the appropriate platform-specific configuration
func GetPlatformConfig() (*PlatformConfig, error) {
	platformInfo, err := DetectPlatform()
	if err != nil {
		return nil, fmt.Errorf("failed to detect platform: %w", err)
	}

	return NewPlatformConfig(platformInfo)
}

// NewPlatformConfig creates a new platform configuration based on detected platform
func NewPlatformConfig(platform *types.PlatformInfo) (*PlatformConfig, error) {
	config := &PlatformConfig{
		Timing: &PlatformTimingConfig{
			EnableHRTimer:         true,
			HRTimerFallbackUS:     1000, // 1ms fallback
			TargetPrecision:       "microsecond",
			ClockSyncEnabled:      true,
			ClockSyncInterval:     60, // 1 minute
			OptimizationLevel:     "basic",
			CacheFriendly:         true,
		},
		Security: &PlatformSecurityConfig{
			ICMPPrivilegeRequired: false,
			MinimumPrivilegeLevel: "user",
			EnableASLR:            true,
			EnableDEP:             true,
		},
	}

	// Configure based on operating system
	switch platform.OS {
	case "linux":
		config.LinuxConfig = &LinuxConfig{
			BindInterface: "",
			UseRawSockets: false,
			ClockSource:   "monotonic",
			RequireRoot:   false,
			CapNetAdmin:   false,
			AllowedUsers:  []string{},
		}
		
		// Linux typically requires root for raw sockets
		config.Security.ICMPPrivilegeRequired = true
		config.Security.MinimumPrivilegeLevel = "root"
		
	case "darwin":
		config.MacOSConfig = &MacOSConfig{
			BindInterface: "",
			UseICMP:       true,
			TimingMode:    "high_resolution",
			RequireAdmin:  false,
			AllowedUsers:  []string{},
		}
		
		// macOS requires admin for raw ICMP
		config.Security.MinimumPrivilegeLevel = "user"
		
	case "windows":
		config.WindowsConfig = &WindowsConfig{
			BindInterface:           "",
			UseWSA:                  true,
			QueryPerformanceCounter: true,
			RequireAdmin:            false,
		}
		
		// Windows rarely requires admin for basic ICMP
		config.Security.MinimumPrivilegeLevel = "user"
		
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", platform.OS)
	}

	// Configure timing based on architecture
	switch platform.Arch {
	case "amd64", "x86_64":
		config.Timing.OptimizationLevel = "aggressive"
		config.Timing.CacheFriendly = true
	case "arm64":
		config.Timing.OptimizationLevel = "basic"
		config.Timing.CacheFriendly = true
	case "arm":
		config.Timing.OptimizationLevel = "basic"
		config.Timing.CacheFriendly = false
	}

	return config, nil
}

// ApplyConfig applies platform-specific configuration to the base config
func (pc *PlatformConfig) ApplyConfig(baseConfig *types.Configuration) error {
	platformInfo, err := DetectPlatform()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	// Apply network settings
	if baseConfig.Network == nil {
		baseConfig.Network = &types.NetworkConfig{}
	}

	switch platformInfo.OS {
	case "linux":
		if pc.LinuxConfig != nil {
			baseConfig.Network.Interface = pc.LinuxConfig.BindInterface
		}
	case "darwin":
		if pc.MacOSConfig != nil {
			baseConfig.Network.Interface = pc.MacOSConfig.BindInterface
		}
	case "windows":
		if pc.WindowsConfig != nil {
			baseConfig.Network.Interface = pc.WindowsConfig.BindInterface
		}
	}

	// Apply security settings
	if baseConfig.Auth == nil {
		baseConfig.Auth = &types.AuthConfig{}
	}

	// Note: Actual privilege checking is done at runtime, not config time
	return nil
}

// ValidateConfig validates platform-specific configuration
func (pc *PlatformConfig) ValidateConfig(platform *types.PlatformInfo) []error {
	var errors []error

	// Validate timing settings
	if pc.Timing != nil {
		if pc.Timing.TargetPrecision != "nanosecond" && 
		   pc.Timing.TargetPrecision != "microsecond" && 
		   pc.Timing.TargetPrecision != "millisecond" {
			errors = append(errors, fmt.Errorf("invalid target precision: %s", pc.Timing.TargetPrecision))
		}
		
		if pc.Timing.ClockSyncInterval < 1 || pc.Timing.ClockSyncInterval > 3600 {
			errors = append(errors, fmt.Errorf("invalid clock sync interval: %d (must be 1-3600 seconds)", pc.Timing.ClockSyncInterval))
		}
		
		if pc.Timing.OptimizationLevel != "none" && 
		   pc.Timing.OptimizationLevel != "basic" && 
		   pc.Timing.OptimizationLevel != "aggressive" {
			errors = append(errors, fmt.Errorf("invalid optimization level: %s", pc.Timing.OptimizationLevel))
		}
	}

	// Validate security settings
	if pc.Security != nil {
		if pc.Security.MinimumPrivilegeLevel != "none" && 
		   pc.Security.MinimumPrivilegeLevel != "user" && 
		   pc.Security.MinimumPrivilegeLevel != "elevated" && 
		   pc.Security.MinimumPrivilegeLevel != "root" {
			errors = append(errors, fmt.Errorf("invalid minimum privilege level: %s", pc.Security.MinimumPrivilegeLevel))
		}
	}

	// Platform-specific validation
	switch platform.OS {
	case "linux":
		if pc.LinuxConfig != nil {
			if pc.LinuxConfig.ClockSource != "" && 
			   pc.LinuxConfig.ClockSource != "monotonic" && 
			   pc.LinuxConfig.ClockSource != "hpet" && 
			   pc.LinuxConfig.ClockSource != "tsc" {
				errors = append(errors, fmt.Errorf("invalid Linux clock source: %s", pc.LinuxConfig.ClockSource))
			}
		}
	case "darwin":
		if pc.MacOSConfig != nil {
			if pc.MacOSConfig.TimingMode != "high_resolution" && 
			   pc.MacOSConfig.TimingMode != "monotonic" {
				errors = append(errors, fmt.Errorf("invalid macOS timing mode: %s", pc.MacOSConfig.TimingMode))
			}
		}
	case "windows":
		if pc.WindowsConfig != nil {
			if pc.WindowsConfig.PerformanceFrequency < 0 {
				errors = append(errors, fmt.Errorf("invalid Windows performance frequency: %d", pc.WindowsConfig.PerformanceFrequency))
			}
		}
	}

	return errors
}

// GetOptimalTimingSource returns the optimal timing source for the current platform
func (pc *PlatformConfig) GetOptimalTimingSource(platform *types.PlatformInfo) (TimingSource, error) {
	switch platform.OS {
	case "linux":
		return pc.getLinuxTimingSource()
	case "darwin":
		return pc.getMacOSTimingSource()
	case "windows":
		return pc.getWindowsTimingSource()
	default:
		return SimpleTimingSource{}, fmt.Errorf("unsupported operating system for timing: %s", platform.OS)
	}
}

// getLinuxTimingSource returns the optimal Linux timing source
func (pc *PlatformConfig) getLinuxTimingSource() (TimingSource, error) {
	clockSource := "monotonic"
	if pc.LinuxConfig != nil && pc.LinuxConfig.ClockSource != "" {
		clockSource = pc.LinuxConfig.ClockSource
	}

	switch clockSource {
	case "monotonic":
		return SimpleTimingSource{}, nil
	case "hpet":
		return SimpleTimingSource{}, nil // HPET detection would be done at runtime
	case "tsc":
		return SimpleTimingSource{}, nil // TSC detection would be done at runtime
	default:
		return SimpleTimingSource{}, fmt.Errorf("unknown Linux clock source: %s", clockSource)
	}
}

// getMacOSTimingSource returns the optimal macOS timing source
func (pc *PlatformConfig) getMacOSTimingSource() (TimingSource, error) {
	if pc.MacOSConfig != nil && pc.MacOSConfig.TimingMode == "high_resolution" {
		return SimpleTimingSource{}, nil // macOS high-res timing
	}
	return SimpleTimingSource{}, nil
}

// getWindowsTimingSource returns the optimal Windows timing source
func (pc *PlatformConfig) getWindowsTimingSource() (TimingSource, error) {
	if pc.WindowsConfig != nil && pc.WindowsConfig.QueryPerformanceCounter {
		return SimpleTimingSource{}, nil // Windows performance counter
	}
	return SimpleTimingSource{}, nil
}

// TimingSource represents a source of timing information
type TimingSource interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

// SimpleTimingSource provides basic timing functionality
type SimpleTimingSource struct{}

func (s SimpleTimingSource) Now() time.Time {
	return time.Now()
}

func (s SimpleTimingSource) Since(t time.Time) time.Duration {
	return time.Since(t)
}

// GetRecommendedSettings returns platform-recommended settings for the current platform
func (pc *PlatformConfig) GetRecommendedSettings(platform *types.PlatformInfo) (*RecommendedSettings, error) {
	settings := &RecommendedSettings{
		Timeout:        5 * time.Second,
		BufferSize:     65535,
		TTL:           64,
		RetryAttempts: 3,
		RetryDelay:    time.Second,
		UseRawSockets: false,
		RequiresElevatedPrivs: false,
	}

	switch platform.OS {
	case "linux":
		settings.Timeout = 3 * time.Second
		settings.BufferSize = 32768
		settings.UseRawSockets = true
		settings.RequiresElevatedPrivs = true
		
	case "darwin":
		settings.Timeout = 5 * time.Second
		settings.BufferSize = 65535
		settings.RequiresElevatedPrivs = true
		
	case "windows":
		settings.Timeout = 6 * time.Second
		settings.BufferSize = 65535
		settings.RequiresElevatedPrivs = false // Basic ICMP doesn't require admin
	}

	return settings, nil
}

// RecommendedSettings contains platform-recommended configuration settings
type RecommendedSettings struct {
	Timeout              time.Duration `json:"timeout"`
	BufferSize          int           `json:"buffer_size"`
	TTL                 int           `json:"ttl"`
	RetryAttempts       int           `json:"retry_attempts"`
	RetryDelay          time.Duration `json:"retry_delay"`
	UseRawSockets       bool          `json:"use_raw_sockets"`
	RequiresElevatedPrivs bool        `json:"requires_elevated_privileges"`
}

// CheckCapability checks if the platform supports a specific capability
func (pc *PlatformConfig) CheckCapability(platform *types.PlatformInfo, capability string) (bool, error) {
	switch capability {
	case "icmp_raw":
		switch platform.OS {
		case "linux":
			return !pc.Security.ICMPPrivilegeRequired, nil // Available if not requiring root
		case "darwin":
			return true, nil // Usually available with admin
		case "windows":
			return false, nil // Windows raw sockets have restrictions
		}
	case "high_res_timing":
		switch platform.OS {
		case "linux":
			return runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64", nil
		case "darwin":
			return true, nil // macOS typically has high-res timing
		case "windows":
			return runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64", nil
		}
	case "packet_capture":
		return false, fmt.Errorf("packet capture capability not implemented")
	default:
		return false, fmt.Errorf("unknown capability: %s", capability)
	}
	return false, nil
}

// GetPrivilegesRequired returns the privilege requirements for the current platform
func (pc *PlatformConfig) GetPrivilegesRequired(platform *types.PlatformInfo) (*PrivilegeRequirements, error) {
	pr := &PrivilegeRequirements{
		RequiredPrivilegeLevel: pc.Security.MinimumPrivilegeLevel,
		SpecificCapabilities:   []string{},
		OperatingSystemSpecific: map[string]interface{}{},
	}

	switch platform.OS {
	case "linux":
		if pc.LinuxConfig != nil {
			if pc.LinuxConfig.RequireRoot {
				pr.RequiredPrivilegeLevel = "root"
			} else if pc.LinuxConfig.CapNetAdmin {
				pr.RequiredPrivilegeLevel = "elevated"
				pr.SpecificCapabilities = append(pr.SpecificCapabilities, "CAP_NET_ADMIN")
			}
			
			pr.OperatingSystemSpecific["raw_sockets"] = pc.LinuxConfig.UseRawSockets
			pr.OperatingSystemSpecific["allowed_users"] = pc.LinuxConfig.AllowedUsers
		}
		
	case "darwin":
		if pc.MacOSConfig != nil {
			if pc.MacOSConfig.RequireAdmin {
				pr.RequiredPrivilegeLevel = "elevated"
			}
			pr.OperatingSystemSpecific["use_icmp"] = pc.MacOSConfig.UseICMP
			pr.OperatingSystemSpecific["allowed_users"] = pc.MacOSConfig.AllowedUsers
		}
		
	case "windows":
		if pc.WindowsConfig != nil {
			if pc.WindowsConfig.RequireAdmin {
				pr.RequiredPrivilegeLevel = "elevated"
			}
			pr.OperatingSystemSpecific["use_wsa"] = pc.WindowsConfig.UseWSA
		}
	}

	return pr, nil
}

// PrivilegeRequirements describes the privilege requirements for a platform
type PrivilegeRequirements struct {
	RequiredPrivilegeLevel string                 `json:"required_privilege_level"`
	SpecificCapabilities   []string               `json:"specific_capabilities"`
	OperatingSystemSpecific map[string]interface{} `json:"os_specific"`
}

// DetectPlatform is a local helper function that returns platform detection
func DetectPlatform() (*types.PlatformInfo, error) {
	return platform.DetectPlatform()
}