# Threat model (v0)

> Living document. Update whenever a new container/capability is added. STRIDE per container, mapped to the system context in `docs/MASTER_PLAN.md §5.2`.

## Why this matters

CyberDitto holds configuration data about real machines and can execute attack tooling against lab VMs it controls. It is both a **high-value target** (the data it holds) and a **dangerous capability** (what it can do) if compromised or misused. v1 shipped with every API endpoint unauthenticated and a client-signed JWT — this document exists so that class of failure cannot recur silently.

## Containers and primary threats (STRIDE)

| Container | Spoofing | Tampering | Repudiation | Info disclosure | DoS | Elevation of privilege |
|---|---|---|---|---|---|---|
| Web app | Session fixation | — (server renders/validates) | — | XSS exfiltrating session cookie if CSP fails | Client-side only | — |
| Control-plane API | Forged session / stolen cookie | Unvalidated input mutating state | Missing/incomplete audit log | Workspace data leak via broken authZ (v1's exact failure) | Unbounded uploads/requests | Role-check bypass on a handler |
| Lab gateway (mTLS) | Rogue agent without a valid cert | Forged `runId`/`testId` markers | — | Telemetry/evidence leak off the lab network | Twin flooding the gateway | Agent executing an RPC outside its allow-list |
| Twin VM | — | Attack test modifies something outside the twin | — | Twin reaching LAN/internet and exfiltrating | Twin resource exhaustion affecting host | Attack test escapes the twin (hypervisor escape) |
| Postgres | — | Direct DB access bypassing RLS | Logs scrubbed/deleted | Backup left unencrypted | — | App connecting as superuser |
| Anthropic API calls | — | Prompt injection via tool output | — | Hostnames/users/IPs sent unredacted | Budget exhaustion | Model-proposed action executes without approval |

## Primary controls (see `docs/MASTER_PLAN.md §6` for full detail)

- Deny-by-default authZ on every route; RBAC + `approver` capability; authZ-matrix tests (every route × every role).
- Server-side sessions only, httpOnly/Secure/SameSite cookies, no tokens in `localStorage`.
- Lab network is internal-only (no NAT/bridge); twin→control-plane traffic only via the mTLS lab gateway; egress blocked at the hypervisor and in the twin's firewall; automated egress probe run before every attack run.
- Attack execution restricted to `TWIN` targets by the type system (no code path accepts an `ASSET`); `destructive` risk tier hard-blocked in code, not config.
- Catalog pinned to a verified commit SHA; twin agent only executes signed catalog entries or signed remediation plans — no arbitrary-exec RPC.
- Append-only, hash-chained audit log for every mutating action.
- LLM tool results treated as untrusted data; mutating tools always produce a human-gated `ApprovalRequest`; redaction/pseudonymisation before anything leaves the host; AI off by default.

## Known gaps at this stage (Phase 0)

No code exists yet beyond this document and the repo scaffold — the control plane, auth, lab gateway, and twin agent described above are design targets for Phases 1–4. This document will be updated as each lands, and the §6.4 safety invariants get dedicated automated tests (not just this document) before Phase 4 ships attack execution.
