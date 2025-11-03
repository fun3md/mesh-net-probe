# Mesh Probe System - Complete Deployment Guide

## Overview

The Mesh Probe System is a comprehensive, enterprise-grade network monitoring solution that provides high-precision ICMP measurements across multiple platforms with centralized configuration management and real-time web-based administration.

## System Architecture

### Core Components

1. **Probe Engine** (`cmd/probe/`): Core ICMP measurement engine
2. **Admin Web Interface** (`cmd/admin-web/`): Web-based management dashboard
3. **Configuration Management**: etcd/Consul-based centralized configuration
4. **Cross-Platform Support**: Linux, macOS, Windows (x64/ARM64)
5. **Real-time Monitoring**: WebSocket-based live data streaming

### Architecture Diagram

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Admin Web     │    │   Configuration  │    │   Probe Fleet   │
│   Interface     │◄──►│   Management     │◄──►│   (100+ nodes)  │
│                 │    │   (etcd/Consul)  │    │                 │
│ • Dashboard     │    │                  │    │ • ICMP Engine   │
│ • Configuration │    │ • SLA tracking   │    │ • Health Monitor│
│ • Real-time     │    │ • Conflict res.  │    │ • Auto-register │
│   Monitoring    │    │ • Auth & audit   │    │ • Cross-platform│
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Prerequisites

### System Requirements

- **Operating Systems**: Linux, macOS, Windows
- **Architecture**: x64, ARM64
- **Memory**: 256MB RAM minimum per probe
- **Network**: ICMP (ping) access to target networks
- **Go Version**: 1.21 or later

### Dependencies

- **etcd** (optional): For centralized configuration
- **Consul** (optional): Alternative configuration backend
- **Docker**: For containerized deployment
- **Kubernetes**: For orchestration (optional)

## Installation Methods

### 1. Docker Deployment (Recommended)

#### Multi-Platform Build

```bash
# Build for all platforms
docker buildx build --platform linux/amd64,linux/arm64 -t mesh-probe:latest .

# Or use pre-built images
docker pull mesh-probe:latest
```

#### Docker Compose (Single Probe)

```yaml
version: '3.8'
services:
  probe:
    image: mesh-probe:latest
    container_name: mesh-probe
    network_mode: host
    environment:
      - PROBE_TARGETS=8.8.8.8,1.1.1.1
      - PROBE_INTERVAL=30s
      - LOG_LEVEL=info
    volumes:
      - ./config.json:/app/config.json:ro
      - probe-data:/app/data
    restart: unless-stopped

volumes:
  probe-data:
```

#### Docker Compose (Full Stack)

```yaml
version: '3.8'
services:
  # etcd for configuration management
  etcd:
    image: quay.io/coreos/etcd:v3.5.10
    container_name: etcd
    environment:
      - ETCD_AUTO_COMPACTION_MODE=revision
      - ETCD_AUTO_COMPACTION_RETENTION=1000
      - ETCD_QUOTA_BACKEND_BYTES=4294967296
    volumes:
      - etcd-data:/etcd
    command: [
      "etcd",
      "--data-dir=/etcd",
      "--listen-client-urls=http://0.0.0.0:2379",
      "--advertise-client-urls=http://localhost:2379",
      "--max-txn-ops=128",
      "--max-request-bytes=1048576"
    ]

  # Admin Web Interface
  admin-web:
    image: mesh-probe-web:latest
    container_name: mesh-probe-web
    ports:
      - "8080:8080"
    environment:
      - ETCD_ENDPOINTS=http://etcd:2379
      - JWT_SECRET=your-jwt-secret
    depends_on:
      - etcd
    restart: unless-stopped

  # Mesh Probe
  probe:
    image: mesh-probe:latest
    container_name: mesh-probe
    network_mode: host
    environment:
      - CONFIG_SOURCE=etcd
      - ETCD_ENDPOINTS=http://etcd:2379
      - PROBE_INTERVAL=30s
      - LOG_LEVEL=info
    depends_on:
      - etcd
    restart: unless-stopped

volumes:
  etcd-data:
```

### 2. Native Installation

#### Build from Source

```bash
# Clone repository
git clone https://github.com/mesh-net-probe/probe.git
cd probe

# Build for current platform
go build -o bin/probe ./cmd/probe
go build -o bin/admin-web ./cmd/admin-web

# Install system-wide
sudo cp bin/probe /usr/local/bin/
sudo cp bin/admin-web /usr/local/bin/
```

