package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mesh-net-probe/probe/internal/icmp"
	"github.com/mesh-net-probe/probe/internal/platform"
	"github.com/mesh-net-probe/probe/pkg/types"
)

var (
	// Global flags
	target     string
	count      int
	interval   time.Duration
	timeout    time.Duration
	verbose    bool
	jsonOutput bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "probe",
	Short: "Mesh Probe - High-precision network measurement tool",
	Long: `Mesh Probe performs ICMP measurements with microsecond precision
across multiple platforms (Linux, macOS, Windows, x86_64, ARM64).

Examples:
  probe ping 8.8.8.8
  probe ping --count 10 --interval 1s 192.168.1.1
  probe platform
  probe --version`,
	Version: "1.0.0",
}

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping <target>",
	Short: "Perform ICMP measurement to target",
	Long: `Perform high-precision ICMP measurements to a target.

Examples:
  probe ping 8.8.8.8
  probe ping --count 5 --timeout 3s google.com
  probe ping --json 192.168.1.1`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runPing,
}

// platformCmd represents the platform command
var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Display platform and capability information",
	Long: `Display detailed information about the current platform including
operating system, architecture, CPU capabilities, and measurement precision.`,
	Run: runPlatform,
}

// init initializes the application
func init() {
	// Global flags for all commands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")

	// Ping command flags
	pingCmd.Flags().StringVarP(&target, "target", "t", "", "Target IP address or hostname")
	pingCmd.Flags().IntVarP(&count, "count", "c", 1, "Number of measurements to perform")
	pingCmd.Flags().DurationVarP(&interval, "interval", "i", time.Second, "Interval between measurements")
	pingCmd.Flags().DurationVarP(&timeout, "timeout", "T", 5*time.Second, "Timeout for each measurement")

	// Add subcommands
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(platformCmd)
}

// runPing executes the ping command
func runPing(cmd *cobra.Command, args []string) error {
	// Determine target
	targetHost := target
	if len(args) > 0 {
		targetHost = args[0]
	}

	if targetHost == "" {
		return fmt.Errorf("target is required (provide as argument or use --target flag)")
	}

	// Resolve target address
	targetIP, err := resolveTarget(targetHost)
	if err != nil {
		return fmt.Errorf("failed to resolve target: %w", err)
	}

	// Initialize platform detection
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	// Validate platform capabilities
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		return fmt.Errorf("failed to validate capabilities: %w", err)
	}

	if !capabilities.CanMeasure {
		return fmt.Errorf("ICMP measurement not supported on this platform: %v", capabilities.Limitations)
	}

	// Initialize ICMP engine
	icmpEngine, err := icmp.NewEngine(
		icmp.WithTimeout(timeout),
		icmp.WithBufferSize(65535),
	)
	if err != nil {
		return fmt.Errorf("failed to create ICMP engine: %w", err)
	}
	defer icmpEngine.Close()

	ctx := context.Background()

	// Perform measurements
	if count == 1 {
		// Single measurement
		return performSingleMeasurement(ctx, icmpEngine, platformInfo, targetIP)
	} else {
		// Batch measurements
		return performBatchMeasurements(ctx, icmpEngine, platformInfo, targetIP, count, interval)
	}
}

// runPlatform executes the platform command
func runPlatform(cmd *cobra.Command, args []string) {
	// Detect platform
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to detect platform: %v\n", err)
		os.Exit(1)
	}

	// Validate capabilities
	capabilities, err := platform.ValidateCapabilities()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to validate capabilities: %v\n", err)
		os.Exit(1)
	}

	// Get platform stats
	stats, err := platform.GetPlatformStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get platform stats: %v\n", err)
		os.Exit(1)
	}

	// Check privilege requirements
	needsPrivs, missing, err := platform.CheckPrivilegeRequirements()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check privilege requirements: %v\n", err)
		os.Exit(1)
	}

	// Get optimal timing precision
	precision, err := platform.GetOptimalTimingPrecision()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to determine timing precision: %v\n", err)
		os.Exit(1)
	}

	// Output platform information
	if jsonOutput {
		outputJSONPlatformInfo(platformInfo, capabilities, stats, needsPrivs, missing, precision)
	} else {
		outputTextPlatformInfo(platformInfo, capabilities, stats, needsPrivs, missing, precision)
	}
}

// resolveTarget resolves a target hostname or IP address
func resolveTarget(target string) (net.IP, error) {
	// Check if it's already an IP address
	if ip := net.ParseIP(target); ip != nil {
		return ip, nil
	}

	// Resolve hostname
	addrs, err := net.LookupIP(target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target: %w", err)
	}

	// Prefer IPv4 addresses if available
	for _, addr := range addrs {
		if addr.To4() != nil {
			return addr, nil
		}
	}

	// Return first address if no IPv4 available
	return addrs[0], nil
}

// performSingleMeasurement performs a single ICMP measurement
func performSingleMeasurement(ctx context.Context, engine *icmp.Engine, platformInfo *types.PlatformInfo, target net.IP) error {
	measurement, err := engine.Ping(ctx, target)
	if err != nil {
		return fmt.Errorf("measurement failed: %w", err)
	}

	// Add platform information to measurement
	measurement.PlatformData.OS = platformInfo.OS
	measurement.PlatformData.Architecture = platformInfo.Arch

	// Output result
	if jsonOutput {
		return outputJSONMeasurement(measurement)
	}

	return outputTextMeasurement(measurement)
}

