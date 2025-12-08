# Performance Baseline (ent-backed)

- DSN: `host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=Abcd1234 dbname=postgres sslmode=disable`
- Date: 2025-12-06
- Method: direct SQL via gorm against current ent-backed database

| Query | Sample Size | Duration (ms) |
|-------|-------------|---------------|
| `SELECT COUNT(*) FROM sys_users` | all | 37.9 |
| `SELECT * FROM sys_admin_operation_logs ORDER BY id DESC LIMIT 50` | 50 rows | 3.9 |

Notes: Use these as baselines; gorm implementation should stay within ±5% p95/error budgets relative to these measurements, adjusting for test environment variance.
