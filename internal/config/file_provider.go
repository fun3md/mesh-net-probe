package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// fileProvider implements configuration management for local files
type fileProvider struct {
	configDir string
	fileName  string
	timeout   time.Duration
}

// NewFileProvider creates a new file-based configuration provider
func NewFileProvider(configDir string, fileName string, timeout time.Duration) (Provider, error) {
	return &fileProvider{
		configDir: configDir,
		fileName:  fileName,
		timeout:   timeout,
	}, nil
}

// Name returns the provider name
func (p *fileProvider) Name() string {
	return "file"
}

// Initialize sets up the file provider
func (p *fileProvider) Initialize(ctx context.Context, authConfig *types.AuthConfig) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(p.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// Get retrieves configuration from file
func (p *fileProvider) Get(ctx context.Context) (*types.Configuration, error) {
	filePath := filepath.Join(p.configDir, p.fileName)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Unmarshal JSON
	var config types.Configuration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &config, nil
}

// Set stores configuration to file
func (p *fileProvider) Set(ctx context.Context, config *types.Configuration) error {
	filePath := filepath.Join(p.configDir, p.fileName)

	// Marshal to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// Watch monitors configuration file changes
func (p *fileProvider) Watch(ctx context.Context, handler ConfigurationHandler) error {
	go func() {
		filePath := filepath.Join(p.configDir, p.fileName)
		
		// Create a channel to receive file system events
		// Since we don't have a file watcher dependency, we'll poll for changes
		var lastModTime time.Time

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second): // Poll every second
				// Get file info
				info, err := os.Stat(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						continue // File doesn't exist yet
					}
					handler.HandleConfigurationError(ctx, p.Name(), fmt.Errorf("failed to stat file: %w", err))
					continue
				}

				// Check if file was modified
				if lastModTime.IsZero() {
					lastModTime = info.ModTime()
					continue
				}

				if info.ModTime().After(lastModTime) {
					lastModTime = info.ModTime()
					
					// Load and process the updated configuration
					config, err := p.Get(ctx)
					if err != nil {
						handler.HandleConfigurationError(ctx, p.Name(), fmt.Errorf("failed to load updated configuration: %w", err))
						continue
					}
					
					handler.HandleConfigurationUpdate(ctx, config, p.Name())
				}
			}
		}
	}()

	return nil
}

// HealthCheck verifies file accessibility
func (p *fileProvider) HealthCheck(ctx context.Context) error {
	filePath := filepath.Join(p.configDir, p.fileName)
	_, err := os.Stat(filePath)
	return err
}

// Close shuts down the file provider
func (p *fileProvider) Close(ctx context.Context) error {
	// File provider doesn't require explicit closing
	return nil
}