#### Package Installation

```bash
# Debian/Ubuntu
dpkg -i mesh-probe_1.0.0_amd64.deb

# RHEL/CentOS
rpm -ivh mesh-probe-1.0.0.x86_64.rpm

# macOS (Homebrew)
brew install mesh-probe
```

### 3. Kubernetes Deployment

#### Namespace and Configuration

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: mesh-probe

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: probe-config
  namespace: mesh-probe
data:
  config.json: |
    {
      "targets": [
        "8.8.8.8",
        "1.1.1.1"
      ],
      "interval": "30s",
      "count": 5,
      "timeout": "5s"
    }
```

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mesh-probe
  namespace: mesh-probe
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mesh-probe
  template:
    metadata:
      labels:
        app: mesh-probe
    spec:
      containers:
      - name: probe
        image: mesh-probe:latest
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        env:
        - name: PROBE_CONFIG_PATH
          value: "/app/config.json"
        volumeMounts:
        - name: config
          mountPath: /app/config.json
          subPath: config.json
        - name: probe-data
          mountPath: /app/data
      volumes:
      - name: config
        configMap:
          name: probe-config
      - name: probe-data
        emptyDir: {}
      hostNetwork: true
```

#### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: mesh-probe-service
  namespace: mesh-probe
spec:
  selector:
    app: mesh-probe
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

## Configuration

### Configuration File Format

```json
{
  "targets": [
    {
      "id": "google-dns",
      "display_name": "Google DNS",
      "address": "8.8.8.8",
      "port": 0,
      "enabled": true,
      "priority": 5,
      "timeout": "5s",
      "count": 5,
      "interval": "1s"
    },
    {
      "id": "cloudflare-dns",
      "display_name": "Cloudflare DNS", 
      "address": "1.1.1.1",
      "port": 0,
      "enabled": true,
      "priority": 5,
      "timeout": "5s",
      "count": 5,
      "interval": "1s"
    }
  ],
  "measurement": {
    "timeout": "5s",
    "buffer_size": 65535,
    "ttl": 64,
    "dscp": 0
  },
  "output": {
    "format": "json",
    "file_path": "/app/data/measurements.json",
    "max_size": "100MB",
    "max_age": 7
  },
  "telemetry": {
    "enabled": true,
    "endpoint": "http://localhost:4317",
    "service_name": "mesh-probe"
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  }
}
```

### Environment Variables

```bash
# Probe Configuration
export PROBE_TARGETS="8.8.8.8,1.1.1.1"
export PROBE_INTERVAL="30s"
export PROBE_COUNT="5"
export PROBE_TIMEOUT="5s"

# Configuration Source
export CONFIG_SOURCE="file"              # file, etcd, consul
export ETCD_ENDPOINTS="localhost:2379"
export CONSUL_ENDPOINTS="localhost:8500"

# Output Configuration  
export OUTPUT_FORMAT="json"              # json, text
export OUTPUT_FILE="/app/data/out.json"
export MAX_OUTPUT_SIZE="100MB"

# Telemetry
export TELEMETRY_ENABLED="true"
export OTEL_ENDPOINT="http://localhost:4317"
export OTEL_SERVICE_NAME="mesh-probe"

# Logging
export LOG_LEVEL="info"                  # debug, info, warn, error
export LOG_FORMAT="json"                 # json, text
export LOG_OUTPUT="stdout"               # stdout, stderr, file

# Admin Web Interface
export WEB_PORT="8080"
export WEB_HOST="0.0.0.0"
export JWT_SECRET="your-secret-key"
export AUTH_ENABLED="true"
```

### Centralized Configuration (etcd)

```bash
# Set configuration in etcd
etcdctl put /mesh-probe/config '{"targets": ["8.8.8.8"], "interval": "30s"}'

# Probe configuration
export CONFIG_SOURCE="etcd"
export ETCD_ENDPOINTS="etcd1:2379,etcd2:2379,etcd3:2379"
```

### Centralized Configuration (Consul)

```bash
# Set configuration in Consul
consul kv put mesh-probe/config '{"targets": ["8.8.8.8"], "interval": "30s"}'

# Probe configuration
export CONFIG_SOURCE="consul"
export CONSUL_ENDPOINTS="consul1:8500,consul2:8500"
```

## Usage

### Command Line Interface

#### Basic Probe Operation

