# Mesh Probe System

A high-precision ICMP measurement probe with statistical averaging and continuous monitoring capabilities, designed for cross-platform operation with Windows, Linux, and macOS support.

## 🚀 Features

- **Single Measurements**: Precise ICMP ping measurements with microsecond precision
- **Statistical Averaging**: Average multiple measurements with min, max, and standard deviation
- **Continuous Monitoring**: Real-time monitoring with configurable intervals
- **Network Path Tracing**: Complete traceroute with configurable hops and DNS control
- **Cross-Platform**: Automatic adaptation for Windows, Linux, and macOS
- **Multiple Output Formats**: Both human-readable text and structured JSON
- **Configuration-Driven**: JSON-based configuration with CLI override support

## 🎯 Quick Start

### Basic Usage

```bash
# Single ping measurement
./probe.exe ping 8.8.8.8

# Average 10 measurements
./probe.exe ping 8.8.8.8 -n 10

# Continuous monitoring (default 1-second intervals)
./probe.exe ping 8.8.8.8 --continuous

# Custom interval (5 seconds)
./probe.exe ping 8.8.8.8 --continuous --interval 5s

# JSON output
./probe.exe ping 8.8.8.8 -n 5 -f json

# Traceroute to a target
./probe.exe traceroute 8.8.8.8

# Traceroute with limited hops
./probe.exe traceroute 8.8.8.8 -m 10

# Traceroute without DNS resolution
./probe.exe traceroute 8.8.8.8 --no-dns
```

### Configuration File

Create `config.json`:
```json
{
  "name": "Network Monitor",
  "version": 1,
  "network": {
    "buffer_size": 2048,
    "ttl": 64
  },
  "targets": [
    {
      "id": "google_dns",
      "address": "8.8.8.8",
      "enabled": true,
      "timeout": 5
    },
    {
      "id": "cloudflare_dns",
      "address": "1.1.1.1",
      "enabled": true,
      "timeout": 5
    }
  ],
  "telemetry": {
    "enable_metrics": true,
    "enable_tracing": true,
    "enable_logging": true,
    "log_level": "info",
    "log_format": "text"
  }
}
```

### Using Configuration File

```bash
# Use targets from config
./probe.exe -c config.json ping

# Average 3 measurements for config targets
./probe.exe -c config.json ping -n 3

# Continuous monitoring of config targets
./probe.exe -c config.json ping --continuous

# Override interval
./probe.exe -c config.json ping --continuous --interval 30s

# Traceroute targets from config
./probe.exe -c config.json traceroute

# Traceroute with limited hops
./probe.exe -c config.json traceroute -m 15
```

## 📖 Command Line Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-c, --config` | Configuration file path | None |
| `-l, --log-level` | Log level (debug, info, warn, error) | info |
| `-f, --log-format` | Output format (text, json) | text |
| `-v, --verbose` | Enable verbose output | false |
| `-p, --probe-id` | Custom probe identifier | Auto-generated |

### Ping Control Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-n, --count` | Number of measurements for averaging | 1 |
| `--continuous` | Enable continuous measurement mode | false |
| `--interval` | Interval for continuous measurements | 1s |

### Traceroute Control Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-m, --max-hops` | Maximum number of hops | 30 |
| `--no-dns` | Disable DNS hostname resolution | false |

### Interval Formats

Supported duration formats:
- `ns`, `µs`, `ms`, `s`, `m`, `h`
- Examples: `500ms`, `1s`, `5m`, `2h`

## 📊 Output Examples

### Single Ping Measurement (Text)
```bash
$ ./probe.exe ping 8.8.8.8
SUCCESS: cmd_target_1 -> 8.8.8.8: 10.5154ms
```

