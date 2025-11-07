# Implementation Plan: Distributed Mesh Probe System

**Branch**: `001-mesh-probe-system` | **Date**: 2025-11-03 | **Spec**: specs/001-mesh-probe-system/spec.md
**Input**: Feature specification from `/specs/001-mesh-probe-system/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a distributed mesh probe system that collects high-resolution ICMP measurements with microsecond precision across Linux, macOS, and Windows on both x64 and ARM architectures. The system provides centralized configuration management with real-time web interface for administration and monitoring. The solution includes both CLI tools and a comprehensive admin web interface that exports all metrics and logs to OpenTelemetry Collector via OTLP protocol for comprehensive observability.

### System Components Overview

1. **Core Probe CLI**: Command-line interface for individual probe operations (ping, traceroute)
2. **Configuration Management**: Centralized etcd/Consul-based configuration with real-time updates
3. **Admin Web Interface**: Web-based management dashboard for configuration and monitoring
4. **Mesh Coordination**: Distributed probe coordination and measurement aggregation
5. **Telemetry Export**: OpenTelemetry integration for comprehensive observability
6. **Cross-Platform Support**: Native support for x86_64 and ARM64 architectures

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.21+ (for cross-platform compatibility and high performance), React 18 + TypeScript (for web interface)
**Primary Dependencies**: ICMP library for Go, OpenTelemetry Go SDK, configuration management client library, WebSocket libraries, React UI framework
**Storage**: Configuration management system (etcd/Consul) for centralized configuration, local buffer for measurement data during network outages, web interface state management
**Testing**: Go test suite with cross-platform testing framework, performance benchmarking tools, React component testing, E2E testing for web interface
**Target Platform**: Containerized deployment on Linux, macOS, Windows for both x64 and ARM64 architectures
**Project Type**: multi (Go CLI application + React web interface + shared configuration)
**Performance Goals**: Microsecond-level measurement precision, support 100+ concurrent probes, sub-second configuration propagation, sub-100ms web interface response times
**Constraints**: ICMP requires elevated privileges, multi-platform consistency mandatory, OTLP export must be reliable, web interface must support real-time updates
**Scale/Scope**: Support 100+ concurrent probe instances, handle high-frequency ICMP measurements without network degradation, manage multiple web interface clients simultaneously

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Code Quality Excellence Requirements
- [x] Go code follows strict quality standards with comprehensive error handling
- [x] Clean architecture with separation of concerns implemented
- [x] All public APIs and complex logic have comprehensive documentation
- [x] golangci-lint configuration enforces consistent coding style
- [x] Memory-safe practices prevent leaks and race conditions

### Testing Standards Compliance
- [x] MANDATORY test coverage: 80% minimum for all packages
- [x] Unit tests cover all business logic and utilities
- [x] Integration tests validate network protocols and OS interactions
- [x] Cross-platform testing plan includes Linux, macOS, Windows
- [x] Performance benchmarks included for measurement accuracy and latency
- [x] End-to-end tests simulate real operational scenarios

### User Experience Consistency Requirements
- [x] CLI interface provides consistent command patterns and flag structures
- [x] Text-based output supports both human-readable and JSON formats
- [x] Configuration is intuitive with sensible defaults and clear validation
- [x] Error messages are actionable with suggested resolutions
- [x] Progress indicators and status reporting are informative yet non-verbose

### Performance Requirements
- [x] Measurement precision achieves domain-appropriate accuracy (microsecond-level for network measurement)
- [x] Resource consumption optimized for long-running deployments
- [x] Scalable architecture handles high-frequency operations without degradation
- [x] Hot code paths optimized (zero-allocation where applicable)
- [x] Memory footprint remains bounded regardless of operation duration

### Cross-Platform Compatibility
- [x] Identical behavior maintained across target platforms (Linux, macOS, Windows)
- [x] Platform-specific optimizations preserve functional equivalence
- [x] Architecture support verified for target architectures (x64, ARM64)
- [x] Platform detection and capability checking are automatic and transparent

## Project Structure

### Documentation (this feature)

```text
specs/001-mesh-probe-system/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── admin-web-interface.md # Admin web interface specification
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
src/
├── cmd/
│   ├── probe/          # Main CLI application (ping, traceroute)
│   └── admin-web/      # Admin web interface backend service
├── internal/
│   ├── config/         # Configuration management (etcd/Consul providers)
│   ├── icmp/           # ICMP measurement engine
│   ├── telemetry/      # OpenTelemetry integration
│   ├── mesh/           # Mesh network coordination
│   ├── platform/       # Cross-platform utilities
│   ├── web/            # Web interface services (WebSocket, REST API)
│   │   ├── websocket/  # Real-time WebSocket communication
│   │   ├── api/        # REST API endpoints
│   │   └── auth/       # Authentication and authorization
│   └── monitoring/     # Probe registry and health monitoring
├── web/                # React TypeScript frontend
│   ├── src/
│   │   ├── components/ # Reusable React components
│   │   ├── pages/      # Main application pages
│   │   ├── hooks/      # Custom React hooks
│   │   ├── services/   # API and WebSocket clients
│   │   └── types/      # TypeScript type definitions
│   ├── public/         # Static assets
│   └── package.json    # Frontend dependencies
└── pkg/
    └── types/          # Shared data types and contracts
tests/
├── unit/               # Unit tests (Go + React)
├── integration/        # Cross-platform integration tests
├── performance/        # Benchmark tests
├── contract/           # API contract tests
├── web/                # Web interface tests (Jest, Cypress)
│   ├── components/     # React component tests
│   ├── e2e/            # End-to-end tests
│   └── api/            # API integration tests
└── load/               # Load testing for web interface
```

**Structure Decision**: Multi-component system with separate Go CLI application, web interface backend, and React frontend. Clean separation between probe operations (cmd/probe), web management services (cmd/admin-web, internal/web), and configuration management. Testing covers both Go backend and React frontend with comprehensive cross-platform validation.

## Admin Web Interface Integration

### Integration Architecture

The admin web interface integrates seamlessly with the existing mesh probe toolset through the following mechanisms:

#### 1. **Configuration Provider Integration**
- **etcd Provider**: Web interface directly manages configurations in the same etcd cluster used by CLI probes
- **Real-time Sync**: Changes made through web interface are immediately available to CLI tools
- **Version Control**: Configuration versions tracked for both web and CLI access
- **Validation**: Same validation logic used by both web interface and CLI tools

#### 2. **Probe Registry Integration**
- **Self-Registration**: CLI probes register with the admin web interface via API endpoints
- **Health Reporting**: Probes send periodic health status updates to the web interface
- **Measurement Streaming**: Real-time measurement data streamed from probes to web dashboard
- **Configuration Distribution**: Web interface pushes configurations to registered probes

#### 3. **Unified Authentication**
```yaml
authentication_flow:
  web_interface:
    - User authenticates via web interface
    - JWT token issued for API access
    - Token used for probe registration and configuration access
  
  cli_integration:
    - CLI probes use same authentication endpoints
    - Service accounts for probe-to-web communication
    - Token-based configuration access
    - Consistent permission model across web and CLI
```

#### 4. **Tool Integration Examples**

**CLI Tool Integration**:
```bash
# Existing CLI tools continue to work unchanged
./probe ping 8.8.8.8
./probe traceroute 8.8.8.8

