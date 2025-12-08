# Data

Cutover checklist (ent -> gorm):
- Keep ent client available until gorm parity is validated; `Data.UseGorm()` gates behavior.
- Config switch: `USE_GORM=true` or driver `gorm-postgres` to enable gorm; revert to `postgres` to fall back.
- Before removing ent code, ensure all repos have gorm equivalents, goldens pass, and schema parity checks succeed.
- Post-cutover cleanup: delete ent-generated folders only after a stable period and when rollback is no longer required.
