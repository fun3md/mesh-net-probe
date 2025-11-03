# Research: Distributed Mesh Probe System

**Phase 0 Output**: Technical research findings for distributed mesh probe system
**Created**: 2025-11-03

## Research Objectives

1. **Go ICMP Implementation**: Determine best approach for microsecond-precision ICMP measurements
2. **OpenTelemetry Integration**: Research OTLP protocol implementation for Go
3. **Configuration Management**: Evaluate etcd/Consul client libraries for Go
4. **Cross-Platform Testing**: Research Go testing frameworks for multi-platform validation
5. **Docker Multi-Platform**: Determine optimal build strategies for linux/amd64 and linux/arm64

## Research Findings

### 1. Go ICMP Implementation

**Decision**: Use `golang.org/x/net/icmp` package with custom timing implementation

**Rationale**: The `golang.org/x/net/icmp` package provides standard ICMPv4 and ICMPv6 support while allowing custom timestamp precision. For microsecond-level timing, we need to combine:
- High-resolution timers (`time.Now()` with microsecond precision)
- Custom ICMP packet structure with timestamps
- Platform-specific timing optimizations

**Alternatives Considered**:
- Raw sockets (`syscall` package): Better precision but more complex and permission-heavy
- External C libraries via cgo: Higher performance but breaks cross-platform portability
- Third-party ICMP libraries: Less control over timing precision

**Implementation Approach**:
```go
// Custom ICMP packet with timestamp tracking
type ICMPEchoRequest struct {
    Timestamp time.Time
    Sequence  uint16
    Data      []byte
}
```

### 2. OpenTelemetry Go SDK Integration

**Decision**: Use OpenTelemetry Go SDK with OTLP/gRPC exporter

**Rationale**: OpenTelemetry provides standardized observability with excellent Go support. OTLP/gRPC offers reliable, bidirectional communication suitable for real-time metric export.

**Key Components**:
- `go.opentelemetry.io/otel` (core SDK)
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` for tracing
- `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc` for metrics

**Implementation Strategy**:
```go
// Measurement metrics export
meter := otel.Meter("mesh-probe")
probeLatency, _ := meter.Int64Histogram("probe_latency",
    otel.WithDescription("ICMP measurement latency"))
```

### 3. Configuration Management Integration

**Decision**: Support both etcd and Consul with pluggable provider interface

**Rationale**: Different environments prefer different configuration systems. Providing abstraction allows deployment flexibility while maintaining centralized management capabilities.

**Provider Interface**:
```go
type ConfigProvider interface {
    Get(key string) (value []byte, err error)
    Watch(key string, handler func([]byte)) error
    Put(key string, value []byte) error
}
```

**Supported Providers**:
- etcd: For Kubernetes environments
- Consul: For HashiCorp deployments
- Static file: For simple single-probe deployments

### 4. Cross-Platform Testing Framework

**Decision**: Use Go's built-in testing with custom cross-platform test suite

**Rationale**: Go's testing framework is excellent for cross-platform compatibility. Additional tools needed:
- `golang.org/x/sys/unix` for platform-specific ICMP testing
- Docker multi-platform testing via CI/CD
- Performance benchmarking with platform comparison

**Testing Strategy**:
- Unit tests: Platform-agnostic logic validation
- Integration tests: Platform-specific ICMP behavior
- Performance tests: Cross-platform latency comparison
- Contract tests: Configuration and telemetry APIs

### 5. Docker Multi-Platform Build Strategy

**Decision**: Use BuildKit multi-stage builds with platform-specific optimization

**Rationale**: Docker BuildKit provides native multi-platform support with efficient layer caching and minimal image size.

**Build Strategy**:
```dockerfile
# Multi-platform build
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /app

# Final image
FROM --platform=$TARGETPLATFORM alpine:latest
COPY --from=builder /app /usr/local/bin/
```

**Platform Support**:
- Base image: Alpine Linux (minimal, multi-arch support)
- Binary: Static linking for maximum compatibility
- Runtime: Non-root user with ICMP capabilities

## Technical Implementation Summary

**Core Architecture**:
- Single Go binary with CLI interface
- Modular design with clear separation of concerns
- Cross-platform compatibility through abstraction layers
- High-performance measurement engine with microsecond precision

**Key Dependencies**:
- `golang.org/x/net/icmp`: ICMP protocol implementation
- `go.opentelemetry.io/otel`: Observability and metrics
- `github.com/coreos/etcd/clientv3`: Configuration management
- `github.com/hashicorp/consul/api`: Alternative configuration provider

**Performance Considerations**:
- Zero-allocation paths for hot measurement code
- Efficient ICMP packet handling with minimal overhead
- Buffered telemetry export to prevent measurement disruption
- Platform-specific optimizations while maintaining consistency

## Constitutional Compliance Review

**Code Quality**: Clean architecture with separation of concerns, comprehensive error handling, and documentation requirements addressed
**Testing**: Cross-platform test coverage with performance benchmarks and integration testing
**User Experience**: CLI consistency with JSON output support and intuitive configuration management
**Performance**: Microsecond-level precision with resource optimization and bounded memory footprint
**Cross-Platform**: Identical measurement semantics across all target platforms and architectures

All constitutional principles are satisfied by the research findings and proposed implementation approach.