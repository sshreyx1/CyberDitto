# 0002: Backend language and core stack

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v1 used Go 1.23 with `gorilla/mux`. Needed: a boring, testable, strongly-typed backend that can shell out safely to hypervisor/attack tooling without shell-injection risk, and that supports clean/hexagonal layering (MASTER_PLAN §5.1, §5.3).

## Decision

- **Go 1.26.10** — the native Windows toolchain on this host is `go1.26.3`; CI's first run (`security.yml`) caught that `go1.26.3` has 14 known-fixed stdlib vulnerabilities (govulncheck, `GO-2026-*`), all fixed by `go1.26.10` (released 2026-10-19, confirmed via `go.dev/doc/devel/release`). Pinned `go.mod`/`go.work`/CI to `1.26.10` and verified clean (`govulncheck`: no vulnerabilities found) using a separately-installed `go1.26.10` toolchain in WSL2 — this is exactly the "verify at build time, security fix can override the 14-day-old rule" case §0.3 anticipates. The native Windows Go install was left untouched (upgrading it would be a host-level change outside the project directory); Go's toolchain auto-switch (via the `go` directive) or a WSL2/CI toolchain satisfies the pin without touching the host.
- Standard library `net/http` `ServeMux` (Go 1.22+ method+path routing) instead of a router framework; add `chi` only if middleware ergonomics demand it later.
- `log/slog` for structured logging from day one.
- Modular monolith: one control-plane binary (`services/api`), internal packages split into `domain` (no I/O), `app` (use cases), `ports` (interfaces), `adapters` (concrete I/O), `transport` (HTTP/MCP/lab-gateway), enforced by `depguard` in CI once lint is wired up.

## Rejected alternatives

- `gorilla/mux` (v1's choice) — stdlib routing is sufficient since Go 1.22 and removes a dependency.
- A microservice split — unjustified complexity at this stage; ports/adapters already give an extraction path if needed later.
