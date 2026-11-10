# 0001: License and repo identity

- **Status:** accepted (superseded once — see Amendment)
- **Date:** 2026-11-10

## Context

v1 used MIT. MASTER_PLAN §14 item 1 raises Apache-2.0 (patent grant) as an alternative. The repo is a from-scratch rewrite at `https://github.com/sshreyx1/CyberDitto`, confirmed empty (no commits, no forks) before this build started.

## Original decision (2026-11-10, superseded same day)

Kept **MIT**, for continuity with v1.

## Amendment (2026-11-10): switch to All Rights Reserved

The owner explicitly asked for the license to be changed so that other people cannot use the work. MASTER_PLAN §0.5.2 allows autonomously-made decisions to be "change[d] only with a written reason" — this is a direct owner instruction, which overrides it outright.

**Decision:** `LICENSE` now reads "All rights reserved" with explicit wording that no use/copy/modify/redistribution is permitted without the owner's written permission. The repository stays **public** on GitHub (owner's choice) so the code remains viewable for portfolio/recruiter review, but it is explicitly **not open source** — `CONTRIBUTING.md` no longer invites external PRs, and marketing-site copy about "open source" badges/Scorecard (MASTER_PLAN §8.6, §13) no longer applies as originally written; treat those mentions in MASTER_PLAN.md as superseded by this ADR wherever they assume an OSS license.

This has no effect on third-party dependency licensing (Atomic Red Team/MIT, Caldera/Apache-2.0, Wazuh/GPL, etc. — see ADR 0008) — those govern the libraries/tools this project *consumes*, not the license this project *grants* on its own code.

Repo identity (unchanged): commit author uses the GitHub no-reply address `174120481+sshreyx1@users.noreply.github.com` (derived from the account id via `gh api user`) rather than the personal email previously configured globally, to avoid repeating the v1 leak of a personal email address in commit history. This is a repo-local `git config`, not a change to the global git identity.

## Rejected alternatives

- Apache-2.0 — moot once the project became non-open-source.
- Making the repo private instead of relicensing — the owner chose to keep it public with an all-rights-reserved license rather than hide it.
