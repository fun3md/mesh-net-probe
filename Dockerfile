# Single-platform Dockerfile for mesh probe system
# Optimized for deployment on Linux x86_64

# Multi-stage build for smaller final image
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies for cross-compilation
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" \
    -o mesh-probe \
    ./cmd/probe/main.go

# Runtime stage - minimal Alpine Linux image
FROM alpine:latest

# Install ca-certificates for HTTPS/TLS and curl for health checks
RUN apk --no-cache add ca-certificates curl

# Create non-root user for security
RUN addgroup -g 1001 probe && \
    adduser -D -s /bin/sh -u 1001 -G probe probe

# Create working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/mesh-probe /usr/local/bin/

# Copy configuration files if they exist
RUN mkdir -p /etc/mesh-probe && \
    cp /app/config.json /etc/mesh-probe/config.json 2>/dev/null || true

# Change ownership to non-root user
RUN chown -R probe:probe /app /etc/mesh-probe 2>/dev/null || true

# Switch to non-root user
USER probe

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV PATH=/usr/local/bin:$PATH
ENV MESH_PROBE_HOME=/app
ENV MESH_PROBE_CONFIG=/etc/mesh-probe/config.json

# Expose default ports
EXPOSE 8080 8081

# Default command
CMD ["mesh-probe"]

# Metadata labels
LABEL org.opencontainers.image.title="Mesh Probe System"
LABEL org.opencontainers.image.description="High-precision network measurement tool with distributed coordination"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.vendor="Mesh Net Probe"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/mesh-net-probe/probe"
LABEL org.opencontainers.image.architecture="amd64"
LABEL org.opencontainers.image.os="linux"
LABEL maintainer="mesh-net-probe@example.com"