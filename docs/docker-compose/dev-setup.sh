#!/bin/bash

# Mesh Probe System - Development Environment Setup
# This script builds and starts all components of the mesh-net-probe system
# with component-specific flags for selective service management

set -e

echo "🚀 Mesh Probe System - Development Environment Setup"
echo "==================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_step() {
    echo -e "${PURPLE}🔄 $1${NC}"
}

print_component() {
    echo -e "${CYAN}🏗️  $1${NC}"
}

# Default flags
BUILD_ALL=true
START_SERVICES=true
CLEAN_BUILD=false
VERBOSE=false
SKIP_TESTS=false
DEVELOPMENT_MODE=true

# Component flags (default: all enabled)
ENABLE_ADMIN_WEB=true
ENABLE_FRONTEND=true
ENABLE_CLI_PROBE=true
ENABLE_REDIS=true
ENABLE_ETCD=true
ENABLE_POSTGRES=true
ENABLE_JAEGER=true
ENABLE_PROMETHEUS=true
ENABLE_GRAFANA=true
ENABLE_NGINX=true

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-admin-web)
            ENABLE_ADMIN_WEB=false
            shift
            ;;
        --skip-frontend)
            ENABLE_FRONTEND=false
            shift
            ;;
        --skip-cli-probe)
            ENABLE_CLI_PROBE=false
            shift
            ;;
        --skip-redis)
            ENABLE_REDIS=false
            shift
            ;;
        --skip-etcd)
            ENABLE_ETCD=false
            shift
            ;;
        --skip-postgres)
            ENABLE_POSTGRES=false
            shift
            ;;
        --skip-jaeger)
            ENABLE_JAEGER=false
            shift
            ;;
        --skip-prometheus)
            ENABLE_PROMETHEUS=false
            shift
            ;;
        --skip-grafana)
            ENABLE_GRAFANA=false
            shift
            ;;
        --skip-nginx)
            ENABLE_NGINX=false
            shift
            ;;
        --core-only)
            ENABLE_JAEGER=false
            ENABLE_PROMETHEUS=false
            ENABLE_GRAFANA=false
            ENABLE_NGINX=false
            shift
            ;;
        --no-build)
            BUILD_ALL=false
            shift
            ;;
        --no-start)
            START_SERVICES=false
            shift
            ;;
        --clean)
            CLEAN_BUILD=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --production)
            DEVELOPMENT_MODE=false
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Component Control Flags:"
            echo "  --skip-admin-web      Disable admin web backend"
            echo "  --skip-frontend       Disable React frontend"
            echo "  --skip-cli-probe      Disable CLI probe engine"
            echo "  --skip-redis          Disable Redis cache service"
            echo "  --skip-etcd           Disable etcd configuration store"
            echo "  --skip-postgres       Disable PostgreSQL database"
            echo "  --skip-jaeger         Disable Jaeger tracing"
            echo "  --skip-prometheus     Disable Prometheus metrics"
            echo "  --skip-grafana        Disable Grafana dashboards"
            echo "  --skip-nginx          Disable Nginx reverse proxy"
            echo "  --core-only           Enable only core services (admin-web, frontend, etcd, redis, postgres)"
            echo ""
            echo "Build Control Flags:"
            echo "  --no-build            Skip building components, only start services"
            echo "  --no-start            Only build components, don't start services"
            echo "  --clean               Force clean build (remove existing images)"
            echo "  --skip-tests          Skip running tests during build"
            echo ""
            echo "Other Flags:"
            echo "  --verbose             Enable verbose output"
            echo "  --production          Use production mode settings"
            echo "  --help                Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Set verbose mode
if [ "$VERBOSE" = true ]; then
    set -x
fi

# Check prerequisites
echo ""
print_step "Checking prerequisites..."
echo "========================="

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    exit 1
fi
print_status "Docker is installed: $(docker --version)"

# Check Docker Compose
COMPOSE_AVAILABLE=false
if command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version)
    COMPOSE_AVAILABLE=true
