package config

import (
	"context"
	"fmt"
	"time"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// Manager interface provides unified access to configuration from multiple sources
type Manager interface {
	// Initialize sets up the configuration manager with the given configuration
	Initialize(ctx context.Context, config *types.Configuration) error

	// GetConfiguration retrieves the current configuration
	GetConfiguration(ctx context.Context) (*types.Configuration, error)

	// WatchConfiguration monitors configuration changes
	WatchConfiguration(ctx context.Context, handler ConfigurationHandler) error

	// ReloadConfiguration forces a configuration reload from all sources
	ReloadConfiguration(ctx context.Context) error

	// GetStatus returns the health status of all configuration sources
	GetStatus(ctx context.Context) (*ManagerStatus, error)

	// Close shuts down the configuration manager
	Close(ctx context.Context) error
}

// ConfigurationHandler processes configuration updates
type ConfigurationHandler interface {
	// HandleConfigurationUpdate processes a configuration update
	HandleConfigurationUpdate(ctx context.Context, config *types.Configuration, source string) error

	// HandleConfigurationError handles configuration source errors
	HandleConfigurationError(ctx context.Context, source string, err error)
}

// ManagerStatus contains health information about configuration sources
type ManagerStatus struct {
	Sources         map[string]SourceStatus `json:"sources"`          // Status of each configuration source
	CurrentConfigID string                  `json:"current_config_id"` // Currently active configuration ID
	LastUpdate      time.Time               `json:"last_update"`      // Last successful configuration update
	UpdateCount     int64                   `json:"update_count"`     // Total number of configuration updates
	HealthScore     float64                 `json:"health_score"`     // Overall configuration health (0.0-1.0)
}

// SourceStatus represents the health status of a single configuration source
type SourceStatus struct {
	Name        string    `json:"name"`        // Source name (etcd, consul, file, etc.)
	Enabled     bool      `json:"enabled"`     // Whether this source is enabled
	Healthy     bool      `json:"healthy"`     // Whether this source is responding
	LastSeen    time.Time `json:"last_seen"`   // Last successful response time
	LastError   string    `json:"last_error"`  // Last error message
	ResponseTime time.Duration `json:"response_time"` // Average response time
	UpdateCount int64     `json:"update_count"`      // Number of successful updates from this source
	Priority    int       `json:"priority"`    // Priority (higher number = higher priority)
}

// DefaultConfigurationHandler implements ConfigurationHandler interface
type DefaultConfigurationHandler struct {
	onUpdate func(ctx context.Context, config *types.Configuration, source string) error
	onError  func(ctx context.Context, source string, err error)
}

// NewDefaultConfigurationHandler creates a new default configuration handler
func NewDefaultConfigurationHandler(
	onUpdate func(ctx context.Context, config *types.Configuration, source string) error,
	onError func(ctx context.Context, source string, err error),
) *DefaultConfigurationHandler {
	return &DefaultConfigurationHandler{
		onUpdate: onUpdate,
		onError:  onError,
	}
}

// HandleConfigurationUpdate processes a configuration update
func (h *DefaultConfigurationHandler) HandleConfigurationUpdate(ctx context.Context, config *types.Configuration, source string) error {
	if h.onUpdate != nil {
		return h.onUpdate(ctx, config, source)
	}
	fmt.Printf("Configuration updated from source: %s\n", source)
	return nil
}

// HandleConfigurationError handles configuration source errors
func (h *DefaultConfigurationHandler) HandleConfigurationError(ctx context.Context, source string, err error) {
	if h.onError != nil {
		h.onError(ctx, source, err)
		return
	}
	fmt.Printf("Configuration error from source %s: %v\n", source, err)
}