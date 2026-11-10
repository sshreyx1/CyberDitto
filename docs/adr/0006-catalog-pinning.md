# 0006: Attack catalog pinning and risk tiers

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v1 never actually wired up an attack catalog — the Emulation screen used a hardcoded array and a fake progress bar (MASTER_PLAN §2.2). v2 must execute real, ATT&CK-mapped tests, but a third-party catalog is untrusted input until verified, and some tests in any such catalog are destructive by nature (MASTER_PLAN §6.4).

## Decision

- Import **Atomic Red Team** (MIT) YAML **pinned to a specific commit SHA**, verified against a recorded hash manifest at import time; updating the pin is itself a reviewed change, never an automatic update.
- Every imported test gets a `risk_tier` (`benign` | `intrusive` | `destructive`). `destructive` tests are excluded from the schedulable set in code (a hard block in the domain layer), not a configuration toggle — MASTER_PLAN §6.4 invariant 3.
- The twin agent only ever executes tests referenced by this signed, pinned catalog or by a signed remediation plan produced by the control plane — there is no "run arbitrary command" RPC (§5.11).
- MITRE ATT&CK technique metadata (STIX) is imported separately, also pinned to a specific release.

## Rejected alternatives

- Executing catalog tests directly from a live upstream feed — rejected; an unpinned, unverified feed is a supply-chain and safety risk given this product's lab-execution capability.
