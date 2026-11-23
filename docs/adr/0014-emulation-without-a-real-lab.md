# 0014: Build the emulation pipeline before the real lab exists; never simulate outcomes

- **Status:** accepted
- **Date:** 2026-11-23

## Context

Phase 4 executes Atomic Red Team tests inside a twin and classifies each as blocked / detected / logged / missed from telemetry (MASTER_PLAN §5.12). The real lab — Hyper-V provider, twin agent, lab gateway, Sysmon/Wazuh telemetry — is owner-blocked (`docs/OWNER_TODO.md` items 6–7). Most of Phase 4 doesn't depend on it: the pinned catalog, the destructive denylist, the preflight policy engine that enforces the §6.4 invariants, authorisation, the kill switch, and run records.

The `fake` provider could plausibly *predict* outcomes from the twin's configuration (e.g. "LSA protection off, so the LSASS dump was missed"). That would make demos richer, but MASTER_PLAN §0.4.3 forbids fabricated results, and a predicted outcome stored in the same column as an observed one is exactly that.

## Decision

- Build the whole pipeline now against the `LabProvider` port: `VerifyIsolation`, `Execute(twinID, test)`, `Revert`. The real provider later replaces only the adapter.
- The `fake` provider executes nothing. Every execution it reports has outcome **`not_executed`**, with a detail saying a real lab is needed. It never reports blocked, detected, logged or missed.
- Configuration-based reasoning about which techniques a weakness enables belongs in Phase 5 gap correlation, labelled as such (a control failing makes a technique *more likely* to succeed), not in run outcomes.
- Every §6.4 invariant that has code today has a dedicated test that fails the build if weakened (`internal/app/emulation/invariants_test.go`, `internal/domain/emulation`, `internal/adapters/atomic`).

## Consequences

- A run on the fake provider demonstrates the safety machinery end to end (preflight, authorisation, audit, kill switch, revert) but produces no attack-outcome evidence. The console says so on every such run.
- Gap priority (MASTER_PLAN §4.4) uses evidence factor 0.3 ("untested") for every technique until a real lab produces observed outcomes.

## Rejected alternatives

- Config-predicted outcomes stored as run results — rejected (§0.4.3, and it would make the tool's own evidence untrustworthy).
- Waiting for the real lab before building any of Phase 4 — rejected; the safety-critical parts are independent of the hypervisor and benefit from being built and tested first.
