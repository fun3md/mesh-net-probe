#!/bin/bash

# Mesh Probe System - Development Environment Helper
# Focus: simple, reliable dev workflow aligned with:
# - docs/docker-compose/full-stack.yml
# - docs/docker-compose/development.yml

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_DIR="${ROOT_DIR}/docs/docker-compose"

FULL_STACK_YML="${COMPOSE_DIR}/full-stack.yml"
DEV_YML="${COMPOSE_DIR}/development.yml"
DEV_CONFIG="${ROOT_DIR}/docs/example-config/development.json"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log()       { echo -e "${BLUE}ℹ${NC} $*"; }
ok()        { echo -e "${GREEN}✔${NC} $*"; }
warn()      { echo -e "${YELLOW}⚠${NC} $*"; }
err()       { echo -e "${RED}✘${NC} $*"; }

usage() {
  cat <<EOF
Mesh Probe System - Development Environment

Usage:
  $(basename "$0") up [minimal|stack] [--build]   # Start dev services; optionally build images first
  $(basename "$0") down                           # Stop and remove dev services
  $(basename "$0") ps                             # List dev containers
  $(basename "$0") logs [service]                 # Tail logs (default: all)
  $(basename "$0") config                         # Validate compose + config
  $(basename "$0") build                          # Build Go binaries, frontend, and Docker images
  $(basename "$0") help                           # Show this help

Modes:
  minimal  - admin-web + frontend + core infra from development.yml
  stack    - full local stack from full-stack.yml

Notes:
  - This script is DEV-ONLY; it does not manage production.yml.
  - It assumes Docker and docker compose (v2) are available.
  - 'build' uses Dockerfile.multi-platform and web/Dockerfile: keep those in sync.
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "Missing required command: $1"
    exit 1
  fi
}

detect_compose() {
  if docker compose version >/dev/null 2>&1; then
    echo "docker compose"
  elif command -v docker-compose >/dev/null 2>&1; then
    echo "docker-compose"
  else
    err "Docker Compose v2 ('docker compose') or v1 ('docker-compose') is required."
    exit 1
  fi
}

validate_files() {
  local missing=0

  if [ ! -f "${FULL_STACK_YML}" ]; then
    err "Missing ${FULL_STACK_YML} (required for 'stack' mode)"
    missing=1
  fi

  if [ ! -f "${DEV_YML}" ]; then
    err "Missing ${DEV_YML} (required for 'minimal' mode)"
    missing=1
  fi

  if [ ! -f "${DEV_CONFIG}" ]; then
    err "Missing dev config ${DEV_CONFIG}"
    warn "Generate or copy example-config/development.json before running the stack."
    missing=1
  fi

  if [ "${missing}" -eq 0 ]; then
    ok "Compose and config files present."
  else
    exit 1
  fi
}

validate_compose() {
  local COMPOSE_BIN
  COMPOSE_BIN="$(detect_compose)"

  validate_files

  log "Validating ${FULL_STACK_YML}"
  if ${COMPOSE_BIN} -f "${FULL_STACK_YML}" config >/dev/null 2>&1; then
    ok "full-stack.yml syntax is valid."
  else
    err "full-stack.yml has errors."
    exit 1
  fi

  log "Validating ${DEV_YML}"
  if ${COMPOSE_BIN} -f "${DEV_YML}" config >/dev/null 2>&1; then
    ok "development.yml syntax is valid."
  else
    err "development.yml has errors."
    exit 1
  fi
}

cmd_up() {
  local mode="stack"
  local build=false

  # Parse args: first non-flag is mode, --build toggles build
  while [ $# -gt 0 ]; do
    case "$1" in
      minimal|stack)
        mode="$1"
        ;;
      --build)
        build=true
        ;;
      *)
        err "Unknown up option: $1"
        usage
        exit 1
        ;;
    esac
    shift
  done

  local COMPOSE_BIN
  COMPOSE_BIN="$(detect_compose)"

  validate_files

  if [ "${build}" = true ]; then
    log "Running build before 'up ${mode}'..."
    cmd_build
  fi

  case "${mode}" in
    minimal)
      log "Starting minimal development stack (development.yml)..."
      ${COMPOSE_BIN} -f "${DEV_YML}" up -d
      ;;
    stack)
      log "Starting full development stack (full-stack.yml)..."
      ${COMPOSE_BIN} -f "${FULL_STACK_YML}" up -d
      ;;
    *)
      err "Unknown mode '${mode}'. Use 'minimal' or 'stack'."
      exit 1
      ;;
  esac

  ok "Services started."

  cat <<EOF

