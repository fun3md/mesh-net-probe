# AGENTS.md

This file provides guidance to agents when working with code in this repository.

- Run Go tests and tools from the repo root; run frontend commands only inside `web/`.
- Admin Web is served by the Go service in [`cmd/admin-web/main.go`](cmd/admin-web/main.go) with routes defined in [`internal/web/api/routes.go`](internal/web/api/routes.go:62); the React SPA under `web/` is a thin client over these routes and MUST not invent new endpoints.
- The authoritative API contract for admin-web is [`specs/001-mesh-probe-system/contracts/admin-web.openapi.yaml`](specs/001-mesh-probe-system/contracts/admin-web.openapi.yaml:1); align all frontend types and calls with it instead of guessing shapes.
- Probe state and configuration:
  - Probes are managed via `ProbeRegistry` through `/probes/admin` routes; do not use legacy or direct maps.
  - Configuration is owned by `config.Manager`; UI writes should go through `/config` + `/config/propagate` and treat `types.Configuration.data` as the extensibility surface.
  - Probe configuration rollout tracking uses `/probes/:id/config-applied`; frontend should read config metadata from probe objects, not synthesize it.
- For probe configuration UI and tasks:
  - Store build configurations and measurement-task definitions inside the single active configuration’s `data` (e.g. `data.buildConfigs`, `data.measurementTasks`) so that daemon/agent probes can consume them via the existing config pull.
  - When assigning build configurations to probes, update them via `/probes/admin/:id` using metadata fields that match the backend contract (snake_case to camelCase mapping happens in the client).
- HTTP auth:
  - All real admin operations are expected to use JWT via `auth.Middleware`; keep the frontend’s `apiService` token handling aligned with `/auth/login`, `/auth/me`, `/auth/refresh` responses (see `routes.go` and OpenAPI).
- Testing and commands (non-obvious locations):
  - Contract and integration tests for admin-web + probes live under `tests/contract/` and `tests/integration/`; use these as the source of truth for behavior, not only docs.
  - Frontend package.json is in `web/`; from repo root use:
    - `cd web && npm test` or `cd web && npm run build` instead of running npm at the root (no package.json there).
- When adding new behavior:
  - Prefer updating contracts (`specs/.../admin-web.openapi.yaml`) and Go handlers first, then align `web/src/types` and `web/src/services/api.ts`; never diverge types from actual handler responses.