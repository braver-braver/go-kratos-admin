# API Contracts: ent-to-gorm API parity

- Objective: preserve existing API contracts while replacing ent with gorm. No request/response changes are allowed.
- Scope: all admin service endpoints under `api/admin/service/v1`. Use existing protobuf/OpenAPI definitions as the source of truth.
- Method:
  - Capture ent-backed responses as golden fixtures per critical endpoint (success + validation/error cases).
  - Run the same tests against gorm-backed build; compare status codes, headers, payload fields, ordering, and error semantics.
  - Track any schema-dependent differences (defaults, nullability) and block release until aligned.
- Deliverables:
  - Golden contract tests in backend/app/admin/service/internal/data or appropriate test package.
  - Regression cases for validation errors and transaction rollbacks.
  - Documentation in PR/notes citing any DeepWiki findings used for gorm behavior tuning（此结论来自 DeepWiki 文档）.

## Rollout & Rollback Playbook

- **Enable gorm**: set `USE_GORM=true` env or set `data.database.driver: gorm-postgres` in `backend/app/admin/service/configs/data.yaml`; restart service.
- **Validate**: run `go test ./app/admin/service/internal/data/gormcli/...` plus contract goldens and perf script `backend/script/perf/ent_vs_gorm.sh`; compare `testdata/gorm_results.json` vs `ent_goldens.json`.
- **Rollback**: switch driver/env back to ent (`postgres`), restart service, rerun goldens to confirm parity; keep ent client compiled in for immediate fallback.
- **Data safety**: migrations must remain reversible; use transaction rollback tests and schema parity checks before promoting gorm build.
