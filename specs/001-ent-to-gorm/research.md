# Research: ent-to-gorm API parity

## GORM generic API + gorm CLI (codegen)
- **Decision**: Use gorm generic API (`gorm.G[T]`) with code generated models/repos from gorm CLI to replace ent repositories; keep DB access behind interfaces for mocking and DryRun/ToSQL contract checks.  
- **Rationale**: Generics provide typed queries; CLI generation reduces boilerplate and keeps schema in sync; interfaces enable deterministic tests and swap in DryRun or fake implementations. Guidance aligns with DeepWiki on sessions, transactions, DryRun, and interfaces（此结论来自 DeepWiki 文档）.  
- **Alternatives considered**: Keep ent and wrap gorm separately (more dual maintenance); raw SQL (loses type-safety and increases risk of drift).

## Database target and drivers
- **Decision**: Primary target PostgreSQL (compose uses pg15/timescaledb); retain MySQL compatibility only if existing deployments require it by keeping field parity and running smoke tests against a MySQL DSN before release.  
- **Rationale**: Current tests reference postgres; ent baseline includes mysql driver. Prioritizing postgres reduces variance while allowing optional MySQL verification.  
- **Alternatives considered**: Drop MySQL entirely (simpler but risks breaking mysql users); dual-first-class support (higher test matrix/time).

## API contract parity and testing
- **Decision**: Capture ent baseline responses as goldens and compare gorm output across critical endpoints; add regression tests for validation errors and transaction rollback behavior.  
- **Rationale**: Spec demands identical API behavior; golden diffing prevents silent changes; txn tests ensure no partial writes.  
- **Alternatives considered**: Manual QA only (insufficient coverage); relying on integration tests without goldens (higher miss risk).

## Performance and observability
- **Decision**: Validate p95 latency/error rates within ±5% of ent baseline on representative endpoints; enable tracing/logging of query SQL and duration in gorm; monitor slow queries and N+1 via preloading and query logging.  
- **Rationale**: Constitution requires budgets; migration could change query patterns; observability keeps regressions detectable.  
- **Alternatives considered**: Skip perf check (unacceptable); rely solely on DB metrics (less granular).

## Rollout and rollback
- **Decision**: Introduce config/flag to select gorm vs ent build during rollout; keep migrations reversible; run rehearsal with rollback path verified.  
- **Rationale**: Migration risk is high; fast rollback protects API consumers.  
- **Alternatives considered**: One-shot cutover (higher downtime risk).
