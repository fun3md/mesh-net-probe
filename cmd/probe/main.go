package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/mesh-net-probe/probe/internal/agentclient"
	"github.com/mesh-net-probe/probe/internal/config"
	"github.com/mesh-net-probe/probe/internal/icmp"
	"github.com/mesh-net-probe/probe/internal/logger"
	"github.com/mesh-net-probe/probe/internal/platform"
	"github.com/mesh-net-probe/probe/internal/telemetry"
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
	maxHops    int
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

// tracerouteCmd represents the traceroute command
var tracerouteCmd = &cobra.Command{
	Use:   "traceroute <target>",
	Short: "Perform traceroute to target",
	Long: `Perform a traceroute to the target host to determine the path taken.
This command will display each hop in the route to the destination.

Examples:
  probe traceroute 8.8.8.8
  probe traceroute --max-hops 30 google.com`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runTraceroute,
}

// agentCmd represents the long-running probe daemon/agent mode
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run probe in long-running daemon/agent mode integrated with central backend",
	Long: `Run Mesh Probe as a managed agent:
- Registers with backend ProbeRegistry
- Sends periodic heartbeats
- Pulls configuration from central config manager/control plane
- Applies configuration without restart
- Reports config-applied and emits telemetry for observability`,
	RunE: runAgent,
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

	// Traceroute command flags
	tracerouteCmd.Flags().IntVarP(&maxHops, "max-hops", "m", 30, "Maximum number of hops")
	tracerouteCmd.Flags().DurationVarP(&timeout, "timeout", "T", 5*time.Second, "Timeout for each probe")

	// Add subcommands
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(platformCmd)
	rootCmd.AddCommand(tracerouteCmd)

	// Agent/daemon mode flags
	agentCmd.Flags().String("backend-url", "", "Admin backend base URL (e.g. http://admin-web:8080/api/v1)")
	agentCmd.Flags().String("backend-api-key", "", "API key or token for backend authentication (overrides env)")
	agentCmd.Flags().Duration("heartbeat-interval", 10*time.Second, "Interval between heartbeats")
	agentCmd.Flags().Duration("config-interval", 30*time.Second, "Interval between configuration syncs")
	agentCmd.Flags().Duration("shutdown-timeout", 30*time.Second, "Graceful shutdown timeout")
	rootCmd.AddCommand(agentCmd)
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
		icmp.WithVerbose(verbose),
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

