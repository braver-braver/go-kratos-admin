<!--
Sync Impact Report
- Version change: N/A -> 1.0.0
- Modified principles: Template placeholders -> I. Production-Ready Code Quality; II. Testable-First Delivery; III. Consistent Admin UX; IV. Performance & Reliability Budgets; V. External Dependencies & Research Discipline
- Added sections: Engineering Constraints & Non-Functional Standards; Development Workflow & Quality Gates
- Removed sections: None
- Templates requiring updates: ✅ .specify/templates/plan-template.md; ✅ .specify/templates/spec-template.md; ✅ .specify/templates/tasks-template.md; ⚠️ No commands templates present
- Follow-up TODOs: None
-->
# Go Kratos Admin Constitution

## Core Principles

### I. Production-Ready Code Quality
- Code MUST stay readable, idiomatic Go/TypeScript, and pass gofmt/golangci-lint plus eslint/prettier gates; reject dead, commented-out, or untyped code paths.
- Boundaries stay explicit: domain services expose clear interfaces, errors are wrapped with actionable context, and structured logs include correlation/trace IDs.
- Concurrency and lifecycle are intentional: every goroutine honors context cancellation/timeouts; shared state guarded or avoided; recover only with logging and metrics.
*Rationale*: Maintainability and operational safety depend on clarity, observability, and predictable execution.

### II. Testable-First Delivery
- No feature, refactor, or bug fix merges without automated tests that cover the new or changed logic; regression tests accompany every defect fix.
- Design for testing: isolate external calls behind interfaces, inject dependencies, and provide fakes/mocks; tests must be deterministic (no real network/time reliance without control).
- Coverage balance: unit tests for logic, integration/contract tests for service and DB boundaries, and UI tests for critical user journeys; document gaps with justification.
*Rationale*: Reliable change velocity requires executable proof and guardrails.

### III. Consistent Admin UX
- UI changes MUST use the established Vben/Ant Design Vue components, spacing, typography, and interaction patterns; avoid bespoke one-off styling.
- Every user flow includes loading, empty, and error states; inputs include validation and accessible labels; strings belong in locale files to preserve i18n.
- Behavior stays predictable across tenants: navigation, permissions messaging, and feedback patterns mirror existing screens unless explicitly redesigned.
*Rationale*: Consistency reduces user friction and lowers support cost.

### IV. Performance & Reliability Budgets
- Backend APIs target p95 latency ≤300ms for standard read/list operations under expected load; set explicit budgets for heavier flows and track them.
- Prevent waste: avoid N+1 queries, ensure indexes/migrations exist before rollout, use caching thoughtfully, and apply timeouts/retries with backoff where safe.
- Instrument everything new with metrics and traces; add lightweight load/perf checks for critical paths before merging.
*Rationale*: Predictable performance keeps the admin experience responsive and stable.

### V. External Dependencies & Research Discipline
- When introducing or changing external packages (e.g., gorm, go-kratos middleware, ent, Vue plugins), consult Context7 DeepWiki for best practices, record the finding with “此结论来自 DeepWiki 文档” in notes/PRs, and follow the guidance.
- Wrap dependencies behind interfaces to keep code testable; pin versions and document upgrade plans; prefer minimal surface area and remove unused packages promptly.
- Provide mocks/fakes for dependency interactions and validate integration via contract tests.
*Rationale*: Disciplined dependency use avoids lock-in, surprises, and untestable code.

## Engineering Constraints & Non-Functional Standards

- Technology stack: Go + go-kratos with wire; data via ent or gorm when justified by DeepWiki-backed guidance; frontend TypeScript + Vue + Vben Admin + Ant Design Vue.
- Observability: structured logging, tracing propagation with context, and metrics for every new endpoint or background job; error logs must not leak secrets.
- Data and API: migrations are versioned and reversible; API changes are documented in OpenAPI/Swagger with backward-compatibility plans or versioning.
- Security and compliance: enforce authentication/authorization on backend endpoints; secrets live in env/secret stores, never in repo; validate inputs/outputs for admin safety.
- Documentation: PRs/specs capture performance budgets, UX acceptance criteria, and testing strategy; inline comments explain non-obvious decisions only.

## Development Workflow & Quality Gates

- Planning artifacts (plan/spec/tasks) MUST list: test strategy per user story, UX acceptance (states, accessibility, locale updates), performance budgets, and dependency research outcomes (with DeepWiki citations when used).
- Constitution Check for every PR:
  - Code quality: lint/format clean; clear interfaces and context-aware lifecycle; no dead or panicking code without recovery and logging.
  - Testing: unit + integration/contract tests updated or added; deterministic CI; regression tests for fixes; test debt documented with owner/date.
  - UX: uses Vben/AntD patterns; loading/error/empty states; locale strings updated; screenshot or demo notes for impactful UI changes.
  - Performance/reliability: budgets stated; traces/metrics added; evidence of perf check or reason it is unnecessary.
  - External dependencies: DeepWiki research noted for gorm/other libs; dependencies wrapped for testability; versions pinned and reviewed.
- Pre-merge evidence includes lint/test results, perf/UX notes, and any rollout/rollback considerations for risky changes.

## Governance

- This constitution supersedes conflicting process docs; deviations require a written justification in the PR description and sign-off from a maintainer.
- Amendments require: proposed diff, impact analysis (principles and templates), and a version bump per semantic rules; migration/communication plans accompany breaking governance changes.
- Compliance reviews occur on every PR and during periodic audits; findings are tracked and resolved before release where risk is high.

**Version**: 1.0.0 | **Ratified**: 2025-12-05 | **Last Amended**: 2025-12-05
