# Feature Specification: ent-to-gorm API parity

**Feature Branch**: `001-ent-to-gorm`  
**Created**: 2025-12-05  
**Status**: Draft  
**Input**: User description: "这是一个重构项目，重点在于将backend 后端服务中依赖的 ent orm 替换为 gorm 生态. 同时保证对外提供的api 严格一致。"

## User Scenarios & Testing *(mandatory)*

Include UX acceptance (loading/error/empty states, accessibility basics, locale updates) and any performance expectations tied to the story.

### User Story 1 - Preserve API contracts post-ORM swap (Priority: P1)

API consumers experience identical behavior after the backend moves from ent to gorm.

**Why this priority**: Any change in request/response contracts breaks clients and integration partners.

**Independent Test**: Run golden contract tests comparing current ent-based responses to gorm-based responses; validate status codes, payload fields, ordering, and validation errors are unchanged.

**Acceptance Scenarios**:

1. **Given** an existing client calling any documented endpoint with valid data, **When** the service is backed by gorm, **Then** the status code, headers, and payload structure match the ent baseline.
2. **Given** an invalid request that previously returned a specific validation error, **When** processed through the gorm-backed service, **Then** the same error code/message is returned and no partial writes occur.

---

### User Story 2 - Preserve data integrity during migration (Priority: P1)

All existing records remain readable/writable after migrating models and queries to gorm.

**Why this priority**: Tenant and user data loss or corruption is unacceptable.

**Independent Test**: Snapshot representative tables before/after migration and compare row counts, key relationships, and business-critical fields; perform read/write round trips on migrated endpoints.

**Acceptance Scenarios**:

1. **Given** existing tenant/user/role/menu data, **When** the service runs on gorm, **Then** reads return the same values and associations as before the migration.
2. **Given** new records created post-migration, **When** they are read by legacy-compatible clients, **Then** no schema drift or constraint violations occur.

---

### User Story 3 - Maintain operational performance and observability (Priority: P2)

Runtime characteristics remain stable and regressions are detectable after the ORM change.

**Why this priority**: Performance or reliability regressions degrade the admin experience and increase ops load.

**Independent Test**: Run load checks on representative endpoints; compare p95 latency and error rates to pre-migration baselines; verify tracing/metrics still emit.

**Acceptance Scenarios**:

1. **Given** steady-state traffic levels, **When** the gorm-backed service handles requests, **Then** p95 latency and error rates stay within agreed budgets versus the ent baseline.
2. **Given** an elevated error or slow query, **When** operators inspect logs/traces/metrics, **Then** they can attribute the issue to the relevant gorm query or transaction without missing context.

### Edge Cases

- Long-running transactions or batch operations during migration and how they are rolled back without partial data writes.
- Concurrent writes during deployment and how locking/versioning prevents lost updates.
- Schema drift between environments (local/stage/prod) and how mismatches are detected before rollout.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All existing backend API endpoints MUST return identical response structures, status codes, error semantics, and ordering compared to the ent implementation.
- **FR-002**: Database schema and constraints exposed to clients (field names, types, defaults, indexes) MUST remain compatible; any migration MUST include rollback steps and validation before release.
- **FR-003**: Create/update flows MUST enforce the same validation rules as before and prevent partial writes on failure.
- **FR-004**: New records written via gorm MUST remain readable by existing clients without additional transformations.
- **FR-005**: Deployment MUST provide a safe switch/flag or rollout plan that allows reverting to the ent-backed build if critical regressions appear.

### Non-Functional Requirements *(mandatory)*

- **NFR-Quality**: Code MUST pass gofmt/golangci-lint; keep clear service interfaces; all DB calls honor request context for cancellation/timeouts; concurrency safety retained where goroutines touch shared state.
- **NFR-Testing**: Provide contract/regression tests comparing ent vs gorm responses for critical endpoints, integration tests for transactions and associations, and deterministic unit tests using mocks/fakes; document any remaining gaps with owners.
- **NFR-UX**: No UI changes are expected; if error surfaces differ, ensure the admin frontend still shows the same loading/error/empty states and locale strings.
- **NFR-Performance**: p95 latency and error rates for representative endpoints MUST stay within ±5% of ent baselines; detect and prevent N+1 query regressions and unbounded scans.
- **NFR-Observability**: Traces/logs/metrics MUST continue to emit correlation IDs, DB latency, error classifications, and slow-query indicators to attribute issues to specific gorm operations.
- **NFR-Dependencies**: The gorm adoption MUST follow Context7 DeepWiki guidance on sessions, transactions, migrations, and DryRun/mocking; record findings with “此结论来自 DeepWiki 文档” in notes/PRs; wrap DB access for testability and pin versions.

### Key Entities *(include if feature involves data)*

- **User/Tenant/Role/Permission/Menu**: Existing relationships and identifiers remain unchanged; access control data stays consistent across reads/writes.
- **Audit/Log Records**: Operational and security logs retain the same fields and retention expectations after the ORM swap.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of existing API contract/golden tests pass against the gorm-backed service compared to the ent baseline.
- **SC-002**: Data integrity checks show 0 unexpected schema diffs and 0 data-loss incidents in pre-production migration rehearsals.
- **SC-003**: p95 latency and error rates for agreed representative endpoints stay within ±5% of baseline across load tests.
- **SC-004**: Rollback rehearsal completes successfully within the planned maintenance window with no residual schema or data drift.