### Averaged Ping Measurements (Text)
```bash
$ ./probe.exe ping 8.8.8.8 -n 5 -v
Using 1 targets from command line
Performing 5 measurements per target for averaging
Measurement round 1/5
Measurement round 2/5
Measurement round 3/5
Measurement round 4/5
Measurement round 5/5

=== Averaged Results (5 measurements) ===
cmd_target_1 -> 8.8.8.8:
  Average: 10.34202ms
  Min: 10.0007ms
  Max: 10.571ms
  StdDev: 285.12µs
  Success Rate: 100.0% (5/5)
```

### Ping Measurement (JSON)
```bash
$ ./probe.exe ping 8.8.8.8 -f json
{
  "results": [
    {
      "success": true,
      "target": "cmd_target_1",
      "address": "8.8.8.8",
      "rtt": "10ms"
    }
  ],
  "total_targets": 1,
  "success_count": 1,
  "failure_count": 0
}
```

### Averaged Ping Measurements (JSON)
```bash
$ ./probe.exe ping 8.8.8.8 -n 3 -f json
{
  "measurement_count": 3,
  "results": [
    {
      "target": "cmd_target_1",
      "address": "8.8.8.8",
      "average_rtt": "10.344366ms",
      "min_rtt": "10.0001ms",
      "max_rtt": "10.5192ms",
      "std_dev": "298.155µs",
      "success_rate": 1,
      "total_measurements": 3,
      "successful_count": 3,
      "failed_count": 0
    }
  ]
}
```

### Continuous Ping Mode
```bash
$ ./probe.exe ping 8.8.8.8 --continuous
Starting continuous measurements with 1s interval
Continuous measurement round 1
SUCCESS: cmd_target_1 -> 8.8.8.8: 10.007ms
Continuous measurement round 2
SUCCESS: cmd_target_1 -> 8.8.8.8: 10.1034ms
Continuous measurement round 3
SUCCESS: cmd_target_1 -> 8.8.8.8: 10.5172ms
...
```

### Traceroute Output
```bash
$ ./probe.exe traceroute 8.8.8.8
 1  192.168.1.1 (192.168.1.1)         1.234ms
 2  10.0.0.1 (10.0.0.1)               5.678ms
 3  8.8.8.8 (google-public-dns-a.google.com)  10.123ms
```

### Traceroute without DNS
```bash
$ ./probe.exe traceroute 8.8.8.8 --no-dns
 1  192.168.1.1                       1.234ms
 2  10.0.0.1                          5.678ms
 3  8.8.8.8                           10.123ms
```

## 🔧 Command Examples

### Single Target
```bash
# Basic ping
./probe.exe ping 8.8.8.8

# 10 measurements average
./probe.exe ping 8.8.8.8 -n 10

# JSON output
./probe.exe ping 8.8.8.8 -f json
```

### Multiple Targets
```bash
# Multiple CLI targets
./probe.exe ping 8.8.8.8 1.1.1.1 8.8.4.4

# Average 5 measurements
./probe.exe ping 8.8.8.8 1.1.1.1 -n 5

# Mixed configuration and CLI
./probe.exe -c config.json ping 1.1.1.1
```

### Continuous Monitoring
```bash
# Default 1-second intervals
./probe.exe ping 8.8.8.8 --continuous

# 5-second intervals
./probe.exe ping 8.8.8.8 --continuous --interval 5s

# Monitor config targets
./probe.exe -c config.json ping --continuous --interval 30s
```

### Traceroute Examples
```bash
# Basic traceroute
./probe.exe traceroute 8.8.8.8

# Traceroute with limited hops
./probe.exe traceroute 8.8.8.8 -m 10

# Traceroute without DNS resolution
./probe.exe traceroute 8.8.8.8 --no-dns

# Traceroute with verbose output
./probe.exe traceroute 8.8.8.8 -v
```

### Advanced Examples
```bash
# Verbose continuous mode with custom interval
./probe.exe -c config.json ping -v --continuous --interval 10s

# High-precision averaging (20 samples) with JSON output
./probe.exe ping 8.8.8.8 -n 20 -f json

# Debug logging
./probe.exe ping 8.8.8.8 -l debug -n 3

# Traceroute configuration targets
./probe.exe -c config.json traceroute

# Traceroute with custom hop limit for config
./probe.exe -c config.json traceroute -m 15
```

