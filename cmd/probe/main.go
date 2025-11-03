package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
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
	configFile        string
	logLevel          string
	logFormat         string
	otlpEndpoint      string
	probeID           string
	verbose           bool
	daemon            bool
	measureCount      int
	continuousMode    bool
	continuousInterval time.Duration
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
	app.config = loadConfiguration()
	
	// Generate probe ID if not set
	app.probeID = generateProbeID()
	
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
	fmt.Printf("Starting interactive mode with configured targets\n")
	
	targets := toPointerSlice(app.config.Targets)
	if len(targets) == 0 {
		return fmt.Errorf("no targets configured in config file")
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
	// Use JSON format if specified
	if logFormat == "json" {
		return app.outputJSONResults(measurements)
	} else {
		return app.outputTextResults(measurements)
	}
}

func (app *ProbeApplication) outputTextResults(measurements []*types.MeasurementData) error {
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

func (app *ProbeApplication) outputJSONResults(measurements []*types.MeasurementData) error {
	// Create a structured JSON response
	type MeasurementResult struct {
		Success bool   `json:"success"`
		Target  string `json:"target"`
		Address string `json:"address"`
		RTT     string `json:"rtt,omitempty"`
		Error   string `json:"error,omitempty"`
	}
	
	type OutputResponse struct {
		Results       []MeasurementResult `json:"results"`
		TotalTargets  int                 `json:"total_targets"`
		SuccessCount  int                 `json:"success_count"`
		FailureCount  int                 `json:"failure_count"`
	}
	
	var response OutputResponse
	response.Results = make([]MeasurementResult, 0, len(measurements))
	
	for _, measurement := range measurements {
		result := MeasurementResult{
			Success: measurement.Success,
			Target:  measurement.Target.ID,
			Address: measurement.DestIP.String(),
		}
		
		if measurement.Success {
			result.RTT = measurement.RTT.String()
			response.SuccessCount++
		} else {
			result.Error = measurement.ErrorMessage
			response.FailureCount++
		}
		
		response.Results = append(response.Results, result)
	}
	
	response.TotalTargets = len(measurements)
	
	// Marshal and output JSON
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON output: %w", err)
	}
	
	fmt.Printf("%s\n", string(jsonData))
	return nil
}

func (app *ProbeApplication) outputAveragedResults(allMeasurements [][]*types.MeasurementData, targets []*types.NetworkTarget) error {
	// Calculate statistics for each target
	stats := make([]*TargetStats, len(targets))
	
	for i := range targets {
		stats[i] = calculateTargetStats(allMeasurements, i)
	}
	
	// Output results based on format
	if logFormat == "json" {
		return app.outputAveragedJSONResults(stats, targets)
	} else {
		return app.outputAveragedTextResults(stats, targets)
	}
}

func (app *ProbeApplication) outputAveragedTextResults(stats []*TargetStats, targets []*types.NetworkTarget) error {
	fmt.Printf("\n=== Averaged Results (%d measurements) ===\n", measureCount)
	
	for i, target := range targets {
		s := stats[i]
		fmt.Printf("%s -> %s:\n", target.ID, target.Address.String())
		fmt.Printf("  Average: %v\n", s.Average)
		fmt.Printf("  Min: %v\n", s.Min)
		fmt.Printf("  Max: %v\n", s.Max)
		fmt.Printf("  StdDev: %v\n", s.StandardDeviation)
		fmt.Printf("  Success Rate: %.1f%% (%d/%d)\n", s.SuccessRate*100, s.SuccessCount, s.TotalCount)
	}
	
	return nil
}