func runTraceroute(cmd *cobra.Command, args []string) error {
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

	// Initialize ICMP engine
	icmpEngine, err := icmp.NewEngine(
		icmp.WithTimeout(timeout),
		icmp.WithBufferSize(65535),
		icmp.WithVerbose(verbose),
	)
	if err != nil {
		return fmt.Errorf("failed to create ICMP engine: %w", err)
	}
	defer icmpEngine.Close()

	ctx := context.Background()

	// Perform traceroute
	if jsonOutput {
		return performTracerouteJSON(ctx, icmpEngine, targetIP)
	} else {
		return performTracerouteText(ctx, icmpEngine, targetIP)
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

	// Output platform information (simple, non-broken version)
	if jsonOutput {
		payload := map[string]interface{}{
			"platform":             platformInfo,
			"capabilities":         capabilities,
			"stats":                stats,
			"needs_privileges":     needsPrivs,
			"missing_capabilities": missing,
			"timing_precision":     precision,
		}
		data, jerr := json.MarshalIndent(payload, "", "  ")
		if jerr != nil {
			fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", jerr)
			os.Exit(1)
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("Platform: %+v\n", platformInfo)
		fmt.Printf("Capabilities: %+v\n", capabilities)
		fmt.Printf("Stats: %+v\n", stats)
		fmt.Printf("Needs elevated privileges: %v (missing: %v)\n", needsPrivs, missing)
		fmt.Printf("Optimal timing precision: %s\n", precision)
	}
}

// runAgent starts the long-running daemon/agent loop.
func runAgent(cmd *cobra.Command, args []string) error {
	// Context with signal-based cancellation for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	backendURL, _ := cmd.Flags().GetString("backend-url")
	apiKeyFlag, _ := cmd.Flags().GetString("backend-api-key")
	heartbeatInterval, _ := cmd.Flags().GetDuration("heartbeat-interval")
	configInterval, _ := cmd.Flags().GetDuration("config-interval")
	shutdownTimeout, _ := cmd.Flags().GetDuration("shutdown-timeout")

	// Initialize configuration manager.
	// For Phase 5.3 we use config.Manager as the preferred source of effective configuration,
	// but agent mode MUST be able to start even when no external providers are configured.
	baseCfg := &types.Configuration{}
	cfgManager := config.NewManager()
	if err := cfgManager.Initialize(ctx, baseCfg); err != nil {
		// If there are no healthy providers, log and continue with in-memory/default config
		fmt.Fprintf(os.Stderr, "Warning: configuration manager initialization degraded: %v\n", err)
	} else {
		defer cfgManager.Close(context.Background())
	}

	// Initialize logging manager
	logMgr, err := logger.NewManager(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}
	defer logMgr.Close()
	log := logMgr.WithComponent("probe-agent")
	log.Info(ctx, "startup", "starting probe agent daemon")

	// Initialize telemetry based on configuration (if present)
	var telemMgr *telemetry.Manager
	if baseCfg.Telemetry != nil {
		tm, terr := telemetry.NewManager(baseCfg.Telemetry)
		if terr != nil {
			log.Error(ctx, terr, "startup", "failed to initialize telemetry manager from config")
			return terr
		}
		telemMgr = tm
	} else {
		tm, terr := telemetry.NewManager(nil)
		if terr != nil {
			log.Error(ctx, terr, "startup", "failed to initialize telemetry manager with defaults")
			return terr
		}
		telemMgr = tm
	}
	if err := telemMgr.Start(ctx); err != nil {
		log.Error(ctx, err, "startup", "failed to start telemetry")
		return err
	}
	defer telemMgr.Stop(context.Background())

	// Derive backend settings from flags or environment.
	if backendURL == "" {
		if v := os.Getenv("PROBE_BACKEND_URL"); v != "" {
			backendURL = v
		}
	}
	if backendURL == "" {
		log.Error(ctx, fmt.Errorf("missing backend URL"), "startup", "PROBE_BACKEND_URL or --backend-url must be set for agent mode")
		return fmt.Errorf("backend URL required for agent mode")
	}

	apiKey := apiKeyFlag
	if apiKey == "" && baseCfg.Auth != nil {
		apiKey = baseCfg.Auth.APIKey
	}
	if apiKey == "" {
		apiKey = os.Getenv("PROBE_BACKEND_API_KEY")
	}

	// Derive a stable probe ID (trusted single-tenant model).
	probeID := os.Getenv("PROBE_ID")
	if probeID == "" {
		if h, err := os.Hostname(); err == nil && h != "" {
			probeID = h
		} else {
			probeID = fmt.Sprintf("probe-%d", time.Now().UnixNano())
		}
	}

	// Initialize agent backend client
	agentClient, err := agentclient.NewClient(agentclient.Config{
		BaseURL:       backendURL,
		APIKey:        apiKey,
		Timeout:       5 * time.Second,
		ProbeID:       probeID,
		LoggerManager: logMgr,
	})
	if err != nil {
		log.Error(ctx, err, "startup", "failed to create agent backend client")
		return err
	}

	// Minimal probe identity; can be enriched as types evolve.
	probeInfo := &types.ProbeInstance{
		ID: probeID,
	}

	// Initial registration (best-effort; do not exit on failure).
	if err := agentClient.RegisterProbe(ctx, probeInfo); err != nil {
		log.Warn(ctx, "startup", "probe registration failed; will rely on heartbeats/config-applied to converge",
			"probe_id", probeID, "error", err.Error())
	}

	// Tickers for heartbeats and config sync
	if heartbeatInterval <= 0 {
		heartbeatInterval = 10 * time.Second
	}
	if configInterval <= 0 {
		configInterval = 30 * time.Second
	}
	heartbeatTicker := time.NewTicker(heartbeatInterval)
	configTicker := time.NewTicker(configInterval)
	defer heartbeatTicker.Stop()
	defer configTicker.Stop()

	log.Info(ctx, "runtime", "probe agent started", "probe_id", probeID, "backend", backendURL)

	// Main loop: heartbeats + config sync + graceful shutdown
	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown: attempt final heartbeat, stop telemetry, flush logs.
			shCtx, shCancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer shCancel()
			_ = agentClient.SendHeartbeat(shCtx, probeID)
			_ = telemMgr.Stop(shCtx)
			_ = logMgr.Flush()
			return nil

		case <-heartbeatTicker.C:
			hbCtx, hbCancel := context.WithTimeout(ctx, heartbeatInterval/2)
			if err := agentClient.SendHeartbeat(hbCtx, probeID); err != nil {
				log.Warn(hbCtx, "heartbeat", "heartbeat failed", "probe_id", probeID, "error", err.Error())
			}
			hbCancel()

		case <-configTicker.C:
			cfgCtx, cfgCancel := context.WithTimeout(ctx, configInterval/2)
			cfg, err := cfgManager.GetConfiguration(cfgCtx)
			if err != nil {
				log.Warn(cfgCtx, "config-sync", "failed to get configuration", "error", err.Error())
				cfgCancel()
				continue
			}

			if cfg != nil && cfg.ID != "" {
				if err := agentClient.ReportConfigApplied(cfgCtx, probeID, cfg, "config-manager", time.Now()); err != nil {
					log.Warn(cfgCtx, "config-applied", "failed to report config applied",
						"probe_id", probeID, "config_id", cfg.ID, "error", err.Error())
				}
			}
			cfgCancel()
		}
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
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no IP addresses found for target %q", target)
	}

	// Prefer IPv4 addresses if available
	for _, addr := range addrs {
		if addr.To4() != nil {
			return addr, nil
		}
	}

	// Return first address if no IPv4 is available
	return addrs[0], nil
}

