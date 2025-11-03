# Mesh Net Probe

A high-performance, cross-platform ICMP measurement tool that automatically adapts to Windows, Linux, and macOS on x86_64 and ARM64 architectures with automatic optimization for each platform.

## 🚀 Quick Start

### Platform Support

| Platform | Architecture | Support Level | Key Features |
|----------|-------------|---------------|--------------|
| **Windows** | x86_64, ARM64 | ✅ Full | Admin privileges, Winsock, AVX2/NEON |
| **Linux** | x86_64, ARM64, ARM | ✅ Full | Raw sockets, CAP_NET_RAW, All SIMD |
| **macOS** | x86_64, ARM64 | ✅ Full | BPF, System permissions, All SIMD |

### Choose Your Deployment Method

#### Docker Deployment (Recommended)
```bash
# Quick Docker setup
docker run --rm -it \
  --cap-add=NET_RAW \
  -v $(pwd)/config.json:/config/config.json \
  mesh-probe start --config /config/config.json
```

#### Binary Deployment
```bash
# Linux (with capabilities)
sudo setcap cap_net_raw+ep probe
./probe start --config config.json

# Windows (run as Administrator)
probe.exe start --config config.json

# macOS (remove quarantine)
xattr -rd com.apple.quarantine probe
./probe start --config config.json
```

### Basic Configuration

#### Minimal Configuration
```json
{
  "id": "basic_probe",
  "name": "Basic Probe",
  "network": {
    "ttl": 64
  },
  "targets": [
    {
      "id": "local_test",
      "address": "127.0.0.1",
      "interval": "5s"
    }
  ]
}
```

#### Production Configuration
```json
{
  "id": "production_probe",
  "name": "Production Mesh Probe",
  "network": {
    "ttl": 64,
    "buffer_size": 8192,
    "interface": "auto"
  },
  "targets": [
    {
      "id": "gateway",
      "address": "192.168.1.1",
      "interval": "5s",
      "priority": 10
    },
    {
      "id": "dns_primary",
      "address": "8.8.8.8",
      "interval": "30s",
      "priority": 9
    }
  ],
  "telemetry": {
    "otlp_endpoint": "localhost:4317",
    "enable_tracing": true,
    "enable_metrics": true,
    "log_level": "info"
  }
}
```

### Platform Validation

```bash
# Test your setup
./probe --test-platform

# Expected output for Linux/amd64:
# OS: linux
# Architecture: amd64
# Capabilities: [icmp udp tcp raw_sockets sse2 avx2 aes]
# Precision: nanosecond
# Optimizations: {use_rdtsc: true, batch_size: 256}
```

## 📖 Documentation

### 📋 Core Documentation

- **[Quick Reference Guide](docs/quick-reference.md)** - Essential configuration patterns and commands
- **[Cross-Platform Configuration Guide](docs/cross-platform-configuration.md)** - Comprehensive configuration options
- **[Troubleshooting Guide](docs/troubleshooting.md)** - Common issues and solutions

### 🎯 Key Features

#### Cross-Platform Compatibility
- **Automatic Platform Detection**: Detects OS and architecture automatically
- **Architecture Optimizations**: 
  - **x86_64**: RDTSC timing, AVX2 vectorization, 256-packet batches
  - **ARM64**: ARM counter timing, NEON SIMD, 128-packet batches
- **Platform-Specific APIs**: Raw sockets (Linux), Winsock (Windows), BPF (macOS)

#### High-Performance ICMP Measurements
- **Microsecond Precision**: Optimized timing across all platforms
- **High-Frequency Monitoring**: Support for sub-second measurement intervals
- **Batch Processing**: Efficient packet processing with SIMD optimizations
- **Low Latency**: Minimal overhead measurements

#### Enterprise-Ready Features
- **OpenTelemetry Integration**: Built-in metrics, tracing, and logging
- **Mesh Networking**: Multi-probe coordination and discovery
- **Container Support**: Docker and Kubernetes deployment
- **Configuration Management**: Dynamic config updates with etcd/Consul support

