# Cross-Platform Configuration Guide

This guide provides comprehensive documentation for configuring the Mesh Net Probe across different operating systems and architectures, including Windows, Linux, and macOS on x86_64 and ARM64 systems.

## Table of Contents

1. [Overview](#overview)
2. [Platform Detection](#platform-detection)
3. [Configuration Options](#configuration-options)
4. [Platform-Specific Settings](#platform-specific-settings)
5. [Architecture Optimizations](#architecture-optimizations)
6. [Examples](#examples)
7. [Validation](#validation)
8. [Troubleshooting](#troubleshooting)

## Overview

The Mesh Net Probe includes a sophisticated cross-platform configuration system that automatically detects the operating system and architecture, then applies appropriate optimizations and settings for optimal performance and compatibility.

### Supported Platforms

| Platform | Architecture | Status | Key Features |
|----------|-------------|--------|-------------|
| Windows 10+ | x86_64 | ✅ Full Support | AVX2, Winsock, Admin privileges |
| Windows 10+ | ARM64 | ✅ Full Support | NEON, Winsock, Admin privileges |
| Linux (Ubuntu, CentOS, etc.) | x86_64 | ✅ Full Support | Raw sockets, AVX2, Nanosecond precision |
| Linux (Ubuntu, CentOS, etc.) | ARM64 | ✅ Full Support | Raw sockets, NEON, Nanosecond precision |
| macOS 10.15+ | x86_64 | ✅ Full Support | BPF, SSE, Microsecond precision |
| macOS 10.15+ | ARM64 | ✅ Full Support | BPF, NEON, Microsecond precision |

## Platform Detection

The system automatically detects your platform and applies appropriate settings:

```go
// Example platform detection output
OS: windows
Architecture: amd64
Version: Windows 10
Kernel: Windows Kernel 10.0.19045
Hostname: workstation01
Container: false
```

### Detection Features

- **Operating System**: Windows, Linux, macOS identification
- **Architecture**: x86_64, ARM64, ARM detection
- **Container Awareness**: Docker, Kubernetes environments
- **System Information**: Kernel version, hostname, CPU features

## Configuration Options

### Basic Network Configuration

```json
{
  "network": {
    "source_ip": "0.0.0.0",
    "interface": "auto",
    "ttl": 64,
    "buffer_size": 4096,
    "bind_port": 0,
    "qos": "normal",
    "dscp": 0
  }
}
```

### Telemetry Configuration

```json
{
  "telemetry": {
    "otlp_endpoint": "localhost:4317",
    "enable_tracing": true,
    "enable_metrics": true,
    "enable_logging": true,
    "log_level": "info",
    "export_interval": "30s",
    "buffer_size": 1000
  }
}
```

### Authentication Configuration

```json
{
  "auth": {
    "provider": "local",
    "timeout": "10s",
    "retry_attempts": 3
  }
}
```

## Platform-Specific Settings

### Windows Configuration

Windows requires administrator privileges for ICMP operations and uses the Winsock API instead of raw sockets.

#### Configuration Example

```json
{
  "network": {
    "source_ip": "0.0.0.0",
    "interface": "auto",
    "ttl": 64,
    "buffer_size": 2048,
    "bind_port": 0,
    "qos": "normal",
    "dscp": 0
  },
  "targets": [
    {
      "id": "local_test",
      "display_name": "Local Test Target",
      "address": "127.0.0.1",
      "enabled": true,
      "timeout": "5s",
      "interval": "1s"
    }
  ]
}
```

#### Windows-Specific Behavior

- **Raw Sockets**: ❌ Disabled (requires elevated privileges)
- **Network API**: Winsock
- **Precision**: Millisecond
- **Timeout**: 5 seconds
- **Buffer Size**: 2048 bytes
- **Requirements**: Administrator privileges

#### Setting up Windows Permissions

1. **Run as Administrator**:
   ```cmd
   # Right-click Command Prompt → "Run as administrator"
   # Then run probe with elevated privileges
   probe.exe start --config config.json
   ```

2. **Windows Defender Firewall**:
   - Allow ICMP echo requests (ping)
   - Configure firewall rules for probe ports

### Linux Configuration

Linux provides the highest performance with raw socket access and nanosecond precision timing.

#### Configuration Example

```json
{
  "network": {
    "source_ip": "0.0.0.0",
    "interface": "auto",
    "ttl": 64,
    "buffer_size": 8192,
    "bind_port": 0,
    "qos": "normal",
    "dscp": 0
  }
}
```

#### Linux-Specific Behavior

- **Raw Sockets**: ✅ Enabled
- **Network API**: Raw sockets with CAP_NET_RAW
- **Precision**: Nanosecond
- **Timeout**: 30 seconds
- **Buffer Size**: 8192 bytes
- **Requirements**: CAP_NET_RAW capability or root

#### Setting up Linux Permissions

1. **Using Capabilities** (Recommended):
   ```bash
   # Set capabilities for non-root operation
   sudo setcap cap_net_raw+ep /path/to/probe
   
   # Run probe as regular user
   ./probe start --config config.json
   ```

2. **Using Root** (Legacy):
   ```bash
   # Run as root
   sudo ./probe start --config config.json
   ```

3. **Docker Configuration**:
   ```dockerfile
   # Add capabilities to container
   FROM ubuntu:22.04
   COPY probe /usr/local/bin/probe
   RUN chmod +x /usr/local/bin/probe
   ENTRYPOINT ["probe"]
   
   # Run with network capabilities
   docker run --cap-add=NET_RAW mesh-probe
   ```

### macOS Configuration

macOS uses the BPF (Berkeley Packet Filter) system with microsecond precision.

#### Configuration Example

```json
{
  "network": {
    "source_ip": "0.0.0.0",
    "interface": "auto",
    "ttl": 64,
    "buffer_size": 4096,
    "bind_port": 0,
    "qos": "normal",
    "dscp": 0
  }
}
```

#### macOS-Specific Behavior

- **Raw Sockets**: ❌ Disabled
- **Network API**: BPF (Berkeley Packet Filter)
- **Precision**: Microsecond
- **Timeout**: 10 seconds
- **Buffer Size**: 4096 bytes
- **Requirements**: System permissions

#### Setting up macOS Permissions

1. **System Preferences**:
   - Go to System Preferences → Security & Privacy
   - Allow the application to run

2. **Command Line**:
   ```bash
   # Remove quarantine attribute
   xattr -rd com.apple.quarantine /path/to/probe
   
   # Run probe
   ./probe start --config config.json
   ```

## Architecture Optimizations

The system automatically applies architecture-specific optimizations for maximum performance.

### x86_64 (AMD64) Optimizations

```json
{
  "optimizations": {
    "use_rdtsc": true,
    "packet_alignment": 64,
    "batch_size": 256,
    "vectorization": "avx2",
    "cache_friendly_allocation": true,
    "prefer_local_memory": true
  }
}
```

**Features Detected**:
- SSE2, SSE4.1, SSE4.2 support
- AVX, AVX2, AVX512 vectorization
- AES-NI encryption acceleration
- RDRAND random number generation

### ARM64 Optimizations

```json
{
  "optimizations": {
    "use_arm_counter": true,
    "packet_alignment": 64,
    "batch_size": 128,
    "vectorization": "neon",
    "cache_friendly_allocation": true,
    "prefer_local_memory": true
  }
}
```

**Features Detected**:
- NEON SIMD instructions
- AES acceleration
- SHA256 hardware acceleration
- CRC32 checksum acceleration

### ARM32 Optimizations

```json
{
  "optimizations": {
    "packet_alignment": 32,
    "batch_size": 64,
    "vectorization": "arm_neon",
    "cache_friendly_allocation": true,
    "prefer_local_memory": true
  }
}
```

## Examples

### Complete Configuration Examples

#### Linux Server Configuration

```json
{
  "id": "production_probe",
  "name": "Production Probe Server",
  "version": 1,
  
  "network": {
    "source_ip": "0.0.0.0",
    "interface": "eth0",
    "ttl": 64,
    "buffer_size": 8192,
    "bind_port": 0
  },
  
  "targets": [
    {
      "id": "gateway",
      "display_name": "Network Gateway",
      "address": "192.168.1.1",
      "enabled": true,
      "priority": 10,
      "timeout": "10s",
      "interval": "5s",
      "expected_rtt": "1ms"
    },
    {
      "id": "dns_server",
      "display_name": "Primary DNS",
      "address": "8.8.8.8",
      "enabled": true,
      "priority": 9,
      "timeout": "5s",
      "interval": "10s"
    }
  ],
  
  "telemetry": {
    "otlp_endpoint": "localhost:4317",
    "enable_tracing": true,
    "enable_metrics": true,
    "log_level": "info",
    "export_interval": "30s"
  },
  
  "mesh": {
    "enabled": true,
    "discovery_method": "multicast",
    "heartbeat_interval": "30s",
    "max_neighbors": 10
  }
}
```

#### Windows Workstation Configuration

```json
{
  "id": "workstation_probe",
  "name": "Workstation Probe",
  "version": 1,
  
  "network": {
    "source_ip": "0.0.0.0",
    "interface": "auto",
    "ttl": 64,
    "buffer_size": 2048,
    "bind_port": 0
  },
  
  "targets": [
    {
      "id": "router",
      "display_name": "Home Router",
      "address": "192.168.1.1",
      "enabled": true,
      "priority": 10,
      "timeout": "5s",
      "interval": "2s"
    },
    {
      "id": "isp_gateway",
      "display_name": "ISP Gateway",
      "address": "1.1.1.1",
      "enabled": true,
      "priority": 8,
      "timeout": "8s",
      "interval": "5s"
    }
  ],
  
  "telemetry": {
    "otlp_endpoint": "localhost:4317",
    "enable_tracing": true,
    "enable_metrics": true,
    "log_level": "debug",
    "export_interval": "60s"
  }
}
```

#### ARM64 Embedded Configuration

```json
{
  "id": "embedded_probe",
  "name": "Embedded ARM64 Probe",
  "version": 1,
  
  "network": {
    "source_ip": "0.0.0.0",
    "interface": "eth0",
    "ttl": 64,
    "buffer_size": 2048,
    "bind_port": 0
  },
  
  "targets": [
    {
      "id": "sensor_gateway",
      "display_name": "Sensor Network Gateway",
      "address": "10.0.1.1",
      "enabled": true,
      "priority": 10,
      "timeout": "3s",
      "interval": "1s"
    }
  ],
  
  "telemetry": {
    "otlp_endpoint": "192.168.1.100:4317",
    "enable_tracing": false,
    "enable_metrics": true,
    "log_level": "warn",
    "export_interval": "300s"
  },
  
  "mesh": {
    "enabled": false,
    "discovery_method": "manual",
    "heartbeat_interval": "60s"
  }
}
```

### Command Line Examples

#### Starting the Probe

```bash
# Linux/Unix
./probe start --config /etc/mesh-probe/config.json

# Windows (run as administrator)
probe.exe start --config C:\ProgramData\mesh-probe\config.json

# With custom log level
./probe start --config config.json --log-level debug

# With custom data directory
./probe start --config config.json --data-dir /var/lib/mesh-probe
```

#### Platform Detection Testing

```bash
# Test platform detection
./probe --test-platform

# Output shows:
# OS: linux
# Architecture: amd64
# Capabilities: [icmp udp tcp raw_sockets sse2 avx2 aes]
# Precision: nanosecond
# Optimizations: {use_rdtsc: true, batch_size: 256}
```

#### Configuration Validation

```bash
# Validate configuration
./probe validate-config --config config.json

# Test specific platform compatibility
./probe test-platform --config config.json --platform linux/amd64
```

## Validation

The system provides comprehensive validation to ensure configuration compatibility:

### Platform Compatibility Check

```bash
./probe validate-platform
```

**Output Example**:
```
✅ Platform: windows/amd64
✅ CPU Features: SSE2, AVX2, AES, RDRAND
✅ Network Support: Winsock, ICMP, UDP, TCP
⚠️  Requirements: Administrator privileges required for ICMP
✅ Optimizations: AVX2 vectorization, 256 batch size
```

### Configuration Validation

```json
{
  "validation_results": {
    "platform_compatible": true,
    "permissions_sufficient": false,
    "configuration_valid": true,
    "warnings": [
      "Windows requires administrator privileges for ICMP operations",
      "Raw sockets not available on this platform, using Winsock"
    ]
  }
}
```

### Network Interface Detection

```bash
# List available network interfaces
./probe list-interfaces

# Output:
# eth0     192.168.1.10  UP    ✅ Suitable
# lo       127.0.0.1     UP    ✅ Suitable  
# wlan0    192.168.1.15  UP    ✅ Suitable
# docker0  172.17.0.1    UP    ⚠️  Container bridge
```

## Troubleshooting

### Common Issues and Solutions

#### Windows: "Permission Denied" for ICMP

**Problem**: ICMP operations fail with permission errors.

**Solution**:
1. Run Command Prompt as Administrator
2. Or configure Windows Defender Firewall to allow ICMP
3. Use `netsh advfirewall firewall` commands:

```cmd
# Allow ICMP echo requests
netsh advfirewall firewall add rule name="Mesh Probe ICMP" dir=in action=allow protocol=icmpv4:8,any

# Allow probe ports
netsh advfirewall firewall add rule name="Mesh Probe UDP" dir=in action=allow protocol=udp localport=8080
```

#### Linux: "Operation Not Permitted" for Raw Sockets

**Problem**: Raw socket creation fails.

**Solution**:
1. Run with proper capabilities:
   ```bash
   sudo setcap cap_net_raw+ep /path/to/probe
   ./probe start --config config.json
   ```

2. Or run as root:
   ```bash
   sudo ./probe start --config config.json
   ```

3. Check capabilities:
   ```bash
   getcap /path/to/probe
   ```

#### macOS: "Sandbox" Restrictions

**Problem**: macOS security restrictions prevent network access.

**Solution**:
1. Remove quarantine attribute:
   ```bash
   xattr -rd com.apple.quarantine /path/to/probe
   ```

2. Allow in System Preferences → Security & Privacy

#### ARM64: Performance Issues

**Problem**: Lower than expected performance on ARM64.

**Solution**:
1. Verify NEON support:
   ```bash
   grep -i neon /proc/cpuinfo
   ```

2. Check architecture detection:
   ```bash
   ./probe --test-platform
   ```

3. Ensure proper memory alignment:
   ```json
   {
     "optimizations": {
       "packet_alignment": 64,
       "cache_friendly_allocation": true
     }
   }
   ```

### Debug Mode

Enable detailed logging for troubleshooting:

```json
{
  "telemetry": {
    "log_level": "debug",
    "enable_tracing": true,
    "enable_metrics": true
  }
}
```

```bash
# Run with debug output
./probe start --config config.json --log-level debug

# Follow real-time logs
./probe logs --follow
```

### Performance Monitoring

Monitor platform-specific performance:

```bash
# Check timing precision
./probe test-timing

# Monitor CPU usage
top -p $(pgrep probe)

# Check network interface statistics
netstat -i
# or on Linux:
cat /proc/net/dev
```

### Getting Help

For additional support:

1. Check system logs:
   ```bash
   # Linux/macOS
   journalctl -u mesh-probe
   
   # Windows
   eventvwr.msc
   ```

2. Run diagnostic tools:
   ```bash
   ./probe diagnose --full-report
   ```

3. Review configuration validation:
   ```bash
   ./probe validate-config --verbose --config config.json
   ```

---

This comprehensive guide ensures successful deployment and configuration of the Mesh Net Probe across all supported platforms and architectures.