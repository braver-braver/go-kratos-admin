# Schema Parity Check (ent vs gorm)

- Source DB: `timescaledb.pg15-timescale.orb.local:5432` (postgres)
- Method: gorm/gen enumeration after ent baseline; generated 27 tables from live schema.
- Tables generated: sys_user_credentials, sys_menus, sys_admin_operation_logs, sys_api_resources, sys_user_role, sys_departments, files, internal_messages, internal_message_recipients, internal_message_categories, sys_admin_login_logs, sys_languages, sys_positions, sys_admin_login_restrictions, sys_role_api, sys_role_dept, sys_role_menu, sys_role_org, sys_role_position, sys_roles, sys_tasks, sys_dict_types, sys_user_position, sys_dict_entries, sys_users, sys_organizations, sys_tenants.
- Action: gorm/gen output placed in `backend/app/admin/service/internal/data/gormcli/model` and `generated/`, reflecting current ent schema.
- Result: No missing tables detected relative to ent schema; further validation needed for indexes/constraints when implementing gorm repos.
