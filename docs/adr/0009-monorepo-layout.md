# 0009: Monorepo layout and tooling

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v2 spans a Go control plane, a Go collector, a Go twin agent, a Next.js web app, a generated TS API client, and a shared design system. These need to build/test together with one command and share CI without becoming entangled (MASTER_PLAN §5.4, §5.5).

## Decision

- Single repository, laid out per MASTER_PLAN §5.5: `apps/web` (Next.js), `services/{api,collector,twin-agent}` (Go, joined by `go.work`), `packages/{api-client,ui,config}` (TS), `content/`, `deploy/`, `docs/`, `fixtures/`.
- **pnpm workspaces + Turborepo** for the JS/TS side; `go.work` for the Go side; a top-level `Taskfile.yml` as the single entry point (`task dev | test | lint | gen | demo | e2e | build | release`) so contributors don't need to know which tool to run where.
- Node toolchain: v24.14.0. Enabling `pnpm` via Corepack on the native Windows Node install failed (`EPERM` writing into `C:\Program Files\nodejs\`, a permissions issue outside the project directory, not something to force). As with the Go toolchain (ADR 0002), the workaround is a user-space Node v24.14.0 + pnpm (pinned `12.15.0`, Corepack-managed) install inside the existing WSL2 Ubuntu distro — no host changes, consistent with §0.5.3.2. `apps/web` (Next.js 16.4.3 + React 19.3.1) is scaffolded and verified to `pnpm install` and `next build` cleanly there.

## Rejected alternatives

- Separate repos per component — would fragment PRs that touch the API contract and both the Go server and the generated TS client at once; rejected for a project this size.
