# API Overview / 接口说明

## Conventions / 约定
- **Client**: `utils/request.ts` (`requestClient`) adds `Authorization` Bearer token and `Accept-Language`; unwraps responses and retries refresh when enabled.
- **Data Models**: Types derive from generated protobuf DTOs under `src/generated/api/**`; services translate to REST endpoints.
- **Base URL**: From `useAppConfig(import.meta.env, import.meta.env.PROD)`; usually set via `VITE_GLOB_API_URL` in env files.

## Authentication / 认证
| Method | Path | Request | Response | Notes |
| --- | --- | --- | --- | --- |
| POST | `/login` | `LoginRequest` (username, AES-encrypted password, `grant_type=password`) | `LoginResponse` with `access_token`, `refresh_token` | Password encrypted using `VITE_AES_KEY`; token stored in `useAccessStore`. |
| POST | `/logout` | `Empty` | `Empty` | Clears stores and redirects to login. |
| POST | `/refresh_token` | `LoginRequest` (grant_type, refresh_token) | `LoginResponse` | Used by auth interceptor when refresh enabled. |

## User & Access / 用户与权限
| Method | Path | Request | Response | Notes |
| --- | --- | --- | --- | --- |
| GET | `/users` | `PagingRequest` (page,size,filters) | `ListUserResponse` | List users. |
| POST | `/users` | `CreateUserRequest` | `Empty` | Create user. |
| GET | `/users/{id}` | `GetUserRequest` | `User` | Fetch user detail. |
| PUT | `/users/{id}` | `UpdateUserRequest` (data.id stripped before send) | `Empty` | Update user. |
| DELETE | `/users/{id}` | `DeleteUserRequest` | `Empty` | Remove user. |
| POST | `/users/change-password` | `ChangePasswordRequest` | `Empty` | User self password change. |
| POST | `/users/{user_id}/password` | `EditUserPasswordRequest` | `Empty` | Admin password reset. |
| GET | `/users_exists` | `UserExistsRequest` | `UserExistsResponse` | Username/field existence check. |

## System Modules / 系统模块
Services follow similar CRUD REST patterns via `requestClient`:
- **Roles / 角色** (`role.service.ts`): list/create/update/delete roles; manage role permissions.
- **Menus / 菜单** (`menu.service.ts`): menu tree retrieval and CRUD.
- **Departments/Positions/Organizations / 部门、岗位、组织**: hierarchy & membership management.
- **Tenants / 租户** (`tenant.service.ts`): tenant lifecycle operations.
- **Dictionaries / 字典** (`dict.service.ts`): category/item CRUD for typed enums.
- **API Resources / 接口资源** (`api_resource.service.ts`): manage backend API definitions for access control.
- **Admin Logs / 管理日志** (`admin_operation_log.service.ts`, `admin_login_log.service.ts`): paginated log queries and cleanup.
- **Internal Messages / 站内信** (`internal_message*.service.ts`): categories, message delivery, read/unread status.
- **Tasks / 任务** (`task.service.ts`): scheduled/async task listing and operations.
- **Files / 文件** (`file.service.ts`): upload/download helpers using shared request client.
- **Router / 路由** (`router.service.ts`): dynamic menu/route payloads synced with backend access rules.

## Error Handling / 错误处理
- Responses with HTTP 2xx/3xx return `response.data`; non-success with `code` fields throw structured error objects.
- Auth interceptor attempts refresh/reauthenticate; on failure triggers logout or login-expired modal depending on preferences.
- Generic errors show via `ant-design-vue` `message.error`; customize in `utils/request.ts` if backend codes change.

## Internationalization / 国际化
- Requests carry `Accept-Language` from preferences; backend should honor locale for messages and data.