// performTracerouteJSON performs a traceroute and outputs in JSON format
func performTracerouteJSON(ctx context.Context, engine *icmp.Engine, target net.IP) error {
	hops, err := engine.TraceRoute(ctx, target, maxHops)
	if err != nil {
		return fmt.Errorf("traceroute failed: %w", err)
	}

	hopData := make([]map[string]interface{}, 0, len(hops))
	for _, h := range hops {
		entry := map[string]interface{}{
			"ttl":     h.TTL,
			"ip":      h.IP.String(),
			"success": h.Success,
			"rtt":     h.RTT.String(),
		}
		if h.Host != "" {
			entry["hostname"] = h.Host
		}
		if h.ErrorMessage != "" {
			entry["error"] = h.ErrorMessage
		}
		hopData = append(hopData, entry)
	}

	out := map[string]interface{}{
		"target":   target.String(),
		"max_hops": maxHops,
		"hops":     hopData,
	}

	enc, jerr := json.MarshalIndent(out, "", "  ")
	if jerr != nil {
		return fmt.Errorf("failed to marshal traceroute json: %w", jerr)
	}
	fmt.Println(string(enc))
	return nil
}

// performTracerouteText performs a traceroute and outputs in text format
func performTracerouteText(ctx context.Context, engine *icmp.Engine, target net.IP) error {
	hops, err := engine.TraceRoute(ctx, target, maxHops)
	if err != nil {
		return fmt.Errorf("traceroute failed: %w", err)
	}

	if !jsonOutput {
		fmt.Printf("Tracing route to %s [%s]\n", target.String(), target.String())
		fmt.Printf("over a maximum of %d hops:\n\n", maxHops)
	}

	for _, h := range hops {
		if jsonOutput {
			continue
		}

		var rtt1, rtt2, rtt3 string
		if h.Success {
			rtt := h.RTT.Round(time.Millisecond)
			rtt1 = rtt.String()
			rtt2 = rtt.String()
			rtt3 = rtt.String()
		} else {
			rtt1, rtt2, rtt3 = "*", "*", "*"
		}

		if h.Host != "" {
			fmt.Printf("  %2d    %s    %s    %s  %s [%s]\n",
				h.TTL, rtt1, rtt2, rtt3, h.Host, h.IP.String())
		} else {
			fmt.Printf("  %2d    %s    %s    %s  %s\n",
				h.TTL, rtt1, rtt2, rtt3, h.IP.String())
		}
	}

	if !jsonOutput {
		fmt.Println("\nTrace complete.")
	}
	return nil
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
	if !jsonOutput {
		fmt.Printf("PING %s (%d measurements, interval %s):\n\n", target.String(), count, interval)
	}

	successful := 0
	failed := 0
	var totalRTT time.Duration
	var minRTT time.Duration = -1
	var maxRTT time.Duration

	var rttValues []time.Duration

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
				if !jsonOutput {
					fmt.Printf("Request %d: FAILED - %v\n", i+1, err)
				}
			} else {
				successful++

				measurement.PlatformData.OS = platformInfo.OS
				measurement.PlatformData.Architecture = platformInfo.Arch

				rtt := measurement.RTT
				totalRTT += rtt

				if minRTT == -1 || rtt < minRTT {
					minRTT = rtt
				}
				if rtt > maxRTT {
					maxRTT = rtt
				}

				rttValues = append(rttValues, rtt)

				if jsonOutput {
					_ = outputJSONMeasurement(measurement)
				} else {
					fmt.Printf("Request %d: time=%s ttl=%d size=%d bytes\n", i+1,
						rtt.Round(time.Microsecond), measurement.TTL, measurement.PacketSize)
				}
			}
		}
	}

	// Output summary for JSON mode
	if jsonOutput && successful > 0 {
		avgRTT := totalRTT / time.Duration(successful)
		lossRate := float64(failed) / float64(count) * 100

		// Calculate standard deviation
		var stdDevRTT time.Duration
		if len(rttValues) > 1 {
			var sumSeconds float64
			for _, rtt := range rttValues {
				sumSeconds += rtt.Seconds()
			}
			meanSeconds := sumSeconds / float64(len(rttValues))

			var varianceSum float64
			for _, rtt := range rttValues {
				deviation := rtt.Seconds() - meanSeconds
				varianceSum += deviation * deviation
			}
			variance := varianceSum / float64(len(rttValues)-1)
			stdDevSeconds := math.Sqrt(variance)
			stdDevRTT = time.Duration(stdDevSeconds * float64(time.Second))
		} else {
			stdDevRTT = 0
		}

		// RESTORE ORIGINAL SINGLE-LINE PRINTF (no trailing comma/paren bug)
		fmt.Printf(`{
  "summary": {
    "total_packets": %d,
    "successful_packets": %d,
    "failed_packets": %d,
    "packet_loss": %0.2f,
    "min_rtt": "%s",
    "avg_rtt": "%s",
    "max_rtt": "%s",
    "stddev_rtt": "%s"
  }
}`,
			count,
			successful,
			failed,
			lossRate,
			minRTT.Round(time.Microsecond),
			avgRTT.Round(time.Microsecond),
			maxRTT.Round(time.Microsecond),
			stdDevRTT.Round(time.Microsecond))
	}

	var stdDevRTT time.Duration
	if len(rttValues) > 1 {
		var sumSeconds float64
		for _, rtt := range rttValues {
			sumSeconds += rtt.Seconds()
		}
		meanSeconds := sumSeconds / float64(len(rttValues))

		var varianceSum float64
		for _, rtt := range rttValues {
			deviation := rtt.Seconds() - meanSeconds
			varianceSum += deviation * deviation
		}
		variance := varianceSum / float64(len(rttValues)-1)
		stdDevSeconds := math.Sqrt(variance)
		stdDevRTT = time.Duration(stdDevSeconds * float64(time.Second))
	}

	if !jsonOutput && successful > 0 {
		avgRTT := totalRTT / time.Duration(successful)
		lossRate := float64(failed) / float64(count) * 100
		fmt.Printf("\n--- %s ping statistics ---\n", target.String())
		fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n", count, successful, lossRate)
		fmt.Printf("round-trip min/avg/max ")
		fmt.Printf("= %s/%s/%s\n",
			minRTT.Round(time.Microsecond),
			avgRTT.Round(time.Microsecond),
			maxRTT.Round(time.Microsecond))
		fmt.Printf("stddev = %s\n", stdDevRTT.Round(time.Microsecond))
	}

	return nil
}