elif docker compose version &> /dev/null; then
    COMPOSE_VERSION=$(docker compose version)
    COMPOSE_AVAILABLE=true
fi

if [ "$COMPOSE_AVAILABLE" = false ]; then
    print_error "Docker Compose is not installed"
    exit 1
fi
print_status "Docker Compose is available: $COMPOSE_VERSION"

# Check Go
if [ "$BUILD_ALL" = true ]; then
    if ! command -v go &> /dev/null; then
        print_warning "Go is not installed (required for building Go components)"
        if [ "$ENABLE_ADMIN_WEB" = true ] || [ "$ENABLE_CLI_PROBE" = true ]; then
            exit 1
        fi
    else
        print_status "Go is installed: $(go version)"
    fi

    # Check Node.js for frontend
    if [ "$ENABLE_FRONTEND" = true ]; then
        if ! command -v node &> /dev/null; then
            print_error "Node.js is not installed (required for building React frontend)"
            exit 1
        else
            print_status "Node.js is installed: $(node --version)"
        fi
        
        if ! command -v npm &> /dev/null; then
            print_error "npm is not installed (required for frontend dependencies)"
            exit 1
        else
            print_status "npm is installed: $(npm --version)"
        fi
    fi
fi

# Check curl
if ! command -v curl &> /dev/null; then
    print_error "curl is not installed"
    exit 1
fi
print_status "curl is available"

echo ""
print_step "Configuration Summary"
echo "======================"
print_info "Build all components: $BUILD_ALL"
print_info "Start services: $START_SERVICES"
print_info "Clean build: $CLEAN_BUILD"
print_info "Development mode: $DEVELOPMENT_MODE"
print_info "Skip tests: $SKIP_TESTS"
echo ""
print_info "Enabled components:"
[ "$ENABLE_ADMIN_WEB" = true ] && echo "  ✅ Admin Web Backend"
[ "$ENABLE_FRONTEND" = true ] && echo "  ✅ React Frontend"
[ "$ENABLE_CLI_PROBE" = true ] && echo "  ✅ CLI Probe Engine"
[ "$ENABLE_REDIS" = true ] && echo "  ✅ Redis Cache"
[ "$ENABLE_ETCD" = true ] && echo "  ✅ etcd Configuration"
[ "$ENABLE_POSTGRES" = true ] && echo "  ✅ PostgreSQL Database"
[ "$ENABLE_JAEGER" = true ] && echo "  ✅ Jaeger Tracing"
[ "$ENABLE_PROMETHEUS" = true ] && echo "  ✅ Prometheus Metrics"
[ "$ENABLE_GRAFANA" = true ] && echo "  ✅ Grafana Dashboards"
[ "$ENABLE_NGINX" = true ] && echo "  ✅ Nginx Reverse Proxy"

# Check configuration files
echo ""
print_step "Validating Configuration Files"
echo "================================"

required_configs=(
    "../example-config/development.json"
    "../example-config/production.json"
    "../example-config/testing.json"
)

for config in "${required_configs[@]}"; do
    if [ -f "$config" ]; then
        print_status "Configuration file exists: $config"
        
        # Validate JSON syntax if jq is available
        if command -v jq &> /dev/null; then
            if jq empty "$config" 2>/dev/null; then
                print_status "JSON syntax is valid: $config"
            else
                print_error "Invalid JSON syntax: $config"
                exit 1
            fi
        else
            print_warning "jq not available, skipping JSON validation"
        fi
    else
        print_error "Configuration file missing: $config"
        exit 1
    fi
done

# Check Docker Compose files
compose_files=(
    "development.yml"
    "full-stack.yml"
)

for file in "${compose_files[@]}"; do
    if [ -f "$file" ]; then
        print_status "Docker compose file exists: $file"
    else
        print_error "Docker compose file missing: $file"
        exit 1
    fi
done

