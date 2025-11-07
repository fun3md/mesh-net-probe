# Mesh Probe Admin Web API (Preliminary)

Status: PRELIMINARY - subject to change during Phase 5.1/6 hardening.

This document describes the current Admin Web HTTP API exposed by the backend service.
It reflects the implemented behavior as of this commit and will be updated as Phase 5.1 tasks T093–T101 evolve.

Base URL:
- Default (example): http://localhost:8080/api

All examples assume JSON request/response unless stated otherwise.

## Authentication

The API uses JWT-based authentication middleware.

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

Successful response (example):

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

Effective behavior (current):

- /auth/*: public for login; logout/me/refresh require JWT
- /config (read): requires authenticated user (any valid role)
- /config/status: requires authenticated user (any valid role)
- /config (write/propagate): requires operator/admin as detailed below
- /probes:
  - Read endpoints: require authenticated user
  - Mutating endpoints: require operator/admin
  - Heartbeat / config-applied: require JWT; intended for probe service accounts

This RBAC model is preliminary and may change as real auth integration is completed.

## Endpoints

### 1. Auth

1. POST /auth/login (PRELIMINARY)

- Description:
  - Demo login endpoint; returns a JWT-like token for testing.
- Auth:
  - Public.
- Request:
  - { "username": "admin", "password": "admin" }
- Response:
  - 200 with token and user info on success
  - 401 on invalid credentials

2. POST /auth/logout

- Description:
  - Demo logout; no real token revocation.
- Auth:
  - Requires JWT.
- Response:
  - 200 on success

3. GET /auth/me

- Description:
  - Returns demo current user info from JWT context.
- Auth:
  - Requires JWT.

4. POST /auth/refresh

- Description:
  - Returns a new demo token.
- Auth:
  - Requires JWT.

All auth endpoints are preliminary and not production-secure.

### 2. Configuration Management

All configuration endpoints are now conceptually backed by config.Manager.
The current implementation focuses on read/status exposure and controlled reload triggers.
Direct writes are placeholders, as configuration is expected to be maintained via providers (file, etcd, Consul).

1. GET /config

- Description:
  - Returns the currently active configuration from config.Manager.
- Auth:
  - Requires JWT (any valid role).
- Response:
  - 200 with types.Configuration
  - 404 if no configuration is available

2. GET /config/status

- Description:
  - Returns ManagerStatus including:
    - Sources: provider health and priority
    - CurrentConfigID
    - LastUpdate
    - UpdateCount
    - HealthScore
- Auth:
  - Requires JWT (any valid role).
- Response:
  - 200 with status
  - 500 on internal errors

3. POST /config (PRELIMINARY)

- Description:
  - Accepts a configuration payload and attempts to trigger a reload via config.Manager.ReloadConfiguration.
  - Intended as a façade over provider-backed configuration; not a direct persistent write.
- Auth:
  - Requires JWT + role: operator or admin.
- Response:
  - 201 / 202 on acceptance
  - Includes information/warnings if reload fails

4. PUT /config/:id (PRELIMINARY)

- Description:
  - Updates configuration metadata and triggers Manager.ReloadConfiguration.
  - Implementation is a compatibility shim; backing stores remain authoritative.
- Auth:
  - Requires JWT + role: operator or admin.

5. DELETE /config/:id (PRELIMINARY)

- Description:
  - Placeholder for deleting a configuration via providers and reloading.
- Auth:
  - Requires JWT + role: admin.

6. POST /config/propagate

- Description:
  - Forces a reload from configured providers and returns updated /config/status.
- Auth:
  - Requires JWT + role: operator or admin.

IMPORTANT:
- All write-style configuration endpoints are PRELIMINARY.
- For production, write through the configured providers (file/etcd/Consul); these endpoints will evolve into safe, auditable operations.

### 3. Probe Management

Probe state is now backed by internal/monitoring.ProbeRegistry as the single authoritative registry.

1. GET /probes

- Description:
  - List all registered probes.
- Auth:
  - Requires JWT.
- Response:
  - 200 { "probes": [ Probe ] }

2. GET /probes/:id

- Description:
  - Get details for a specific probe.
- Auth:
  - Requires JWT.

3. GET /probes/:id/health

- Description:
  - Returns health information if available for a probe.
- Auth:
  - Requires JWT.

4. POST /probes

- Description:
  - Register a new probe.
- Auth:
  - Requires JWT + role: operator or admin.
- Request (example):
  - {
  -   "id": "probe-1",
  -   "name": "edge-probe-1",
  -   "version": "v1.0.0",
  -   "platform": "linux",
  -   "arch": "amd64",
  -   "ip_address": "192.168.1.10",
  -   "tags": ["edge","dc1"],
  -   "metadata": {"env": "dev"}
  - }

5. PUT /probes/:id

- Description:
  - Update probe metadata (name/version/platform/arch/ip/tags/metadata).
- Auth:
  - Requires JWT + role: operator or admin.

6. DELETE /probes/:id

- Description:
  - Unregister a probe from the registry.
- Auth:
  - Requires JWT + role: admin.

7. POST /probes/:id/heartbeat

- Description:
  - Record a probe heartbeat; updates last_seen and status.
- Auth:
  - Requires JWT.
- Intended usage:
  - Probes or trusted agents call periodically with a service-token.

8. POST /probes/:id/config-applied

- Description:
  - Probes report which configuration they have successfully applied.
  - Updates ProbeRegistry with:
    - ConfigID, ConfigVersion, ConfigSource, ConfigAppliedAt
- Auth:
  - Requires JWT.
- Request:
  - {
  -   "config_id": "central-config-v3",
  -   "config_version": 3,
  -   "config_source": "etcd",
  -   "applied_at": "2025-11-07T08:55:00Z"
  - }
- Response:
  - 200 on success with updated probe
  - 404 if probe does not exist

NOTE:
- This endpoint is central for rollout tracking and is part of Phase 5.1 (T097).
- Semantics are PRELIMINARY but aligned with the new registry fields.

### 4. Measurements and Monitoring (Demo / Preliminary)

The following endpoints still rely on in-memory/demo data structures and are not yet production-grade.

- /measurements
  - GET /measurements
  - GET /measurements/:id
  - POST /measurements
  - GET /measurements/stream/:probe_id
  - GET /measurements/statistics

- /monitoring
  - GET /monitoring/dashboard
    - Uses ProbeRegistry counts plus demo metrics.
  - GET /monitoring/health/summary
    - Summarized health info (currently partly demo).
  - GET /monitoring/stats
    - Demo monitoring stats.
  - POST /monitoring/alerts
  - GET /monitoring/alerts

All measurement and monitoring endpoints are PRELIMINARY and will be aligned with real telemetry, persistence, and auth in future phases.

## Authentication Examples

1. Login and use token (Windows cmd-style curl)

- curl -X POST http://localhost:8080/api/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"admin\"}"

Copy the token from the response and use:

- set TOKEN=demo-jwt-token-1730970000
- curl http://localhost:8080/api/config ^
  -H "Authorization: Bearer %TOKEN%"

2. Protected write example (operator/admin)

- curl -X POST http://localhost:8080/api/probes ^
  -H "Authorization: Bearer %TOKEN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"id\":\"probe-1\",\"platform\":\"linux\",\"arch\":\"amd64\"}"

3. Probe reporting applied configuration

- curl -X POST http://localhost:8080/api/probes/probe-1/config-applied ^
  -H "Authorization: Bearer %TOKEN%" ^
  -H "Content-Type: application/json" ^
  -d "{\"config_id\":\"central-config-v3\",\"config_version\":3,\"config_source\":\"etcd\"}"

## Notes

- This documentation reflects current implementation, not the final design.
- Endpoints marked PRELIMINARY may change:
  - Request/response schemas
  - Required roles
  - Backing storage semantics
- For production:
  - Replace demo auth with real JWT signing and validation.
  - Use config providers (file/etcd/Consul) as primary write path.
  - Harden all monitoring/measurement endpoints with proper RBAC.