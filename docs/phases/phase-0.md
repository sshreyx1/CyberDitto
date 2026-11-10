# Phase 0: Foundations and hygiene — plan

Full acceptance criteria: `docs/MASTER_PLAN.md §10` (Phase 0).

## Files/decisions

- State files (`PROGRESS.md`, `DECISIONS.md`, `OWNER_TODO.md`, `BLOCKERS.md`), `CLAUDE.md`, `.claude/settings.json` — done.
- Repo hygiene docs (LICENSE, README, SECURITY, CONTRIBUTING, CODE_OF_CONDUCT) — done.
- Commit-msg attribution guard: `.githooks/commit-msg` + `core.hooksPath` set locally; mirrored in CI (`commit-hygiene.yml`) — done.
- `.gitignore` repo-hygiene guard (blocks `security_posture.json`, `.vagrant*`, `node_modules/`, `.env*` except example, anything under `legacy/`) — done.
- ADRs 0001–0010 — done.
- `docs/threat-model.md` v0, `docs/architecture.md` — done.
- `api/openapi.yaml` with a health endpoint — done; auth stubs land with Phase 1 auth work.
- Go side of the monorepo (`go.work`, `services/api` with a real health endpoint + a domain-layer unit test) — done.
- JS/TS side of the monorepo (pnpm workspace, Turborepo, `apps/web`) — **not done this session**: `pnpm` was not installed on this machine; Corepack enablement + `apps/web` scaffold is the next action.
- `Taskfile.yml`, devcontainer, `deploy/compose/dev.yml`, CI workflows (`ci.yml`, `security.yml`, `commit-hygiene.yml`) — Go-side CI and commit-hygiene land this session; devcontainer/compose/full web CI are next actions.

## Risks / open items

- `gitleaks`, `lefthook`, `golangci-lint` are not installed locally; CI installs them, so secret-scanning and lint are enforced there even before they're runnable locally. Local pre-commit enforcement for anything beyond the attribution-guard hook is a near-term follow-up (tracked in `docs/PROGRESS.md`).
- Branch protection on `main` needs the owner (see `docs/OWNER_TODO.md` item 4); until then Phase 0 commits land directly on `main` per MASTER_PLAN §0.5.4.
