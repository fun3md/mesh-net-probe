# Implementation Plan: Admin Web System

**Branch**: `admin-web-functionality` | **Date**: January 2026 | **Spec**: https://github.com/example/admin-web-spec
**Input**: Feature specification from `/specs/001-mesh-probe-system/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a production-ready Admin Web Interface (Phase 4/5.1) that uses the real Go backend as the single source of truth for:
- Authentication (JWT-like demo flow)
- Central configuration (config.Manager via /config, /config/status, /config/propagate)
- Probe registry and health (ProbeRegistry via /probes, /probes/:id/health, /probes/:id/heartbeat, /probes/:id/config-applied)
- Measurements and monitoring (/measurements*, /monitoring*)

This phase is part of the larger Distributed Mesh Probe System described in [`specs/001-mesh-probe-system/intialplan.md`](specs/001-mesh-probe-system/intialplan.md), which defines:
- A cross-platform ICMP probe engine with microsecond-level precision.
- Centralized configuration via etcd/Consul/file providers with sub-second propagation.
- A probe registry and mesh coordination layer for high-frequency, distributed measurements.
- OpenTelemetry-based telemetry export (metrics, logs, traces) for probes and admin web.
- A unified CLI + Admin Web UX, sharing authentication, configuration, and monitoring semantics.

The goal of this plan is to:
- Complete the Admin Web backend and SPA as a real, production-grade surface over the existing mesh probe capabilities.
- Ensure all Admin Web contracts, flows, and UX remain fully aligned with the mesh probe system architecture and quality constraints from the initial plan.

## Technical Context

**Language/Version**: Go 1.21+ (backend, mesh probe services), TypeScript 5.x + React 18 + Vite (frontend)  
**Primary Dependencies**: 
- Backend: Gin, `config.Manager`, `monitoring.Manager`, `ProbeRegistry`, OpenTelemetry SDK, WebSocket/HTTP upgrades, platform detection utilities.
- Frontend: axios, React Query or similar (optional), socket.io-client or native WebSocket client (for future phases), charting libs.
**Storage**: 
- Configuration: file/etcd/Consul providers via `config.Manager` per initial plan.
- Telemetry: exported via OTLP/OpenTelemetry to external collectors (Prometheus, Jaeger, etc.) as described in initial plan.
- Frontend: stateless; relies on backend APIs as source of truth.
**Testing**: 
- Go test suite (unit/integration/contract) for config, ICMP engine, mesh/monitoring, and admin-web APIs.
- Frontend: Jest/Vitest + React Testing Library for web components; Cypress/Playwright for E2E.
**Target Platform**: 
- Backend and probes: Linux/macOS/Windows (x64/ARM64), containers.
- Frontend: browser, served via admin-web or separate static host.
**Project Type**: Multi-component system (Go CLI probes + admin-web backend + React SPA), with Admin Web as orchestration and observability layer.
**Performance Goals**: 
- Backend/admin-web: <50ms p95 for core config/probe operations; efficient aggregation endpoints for dashboards.
- Probes: microsecond precision ICMP measurements; support 100+ concurrent probes without degradation.
**Constraints**: 
- Must respect constitution quality gates (linting, tests, security).
- Real-time UX MUST NOT rely on demo-only in-memory state.
- Auth initially demo-grade but aligned with unified security model; clear seam for real IdP.
- Cross-platform behavior and telemetry guarantees inherited from the initial mesh probe plan.
**Scale/Scope**: 
- O(10^2) probes, O(10^3-10^4) measurements visible via UI.
- Designed to integrate into the broader mesh coordination and telemetry architecture.

## Constitution Check

### Code Quality Excellence Requirements
- [x] Go code follows strict quality standards with comprehensive error handling (routes.go already structured; extend with consistent error envelopes for web API).
- [x] Clean architecture with separation of concerns implemented (handlers delegate to config.Manager, ProbeRegistry, monitoring.Manager, ICMP/mesh components where relevant).
- [ ] All public APIs and complex logic have comprehensive documentation (sync docs/API_ADMIN_WEB.md and contracts with both initial mesh probe APIs and admin-web routes).
- [x] golangci-lint configuration enforces consistent coding style.
- [x] Memory-safe practices prevent leaks and race conditions.

### Testing Standards Compliance
- [ ] MANDATORY test coverage: 80% minimum for all packages (backend, probes, admin-web, and shared types).
- [x] Unit tests cover core config/monitoring/ICMP components.
- [x] Integration tests validate /config, /probes via Phase 5.1 tasks (config_and_probe_admin_test.go).
- [ ] Cross-platform testing plan includes Linux, macOS, Windows (ensure admin-web integration tests exercise mesh probe flows).
- [ ] Performance benchmarks included for measurement accuracy and latency (ICMP engine + selected admin-web endpoints).
- [ ] End-to-end tests simulate real operational scenarios:
      - CLI probes + admin-web + config propagation + telemetry export.

### User Experience Consistency Requirements
- [x] CLI interface consistency (unchanged; remains primary tool for direct probe operations).
- [ ] Web UI mirrors backend/mesh semantics with clear, actionable errors and statuses.
- [x] Configuration is intuitive with sensible defaults and validation at backend.
- [ ] Error messages are actionable with suggested resolutions in frontend.
- [ ] Progress indicators/reporting for async actions (propagate config, probe heartbeats, measurement streams).
- [x] Text-based and JSON outputs from backend APIs remain stable for both CLI and web.

### Performance Requirements
- [x] Measurement precision and probe performance handled in backend ICMP/mesh engine.
- [x] Resource consumption optimized for long-running probe deployments.
- [ ] Frontend must avoid polling storms; prefer aggregated endpoints (/monitoring/dashboard, /monitoring/health/summary) and eventual WebSocket streams.

### Cross-Platform Compatibility
- [x] Identical behavior maintained across target platforms at backend and probes.
- [x] Platform-specific optimizations preserve functional equivalence.
- [x] Architecture support verified for target architectures (x64, ARM64).
- [x] Platform detection/capability checking reused consistently (admin-web surfaces these states).

> Gate Evaluation: Admin web Phase 4/5.1 work proceeds as a first-class surface over the distributed mesh probe system. Remaining NEEDS CLARIFICATION items must be resolved in research.md and contracts before finalizing E2E flows.

## Project Structure

### Documentation (this feature within mesh probe system)

```text
specs/001-mesh-probe-system/
├── plan.md                 # This integrated admin-web + mesh system plan
├── intialplan.md           # Original distributed mesh probe system plan (authoritative system context)
├── research.md
├── data-model.md
├── quickstart.md
├── admin-web-interface.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
.
├── cmd/
│   ├── probe/               # Core Mesh Probe CLI (ICMP, traceroute, etc.)
│   └── admin-web/           # Admin web backend service
├── internal/
│   ├── icmp/                # ICMP measurement engine
│   ├── config/              # Centralized configuration management
│   ├── mesh/                # Mesh coordination and aggregation (planned/implemented per initial plan)
│   ├── telemetry/           # OpenTelemetry integration
│   ├── monitoring/          # Probe registry and health monitoring
│   ├── platform/            # Cross-platform utilities
│   ├── web/                 # Web interface services (REST, WebSocket, auth)
│   │   ├── api/
│   │   ├── auth/
│   │   └── websocket/
├── pkg/
│   └── types/               # Shared data types and contracts
├── web/                     # React TypeScript SPA for admin interface
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── services/
│   │   ├── hooks/
│   │   └── types/
└── tests/
    ├── unit/
    ├── integration/
    ├── contract/
    ├── performance/
    ├── security/
    └── web/                 # To include admin web contracts/E2E tests