# New admin commands for web interface integration
./probe register --web-endpoint=http://admin-web:8080
./probe sync-config --from-web
./probe report-health --to-web
```

**Web Interface Workflow**:
1. **Configuration Creation**: User creates configuration in web interface
2. **Probe Registration**: Probes automatically discover and register with web interface
3. **Configuration Deployment**: Web interface pushes configuration to registered probes
4. **Real-time Monitoring**: Web interface displays live probe status and measurements
5. **Alert Management**: Web interface receives and displays alerts from CLI tools

#### 5. **Cross-Platform Deployment**

**Containerized Deployment**:
```yaml
services:
  probe-cli:
    image: mesh-probe/cli:latest
    command: ["probe", "ping", "8.8.8.8"]
    
  admin-web:
    image: mesh-probe/admin-web:latest
    ports:
      - "8080:8080"  # HTTP API
      - "8081:8081"  # WebSocket
    
  etcd:
    image: etcd:latest
    # Central configuration store
```

**Native Binary Deployment**:
```bash
# Cross-platform builds for admin web interface
GOOS=linux GOARCH=amd64 go build -o admin-web-amd64 ./cmd/admin-web/
GOOS=linux GOARCH=arm64 go build -o admin-web-arm64 ./cmd/admin-web/
GOOS=windows GOARCH=amd64 go build -o admin-web-amd64.exe ./cmd/admin-web/

# Same builds for CLI tools (already implemented)
GOOS=linux GOARCH=amd64 go build -o probe-amd64 ./cmd/probe/
GOOS=linux GOARCH=arm64 go build -o probe-arm64 ./cmd/probe/
```

### Data Flow Integration

#### Configuration Management Flow:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Interface │    │   etcd/Consul   │    │   CLI Probes    │
│                 │    │                 │    │                 │
│ 1. Create Config│◄──►│ 2. Store Config │◄──►│ 3. Load Config  │
│                 │    │                 │    │                 │
│ 4. Update Config│◄──►│ 5. Version Ctrl │◄──►│ 6. Apply Config │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

#### Real-time Monitoring Flow:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Probes    │    │ Admin Web       │    │ Web Dashboard   │
│                 │    │ Backend         │    │                 │
│ 1. Collect Data │───►│ 2. Aggregate    │───►│ 3. Display      │
│                 │    │                 │    │                 │
│ 4. Health Check │───►│ 5. WebSocket    │───►│ 6. Real-time UI │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### API Integration

#### Shared API Endpoints:
```go
// Configuration management (shared by web interface and CLI tools)
type ConfigAPI interface {
    GetConfig(ctx context.Context, id string) (*types.Configuration, error)
    CreateConfig(ctx context.Context, config *types.Configuration) error
    UpdateConfig(ctx context.Context, id string, config *types.Configuration) error
    DeployConfig(ctx context.Context, id string, targetProbes []string) error
}

// Probe registry (used by web interface for monitoring)
type ProbeRegistryAPI interface {
    RegisterProbe(ctx context.Context, probe *ProbeRegistration) error
    ReportHealth(ctx context.Context, probeID string, health *ProbeHealth) error
    StreamMeasurements(ctx context.Context, probeID string) (<-chan *Measurement, error)
}
```

### Security Integration

#### Unified Security Model:
- **Consistent Authentication**: Same JWT tokens work for web interface and CLI tools
- **Role-Based Access**: Admin, operator, viewer roles apply to both interfaces
- **Audit Logging**: All configuration changes logged regardless of source
- **Network Security**: TLS encryption for all inter-component communication

### Monitoring and Observability Integration

#### Unified Telemetry:
```yaml
telemetry_pipeline:
  sources:
    - cli_probes: "ICMP measurements, health status"
    - web_interface: "User actions, API calls"
    - admin_backend: "System metrics, WebSocket events"
  
  processing:
    - open_telemetry: "Unified metrics and traces"
    - alerts: "Cross-component alerting"
    - dashboards: "Unified visualization"
  
  export:
    - prometheus: "Metrics collection"
    - jaeger: "Distributed tracing"
    - elasticsearch: "Log aggregation"
```

This integration ensures the admin web interface becomes a natural extension of the existing toolset while providing powerful management capabilities for complex deployments.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
