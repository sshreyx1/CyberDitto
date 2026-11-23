# Progress

> Read this first at the start of every session (per `CLAUDE.md`). Then `docs/BLOCKERS.md` and `docs/OWNER_TODO.md`.

**Current phase:** 1 (Walking skeleton) — complete. Phase 2 (collector and real twin) starting; the real `LabProvider` work within it is owner-blocked (see `docs/OWNER_TODO.md` items 6–7), so the collector is the active front.
**Last updated:** 2026-11-23.

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
- [x] HTTP transport: `GET /healthz`, `GET /api/v1/auth/setup-status`, `POST /api/v1/auth/{setup,login,logout}`, `GET /api/v1/me`; session-cookie middleware (`__Host-` prefix when `CD_ENV≠dev`, httpOnly, SameSite=Lax), Origin-based CSRF check on state-changing methods, 1 MiB request body cap, `requireCapability()` as the single authZ checkpoint (not yet exercised by a real mutating route — Assets/Twins handlers will be the first). `api/openapi.yaml` updated to match. **Verified** with fast unit tests against new in-memory `internal/adapters/fake` repos (`fake.IAMRepo`/`fake.AuditRepo`, reusing the real domain hash-chain logic) — full setup→login→/me→logout flow, wrong-password/unknown-email giving the same response, double-setup rejection, session-cookie flags, and a cross-origin CSRF rejection. `cmd/api/main.go` wires the real Postgres adapters.
- [x] Rate limiting on `/auth/login` and `/auth/setup`: in-memory sliding-window limiter (`internal/transport/http/ratelimit.go`), 10 attempts/minute, shared budget across both routes keyed by caller IP (deliberately not `X-Forwarded-For` — nothing establishes a trusted proxy yet). Single-instance only; revisit before running more than one control-plane replica. Unit-tested, including the shared-budget behaviour.
- [x] `fake` LabProvider + twin lifecycle state machine + SSE progress events: `internal/domain/twin` (pure state machine mirroring the `twins` table's CHECK constraint), `internal/adapters/fake.LabProvider` (timer-driven build simulation) + `fake.TwinRepo` (in-memory — Phase 1's ask was explicitly the fake provider, not real persistence; that's Phase 2), `internal/app/twins.Service` (create/start/stop/destroy, append-only event log, in-process pub/sub for live SSE). HTTP transport: create/list/get/start/stop/destroy plus `GET .../events` (SSE, replays recorded events then streams live, deduplicated by sequence). All unit- and integration-tested, including a real SSE stream read in a test.
- [x] Console shell in `apps/web`: hand-written typed API client (`lib/api.ts`), `/setup` and `/login` pages, a `(console)` route group with a nav layout + session gate, and a `/twins` page (create, list, live SSE build log, start/stop/destroy). **Verified end to end in a real browser** against the actual Go API (fake lab provider) and a throwaway Postgres container: first-run setup → login → create twin → watch live SSE progress to Ready → start → stop → destroy → logout → unauthenticated visit to `/twins` redirects to `/login`. No CSP or CORS console errors during any of it.
- [x] **CORS** (not originally itemised here, but required for the above to work at all): explicit allow-list, credentialed, no wildcard (`internal/transport/http/cors.go`) — there was no CORS support before this, which would have silently broken every browser call from the web app's origin to the API's.
- [x] The first HTTP-level authZ-matrix test (every **twins** route × every role), driven through the real middleware chain now that a real capability-gated route exists to exercise it against (`internal/transport/http/twin_handlers_test.go`). The route × role matrix grows as more capability-gated routes (Assets, membership management, emulation/remediation) land.
- [ ] TOTP, WebAuthn passkeys — deliberately deferred within Phase 1: password + Argon2id + sessions + RBAC + audit is the security-critical core (directly fixes the v1 defects in MASTER_PLAN §2.3 #1-2) and is built and tested first; 2FA layers on top without changing what's already verified. Tracked here, not forgotten.
- [ ] Assets/Snapshots upload + schema validation + signature check (fixture snapshots) — schema exists (`assets`, `snapshots` tables + queries); upload handler not built yet. Twins currently accept any UUID as `snapshot_id` without validating it against a real snapshot — a known, documented gap until this lands.

Phase 1 is otherwise complete: the full journey (login → create a fake twin → watch its build → start/stop/destroy it) works end to end in a real browser against the real API, matching the phase's acceptance criterion. Assets/Snapshots upload and 2FA are deliberately deferred past Phase 1, not blocking it — see the two items above.

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

## Owner-directed changes (2026-11-21)

- **`CODE_OF_CONDUCT.md` and `CONTRIBUTING.md` removed.** This is a solo portfolio project, not one accepting outside contributions, and the all-rights-reserved license already makes that explicit — a contributor workflow and conduct policy for a project that doesn't take PRs was dead weight. `README.md` rewritten to lead with product/architecture/security content instead of repo-hygiene boilerplate (diagrams reused from `docs/MASTER_PLAN.md §5`; no screenshots yet since there's no real UI to show honestly — the README's own Status section says so directly).

## Phase 1 verification (2026-11-23)

Full manual E2E pass in a real browser, against the real Go API (not a mock), with a throwaway Postgres container (migrated, then discarded — the persistent dev volume's leftover test-fixture data was left untouched rather than truncated): first-run setup → login → create a twin → watch the fake provider's build (imaging → provisioning → applying → verifying → ready) stream live over SSE → start → stop → destroy → logout → confirmed an unauthenticated visit to `/twins` redirects to `/login`. Zero CSP or CORS errors in the browser console throughout. `go build/vet/test -race` and `pnpm typecheck/build` both clean on everything touched this session.

## Next action (Phase 2: collector and real twin)

Phase 2's real `LabProvider` (Hyper-V) work is **owner-blocked**: it needs a decision on a Windows 11 evaluation image/VHDX and a Wazuh hardware-layout decision (`docs/OWNER_TODO.md` items 6–7), plus a real Hyper-V-capable Windows host to build and test against, neither of which this automation can resolve on its own. The collector, by contrast, is self-contained Go code that runs against the actual host and needs neither — it's the active front:

1. Go collector (replaces `scripts/config_collector.ps1` from v1): consent prompt, a manifest of exactly what it gathers, dry-run mode, L1 categories (OS/patch level, services, firewall profiles/rules, Defender config, local users/groups without hashes, installed software — MASTER_PLAN §4.3/§5.10), golden-data unit tests, pseudonymisation of hostnames/users/IPs before anything is exported. Must not repeat v1's defects (MASTER_PLAN §2.3 #3: hardcoded `firewall.enabled = true`) — every value is a real read, with evidence.
2. Once the collector exists: an Assets/Snapshots upload handler (schema validation, signature check, content-addressed storage) so a real snapshot can replace the synthetic `snapshot_id` twins currently accept unvalidated.
3. Hyper-V `LabProvider`, mTLS lab gateway, twin-agent enrolment, `ApplyConfig` + re-collect for the Fidelity Score — blocked on the owner items above; do not start without them (building against an assumed image/licensing setup risks throwaway work).
4. River queue, MinIO (`CD_BLOB_BACKEND=s3`) — add to `deploy/compose/dev.yml` once snapshot upload storage or async twin builds actually need them.
5. Branch protection on `main` stays on `docs/OWNER_TODO.md` (item 4) — current `gh` token scope doesn't include repo-admin.
6. TOTP and WebAuthn passkeys (deferred within Phase 1, see checklist above) — pick up whenever convenient; nothing downstream depends on them yet.
