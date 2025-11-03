# Admin Web Server - Local Development & Testing Guide

## 📋 Overview

This guide covers local development, testing, and deployment of the Admin Web Server (`cmd/admin-web`) component of the Mesh Probe System. The admin web server provides HTTP APIs for configuration management, probe monitoring, and real-time measurement data.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │  Admin Web      │    │  Probe Engine   │
│   (React/Vite)  │◄──►│  Server (Go)    │◄──►│  (ICMP Engine)  │
│   Port: 5173    │    │  Port: 8080     │    │  CLI Tools      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                │
                    ┌─────────────────┐
                    │ Configuration   │
                    │ Sources         │
                    │ • JSON Files    │
                    │ • etcd          │
                    │ • Consul        │
                    └─────────────────┘
```

## 🚀 Quick Start

### 1. Prerequisites

- **Go 1.21+** - For building the admin web server
- **Docker & Docker Compose** - For containerized deployment
- **curl** or **Postman** - For API testing
- **Node.js 18+** (optional) - For frontend development

### 2. Build and Run

```bash
# Clone and build the project
git clone <repository-url>
cd mesh-net-probe

# Build admin web server
cd cmd/admin-web
go build -o admin-web-test.exe

# Run the server
./admin-web-test.exe

# Server will start on http://localhost:8080
```

### 3. Test the Server

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","time":"2025-11-03T19:55:00.2246869Z"}
```

## 📁 Project Structure

```
cmd/admin-web/
├── main.go                 # Main server implementation
└── admin-web-test.exe      # Compiled binary (after build)

docs/
├── ADMIN_WEB_LOCAL_TESTING.md  # This guide
├── example-config/              # Example configurations
│   ├── development.json        # Development config
│   ├── production.json         # Production config
│   └── testing.json            # Testing config
└── docker-compose/             # Docker compose files
    ├── development.yml         # Development environment
    ├── production.yml          # Production setup
    └── full-stack.yml          # Complete stack with frontend
```

## 🔧 Development Setup

### Environment Variables

```bash
# Core configuration
export ADMIN_WEB_PORT=8080                    # Server port
export ADMIN_WEB_LOG_LEVEL=info               # Log level (debug, info, warn, error)
export ADMIN_WEB_HOST=0.0.0.0                 # Bind address

# Optional: Authentication (future feature)
export ADMIN_WEB_JWT_SECRET=your-secret-key   # JWT signing secret
export ADMIN_WEB_SESSION_TIMEOUT=24h          # Session timeout

# Optional: External config providers
export ETCD_ENDPOINTS=http://localhost:2379   # etcd configuration
export CONSUL_ADDR=localhost:8500             # Consul configuration

# Optional: Monitoring
export ADMIN_WEB_METRICS_ENABLED=true         # Enable metrics endpoint
export ADMIN_WEB_TRACING_ENABLED=true         # Enable distributed tracing
```

### VS Code Launch Configuration

Create `.vscode/launch.json` for debugging:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Admin Web",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/admin-web",
            "env": {
                "ADMIN_WEB_PORT": "8080",
                "ADMIN_WEB_LOG_LEVEL": "debug"
            },
            "args": [],
            "console": "integratedTerminal"
        }
    ]
}
```

## 🧪 Testing

### API Testing

#### 1. Health Endpoints

```bash
# Basic health check
curl -s http://localhost:8080/health | jq

# Monitoring health (future feature)
curl -s http://localhost:8080/api/v1/monitoring/health | jq
```

#### 2. Configuration Management

```bash
# Get current configuration
curl -s http://localhost:8080/api/v1/config | jq

# Update configuration (future feature)
curl -X POST http://localhost:8080/api/v1/config \
  -H "Content-Type: application/json" \
  -d '{"targets":[...]}'
```

#### 3. Probe Management

```bash
# List configured probes
curl -s http://localhost:8080/api/v1/probes | jq

# Add new probe (future feature)
curl -X POST http://localhost:8080/api/v1/probes \
  -H "Content-Type: application/json" \
  -d '{"id":"test-probe","address":"8.8.8.8"}'
```

#### 4. Measurement Data

```bash
# Get recent measurements
curl -s http://localhost:8080/api/v1/measurements | jq

