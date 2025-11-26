# Admin Service

This module hosts the Kratos-based admin back-end service. It is organised in the following layers:

- `cmd/server` – bootstraps the HTTP/gRPC workers via Kratos Wire wiring.
- `configs` – default configuration templates consumed by the bootstrap pipeline.
- `internal/server` – transport initialisation (REST, SSE, background workers) and middleware wiring.
- `internal/service` – application use-cases exposed over generated protobuf endpoints.
- `internal/data` – data-access layer (redis, ORM repositories, authorisation helpers).

When adding new business capabilities place protoc-generated stubs under `api/` first, implement the behaviour inside `internal/service`, and keep cross-cutting helpers inside `internal/data` or `pkg/` depending on their reuse scope.