## 🔧 Detailed Configuration

### Deployment Scenarios

#### 1. Desktop/Laptop Monitoring
```json
{
  "network": {
    "buffer_size": 4096
  },
  "targets": [
    {
      "id": "home_network",
      "address": "192.168.1.1",
      "interval": "10s",
      "timeout": "5s"
    }
  ],
  "telemetry": {
    "export_interval": "60s",
    "log_level": "info"
  }
}
```

#### 2. Server Monitoring
```json
{
  "network": {
    "buffer_size": 8192,
    "interface": "eth0"
  },
  "targets": [
    {
      "id": "load_balancer",
      "address": "10.0.1.10",
      "interval": "2s",
      "priority": 10
    },
    {
      "id": "database",
      "address": "10.0.1.20",
      "interval": "5s",
      "priority": 9
    }
  ],
  "telemetry": {
    "export_interval": "30s",
    "enable_tracing": true
  },
  "mesh": {
    "enabled": true,
    "discovery_method": "multicast"
  }
}
```

#### 3. Container/Orchestration
```json
{
  "network": {
    "interface": "eth0",
    "buffer_size": 4096
  },
  "targets": [
    {
      "id": "service_mesh",
      "address": "10.244.0.1",
      "interval": "15s"
    }
  ],
  "telemetry": {
    "enable_logging": false,
    "export_interval": "300s"
  },
  "mesh": {
    "enabled": false
  }
}
```

### Performance Tuning

#### High-Frequency Monitoring (Linux)
```json
{
  "network": {
    "buffer_size": 16384,
    "ttl": 64
  },
  "telemetry": {
    "export_interval": "5s"
  },
  "optimizations": {
    "batch_size": 512,
    "packet_alignment": 64
  }
}
```

#### Resource-Constrained Environments
```json
{
  "network": {
    "buffer_size": 2048
  },
  "telemetry": {
    "enable_logging": false,
    "export_interval": "300s"
  },
  "optimizations": {
    "batch_size": 32
  }
}
```

### Container Orchestration

#### Docker Compose
```yaml
version: '3.8'
services:
  mesh-probe:
    image: mesh-probe:latest
    network_mode: host
    cap_add:
      - NET_RAW
    volumes:
      - ./config:/config:ro
      - ./data:/data
    environment:
      - PROBE_LOG_LEVEL=info
    restart: unless-stopped
```

#### Kubernetes DaemonSet
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: mesh-probe
spec:
  selector:
    matchLabels:
      app: mesh-probe
  template:
    spec:
      hostNetwork: true
      containers:
      - name: mesh-probe
        image: mesh-probe:latest
        securityContext:
          capabilities:
            add: [NET_RAW]
```

## 📊 Monitoring and Metrics

### OpenTelemetry Integration

```json
{
  "telemetry": {
    "otlp_endpoint": "localhost:4317",
    "enable_tracing": true,
    "enable_metrics": true,
    "export_interval": "30s"
  }
}
```

### Prometheus Metrics

Access metrics at `http://localhost:8080/metrics`:

```
# HELP probe_measurements_total Total number of ICMP measurements
# TYPE probe_measurements_total counter
probe_measurements_total{success="true"} 150
probe_measurements_total{success="false"} 5

# HELP probe_rtt_microseconds ICMP round-trip time in microseconds
# TYPE probe_rtt_microseconds histogram
probe_rtt_microseconds_bucket{le="100"} 45
probe_rtt_microseconds_bucket{le="500"} 120
probe_rtt_microseconds_bucket{le="1000"} 150
```

## 🛠️ Troubleshooting

### Common Issues by Platform

| Platform | Issue | Quick Fix |
|----------|-------|-----------|
| Windows | "Access denied" for ICMP | Run as Administrator |
| Linux | "Operation not permitted" | Add `cap_net_raw` capability |
| macOS | Sandbox restrictions | Remove quarantine attribute |

