# 0008: Third-party licensing posture (CIS, Wazuh, Atomic, Caldera, Sigma)

- **Status:** accepted
- **Date:** 2026-11-10

## Context

Several components this project depends on or draws benchmark IDs from have licensing terms that matter for a publicly-hosted open-source project (MASTER_PLAN §6.7, §14 item 5): CIS Benchmarks are copyrighted by CIS; Wazuh's agent/manager components are GPL; Atomic Red Team is MIT; Caldera is Apache-2.0; Sigma rules have their own Detection Rule License terms.

## Decision

- Never bundle or redistribute CIS benchmark text. The `native` compliance engine refers to control IDs and uses original wording for check descriptions and remediation guidance.
- Never claim or imply CIS certification anywhere in the product or marketing.
- Wazuh runs as a separate process/container (its own GPL-licensed components are not linked into our binaries); the `wazuh-sca` adapter talks to it over its API.
- Atomic Red Team (MIT) and Caldera (Apache-2.0, deferred to stretch per §0.5.2) are consumed as data/external processes, not statically linked in a way that would taint licensing.
- A `THIRD_PARTY_NOTICES.md` is generated in CI (Phase 0/ongoing) rather than hand-maintained, to stay accurate as dependencies change.
- Sigma rule usage (detection-engineering stretch, §11 item 4) is deferred until its Detection Rule License terms are reviewed for this project's intended use.

## Rejected alternatives

- Bundling CIS benchmark PDFs/text for convenience — rejected outright on licensing grounds.
