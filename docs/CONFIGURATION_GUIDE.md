# Configuration Provider Usage Guide

This guide explains how to configure the Mesh Probe System to use external configuration providers, with a focus on etcd as a configuration source.

## Overview

The Mesh Probe System supports multiple configuration sources through a unified configuration manager:

- **File-based**: Local JSON configuration files
- **etcd**: Distributed key-value store
- **Consul**: HashiCorp's service discovery and configuration
- **Environment Variables**: Runtime configuration
- **CLI Flags**: Command-line overrides

## etcd Configuration Provider

etcd is a distributed, reliable key-value store for critical application data. It provides:

- ✅ **High Availability**: Clustered deployment for reliability
- ✅ **Real-time Updates**: Watch for configuration changes
- ✅ **Versioning**: Built-in revision tracking
- ✅ **Atomic Operations**: Consistent configuration updates
- ✅ **TLS Security**: Encrypted communication and authentication

### Requirements

#### System Requirements
- etcd cluster (v3.0+ recommended)
- Network access from probe nodes to etcd cluster
- Firewall rules allowing etcd client communication (default port 2379)

#### Data Requirements
- JSON format for configuration storage
- Consistent key structure (`{prefix}/probe/config`)
- Valid configuration JSON structure matching `types.Configuration`

### Configuration Example

#### 1. etcd Setup

First, ensure you have an etcd cluster running:

```bash
# Single-node etcd for testing
etcd --data-dir=/tmp/etcd --listen-client-urls=http://0.0.0.0:2379 --advertise-client-urls=http://localhost:2379

# Production cluster (3-node example)
# Node 1
etcd --name=node1 --initial-advertise-peer-urls=http://node1:2380 --listen-peer-urls=http://node1:2380 \
  --listen-client-urls=http://node1:2379 --advertise-client-urls=http://node1:2379 \
  --initial-cluster=node1=http://node1:2380,node2=http://node2:2380,node3=http://node3:2380 \
  --initial-cluster-state=new --initial-cluster-token=etcd-cluster

# Node 2 & 3 (similar configuration with appropriate names/IPs)
```

#### 2. Configuration Structure

The configuration is stored in etcd as JSON at the key `{keyPrefix}/probe/config`:

```bash
# Store configuration in etcd
etcdctl put /probe/config '{"id":"production-v1","name":"Production Config","version":1,"targets":[...]}'
```

#### 3. Probe Configuration

Update your probe's configuration to use etcd:

**Option A: Environment Variables**
```bash
export ETCD_ENDPOINTS=http://etcd1:2379,http://etcd2:2379,http://etcd3:2379
export ETCD_KEY_PREFIX=/probe
export ETCD_TIMEOUT=10s
export ETCD_TLS_ENABLED=false

./probe.exe ping 8.8.8.8
```

**Option B: Configuration File**
```json
{
  "provider": {
    "type": "etcd",
    "endpoints": ["http://etcd1:2379", "http://etcd2:2379", "http://etcd3:2379"],
    "key_prefix": "/probe",
    "timeout": "10s",
    "tls": {
      "enabled": false,
      "cert_file": "/path/to/client.crt",
      "key_file": "/path/to/client.key",
      "ca_cert_file": "/path/to/ca.crt",
      "server_name": "etcd-cluster"
    }
  },
  "config": {
    "id": "production-v1",
    "name": "Production Configuration",
    "version": 1,
    "targets": [...],
    "network": {...},
    "telemetry": {...}
  }
}
```

#### 4. TLS Authentication (Production)

For production deployments, enable TLS authentication:

**etcd Server Configuration:**
```bash
etcd --name=node1 \
  --cert-file=/path/to/server.crt \
  --key-file=/path/to/server.key \
  --trusted-ca-file=/path/to/ca.crt \
  --client-cert-auth \
  --advertise-client-urls=https://node1:2379 \
  --listen-client-urls=https://0.0.0.0:2379
```

**Client Configuration:**
```json
{
  "provider": {
    "type": "etcd",
    "endpoints": ["https://etcd1:2379", "https://etcd2:2379", "https://etcd3:2379"],
    "key_prefix": "/probe",
    "timeout": "10s",
    "auth": {
      "cert_file": "/path/to/client.crt",
      "key_file": "/path/to/client.key",
      "ca_cert_file": "/path/to/ca.crt",
      "server_name": "etcd-cluster"
    }
  }
}
```

### Configuration Data Structure

The etcd provider stores the complete `types.Configuration` structure:

