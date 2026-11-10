# Progress

> Read this first at the start of every session (per `CLAUDE.md`). Then `docs/BLOCKERS.md` and `docs/OWNER_TODO.md`.

**Current phase:** 0 (Foundations and hygiene) — functionally complete, a few nice-to-haves rolled into Phase 1 (see "Next action"). Moving into Phase 1 (Walking skeleton).
**Last updated:** 2026-11-10.

## Phase 0 checklist

- [x] Fresh repo (confirmed empty on GitHub before starting — no old commits, no forks), `LICENSE` (All Rights Reserved — see note below), `README.md` stub, `SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`
- [x] `CLAUDE.md` (`.claude/` was later removed per owner request — see note below)
- [x] Commit-msg attribution guard (`.githooks/commit-msg` + `core.hooksPath` set locally) — **verified**: rejects a message containing a `Co-Authored-By: Claude` trailer, accepts a clean one
- [x] Repo-local commit identity set to the GitHub no-reply address (`174120481+sshreyx1@users.noreply.github.com`), not the personal email that was in global git config
- [x] `.gitignore` repo-hygiene guard (blocks `security_posture.json`, `*.vagrant*`, `.vagrant/`, `node_modules/`, `.env*` except `.env.example`, anything under `legacy/`)
- [x] `docs/threat-model.md` v0, `docs/architecture.md`
- [x] ADRs 0001–0011 (`docs/adr/`), indexed in `docs/DECISIONS.md`
- [x] `api/openapi.yaml` with a health endpoint (auth stubs land with Phase 1 auth work)
- [x] Go side of the monorepo: `go.work` + `services/api` (Go 1.26.10) with `cmd/api` (health endpoint on stdlib `ServeMux`) and `internal/domain`, `internal/transport/http` — **build, vet and tests verified green** (see note below on how)
- [x] `Taskfile.yml` (dev/build/test/vet/lint wired to the Go module; gen/demo/e2e/release are placeholders until the phases that need them)
- [x] CI skeleton: `.github/workflows/ci.yml` (Go build/vet/test + repo-hygiene file-path guard), `security.yml` (gitleaks + govulncheck), `commit-hygiene.yml` (rejects AI-attribution trailers on PRs) — action versions pinned to verified commit SHAs (fetched via `gh api`, not fabricated), each ≥ 14 days old per §0.3
- [x] State files: this file, `docs/DECISIONS.md`, `docs/OWNER_TODO.md`, `docs/BLOCKERS.md`
- [x] Monorepo scaffold, JS/TS side: root `package.json`/`pnpm-workspace.yaml`/`turbo.json`, `apps/web` (Next.js 16.4.3, React 19.3.1, TypeScript 7.0.2, Tailwind v4.3.3, one placeholder page) — **verified**: `pnpm install` and `next build` both succeed (via WSL2, see note below); `ci.yml` gained a `web` job (typecheck + build). `packages/{ui,config}` deliberately **not** created yet — nothing exists to extract into a shared design system/config until Phase 1 builds real console UI; adding them now would be premature scaffolding.
- [ ] Devcontainer, `deploy/compose/dev.yml` dev stack — not started (no Postgres/MinIO consumer exists yet; follows Phase 1 auth/DB work)
- [ ] `lefthook` pre-commit hooks (gofmt/goimports/golangci-lint/eslint/prettier/tsc) — not installed locally; the attribution guard above already runs via a plain git hook independent of lefthook. Full lefthook wiring is a near-term follow-up once the JS toolchain exists.
- [x] First real commits pushed to `main`; confirmed via `gh api` that `git log` on GitHub shows only fresh history, correct author (`174120481+sshreyx1@users.noreply.github.com`), and no attribution text in any commit message; `ci` and `security` workflows both green on GitHub Actions (`fdb7e56`)

## Important environment note (read before running Go tests on this Windows host)

