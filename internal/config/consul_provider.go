package config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	consul "github.com/hashicorp/consul/api"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// consulProvider implements configuration management for Consul
type consulProvider struct {
	client    *consul.Client
	config    *types.Configuration
	keyPrefix string
	timeout   time.Duration
}

// NewConsulProvider creates a new Consul configuration provider
func NewConsulProvider(address string, keyPrefix string, timeout time.Duration) (Provider, error) {
	config := consul.DefaultConfig()
	config.Address = address

	client, err := consul.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	return &consulProvider{
		client:    client,
		keyPrefix: keyPrefix,
		timeout:   timeout,
	}, nil
}

// Name returns the provider name
func (p *consulProvider) Name() string {
	return "consul"
}

// Initialize sets up the Consul provider
func (p *consulProvider) Initialize(ctx context.Context, authConfig *types.AuthConfig) error {
	// If auth is configured, set up authentication
	if authConfig != nil && authConfig.TLSConfig != nil && authConfig.TLSConfig.Enabled {
		// TLS authentication would be configured here
		// This is a simplified implementation
	}

	// Test connectivity
	_, _, err := p.client.KV().Get("health", &consul.QueryOptions{})
	if err != nil {
		return fmt.Errorf("failed to connect to Consul: %w", err)
	}

	return nil
}

// Get retrieves configuration from Consul
func (p *consulProvider) Get(ctx context.Context) (*types.Configuration, error) {
	key := p.configKey()
	kvPair, _, err := p.client.KV().Get(key, &consul.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration from Consul: %w", err)
	}

	if kvPair == nil {
		return nil, fmt.Errorf("configuration not found in Consul")
	}

	var config types.Configuration
	if err := json.Unmarshal(kvPair.Value, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &config, nil
}

// Set stores configuration in Consul
func (p *consulProvider) Set(ctx context.Context, config *types.Configuration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	key := p.configKey()
	_, err = p.client.KV().Put(&consul.KVPair{
		Key:   key,
		Value: data,
	}, &consul.WriteOptions{})
	if err != nil {
		return fmt.Errorf("failed to store configuration in Consul: %w", err)
	}

	return nil
}

// Watch monitors configuration changes in Consul
func (p *consulProvider) Watch(ctx context.Context, handler ConfigurationHandler) error {
	watcher := NewConsulWatcher(p.client)

	go func() {
		defer watcher.Stop()

		// Watch the configuration key
		key := p.configKey()
		watchChan := make(chan consul.KVPairs, 1)
		stopChan := make(chan struct{}, 1)

		opts := &consul.QueryOptions{}
		if p.timeout > 0 {
			opts.WaitTime = p.timeout
		}

		err := watcher.WatchKey(key, opts, watchChan, stopChan)
		if err != nil {
			handler.HandleConfigurationError(ctx, p.Name(), fmt.Errorf("failed to watch Consul key: %w", err))
			return
		}

		for {
			select {
			case <-ctx.Done():
				close(stopChan)
				return
			case kvPairs := <-watchChan:
				if len(kvPairs) == 0 {
					continue
				}

				var config types.Configuration
				if err := json.Unmarshal(kvPairs[0].Value, &config); err != nil {
					handler.HandleConfigurationError(ctx, p.Name(), fmt.Errorf("failed to unmarshal configuration: %w", err))
					continue
				}
				handler.HandleConfigurationUpdate(ctx, &config, p.Name())
			}
		}
	}()

	return nil
}

// HealthCheck verifies Consul connectivity
func (p *consulProvider) HealthCheck(ctx context.Context) error {
	_, _, err := p.client.KV().Get("health", &consul.QueryOptions{})
	return err
}

// Close shuts down the Consul provider
func (p *consulProvider) Close(ctx context.Context) error {
	// Consul client doesn't require explicit closing
	return nil
}

// configKey returns the configuration key in Consul
func (p *consulProvider) configKey() string {
	return fmt.Sprintf("%s/probe/config", p.keyPrefix)
}

// ConsulWatcher is a simple wrapper for Consul key watching
type ConsulWatcher struct {
	client *consul.Client
	stopCh chan struct{}
}

// NewConsulWatcher creates a new Consul watcher
func NewConsulWatcher(client *consul.Client) *ConsulWatcher {
	return &ConsulWatcher{
		client: client,
		stopCh: make(chan struct{}),
	}
}

// WatchKey watches a Consul key for changes
func (w *ConsulWatcher) WatchKey(key string, opts *consul.QueryOptions, updates chan consul.KVPairs, stopCh chan struct{}) error {
	go func() {
		var index uint64

		for {
			select {
			case <-w.stopCh:
				return
			case <-stopCh:
				return
			default:
				opts.WaitIndex = index
				kvPairs, meta, err := w.client.KV().List(key, opts)
				if err != nil {
					// For now, just log the error and continue
					continue
				}

				if meta.LastIndex != index {
					index = meta.LastIndex
					updates <- kvPairs
				}

				time.Sleep(time.Second) // Avoid tight polling
			}
		}
	}()

	return nil
}

// Stop stops the watcher
func (w *ConsulWatcher) Stop() {
	close(w.stopCh)
}