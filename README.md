# CyberDitto

> Attack your twin, not your production.

CyberDitto builds a configuration-faithful digital twin of a Windows endpoint, then safely assesses and attacks **the twin** (never the real machine) to answer three questions: how hardened is it, would real attacker techniques work and would you see them, and does a proposed fix actually close the gap.

```
Collect → Build Twin → Assess (CIS-style) → Attack (ATT&CK) → Correlate → Remediate on Twin → Re-test → Export fix for Prod → Report
```

This is a from-scratch rewrite ("v2") of an earlier prototype. The full plan, architecture, phase roadmap and safety model live in [`docs/MASTER_PLAN.md`](docs/MASTER_PLAN.md) — read that first for anything non-trivial.

## Status

Early build. Phase 0 (foundations and repo hygiene) is in progress. See [`docs/PROGRESS.md`](docs/PROGRESS.md) for the live checklist and [`docs/OWNER_TODO.md`](docs/OWNER_TODO.md) for items that need the project owner.

## Safety model

CyberDitto runs attack tooling. It only ever targets an isolated lab "twin" VM or a `fake` provider used in CI/demo mode — never the host it runs on, never a production asset. See [`docs/MASTER_PLAN.md §6.4`](docs/MASTER_PLAN.md#64-lab-safety-invariants-non-negotiable-each-has-a-test-that-fails-the-build-if-violated) for the non-negotiable invariants and [`SECURITY.md`](SECURITY.md) for the disclosure process.

**For authorised use only** — on systems you own or have explicit written permission to test.

## License

[MIT](LICENSE)
