package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mesh-net-probe/probe/internal/icmp"
	"github.com/mesh-net-probe/probe/internal/logger"
	"github.com/mesh-net-probe/probe/internal/platform"
	"github.com/mesh-net-probe/probe/pkg/types"
)

// Global variables for the application
var (
	configFile   string
	logLevel     string
	logFormat    string
	otlpEndpoint string
	probeID      string
	verbose      bool
	daemon       bool
)

// ProbeApplication represents the main application instance
type ProbeApplication struct {
	// Core components
	engine    icmp.Engine
	timing    icmp.TimingEngine
	logger    logger.Logger
	
	// Configuration
	config    *types.Configuration
	
	// Runtime state
	ctx       context.Context
	cancel    context.CancelFunc
	probeID   string
	
	// Platform info
	platform  *types.PlatformInfo
}

// NewProbeApplication creates a new probe application instance
func NewProbeApplication() *ProbeApplication {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ProbeApplication{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize sets up all application components
func (app *ProbeApplication) Initialize() error {
	// Initialize logger first
	app.logger = logger.GetGlobalLogger()
	
	// Initialize platform detection
	var err error
	app.platform, err = platform.DetectPlatform()
	if err != nil {
		return fmt.Errorf("platform detection failed: %w", err)
	}
	
	// Validate platform compatibility
	if err := platform.ValidatePlatformCompatibility(app.platform); err != nil {
		return fmt.Errorf("platform validation failed: %w", err)
	}
	
	// Initialize timing engine
	app.timing = icmp.NewTimingEngine()
	if err := app.timing.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize timing: %w", err)
	}
	
	// Initialize ICMP engine
	app.engine = icmp.NewEngine()
	
	// Configure ICMP engine
	networkConfig := &types.NetworkConfig{
		BufferSize: 8192,
		TTL:        64,
	}
	
	if err := app.engine.Initialize(app.ctx, networkConfig); err != nil {
		return fmt.Errorf("failed to initialize ICMP engine: %w", err)
	}
	
	// Load configuration
	app.config = createDefaultConfiguration()
	
	// Generate probe ID if not set
	if probeID == "" {
		app.probeID = generateProbeID()
	} else {
		app.probeID = probeID
	}
	
	return nil
}

// Run starts the probe application
func (app *ProbeApplication) Run() error {
	// For daemon mode, start the background service
	if daemon {
		return app.runDaemon()
	}
	
	// For interactive mode, perform single measurements based on config
	return app.runInteractive()
}

// Shutdown gracefully shuts down the application
func (app *ProbeApplication) Shutdown() error {
	if app.cancel != nil {
		app.cancel()
	}
	
	// Close ICMP engine
	if app.engine != nil {
		return app.engine.Close(app.ctx)
	}
	
	return nil
}

// Runtime methods

func (app *ProbeApplication) runDaemon() error {
	fmt.Printf("Starting probe daemon mode\n")
	
	targets := toPointerSlice(app.config.Targets)
	if len(targets) == 0 {
		fmt.Printf("Warning: No targets configured for daemon mode\n")
		return nil
	}
	
	// Start continuous measurements
	if err := app.engine.StartContinuous(app.ctx, targets, 1*time.Minute); err != nil {
		return fmt.Errorf("failed to start continuous measurements: %w", err)
	}
	
	// Run until context is cancelled
	<-app.ctx.Done()
	
	return app.engine.StopContinuous(app.ctx)
}

func (app *ProbeApplication) runInteractive() error {
	fmt.Printf("Starting interactive mode\n")
	
	targets := toPointerSlice(app.config.Targets)
	if len(targets) == 0 {
		return fmt.Errorf("no targets configured")
	}
	
	// Perform measurements for each target
	measurements, err := app.engine.MeasureBatch(app.ctx, targets)
	if err != nil {
		return fmt.Errorf("measurement batch failed: %w", err)
	}
	
	// Output results
	return app.outputResults(measurements)
}

func (app *ProbeApplication) outputResults(measurements []*types.MeasurementData) error {
	for _, measurement := range measurements {
		if measurement.Success {
			fmt.Printf("SUCCESS: %s -> %s: %v\n",
				measurement.Target.ID,
				measurement.DestIP.String(),
				measurement.RTT,
			)
		} else {
			fmt.Printf("FAILED: %s -> %s: %s\n",
				measurement.Target.ID,
				measurement.DestIP.String(),
				measurement.ErrorMessage,
			)
		}
	}
	return nil
}

// ShowSystemInfo displays detailed system information
func ShowSystemInfo() error {
	app := NewProbeApplication()
	if err := app.Initialize(); err != nil {
		return err
	}
	defer app.Shutdown()
	
	// Get platform information
	platformInfo, err := platform.DetectPlatform()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}
	
	// Get timing system information
	timingSystem := app.timing.GetSystemInfo()
	timingStats := app.timing.GetStats()
	
	fmt.Printf("=== Mesh Probe System Information ===\n")
	fmt.Printf("Probe ID: %s\n", app.probeID)
	fmt.Printf("Platform: %s %s (%s)\n", platformInfo.OS, platformInfo.Version, platformInfo.Arch)
	fmt.Printf("Hostname: %s\n", platformInfo.Hostname)
	fmt.Printf("Container: %t\n", platformInfo.Container)
	fmt.Printf("\n=== Timing System ===\n")
	fmt.Printf("Timer Resolution: %v\n", timingSystem.TimerResolution)
	fmt.Printf("Achieved Precision: %s\n", timingStats.PrecisionAchieved)
	fmt.Printf("CPU Logical Cores: %d\n", timingSystem.CPULogicalCores)
	fmt.Printf("Go Version: %s\n", timingSystem.GoVersion)
	fmt.Printf("High-Res Timer: %t\n", timingSystem.HasHighResTimer)
	fmt.Printf("Monotonic Clock: %t\n", timingSystem.HasMonotonicClock)
	
	return nil
}

