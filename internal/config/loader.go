package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// Loader handles loading and validation of probe configurations
type Loader struct {
	validator *Validator
	providers map[string]Provider
}

// Validator validates configuration data
type Validator struct {
	rules []ValidationRule
}

// ValidationRule defines a validation rule for configuration
type ValidationRule struct {
	Path      string
	Required  bool
	Type      string
	MinValue  interface{}
	MaxValue  interface{}
	Validator func(interface{}) error
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		validator: NewValidator(),
		providers: make(map[string]Provider),
	}
}

// Load loads configuration from various sources
func (l *Loader) Load(source string, key string) (*types.Configuration, error) {
	// Determine configuration source type
	sourceType := determineSourceType(source)

	switch sourceType {
	case "file":
		return l.loadFromFile(source, key)
	case "etcd":
		return l.loadFromETCD(source, key)
	case "consul":
		return l.loadFromConsul(source, key)
	default:
		return nil, fmt.Errorf("unsupported configuration source: %s", source)
	}
}

// determineSourceType determines the configuration source type
func determineSourceType(source string) string {
	// Check if it's a file path
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "/") || 
	   strings.HasSuffix(source, ".json") || strings.HasSuffix(source, ".yaml") || 
	   strings.HasSuffix(source, ".yml") || strings.HasSuffix(source, ".toml") {
		return "file"
	}

	// Check if it's a etcd/Consul URL
	if strings.HasPrefix(source, "etcd://") || strings.HasPrefix(source, "consul://") {
		parts := strings.Split(source, "://")
		if len(parts) >= 2 {
			return parts[0]
		}
	}

	// Default to file
	return "file"
}

// loadFromFile loads configuration from a file
func (l *Loader) loadFromFile(filePath, key string) (*types.Configuration, error) {
	// Expand environment variables
	expandedPath := os.ExpandEnv(filePath)

	// Create Viper instance
	v := viper.New()
	v.SetConfigFile(expandedPath)
	v.SetConfigType(determineFileType(expandedPath))

	// Set defaults and environment variable overrides
	l.setViperDefaults(v)

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", expandedPath, err)
	}

	// Load the specific configuration key if provided
	if key != "" {
		configMap := v.GetStringMap(key)
		if len(configMap) == 0 {
			return nil, fmt.Errorf("configuration key '%s' not found in file %s", key, expandedPath)
		}
		
		// Convert map to configuration
		config, err := l.mapToConfiguration(configMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse configuration key '%s': %w", key, err)
		}
		
		// Validate configuration
		if err := l.validator.Validate(config); err != nil {
			return nil, fmt.Errorf("configuration validation failed: %w", err)
		}
		
		return config, nil
	}

	// Load entire configuration
	var config types.Configuration
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Validate configuration
	if err := l.validator.Validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// determineFileType determines the file type from the extension
func determineFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	default:
		return "json" // Default to JSON
	}
}

// setViperDefaults sets default configuration values
func (l *Loader) setViperDefaults(v *viper.Viper) {
	// Set default values
	v.SetDefault("targets", []interface{}{})
	v.SetDefault("intervals.default_interval", "1s")
	v.SetDefault("intervals.min_interval", "100ms")
	v.SetDefault("intervals.max_interval", "30s")
	v.SetDefault("network.buffer_size", 65535)
	v.SetDefault("network.ttl", 64)
	v.SetDefault("telemetry.enable_metrics", true)
	v.SetDefault("telemetry.enable_tracing", true)
	v.SetDefault("telemetry.log_level", "info")
	v.SetDefault("mesh.enabled", false)
	v.SetDefault("mesh.heartbeat_interval", "30s")
	v.SetDefault("mesh.max_neighbors", 10)
}

