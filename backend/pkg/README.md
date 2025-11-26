# Shared Packages

Utilities shared across multiple services live under `pkg/`. Each sub-package focuses on a single cross-cutting concern:

- `entgo/` – legacy Ent helpers (codecs, mixins) kept until the ORM migration completes.
- `jwt/` – JWT helpers used by authentication middleware.
- `metadata/` – metadata helpers for propagating request-scoped values.
- `middleware/` – Kratos middleware implementations (auth, logging, tracing).
- `oss/` – MinIO/S3 storage adapters and uploader utilities.
- `service/` – shared service helpers (e.g. dynamic service naming).
- `task/` – background task definitions/integration with Asynq.
- `utils/` – generic helpers that do not fit a dedicated domain yet.

When introducing a reusable helper prefer placing it in an existing specialised folder; only extend `utils/` as a last resort to avoid accumulating unrelated glue code.