Access URLs (when running):
  Frontend (dev via full-stack):    http://localhost:5173
  Frontend (nginx/prod-like):       http://localhost:3001 or http://localhost (via nginx)
  Admin API:                        http://localhost:8080
  Grafana:                          http://localhost:3000  (admin/admin)
  Prometheus:                       http://localhost:9090
  Jaeger:                           http://localhost:16686

Use:
  $(basename "$0") logs [service]
  $(basename "$0") ps
  $(basename "$0") down
EOF
}

cmd_down() {
  local COMPOSE_BIN
  COMPOSE_BIN="$(detect_compose)"

  log "Stopping dev stack (full-stack.yml + development.yml if present)..."

  # Stop full-stack if exists
  if [ -f "${FULL_STACK_YML}" ]; then
    ${COMPOSE_BIN} -f "${FULL_STACK_YML}" down -v || true
  fi

  # Stop minimal/dev if exists
  if [ -f "${DEV_YML}" ]; then
    ${COMPOSE_BIN} -f "${DEV_YML}" down -v || true
  fi

  ok "Dev containers removed."
}

cmd_ps() {
  local COMPOSE_BIN
  COMPOSE_BIN="$(detect_compose)"

  if [ -f "${FULL_STACK_YML}" ]; then
    ${COMPOSE_BIN} -f "${FULL_STACK_YML}" ps
  fi
  if [ -f "${DEV_YML}" ]; then
    ${COMPOSE_BIN} -f "${DEV_YML}" ps
  fi
}

cmd_logs() {
  local service="${1:-}"
  local COMPOSE_BIN
  COMPOSE_BIN="$(detect_compose)"

  if [ -f "${FULL_STACK_YML}" ]; then
    if [ -n "${service}" ]; then
      ${COMPOSE_BIN} -f "${FULL_STACK_YML}" logs -f "${service}" || true
    else
      ${COMPOSE_BIN} -f "${FULL_STACK_YML}" logs -f
    fi
    return
  fi

  if [ -f "${DEV_YML}" ]; then
    if [ -n "${service}" ]; then
      ${COMPOSE_BIN} -f "${DEV_YML}" logs -f "${service}" || true
    else
      ${COMPOSE_BIN} -f "${DEV_YML}" logs -f
    fi
  fi
}

cmd_build() {
  local COMPOSE_BIN
  COMPOSE_BIN="$(detect_compose)"

  log "Building Mesh Probe dev components (Go, frontend, Docker images)..."

  # Build Go components (admin-web, probe) using multi-platform Dockerfile targets
  (
    cd "${ROOT_DIR}"
    require_cmd go || true

    log "Building admin-web Docker image (runtime-web target)..."
    ${COMPOSE_BIN} build admin-web || docker build \
      -f Dockerfile.multi-platform \
      --target runtime-web \
      -t mesh-probe:admin-web-dev .

    log "Building probe-cli Docker image (runtime-cli target)..."
    ${COMPOSE_BIN} build probe-cli || docker build \
      -f Dockerfile.multi-platform \
      --target runtime-cli \
      -t mesh-probe:cli-dev .
  )

  # Build frontend image
  (
    cd "${ROOT_DIR}/web"
    if command -v npm >/dev/null 2>&1; then
      if [ ! -d node_modules ]; then
        log "Installing frontend dependencies..."
        npm install
      fi
      log "Ensuring frontend builds successfully..."
      npm run build
    else
      warn "npm not found; skipping local frontend build check."
    fi

    log "Building frontend Docker image..."
    docker build -t mesh-probe:frontend-dev .
  )

  ok "Build completed. Images ready for compose up."
}

main() {
  require_cmd docker

  case "${1:-help}" in
    up)
      shift
      cmd_up "$@"
      ;;
    down)
      cmd_down
      ;;
    ps)
      cmd_ps
      ;;
    logs)
      shift || true
      cmd_logs "$@"
      ;;
    config)
      validate_compose
      ;;
    build)
      cmd_build
      ;;
    help|-h|--help)
      usage
      ;;
    *)
      err "Unknown command: ${1:-}"
      usage
      exit 1
      ;;
  esac
}

main "$@"