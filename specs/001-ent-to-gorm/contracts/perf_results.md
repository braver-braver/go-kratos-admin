# Perf Results: ent vs gorm

- DSN: host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=Abcd1234 dbname=postgres sslmode=disable
- Endpoint: /admin/v1/users?page=1&pageSize=20
- Concurrency: 10
- Requests: 100
- Ent base URL: http://localhost:8000
- Gorm base URL: http://localhost:8001

> Run `backend/script/perf/ent_vs_gorm.sh` with the ent (baseline) service on port 8000 and the gorm-enabled service on port 8001. Attach outputs next to this file for recordkeeping.
