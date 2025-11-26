# Data Layer

The data package encapsulates every persistence concern for the admin service. Key directories:

- `gorm/` – auto-generated models and query builders produced by `cmd/gormgen` (gorm/gen). Generated code lives in `model/` and `query/` subfolders; do not edit by hand.
- `repositories` (under `gorm/`) – handwritten repository facades built on top of the generated query API to keep business-friendly interfaces.
- Entire data layer now uses gorm/gen; legacy Ent artefacts have been removed.
- Root Go files (e.g. `user_repo.go`, `department_repo.go`) expose repository implementations currently used by the service layer. During the migration each of these will be replaced with equivalents that delegate to the `gorm/` repositories.

Infrastructure helpers such as `NewGormClient`, Redis bootstrap, authentication token stores, and MinIO clients also live here. Any new external integration should be initialised in this package and injected into services through constructors.