// Output functions (restore to use types.MeasurementData correctly)

// outputJSONMeasurement outputs a MeasurementData as JSON
func outputJSONMeasurement(measurement *types.MeasurementData) error {
	payload := map[string]interface{}{
		"success":      measurement.Success,
		"target":       measurement.Target.Address.String(),
		"rtt":          measurement.RTT.Round(time.Microsecond).String(),
		"ttl":          measurement.TTL,
		"packet_size":  measurement.PacketSize,
		"timestamp":    measurement.Timestamp.Format(time.RFC3339Nano),
		"source_ip":    measurement.SourceIP.String(),
		"dest_ip":      measurement.DestIP.String(),
		"error_message": func() string {
			if measurement.Success {
				return ""
			}
			return measurement.ErrorMessage
		}(),
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal measurement json: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// outputTextMeasurement outputs a MeasurementData as text
func outputTextMeasurement(measurement *types.MeasurementData) error {
	fmt.Printf("PING %s:\n", measurement.Target.Address.String())
	if measurement.Success {
		fmt.Printf("time=%s ttl=%d size=%d bytes\n", measurement.RTT.Round(time.Microsecond), measurement.TTL, measurement.PacketSize)
	} else {
		fmt.Printf("REQUEST FAILED: %s\n", measurement.ErrorMessage)
	}
	return nil
}

// main is unchanged; it simply executes rootCmd (with the new agent subcommand included).
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}