Running a freshly-built Go test binary directly on this Windows machine is blocked by a host **Application Control policy** ("An Application Control policy has blocked this file") — this is outside the project directory and not something this automation can or should change (MASTER_PLAN §0.5.3.2). **Workaround that works and was used to verify Phase 0:** run Go inside the existing WSL2 **Ubuntu** distro (already present on this machine), where `go build`/`go test` execute normally. Go 1.26.3 was installed there under `~/sdk/go` (user-space, no `sudo`, consistent with §0.5.3.2's allowance for installing dev dependencies inside a WSL2 sandbox). Example:
```
wsl -e bash -lc 'export PATH=$HOME/sdk/go/bin:$PATH && cd /mnt/c/Users/shreya/Documents/GitHub/CyberDitto/services/api && go build ./... && go test ./... -race -count=1'
```
GitHub Actions (Linux runners) are unaffected by this and are the authoritative CI signal either way. The same applies to Node/pnpm: Corepack couldn't write into the native `C:\Program Files\nodejs\`, so Node v24.14.0 + pnpm 12.15.0 are installed user-space inside the same WSL2 Ubuntu distro — `docs/adr/0009-monorepo-layout.md` has the detail.

**First CI run caught a real issue**: `security.yml`'s `govulncheck` job failed on the first push because `go.mod`/`go.work` were pinned to `go1.26.3` (the version installed on this host), which has 14 stdlib vulnerabilities already fixed in `go1.26.10` (released 2026-10-19). Re-pinned to `go1.26.10` everywhere (go.mod, go.work, both CI workflows), re-verified `build`/`vet`/`test`/`govulncheck` clean via a `go1.26.10` toolchain installed in WSL2, and recorded the reasoning in `docs/adr/0002-backend-stack.md`.

## Owner-directed changes (2026-11-10, after initial Phase 0 push)

- **License changed to All Rights Reserved** (was MIT). Repo stays public; code is explicitly not open source. Updated: `LICENSE`, `README.md`, `CONTRIBUTING.md` (no longer invites external PRs), `SECURITY.md` wording, `docs/adr/0001-license-and-repo-identity.md` (amended), `docs/DECISIONS.md`. Mentions of "open source"/Scorecard badges/contribution flow elsewhere in `docs/MASTER_PLAN.md` (§8.6, §13) are superseded by this ADR — treat the ADR as authoritative over those passages.
- **Caldera brought into main v2.0 scope** (was deferred to stretch). Now an in-scope Phase 4 deliverable alongside Atomic Red Team — see `docs/adr/0011-caldera-in-scope.md`. No code impact yet (Phase 4 is still several phases away); recorded now so the Phase 4 plan is built around it from the start rather than retrofitted.
- **`.claude/` removed** at the owner's request. Note for future sessions: this previously carried the `attribution: {commit: "", pr: "", sessionUrl: false}` config and a `permissions.deny` list (blocking force-push, remote changes, `gh repo delete`, etc.) — the no-attribution rule is still enforced independently by `.githooks/commit-msg` + `commit-hygiene.yml` + the owner's standing chat instruction, but the `permissions.deny` guardrails no longer exist as repo-committed config. Nothing else currently depends on `.claude/` existing.

## Next action

Phase 0 is functionally complete: repo hygiene, state files, ADRs 0001–0011, CI (`ci`/`security`/`commit-hygiene`, all green), Go API module, and the Next.js web scaffold are all in place and verified. Remaining Phase 0 nice-to-haves (devcontainer, `deploy/compose/dev.yml`, `lefthook`) are low-value until Phase 1 actually needs Postgres/MinIO and a JS toolchain for pre-commit hooks, so the plan is to pick them up *as part of* Phase 1 rather than first. Next steps, in order:

1. Start Phase 1 (MASTER_PLAN §10): Postgres schema + `goose` migrations + RLS, `sqlc` repos, River queue, first-run auth wizard (Argon2id, sessions, TOTP, passkeys), RBAC + authZ-matrix tests, hash-chained audit log.
2. Stand up `deploy/compose/dev.yml` (Postgres, MinIO) alongside the DB work above, since that's the first point it's actually needed.
3. `fake` `LabProvider` with the twin lifecycle state machine + SSE progress events.
4. Console shell in `apps/web` (layout, nav, generated API client once `oapi-codegen`/`openapi-typescript` are wired up, auth pages, Assets/Twins pages).
5. Branch protection on `main` stays on `docs/OWNER_TODO.md` (item 4) — current `gh` token scope doesn't include repo-admin.