```

**Structure Decision**: Single repository containing:
- Core probe CLI and mesh/ICMP/telemetry components (from initial plan).
- Admin-web backend exposing configuration, probe registry, monitoring, and measurements.
- React SPA as thin client over these APIs.
This plan builds on and does not supersede the system-level requirements from [`intialplan.md`](specs/001-mesh-probe-system/intialplan.md).

## Phase 4 / 5.1: Admin Web Backend + Frontend Real Functionality Plan

### 1. Backend: Solidify Admin Web API Contracts

Use internal/web/api/routes.go as the authoritative contract and close gaps between types and responses:

1. Auth:
   - POST /auth/login → { token: string, user: { id, username, role }, expiresAt } (already implemented).
   - GET /auth/me → returns current user (demo static; treat as admin).
   - POST /auth/refresh → returns { token, expiresAt }.
   - Action: Document these shapes in contracts/auth.openapi.yaml and update frontend types accordingly.

2. Config:
   - GET /config → current effective configuration (types.Configuration).
   - GET /config/status → manager status (provider health, last reload).
   - POST /config, PUT /config/:id, DELETE /config/:id, POST /config/propagate → drive configManager.ReloadConfiguration.
   - Action: Model these as admin-only operations in OpenAPI and ensure handlers consistently return JSON with error/message fields for UI.

3. Probes:
   - GET /probes → { probes: Probe[] } mapped from monitoring.ProbeRegistry.
   - GET /probes/:id, DELETE /probes/:id, etc.
   - POST /probes/:id/heartbeat, POST /probes/:id/config-applied → used by probes for status/rollout tracking.
   - Action: Define normalized probe DTO in contracts to match web/src/types.Probe (align field names: ipAddress vs IPAddress, lastSeen, status, health).

4. Measurements:
   - /measurements, /measurements/:id, /measurements/statistics, /measurements/stream/:probe_id.
   - Currently demo/in-memory; Phase 4+ treats them as optional but exposes consistent schema for UI charts.

5. Monitoring:
   - GET /monitoring/dashboard → DashboardStats.
   - GET /monitoring/health/summary → health overview.
   - GET/POST /monitoring/alerts → Alert list/create.

Deliverables:
- contracts/admin-web.openapi.yaml aligned with routes.go semantics.
- docs/API_ADMIN_WEB.md updated to exactly match handlers.

### 2. Frontend: Refactor to Real API-Driven Flows

Key problems today:
- useAuth.ts assumes /auth/me returns a user bound to stored token, but login currently stores token and user from response without verifying persistence semantics.
- api.ts types (AuthToken, LoginResponse, DashboardStats, Alert etc.) do not perfectly match backend responses.
- Dashboard and other pages use mostly static/demo components; they are not wired to apiService.

Plan:

1. Auth wiring:
   - Update apiService.login to type its response: { token: string; user: { ... }; expiresAt: number } to match handleLogin.
   - Ensure apiService.setAuthToken stores raw token and sets Authorization: Bearer <token>.
   - Update useAuth.ts to:
     - On mount, read auth_token and attempt GET /auth/me.
     - Handle 401 by clearing auth state.
     - Expose loading state; App.tsx should render a loading screen until auth state resolved, then gate routes.

2. Routing and guards:
   - Keep App.tsx router structure, but:
     - Add explicit /login route.
     - Show Login page when !isAuthenticated (instead of replacing whole app without Router).
     - Optionally use a ProtectedRoute wrapper to centralize guard logic.

3. Dashboard page wiring:
   - Replace static placeholders in Dashboard.tsx with real calls:
     - useEffect to call apiService.getDashboardStats and apiService.getHealthSummary.
     - Pass data into StatCards, ProbesTable, LatencyChart, PacketLossChart.
   - Add error/loading states consistent with Constitution UX requirements.

4. Probes and Config pages:
   - Targets.tsx / NetworkConfiguration.tsx / VisualConfigurator.tsx / OpenTelemetryConfiguration.tsx:
     - Implement CRUD flows using apiService:
       - Config list/create/update/delete → /config*, plus /config/propagate.
       - Probe list/details → /probes, /probes/:id, /probes/:id/health.
     - Add optimistic UI or clear status messages.

5. WebSocket / real-time:
   - web/src/services/websocket.ts currently assumes socket.io server; backend exposes HTTP/WebSocket endpoints differently (NEEDS CLARIFICATION).
   - Interim Phase 4 plan:
     - Use polling for /monitoring/dashboard and /monitoring/alerts (e.g., 5-10s interval).
     - Gate WebSocketService behind feature flag and update once internal/web/websocket/service.go is aligned (Phase 5+).
   - Mark current socket.io usage as experimental; do not make it a hard dependency for core flows.

### 3. Alignment Tasks (Backend + Frontend)

Define concrete refactor tasks (to be mirrored into tasks.md):

- Align types:
  - Update web/src/types/index.ts to match backend responses for:
    - Probe (fields from monitoring.Probe).
    - Alert (fields from Alert in routes.go).
    - DashboardStats (fields from DashboardStats in routes.go).
  - Where backend uses snake_case, normalize via apiService mapping.

- Align apiService methods:
  - Ensure paths and shapes match internal/web/api/routes.go handlers exactly.
  - Add wrapper methods for:
    - getDashboardStats(), getHealthSummary(), getMonitoringStats(), getAlerts(), createAlert().
    - getConfigStatus(), propagateCurrentConfig(), etc.
  - Implement consistent error handling/logging for debugging.

- App initialization:
  - On app start: auth check → fetch basic health/dashboard data.

### 4. Tests and Observability Hooks

- Backend:
  - Extend tests/contract/config_and_probe_admin_test.go to:
    - Assert /auth/login + /auth/me flow shape.
    - Assert /config/status and /probes/:id/config-applied behavior.
  - Add tests/security cases ensuring auth middleware enforced on admin routes.

- Frontend:
  - Add lightweight contract tests (TypeScript) to validate that:
    - apiService response mappers align with OpenAPI schemas.
    - useAuth handles token lifecycle correctly.

- E2E (Phase 4 gate):
  - Scenario: login via UI → load dashboard from real backend → list probes from ProbeRegistry → show health summary.
  - Scenario: create/update configuration via UI → POST/PUT /config → trigger /config/propagate → verify status is surfaced in UI.

- Observability:
  - Add logging in backend handlers around auth, config, probes to aid debugging.
  - In frontend, centralize API error logging with correlation IDs (if available in responses).

## Gate Re-evaluation (Post-Plan)

Once:
- contracts/admin-web.openapi.yaml is defined,
- apiService and types are aligned,
- Dashboard/Config/Probes pages are wired to these endpoints,
- and basic E2E is in place,

then:
- Constitutional gates for UX, testing, and cross-platform for Phase 4/5.1 are satisfied for the admin web surface.