// mapToConfiguration converts a map to a Configuration struct
func (l *Loader) mapToConfiguration(configMap map[string]interface{}) (*types.Configuration, error) {
	// This is a simplified conversion - in practice, you'd want more robust mapping
	config := &types.Configuration{
		ID:        generateConfigID(),
		Name:      "Default Configuration",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Extract targets
	if targets, ok := configMap["targets"].([]interface{}); ok {
		networkTargets, err := l.parseTargets(targets)
		if err != nil {
			return nil, err
		}
		config.Targets = networkTargets
	}

	// Extract intervals
	if intervals, ok := configMap["intervals"].(map[string]interface{}); ok {
		config.Intervals = &types.MeasurementIntervals{
			DefaultInterval: parseDuration(intervals["default_interval"], time.Second),
			MinInterval:     parseDuration(intervals["min_interval"], 100*time.Millisecond),
			MaxInterval:     parseDuration(intervals["max_interval"], 30*time.Second),
		}
	}

	// Extract network config
	if network, ok := configMap["network"].(map[string]interface{}); ok {
		config.Network = &types.NetworkConfig{
			BufferSize:    getInt(network["buffer_size"], 65535),
			TTL:           getInt(network["ttl"], 64),
			BindPort:      getInt(network["bind_port"], 0),
		}

		if sourceIP, ok := network["source_ip"].(string); ok && sourceIP != "" {
			if ip := net.ParseIP(sourceIP); ip != nil {
				config.Network.SourceIP = ip
			}
		}
	}

	// Extract telemetry config
	if telemetry, ok := configMap["telemetry"].(map[string]interface{}); ok {
		config.Telemetry = &types.TelemetryConfig{
			EnableMetrics:   getBool(telemetry["enable_metrics"], true),
			EnableTracing:   getBool(telemetry["enable_tracing"], true),
			EnableLogging:   getBool(telemetry["enable_logging"], true),
			LogLevel:        getString(telemetry["log_level"], "info"),
			OTLPEndpoint:    getString(telemetry["otlp_endpoint"], ""),
		}
	}

	// Extract mesh config
	if mesh, ok := configMap["mesh"].(map[string]interface{}); ok {
		config.Mesh = &types.MeshConfig{
			Enabled:            getBool(mesh["enabled"], false),
			DiscoveryMethod:    getString(mesh["discovery_method"], "manual"),
			DiscoveryPort:      getInt(mesh["discovery_port"], 9000),
			HeartbeatInterval:  parseDuration(mesh["heartbeat_interval"], 30*time.Second),
			MaxNeighbors:       getInt(mesh["max_neighbors"], 10),
		}
	}

	return config, nil
}

// parseTargets converts interface{} slice to NetworkTarget slice
func (l *Loader) parseTargets(targets []interface{}) ([]types.NetworkTarget, error) {
	networkTargets := make([]types.NetworkTarget, 0, len(targets))

	for i, target := range targets {
		targetMap, ok := target.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("target %d is not a map", i)
		}

		networkTarget, err := l.parseTarget(targetMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target %d: %w", i, err)
		}

		networkTargets = append(networkTargets, networkTarget)
	}

	return networkTargets, nil
}

// parseTarget converts a map to a NetworkTarget
func (l *Loader) parseTarget(targetMap map[string]interface{}) (types.NetworkTarget, error) {
	target := types.NetworkTarget{
		ID:          getString(targetMap["id"], generateTargetID()),
		DisplayName: getString(targetMap["display_name"], ""),
		Enabled:     getBool(targetMap["enabled"], true),
		Priority:    getInt(targetMap["priority"], 5),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Parse IP address
	if address, ok := targetMap["address"].(string); ok && address != "" {
		if ip := net.ParseIP(address); ip != nil {
			target.Address = ip
		} else {
			return target, fmt.Errorf("invalid IP address: %s", address)
		}
	}

	// Parse port
	if port, ok := targetMap["port"].(int); ok {
		target.Port = port
	} else if portFloat, ok := targetMap["port"].(float64); ok {
		target.Port = int(portFloat)
	}

	// Parse timing configuration
	target.Timeout = parseDuration(targetMap["timeout"], 5*time.Second)
	target.Interval = parseDuration(targetMap["interval"], time.Second)
	target.Jitter = getFloat64(targetMap["jitter"], 0.0)

	// Parse performance expectations
	target.ExpectedRTT = parseDuration(targetMap["expected_rtt"], time.Millisecond)
	target.MinRTT = parseDuration(targetMap["min_rtt"], 0)
	target.MaxRTT = parseDuration(targetMap["max_rtt"], time.Second*10)
	target.MaxLossPct = getFloat64(targetMap["max_loss_pct"], 0.0)

	// Parse metadata
	if tags, ok := targetMap["tags"].([]interface{}); ok {
		target.Tags = make([]string, len(tags))
		for i, tag := range tags {
			if str, ok := tag.(string); ok {
				target.Tags[i] = str
			}
		}
	}

	if labels, ok := targetMap["labels"].(map[string]interface{}); ok {
		target.Labels = make(map[string]string)
		for k, v := range labels {
			if str, ok := v.(string); ok {
				target.Labels[k] = str
			}
		}
	}

	target.Region = getString(targetMap["region"], "")
	target.Environment = getString(targetMap["environment"], "prod")

	return target, nil
}

// Helper functions for type conversion
func getString(val interface{}, defaultValue string) string {
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

func getInt(val interface{}, defaultValue int) int {
	if i, ok := val.(int); ok {
		return i
	}
	if f, ok := val.(float64); ok {
		return int(f)
	}
	if i, ok := val.(int64); ok {
		return int(i)
	}
	return defaultValue
}

func getFloat64(val interface{}, defaultValue float64) float64 {
	if f, ok := val.(float64); ok {
		return f
	}
	if i, ok := val.(int); ok {
		return float64(i)
	}
	return defaultValue
}

func getBool(val interface{}, defaultValue bool) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

func parseDuration(val interface{}, defaultValue time.Duration) time.Duration {
	if d, ok := val.(time.Duration); ok {
		return d
	}
	if str, ok := val.(string); ok && str != "" {
		if d, err := time.ParseDuration(str); err == nil {
			return d
		}
	}
	return defaultValue
}

// loadFromETCD loads configuration from etcd
func (l *Loader) loadFromETCD(source, key string) (*types.Configuration, error) {
	// This would implement etcd loading logic
	return nil, fmt.Errorf("etcd provider not yet implemented")
}

// loadFromConsul loads configuration from Consul
func (l *Loader) loadFromConsul(source, key string) (*types.Configuration, error) {
	// This would implement Consul loading logic
	return nil, fmt.Errorf("consul provider not yet implemented")
}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
	v := &Validator{
		rules: make([]ValidationRule, 0),
	}

	// Add validation rules
	v.addRules()
	return v
}

// addRules adds default validation rules
func (v *Validator) addRules() {
	v.rules = []ValidationRule{
		{
			Path:     "ID",
			Required: true,
			Type:     "string",
			Validator: func(val interface{}) error {
				if str, ok := val.(string); ok && len(str) > 0 {
					return nil
				}
				return fmt.Errorf("ID must be a non-empty string")
			},
		},
		{
			Path:     "Targets",
			Required: false,
			Type:     "array",
			Validator: func(val interface{}) error {
				if arr, ok := val.([]types.NetworkTarget); ok {
					for i, target := range arr {
						if target.Address == nil {
							return fmt.Errorf("target %d has no address", i)
						}
						if target.Timeout <= 0 {
							return fmt.Errorf("target %d has invalid timeout", i)
						}
					}
					return nil
				}
				return fmt.Errorf("Targets must be an array")
			},
		},
	}
}

// Validate validates a configuration
func (v *Validator) Validate(config *types.Configuration) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate required fields
	for _, rule := range v.rules {
		if rule.Required {
			if !v.hasField(config, rule.Path) {
				return fmt.Errorf("required field %s is missing", rule.Path)
			}
		}
	}

	// Apply validation rules
	for _, rule := range v.rules {
		if value := v.getField(config, rule.Path); value != nil {
			if err := rule.Validator(value); err != nil {
				return fmt.Errorf("validation failed for %s: %w", rule.Path, err)
			}
		}
	}

	// Custom validation
	return v.customValidate(config)
}