## 🏗️ Architecture

### Key Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Mesh Probe Application                   │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐  │
│  │   Platform      │  │  Configuration  │  │  Command     │  │
│  │   Detection     │  │    Parser       │  │   Parser     │  │
│  └─────────────────┘  └─────────────────┘  └──────────────┘  │
│           │                     │                    │       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐  │
│  │    ICMP         │  │   Timing        │  │  Output      │  │
│  │   Engine        │  │   Engine        │  │  Formatter   │  │
│  └─────────────────┘  └─────────────────┘  └──────────────┘  │
│           │                     │                    │       │
└─────────────────────────────────────────────────────────────┘
            │                     │                    │
     ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
     │   Raw/      │      │   High      │      │    Text/    │
     │  Winsock    │      │Precision    │      │    JSON     │
     │  Sockets    │      │   Timing    │      │  Output     │
     └─────────────┘      └─────────────┘      └─────────────┘
```

### Platform Support Matrix

| Platform | Architecture | ICMP Method | Timing Precision |
|----------|-------------|-------------|------------------|
| **Windows** | x86_64, ARM64 | Winsock Simulation | Millisecond |
| **Linux** | x86_64, ARM64 | Raw Sockets | Nanosecond |
| **macOS** | x86_64, ARM64 | BPF | Microsecond |

### Statistical Features

- **Sample Standard Deviation**: Unbiased sample formula (n-1 denominator)
- **Success Rate**: Percentage of successful vs total measurements
- **Range Analysis**: Min/max values across measurement series
- **Robust Statistics**: Handles mixed success/failure scenarios

## 🛠️ Development

### Building

```bash
# Build for current platform
go build -o probe ./cmd/probe/

# Cross-compile
GOOS=windows GOARCH=amd64 go build -o probe.exe ./cmd/probe/
GOOS=linux GOARCH=arm64 go build -o probe-linux-arm64 ./cmd/probe/
```

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Cross-platform tests
go test -race ./...
```

### Code Structure

```
cmd/probe/
├── main.go              # Command line interface and orchestration (ping + traceroute commands)

internal/
├── icmp/
│   └── engine.go        # ICMP measurement engine with simulation mode and traceroute support
├── platform/
│   └── detection.go     # Cross-platform detection and compatibility
├── config/
│   └── manager.go       # Configuration loading and validation
└── telemetry/
    └── metrics.go       # Metrics collection and reporting

pkg/types/
├── config.go            # Configuration data structures
├── measurement.go       # Measurement result structures
├── probe.go             # Probe command structures and definitions
└── target.go            # Target configuration structures
```

## 🐛 Troubleshooting

### Common Issues

#### Windows ICMP Permission
- **Issue**: `platform validation failed: Windows platform requires additional configuration`
- **Solution**: Run as Administrator or use simulation mode (default)

#### No Output
- **Issue**: Command runs but produces no output
- **Solution**: Check target addresses and network connectivity

#### High Latency Values
- **Issue**: Measurements showing unrealistic high latencies
- **Solution**: Normal for simulation mode; real ICMP requires admin privileges

### Platform-Specific Notes

**Windows:**
- Simulation mode is default (no admin privileges required)
- For real ICMP, run as Administrator
- Precision limited to millisecond timing

**Linux:**
- Best performance with raw sockets and CAP_NET_RAW
- Nanosecond timing precision available
- Multi-architecture support (x86_64, ARM64, ARM)

**macOS:**
- BPF-based measurements
- Microsecond timing precision
- Requires permission removal for non-signed binaries

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📞 Support

For support, please open an issue in the repository or contact the development team.

---

*Mesh Probe System - High-precision ICMP measurements with statistical analysis, continuous monitoring, and network path tracing*