### Diagnostic Commands

```bash
# Platform capability check
./probe --test-platform

# Configuration validation
./probe validate-config --config config.json

# Network diagnostics
./probe ping --target 8.8.8.8 --count 5

# Performance monitoring
./probe metrics --format prometheus
```

See the [Troubleshooting Guide](docs/troubleshooting.md) for detailed solutions.

## 🏗️ Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Mesh Net Probe                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐  │
│  │   Platform      │  │  Configuration  │  │   Network    │  │
│  │   Detection     │  │    Manager      │  │   Interface  │  │
│  └─────────────────┘  └─────────────────┘  └──────────────┘  │
│           │                     │                    │       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐  │
│  │    ICMP         │  │   Timing        │  │  Telemetry   │  │
│  │   Engine        │  │   Engine        │  │   Provider   │  │
│  └─────────────────┘  └─────────────────┘  └──────────────┘  │
│           │                     │                    │       │
└─────────────────────────────────────────────────────────────┘
           │                     │                    │
    ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
    │   Raw/      │      │   High      │      │ OpenTelemetry│
    │  Winsock    │      │Precision    │      │   Collector  │
    │  Sockets    │      │   Timing    │      │             │
    └─────────────┘      └─────────────┘      └─────────────┘
```

### Platform Abstraction Layer

The probe automatically adapts to platform capabilities:

- **Windows**: Winsock API with millisecond precision
- **Linux**: Raw sockets with nanosecond precision  
- **macOS**: BPF with microsecond precision

## 🧪 Testing and Validation

### Platform Compatibility Testing

```bash
# Test all supported platforms
./probe test-platform --cross-platform

# Architecture-specific tests
./probe test-architecture --arch amd64
./probe test-architecture --arch arm64

# Performance benchmarking
./probe benchmark --duration 60s
```

### Configuration Validation

```bash
# Strict validation
./probe validate-config --strict --config config.json

# Platform-specific validation
./probe validate-platform --platform linux/amd64

# Network interface validation
./probe check-interfaces
```

## 📚 Additional Resources

### External Documentation
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Docker Networking Guide](https://docs.docker.com/network/)
- [Kubernetes Networking Concepts](https://kubernetes.io/docs/concepts/cluster-networking/)

### Community and Support
- **Issues**: Report bugs and feature requests
- **Discussions**: Ask questions and share configurations
- **Wiki**: Community-contributed configurations and scripts

### Development
- **API Reference**: See [pkg/types/](pkg/types/) for type definitions
- **Source Code**: Architecture and implementation details
- **Contributing**: Guidelines for contributing to the project

---

## Quick Links

- **[Quick Start](docs/quick-reference.md)** - Essential commands and patterns
- **[Configuration Guide](docs/cross-platform-configuration.md)** - Complete configuration reference  
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions
- **[GitHub Repository](https://github.com/mesh-net-probe/probe)** - Source code and issues

---

## Getting Started with Configuration

For immediate deployment, see the [Quick Reference Guide](docs/quick-reference.md) for common configuration patterns and platform-specific commands.

For comprehensive configuration options, see the [Cross-Platform Configuration Guide](docs/cross-platform-configuration.md#configuration-options) including:
- [Platform-Specific Settings](docs/cross-platform-configuration.md#platform-specific-settings)
- [Architecture Optimizations](docs/cross-platform-configuration.md#architecture-optimizations)
- [Performance Tuning](docs/cross-platform-configuration.md#performance-tuning)

For troubleshooting common issues, see the [Troubleshooting Guide](docs/troubleshooting.md) covering:
- Platform-specific problems and solutions
- Network configuration issues
- Permission and deployment errors

---

*For the latest updates and detailed API documentation, visit the [project repository](https://github.com/mesh-net-probe/probe).*