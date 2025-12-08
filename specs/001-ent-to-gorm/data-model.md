# Data Model: ent-to-gorm API parity

## Entities

- **User**: id, username, email/phone, hashed password, status, tenant_id, dept_id, roles; audit fields (created_at, updated_at, deleted_at). Validation: unique username; active status required for auth.
- **Tenant**: id, name, package/tier, admin_user_id, status, metadata json; relations to departments, roles, menus.
- **Role**: id, name, code, scope, status; relations to permissions/menus and users.
- **Permission/Menu**: id, type (menu/button/api), path/key, parent_id, sort, visibility, i18n label; relations to roles and API resources.
- **Organization/Department**: id, name, parent_id, order, status; relations to users and tenants.
- **Dict/Config**: id, category, key, value, sort, status; localization fields where present.
- **Task/Scheduler Job**: id, name, cron/trigger, payload JSON, status, next_run_at/last_run_at; logs capture run status, duration, error.
- **Audit/Logs**: login_log and operation_log retain user_id/tenant_id, request path, method, status, latency, IP/location, user agent; immutable once written.
- **File/Object storage entry**: id, path/url, provider, content_type, size, checksum, owner/tenant; soft-delete supported.

## Relationships

- Tenant 1..N Departments, Roles, Users, Menus/Permissions.
- User N..M Roles; Role N..M Permissions/Menus.
- Menu hierarchical via parent_id; Permissions tied to routes/actions.
- Tasks produce TaskLogs 1..N; Task references tenant/user that created/owns it.
- Files belong to tenants and optionally to users or domain entities via foreign keys.

## Validation Rules (from spec)

- API contracts must remain identical: field names/types/defaults unchanged.
- Create/update flows must enforce prior validation semantics and prevent partial writes on failure (transactional).
- New records written via gorm must stay readable by existing clients with no extra transformation.
- Schema and constraints remain compatible; migrations reversible.
