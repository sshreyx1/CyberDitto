# Progress

> Read this first at the start of every session (per `CLAUDE.md`). Then `docs/BLOCKERS.md` and `docs/OWNER_TODO.md`.

**Current phase:** 0 (Foundations and hygiene) — in progress.
**Last updated:** 2026-11-10.

## Phase 0 checklist

- [x] Fresh repo (confirmed empty on GitHub before starting — no old commits, no forks), MIT `LICENSE`, `README.md` stub, `SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`
- [x] `.claude/settings.json`, `CLAUDE.md`
- [x] Commit-msg attribution guard (`.githooks/commit-msg` + `core.hooksPath` set locally) — **verified**: rejects a message containing a `Co-Authored-By: Claude` trailer, accepts a clean one
- [x] Repo-local commit identity set to the GitHub no-reply address (`174120481+sshreyx1@users.noreply.github.com`), not the personal email that was in global git config
- [x] `.gitignore` repo-hygiene guard (blocks `security_posture.json`, `*.vagrant*`, `.vagrant/`, `node_modules/`, `.env*` except `.env.example`, anything under `legacy/`)
- [x] `docs/threat-model.md` v0, `docs/architecture.md`
- [x] ADRs 0001–0010 (`docs/adr/`), indexed in `docs/DECISIONS.md`
- [x] `api/openapi.yaml` with a health endpoint (auth stubs land with Phase 1 auth work)
- [x] Go side of the monorepo: `go.work` + `services/api` (Go 1.26.3) with `cmd/api` (health endpoint on stdlib `ServeMux`) and `internal/domain`, `internal/transport/http` — **build, vet and tests verified green** (see note below on how)
- [x] `Taskfile.yml` (dev/build/test/vet/lint wired to the Go module; gen/demo/e2e/release are placeholders until the phases that need them)
- [x] CI skeleton: `.github/workflows/ci.yml` (Go build/vet/test + repo-hygiene file-path guard), `security.yml` (gitleaks + govulncheck), `commit-hygiene.yml` (rejects AI-attribution trailers on PRs) — action versions pinned to verified commit SHAs (fetched via `gh api`, not fabricated), each ≥ 14 days old per §0.3
- [x] State files: this file, `docs/DECISIONS.md`, `docs/OWNER_TODO.md`, `docs/BLOCKERS.md`
- [ ] Monorepo scaffold, JS/TS side (`apps/web`, `packages/*`, pnpm workspace + Turborepo) — **not started**. `pnpm` is not installed on this machine (Node v24.14.0 is). **Next action.**
- [ ] Devcontainer, `deploy/compose/dev.yml` dev stack — not started (no Postgres/MinIO consumer exists yet; follows the web scaffold and Phase 1 auth/DB work)
- [ ] `lefthook` pre-commit hooks (gofmt/goimports/golangci-lint/eslint/prettier/tsc) — not installed locally; the attribution guard above already runs via a plain git hook independent of lefthook. Full lefthook wiring is a near-term follow-up once the JS toolchain exists.
- [x] First real commits pushed to `main`; confirmed via `gh api` that `git log` on GitHub shows only fresh history, correct author (`174120481+sshreyx1@users.noreply.github.com`), and no attribution text in any commit message; `ci` and `security` workflows both green on GitHub Actions (`fdb7e56`)

## Important environment note (read before running Go tests on this Windows host)

Running a freshly-built Go test binary directly on this Windows machine is blocked by a host **Application Control policy** ("An Application Control policy has blocked this file") — this is outside the project directory and not something this automation can or should change (MASTER_PLAN §0.5.3.2). **Workaround that works and was used to verify Phase 0:** run Go inside the existing WSL2 **Ubuntu** distro (already present on this machine), where `go build`/`go test` execute normally. Go 1.26.3 was installed there under `~/sdk/go` (user-space, no `sudo`, consistent with §0.5.3.2's allowance for installing dev dependencies inside a WSL2 sandbox). Example:
```
wsl -e bash -lc 'export PATH=$HOME/sdk/go/bin:$PATH && cd /mnt/c/Users/shreya/Documents/GitHub/CyberDitto/services/api && go build ./... && go test ./... -race -count=1'
```
GitHub Actions (Linux runners) are unaffected by this and are the authoritative CI signal either way.

**First CI run caught a real issue**: `security.yml`'s `govulncheck` job failed on the first push because `go.mod`/`go.work` were pinned to `go1.26.3` (the version installed on this host), which has 14 stdlib vulnerabilities already fixed in `go1.26.10` (released 2026-10-19). Re-pinned to `go1.26.10` everywhere (go.mod, go.work, both CI workflows), re-verified `build`/`vet`/`test`/`govulncheck` clean via a `go1.26.10` toolchain installed in WSL2, and recorded the reasoning in `docs/adr/0002-backend-stack.md`.

## Next action

1. Enable `pnpm` (via Corepack, bundled with Node 24) and scaffold `apps/web` as a minimal Next.js 16 app (health/placeholder page) plus `packages/{ui,config}` stubs and the root `pnpm-workspace.yaml` + `turbo.json`.
2. Extend `ci.yml` with a web job (typecheck/lint/build) once `apps/web` exists.
3. Make the first commit (this Phase 0 foundation work) and push directly to `main` per MASTER_PLAN §0.5.4; after that, Phase 0 continues on `phase-0/*` branches for anything further, and branch protection goes on the owner's list (`docs/OWNER_TODO.md` item 4) since the current `gh` token scope doesn't include repo-admin.
4. Devcontainer + `deploy/compose/dev.yml`, then move into Phase 1 (auth, Postgres/RLS, River queue, fake `LabProvider`, console shell).