func (app *ProbeApplication) outputAveragedJSONResults(stats []*TargetStats, targets []*types.NetworkTarget) error {
	type AveragedMeasurementResult struct {
		Target              string  `json:"target"`
		Address             string  `json:"address"`
		AverageRTT          string  `json:"average_rtt"`
		MinRTT              string  `json:"min_rtt"`
		MaxRTT              string  `json:"max_rtt"`
		StandardDeviation   string  `json:"std_dev"`
		SuccessRate         float64 `json:"success_rate"`
		TotalMeasurements   int     `json:"total_measurements"`
		SuccessfulCount     int     `json:"successful_count"`
		FailedCount         int     `json:"failed_count"`
	}
	
	type OutputResponse struct {
		MeasurementCount int                     `json:"measurement_count"`
		Results          []AveragedMeasurementResult `json:"results"`
	}
	
	var response OutputResponse
	response.MeasurementCount = measureCount
	response.Results = make([]AveragedMeasurementResult, len(targets))
	
	for i, target := range targets {
		s := stats[i]
		response.Results[i] = AveragedMeasurementResult{
			Target:              target.ID,
			Address:             target.Address.String(),
			AverageRTT:          s.Average.String(),
			MinRTT:              s.Min.String(),
			MaxRTT:              s.Max.String(),
			StandardDeviation:   s.StandardDeviation.String(),
			SuccessRate:         s.SuccessRate,
			TotalMeasurements:   s.TotalCount,
			SuccessfulCount:     s.SuccessCount,
			FailedCount:         s.FailedCount,
		}
	}
	
	// Marshal and output JSON
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON output: %w", err)
	}
	
	fmt.Printf("%s\n", string(jsonData))
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

func loadConfiguration() *types.Configuration {
	// Start with default configuration
	config := &types.Configuration{
		Name:      "Default Configuration",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		
		Network: &types.NetworkConfig{
			BufferSize: 2048,
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
	
	// Try to load from config file if specified
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Printf("Warning: Could not load config file %s: %v\n", configFile, err)
			return applyLogLevelOverride(config)
		}
		
		// Parse the JSON config file
		var fileConfig map[string]interface{}
		if err := json.Unmarshal(data, &fileConfig); err != nil {
			fmt.Printf("Warning: Could not parse config file %s: %v\n", configFile, err)
			return applyLogLevelOverride(config)
		}
		
		// Extract targets from config
		if targetsSlice, ok := fileConfig["targets"].([]interface{}); ok {
			config.Targets = make([]types.NetworkTarget, 0, len(targetsSlice))
			
			for i, targetData := range targetsSlice {
				targetMap, ok := targetData.(map[string]interface{})
				if !ok {
					fmt.Printf("Warning: Invalid target data at index %d\n", i)
					continue
				}
				
				target := types.NetworkTarget{
					ID:      fmt.Sprintf("target_%d", i+1),
					Enabled: true,
					Timeout: 5 * time.Second,
				}
				
				// Parse address
				if addrStr, ok := targetMap["address"].(string); ok {
					target.Address = parseIP(addrStr)
				}
				
				// Parse ID
				if idStr, ok := targetMap["id"].(string); ok {
					target.ID = idStr
				}
				
				// Parse enabled flag
				if enabled, ok := targetMap["enabled"].(bool); ok {
					target.Enabled = enabled
				}
				
				// Parse timeout
				if timeoutSec, ok := targetMap["timeout"].(float64); ok {
					target.Timeout = time.Duration(timeoutSec) * time.Second
				}
				
				config.Targets = append(config.Targets, target)
			}
			
			fmt.Printf("Loaded %d targets from configuration\n", len(config.Targets))
		}
		
		// Extract network config
		if networkMap, ok := fileConfig["network"].(map[string]interface{}); ok {
			if bufferSize, ok := networkMap["buffer_size"].(float64); ok {
				config.Network.BufferSize = int(bufferSize)
			}
		}
		
		fmt.Printf("Loaded configuration from %s\n", configFile)
	}
	
	return applyLogLevelOverride(config)
}

