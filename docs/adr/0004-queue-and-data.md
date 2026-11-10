# 0004: Jobs/workflow queue and primary datastore

- **Status:** accepted
- **Date:** 2026-11-10

## Context

Twin builds, assessments and emulation runs are long-running, multi-step, resumable workflows that need durable state and progress events (MASTER_PLAN §5.1, §5.7). v1 kept everything in memory, so scans/deployments were lost on restart (§2.3 #6).

## Decision

- **PostgreSQL** as the single datastore for both application state and the job queue — avoids running Redis/Kafka for an app at this scale.
- **River** (Postgres-backed queue) for background jobs, paired with explicit state machines in the `domain` layer for twin and run lifecycles (§5.7).
- `pgx` driver, **`sqlc`** for typed queries (no ORM), **`goose`** for migrations.
- Realtime progress to the browser over **SSE** (not WebSockets) — simpler, proxy-friendly, and sufficient since the client never needs to push back mid-stream.

## Rejected alternatives

- Redis/Celery-style queue — adds an operational dependency with no measured need yet; revisit only if River's throughput becomes a bottleneck.
- Temporal — too much operational weight for the current workflow complexity; revisit only if workflows grow significantly more branchy.
- WebSockets for run/build streams — no bidirectional requirement exists yet.