```bash
# Simple ping measurement
probe ping 8.8.8.8

# Multiple targets with custom settings
probe measure \
  --targets 8.8.8.8,1.1.1.1 \
  --count 10 \
  --interval 1s \
  --timeout 5s

# Output to file
probe measure \
  --config /etc/probe/config.json \
  --output /var/log/probe/measurements.json \
  --format json
```

#### Advanced Features

```bash
# Average multiple measurements
probe measure \
  --targets google.com \
  --average 10 \
  --count 100 \
  --output-file measurements.json

# Continuous monitoring
probe measure \
  --targets 8.8.8.8 \
  --continuous \
  --interval 30s \
  --duration 1h

# JSON output with statistics
probe measure \
  --targets 8.8.8.8 \
  --count 50 \
  --format json \
  --output measurements.json
```

#### Configuration Management

```bash
# Generate sample configuration
probe config init --output /etc/probe/config.json

# Validate configuration
probe config validate --config /etc/probe/config.json

# Sync configuration from etcd
probe config sync --source etcd --endpoints localhost:2379
```

### Admin Web Interface

#### Starting the Web Interface

```bash
# Start with default settings
admin-web start

# Custom configuration
admin-web start \
  --port 8080 \
  --host 0.0.0.0 \
  --etcd-endpoints localhost:2379 \
  --jwt-secret "your-secret"

# Background mode
admin-web start --daemon --pid-file /var/run/admin-web.pid
```

#### Web Interface Features

1. **Dashboard**: Real-time monitoring of all probes
2. **Configuration Management**: Create, edit, and deploy configurations
3. **Probe Management**: Register, monitor, and manage probe fleet
4. **Measurement Visualization**: Live charts and historical data
5. **Alert Management**: Real-time alerts and notifications
6. **User Management**: Role-based access control

#### API Endpoints

```bash
# Health check
curl http://localhost:8080/api/health

# Get probe status
curl http://localhost:8080/api/probes

# Get measurements
curl http://localhost:8080/api/measurements?limit=100

# Create configuration
curl -X POST http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{"targets": ["8.8.8.8"], "interval": "30s"}'

# WebSocket for real-time updates
ws://localhost:8080/ws
```

### Systemd Service Installation

#### Probe Service

```ini
# /etc/systemd/system/mesh-probe.service
[Unit]
Description=Mesh Probe Network Monitoring
After=network.target

[Service]
Type=simple
User=probe
Group=probe
ExecStart=/usr/local/bin/probe measure --config /etc/probe/config.json
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ReadWritePaths=/var/lib/probe /var/log/probe

[Install]
WantedBy=multi-user.target
```

#### Admin Web Service

```ini
# /etc/systemd/system/mesh-probe-web.service
[Unit]
Description=Mesh Probe Admin Web Interface
After=network.target etcd.service

[Service]
Type=simple
User=web
Group=web
ExecStart=/usr/local/bin/admin-web start \
  --port 8080 \
  --etcd-endpoints localhost:2379
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start services:

```bash
sudo systemctl daemon-reload
sudo systemctl enable mesh-probe
sudo systemctl start mesh-probe
sudo systemctl enable mesh-probe-web
sudo systemctl start mesh-probe-web
```

## Monitoring and Metrics

### OpenTelemetry Integration

```yaml
# otel-collector-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:

exporters:
  prometheus:
    endpoint: "0.0.0.0:8889"

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheus]
```

### Prometheus Metrics

```bash
# Scrape configuration for Prometheus
scrape_configs:
  - job_name: 'mesh-probe'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/api/metrics'
    scrape_interval: 30s
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Mesh Probe Network Monitoring",
    "panels": [
      {
        "title": "Network Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "probe_rtt_seconds",
            "legendFormat": "{{target}}"
          }
        ]
      },
      {
        "title": "Probe Health",
        "type": "stat",
        "targets": [
          {
            "expr": "probe_up",
            "legendFormat": "Probes Online"
          }
        ]
      }
    ]
  }
}
```

## Troubleshooting

### Common Issues

#### ICMP Permission Errors

```bash
# Linux - Check capabilities
sudo capsh --print | grep cap_net_raw

# Grant capabilities
sudo setcap cap_net_raw+ep /usr/local/bin/probe

# Alternative - run as root (not recommended for production)
sudo probe ping 8.8.8.8
```

#### Configuration Loading Errors

```bash
# Validate configuration syntax
probe config validate --config config.json

# Check etcd connectivity
etcdctl endpoint health

