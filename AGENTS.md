# Repository Guidelines

## Project Structure & Module Organization
Root folders separate concerns: `backend` hosts the Go services (generated APIs in `api/`, service logic in `app/<service>/`, shared helpers under `pkg/`, and deployment scripts in `script/`), while `frontend` provides the Vue-based admin shell (feature apps in `apps/`, shared UI packages in `packages/`, and tooling configs inside `internal/`). Database fixtures and migrations live in `backend/sql/`, asset documentation resides in `docs/`, and cross-cutting workspace settings sit alongside this guide. Keep new modules colocated with their domain to preserve this layout.

## Build, Test, and Development Commands
- Backend quickstart: `cd backend && make init && make compose-up` installs toolchains and dependent containers.
- Backend iteration: `make run` inside a service folder hot-reloads Go binaries; `make docker` builds images for deployment.
- Frontend setup: `cd frontend && pnpm install && pnpm dev` launches the Vben shell on port 5666.
- Quality gates: `make lint` (Go) and `pnpm lint` (frontend) must pass before review.
- Release assets: `pnpm build` emits production bundles, while `make docker` plus `docker compose up` ship the full stack.

## Coding Style & Naming Conventions
Go code follows `gofmt` defaults with modules under `go-kratos-admin/<component>`; exported symbols should read `PackageVerb`, and ent schemas stay singular (`User`, `Role`). `golangci-lint` enforces style and vetting—run it prior to commits. Frontend TypeScript adopts the workspace ESLint/Prettier/Stylelint presets; name Vue components in `PascalCase.vue`, composables as `useFoo.ts`, and store packages using kebab-case directories.

## Testing Guidelines
Backend unit coverage relies on `make test`; use `make cover` when you need a quick profile. When adding services, scaffold deterministic table-driven tests and mock external clients through the existing interfaces. Frontend unit specs run with `pnpm test:unit --dom`, and Playwright end-to-ends flow through `pnpm test:e2e`; place specs beside the source in `__tests__` folders. New features should include at least one backend assertion and, when UI-facing, a Vitest or Playwright scenario.

## Commit & Pull Request Guidelines
Adopt the conventional commit scheme already in history. Use `pnpm commit` (czg) for interactive prompts and scopes drawn from workspace packages; backend changes can use the nearest module as scope (e.g., `feat(app-user): ...`). Keep PRs focused, note required migrations or env tweaks, link related issues, and attach screenshots or cURL snippets for UI/API updates. Confirm `make lint`, `make test`, `pnpm lint`, and relevant test suites before requesting review.
