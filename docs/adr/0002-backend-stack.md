# 0002: Backend language and core stack

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v1 used Go 1.23 with `gorilla/mux`. Needed: a boring, testable, strongly-typed backend that can shell out safely to hypervisor/attack tooling without shell-injection risk, and that supports clean/hexagonal layering (MASTER_PLAN §5.1, §5.3).

## Decision

- **Go 1.26** (local toolchain confirmed at `go1.26.3`, matching the plan's target — satisfies the "verify before pinning" rule in §0.3).
- Standard library `net/http` `ServeMux` (Go 1.22+ method+path routing) instead of a router framework; add `chi` only if middleware ergonomics demand it later.
- `log/slog` for structured logging from day one.
- Modular monolith: one control-plane binary (`services/api`), internal packages split into `domain` (no I/O), `app` (use cases), `ports` (interfaces), `adapters` (concrete I/O), `transport` (HTTP/MCP/lab-gateway), enforced by `depguard` in CI once lint is wired up.

## Rejected alternatives

- `gorilla/mux` (v1's choice) — stdlib routing is sufficient since Go 1.22 and removes a dependency.
- A microservice split — unjustified complexity at this stage; ports/adapters already give an extraction path if needed later.