# Check Consul connectivity
consul kv get mesh-probe/config
```

#### Network Connectivity Issues

```bash
# Test ICMP manually
ping -c 3 8.8.8.8

# Check firewall rules
sudo iptables -L

# Test network reachability
probe ping --verbose 8.8.8.8
```

#### Performance Issues

```bash
# Enable debug logging
export LOG_LEVEL=debug
probe measure --targets 8.8.8.8

# Check system resources
top
htop
free -h
iostat -x 1

# Profile application
go tool pprof http://localhost:6060/debug/pprof/profile
```

### Logging and Debugging

#### Enable Debug Logging

```bash
# Environment variable
export LOG_LEVEL=debug

# Command line
probe --log-level debug measure 8.8.8.8

# Configuration file
{
  "logging": {
    "level": "debug",
    "format": "json"
  }
}
```

#### Structured Logging

```bash
# View logs with jq
journalctl -u mesh-probe -f | jq .

# Export to file
journalctl -u mesh-probe > probe.log
```

### Health Checks

#### Probe Health

```bash
# Check probe status
probe status

# Test probe functionality
probe ping --self-test

# Check configuration
probe config show
```

#### Web Interface Health

```bash
# Health check
curl http://localhost:8080/api/health

# Detailed status
curl http://localhost:8080/api/status

# Check connected probes
curl http://localhost:8080/api/probes
```

## Security Considerations

### Network Security

1. **ICMP Restrictions**: Ensure ICMP traffic is allowed between probes and targets
2. **Firewall Rules**: Configure appropriate firewall rules for probe communication
3. **Network Segmentation**: Use private networks where possible

### Authentication

1. **JWT Tokens**: Use strong JWT secrets for web interface authentication
2. **API Keys**: Implement API key authentication for configuration management
3. **TLS**: Enable TLS for all web interface communications

### Data Protection

1. **Sensitive Data**: Avoid logging sensitive configuration data
2. **Secure Storage**: Use secure storage for configuration files
3. **Access Control**: Implement proper access controls for probe fleet

### Compliance

1. **Audit Logging**: Enable comprehensive audit logging
2. **Data Retention**: Configure appropriate data retention policies
3. **Privacy**: Ensure compliance with relevant data privacy regulations

## Performance Tuning

### System Optimization

```bash
# Increase file descriptor limits
ulimit -n 65536

# Optimize network settings
echo 1 > /proc/sys/net/ipv4/ip_forward
echo 0 > /proc/sys/net/ipv4/conf/all/rp_filter
```

### Probe Optimization

```json
{
  "measurement": {
    "timeout": "5s",
    "buffer_size": 65535,
    "ttl": 64,
    "dscp": 0
  },
  "output": {
    "buffer_size": 8192,
    "flush_interval": "1s"
  }
}
```

### Scalability

1. **Horizontal Scaling**: Deploy multiple probe instances
2. **Load Balancing**: Use load balancers for web interface
3. **Database Optimization**: Optimize measurement data storage
4. **Caching**: Implement caching for configuration data

## Maintenance

### Backup Procedures

```bash
# Backup configuration
tar -czf probe-config-backup.tar.gz /etc/probe/

# Backup etcd data
etcdctl snapshot save probe-etcd-snapshot.db

# Backup Consul data
consul snapshot save probe-consul-snapshot.snap
```

### Update Procedures

```bash
# Update probe
sudo systemctl stop mesh-probe
docker pull mesh-probe:latest
sudo systemctl start mesh-probe

# Update configuration
probe config update --file new-config.json
```

### Monitoring Health

```bash
# System health check
probe health-check --verbose

# Configuration validation
probe config validate --comprehensive

# Performance check
probe benchmark --duration 30s
```

## Support and Resources

### Documentation

- **API Documentation**: http://localhost:8080/api/docs
- **Configuration Reference**: `/docs/configuration.md`
- **Troubleshooting Guide**: `/docs/troubleshooting.md`

### Getting Help

1. **GitHub Issues**: https://github.com/mesh-net-probe/probe/issues
2. **Documentation**: https://docs.mesh-probe.org
3. **Community Forum**: https://community.mesh-probe.org

### Contributing

1. **Development Setup**: See `CONTRIBUTING.md`
2. **Code Style**: Follow Go community standards
3. **Testing**: Maintain 80%+ test coverage

---

**Version**: 1.0.0  
**Last Updated**: 2025-11-03  
**Support**: community@mesh-probe.org