// hasField checks if a configuration has a specific field
func (v *Validator) hasField(config *types.Configuration, path string) bool {
	return v.getField(config, path) != nil
}

// getField gets a field value from configuration by path
func (v *Validator) getField(config *types.Configuration, path string) interface{} {
	switch path {
	case "ID":
		return config.ID
	case "Name":
		return config.Name
	case "Targets":
		return config.Targets
	case "Intervals":
		return config.Intervals
	case "Network":
		return config.Network
	case "Telemetry":
		return config.Telemetry
	case "Auth":
		return config.Auth
	case "Mesh":
		return config.Mesh
	default:
		return nil
	}
}

// customValidate performs custom validation logic
func (v *Validator) customValidate(config *types.Configuration) error {
	// Validate network configuration consistency
	if config.Network != nil {
		if config.Network.SourceIP != nil && !config.Network.SourceIP.IsUnspecified() {
			if config.Network.SourceIP.To4() == nil && config.Network.SourceIP.To16() == nil {
				return fmt.Errorf("invalid source IP address: %s", config.Network.SourceIP)
			}
		}
	}

	// Validate measurement intervals
	if config.Intervals != nil {
		if config.Intervals.DefaultInterval <= 0 {
			return fmt.Errorf("default interval must be positive")
		}
		if config.Intervals.MinInterval <= 0 {
			return fmt.Errorf("min interval must be positive")
		}
		if config.Intervals.MaxInterval <= 0 {
			return fmt.Errorf("max interval must be positive")
		}
		if config.Intervals.MinInterval > config.Intervals.DefaultInterval {
			return fmt.Errorf("min interval cannot be greater than default interval")
		}
		if config.Intervals.DefaultInterval > config.Intervals.MaxInterval {
			return fmt.Errorf("default interval cannot be greater than max interval")
		}
	}

	// Validate telemetry configuration
	if config.Telemetry != nil {
		if config.Telemetry.OTLPEndpoint != "" {
			if !strings.HasPrefix(config.Telemetry.OTLPEndpoint, "http://") && 
			   !strings.HasPrefix(config.Telemetry.OTLPEndpoint, "https://") {
				return fmt.Errorf("OTLP endpoint must be http:// or https://")
			}
		}
	}

	// Validate mesh configuration
	if config.Mesh != nil && config.Mesh.Enabled {
		if config.Mesh.HeartbeatInterval <= 0 {
			return fmt.Errorf("mesh heartbeat interval must be positive")
		}
		if config.Mesh.MaxNeighbors <= 0 {
			return fmt.Errorf("mesh max neighbors must be positive")
		}
	}

	return nil
}

// generateConfigID generates a unique configuration ID
func generateConfigID() string {
	return fmt.Sprintf("config_%d", time.Now().UnixNano())
}

// generateTargetID generates a unique target ID
func generateTargetID() string {
	return fmt.Sprintf("target_%d", time.Now().UnixNano())
}