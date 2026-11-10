# Decisions log

Running index of decisions taken autonomously during the build. Full rationale and rejected alternatives live in `docs/adr/`.

| # | Decision | ADR |
|---|---|---|
| 1 | MIT license kept; repo-local commit identity uses GitHub no-reply address | [0001](adr/0001-license-and-repo-identity.md) |
| 2 | Go 1.26, stdlib routing, modular-monolith/hexagonal layering | [0002](adr/0002-backend-stack.md) |
| 3 | Server-side sessions + Argon2id + passkeys/TOTP, no client-signed tokens | [0003](adr/0003-auth.md) |
| 4 | PostgreSQL for state + queue (River), SSE for realtime | [0004](adr/0004-queue-and-data.md) |
| 5 | `LabProvider` port; `fake` + `hyperv` adapters; no Vagrant | [0005](adr/0005-provider-abstraction.md) |
| 6 | Atomic Red Team catalog pinned to a commit SHA; `destructive` tier hard-blocked in code | [0006](adr/0006-catalog-pinning.md) |
| 7 | LLM orchestration server-side only, config-driven model routing, approval-gated mutating tools | [0007](adr/0007-llm-routing.md) |
| 8 | No CIS text redistribution; Wazuh/Atomic/Caldera/Sigma licensing posture | [0008](adr/0008-licensing-of-dependencies.md) |
| 9 | pnpm + Turborepo + `go.work` monorepo, single `Taskfile.yml` entry point | [0009](adr/0009-monorepo-layout.md) |
| 10 | Design system in `packages/ui` (shadcn/ui + Radix); Fumadocs for docs | [0010](adr/0010-design-system-and-docs-framework.md) |

Also decided and already in force (not separate ADRs — carried directly from `docs/MASTER_PLAN.md §0.5.2`, §12): compliance engine order (`native` then `wazuh-sca`), Caldera deferred to stretch, analytics = Plausible/Umami, AI off by default, **no AI-attribution lines in any commit or PR for this repository, ever** (owner instruction, 2026-11-10 — see `CLAUDE.md` and `docs/MASTER_PLAN.md §12`).