```json
{
  "id": "prod-measurement-config",
  "name": "Production Measurement Config",
  "version": 1,
  "created_at": "2025-11-03T14:00:00Z",
  "updated_at": "2025-11-03T14:00:00Z",
  "targets": [
    {
      "id": "google_dns",
      "display_name": "Google DNS",
      "address": "8.8.8.8",
      "port": 0,
      "enabled": true,
      "priority": 10,
      "timeout": "5s",
      "interval": "1s",
      "expected_rtt": "20ms",
      "min_rtt": "5ms",
      "max_rtt": "100ms",
      "max_loss_pct": 1.0,
      "tags": ["dns", "public"],
      "region": "us-west",
      "environment": "production"
    },
    {
      "id": "cloudflare_dns",
      "display_name": "Cloudflare DNS",
      "address": "1.1.1.1",
      "port": 0,
      "enabled": true,
      "priority": 8,
      "timeout": "5s",
      "interval": "1s",
      "tags": ["dns", "public"],
      "region": "global",
      "environment": "production"
    }
  ],
  "intervals": {
    "default_interval": "1s",
    "min_interval": "100ms",
    "max_interval": "10m",
    "backoff_multiplier": 1.5
  },
  "network": {
    "source_ip": "",
    "interface": "",
    "ttl": 64,
    "buffer_size": 2048,
    "bind_port": 0
  },
  "telemetry": {
    "enable_metrics": true,
    "enable_tracing": true,
    "enable_logging": true,
    "log_level": "info",
    "log_format": "json",
    "export_interval": "30s"
  },
  "mesh": {
    "enabled": false,
    "discovery_method": "multicast",
    "heartbeat_interval": "30s",
    "max_neighbors": 10
  }
}
```

### Real-time Configuration Updates

The etcd provider supports watching for configuration changes:

#### Watch for Changes
```bash
# In one terminal, start the probe
./probe.exe ping 8.8.8.8

# In another terminal, update the configuration
etcdctl put /probe/config '{"id":"updated-config","name":"Updated Config","version":2,"targets":[...]}'
```

The probe will automatically:
1. Detect the configuration change
2. Validate the new configuration
3. Apply the changes without restart
4. Log the update event

### Health Monitoring

Monitor etcd connectivity:

```bash
# Check provider health
curl -s http://localhost:8080/config/health | jq

# Expected response:
{
  "sources": {
    "etcd": {
      "name": "etcd",
      "enabled": true,
      "healthy": true,
      "last_seen": "2025-11-03T14:23:00Z",
      "response_time": "5ms",
      "update_count": 42
    }
  },
  "current_config_id": "prod-measurement-config",
  "health_score": 1.0
}
```

### Troubleshooting

#### Common Issues

**Connection Timeout**
```bash
# Check etcd connectivity
etcdctl --endpoints=http://etcd:2379 endpoint health

# Test with curl
curl -s http://etcd:2379/health
```

**Authentication Errors**
```bash
# Check TLS certificates
openssl x509 -in /path/to/client.crt -text -noout

# Verify etcd server certificate
curl -s --cacert /path/to/ca.crt https://etcd:2379/health
```

**Configuration Not Found**
```bash
# List all keys under prefix
etcdctl get --prefix /probe/

# Check specific configuration key
etcdctl get /probe/config --print-value-only | jq .
```

**Watch Not Working**
```bash
# Check if etcd watch is enabled (should show events)
etcdctl get /probe/config --watch
```

### Best Practices

#### 1. High Availability
- Use at least 3-node etcd cluster
- Enable TLS in production
- Configure client timeout appropriately
- Monitor cluster health

#### 2. Configuration Management
- Version your configurations
- Use descriptive IDs and names
- Implement configuration validation
- Document configuration changes

#### 3. Security
- Enable TLS authentication
- Use certificate rotation
- Restrict network access to etcd cluster
- Monitor authentication failures

#### 4. Monitoring
- Track configuration update frequency
- Monitor etcd cluster performance
- Alert on provider failures
- Log configuration changes

### Migration from File-based Config

To migrate from file-based to etcd configuration:

1. **Export current configuration:**
   ```bash
   cat config.json | jq '.'
   ```

2. **Store in etcd:**
   ```bash
   etcdctl put /probe/config "$(cat config.json | jq -c .)"
   ```

3. **Update probe startup:**
   ```bash
   # Old: ./probe.exe -c config.json ping 8.8.8.8
   # New: ./probe.exe --provider etcd --etcd-endpoints=http://etcd:2379 ping 8.8.8.8
   ```

4. **Verify configuration loading:**
   ```bash
   ./probe.exe --provider etcd --etcd-endpoints=http://etcd:2379 ping 8.8.8.8 -v
   ```

### Integration Examples

#### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mesh-probe
spec:
  template:
    spec:
      containers:
      - name: probe
        image: mesh-probe:latest
        env:
        - name: ETCD_ENDPOINTS
          value: "http://etcd-service:2379"
        - name: ETCD_KEY_PREFIX
          value: "/probe"
        - name: ETCD_TIMEOUT
          value: "10s"
```

#### Docker Compose
```yaml
version: '3.8'
services:
  etcd:
    image: gcr.io/etcd-development/etcd:v3.5
    ports:
      - "2379:2379"
    command: |
      etcd --listen-client-urls=http://0.0.0.0:2379
      --advertise-client-urls=http://localhost:2379

  probe:
    image: mesh-probe:latest
    environment:
      - ETCD_ENDPOINTS=http://etcd:2379
      - ETCD_KEY_PREFIX=/probe
    depends_on:
      - etcd
```

This comprehensive guide provides everything needed to configure and operate the Mesh Probe System with etcd as a configuration source.