# Real-time measurements (future WebSocket)
curl -s http://localhost:8080/ws/test | jq
```

### Load Testing

```bash
# Install hey for load testing
# Windows: Download from https://github.com/rakyll/hey/releases
# Linux/Mac: go install github.com/rakyll/hey@latest

# Basic load test
hey -n 1000 -c 10 http://localhost:8080/health

# Detailed load test
hey -n 10000 -c 50 -q 10 -z 30s http://localhost:8080/api/v1/config
```

### Integration Testing

Create test scripts:

```bash
# test-api.sh - Comprehensive API testing
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "🧪 Testing Admin Web Server APIs..."

# Test health endpoint
echo "Testing health endpoint..."
response=$(curl -s -w "%{http_code}" -o /tmp/health.json $BASE_URL/health)
if [ "$response" = "200" ]; then
    echo "✅ Health check passed"
    cat /tmp/health.json | jq .
else
    echo "❌ Health check failed (HTTP $response)"
fi

# Test config endpoint
echo "Testing config endpoint..."
response=$(curl -s -w "%{http_code}" -o /tmp/config.json $BASE_URL/api/v1/config)
if [ "$response" = "200" ]; then
    echo "✅ Config endpoint passed"
    cat /tmp/config.json | jq .
else
    echo "❌ Config endpoint failed (HTTP $response)"
fi

# Test probes endpoint
echo "Testing probes endpoint..."
response=$(curl -s -w "%{http_code}" -o /tmp/probes.json $BASE_URL/api/v1/probes)
if [ "$response" = "200" ]; then
    echo "✅ Probes endpoint passed"
    cat /tmp/probes.json | jq .
else
    echo "❌ Probes endpoint failed (HTTP $response)"
fi

echo "🎯 API testing complete!"
```

## 🐳 Docker Development

### Development Docker Compose

See `docker-compose/development.yml` for complete development setup:

```bash
# Start development environment
docker-compose -f docker-compose/development.yml up -d

# View logs
docker-compose -f docker-compose/development.yml logs -f admin-web

# Stop environment
docker-compose -f docker-compose/development.yml down
```

### Full Stack Development

For complete frontend + backend development:

```bash
# Start full stack (frontend + backend + dependencies)
docker-compose -f docker-compose/full-stack.yml up -d

# Access frontend
open http://localhost:5173

# Access admin API
open http://localhost:8080
```

### Production Docker

```bash
# Build production image
docker build -f Dockerfile.multi-platform --target runtime-web -t mesh-probe:web .

# Run production container
docker run -d \
  --name admin-web-prod \
  -p 8080:8080 \
  -e ADMIN_WEB_PORT=8080 \
  mesh-probe:web
```

## 🔍 Debugging

### Common Issues

#### 1. Server Won't Start

```bash
# Check if port is already in use
netstat -an | findstr :8080

# Kill existing processes
taskkill /f /im admin-web-test.exe

# Check for compilation errors
go build -v ./cmd/admin-web
```

#### 2. HTTP Requests Time Out

```bash
# Verify server is listening
netstat -an | findstr :8080

# Check server logs for errors
# Look for middleware issues or dependency problems

# Test with simple curl
curl -v http://localhost:8080/health
```

#### 3. CORS Issues

```bash
# Test with proper headers
curl -H "Origin: http://localhost:5173" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     http://localhost:8080/health
```

### Debug Logging

Enable debug logging:

```bash
# Set environment variable
export ADMIN_WEB_LOG_LEVEL=debug

# Run server with verbose logging
./admin-web-test.exe

# Check Gin debug output
[GIN-debug] GET    /health    --> main.createRouter.func1 (4 handlers)
2025/11/03 20:55:00 Health check requested from 127.0.0.1
```

### Network Debugging

```bash
# Check server connectivity
telnet localhost 8080

# Monitor HTTP traffic
# Windows: Use Fiddler or Wireshark
# Linux/Mac: sudo tcpdump -i lo0 port 8080

# Test WebSocket connections (future feature)
wscat -c ws://localhost:8080/ws/test
```

## 📊 Monitoring

### Health Checks

```bash
# Automated health monitoring
while true; do
    curl -s http://localhost:8080/health | jq .status
    sleep 5
