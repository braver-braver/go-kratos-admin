# Implementation Plan: ent-to-gorm API parity

**Branch**: `001-ent-to-gorm` | **Date**: 2025-12-05 | **Spec**: specs/001-ent-to-gorm/spec.md
**Input**: Feature specification from `/specs/001-ent-to-gorm/spec.md`

**Note**: Constitution Check gates below must be satisfied; record any Context7 DeepWiki findings when external packages are involved.

## Summary

Replace backend repo layer from ent ORM to gorm generic API with gorm CLI code generation while keeping all external API contracts, validation behavior, and performance/observability characteristics identical to the ent baseline. Add contract/perf regression tests, rollback switch, and schema safety to prevent drift.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.24.6  
**Primary Dependencies**: go-kratos, gorm generic API, gorm CLI codegen, wire, ent (baseline for parity tests), golangci-lint, testify  
**Storage**: PostgreSQL (compose uses timescaledb pg15); MySQL driver exists in ent baseline—confirm whether dual support must be preserved  
**Testing**: go test with contract/regression + integration DB tests; golangci-lint; golden API responses; load/perf checks against baseline  
**Target Platform**: Linux containers (docker/compose; deployable to k8s)  
**Project Type**: Backend service (admin APIs)  
**Performance Goals**: API p95 latency/error rates within ±5% of ent baseline; avoid N+1 and unbounded scans  
**Constraints**: No API contract changes; schema drift prohibited; transactions/locking must prevent partial writes; rollout must include revert path to ent build  
**Scale/Scope**: Multi-tenant admin traffic; throughput targets follow existing deployment baselines—confirm expected concurrency for perf checks

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Code quality: gofmt/golangci-lint enforced; repository interfaces remain explicit; DB calls honor context cancellation/timeouts; goroutine safety kept where shared state exists.
- Testing: add golden API contract tests (ent vs gorm), integration tests for transactions/associations, regression tests for known defects; DI with gorm interfaces/mocks; deterministic CI.
- UX: No UI changes expected; if error surfaces differ, ensure admin frontend keeps loading/error/empty states and locale strings aligned.
- Performance/reliability: p95 latency/error rates within ±5% of ent baseline; tracing/metrics preserved; perf checks on representative endpoints; avoid N+1 and add slow-query visibility.
- External dependencies: gorm generic API + gorm CLI; follow Context7 DeepWiki guidance (“此结论来自 DeepWiki 文档”); wrap DB access behind interfaces; pin versions and document upgrade/rollback.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
backend/
├── api/admin/service/v1/         # protobuf/OpenAPI contracts
├── app/admin/service/internal/
│   ├── biz/                      # domain services
│   ├── conf/                     # configuration
│   ├── data/                     # repositories (ent -> gorm generic/CLI)
│   ├── server/                   # transport wiring
│   └── service/                  # handlers
├── pkg/                          # shared helpers
├── sql/                          # migrations/schema
└── script/                       # tooling (prepare, etc.)

frontend/                         # admin UI (unchanged by this migration)
specs/001-ent-to-gorm/            # plan/spec/research/contracts/data-model/tasks
```

**Structure Decision**: Use existing backend service layout; replace `internal/data` ent repos with gorm generic/CLI output while keeping biz/service unchanged. Contracts live in `api/admin/service/v1`, tests under backend/app/admin/service/internal/data/... (add contract/perf/golden tests as needed).

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
