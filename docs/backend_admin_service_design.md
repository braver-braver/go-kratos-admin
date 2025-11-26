# Admin Service Design (GORM)

## Overview
- **Purpose**: Admin backend built on Kratos with full GORM/Gen data layer and Redis/MinIO integrations.
- **Entry**: `app/admin/service/cmd/server` boots REST + SSE + Asynq via Wire.
- **Transports**: REST with swagger (optional), SSE for push, Asynq for background jobs.
- **Security**: JWT authn (kratos-authn), authz (casbin/OPA/noop) policies generated from roles + API resources.

## Architecture
- **Service layer** (`internal/service`): Protobuf HTTP servers per domain (auth, user, tenant, org/dept/position, role/menu/router/api-resource, dict, file/oss/ueditor, task, internal message, admin logs, login restriction, user profile/credential).
- **Data layer** (`internal/data`): GORM/Gen models + handwritten repositories, Redis token cache, MinIO client, authorizer policy builder, password crypto helper.
- **Server wiring** (`internal/server`): Middleware (logging, op/login log hooks, authn/authz), swagger registration, SSE server, Asynq server with task subscribers.

## Key HTTP Interfaces (from protobuf)
- Auth: `POST /admin/v1/login`, `POST /admin/v1/refresh_token`, `POST /admin/v1/logout`.
- Users: `GET /admin/v1/users`, `GET /admin/v1/users/{id}`, `POST /admin/v1/users`, `PUT /admin/v1/users/{data.id}`, `DELETE /admin/v1/users/{id}`, `POST /admin/v1/users/{user_id}/password`, `POST /admin/v1/users/change-password`, `GET /admin/v1/users_exists`.
- Tenants: CRUD under `/admin/v1/tenants`; create-with-admin handled in service logic.
- Org/Dept/Position: CRUD under `/admin/v1/organizations`, `/admin/v1/departments`, `/admin/v1/positions` with tree/list helpers.
- RBAC: Roles `/admin/v1/roles` (bind menus/apis); Menus `/admin/v1/menus`; Routers `/admin/v1/routers`; API resources `/admin/v1/api-resources`.
- Dict: Dict types `/admin/v1/dict/types`, entries `/admin/v1/dict/entries` (batch delete supported).
- Files: OSS `/admin/v1/oss/files`, generic files `/admin/v1/files`, UEditor uploads `/admin/v1/ueditor`.
- Tasks: `/admin/v1/tasks` CRUD + enable/disable/start/stop; async backup subscriber.
- Internal Message: `/admin/v1/internal-messages` CRUD/send; categories `/admin/v1/internal-message-categories`.
- Admin Ops: `/admin/v1/admin-login-logs`, `/admin/v1/admin-operation-logs`; Login restrictions `/admin/v1/admin-login-restrictions`.

## Data Layer Notes
- GORM/Gen code under `internal/data/gorm`; repositories wrap generated queries.
- `NewGormClient` sets default query client; Redis token cache for JWT/refresh tokens.
- Authorizer rebuilds policies from DB roles + API resources into casbin/OPA engines.
- Password crypto uses bcrypt helper.

## Sequence Sketches (PlantUML)
```plantuml
@startuml
title Login (password grant)
actor Client
participant AuthenticationService as Auth
participant UserCredentialRepo as CredRepo
participant UserRepo as UserRepo
participant RoleRepo as RoleRepo
participant UserTokenCacheRepo as TokenStore
Client -> Auth: POST /admin/v1/login (username, password)
Auth -> CredRepo: VerifyCredential(username,password,decrypt)
CredRepo --> Auth: ok / error
Auth -> UserRepo: GetUserByUserName
UserRepo --> Auth: User
Auth -> RoleRepo: GetRoleCodesByRoleIds(user.role_ids)
RoleRepo --> Auth: role codes
Auth -> TokenStore: GenerateToken(user, clientId)
TokenStore --> Auth: access/refresh token
Auth --> Client: tokens (bearer)
@enduml
```

```plantuml
@startuml
title Create User (admin)
actor Client
participant UserService as UserSvc
participant UserRepo as UserRepo
participant UserCredentialRepo as CredRepo
participant TenantRepo as TenantRepo
Client -> UserSvc: POST /admin/v1/users (payload)
UserSvc -> TenantRepo: Get(id) [optional]
TenantRepo --> UserSvc: Tenant
UserSvc -> UserRepo: Create(request)
UserSvc -> CredRepo: Create password credential (if provided)
CredRepo --> UserSvc: ok
UserSvc --> Client: 200/empty
@enduml
```

```plantuml
@startuml
title Send Internal Message
actor Client
participant InternalMessageService as MsgSvc
participant InternalMessageRepo as MsgRepo
participant InternalMessageRecipientRepo as RcptRepo
participant SseServer as SSE
Client -> MsgSvc: POST /admin/v1/internal-messages (content, recipients)
MsgSvc -> MsgRepo: Create message record
MsgRepo --> MsgSvc: message id
MsgSvc -> RcptRepo: BatchCreate recipients (status=unread)
RcptRepo --> MsgSvc: ok
MsgSvc -> SSE: push event per recipient stream (if connected)
MsgSvc --> Client: message id / ack
@enduml
```

## Testing Guidance
- Use real configs in `configs/` for integration when possible; Redis/MinIO/DB endpoints must be reachable.
- Run `make test` or `go test ./...` in `backend/app/admin/service` with env pointing to test DB/Redis.
- Swagger UI available when `server.rest.enable_swagger` is true to inspect routes.
