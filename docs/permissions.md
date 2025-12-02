# Permission Model Overview (API / Menu / Data Scope)

## Goals
- Make authorization expectations explicit for API endpoints, menu rendering, and data-scope filters.
- Provide a minimal, repeatable flow for adding new resources without surprising auth gaps.
- Ensure operators can reason about which identifiers and claims drive each layer.

## Concepts
- **Authentication**: JWT (`Authorization: Bearer <token>`) validated by authn middleware.
- **Authorization (API)**: Casbin-based rules mapping _subject_ (user/role) → _object_ (route or RPC) → _action_ (method).
- **Menu Permission**: UI navigation and button visibility driven by backend-provided menu tree filtered by the user’s roles.
- **Data Scope (Department/Organization)**: Row-level filter applied to queries based on the current user’s data range (e.g., own dept, dept + children, custom).

## API Permissions
- Enforcement: REST/gRPC pass through `authn` + `authz` middlewares (Casbin).
- Whitelist (no auth): `admin.service.v1.AuthenticationService/Login`, `admin.service.v1.TenantService/TenantExists`.
- Identity source: JWT claims (user id, tenant id, roles). Subject typically resolves to role(s) or user id in Casbin.
- Objects/actions: API resources registered in DB (see `backend/sql/default-data.sql`) with path + method (e.g., `/admin/v1/users` + `GET`).
- Adding a new API:
  1) Define route/RPC.
  2) Insert resource definition (path, method, module, action) into authz policy store.
  3) Bind resource to roles in Casbin policy.
  4) If the API must be public, add to REST whitelist.

## Menu Permissions
- The backend returns a menu tree already filtered by user roles/permissions.
- Frontend renders routes/buttons from that tree; hidden items are omitted, avoiding client-side hard-coding.
- Button/operation-level control is modeled as resources bound to roles; the backend omits entries the user lacks.
- Adding a menu entry:
  1) Create menu node in DB (title, path, component, parent).
  2) Link node to required resource/permission.
  3) Assign to roles; ensure API resources it links to are also bound.

## Data Permissions (Department/Organization Scope)
- Purpose: limit query results to an allowed organization range for the current user.
- Scope derivation: from user profile + role configuration (e.g., self-only, own dept, dept + children, custom dept set, or all).
- Enforcement: service/repo layer applies filters (e.g., `dept_id IN (...)`) based on derived scope before DB queries.
- Adding data-scope-aware queries:
  1) Determine the org field (e.g., `dept_id`, `tenant_id`).
  2) Fetch current user’s scope once per request.
  3) Apply scope filter to list/count/export queries; avoid bypass paths (raw SQL, caches).

## Sequence Diagrams

### API AuthZ (JWT + Casbin)
```plantuml
@startuml
actor Client
participant Gateway as REST/gRPC
participant Authn as Authn MW
participant Authz as Casbin
participant Handler

Client -> Gateway: HTTP/gRPC + Authorization: Bearer
Gateway -> Authn: validate JWT
Authn -> Gateway: subject/claims (user, roles, tenant)
Gateway -> Authz: enforce(subject, object=route, action=method)
Authz --> Gateway: allow/deny
Gateway -> Handler: on allow
Handler --> Client: response
Gateway --> Client: 401/403 on fail
@enduml
```

### Menu Permission Flow
```plantuml
@startuml
actor User
participant Frontend
participant Backend as Menu API
participant Authz

User -> Frontend: open app
Frontend -> Backend: GET /menus (with JWT)
Backend -> Authz: filter menu items by user roles/permissions
Authz --> Backend: allowed menu tree
Backend --> Frontend: filtered menu tree
Frontend -> User: render only allowed routes/buttons
@enduml
```

### Data Scope Enforcement
```plantuml
@startuml
actor User
participant Service
participant Scope as DataScope Resolver
participant Repo as Repository/DB

User -> Service: List request
Service -> Scope: resolve user data scope (dept/tenant/custom)
Scope --> Service: allowed ids
Service -> Repo: query with scope filter (WHERE dept_id IN scope)
Repo --> Service: rows
Service --> User: filtered result
@enduml
```

## Operational Checklist
- JWT valid and not expired; correct roles in claims.
- Casbin policies present for every protected API; whitelist only where intended.
- Menu API tested with different roles to ensure proper filtering.
- Data-scope helpers applied to all list/count/export endpoints that expose org-bound data.
- Logs emit `trace_id`, `span_id`, and `X-Request-ID` for correlation; use `Trace-Id` for Jaeger lookup.