done
```

### Performance Monitoring

```bash
# Monitor response times
curl -w "@curl-format.txt" -s -o /dev/null http://localhost:8080/health

# Create curl-format.txt:
#      time_namelookup:  %{time_namelookup}\n
#         time_connect:  %{time_connect}\n
#      time_appconnect:  %{time_appconnect}\n
#     time_pretransfer:  %{time_pretransfer}\n
#        time_redirect:  %{time_redirect}\n
#   time_starttransfer:  %{time_starttransfer}\n
#                     ----------\n
#           time_total:  %{time_total}\n
```

### Log Analysis

```bash
# Monitor server logs in real-time
tail -f server.log | jq .

# Search for errors
grep -i error server.log | jq .

# Analyze request patterns
grep "Health check requested" server.log | jq .
```

## 🔧 Configuration

### Example Configurations

See `example-config/` directory for sample configurations:

- **development.json** - Development environment settings
- **production.json** - Production deployment configuration  
- **testing.json** - Testing environment setup

### Environment-Specific Setup

#### Development
- Debug logging enabled
- CORS allows localhost origins
- Hot reloading supported
- Verbose error messages

#### Production
- Info logging level
- Restricted CORS origins
- Request rate limiting
- Structured logging

#### Testing
- Mock data responses
- Simulated delays
- Error injection
- Performance baselines

## 🚀 Deployment

### Local Deployment

```bash
# Build and run locally
cd cmd/admin-web
go build -o admin-web-prod.exe
./admin-web-prod.exe

# Background execution (Linux/Mac)
nohup ./admin-web-prod.exe > admin-web.log 2>&1 &

# Background execution (Windows)
start /B admin-web-prod.exe
```

### Production Deployment

See Docker deployment options and Kubernetes configurations in the main project documentation.

### Service Integration

```bash
# Register as Windows service (future)
# Configure systemd service (Linux)
# Deploy to cloud platforms (AWS, GCP, Azure)
```

## 🔗 API Reference

### Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check | No |
| GET | `/api/v1/config` | Get configuration | No* |
| POST | `/api/v1/auth/login` | User authentication | No |
| GET | `/api/v1/probes` | List probes | No* |
| GET | `/api/v1/measurements` | Get measurements | No* |
| GET | `/api/v1/monitoring/health` | System health | No* |
| GET | `/ws/test` | WebSocket test | No |

*Future versions will require authentication

### Response Formats

```json
// Success response
{
  "status": "ok",
  "data": {...},
  "timestamp": "2025-11-03T20:00:00Z"
}

// Error response
{
  "error": "error_description",
  "code": "ERROR_CODE",
  "timestamp": "2025-11-03T20:00:00Z"
}
```

## 🐛 Troubleshooting

### Server Issues

1. **Port binding errors**
   - Check if port is available
   - Verify firewall settings
   - Use `netstat -an | findstr :8080`

2. **Request timeouts**
   - Check middleware dependencies
   - Verify logging configuration
   - Test with simplified endpoints

3. **CORS issues**
   - Verify Origin headers
   - Check allowed origins configuration
   - Test preflight requests

### Docker Issues

1. **Container won't start**
   - Check Docker logs: `docker logs admin-web`
   - Verify port conflicts
   - Check environment variables

2. **Build failures**
   - Verify Go version compatibility
   - Check dependency downloads
   - Review multi-stage build output

### Network Issues

1. **Connection refused**
   - Verify server is running
   - Check bind address configuration
   - Test local connectivity

2. **Slow responses**
   - Monitor server resource usage
   - Check database/config provider latency
   - Analyze request patterns

## 📚 Additional Resources

- [Main README](../README.md) - Project overview
- [Configuration Guide](CONFIGURATION_GUIDE.md) - External config providers
- [Docker Validation](DOCKER_VALIDATION.md) - Container deployment
- [API Documentation](http://localhost:8080/docs) - Interactive API docs (future)

---

**Last Updated**: 2025-11-03 20:00:00 UTC  
**Version**: 1.0.0  
**Maintainer**: Development Team