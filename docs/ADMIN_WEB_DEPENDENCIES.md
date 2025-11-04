# Admin Web Component Dependencies

This document provides a comprehensive explanation of all dependencies used by the `cmd/admin-web` component, including external libraries, internal packages, and their relationships.

## Table of Contents

- [Overview](#overview)
- [External Dependencies](#external-dependencies)
- [Internal Dependencies](#internal-dependencies)
- [Standard Library Dependencies](#standard-library-dependencies)
- [Dependency Architecture](#dependency-architecture)
- [Configuration and Environment](#configuration-and-environment)
- [Development Notes](#development-notes)

## Overview

The admin-web component is a Go-based HTTP server that provides a RESTful API for managing the mesh probe system. It handles authentication, configuration management, monitoring data, and probe registration through a web interface.

**Component Location:** `cmd/admin-web/main.go`  
**Primary Purpose:** Web administration interface and API gateway  
**Key Features:** JWT authentication, CORS support, monitoring endpoints, configuration management

## External Dependencies

### Core Web Framework

#### `github.com/gin-gonic/gin v1.11.0`
- **Purpose:** HTTP web framework for building RESTful APIs
- **Usage in Code:** Router creation, middleware handling, request routing
- **Key Features Used:**
  - HTTP request/response handling
  - Middleware system for authentication and CORS
  - Parameter binding and validation
  - JSON marshaling
  - WebSocket support (future implementation)

```go
// Used in createRouter function
router := gin.Default()
router.Use(gin.Recovery())
```

### Logging and Observability

#### `github.com/go-logr/logr v1.2.4` + `github.com/go-logr/stdr v1.2.2`
- **Purpose:** Structured logging framework with OpenTelemetry integration
- **Usage in Code:** Global logger initialization, component-specific logging
- **Key Features Used:**
  - Structured logging with contextual information
  - Integration with OpenTelemetry for distributed tracing
  - Different log levels (Debug, Info, Warn, Error)

```go
// Used in main.go and internal/logger/manager.go
logger.InitGlobalLogger(logger.InfoLevel)
```

#### `go.opentelemetry.io/otel v1.19.0`
- **Purpose:** OpenTelemetry observability framework
- **Sub-packages:** `metric`, `trace`
- **Usage in Code:** Distributed tracing and metrics collection
- **Key Features Used:**
  - Request tracing
  - Performance metrics
  - Span correlation across services

### Authentication and Security

#### `github.com/golang-jwt/jwt/v5 v5.3.0`
- **Purpose:** JWT token creation and validation
- **Usage in Code:** User authentication middleware, token generation
- **Key Features Used:**
  - HS256 signing method
  - Custom claims structure
  - Token validation and parsing

```go
// Used in internal/web/auth/middleware.go
type Claims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

### Configuration Management

#### `github.com/spf13/viper v1.18.2`
- **Purpose:** Configuration file parsing and management
- **Usage in Code:** Configuration loading, environment variable handling
- **Key Features Used:**
  - JSON/YAML file parsing
  - Environment variable binding
  - Configuration watching and reloading

#### `github.com/spf13/cobra v1.8.0`
- **Purpose:** CLI command framework
- **Usage in Code:** Command-line interface support (likely for admin commands)
- **Key Features Used:**
  - Command hierarchy
  - Flag parsing
  - Subcommand support

### Service Discovery and Configuration Storage

#### `github.com/hashicorp/consul/api v1.33.0`
- **Purpose:** HashiCorp Consul client for service discovery
- **Usage in Code:** Dynamic configuration from Consul
- **Key Features Used:**
  - Key-value store access
  - Service registration
  - Health checking integration

#### `go.etcd.io/etcd/client/v3 v3.5.10`
- **Purpose:** etcd v3 client for distributed configuration
- **Usage in Code:** Cluster configuration management
- **Key Features Used:**
  - Key-value operations
  - Watch functionality for config changes
  - Lease management for temporary configurations

### Networking and Protocols

#### `golang.org/x/net v0.43.0`
- **Purpose:** Extended networking utilities
- **Usage in Code:** Advanced networking features, HTTP/2 support
- **Key Features Used:**
  - HTTP/2 support
  - Network address parsing
  - Advanced protocol support

## Internal Dependencies

The admin-web component relies on several internal packages that provide core functionality:

### Core System Components

#### `github.com/mesh-net-probe/probe/internal/config`
- **Purpose:** Unified configuration management interface
- **Key Components:**
  - `Manager` interface: Configuration orchestration
  - `ManagerStatus`: Health monitoring of config sources
  - `SourceStatus`: Individual configuration source health
  - `DefaultConfigurationHandler`: Default implementation

```go
// Used in main.go
configManager := initConfigManager()
// Interface methods:
- Initialize(ctx, config)
- GetConfiguration(ctx)
- WatchConfiguration(ctx, handler)
- GetStatus(ctx)
```

#### `github.com/mesh-net-probe/probe/internal/monitoring`
- **Purpose:** Monitoring data management and probe registry
- **Key Components:**
  - `Manager`: Measurement collection and storage
  - `ProbeRegistry`: Probe registration and lifecycle management
  - `Measurement`: Network measurement data structure
  - `MeasurementStream`: Real-time measurement streaming

```go
// Used in main.go
monitoringMgr := initMonitoringManager()
probeRegistry := initProbeRegistry()

// Manager capabilities:
- AddMeasurement(measurement)
- GetMeasurements(probeID, limit)
- GetMeasurementStatistics()
- StartMeasurementStream()
```

#### `github.com/mesh-net-probe/probe/internal/logger`
- **Purpose:** Structured logging service
- **Key Components:**
  - `Manager`: Centralized logging management
  - `ComponentLogger`: Component-specific logging
  - `LogHandler`: Interface for different output backends
  - `LoggingConfig`: Configuration for logging behavior

```go
// Used in main.go and auth middleware
logger.InitGlobalLogger(logger.InfoLevel)
logger.GetGlobalLogger()

// Logging features:
- Structured logging with context
- Component-specific loggers
- Multiple output handlers (console, file)
- Log level management
```

### Web API Components

#### `github.com/mesh-net-probe/probe/internal/web/api`
- **Purpose:** RESTful API route definitions and handlers
- **Key Components:**
  - Route registration functions
  - Handler functions for all API endpoints
  - Data models for API responses
  - In-memory storage for demo purposes

```go
// Used in main.go
api.RegisterAuthRoutes(apiRouter, authMiddleware)
api.RegisterConfigRoutes(apiRouter, configManager, authMiddleware)
api.RegisterProbeRoutes(apiRouter, probeRegistry, configManager, authMiddleware)
api.RegisterMeasurementRoutes(apiRouter, monitoringMgr, authMiddleware)
api.RegisterMonitoringRoutes(apiRouter, probeRegistry, monitoringMgr, authMiddleware)
```

#### `github.com/mesh-net-probe/probe/internal/web/auth`
- **Purpose:** Authentication and authorization middleware
- **Key Components:**
  - `Middleware`: JWT-based authentication
  - Role-based access control
  - Rate limiting
  - CORS handling

```go
// Used in main.go
authMiddleware := initAuthMiddleware()

// Middleware features:
- JWT token validation
- CORS support
- Rate limiting (100 requests/minute)
- Role-based access control
- User authentication/authorization
```

## Standard Library Dependencies

### Core System Functions

#### `context`
- **Purpose:** Request context and cancellation
- **Usage:** HTTP request handling, configuration context
- **Key Functions:** `context.Background()`, `context.WithTimeout()`

#### `net/http`
- **Purpose:** HTTP server implementation
- **Usage:** Server creation, HTTP response handling
- **Key Features:** Request/response handling, server lifecycle

```go
// Used in main.go
srv := &http.Server{
    Addr:    ":" + port,
    Handler: createRouter(...),
}
```

#### `os`, `os/signal`, `syscall`
- **Purpose:** System interaction and signal handling
- **Usage:** Environment variables, graceful shutdown
- **Key Features:** Environment variable reading, signal handling for shutdown

```go
// Used in main.go
port := os.Getenv("ADMIN_WEB_PORT")
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
```

#### `time`
- **Purpose:** Time operations and scheduling
- **Usage:** Timestamps, rate limiting, health checks
- **Key Features:** Time formatting, duration calculations

#### `sync`
- **Purpose:** Synchronization primitives
- **Usage:** Concurrent access to shared data
- **Key Features:** Mutex, RWMutex for thread safety

#### `log`
- **Purpose:** Basic logging (fallback to structured logging)
- **Usage:** Initial logging before structured logger setup
- **Key Features:** Simple log output

#### `encoding/json`
- **Purpose:** JSON serialization
- **Usage:** API responses, configuration parsing
- **Key Features:** JSON marshaling/unmarshaling

## Dependency Architecture

### Component Interaction Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    main.go (Entry Point)                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
            ┌─────────▼─────────┐
            │  Component Setup  │
            └─────────┬─────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
┌───────▼────┐  ┌────▼────┐  ┌────▼────┐
│   Config   │  │Monitor  │  │  Auth   │
│  Manager   │  │  Manager│  │Middleware│
└────────────┘  └─────────┘  └─────────┘
        │             │             │
        └─────────────┼─────────────┘
                      │
            ┌─────────▼─────────┐
            │   Gin Router      │
            │   (HTTP Server)   │
            └───────────────────┘
```

### Data Flow

1. **HTTP Request** → Gin Router
2. **Authentication** → Auth Middleware (JWT validation)
3. **Route Handler** → API Package
4. **Business Logic** → Internal Packages
5. **Response** → HTTP Response

### Dependency Injection Pattern

The admin-web component uses dependency injection for better testability and modularity:

```go
// Components are initialized and passed to router
configManager := initConfigManager()
monitoringMgr := initMonitoringManager()
probeRegistry := initProbeRegistry()
authMiddleware := initAuthMiddleware()

router := createRouter(configManager, monitoringMgr, probeRegistry, authMiddleware)
```

## Configuration and Environment

### Environment Variables

- `ADMIN_WEB_PORT`: Server port (default: 8080)
- JWT secret key: Should be set from environment (currently hardcoded in demo)

### Configuration Sources

The component supports multiple configuration sources:
1. **File-based**: JSON/YAML configuration files
2. **Consul**: HashiCorp Consul for dynamic configuration
3. **etcd**: etcd for distributed configuration
4. **Environment**: Direct environment variable binding

### Port Configuration

```go
port := os.Getenv("ADMIN_WEB_PORT")
if port == "" {
    port = "8080"
}
```

## Development Notes

### Security Considerations

1. **JWT Secret**: Currently hardcoded as "your-secret-key" - must be externalized
2. **CORS**: Allows localhost development origins
3. **Rate Limiting**: Basic IP-based rate limiting implemented
4. **Input Validation**: Uses Gin binding for request validation

### Performance Considerations

1. **Memory Management**: Measurements limited to 1000 per probe
2. **Cleanup**: Automatic cleanup of old measurements (24-hour retention)
3. **Concurrency**: Thread-safe operations using mutexes
4. **Streaming**: WebSocket endpoints prepared for real-time data

### Future Enhancements

1. **WebSocket Implementation**: Real-time updates for dashboard
2. **Database Integration**: Replace in-memory storage with persistent storage
3. **Advanced Metrics**: Integration with Prometheus
4. **Enhanced Security**: OAuth2 integration, API key management

### Testing Strategy

The modular design supports:
- Unit testing of individual components
- Integration testing of API endpoints
- Mock-based testing of external dependencies
- Performance testing of concurrent operations

## Summary

The admin-web component demonstrates a well-structured Go application with:

- **Clear separation of concerns** through internal package organization
- **Modern web development practices** using Gin framework
- **Comprehensive observability** through OpenTelemetry integration
- **Flexible configuration management** supporting multiple backends
- **Security-first design** with JWT authentication and CORS handling
- **Scalable architecture** ready for production deployment

The dependency choices reflect best practices for building reliable, observable, and maintainable web services in Go.