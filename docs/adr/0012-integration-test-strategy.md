# 0012: Integration tests run against a DSN-configured Postgres, not testcontainers-go

- **Status:** accepted
- **Date:** 2026-11-11

## Context

MASTER_PLAN §9.4 names `testcontainers-go` for adapter integration tests (Postgres, MinIO). While building Phase 1's `postgres` adapter, local verification surfaced two environment realities on this Windows 11 + WSL2 + Docker Desktop host (also recorded in `docs/PROGRESS.md`):

1. Windows Application Control blocks freshly-built Go binaries from running natively — Go work happens inside WSL2 or a throwaway `golang` container instead (ADR 0002).
2. WSL2's own network namespace cannot reach `localhost:<port>` for ports Docker Desktop publishes to the Windows host, even though the `docker` CLI itself works from both sides via the shared daemon socket. Container-to-container traffic on a compose network works reliably; a bare `go test` process's own TCP connections to `localhost` from WSL2 do not reach those published ports.

`testcontainers-go` has the test process start its own Postgres container and connect to its mapped port — exactly the connection path that doesn't work from this host's WSL2 side. It would work fine on a Linux GitHub Actions runner, but could not be verified locally on this machine without nesting Docker-in-Docker (mounting the socket into a `golang` container), which is more moving parts than the problem warrants.

## Decision

Integration tests that need Postgres read a DSN from `CD_TEST_DATABASE_URL` and skip (not fail) if it's unset. Locally, that DSN points at the existing `deploy/compose/dev.yml` Postgres (reached via the compose network from a `golang` container, or from WSL2 once/if its networking is reconfigured — not required). In CI, `ci.yml`'s Go job gains a `postgres` service container (a GitHub Actions `services:` entry, which runs as an actual sibling container reachable at `localhost` from the job's own containerized runner — no WSL involved there) and sets `CD_TEST_DATABASE_URL` accordingly.

This is simpler than testcontainers-go for this project's current scale (one database, no dynamic container lifecycle needed per test) and sidesteps the local networking issue entirely. Revisit testcontainers-go if/when tests need to spin up multiple distinct database instances per test case, or once development regularly happens inside a devcontainer (ADR 0009) where the networking issue wouldn't exist.

## Rejected alternatives

- `testcontainers-go` as originally planned — works in CI, but could not be locally verified without Docker-in-Docker nesting on this host; revisit later per above.
- Docker-in-Docker (mounting the host's Docker socket into a `golang` build container) purely to make `testcontainers-go` work locally — adds real complexity to solve a problem the simpler DSN-based approach avoids entirely.