// performBatchMeasurements performs multiple ICMP measurements
func performBatchMeasurements(ctx context.Context, engine *icmp.Engine, platformInfo *types.PlatformInfo, target net.IP, count int, interval time.Duration) error {
	fmt.Printf("PING %s (%d measurements, interval %s):\n\n", target.String(), count, interval)

	successful := 0
	failed := 0
	var totalRTT time.Duration
	var minRTT time.Duration = -1
	var maxRTT time.Duration

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			measurement, err := engine.Ping(ctx, target)
			
			if err != nil {
				failed++
				fmt.Printf("Request %d: FAILED - %v\n", i+1, err)
			} else {
				successful++
				
				// Add platform information
				measurement.PlatformData.OS = platformInfo.OS
				measurement.PlatformData.Architecture = platformInfo.Arch
				
				// Update statistics
				totalRTT += measurement.RTT
				if minRTT == -1 || measurement.RTT < minRTT {
					minRTT = measurement.RTT
				}
				if measurement.RTT > maxRTT {
					maxRTT = measurement.RTT
				}
				
				if jsonOutput {
					outputJSONMeasurement(measurement)
				} else {
					fmt.Printf("Request %d: time=%s ttl=%d size=%d bytes\n", i+1, 
						measurement.RTT.Round(time.Microsecond), measurement.TTL, measurement.PacketSize)
				}
			}
		}
	}

	// Output summary
	if !jsonOutput && successful > 0 {
		avgRTT := totalRTT / time.Duration(successful)
		lossRate := float64(failed) / float64(count) * 100
		
		fmt.Printf("\n--- %s ping statistics ---\n", target.String())
		fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n", count, successful, lossRate)
		fmt.Printf("round-trip min/avg/max = %s/%s/%s\n", 
			minRTT.Round(time.Microsecond), 
			avgRTT.Round(time.Microsecond), 
			maxRTT.Round(time.Microsecond))
	}

	return nil
}

// Output functions
func outputJSONPlatformInfo(platformInfo *types.PlatformInfo, capabilities *types.PlatformCapabilities, stats *types.PlatformStats, needsPrivs bool, missing []string, precision types.MeasurementPrecision) error {
	fmt.Printf(`{
  "platform": {
    "os": "%s",
    "arch": "%s", 
    "version": "%s",
    "hostname": "%s",
    "container": %v
  },
  "capabilities": {
    "can_measure": %v,
    "max_precision": "%s",
    "features": [%s],
    "limitations": [%s]
  },
  "stats": {
    "cpu_count": %d,
    "memory_mb": %d,
    "uptime": "%s"
  },
  "privileges": {
    "required": %v,
    "missing": [%s]
  }
}`, 
		platformInfo.OS, platformInfo.Arch, platformInfo.Version, platformInfo.Hostname, platformInfo.Container,
		capabilities.CanMeasure, precision, formatStringArray(capabilities.Features), formatStringArray(capabilities.Limitations),
		stats.CPUCount, stats.MemorySys/(1024*1024), stats.Uptime,
		needsPrivs, formatStringArray(missing))
	return nil
}

func outputTextPlatformInfo(platformInfo *types.PlatformInfo, capabilities *types.PlatformCapabilities, stats *types.PlatformStats, needsPrivs bool, missing []string, precision types.MeasurementPrecision) {
	fmt.Println("=== Platform Information ===")
	fmt.Printf("Operating System: %s %s\n", platformInfo.OS, platformInfo.Version)
	fmt.Printf("Architecture: %s\n", platformInfo.Arch)
	fmt.Printf("Hostname: %s\n", platformInfo.Hostname)
	fmt.Printf("Container: %v\n", platformInfo.Container)
	fmt.Printf("CPU Count: %d\n", stats.CPUCount)
	fmt.Printf("Memory: %d MB\n", stats.MemorySys/(1024*1024))
	fmt.Printf("Uptime: %s\n", stats.Uptime)
	
	fmt.Println("\n=== Capabilities ===")
	fmt.Printf("ICMP Measurement: %v\n", capabilities.CanMeasure)
	fmt.Printf("Timing Precision: %s\n", precision)
	
	if len(capabilities.Features) > 0 {
		fmt.Println("Features:")
		for _, feature := range capabilities.Features {
			fmt.Printf("  • %s\n", feature)
		}
	}
	
	if len(capabilities.Limitations) > 0 {
		fmt.Println("\nLimitations:")
		for _, limitation := range capabilities.Limitations {
			fmt.Printf("  • %s\n", limitation)
		}
	}
	
	if needsPrivs {
		fmt.Println("\n=== Privileges Required ===")
		if len(missing) > 0 {
			fmt.Println("Missing capabilities:")
			for _, capability := range missing {
				fmt.Printf("  • %s\n", capability)
			}
		} else {
			fmt.Println("No special privileges required")
		}
	}
}

func outputJSONMeasurement(measurement *types.MeasurementData) error {
	fmt.Printf(`{
  "success": %v,
  "target": "%s",
  "rtt": "%s",
  "ttl": %d,
  "packet_size": %d,
  "timestamp": "%s",
  "source_ip": "%s",
  "dest_ip": "%s"
}`, 
		measurement.Success, measurement.Target.Address.String(), 
		measurement.RTT.Round(time.Microsecond), measurement.TTL, 
		measurement.PacketSize, measurement.Timestamp.Format(time.RFC3339Nano),
		measurement.SourceIP.String(), measurement.DestIP.String())
	return nil
}

func outputTextMeasurement(measurement *types.MeasurementData) error {
	fmt.Printf("PING %s:\n", measurement.Target.Address.String())
	if measurement.Success {
		fmt.Printf("time=%s ttl=%d size=%d bytes\n", measurement.RTT.Round(time.Microsecond), measurement.TTL, measurement.PacketSize)
	} else {
		fmt.Printf("REQUEST FAILED: %s\n", measurement.ErrorMessage)
	}
	return nil
}

func formatStringArray(items []string) string {
	if len(items) == 0 {
		return ""
	}
	
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf(`"%s"`, item)
	}
	return result
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}