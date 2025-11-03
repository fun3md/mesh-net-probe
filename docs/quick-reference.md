# Quick Reference Guide

This quick reference provides essential configuration patterns and common tasks for the Mesh Net Probe.

## Common Configuration Patterns

### Basic ICMP Monitoring

```json
{
  "network": {
    "ttl": 64,
    "buffer_size": 4096
  },
  "targets": [
    {
      "id": "gateway",
      "address": "192.168.1.1",
      "interval": "5s",
      "timeout": "10s"
    }
  ]
}
```

### High-Frequency Monitoring

```json
{
  "network": {
    "ttl": 64,
    "buffer_size": 8192
  },
  "targets": [
    {
      "id": "server1",
      "address": "10.0.1.10",
      "interval": "1s",
      "timeout": "3s",
      "priority": 10
    }
  ],
  "telemetry": {
    "export_interval": "10s",
    "enable_metrics": true
  }
}
```

### Multi-Target Monitoring

```json
{
  "network": {
    "ttl": 64,
    "buffer_size": 4096
  },
  "targets": [
    {
      "id": "primary_dns",
      "address": "8.8.8.8",
      "interval": "30s",
      "priority": 9,
      "labels": {"type": "dns"}
    },
    {
      "id": "backup_dns", 
      "address": "1.1.1.1",
      "interval": "30s",
      "priority": 8,
      "labels": {"type": "dns"}
    },
    {
      "id": "gateway",
      "address": "192.168.1.1",
      "interval": "5s",
      "priority": 10,
      "labels": {"type": "gateway"}
    }
  ]
}
```

## Platform-Specific Commands

### Linux Commands

```bash
# Start probe with configuration
sudo ./probe start --config /etc/mesh-probe/config.json

# Set capabilities for non-root operation
sudo setcap cap_net_raw+ep /usr/local/bin/probe

# Check permissions
getcap /usr/local/bin/probe

# Run as service
sudo systemctl enable mesh-probe
sudo systemctl start mesh-probe

# View logs
journalctl -u mesh-probe -f

# Test platform detection
./probe --test-platform
```

### Windows Commands

```cmd
REM Run as Administrator
probe.exe start --config C:\ProgramData\mesh-probe\config.json

REM Check Windows Defender Firewall
netsh advfirewall firewall show rule name="Mesh Probe ICMP"

REM Allow ICMP through firewall
netsh advfirewall firewall add rule name="Mesh Probe ICMP" dir=in action=allow protocol=icmpv4:8,any

REM Check service status
sc query MeshProbe

REM Start as Windows service
sc create MeshProbe binPath= "C:\Program Files\MeshProbe\probe.exe start"
sc start MeshProbe
```

### macOS Commands

```bash
# Remove quarantine attribute
xattr -rd com.apple.quarantine /Applications/mesh-probe.app/Contents/MacOS/probe

# Run probe
/Applications/mesh-probe.app/Contents/MacOS/probe start --config ~/config.json

# Check permissions
ls -la /Applications/mesh-probe.app/

# View system logs
log show --predicate 'process == "probe"' --last 1h
```

## Docker Deployment

### Basic Docker Configuration

```dockerfile
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy probe binary
COPY probe /usr/local/bin/probe
RUN chmod +x /usr/local/bin/probe

# Set working directory
WORKDIR /app

# Create configuration directory
RUN mkdir -p /app/config

# Copy configuration
COPY config.json /app/config/

# Add capabilities
RUN setcap cap_net_raw+ep /usr/local/bin/probe

# Expose ports
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["probe"]
CMD ["start", "--config", "/app/config/config.json"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  mesh-probe:
    build: .
    container_name: mesh-probe
    restart: unless-stopped
    network_mode: host
    cap_add:
      - NET_RAW
    volumes:
      - ./config:/app/config:ro
      - ./data:/app/data
    environment:
      - PROBE_LOG_LEVEL=info
    command: start --config /app/config/config.json

  otel-collector:
    image: otel/opentelemetry-collector:latest
    container_name: otel-collector
    restart: unless-stopped
    ports:
      - "4317:4317"
      - "4318:4318"
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-config.yaml:/etc/otel-collector-config.yaml:ro
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: mesh-probe
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: mesh-probe
  template:
    metadata:
      labels:
        app: mesh-probe
    spec:
      hostNetwork: true
      containers:
      - name: mesh-probe
        image: mesh-probe:latest
        args:
        - start
        - --config
        - /etc/probe/config.json
        securityContext:
          capabilities:
            add:
            - NET_RAW
        volumeMounts:
        - name: config
          mountPath: /etc/probe
          readOnly: true
        - name: data
          mountPath: /app/data
        resources:
          limits:
            cpu: 100m
            memory: 128Mi
          requests:
            cpu: 50m
            memory: 64Mi
      volumes:
      - name: config
        configMap:
          name: mesh-probe-config
      - name: data
        emptyDir: {}
```

