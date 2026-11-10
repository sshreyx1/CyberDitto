# 0011: Bring Caldera into the main v2.0 build (not deferred)

- **Status:** accepted
- **Date:** 2026-11-10

## Context

MASTER_PLAN §0.5.2 and §11 originally deferred MITRE Caldera to a stretch feature, outside v2.0 scope, because of a past max-severity RCE (CVE-2025-27364, fixed in 5.1.0+) and the extra isolation work a second attack-execution engine needs on top of the primary Atomic Red Team pipeline (Phase 4). The owner explicitly asked for it to be included in the working build rather than deferred.

## Decision

Caldera is now **in scope for Phase 4**, as a second `AttackCatalog`/execution path alongside Atomic Red Team, not a later stretch item. This does not relax any §6.4 safety invariant — the existing invariant 9 already required: Caldera runs in its own container/VM on the lab network only, pinned to a release that fixes CVE-2025-27364 (i.e. **5.1.0 or later**, verified against the actual latest patched release at Phase 4 build time per §0.3), never exposed beyond the lab, and never given a route to the host or production assets. The Sandcat agent and `emu` plugin (for CTID plans) are evaluated for inclusion at that time, same provenance/pinning rules as the Atomic catalog (ADR 0006): pinned version, verified hash, no auto-update.

Phase 4's acceptance criteria and task list (MASTER_PLAN §10) gain a Caldera sub-item; Phase 4 is not considered complete until both engines pass the same safety-invariant test suite (destructive-tier block, twin-only targeting, lab-isolation preflight, kill switch).

## Rejected alternatives

- Leaving Caldera deferred — rejected per explicit owner instruction, 2026-11-10.
- Replacing Atomic Red Team with Caldera instead of running both — rejected; Atomic Red Team remains the simpler, lower-risk primary engine (MIT license, no history of a critical RCE), Caldera is additive.