// Utility functions

func generateProbeID() string {
	// Generate unique probe ID based on hostname, process ID, and random data
	hostname, _ := os.Hostname()
	pid := os.Getpid()
	
	// Create seed from hostname and PID
	seed := fmt.Sprintf("%s-%d-%d", hostname, pid, time.Now().UnixNano())
	
	// Hash the seed to create a consistent ID
	hash := sha256.Sum256([]byte(seed))
	
	// Convert to short ID (first 16 characters)
	return "probe_" + fmt.Sprintf("%x", hash[:8])
}

func createDefaultConfiguration() *types.Configuration {
	return &types.Configuration{
		Name:      "Default Configuration",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		
		Network: &types.NetworkConfig{
			BufferSize: 8192,
			TTL:        64,
		},
		
		Telemetry: &types.TelemetryConfig{
			EnableMetrics: true,
			EnableTracing: true,
			EnableLogging: true,
			LogLevel:     "info",
			LogFormat:    "text",
		},
	}
}

// Command definitions

var rootCmd = &cobra.Command{
	Use:   "mesh-probe",
	Short: "High-precision ICMP measurement probe",
	Long:  "Mesh Probe provides microsecond-precision ICMP measurements with distributed coordination capabilities.",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := NewProbeApplication()
		if err := app.Initialize(); err != nil {
			return err
		}
		defer app.Shutdown()
		
		return app.Run()
	},
}

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Show system information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ShowSystemInfo()
	},
}

var measureCmd = &cobra.Command{
	Use:   "measure [target...]",
	Short: "Perform ICMP measurements to specified targets",
	Long:  "Perform single ICMP measurements to the specified target IP addresses or hostnames.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("at least one target is required")
		}
		
		app := NewProbeApplication()
		if err := app.Initialize(); err != nil {
			return err
		}
		defer app.Shutdown()
		
		return app.performMeasurements(args)
	},
}

func (app *ProbeApplication) performMeasurements(targets []string) error {
	// Convert targets to NetworkTarget format
	networkTargets := make([]*types.NetworkTarget, len(targets))
	for i, target := range targets {
		networkTargets[i] = &types.NetworkTarget{
			ID:      fmt.Sprintf("target_%d", i+1),
			Address: parseIP(target),
			Enabled: true,
			Timeout: 5 * time.Second,
		}
	}
	
	measurements, err := app.engine.MeasureBatch(app.ctx, networkTargets)
	if err != nil {
		return fmt.Errorf("measurements failed: %w", err)
	}
	
	return app.outputResults(measurements)
}

func parseIP(target string) net.IP {
	// Parse IP address or hostname
	if ip := net.ParseIP(target); ip != nil {
		return ip
	}
	// Fallback to loopback if parsing fails
	return net.ParseIP("127.0.0.1")
}

func toPointerSlice(targets []types.NetworkTarget) []*types.NetworkTarget {
	ptrSlice := make([]*types.NetworkTarget, len(targets))
	for i := range targets {
		ptrSlice[i] = &targets[i]
	}
	return ptrSlice
}

// Main entry point

func main() {
	// Parse command line flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVarP(&logFormat, "log-format", "f", "text", "Log format (text, json)")
	rootCmd.PersistentFlags().StringVarP(&otlpEndpoint, "otlp-endpoint", "o", "", "OpenTelemetry Collector endpoint")
	rootCmd.PersistentFlags().StringVarP(&probeID, "probe-id", "p", "", "Probe identifier")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&daemon, "daemon", "d", false, "Run as daemon")
	
	// Add subcommands
	rootCmd.AddCommand(systemCmd)
	rootCmd.AddCommand(measureCmd)
	
	// Set up error handling
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}