func applyLogLevelOverride(config *types.Configuration) *types.Configuration {
	// Apply command line log level override
	if logLevel != "" && config.Telemetry != nil {
		config.Telemetry.LogLevel = logLevel
		if verbose {
			fmt.Printf("Log level set to: %s (verbose mode)\n", logLevel)
		} else {
			fmt.Printf("Log level set to: %s\n", logLevel)
		}
	}
	return config
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
	Long:  "Perform single ICMP measurements to the specified target IP addresses or hostnames. If no targets are specified, uses targets from the configuration file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create app with proper configuration loading
		app := NewProbeApplication()
		
		// Load configuration first
		app.config = loadConfiguration()
		
		// Check if we have CLI targets or should use config targets
		if len(args) == 0 {
			// No CLI targets provided, check if config has targets
			if len(app.config.Targets) == 0 {
				return fmt.Errorf("no targets provided in CLI and no targets configured in config file")
			}
			fmt.Printf("Using %d targets from configuration\n", len(app.config.Targets))
		} else {
			fmt.Printf("Using %d targets from command line\n", len(args))
		}
		
		// Initialize platform
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
		
		// Configure ICMP engine with network settings from config
		if app.config.Network != nil {
			networkConfig := &types.NetworkConfig{
				BufferSize: app.config.Network.BufferSize,
				TTL:        app.config.Network.TTL,
			}
			
			if err := app.engine.Initialize(app.ctx, networkConfig); err != nil {
				return fmt.Errorf("failed to initialize ICMP engine: %w", err)
			}
		}
		
		// Generate probe ID
		app.probeID = generateProbeID()
		
		// Perform measurements using CLI targets or fall back to config targets
		if len(args) > 0 {
			return app.performMeasurements(args)
		} else {
			return app.performMeasurementsFromConfig()
		}
	},
}

func (app *ProbeApplication) performMeasurements(targets []string) error {
	// Convert targets to NetworkTarget format
	networkTargets := make([]*types.NetworkTarget, len(targets))
	for i, target := range targets {
		parsedIP := parseIP(target)
		networkTargets[i] = &types.NetworkTarget{
			ID:      fmt.Sprintf("cmd_target_%d", i+1),
			Address: parsedIP,
			Enabled: true,
			Timeout: 5 * time.Second,
		}
	}
	
	return app.performMeasurementLoop(networkTargets)
}

func (app *ProbeApplication) performMeasurementLoop(targets []*types.NetworkTarget) error {
	if continuousMode {
		return app.performContinuousMeasurements(targets)
	} else if measureCount > 1 {
		return app.performAveragedMeasurements(targets)
	} else {
		// Single measurement
		measurements, err := app.engine.MeasureBatch(app.ctx, targets)
		if err != nil {
			return fmt.Errorf("measurements failed: %w", err)
		}
		return app.outputResults(measurements)
	}
}