## Monitoring Commands

### Real-time Status

```bash
# View probe status
./probe status

# Monitor measurements in real-time
./probe measurements --follow

# Check network connectivity
./probe ping --target 8.8.8.8 --count 5

# View platform capabilities
./probe platform-info
```

### Performance Monitoring

```bash
# Check timing precision
./probe timing-test

# Monitor resource usage
./probe metrics --format prometheus

# Check measurement statistics
./probe stats

# Validate configuration
./probe validate-config --verbose
```

### Health Checks

```bash
# System health check
./probe health --detailed

# Network interface status
./probe interfaces

# Configuration validation
./probe check-config --config config.json

# Platform compatibility
./probe platform-check
```

## Troubleshooting Commands

### Linux Troubleshooting

```bash
# Check capabilities
getcap $(which probe)

# Test raw socket creation
sudo tcpdump -i eth0 icmp

# Check network interfaces
ip addr show

# Monitor system resources
top -p $(pgrep probe)

# Check kernel network parameters
sysctl net.ipv4.icmp_echo_ignore_broadcasts
sysctl net.ipv4.icmp_ignore_bogus_error_responses
```

### Windows Troubleshooting

```cmd
REM Check firewall rules
netsh advfirewall firewall show rule name="Mesh Probe ICMP"

REM Test ICMP manually
ping 8.8.8.8

REM Check network adapters
ipconfig /all

REM Monitor network traffic
netstat -s

REM Check Windows service status
sc query MeshProbe
sc queryex MeshProbe
```

### macOS Troubleshooting

```bash
# Check network interfaces
ifconfig

# Monitor network traffic
sudo lsof -i -n -P

# Check system permissions
ls -la /Applications/mesh-probe.app/

# View application logs
log show --predicate 'process == "probe"' --last 1h
```

## Configuration Validation

### Quick Validation Commands

```bash
# Validate basic configuration
./probe validate-config --config config.json

# Test platform compatibility
./probe test-platform --config config.json

# Check network configuration
./probe check-network

# Validate targets
./probe validate-targets --config config.json
```

### Common Validation Issues

| Issue | Command | Solution |
|-------|---------|----------|
| Invalid JSON | `./probe validate-config` | Use `jq` to validate JSON syntax |
| Missing permissions | `./probe test-platform` | Run with appropriate privileges |
| Invalid IP address | `./probe validate-targets` | Check IP address format |
| Port binding error | `./probe check-network` | Ensure port is available |
| Buffer size too small | `./probe check-config` | Increase buffer size for high frequency |

## Performance Tuning

### High-Frequency Monitoring (Linux)

```json
{
  "network": {
    "buffer_size": 16384,
    "ttl": 64
  },
  "telemetry": {
    "export_interval": "5s",
    "buffer_size": 5000
  },
  "optimizations": {
    "batch_size": 512,
    "packet_alignment": 64
  }
}
```

### Low-Resource Environment

```json
{
  "network": {
    "buffer_size": 2048,
    "ttl": 64
  },
  "telemetry": {
    "export_interval": "300s",
    "enable_logging": false
  },
  "optimizations": {
    "batch_size": 32
  }
}
```

### Container Environment

```json
{
  "network": {
    "interface": "eth0",
    "buffer_size": 4096
  },
  "telemetry": {
    "enable_logging": true,
    "log_level": "info"
  },
  "mesh": {
    "enabled": false
  }
}
```

---

For detailed information, see the [Cross-Platform Configuration Guide](cross-platform-configuration.md) and [Troubleshooting Guide](troubleshooting.md).