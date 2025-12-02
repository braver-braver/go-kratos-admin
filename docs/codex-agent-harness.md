# Codex Long-Running Agent Harness / Codex 长时段智能体工作流

Adapted from Anthropic’s “Effective harnesses for long-running agents” to fit this repo’s Codex workflow. Use this alongside the CoT→AoT instructions in `AGENTS.md`.

改编自 Anthropic《Effective harnesses for long-running agents》，用于本仓库的 Codex 工作流；与 `AGENTS.md` 的 CoT→AoT 指南配合使用。

## Roles / 角色
- **Initializer agent (first session only)**: lay down scaffolding and baseline state.
- **Coding agent (every subsequent session)**: make incremental, well-tested progress and leave crisp breadcrumbs.

## Required artifacts / 关键产物
1. `init.sh` — starts required services, installs deps, runs smoke tests; idempotent.
2. `feature-list.json` — structured list of end-to-end features from the user prompt. Fields: `id`, `title`, `description`, `steps` (array of user-visible checks), `status` (`failing|in_progress|passing`), `notes`. Default `status=failing`.
3. `codex-progress.json` — append-only session log: timestamp, agent type, branch, features touched, commits, tests run (commands + results), manual checks, open issues/TODOs.

### Minimal templates
`feature-list.json`
```json
[
  {
    "id": "auth-login-basic",
    "title": "User can log in",
    "description": "Login form accepts valid credentials and navigates to dashboard.",
    "steps": [
      "Open /login",
      "Enter valid username/password",
      "Press submit",
      "See dashboard and session cookie set"
    ],
    "status": "failing",
    "notes": ""
  }
]
```

`codex-progress.json`
```json
[
  {
    "timestamp": "2025-01-01T12:00:00Z",
    "agent": "initializer",
    "branch": "main",
    "features": ["auth-login-basic"],
    "commits": [],
    "tests": [
      {"cmd": "backend/make lint", "result": "pass"},
      {"cmd": "frontend/pnpm lint", "result": "pass"}
    ],
    "manual_checks": ["Opened /login, verified form renders"],
    "notes": "Seeded feature list; no code changes yet.",
    "todos": ["Implement login flow"]
  }
]
```

## Initializer agent checklist / 初始化智能体清单
1. **Get bearings**: `pwd`, `git status`, `git log -5 --oneline`.
2. **Create/refresh `init.sh`** with:
   - Backend: `cd backend && make init` (once), `make compose-up`, `make lint`, `make test`.
   - Frontend: `cd frontend && pnpm install`, `pnpm lint`, `pnpm test:unit --dom`, `pnpm dev` (document port 5666).
   - Any seed data or env var notes.
3. **Derive `feature-list.json`** from the task/spec; include all notable end-to-end behaviors as failing entries.
4. **Seed `codex-progress.json`** with an initial entry recording the scaffold and any smoke-test results.
5. **Baseline git state**: ensure clean working tree (no commit here unless explicitly requested).

## Coding agent loop / 编码智能体循环
1. **Warm up**: `pwd`, `git status`, skim `git log -10`, read `codex-progress.json` and `feature-list.json`.
2. **Pick one feature**: choose highest-priority `failing` item; set it to `in_progress`.
3. **Plan minimally**: outline substeps + planned tests before touching code.
4. **Execute incrementally**: avoid parallel feature work; keep WIP local until tests pass.
5. **Test and self-verify** (preferred order):
   - Backend: `make lint`, `make test` (or targeted), `make cover` if quick.
   - Frontend: `pnpm lint`, `pnpm test:unit --dom`; run relevant Playwright spec if feature is UI-heavy.
   - Manual/smoke: run `init.sh` commands to bring services up, then exercise the feature like a user (document steps).
6. **Update artifacts**:
   - `feature-list.json`: flip `status` to `passing` only after tests + manual checks; never delete or rewrite other entries.
   - Append to `codex-progress.json`: session summary, code areas touched, commands/tests with outcomes, manual steps, regressions, next TODO.
7. **Commit etiquette**: keep working tree clean; use conventional commit messages; avoid bundling multiple features.

## Recovery & handoff / 恢复与交接
- If tests fail, revert the change or mark the feature back to `failing` and record the failure in `codex-progress.json`.
- Next sessions must start by running `init.sh` smoke commands and reading both JSON files before coding.
- Leave TODOs in `codex-progress.json` rather than scattered comments.

## Notes / 备注
- JSON is chosen to reduce accidental rewriting; keep files stable and append-only where possible.
- Keep docs bilingual when updating these artifacts; code comments remain concise and minimal.
