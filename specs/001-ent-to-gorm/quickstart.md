# Quickstart: ent-to-gorm API parity

1) **Setup**
- Ensure Go 1.24.6, docker/compose available.
- Start dependencies: from `backend` run `make init` then `make compose-up` (uses postgres/timescaledb). For this migration we use the existing PG/Redis instances already defined in `backend/app/admin/service/configs/data.yaml` (driver `gorm-postgres`) and `server.yaml`.
- Export DB DSN envs for tests: `export DSN="host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=Abcd1234 dbname=postgres sslmode=disable"`. Add `MYSQL_DSN` only if running optional MySQL smoke tests.

2) **Generate gorm code**
- Install gorm CLI (per project docs) and run codegen for models/repos in `backend/app/admin/service/internal/data/gormcli` targeting the current schema.
- Use gorm generic API (`gorm.G[T]`) in repositories; keep interfaces for biz layer injection and tests.

3) **Run checks**
- Lint: `golangci-lint run ./...` (backend).
- Contract/perf baselines: run ent-backed baselines (captured in `app/admin/service/internal/data/gormcli/testdata/ent_goldens.json` and `specs/001-ent-to-gorm/contracts/perf_baseline.md`). Generate gorm snapshot via `go run ./app/admin/service/internal/data/gormcli/cmd/results` (writes `testdata/gorm_results.json`) and run `go test ./app/admin/service/internal/data/gormcli/...` for parity, transaction rollback, and schema existence checks.
- DB integration: run migration checks and transaction/association tests; schema parity test ensures all generated tables exist.
- Observability: verify traces/logs/metrics emit DB latency and slow-query info.

4) **Rollout rehearsal**
- Provide flag/config to switch ent vs gorm build: set `USE_GORM=true` env or set driver to `gorm-postgres` in `configs/data.yaml` to enable gorm; driver names auto-map back to ent for fallback.
- Rehearse rollback: switch driver/env back to ent (`postgres`), restart service, and confirm schema/data unchanged. Keep ent client available for parity testing during rollout.
