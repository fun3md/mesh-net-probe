package config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/mesh-net-probe/probe/pkg/types"
)

// etcdProvider implements configuration management for etcd
type etcdProvider struct {
	client    *clientv3.Client
	config    *types.Configuration
	keyPrefix string
	timeout   time.Duration
}

// NewEtcdProvider creates a new etcd configuration provider
func NewEtcdProvider(endpoints []string, keyPrefix string, timeout time.Duration) (Provider, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &etcdProvider{
		client:    client,
		keyPrefix: keyPrefix,
		timeout:   timeout,
	}, nil
}

// Provider interface defines the contract for configuration providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Initialize sets up the provider
	Initialize(ctx context.Context, config *types.AuthConfig) error

	// Get retrieves configuration from this provider
	Get(ctx context.Context) (*types.Configuration, error)

	// Set stores configuration in this provider
	Set(ctx context.Context, config *types.Configuration) error

	// Watch monitors configuration changes
	Watch(ctx context.Context, handler ConfigurationHandler) error

	// HealthCheck verifies provider health
	HealthCheck(ctx context.Context) error

	// Close shuts down the provider
	Close(ctx context.Context) error
}

// Name returns the provider name
func (p *etcdProvider) Name() string {
	return "etcd"
}

// Initialize sets up the etcd provider
func (p *etcdProvider) Initialize(ctx context.Context, authConfig *types.AuthConfig) error {
	// If auth is configured, set up authentication
	if authConfig != nil && authConfig.TLSConfig != nil && authConfig.TLSConfig.Enabled {
		// TLS authentication would be configured here
		// This is a simplified implementation
	}

	// Test connectivity
	_, err := p.client.Get(ctx, "health", clientv3.WithLimit(1))
	if err != nil {
		return fmt.Errorf("failed to connect to etcd: %w", err)
	}

	return nil
}

// Get retrieves configuration from etcd
func (p *etcdProvider) Get(ctx context.Context) (*types.Configuration, error) {
	resp, err := p.client.Get(ctx, p.configKey(), clientv3.WithLimit(1))
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration from etcd: %w", err)
	}

	if resp.Count == 0 {
		return nil, fmt.Errorf("configuration not found in etcd")
	}

	var config types.Configuration
	if err := json.Unmarshal(resp.Kvs[0].Value, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &config, nil
}

// Set stores configuration in etcd
func (p *etcdProvider) Set(ctx context.Context, config *types.Configuration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	_, err = p.client.Put(ctx, p.configKey(), string(data))
	if err != nil {
		return fmt.Errorf("failed to store configuration in etcd: %w", err)
	}

	return nil
}

// Watch monitors configuration changes in etcd
func (p *etcdProvider) Watch(ctx context.Context, handler ConfigurationHandler) error {
	watcher := clientv3.NewWatcher(p.client)

	go func() {
		defer watcher.Close()

		// Watch the configuration key
		watchChan := watcher.Watch(ctx, p.configKey(), clientv3.WithRev(0))

		for {
			select {
			case <-ctx.Done():
				return
			case watchResp := <-watchChan:
				for _, event := range watchResp.Events {
					if event.Type == clientv3.EventTypePut {
						var config types.Configuration
						if err := json.Unmarshal(event.Kv.Value, &config); err != nil {
							handler.HandleConfigurationError(ctx, p.Name(), fmt.Errorf("failed to unmarshal configuration: %w", err))
							continue
						}
						handler.HandleConfigurationUpdate(ctx, &config, p.Name())
					}
				}
			}
		}
	}()

	return nil
}

// HealthCheck verifies etcd connectivity
func (p *etcdProvider) HealthCheck(ctx context.Context) error {
	_, err := p.client.Get(ctx, "health", clientv3.WithLimit(1))
	return err
}

// Close shuts down the etcd provider
func (p *etcdProvider) Close(ctx context.Context) error {
	return p.client.Close()
}

// configKey returns the configuration key in etcd
func (p *etcdProvider) configKey() string {
	return fmt.Sprintf("%s/probe/config", p.keyPrefix)
}