func (app *ProbeApplication) performAveragedMeasurements(targets []*types.NetworkTarget) error {
	fmt.Printf("Performing %d measurements per target for averaging\n", measureCount)
	
	// For averaging, we need to run multiple rounds of measurements
	// and calculate statistics
	allMeasurements := make([][]*types.MeasurementData, measureCount)
	
	for round := 0; round < measureCount; round++ {
		if verbose {
			fmt.Printf("Measurement round %d/%d\n", round+1, measureCount)
		}
		
		measurements, err := app.engine.MeasureBatch(app.ctx, targets)
		if err != nil {
			return fmt.Errorf("measurement round %d failed: %w", round+1, err)
		}
		
		allMeasurements[round] = measurements
		
		// Small delay between measurements to avoid rate limiting
		if round < measureCount-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	// Calculate statistics and output results
	return app.outputAveragedResults(allMeasurements, targets)
}

func (app *ProbeApplication) performContinuousMeasurements(targets []*types.NetworkTarget) error {
	fmt.Printf("Starting continuous measurements with %v interval\n", continuousInterval)
	
	measurementRound := 1
	
	for {
		select {
		case <-app.ctx.Done():
			fmt.Printf("Continuous measurement stopped after %d rounds\n", measurementRound-1)
			return nil
		default:
		}
		
		fmt.Printf("Continuous measurement round %d\n", measurementRound)
		
		measurements, err := app.engine.MeasureBatch(app.ctx, targets)
		if err != nil {
			fmt.Printf("Continuous measurement round %d failed: %v\n", measurementRound, err)
		} else {
			app.outputResults(measurements)
		}
		
		measurementRound++
		
		// Wait for next interval
		timer := time.NewTimer(continuousInterval)
		select {
		case <-app.ctx.Done():
			timer.Stop()
			fmt.Printf("Continuous measurement stopped after %d rounds\n", measurementRound-1)
			return nil
		case <-timer.C:
			// Continue to next measurement
		}
	}
}

func (app *ProbeApplication) performMeasurementsFromConfig() error {
	// Use targets from configuration
	targets := toPointerSlice(app.config.Targets)
	if len(targets) == 0 {
		return fmt.Errorf("no targets configured in config file")
	}
	
	return app.performMeasurementLoop(targets)
}

func parseIP(target string) net.IP {
	// Try to parse as IP address first
	if ip := net.ParseIP(target); ip != nil {
		return ip
	}
	
	// Try to resolve as hostname
	if host, err := net.LookupHost(target); err == nil && len(host) > 0 {
		return net.ParseIP(host[0])
	}
	
	// Fallback to loopback if parsing fails
	fmt.Printf("Warning: Could not resolve '%s', using 127.0.0.1\n", target)
	return net.ParseIP("127.0.0.1")
}

func toPointerSlice(targets []types.NetworkTarget) []*types.NetworkTarget {
	ptrSlice := make([]*types.NetworkTarget, len(targets))
	for i := range targets {
		ptrSlice[i] = &targets[i]
	}
	return ptrSlice
}

// TargetStats contains statistical analysis of multiple measurements
type TargetStats struct {
	Average          time.Duration
	Min              time.Duration
	Max              time.Duration
	StandardDeviation time.Duration
	SuccessRate      float64
	TotalCount       int
	SuccessCount     int
	FailedCount      int
}

// calculateTargetStats calculates statistics for a specific target across all measurement rounds
func calculateTargetStats(allMeasurements [][]*types.MeasurementData, targetIndex int) *TargetStats {
	var totalRTT time.Duration
	var successCount int
	var failedCount int
	var minRTT time.Duration
	var maxRTT time.Duration
	var firstSuccess bool
	
	var rtts []time.Duration
	
	// Collect all RTTs and count successes/failures
	for _, roundMeasurements := range allMeasurements {
		if targetIndex < len(roundMeasurements) {
			measurement := roundMeasurements[targetIndex]
			if measurement.Success {
				successCount++
				totalRTT += measurement.RTT
				rtts = append(rtts, measurement.RTT)
				
				if !firstSuccess || measurement.RTT < minRTT {
					minRTT = measurement.RTT
				}
				if !firstSuccess || measurement.RTT > maxRTT {
					maxRTT = measurement.RTT
				}
				firstSuccess = true
			} else {
				failedCount++
			}
		}
	}
	
	totalCount := successCount + failedCount
	successRate := float64(successCount) / float64(totalCount)
	
	var average time.Duration
	var stdDev time.Duration
	
	if successCount > 0 {
		average = totalRTT / time.Duration(successCount)
		
		// Calculate standard deviation
		if len(rtts) > 1 {
			var sumSqDiff time.Duration
			for _, rtt := range rtts {
				diff := rtt - average
				sumSqDiff += diff * diff
			}
			stdDev = time.Duration(math.Sqrt(float64(sumSqDiff) / float64(len(rtts)-1)))
		}
	}
	
	return &TargetStats{
		Average:            average,
		Min:                minRTT,
		Max:                maxRTT,
		StandardDeviation:  stdDev,
		SuccessRate:        successRate,
		TotalCount:         totalCount,
		SuccessCount:       successCount,
		FailedCount:        failedCount,
	}
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
	
	// Measurement control flags
	rootCmd.PersistentFlags().IntVarP(&measureCount, "count", "n", 1, "Number of measurements to average (for averaging mode)")
	rootCmd.PersistentFlags().BoolVar(&continuousMode, "continuous", false, "Enable continuous measurement mode")
	rootCmd.PersistentFlags().DurationVar(&continuousInterval, "interval", 1*time.Second, "Interval for continuous measurements (e.g., 30s, 1m, 5m)")
	
	// Add subcommands
	rootCmd.AddCommand(systemCmd)
	rootCmd.AddCommand(measureCmd)
	
	// Set up error handling
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}