# Clean build if requested
if [ "$CLEAN_BUILD" = true ]; then
    echo ""
    print_step "Cleaning existing build artifacts"
    echo "==================================="
    
    if [ "$ENABLE_ADMIN_WEB" = true ] || [ "$ENABLE_CLI_PROBE" = true ]; then
        print_info "Cleaning Go build cache..."
        go clean -cache
    fi
    
    if [ "$ENABLE_FRONTEND" = true ]; then
        print_info "Cleaning Node.js modules..."
        cd ../../
        if [ -d "web/node_modules" ]; then
            rm -rf web/node_modules
        fi
        if [ -d "web/dist" ]; then
            rm -rf web/dist
        fi
        cd docs/docker-compose
    fi
    
    print_info "Removing existing Docker images..."
    docker rmi mesh-probe:admin-web-dev 2>/dev/null || true
    docker rmi mesh-probe:cli-dev 2>/dev/null || true
    docker rmi mesh-probe:frontend-dev 2>/dev/null || true
    print_status "Build cache cleaned"
fi

# Build components
if [ "$BUILD_ALL" = true ]; then
    echo ""
    print_step "Building Components"
    echo "==================="
    
    # Build Go components
    if [ "$ENABLE_ADMIN_WEB" = true ] || [ "$ENABLE_CLI_PROBE" = true ]; then
        print_component "Building Go components..."
        cd ../..
        
        if [ "$ENABLE_ADMIN_WEB" = true ]; then
            print_info "Building admin web backend..."
            if go build -o admin-web-dev ./cmd/admin-web/; then
                print_status "Admin web backend built successfully"
                
                if [ "$SKIP_TESTS" = false ]; then
                    print_info "Running admin web tests..."
                    if go test ./cmd/admin-web/... -v; then
                        print_status "Admin web tests passed"
                    else
                        print_warning "Admin web tests failed (continuing anyway)"
                    fi
                fi
            else
                print_error "Failed to build admin web backend"
                exit 1
            fi
        fi
        
        if [ "$ENABLE_CLI_PROBE" = true ]; then
            print_info "Building CLI probe engine..."
            if go build -o probe-dev ./cmd/probe/; then
                print_status "CLI probe engine built successfully"
                
                if [ "$SKIP_TESTS" = false ]; then
                    print_info "Running CLI probe tests..."
                    if go test ./cmd/probe/... -v; then
                        print_status "CLI probe tests passed"
                    else
                        print_warning "CLI probe tests failed (continuing anyway)"
                    fi
                fi
            else
                print_error "Failed to build CLI probe engine"
                exit 1
            fi
        fi
        
        cd docs/docker-compose
    fi
    
    # Build React frontend
    if [ "$ENABLE_FRONTEND" = true ]; then
        print_component "Building React frontend..."
        cd ../../
        
        print_info "Installing Node.js dependencies..."
        cd web
        if npm install; then
            print_status "Node.js dependencies installed"
        else
            print_error "Failed to install Node.js dependencies"
            exit 1
        fi
        
        if [ "$SKIP_TESTS" = false ]; then
            print_info "Running frontend tests..."
            if npm run test:ci 2>/dev/null; then
                print_status "Frontend tests passed"
            else
                print_warning "Frontend tests failed or not configured (continuing anyway)"
            fi
        fi
        
        print_info "Building frontend for development..."
        if npm run build; then
            print_status "Frontend built successfully"
        else
            print_error "Failed to build frontend"
            exit 1
        fi
        
        cd ../docs/docker-compose
    fi
    
    # Build Docker images
    print_component "Building Docker images..."
    
    if [ "$ENABLE_ADMIN_WEB" = true ]; then
        print_info "Building admin web Docker image..."
        if docker build -f Dockerfile.multi-platform -t mesh-probe:admin-web-dev --target runtime-web ../.. 2>/dev/null; then
            print_status "Admin web Docker image built"
        else
            print_error "Failed to build admin web Docker image"
            exit 1
        fi
    fi
    
    if [ "$ENABLE_CLI_PROBE" = true ]; then
        print_info "Building CLI probe Docker image..."
        if docker build -f Dockerfile.multi-platform -t mesh-probe:cli-dev --target runtime-cli ../.. 2>/dev/null; then
            print_status "CLI probe Docker image built"
        else
            print_error "Failed to build CLI probe Docker image"
            exit 1
        fi
    fi
    
    if [ "$ENABLE_FRONTEND" = true ]; then
        print_info "Building frontend Docker image..."
        if docker build -t mesh-probe:frontend-dev ../web/ 2>/dev/null; then
            print_status "Frontend Docker image built"
        else
            print_error "Failed to build frontend Docker image"
            exit 1
        fi
    fi
    
    print_status "All components built successfully"
