# API Permission Audit (Frontend → Backend → DB)

Scope: mapping of frontend-admin API calls to backend routes and `sys_api_resources`/`sys_roles` assignments. Status is based on current DB contents queried via db MCP.

## Summary
- Tenant APIs: CRUD routes are registered, but `tenants_exists` and `tenants_with_admin` are missing in `sys_api_resources` (so non-whitelisted calls will 403). Super role currently owns all defined APIs (ids 1–102); other roles (`admin`, `user`, `guest`) have empty `apis`.
- Broad gap: only role `super` is functional; to grant real access, assign `apis` to other roles or introduce role-based policies.

## Tenant Service (frontend/apps/admin/src/services/tenant.service.ts)
| Frontend call | Backend path | Method | sys_api_resources | Role coverage |
| --- | --- | --- | --- | --- |
| `List` | `/admin/v1/tenants` | GET | ✔ id=36 | super only |
| `Create` | `/admin/v1/tenants` | POST | ✔ id=37 | super only |
| `CreateTenantWithAdminUser` | `/admin/v1/tenants_with_admin` | POST | ✔ id=104 | super only |
| `Delete` | `/admin/v1/tenants/{id}` | DELETE | ✔ id=29 | super only |
| `Get` | `/admin/v1/tenants/{id}` | GET | ✔ id=30 | super only |
| `Update` | `/admin/v1/tenants/{data.id}` | PUT | ✔ id=22 | super only |
| `TenantExists` | `/admin/v1/tenants_exists` | GET | ✔ id=103 | super only |

Actions to fix tenant endpoints:
1) DONE: inserted `sys_api_resources` ids 103 (tenants_exists, GET) and 104 (tenants_with_admin, POST).
2) TODO: add these ids (and other APIs as needed) to roles beyond `super` (`admin`, `user`, `guest`) in `sys_roles.apis` per desired policy.

## Role table snapshot
- `sys_roles`: `super` has apis `[1..102]`; `admin/user/guest` have empty `apis`. No API is currently usable by non-super roles.

## Next steps for full coverage
1) Enumerate all frontend service calls (see `frontend/apps/admin/src/services/*.ts`), list expected paths/methods.
2) Cross-check each against `sys_api_resources`; insert missing rows. (InternalMessage* set added: ids 105–113; tenants_exists id=103; tenants_with_admin id=104.)
3) Decide per-role API grants and populate `sys_roles.apis` for `admin`, `user`, `guest` (or derive from Casbin policies if syncing). Currently only `super` owns all APIs including 103–121.
4) Re-run smoke tests with non-super tokens to ensure 200/403 align with design.
