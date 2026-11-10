# Security policy

CyberDitto handles sensitive configuration data about real machines and runs attack-emulation tooling against lab twins. Treat it as a high-value target. See `docs/threat-model.md` and `docs/MASTER_PLAN.md §6` for the full threat model and safety invariants.

## Reporting a vulnerability

Please report suspected vulnerabilities privately via GitHub's "Report a vulnerability" flow on this repository (Security tab → Advisories), rather than opening a public issue. Include:

- A description of the issue and its impact
- Steps to reproduce (or a proof of concept)
- Affected version/commit

We aim to acknowledge reports within 5 business days. This is a personal/portfolio project, not a company with an SLA — please be patient.

## Scope

In scope: the control-plane API, web app, collector, twin agent, and the lab-isolation/attack-execution invariants in `docs/MASTER_PLAN.md §6.4`.

Out of scope: third-party dependencies (report upstream), and anything requiring physical access to the operator's machine.

## Safe harbor

Good-faith security research against this project's own published code (not against any instance someone else is operating without your permission) is welcome and will not be treated as a hostile act, regardless of the project's all-rights-reserved license (see `LICENSE`) — that license governs reuse of the code, not security reporting.