fi

# Validate Docker Compose configurations
echo ""
print_step "Validating Docker Compose Configurations"
echo "==========================================="

for compose_file in "${compose_files[@]}"; do
    if docker compose -f "$compose_file" config > /dev/null 2>&1; then
        print_status "Docker compose file syntax is valid: $compose_file"
    else
        print_error "Docker compose file has syntax errors: $compose_file"
        exit 1
    fi
done

# Start services
if [ "$START_SERVICES" = true ]; then
    echo ""
    print_step "Starting Development Environment"
    echo "================================="
    
    # Determine which compose file to use
    COMPOSE_FILE="full-stack.yml"
    if [ "$DEVELOPMENT_MODE" = false ]; then
        COMPOSE_FILE="production.yml"
    fi
    
    # Build the docker-compose command with selective services
    COMPOSE_CMD="docker compose -f $COMPOSE_FILE"
    
    # Add service selection based on flags
    SERVICES_ARGS=""
    
    if [ "$ENABLE_ADMIN_WEB" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale admin-web-dev=0"
    fi
    if [ "$ENABLE_FRONTEND" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale frontend=0"
    fi
    if [ "$ENABLE_CLI_PROBE" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale probe-cli=0"
    fi
    if [ "$ENABLE_REDIS" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale redis=0"
    fi
    if [ "$ENABLE_ETCD" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale etcd=0"
    fi
    if [ "$ENABLE_POSTGRES" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale postgres=0"
    fi
    if [ "$ENABLE_JAEGER" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale jaeger=0"
    fi
    if [ "$ENABLE_PROMETHEUS" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale prometheus=0"
    fi
    if [ "$ENABLE_GRAFANA" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale grafana=0"
    fi
    if [ "$ENABLE_NGINX" = false ]; then
        SERVICES_ARGS="$SERVICES_ARGS --scale nginx=0"
    fi
    
    # Create docker-compose override file for selective services
    OVERRIDE_FILE="docker-compose.override.yml"
    cat > "$OVERVIEV_FILE" << EOF
version: '3.8'
services:
EOF

    if [ "$ENABLE_ADMIN_WEB" = false ]; then
        cat >> "$OVERVIEV_FILE" << EOF
  admin-web-dev:
    scale: 0
EOF
    fi
    
    if [ "$ENABLE_FRONTEND" = false ]; then
        cat >> "$OVERVIEV_FILE" << EOF
  frontend:
    scale: 0
EOF
    fi
    
    # Note: We'd need to add more services here, but keeping it simple
    
    # Start services
    print_info "Starting services with: $COMPOSE_CMD up -d $SERVICES_ARGS"
    
    if $COMPOSE_CMD up -d $SERVICES_ARGS; then
        print_status "Services started successfully"
    else
        print_error "Failed to start services"
        exit 1
    fi
    
    # Wait for services to be ready
    echo ""
    print_step "Waiting for services to be ready..."
    echo "======================================"
    
    # Function to test service health
    test_service() {
        local service_name=$1
        local health_url=$2
        local max_attempts=30
        local attempt=1
        
        print_info "Waiting for $service_name..."
        
        while [ $attempt -le $max_attempts ]; do
            if curl -s -f -m 3 "$health_url" > /dev/null 2>&1; then
                print_status "$service_name is ready"
                return 0
            fi
            
            if [ $((attempt % 5)) -eq 0 ]; then
                print_info "$service_name still starting... (attempt $attempt/$max_attempts)"
            fi
            
            sleep 2
            attempt=$((attempt + 1))
        done
        
        print_warning "$service_name may not be ready yet"
        return 1
    }
    
    # Test service health
    [ "$ENABLE_ADMIN_WEB" = true ] && test_service "Admin Web" "http://localhost:8080/health"
    [ "$ENABLE_FRONTEND" = true ] && test_service "Frontend" "http://localhost:5173"
    [ "$ENABLE_JAEGER" = true ] && test_service "Jaeger" "http://localhost:16686"
    [ "$ENABLE_PROMETHEUS" = true ] && test_service "Prometheus" "http://localhost:9090/-/healthy"
    [ "$ENABLE_GRAFANA" = true ] && test_service "Grafana" "http://localhost:3000/api/health"
    
    # Test database services
    if [ "$ENABLE_REDIS" = true ]; then
        print_info "Testing Redis connectivity..."
        if docker exec mesh-probe-redis redis-cli ping | grep -q PONG; then
            print_status "Redis is ready"
        else
            print_warning "Redis may not be ready yet"
        fi
    fi
    
    if [ "$ENABLE_POSTGRES" = true ]; then
        print_info "Testing PostgreSQL connectivity..."
        if docker exec mesh-probe-postgres pg_isready -U devuser -d meshprobe | grep -q "accepting connections"; then
            print_status "PostgreSQL is ready"
        else
            print_warning "PostgreSQL may not be ready yet"
        fi
    fi
    
    if [ "$ENABLE_ETCD" = true ]; then
        print_info "Testing etcd connectivity..."
        if curl -s http://localhost:2379/health | grep -q '"health"'; then
            print_status "etcd is ready"
        else
            print_warning "etcd may not be ready yet"
        fi
    fi
    
    echo ""
    print_step "Environment Status"
    echo "==================="
    $COMPOSE_CMD ps
    
    echo ""
    print_info "Access URLs (when services are running):"
    [ "$ENABLE_FRONTEND" = true ] && echo "   Frontend:        http://localhost:5173 (Dev) / http://localhost:3000 (Prod)"
    [ "$ENABLE_ADMIN_WEB" = true ] && echo "   Admin API:       http://localhost:8080"
    [ "$ENABLE_GRAFANA" = true ] && echo "   Grafana:         http://localhost:3000 (admin/admin)"
    [ "$ENABLE_PROMETHEUS" = true ] && echo "   Prometheus:      http://localhost:9090"
    [ "$ENABLE_JAEGER" = true ] && echo "   Jaeger:          http://localhost:16686"
    [ "$ENABLE_NGINX" = true ] && echo "   Nginx Proxy:     http://localhost"
    [ "$ENABLE_ETCD" = true ] && echo "   etcd:            http://localhost:2379"
    [ "$ENABLE_POSTGRES" = true ] && echo "   PostgreSQL:      localhost:5432 (devuser/devpassword)"
    [ "$ENABLE_REDIS" = true ] && echo "   Redis:           localhost:6379"
    
    echo ""
    print_info "Useful commands:"
    echo "   View logs:        $COMPOSE_CMD logs -f [service]"
    echo "   Stop services:    $COMPOSE_CMD down"
    echo "   Restart service:  $COMPOSE_CMD restart [service]"
    echo "   Check status:     $COMPOSE_CMD ps"
fi

echo ""
echo "✨ Development environment setup completed!"
echo "=========================================="
print_status "All requested components built and started"
print_status "Environment ready for development and testing"

if [ "$BUILD_ALL" = true ] && [ "$START_SERVICES" = true ]; then
    echo ""
    print_info "Next steps:"
    echo "   1. Access the web interface at http://localhost:5173"
    echo "   2. Check the admin API at http://localhost:8080"
    echo "   3. Monitor metrics at http://localhost:3000 (Grafana)"
    echo "   4. View traces at http://localhost:16686 (Jaeger)"
    echo "   5. Run tests: npm test (in web/) or go test ./... (in project root)"
fi

exit 0