# Architecture overview

Full detail lives in `docs/MASTER_PLAN.md §5`; this page is a short orientation pointer, kept in sync as the system grows.

## Shape

Modular monolith (one Go control-plane binary: API + workers), hexagonal/ports-and-adapters internally, API-first via `api/openapi.yaml`, event-sourced runs/builds streamed over SSE. See the system-context and module diagrams in `docs/MASTER_PLAN.md §5.2`–`§5.3`.

## Current state (Phase 0)

- `services/api`: Go module scaffold with a health endpoint (`cmd/api`), `internal/domain` package started. No persistence, auth, or lab integration yet — those are Phase 1–2.
- `apps/web`: not yet scaffolded (next action, tracked in `docs/PROGRESS.md`).
- `api/openapi.yaml`: health endpoint only so far.

## Where decisions live

- Why a given technology/approach was chosen: `docs/adr/`, indexed in `docs/DECISIONS.md`.
- Safety invariants that attack execution must never violate: `docs/MASTER_PLAN.md §6.4`.
- Threat model: `docs/threat-model.md`.
