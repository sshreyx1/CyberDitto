# Progress

> Read this first at the start of every session (per `CLAUDE.md`). Then `docs/BLOCKERS.md` and `docs/OWNER_TODO.md`.

**Current phase:** 1 (Walking skeleton) — in progress. IAM (auth, RBAC, hash-chained audit log) is built and verified against a real Postgres; HTTP handlers/middleware, the `fake` LabProvider, and the console shell are next.
**Last updated:** 2026-11-11.

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

## Phase 1 checklist

- [x] Postgres schema + `goose` migrations (`services/api/internal/adapters/postgres/migrations/0001_init.sql`): `workspaces`, `users`, `memberships`, `sessions`, `audit_events`, `assets`, `snapshots`, `twins`; RLS enabled on the five workspace-scoped tables; a dedicated non-superuser `cyberditto_app` role (docs/MASTER_PLAN.md §6.2) — **verified against a real dockerized Postgres 18.6**, including manually confirming RLS actually blocks a cross-workspace INSERT/SELECT as `cyberditto_app` (the owner role is RLS-exempt by default, so the app must never connect as the owner — see `docs/adr/0012-integration-test-strategy.md` and `.env.example`'s `CD_APP_DATABASE_URL`)
- [x] `sqlc` wired up (`internal/adapters/postgres/sqlc.yaml`, generated code in `gen/`) for iam/audit/assets/twins queries
- [x] `cmd/migrate` (`up`/`down`/`status`/`grant-app-login`) — the LOGIN password for `cyberditto_app` is granted out of band (env var `CD_APP_DB_PASSWORD`), never baked into a migration; dev compose does the equivalent via `deploy/compose/initdb/01-app-role.sh` on first container init
- [x] `deploy/compose/dev.yml`: Postgres 18.6-alpine (note: 18+ images need the volume mounted at `/var/lib/postgresql`, not `.../postgresql/data` — the first attempt crash-looped until this was fixed)
- [x] Domain layer: `internal/domain/iam` (Role/Capability/RBAC `Can()` policy function + full authZ-matrix unit tests — every role × every capability, plus the approver-gate invariant), `internal/domain/audit` (hash-chain `ComputeHash`/`Verify`, pure, unit-tested)
- [x] App layer: `internal/app/iam.Service` — `Setup` (first-run, refuses if any user already exists), `Login`/`Logout`/`Authenticate`, same error for wrong-password vs unknown-email (no enumeration)
- [x] Adapters: `internal/adapters/postgres` (`IAMRepo`, `AuditRepo` — the audit repo takes a `pg_advisory_xact_lock` per workspace so the hash chain can't fork under concurrent writers), `internal/adapters/argon2` (Argon2id hasher, OWASP-baseline params, unit-tested), `internal/adapters/token` (256-bit session tokens)
- [x] **Real integration tests, not fabricated**: `internal/adapters/postgres/integration_test.go` (build tag `integration`) runs Setup→Login→Authenticate→Logout, cross-workspace RLS isolation, and audit hash-chain verification against the actual dockerized Postgres — all passing. `ci.yml`'s `go` job now runs a `postgres:18.6-alpine` service container and executes these in CI too.
- [ ] HTTP transport (handlers for setup/login/logout/me, session-cookie middleware, CSRF, rate limiting on auth) — **next action**
- [ ] `fake` LabProvider + twin lifecycle state machine + SSE progress events — not started
- [ ] Console shell in `apps/web` (layout, nav, auth pages) — not started
- [ ] TOTP, WebAuthn passkeys — deliberately deferred within Phase 1: password + Argon2id + sessions + RBAC + audit is the security-critical core (directly fixes the v1 defects in MASTER_PLAN §2.3 #1-2) and is built and tested first; 2FA layers on top without changing what's already verified. Tracked here, not forgotten.
- [ ] Assets/Snapshots upload + schema validation + signature check (fixture snapshots) — schema exists (`assets`, `snapshots` tables + queries); upload handler not built yet
- [ ] AuthZ-matrix tests exist at the domain level (`rbac_test.go`); the HTTP-route × role matrix (MASTER_PLAN §9.4) lands once handlers exist
- [ ] e2e: login → add asset → build fake twin → watch progress — depends on the above

## Important environment note (read before running Go tests on this Windows host)

Running a freshly-built Go test binary directly on this Windows machine is blocked by a host **Application Control policy** ("An Application Control policy has blocked this file") — this is outside the project directory and not something this automation can or should change (MASTER_PLAN §0.5.3.2). **Workaround that works and was used to verify Phase 0:** run Go inside the existing WSL2 **Ubuntu** distro (already present on this machine), where `go build`/`go test` execute normally. Go 1.26.3 was installed there under `~/sdk/go` (user-space, no `sudo`, consistent with §0.5.3.2's allowance for installing dev dependencies inside a WSL2 sandbox). Example:
```
wsl -e bash -lc 'export PATH=$HOME/sdk/go/bin:$PATH && cd /mnt/c/Users/shreya/Documents/GitHub/CyberDitto/services/api && go build ./... && go test ./... -race -count=1'
```
GitHub Actions (Linux runners) are unaffected by this and are the authoritative CI signal either way. The same applies to Node/pnpm: Corepack couldn't write into the native `C:\Program Files\nodejs\`, so Node v24.14.0 + pnpm 12.15.0 are installed user-space inside the same WSL2 Ubuntu distro — `docs/adr/0009-monorepo-layout.md` has the detail.

**First CI run caught a real issue**: `security.yml`'s `govulncheck` job failed on the first push because `go.mod`/`go.work` were pinned to `go1.26.3` (the version installed on this host), which has 14 stdlib vulnerabilities already fixed in `go1.26.10` (released 2026-10-19). Re-pinned to `go1.26.10` everywhere (go.mod, go.work, both CI workflows), re-verified `build`/`vet`/`test`/`govulncheck` clean via a `go1.26.10` toolchain installed in WSL2, and recorded the reasoning in `docs/adr/0002-backend-stack.md`.

**Phase 1 environment notes:**
- The WSL2 service itself hung once mid-session (`wsl -e ...` returned `Error code: Wsl/Service/0x8007274c` even for `echo hello`). Fixed with `wsl --shutdown` (a plain service restart, not a host config change) followed by a normal `wsl -e` call; no data was lost. If this recurs, that's the fix.
- `go run`/`go test` binaries are blocked by the same Application Control policy from `localhost` **and** from a bare `go run` too (confirmed while testing `cmd/migrate`), not just `go test` binaries as first thought.
- WSL2's own network namespace cannot reach `localhost:<port>` for ports Docker Desktop publishes to Windows (confirmed: `docker ps` works from WSL2 via the shared daemon socket, but a raw TCP connect to `localhost:5432` from WSL2 hangs). **The reliable way to run Go code against the dev Postgres, from this host, is a throwaway container on the compose network**, e.g.:
  ```
  docker run --rm --network cyberditto-dev_default -v "$(pwd):/work" -w /work/services/api \
    -e CD_DATABASE_URL="postgres://cyberditto:dev_only_password_change_me@postgres:5432/cyberditto?sslmode=disable" \
    golang:1.26.10 go run ./cmd/migrate up
  ```
  (On Windows Git Bash, prefix with `MSYS_NO_PATHCONV=1` or the container paths in `-v`/`-w` get mangled.) This is purely a local-dev-loop convenience issue; GitHub Actions runners have none of this duality (see `docs/adr/0012-integration-test-strategy.md`).
- `deploy/compose/dev.yml`'s Postgres container crash-looped on first boot: Postgres 18+ images require the data volume mounted at `/var/lib/postgresql`, not the older `/var/lib/postgresql/data`. Fixed; see the comment in that file.

## Owner-directed changes (2026-11-10, after initial Phase 0 push)

- **License changed to All Rights Reserved** (was MIT). Repo stays public; code is explicitly not open source. Updated: `LICENSE`, `README.md`, `CONTRIBUTING.md` (no longer invites external PRs), `SECURITY.md` wording, `docs/adr/0001-license-and-repo-identity.md` (amended), `docs/DECISIONS.md`. Mentions of "open source"/Scorecard badges/contribution flow elsewhere in `docs/MASTER_PLAN.md` (§8.6, §13) are superseded by this ADR — treat the ADR as authoritative over those passages.
- **Caldera brought into main v2.0 scope** (was deferred to stretch). Now an in-scope Phase 4 deliverable alongside Atomic Red Team — see `docs/adr/0011-caldera-in-scope.md`. No code impact yet (Phase 4 is still several phases away); recorded now so the Phase 4 plan is built around it from the start rather than retrofitted.
- **`.claude/` removed** at the owner's request. Note for future sessions: this previously carried the `attribution: {commit: "", pr: "", sessionUrl: false}` config and a `permissions.deny` list (blocking force-push, remote changes, `gh repo delete`, etc.) — the no-attribution rule is still enforced independently by `.githooks/commit-msg` + `commit-hygiene.yml` + the owner's standing chat instruction, but the `permissions.deny` guardrails no longer exist as repo-committed config. Nothing else currently depends on `.claude/` existing.

## Next action

1. HTTP transport layer: `POST /api/v1/auth/setup`, `/auth/login`, `/auth/logout`, `GET /api/v1/me`, session-cookie middleware (httpOnly/Secure/SameSite, `__Host-` prefix), CSRF protection on state-changing requests, rate limiting on auth endpoints, wire into `cmd/api/main.go` (DB pool, services, router). Update `api/openapi.yaml` with these paths.
2. HTTP-level authZ-matrix tests (every route × every role, table-driven, per MASTER_PLAN §9.4) once the handlers exist.
3. `fake` `LabProvider` with the twin lifecycle state machine + SSE progress events (River queue not strictly required yet for a `fake` provider — evaluate when wiring it whether to introduce River now or with the first real background job).
4. Console shell in `apps/web` (layout, nav, generated API client — `oapi-codegen`/`openapi-typescript` codegen is still a placeholder in `Taskfile.yml`'s `gen` task — auth pages, Assets/Twins pages).
5. River queue, MinIO (for `CD_BLOB_BACKEND=s3`) — add to `deploy/compose/dev.yml` when the first feature actually needs them (snapshot upload storage, async twin builds).
6. Branch protection on `main` stays on `docs/OWNER_TODO.md` (item 4) — current `gh` token scope doesn't include repo-admin.
7. TOTP and WebAuthn passkeys (deferred within Phase 1, see checklist above).
