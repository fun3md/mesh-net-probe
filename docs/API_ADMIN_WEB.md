# Mesh Probe Admin Web API (Preliminary)

Status: PRELIMINARY - subject to change during Phase 5.1/6 hardening.

This document describes the current Admin Web HTTP API exposed by the backend service.
It reflects the implemented behavior as of this commit and will be updated as Phase 5.1 tasks T093–T101 evolve.

Base URL:
- Default (example): http://localhost:8080/api

All examples assume JSON request/response unless stated otherwise.

## Authentication

The API uses JWT-based authentication middleware (demo-grade in current implementation).

- Protected endpoints require:
  - Authorization: Bearer &lt;token&gt;

Current implementation:
- A demo login handler exists, returning a simple JWT-like token.
- In production deployments, replace this with a real authentication backend and secure signing keys.

Example: Authenticate and obtain token

Request:
- POST /auth/login
- Body:
  - { "username": "admin", "password": "admin" }

Example (curl):

- curl -X POST http://localhost:8080/api/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"admin\"}"

Successful response (matches routes.go &amp; OpenAPI):

- {
-   "token": "demo-jwt-token-1730970000",
-   "user": {
-     "id": "1",
-     "username": "admin",
-     "role": "admin"
-   },
-   "expiresAt": 1731056400
- }

Use the returned token in subsequent calls:

- Authorization: Bearer demo-jwt-token-1730970000

NOTE:
- Current implementation is demo-grade and must be treated as insecure.
- Do not use the demo credentials or static secret in production.
- Future revisions will enforce proper JWT signing, rotation, and real identity providers.

## Roles and Access Control (Preliminary)

The following helpers are wired into routes:

- auth.JWT() - verifies JWT and injects user + role into context
- auth.RequireAdmin() - allows only admin
- auth.RequireOperator() - allows admin or operator

Effective behavior (current, see routes.go &amp; OpenAPI):

