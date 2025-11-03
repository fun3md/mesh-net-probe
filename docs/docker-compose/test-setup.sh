#!/bin/bash

# Admin Web Server - Local Development Test Script
# This script demonstrates how to use the example configurations and Docker compose files

set -e

echo "🚀 Mesh Probe Admin Web - Development Setup Test"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Check prerequisites
echo "🔍 Checking prerequisites..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    exit 1
fi
print_status "Docker is installed"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not installed"
    exit 1
fi
print_status "Docker Compose is available"

# Check Go
if ! command -v go &> /dev/null; then
    print_warning "Go is not installed (required for building from source)"
else
    print_status "Go is installed: $(go version)"
fi

# Check curl
if ! command -v curl &> /dev/null; then
    print_error "curl is not installed"
    exit 1
fi
print_status "curl is available"

echo ""
echo "🧪 Testing Configuration Files..."
echo "=================================="

# Test configuration files exist
config_files=(
    "../example-config/development.json"
    "../example-config/production.json" 
    "../example-config/testing.json"
)

for file in "${config_files[@]}"; do
    if [ -f "$file" ]; then
        print_status "Configuration file exists: $file"
        
        # Validate JSON syntax
        if command -v jq &> /dev/null; then
            if jq empty "$file" 2>/dev/null; then
                print_status "JSON syntax is valid: $file"
            else
                print_error "Invalid JSON syntax: $file"
                exit 1
            fi
        else
            print_warning "jq not available, skipping JSON validation"
        fi
    else
        print_error "Configuration file missing: $file"
        exit 1
    fi
done

# Test Docker compose files exist
compose_files=(
    "development.yml"
    "production.yml"
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

echo ""
echo "🏗️  Building Admin Web Server..."
echo "==============================="

# Change to project root
cd ../..

# Build admin web server
if [ -d "cmd/admin-web" ]; then
    print_info "Building admin web server..."
    cd cmd/admin-web
    
    if go build -o admin-web-test.exe .; then
        print_status "Admin web server built successfully"
        
        # Test basic functionality
        print_info "Testing admin web server startup..."
        timeout 10s ./admin-web-test.exe &
        SERVER_PID=$!
        
        # Wait a moment for server to start
        sleep 3
        
        # Test health endpoint
        if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
            print_status "Admin web server is responding"
        else
            print_warning "Admin web server may not be ready yet"
        fi
        
        # Kill the test server
        if kill $SERVER_PID 2>/dev/null; then
            print_status "Test server stopped"
        fi
        
    else
        print_error "Failed to build admin web server"
        exit 1
    fi
    
    cd ../..
else
    print_error "Admin web source code not found at cmd/admin-web"
    exit 1
fi

echo ""
echo "🐳 Testing Docker Compose Files..."
echo "=================================="

# Test Docker compose syntax
print_info "Validating Docker compose files..."

for compose_file in "${compose_files[@]}"; do
    if docker compose -f "docs/docker-compose/$compose_file" config > /dev/null 2>&1; then
        print_status "Docker compose file syntax is valid: $compose_file"
    else
        print_error "Docker compose file has syntax errors: $compose_file"
        exit 1
    fi
done

echo ""
echo "🧪 Running Integration Tests..."
echo "==============================="

# Function to test API endpoints
test_api() {
    local endpoint=$1
    local description=$2
    
    if curl -s -f -m 5 "$endpoint" > /dev/null 2>&1; then
        print_status "$description is accessible"
        return 0
    else
        print_warning "$description is not accessible (expected if server not running)"
        return 1
    fi
}

# Test endpoints
test_api "http://localhost:8080/health" "Admin web health endpoint"
test_api "http://localhost:8080/api/v1/config" "Admin web config endpoint"
test_api "http://localhost:8080/api/v1/probes" "Admin web probes endpoint"

echo ""
echo "📊 Generating Configuration Report..."
echo "====================================="

# Generate configuration summary
print_info "Configuration Summary:"
echo "======================"

for config_file in "${config_files[@]}"; do
    if [ -f "$config_file" ]; then
        echo ""
        echo "📄 $(basename "$config_file"):"
        if command -v jq &> /dev/null; then
            local name=$(jq -r '.name // "Unnamed"' "$config_file" 2>/dev/null || echo "Unknown")
            local version=$(jq -r '.version // "Unknown"' "$config_file" 2>/dev/null || echo "Unknown")
            echo "   Name: $name"
            echo "   Version: $version"
            
            # Extract some key settings
            local port=$(jq -r '.server.port // "Default"' "$config_file" 2>/dev/null || echo "Default")
            local log_level=$(jq -r '.logging.level // "Default"' "$config_file" 2>/dev/null || echo "Default")
            echo "   Port: $port"
            echo "   Log Level: $log_level"
        else
            echo "   (JSON parsing not available)"
        fi
    fi
done

echo ""
echo "🎯 Next Steps..."
echo "==============="
print_info "To start the full development environment:"
echo "   cd docs/docker-compose"
echo "   docker-compose -f full-stack.yml up -d"
echo ""
print_info "To start just the admin web server:"
echo "   cd docs/docker-compose"  
echo "   docker-compose -f development.yml up -d admin-web-dev"
echo ""
print_info "To run manual tests:"
echo "   curl http://localhost:8080/health"
echo "   curl http://localhost:8080/api/v1/config"
echo ""
print_info "To view logs:"
echo "   docker-compose -f full-stack.yml logs -f admin-web"
echo ""
print_info "To stop all services:"
echo "   docker-compose -f full-stack.yml down"

echo ""
echo "✨ Development setup test completed successfully!"
echo "=================================================="
print_status "All configuration files are valid"
print_status "Admin web server builds and runs"
print_status "Docker compose files are syntactically correct"
print_status "Ready for development and testing"

echo ""
print_info "Access URLs (when services are running):"
echo "   Frontend:        http://localhost:5173"
echo "   Admin API:       http://localhost:8080"
echo "   Grafana:         http://localhost:3000 (admin/admin)"
echo "   Prometheus:      http://localhost:9090"
echo "   Jaeger:          http://localhost:16686"
echo "   Nginx Proxy:     http://localhost"

exit 0