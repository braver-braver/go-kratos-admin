# Service Layer

The service package hosts gRPC/HTTP handlers generated from protobuf definitions. Each file groups use-cases around a bounded domain (users, roles, departments, etc.). Responsibilities:

- Map transport-level DTOs to repository calls in `internal/data` and orchestrate cross-aggregate logic.
- Enforce light-weight validation and permission checks; heavy validation should live either in the API contract (protovalidate) or dedicated data-layer helpers.
- Maintain name-set filling utilities (`utils.go`) that decorate responses with human readable labels from related repositories.

When adding a new API:
1. Define the protobuf service in `api/gen` and regenerate Go stubs.
2. Create a `<domain>_service.go` here, injecting the required repositories through the constructor.
3. Register the service in `internal/server/init.go` so Kratos wires the transport endpoints.
