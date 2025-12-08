---
description: "Task list for ent-to-gorm API parity"
---

# Tasks: ent-to-gorm API parity

**Input**: Design documents from `/specs/001-ent-to-gorm/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: All new logic requires automated tests. Include unit/integration/contract/regression tasks per user story; only omit tests when there is truly no code change and document that exception.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions
- Include tasks for performance checks against stated budgets and for UX states/locale updates when UI changes affect users

## Path Conventions

- Backend service: `backend/app/admin/service/internal/…`
- Contracts: `backend/api/admin/service/v1/`
- Migrations: `backend/sql/`
- Specs: `specs/001-ent-to-gorm/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and baseline capture

- [X] T001 Verify Go toolchain and lint tooling availability in backend/go.mod and install gorm CLI per README in backend/README.md
- [X] T002 Start postgres services via backend/docker-compose.yaml and export DSN envs for tests
- [X] T003 Capture ent-backed API golden responses into backend/app/admin/service/internal/data/gormcli/testdata/ent_goldens.json
- [X] T004 Establish baseline perf metrics for representative endpoints and record in specs/001-ent-to-gorm/contracts/perf_baseline.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [X] T005 Generate gorm generic models/repos using gorm CLI into backend/app/admin/service/internal/data/gormcli/
- [X] T006 Define repository interfaces and ent/gorm switchable provider in backend/app/admin/service/internal/data/data.go
- [X] T007 Add config/flag for ent vs gorm selection in backend/app/admin/service/internal/conf/conf.proto and wire in backend/app/admin/service/internal/server/
- [X] T008 Prepare deterministic DB test harness (postgres primary, optional MySQL) in backend/app/admin/service/internal/data/gormcli/testdata/dsn.env
- [X] T009 Validate migrations/schema parity for gorm vs ent in backend/sql/ with reversible plan documented in specs/001-ent-to-gorm/research.md

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Preserve API contracts post-ORM swap (Priority: P1) 🎯 MVP

**Goal**: API consumers experience identical behavior after swapping to gorm.

**Independent Test**: Golden contract suite comparing ent vs gorm responses passes with identical status codes, headers, payload fields, ordering, and validation errors.

### Tests for User Story 1 (MANDATORY for new logic) ⚠️

- [X] T010 [P] [US1] Build golden contract tests comparing ent and gorm responses in backend/app/admin/service/internal/data/gormcli/contract_tests.go
- [X] T011 [P] [US1] Add regression tests for validation error parity in backend/app/admin/service/internal/data/gormcli/validation_contract_test.go

### Implementation for User Story 1

- [X] T012 [P] [US1] Implement gorm generic repositories using generated code in backend/app/admin/service/internal/data/gormcli/*.go
- [X] T013 [US1] Wire service layer to use gorm repositories behind interfaces with feature flag fallback in backend/app/admin/service/internal/data/data.go
- [X] T014 [US1] Run golden contract suite against gorm build and store results in backend/app/admin/service/internal/data/gormcli/testdata/gorm_results.json

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Preserve data integrity during migration (Priority: P1)

**Goal**: Existing records remain readable/writable after migrating to gorm.

**Independent Test**: Schema parity checks pass; transactional writes roll back cleanly on failure; ent and gorm data snapshots match.

### Tests for User Story 2 (MANDATORY for new logic) ⚠️

- [X] T015 [P] [US2] Add transaction/rollback integration tests in backend/app/admin/service/internal/data/gormcli/transaction_test.go
- [X] T016 [P] [US2] Add data snapshot comparison tests ent vs gorm in backend/app/admin/service/internal/data/gormcli/data_parity_test.go

### Implementation for User Story 2

- [X] T017 [US2] Validate and adjust gorm-generated schemas/indexes for parity in backend/sql/ with rollback scripts noted in specs/001-ent-to-gorm/contracts/README.md
- [X] T018 [US2] Add optional MySQL smoke path if required in backend/app/admin/service/internal/data/gormcli/mysql_smoke_test.go using ent baseline fixtures
- [X] T019 [US2] Document migration/rollback procedure in specs/001-ent-to-gorm/quickstart.md

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Maintain operational performance and observability (Priority: P2)

**Goal**: Runtime characteristics remain stable and regressions are detectable.

**Independent Test**: Perf checks show p95/error rates within ±5% of ent baseline; tracing/logging/metrics attribute slow queries to gorm operations.

### Tests for User Story 3 (MANDATORY for new logic) ⚠️

- [X] T020 [P] [US3] Add perf/load test script comparing ent vs gorm in backend/script/perf/ent_vs_gorm.sh with results logged to specs/001-ent-to-gorm/contracts/perf_results.md
- [X] T021 [P] [US3] Add observability assertions for DB spans/logs/metrics in backend/app/admin/service/internal/data/gormcli/observability_test.go

### Implementation for User Story 3

- [X] T022 [US3] Implement slow-query logging and metrics hooks for gorm in backend/app/admin/service/internal/data/gormcli/hooks.go
- [X] T023 [US3] Ensure tracing propagation and DB span attributes remain intact in backend/app/admin/service/internal/server/server.go
- [X] T024 [US3] Tune preloading/queries to avoid N+1 or unbounded scans in backend/app/admin/service/internal/data/gormcli/query_optimizations.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T025 Update specs/001-ent-to-gorm/quickstart.md with final commands, flags, and DSN examples
- [X] T026 [P] Run full lint/test suite for backend via backend/Makefile targets and capture results in specs/001-ent-to-gorm/contracts/test_report.md
- [X] T027 Document rollout and rollback playbook in specs/001-ent-to-gorm/contracts/README.md
- [X] T028 Cleanup unused ent artifacts after cutover checklist in backend/app/admin/service/internal/data/README.md

---

## Dependencies & Execution Order

- User stories prioritized: US1 (contract parity) → US2 (data integrity) → US3 (performance/observability).
- Foundation (Phase 2) blocks all stories. US1 and US2 can proceed in parallel after Phase 2, but US1 completion is MVP for delivery.

## Parallel Example: User Story 1

```bash
# Parallelizable within US1 once gorm models exist:
Task: "Build golden contract tests comparing ent and gorm responses in backend/app/admin/service/internal/data/gormcli/contract_tests.go"
Task: "Implement gorm generic repositories using generated code in backend/app/admin/service/internal/data/gormcli/*.go"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently (goldens/regression)
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
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