- /auth/*:
  - POST /auth/login is public
  - POST /auth/logout, GET /auth/me, POST /auth/refresh require JWT
- /config:
  - GET /config, GET /config/status require JWT (any valid role) in production wiring
  - POST/PUT/DELETE /config*, POST /config/propagate require operator/admin roles
- /probes:
  - Read endpoints (GET /probes, GET /probes/:id, GET /probes/:id/health) require JWT
  - Mutating endpoints (POST/PUT/DELETE) require operator/admin
  - Heartbeat and config-applied are intended for authenticated probe/service tokens

This RBAC model is preliminary and may change as real auth integration is completed.

## Response Wrapping vs Plain Objects

To align backend, contracts, and frontend:

- Plain objects:
  - Auth:
    - POST /auth/login → { token, user, expiresAt }
    - GET /auth/me → User
    - POST /auth/refresh → { token, expiresAt }
  - Configuration:
    - GET /config → Configuration
    - GET /config/status → ConfigStatus
  - Probes:
    - GET /probes/:id → Probe
    - GET /probes/:id/health → { probe_id, health, status, last_seen }
  - Monitoring:
    - GET /monitoring/dashboard → DashboardStats (snake_case fields)
    - GET /monitoring/health/summary → HealthSummary
    - GET /monitoring/stats → MonitoringStats
  - Alerts (single):
    - POST /monitoring/alerts → Alert

- Wrapped collections:
  - GET /probes → { "probes": Probe[] }
  - GET /measurements → { "measurements": MeasurementData[] }
  - GET /monitoring/alerts → { "alerts": Alert[] }

Frontend rules:
- API client (`web/src/services/api.ts`) is responsible for:
  - Reading wrapped lists (e.g. { probes: [...] }) and returning arrays.
  - Mapping snake_case backend fields into camelCase view models where desired.

## Endpoints

### 1. Auth

1. POST /auth/login

- Description:
  - Demo login endpoint; returns a JWT-like token for testing.
- Auth:
  - Public.
- Request:
  - { "username": "admin", "password": "admin" }
- Response:
  - 200:
    - { "token": string, "user": { "id": string, "username": string, "role": "admin" }, "expiresAt": number }
  - 401 on invalid credentials

2. POST /auth/logout

- Description:
  - Demo logout; no real token revocation.
- Auth:
  - Requires JWT.
- Response:
  - 200 { "message": "Logged out successfully" }

3. GET /auth/me

- Description:
  - Returns demo current user info.
- Auth:
  - Requires JWT.
- Response:
  - 200 User object (see OpenAPI contract)

4. POST /auth/refresh

- Description:
  - Returns a new demo token.
- Auth:
  - Requires JWT.
- Response:
  - 200 { "token": string, "expiresAt": number }

### 2. Configuration Management

All configuration endpoints are backed by config.Manager.

- GET /config
  - Returns the currently active configuration.
  - 200: Configuration (plain object)
  - 404: { "error": "configuration not available", "details": string }

- GET /config/status
  - Returns ManagerStatus: provider health, currentConfigID, lastUpdate, updateCount, healthScore.
  - 200: ConfigStatus
  - 500: { "error": "failed to get configuration status", "details": string }

- POST /config (PRELIMINARY)
  - Accepts a configuration payload (types.Configuration-compatible).
  - Triggers Manager.ReloadConfiguration.
  - 201: Configuration persisted in memory and reloaded successfully.
  - 202: { "message": "configuration accepted; propagation reported issues", "config": Configuration, "warning": string }

- PUT /config/:id (PRELIMINARY)
  - Updates configuration metadata and triggers reload.
  - 200: Updated Configuration
  - 400 / 500 on error

- DELETE /config/:id (PRELIMINARY)
  - Triggers reload after delete.
  - 204 on success

- POST /config/propagate
  - Forces reload from providers.
  - 200: { "message": string, "status": ConfigStatus }
  - 500 on failure

### 3. Probe Management

Backed by internal/monitoring.ProbeRegistry.

- GET /probes
  - 200: { "probes": [Probe, ...] }

- GET /probes/:id
  - 200: Probe
  - 404: { "error": "probe not found" }

- GET /probes/:id/health
  - 200: { "probe_id": string, "health": object, "status": string, "last_seen": timestamp }
  - 404: { "error": "probe or health status not found" }

- POST /probes
  - Request: Probe registration fields (id required, ip_address, platform, arch, etc.)
  - 201: { "message": "probe registered successfully", "probe": Probe }

- PUT /probes/:id
  - Partial update.
  - 200: Updated Probe

- DELETE /probes/:id
  - 204 on success

- POST /probes/:id/heartbeat
  - Records heartbeat, updates last_seen.
  - 200: { "message": "heartbeat received", "probe_id": string, "timestamp": timestamp }

- POST /probes/:id/config-applied
  - Request:
  -     - { "config_id": string, "config_version": number, "config_source": string, "applied_at"?: timestamp }
  - 200: { "message": "configuration state recorded", "probe": Probe }
  - 400 / 404 / 500 on error

### 4. Measurements and Monitoring

Measurement and monitoring endpoints remain demo/preliminary:

- GET /measurements → { "measurements": MeasurementData[] }
- GET /measurements/:id → MeasurementData or 404
- POST /measurements → MeasurementData (id/timestamp set by server)
- GET /measurements/statistics → MeasurementStatsResponse

- GET /monitoring/dashboard → DashboardStats
  - Fields: snake_case as in routes.go (total_probes, active_probes, etc.).
- GET /monitoring/health/summary → HealthSummary
- GET /monitoring/stats → MonitoringStats
- POST /monitoring/alerts → Alert
- GET /monitoring/alerts → { "alerts": Alert[] }

## Contract Source of Truth

- The canonical machine-readable contract is [`specs/001-mesh-probe-system/contracts/admin-web.openapi.yaml`](specs/001-mesh-probe-system/contracts/admin-web.openapi.yaml).
- The handlers in [`internal/web/api/routes.go`](internal/web/api/routes.go) are the implementation reference.
- Frontend types and api client in `web/src/types/index.ts` and `web/src/services/api.ts` must remain aligned with this contract.

Any future changes MUST:
- Update routes.go, the OpenAPI file, this document, and the frontend types/API together to avoid drift.

## Notes

- This documentation now reflects the aligned Phase 5.2.1 contract state.
- Endpoints marked PRELIMINARY may still evolve in storage semantics and auth strength, but shapes and wrapping rules are stable for the current phase.