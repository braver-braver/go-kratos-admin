# Design & Architecture / 设计与架构

## Overview / 概览
- **Purpose / 目的**: Vue 3 + Vite admin shell for go-kratos-admin, built on Vben shared packages and Ant Design Vue for authenticated, multi-tenant console experiences.
- **Composition / 组成**: Root bootstrap (`src/main.ts` -> `bootstrap.ts`) wires preferences, i18n, Pinia stores, access directives, router, and mounts `app.vue`.
- **Workspace / 工作区**: pnpm + Turbo monorepo; shared UI/logic in `packages` & `internal`, main app in `apps/admin`.
- **State / 状态管理**: Pinia stores via `@vben/stores`; local auth store (`stores/authentication.state.ts`) coordinates tokens, user info, and navigation.
- **Networking / 网络**: `utils/request.ts` wraps `@vben/request` with interceptors for auth headers, token refresh, error messaging; services under `src/services` adapt generated protobuf DTOs to REST endpoints.
- **Routing / 路由**: Vue Router history selectable by env (`VITE_ROUTER_HISTORY`), base from `VITE_BASE`; core routes + dynamic modules merged in `router/routes`.
- **Preferences / 偏好**: `initPreferences` namespaces persisted settings by `VITE_APP_NAMESPACE` + `VITE_APP_VERSION` + env to isolate deployments.

## Key Flows / 关键流程
- **Bootstrap**: init preferences → load component adapter → create app → register global comps → setup i18n → init Pinia stores (namespaced) → install access directive → mount router guards → reactive document title.
- **Auth**: login encrypts password with AES key from env, calls `AuthenticationService`, stores token in `useAccessStore`, fetches user info & access codes, redirects to `DEFAULT_HOME_PATH`. Logout clears stores and redirects to `/login`.
- **Request Lifecycle**: request interceptor injects `Authorization` + `Accept-Language`; response interceptor unwraps data, auth interceptor handles 401 with refresh/reauth; error interceptor surfaces messages via `ant-design-vue` message component.
- **Routing Guard**: guards (in `router/guard.ts`) validate access codes vs route meta, handle whitelists, and redirect unauthenticated users to login with redirect query.

## Environment / 环境
- `.env.development/.env.production/.env.analyze` configure API base (`VITE_GLOB_API_URL` via `useAppConfig`), router history, app namespace/version, AES key for password encryption, locale defaults, and feature toggles (e.g., refresh token).

## Extensibility / 可扩展性
- Add features by creating service wrappers in `src/services`, store modules in `src/stores`, and route modules under `src/router/routes/modules`.
- UI components should prefer shared packages (`@vben/*`) and Ant Design Vue; register global components via `registerGlobComp`.
- Keep DTOs in `src/generated` in sync with backend protobufs to maintain typing across services.

## Build & Delivery / 构建与发布
- Dev: `pnpm dev` (turbo) mounts Vite dev server with hash/history per env.
- Build: `pnpm build` (turbo) → Vite production bundle; `pnpm preview` for local serve.
- Checks: `pnpm lint`, `pnpm check`, `pnpm test:unit`, `pnpm test:e2e` for quality gates.
