<!--
Sync Impact Report - Mesh-Net-Probe Constitution v1.0.1
=======================================================

VERSION CHANGE: 1.0.0 → 1.0.1 (Template consistency update)

PRINCIPLES MODIFIED:
- All existing principles remain unchanged (Code Quality, Testing, UX, Performance, Cross-Platform)

TEMPLATES UPDATED:
✅ plan-template.md: Added detailed Constitution Check section with principle-specific gates
✅ tasks-template.md: Added constitutional compliance framework (Phase 1.5) and principle-driven task categorization

TEMPLATE COMPLIANCE:
✅ plan-template.md: Constitution Check now references all 5 constitutional principles with specific requirements
✅ tasks-template.md: Added Phase 1.5 for constitutional compliance framework with quality gates
✅ tasks-template.md: Updated phase dependencies to include constitutional compliance checkpoint
✅ tasks-template.md: Added principle-aware task organization and compliance tracking

SECTIONS ADDED:
- Plan Template Constitution Check section with detailed gates for each principle
- Tasks Template Constitutional Compliance Framework (Phase 1.5)

SECTIONS REMOVED:
- None (all existing content preserved)

FOLLOW-UP TODOS:
✅ Update plan-template.md Constitution Check section - COMPLETED
✅ Review tasks-template.md for principle-driven task categorization - COMPLETED
✅ Verify command templates align with new CLI consistency requirements - N/A (no commands directory)

CONSTITUTIONAL RATIONALE:
This update ensures template consistency by adding explicit constitutional compliance checkpoints
to the planning and task management processes. Quality gates now exist at multiple stages to prevent
violations of established principles during development.
-->

# Mesh-Net-Probe Constitution

## Core Principles

### I. Code Quality Excellence
All Go code MUST follow strict quality standards: Comprehensive error handling with contextual messages; Clean architecture with separation of concerns; Comprehensive documentation for all public APIs and complex logic; Consistent coding style enforced via golangci-lint; Memory-safe practices preventing leaks and race conditions. Rationale: High-reliability network measurement requires bulletproof code to prevent data corruption and measurement inaccuracies.

### II. Comprehensive Testing Standards  
MANDATORY test coverage: 80% minimum for all packages; Unit tests for all business logic and utilities; Integration tests for network protocols and OS interactions; Cross-platform testing on Linux, macOS, Windows; Performance benchmarks for measurement accuracy and latency; End-to-end tests simulating real mesh network scenarios. Rationale: Network probes operate in diverse environments where failures cascade through measurement data integrity.

### III. User Experience Consistency
CLI interface MUST provide consistent command patterns and flag structures; Text-based output MUST support both human-readable and JSON formats; Configuration MUST be intuitive with sensible defaults and clear validation; Error messages MUST be actionable with suggested resolutions; Progress indicators and status reporting MUST be informative yet non-verbose. Rationale: Network operators require predictable tools that work identically across diverse operational contexts.

### IV. Performance Requirements
Measurement precision MUST achieve microsecond-level accuracy; Resource consumption MUST be optimized for long-running deployments; Scalable architecture MUST handle high-frequency probing without degradation; Zero-allocation paths REQUIRED for hot measurement code; Memory footprint MUST remain bounded regardless of measurement duration. Rationale: Real-time network analysis demands predictable performance characteristics to ensure measurement validity.

### V. Cross-Platform Compatibility
Identical measurement semantics MUST be maintained across Linux, macOS, Windows; Platform-specific optimizations MUST preserve functional equivalence; Architecture support REQUIRED for x64 and ARM64; Containerization MUST work consistently across Linux distributions; Platform detection and capability checking MUST be automatic and transparent. Rationale: Distributed mesh networks span heterogeneous environments requiring consistent measurement behavior.

## Platform & Deployment Standards

### Container & Orchestration Requirements
Docker multi-platform builds REQUIRED for linux/amd64 and linux/arm64; OpenTelemetry OTLP integration MUST be standardized across all metric export paths; Resource limits and requests MUST be configured for production deployments; Health checks and readiness probes MUST be implemented for orchestrators; Graceful shutdown procedures MUST preserve measurement data integrity.

### Measurement Protocol Standards
ICMP packet structure and timing MUST follow RFC standards; Measurement data formats MUST be compatible with common observability platforms; Timestamp generation MUST use monotonic clocks to prevent time drift; Network interface selection MUST prioritize measurement accuracy over convenience; Protocol error handling MUST distinguish between network and measurement failures.

## Development Workflow

### Code Review & Quality Gates
All contributions MUST pass automated linting, formatting, and static analysis; Code reviews MUST verify principle compliance and test coverage; Performance profiling REQUIRED for measurement-critical code paths; Cross-platform testing MUST be validated before merge; Documentation updates MANDATORY for any user-facing changes. Rationale: Quality gates prevent regression in measurement reliability and maintain operational consistency.

### Deployment & Release Process
Semantic versioning REQUIRED with automated changelog generation; Container images MUST be scanned for vulnerabilities; Rolling deployment strategy MUST be supported for zero-downtime updates; Rollback procedures MUST be tested and documented; Release candidates MUST pass full cross-platform test suite.

**Version**: 1.0.1 | **Ratified**: 2025-11-02 | **Last Amended**: 2025-11-03