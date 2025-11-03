package config

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// defaultManager implements the Manager interface
type defaultManager struct {
	providers    map[string]Provider
	handler      ConfigurationHandler
	config       *types.Configuration
	configMutex  sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	updateCount  int64
	lastUpdate   time.Time
	status       *ManagerStatus
}

// ManagerOption allows configuration of the manager
type ManagerOption func(*defaultManager)

// WithProvider adds a configuration provider
func WithProvider(provider Provider) ManagerOption {
	return func(m *defaultManager) {
		if m.providers == nil {
			m.providers = make(map[string]Provider)
		}
		m.providers[provider.Name()] = provider
	}
}

// WithConfigurationHandler sets the configuration change handler
func WithConfigurationHandler(handler ConfigurationHandler) ManagerOption {
	return func(m *defaultManager) {
		m.handler = handler
	}
}

// NewManager creates a new configuration manager
func NewManager(options ...ManagerOption) Manager {
	m := &defaultManager{
		providers: make(map[string]Provider),
		status: &ManagerStatus{
			Sources: make(map[string]SourceStatus),
		},
	}

	for _, option := range options {
		option(m)
	}

	return m
}

// Initialize sets up the configuration manager
func (m *defaultManager) Initialize(ctx context.Context, config *types.Configuration) error {
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.config = config

	// Initialize all providers
	for name, provider := range m.providers {
		// Set up source status
		m.status.Sources[name] = SourceStatus{
			Name:     name,
			Enabled:  true,
			Healthy:  false,
			LastSeen: time.Time{},
			Priority: m.getProviderPriority(name),
		}

		// Initialize provider
		if err := provider.Initialize(m.ctx, config.Auth); err != nil {
			status := m.status.Sources[name]
			status.Healthy = false
			status.LastError = fmt.Sprintf("initialization failed: %v", err)
			m.status.Sources[name] = status
			continue
		}

		// Start watching for configuration changes
		if err := provider.Watch(m.ctx, m.providerHandler(name)); err != nil {
			status := m.status.Sources[name]
			status.LastError = fmt.Sprintf("watch failed: %v", err)
			m.status.Sources[name] = status
			continue
		}

		status := m.status.Sources[name]
		status.Healthy = true
		status.LastSeen = time.Now()
		m.status.Sources[name] = status
	}

	// Load initial configuration from the highest priority healthy provider
	if err := m.loadInitialConfiguration(); err != nil {
		return fmt.Errorf("failed to load initial configuration: %w", err)
	}

	return nil
}

// GetConfiguration retrieves the current configuration
func (m *defaultManager) GetConfiguration(ctx context.Context) (*types.Configuration, error) {
	m.configMutex.RLock()
	defer m.configMutex.RUnlock()

	if m.config == nil {
		return nil, fmt.Errorf("no configuration available")
	}

	return m.config, nil
}

// WatchConfiguration monitors configuration changes
func (m *defaultManager) WatchConfiguration(ctx context.Context, handler ConfigurationHandler) error {
	m.handler = handler
	return nil
}

// ReloadConfiguration forces a configuration reload from all sources
func (m *defaultManager) ReloadConfiguration(ctx context.Context) error {
	return m.loadInitialConfiguration()
}

// GetStatus returns the health status of all configuration sources
func (m *defaultManager) GetStatus(ctx context.Context) (*ManagerStatus, error) {
	// Update health scores for all providers
	healthyCount := 0
	totalCount := len(m.providers)

	for name := range m.providers {
		if status, exists := m.status.Sources[name]; exists {
			if status.Healthy {
				healthyCount++
			}
		}
	}

	if totalCount > 0 {
		m.status.HealthScore = float64(healthyCount) / float64(totalCount)
	}

	return m.status, nil
}

// Close shuts down the configuration manager
func (m *defaultManager) Close(ctx context.Context) error {
	if m.cancel != nil {
		m.cancel()
	}

	// Close all providers
	for _, provider := range m.providers {
		provider.Close(ctx)
	}

	return nil
}

// providerHandler creates a configuration handler for a specific provider
func (m *defaultManager) providerHandler(providerName string) ConfigurationHandler {
	return &providerHandler{
		manager:       m,
		providerName:  providerName,
		originalHandler: m.handler,
	}
}

// loadInitialConfiguration loads configuration from the highest priority healthy provider
func (m *defaultManager) loadInitialConfiguration() error {
	// Find the highest priority healthy provider
	var bestProvider Provider
	var bestPriority int
	bestName := ""

	for name, provider := range m.providers {
		status, exists := m.status.Sources[name]
		if !exists || !status.Enabled || !status.Healthy {
			continue
		}

		priority := m.getProviderPriority(name)
		if bestProvider == nil || priority > bestPriority {
			bestProvider = provider
			bestPriority = priority
			bestName = name
		}
	}

	if bestProvider == nil {
		return fmt.Errorf("no healthy configuration providers available")
	}

	// Load configuration from the best provider
	config, err := bestProvider.Get(m.ctx)
	if err != nil {
		status := m.status.Sources[bestName]
		status.LastError = fmt.Sprintf("load failed: %v", err)
		m.status.Sources[bestName] = status
		return fmt.Errorf("failed to load configuration from %s: %w", bestName, err)
	}

	// Update configuration
	m.configMutex.Lock()
	m.config = config
	m.updateCount++
	m.lastUpdate = time.Now()
	m.configMutex.Unlock()

	// Update status
	if status, exists := m.status.Sources[bestName]; exists {
		status.LastSeen = time.Now()
		status.UpdateCount++
		status.Healthy = true
		m.status.Sources[bestName] = status
	}

	// Notify handler
	if m.handler != nil {
		m.handler.HandleConfigurationUpdate(m.ctx, config, bestName)
	}

	return nil
}

// getProviderPriority returns the priority for a provider
func (m *defaultManager) getProviderPriority(name string) int {
	// Define provider priorities (higher number = higher priority)
	priorities := map[string]int{
		"etcd":    3,
		"consul":  2,
		"file":    1,
	}
	return priorities[name]
}

// providerHandler implements ConfigurationHandler for a specific provider
type providerHandler struct {
	manager          *defaultManager
	providerName     string
	originalHandler  ConfigurationHandler
}

// HandleConfigurationUpdate processes a configuration update
func (h *providerHandler) HandleConfigurationUpdate(ctx context.Context, config *types.Configuration, source string) error {
	// Update provider status
	if status, exists := h.manager.status.Sources[h.providerName]; exists {
		status.LastSeen = time.Now()
		status.UpdateCount++
		status.Healthy = true
		status.LastError = ""
		h.manager.status.Sources[h.providerName] = status
	}

	// Check if this configuration is better than what we currently have
	if h.shouldAcceptConfiguration(config) {
		h.manager.configMutex.Lock()
		h.manager.config = config
		h.manager.updateCount++
		h.manager.lastUpdate = time.Now()
		h.manager.configMutex.Unlock()
	}

	// Forward to original handler
	if h.originalHandler != nil {
		return h.originalHandler.HandleConfigurationUpdate(ctx, config, source)
	}

	return nil
}

// HandleConfigurationError handles configuration source errors
func (h *providerHandler) HandleConfigurationError(ctx context.Context, source string, err error) {
	// Update provider status
	if status, exists := h.manager.status.Sources[h.providerName]; exists {
		status.Healthy = false
		status.LastError = err.Error()
		h.manager.status.Sources[h.providerName] = status
	}

	// Forward to original handler
	if h.originalHandler != nil {
		h.originalHandler.HandleConfigurationError(ctx, source, err)
	}
}

// shouldAcceptConfiguration determines if a configuration update should be accepted
func (h *providerHandler) shouldAcceptConfiguration(config *types.Configuration) bool {
	// For now, accept all updates
	// In a more sophisticated implementation, you might check version numbers,
	// apply conflict resolution, etc.
	return true
}