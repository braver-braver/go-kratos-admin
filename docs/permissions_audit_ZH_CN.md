# API 权限核查（前端 → 后端 → DB）

范围：对照前端 admin API 调用与后端路由、`sys_api_resources`/`sys_roles` 的映射。状态基于当前数据库（db MCP 查询）。

## 摘要
- 租户 API：基础 CRUD 已注册，但 `/admin/v1/tenants_exists` 与 `/admin/v1/tenants_with_admin` 未出现在 `sys_api_resources`，因此非白名单调用会 403。当前只有 super 角色拥有已定义的所有 API（id 1–102）；`admin/user/guest` 的 `apis` 为空。
- 总体缺口：仅 super 可用，其他角色尚未分配任何 API，需要补充角色授权。

## 租户服务（frontend/apps/admin/src/services/tenant.service.ts）
| 前端调用 | 后端路径 | 方法 | sys_api_resources | 角色覆盖 |
| --- | --- | --- | --- | --- |
| `List` | `/admin/v1/tenants` | GET | ✔ id=36 | 仅 super |
| `Create` | `/admin/v1/tenants` | POST | ✔ id=37 | 仅 super |
| `CreateTenantWithAdminUser` | `/admin/v1/tenants_with_admin` | POST | ✔ id=104 | 仅 super |
| `Delete` | `/admin/v1/tenants/{id}` | DELETE | ✔ id=29 | 仅 super |
| `Get` | `/admin/v1/tenants/{id}` | GET | ✔ id=30 | 仅 super |
| `Update` | `/admin/v1/tenants/{data.id}` | PUT | ✔ id=22 | 仅 super |
| `TenantExists` | `/admin/v1/tenants_exists` | GET | ✔ id=103 | 仅 super |

修复建议（租户相关）：
1) 已补充 `sys_api_resources`：id=103 (`/admin/v1/tenants_exists`, GET)、id=104 (`/admin/v1/tenants_with_admin`, POST)。
2) 待办：将这两个 ID（及其他需要的 API）写入除 super 以外的角色（admin/user/guest）`sys_roles.apis`，按策略分配。

## 角色表快照
- `sys_roles`：`super` 的 `apis` 含 `[1..102]`；`admin/user/guest` 的 `apis` 为空，导致全部接口不可用（除白名单）。

## 后续全量覆盖步骤
1) 枚举前端所有 service 调用（`frontend/apps/admin/src/services/*.ts`），整理期望的路径/方法列表。
2) 逐一对照 `sys_api_resources`，补齐缺失条目。（已补：InternalMessage* 105–113，tenants_exists 103，tenants_with_admin 104）
3) 设计各角色的 API 授权，将对应 ID 写入 `sys_roles.apis`（或通过 Casbin/同步机制更新）。目前仅 super 拥有全部 API（含 103–121）。
4) 使用非 super 角色进行冒烟测试，确认 200/403 符合预期。
