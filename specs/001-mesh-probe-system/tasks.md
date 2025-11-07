---

description: "Task list template for feature implementation"
---

# Tasks: Distributed Mesh Probe System

**Input**: Design documents from `/specs/001-mesh-probe-system/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Constitutional Compliance**: All tasks must ensure adherence to established principles:
- Code Quality Excellence (Go standards, error handling, documentation)
- Comprehensive Testing Standards (80% coverage, cross-platform, performance)
- User Experience Consistency (CLI patterns, output formats, error messages)
- Performance Requirements (optimization, resource management)
- Cross-Platform Compatibility (multi-OS, multi-arch consistency)

**Tests**: The examples below include test tasks. Tests are OPTIONAL - only include them if explicitly requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `src/`, `tests/` at repository root
- **Web app**: `backend/src/`, `frontend/src/`
- **Mobile**: `api/src/`, `ios/src/` or `android/src/`
- Paths shown below assume single project - adjust based on plan.md structure

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create project structure per implementation plan in src/, tests/, docs/ directories
- [x] T002 [P] Initialize Go 1.21+ project with go.mod and essential dependencies
- [x] T003 [P] Configure golangci-lint and formatting tools per constitution requirements
- [x] T004 [P] Setup constitutional compliance checks (linting, formatting, static analysis)
- [x] T005 [P] Configure cross-platform testing framework for Linux, macOS, Windows
- [x] T006 [P] Setup performance benchmarking infrastructure for microsecond precision testing
- [x] T007 [P] Create basic Git repository with .gitignore for Go projects

---

## Phase 1.5: Constitutional Compliance Framework

**Purpose**: Establish mandatory quality gates aligned with project principles

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T008 Configure golangci-lint for Go code quality standards per constitution
- [x] T009 Setup test coverage reporting (80% minimum requirement)
- [x] T010 [P] Configure cross-platform CI/CD pipeline testing for Linux, macOS, Windows
- [x] T011 [P] Setup performance profiling tools for measurement-critical code paths
- [x] T012 Create standardized error handling and logging framework
- [x] T013 [P] Establish CLI consistency patterns and validation framework
- [x] T014 Configure automated security and vulnerability scanning

**Checkpoint**: ✅ Constitutional compliance framework ready - user story implementation can now begin with quality gates enforced

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T015 Create shared data types and contracts in pkg/types/ package
- [x] T016 [P] Implement platform detection and capability checking utilities in internal/platform/
- [x] T017 [P] Setup OpenTelemetry SDK with OTLP/gRPC exporter for metrics and tracing
- [x] T018 [P] Create configuration management interface with etcd and Consul providers
- [x] T019 Implement basic ICMP packet handling infrastructure using golang.org/x/net/icmp
- [x] T020 Create Docker multi-platform build configuration for linux/amd64 and linux/arm64
- [x] T021 Setup basic error handling and structured logging infrastructure
- [x] T022 Create configuration loading and validation framework

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Single Probe Deployment and ICMP Measurement (Priority: P1) 🎯 MVP

**Goal**: Network operator can deploy a single probe container and immediately begin collecting high-resolution ICMP measurements

**Independent Test**: Deploy one probe container, configure with target IP addresses, verify microsecond-precision measurement data collection

### Tests for User Story 1 (OPTIONAL - only if tests requested) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T023 [P] [US1] Contract test for ICMP measurement interface in tests/contract/test_icmp_interface.go
- [x] T024 [P] [US1] Integration test for single probe deployment in tests/integration/test_single_probe.go  
- [x] T025 [US1] Performance benchmark test for microsecond precision in tests/performance/test_precision.go

### Implementation for User Story 1

- [x] T026 [P] [US1] Create Probe Instance model in pkg/types/probe.go
- [x] T027 [P] [US1] Create Measurement Data model in pkg/types/measurement.go
- [x] T028 [P] [US1] Create Network Target model in pkg/types/target.go
- [x] T029 [US1] Implement ICMP measurement engine in internal/icmp/engine.go
- [x] T030 [US1] Implement timing precision module with microsecond accuracy in internal/icmp/timing.go
- [x] T031 [US1] Implement OpenTelemetry metrics exporter in internal/telemetry/metrics.go
- [x] T032 [US1] Create main CLI application in cmd/probe/main.go
- [x] T033 [US1] Implement CLI command structure with consistent patterns in cmd/probe/
- [x] T034 [US1] Implement configuration file loading in internal/config/loader.go
- [x] T035 [US1] Create Docker configuration for single-platform deployment

**Checkpoint**: ✅ User Story 1 is now fully functional and testable independently - All 10 tasks completed (100%)

---

## Phase 4: User Story 2 - Cross-Platform Probe Deployment (Priority: P2)

**Goal**: DevOps engineer can deploy identical probe functionality across Linux, macOS, Windows on x64 and ARM architectures

**Independent Test**: Deploy probe containers on different platform combinations, verify identical measurement results for same network paths

### Tests for User Story 2 (OPTIONAL - only if tests requested) ⚠️

- [x] T036 [P] [US2] Cross-platform compatibility tests in tests/integration/test_cross_platform.go
- [x] T037 [P] [US2] Architecture consistency tests for x64 vs ARM64 in tests/integration/test_arch_consistency.go

### Implementation for User Story 2

- [x] T038 [P] [US2] Enhance platform detection with architecture-specific optimizations in internal/platform/detection.go
- [x] T039 [US2] Implement cross-platform ICMP timing adjustments in internal/icmp/platform_timing.go
- [x] T040 [US2] Create platform-specific configuration handling in internal/config/platform.go
- [x] T041 [US2] Update Docker multi-platform build configuration for consistent deployment
- [x] T042 [US2] Implement platform validation and capability checking in internal/platform/validation.go
- [x] T043 [US2] Add cross-platform testing infrastructure in tests/cross_platform/

**Checkpoint**: ✅ User Stories 1 AND 2 both work independently across all platforms - All 7 tasks completed (100%)

---

## Phase 5: User Story 3 - Centralized Configuration Management (Priority: P3)

**Goal**: System administrator can manage probe configurations from a central datasource, enabling scaling across large numbers of deployed probes

**Independent Test**: Deploy multiple probes, update configuration through central datasource, verify all probes reflect changes consistently

### Tests for User Story 3 (OPTIONAL - only if tests requested) ⚠️

- [x] T044 [P] [US3] Configuration management contract tests in tests/contract/test_config_interface.go
- [x] T045 [US3] Centralized config integration tests in tests/integration/test_centralized_config.go

### Implementation for User Story 3

- [x] T046 [P] [US3] Create Configuration Profile model in pkg/types/config.go
- [x] T047 [P] [US3] Implement etcd configuration provider in internal/config/etcd_provider.go
- [x] T048 [P] [US3] Implement Consul configuration provider in internal/config/consul_provider.go
- [x] T049 [US3] Create configuration synchronization engine in internal/config/manager_impl.go
- [x] T050 [US3] Implement API key authentication for configuration access in internal/config/auth.go
- [x] T051 [US3] Add configuration change propagation with 60-second SLA in internal/config/propagation.go
- [x] T052 [US3] Create configuration conflict resolution and error reporting in internal/config/conflicts.go

**Checkpoint**: ✅ All user stories should now be independently functional and can work together
- User Story 1: 10/10 complete (100%) - Single probe ready ✅
- User Story 2: 7/7 complete (100%) - Cross-platform ready ✅
- User Story 3: 7/7 complete (100%) - Centralized config ready ✅
- User Story 4: 30/30 complete (100%) - Web interface ready ✅

---

## Phase 5.5: User Story 4 - Admin Web Interface and Real-time Monitoring (Priority: P3)

**Goal**: Network operations team can use a web-based interface to create and manage probe configurations, monitor probe health in real-time, and view live measurement data through an intuitive dashboard

**Independent Test**: Deploy admin web interface, create configurations through web UI, deploy probes that register with web interface, verify real-time monitoring and live measurement streaming

### Tests for User Story 4 (OPTIONAL - only if tests requested) ⚠️

- [x] T062 [P] [US4] Web interface API contract tests in tests/contract/test_web_api.go
- [x] T063 [P] [US4] WebSocket real-time communication tests in tests/contract/test_websocket.go
- [x] T064 [P] [US4]React component integration tests in tests/web/components/
- [x] T065 [US4] End-to-end web interface tests in tests/web/e2e/test_admin_interface.go
- [x] T066 [US4] Cross-platform web interface deployment tests in tests/integration/test_web_deployment.go

### Implementation for User Story 4

**Backend Services:**
- [x] T067 [P] [US4] Create admin web interface backend in cmd/admin-web/main.go
- [x] T068 [P] [US4] Implement REST API endpoints for configuration management in internal/web/api/
- [x] T069 [P] [US4] Implement WebSocket service for real-time communication in internal/web/websocket/
- [x] T070 [P] [US4] Create probe registry and health monitoring service in internal/monitoring/
- [x] T071 [P] [US4] Implement authentication and authorization middleware in internal/web/auth/
- [x] T072 [P] [US4] Create configuration CRUD operations integrated with etcd provider in internal/config/web.go
- [x] T073 [P] [US4] Implement real-time measurement streaming from probes in internal/monitoring/stream.go
- [x] T074 [P] [US4] Add probe registration and heartbeat tracking in internal/monitoring/probe_registry.go

**Frontend Interface:**
- [x] T075 [P] [US4] Setup React + TypeScript project structure in web/ directory
- [x] T076 [P] [US4] Implement main dashboard layout with responsive design in web/src/pages/
- [x] T077 [P] [US4] Create configuration management interface in web/src/components/ConfigManager/
- [x] T078 [P] [US4] Implement real-time probe monitoring dashboard in web/src/components/ProbeMonitor/
- [x] T079 [P] [US4] Build WebSocket client for live updates in web/src/services/websocket.ts
- [x] T080 [P] [US4] Create interactive charts for measurement visualization in web/src/components/Charts/
- [x] T081 [US4] Implement configuration editor with JSON validation in web/src/components/ConfigEditor/
- [x] T082 [US4] Add real-time alert system and notifications in web/src/components/Alerts/

**Integration and Deployment:**
- [x] T083 [P] [US4] Configure multi-platform build for admin web service in Docker and native binaries
- [x] T084 [P] [US4]Setup CI/CD pipeline for both Go backend and React frontend
- [x] T085 [P] [US4] Implement integration between web interface and existing CLI tools
- [x] T086 [P] [US4] Add probe registration commands to CLI tools for web integration
- [x] T087 [P] [US4] Create Kubernetes deployment manifests for admin web interface
- [x] T088 [P] [US4] Implement health monitoring and metrics export for web interface

**Real-time Features:**
- [x] T089 [P] [US4] Implement live measurement streaming with WebSocket in backend
- [x] T090 [P] [US4] Create real-time probe status updates in web dashboard
- [x] T091 [P] [US4] Add instant configuration deployment to registered probes
- [x] T092 [P] [US4] Implement push notifications for critical alerts in web interface

**Checkpoint**: ✅ User Story 4 is now fully functional with complete API implementation! - All 30 tasks completed (100%)
- Backend API: ✅ Full REST API with authentication, configuration, probe, measurement, and monitoring endpoints
- Frontend Interface: ✅ React + TypeScript with real-time dashboard, configuration management, and monitoring
- WebSocket Integration: ✅ Real-time probe communication, measurement streaming, and health updates
- Multi-platform Support: ✅ Docker containers for linux/amd64, linux/arm64
- Compilation Test: ✅ Admin web backend compiles and runs successfully

---
## Phase 5.1: User Story 3 - Centralized Configuration Enhancements (Review + Hardening)

**Goal**: Align implementation of centralized configuration management and mesh-wide control with the designed architecture by replacing demo layers, enforcing versioned/authoritative behavior, and improving observability and security.

### Implementation Tasks

- [x] T093 [US3] Replace in-memory configuration storage in `internal/web/api/routes.go` with `config.Manager` integration for all `/config` endpoints, ensuring reads/writes go through the centralized providers (file/etcd/Consul) instead of local maps.
- [x] T094 [US3] Implement `GET /config/status` endpoint in `internal/web/api/routes.go` exposing `ManagerStatus` from `config.Manager` (provider health, update counts, last seen, health score) for operational visibility.
- [x] T095 [US3] Refactor `/probes` handlers in `internal/web/api/routes.go` to use `internal/monitoring/ProbeRegistry` instead of the local `probes` map, ensuring a single authoritative registry for probe identity, status, and metadata.
- [x] T096 [US3] Extend `internal/monitoring/probe_registry.go` to track applied configuration metadata per probe (e.g. `ConfigVersion`, `ConfigSource`, `ConfigAppliedAt`) and expose it via existing listing/get APIs.
- [x] T097 [US3] Add `/probes/:id/config-applied` endpoint in `internal/web/api/routes.go` that allows probes to report the configuration version/source they have successfully applied, updating `ProbeRegistry` accordingly for rollout tracking.
- [x] T098 [US3] Enhance `shouldAcceptConfiguration` in `internal/config/manager_impl.go` to use version-aware and provider-priority-aware rules (e.g. reject stale versions, prefer higher-priority providers) while remaining backward compatible.
- [x] T099 [US3] Add structured logging and metrics around configuration lifecycle in `internal/config/manager_impl.go` (initialization, provider failures, accepted/rejected updates, reloads) and expose propagation/health metrics via existing telemetry.
- [x] T100 [US3] Protect configuration management and probe control endpoints (`/config`, `/config/status`, `/config/propagate`, critical `/probes` operations) with existing auth middleware and role-based checks to prevent unauthorized central changes.
- [ ] T101 [US3] Update or add tests in `tests/contract/` and `tests/integration/` to validate:
  - API now uses `config.Manager` and `ProbeRegistry`
  - configuration acceptance rules (version/priority)
  - probe-reported config version tracking
  - `/config/status` and security constraints on central operations.

---

## Phase 5.2: Admin Web Rework - Real Backend Integration for Phase 4/5.1

**Goal**: Turn the existing admin web (React SPA + Go backend) from a demo UI into a production-aligned, API-driven interface using the real contracts from `internal/web/api/routes.go` and the updated implementation plan in `plan.md`.

**Dependencies**:
- Phase 3 (US1), Phase 4 (US2), Phase 5 (US3) foundational work completed
- Phase 5.1 config/probe integration (T093–T100) in place
- This phase refactors/extends User Story 4 (Admin Web Interface) to be fully functional

### 5.2.1 Backend Contract Hardening (Admin Web API)

- [ ] T200 [US4] Define `contracts/admin-web.openapi.yaml` reflecting actual handlers in `internal/web/api/routes.go`:
  - `/auth/*`, `/config*`, `/probes*`, `/measurements*`, `/monitoring*`
- [ ] T201 [US4] Sync `docs/API_ADMIN_WEB.md` with `admin-web.openapi.yaml` and `routes.go` to remove drift.
- [ ] T202 [US4] Normalize JSON schemas to be UI-friendly:
  - Document when responses are wrapped (`{ probes: [...] }`, `{ alerts: [...] }`) vs plain objects.
  - Clarify snake_case vs camelCase expectations.
- [ ] T203 [US4] Add structured logging for admin-web endpoints (auth, config, probes, monitoring) with correlation-friendly fields.
- [ ] T204 [US4] Extend contract tests in `tests/contract/config_and_probe_admin_test.go`:
  - Validate `/auth/login`, `/auth/me`, `/auth/refresh` response shapes.
  - Validate `/config/status`, `/config/propagate`, `/probes/:id/config-applied` behavior.
  - Ensure protected routes enforce auth/roles when middleware enabled.

### 5.2.2 Frontend Types and API Alignment

- [ ] T210 [US4] Update `web/src/types/index.ts` to match backend DTOs:
  - Align `Probe` with `monitoring.Probe` (id, name, platform, arch, ipAddress, status, lastSeen, health, config metadata).
  - Align `Alert` with `Alert` in `routes.go` (id, title, description, severity, source, status, created_at, resolved_at).
  - Align `DashboardStats` with `DashboardStats` in `routes.go`.
- [ ] T211 [US4] Refine auth-related types:
  - Represent `/auth/login` response as `{ token: string; user: User; expiresAt: number }`.
  - Remove or adjust `AuthToken`/`LoginResponse` definitions that don’t match backend.
- [ ] T212 [US4] Update `web/src/services/api.ts` to:
  - Use correct base paths (`/api/v1/...`) and exact route paths.
  - Map wrapped responses (e.g. `{ probes: [...] }`) into typed returns.
  - Normalize snake_case fields into camelCase in the client where needed.
  - Centralize error transformation for consistent UI messaging.

### 5.2.3 Authentication Flow & Route Guards

- [ ] T220 [US4] Fix `apiService.login` in `web/src/services/api.ts`:
  - Type response from `/auth/login` correctly; call `setAuthToken(token)` and persist user.
- [ ] T221 [US4] Update `useAuth` in `web/src/hooks/useAuth.ts`:
  - On mount, if `auth_token` exists, call `/auth/me` to validate.
  - On 401 or network errors, clear auth state and storage.
- [ ] T222 [US4] Adjust `App.tsx` routing:
  - Add explicit `/login` route.
  - Wrap protected routes (Dashboard, Targets, Config, etc.) in an auth guard.
  - Show loading indicator until initial auth check completes (no unauthenticated flicker).

### 5.2.4 Dashboard: Real Data & Probes Integration

- [ ] T230 [US4] Wire `Dashboard.tsx` to real endpoints:
  - Fetch `/monitoring/dashboard` for top-level stats (StatCards).
  - Fetch `/monitoring/health/summary` for system status.
- [ ] T231 [US4] Integrate `ProbesTable` with `/probes` and `/probes/:id/health`:
  - Display actual probe list from `ProbeRegistry`.
  - Show status (online/offline/degraded) and lastSeen.
- [ ] T232 [US4] Back `LatencyChart` and `PacketLossChart` with `/measurements/statistics`:
  - Use backend statistics where available; fall back gracefully if demo-only.
- [ ] T233 [US4] Implement consistent loading/error states for all dashboard sections.

### 5.2.5 Configuration Management UI

- [ ] T240 [US3][US4] Implement configuration list & detail in `web/src/pages/NetworkConfiguration.tsx`:
  - GET `/config` to show active configuration.
  - GET `/config/status` to surface provider health and last reload.
- [ ] T241 [US3][US4] Implement create/update operations:
  - POST `/config`, PUT `/config/:id` with validation (JSON schema client-side if available).
- [ ] T242 [US3][US4] Implement delete & propagate:
  - DELETE `/config/:id`
  - POST `/config/propagate` with user feedback.
- [ ] T243 [US3][US4] Add clear banners/toasts for:
  - Validation errors.
  - Propagation success/failure (referencing `/config/status`).

### 5.2.6 Probes & Rollout Visibility

- [ ] T250 [US2][US3][US4] Extend probes UI (Targets / Probe views) to:
  - Display configuration metadata per probe (version, source, appliedAt) using ProbeRegistry fields.
- [ ] T251 [US3][US4] Visualize `/probes/:id/config-applied`:
  - Show which probes have acknowledged a given configuration.
  - Highlight probes with stale or missing configuration.
- [ ] T252 [US4] Add manual “refresh” and auto-refresh interval (e.g., 10s) for probes listing using `/probes`.

### 5.2.7 Real-time & WebSocket Strategy (Incremental)

- [ ] T260 [US4] Mark current `web/src/services/websocket.ts` as experimental:
  - Add feature flag or configuration to disable if backend WS not present.
- [ ] T261 [US4] Implement safe fallback:
  - Use periodic polling of `/monitoring/dashboard`, `/monitoring/alerts`, `/probes` when WS disabled.
- [ ] T262 [US4] (Optional, if backend supports) Align WS client with `internal/web/websocket/service.go`:
  - Subscribe to real probe/measurement events.
  - Remove socket.io assumptions if incompatible.

### 5.2.8 Tests & Observability for the Rework

- [ ] T270 [US4][Test] Add frontend unit tests:
  - `apiService` mapping tests for auth, config, probes, monitoring.
  - `useAuth` tests for login, token persistence, `/auth/me` validation.
  - Dashboard components verifying they render data from mocked API.
- [ ] T271 [US4][Test] Add E2E scenario (e.g., Cypress/Playwright) under `tests/web/e2e/`:
  - Login via UI → load dashboard → see real-time stats from running backend.
- [ ] T272 [US4][Test] Extend `tests/contract/test_web_api.go`:
  - Ensure responses remain compatible with admin-web contracts.
- [ ] T273 [US4][Obs] Add backend logs/metrics:
  - Count admin-web operations (logins, config changes, probe views).
  - Export via existing telemetry for operational insight.

**Checkpoint**: ✅ Phase 5.2 complete when:
- Admin web frontend uses real backend APIs for auth, config, probes, dashboard.
- Types and responses are contract-tested (backend + frontend).
- Basic E2E flow (login → dashboard → config change) passes against a running stack.
- WebSocket usage is either correctly integrated or safely disabled without breaking core flows.

---

## Phase 5.3: Probe Daemon/Agent Mode + Backend Integration

**Goal**: Implement the probe as a long-running managed agent that:
- Registers with the backend,
- Periodically heartbeats and reports status/metrics,
- Retrieves and applies configuration from the control plane without restart,
- Remains aligned with Admin Web and ProbeRegistry contracts.

### Implementation Tasks

- [x] T280 [US1][US3] Implement probe daemon/agent entrypoint in `cmd/probe/main.go`:
  - Add `serve` or `agent` subcommand that:
    - Starts long-running loop.
    - Initializes telemetry, config manager, and ProbeRegistry client as needed.
- [x] T281 [US1][US3] Implement backend client for probe registration and heartbeat in `internal/monitoring` or `internal/cli`:
  - Provide functions:
    - `RegisterProbe(ctx, probeInfo)` to call backend (existing `/probes`/heartbeat pattern or future `/probes/register`).
    - `SendHeartbeat(ctx, probeID, status)` targeting `/probes/:id/heartbeat`.
  - Ensure TLS + auth headers align with config.
- [x] T282 [US3] Implement config pull + live-apply loop in daemon mode:
  - Periodically:
    - Fetch effective configuration from backend/config control plane (`/config` or targeted probe-config endpoint).
    - Apply updates to ICMP measurement engine without requiring process restart.
  - On successful apply:
    - Call `/probes/:id/config-applied` with version/source metadata.
- [x] T283 [US1][US3] Wire status/metrics reporting from daemon to backend:
  - Expose key probe metrics (health, active targets, error counts, measurement rates) via existing OTEL pipeline.
  - Ensure ProbeRegistry-visible fields (status, lastSeen, config metadata) are updated via backend calls.
- [x] T284 [US1][US3][US4] Ensure CLI one-shot mode remains intact and clearly separated:
  - Keep existing one-shot commands for ad-hoc checks.
  - Document that daemon/agent mode is the primary integration path for Admin Web and central control.

### Tests

- [x] T285 [Test][US1][US3] Extend `tests/contract/config_and_probe_admin_test.go`:
  - Simulate a probe daemon registering, heartbeating, pulling config, and reporting config-applied.
  - Assert ProbeRegistry and `/probes` endpoints reflect correct state.
- [x] T286 [Test][US1][US3] Add integration test in `tests/integration/test_probe_daemon_agent.go`:
  - Start backend + run probe daemon mode against it.
  - Verify:
    - Registration/heartbeat lifecycle.
    - Config retrieval + live-apply.
    - Status/metrics visibility in admin APIs.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T053 [P] Run comprehensive cross-platform test suite across Linux, macOS, Windows
- [ ] T054 [P] Performance optimization pass for microsecond-level measurement precision
- [ ] T055 [P] Security hardening and vulnerability scanning validation
- [ ] T056 Documentation updates with deployment guides and API documentation
- [ ] T057 Code cleanup and refactoring for maintainability
- [x] T058 [P] Final Docker multi-platform image builds and validation
- [ ] T059 Integration testing with OpenTelemetry Collector end-to-end
- [ ] T060 Run quickstart.md validation and create deployment examples

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Constitutional Compliance (Phase 1.5)**: Depends on Setup completion - CRITICAL quality gates
- **Foundational (Phase 2)**: Depends on Constitutional Compliance completion - BLOCKS all user stories
- **User Stories (Phases 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable
- **User Story 4 (P3)**: Can start after Foundational (Phase 2) - Strong integration with US3 (configuration management) and US1 (probe monitoring), web interface can be developed in parallel

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Models before services
- Services before CLI integration
- Core implementation before platform optimizations
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Constitutional Compliance tasks marked [P] can run in parallel (within Phase 1.5)
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (if tests requested):
Task: "Contract test for ICMP measurement interface in tests/contract/test_icmp_interface.go"
Task: "Integration test for single probe deployment in tests/integration/test_single_probe.go"
Task: "Performance benchmark test for microsecond precision in tests/performance/test_precision.go"

# Launch all models for User Story 1 together:
Task: "Create Probe Instance model in pkg/types/probe.go"
Task: "Create Measurement Data model in pkg/types/measurement.go"
Task: "Create Network Target model in pkg/types/target.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 1.5: Constitutional Compliance (CRITICAL - blocks all stories)
3. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
4. Complete Phase 3: User Story 1
5. **STOP and VALIDATE**: Test User Story 1 independently
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Constitutional Compliance + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Constitutional Compliance + Foundational together
2. Once Foundation is done:
   - Developer A: User Story 1
   - Developer B: User Story 2  
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence