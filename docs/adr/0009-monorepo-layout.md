# 0009: Monorepo layout and tooling

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v2 spans a Go control plane, a Go collector, a Go twin agent, a Next.js web app, a generated TS API client, and a shared design system. These need to build/test together with one command and share CI without becoming entangled (MASTER_PLAN §5.4, §5.5).

## Decision

- Single repository, laid out per MASTER_PLAN §5.5: `apps/web` (Next.js), `services/{api,collector,twin-agent}` (Go, joined by `go.work`), `packages/{api-client,ui,config}` (TS), `content/`, `deploy/`, `docs/`, `fixtures/`.
- **pnpm workspaces + Turborepo** for the JS/TS side; `go.work` for the Go side; a top-level `Taskfile.yml` as the single entry point (`task dev | test | lint | gen | demo | e2e | build | release`) so contributors don't need to know which tool to run where.
- Node toolchain confirmed locally: Node v24.14.0. `pnpm` is not yet installed on this machine — enabling it (via Corepack) is tracked as the next action for the web-app scaffold (`docs/PROGRESS.md`), not a blocker for the Go-side Phase 0 work landing first.

## Rejected alternatives

- Separate repos per component — would fragment PRs that touch the API contract and both the Go server and the generated TS client at once; rejected for a project this size.
