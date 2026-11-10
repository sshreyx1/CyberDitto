# 0001: License (MIT) and repo identity

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v1 used MIT. MASTER_PLAN §14 item 1 raises Apache-2.0 (patent grant) as an alternative. The repo is a from-scratch rewrite at `https://github.com/sshreyx1/CyberDitto`, confirmed empty (no commits, no forks) before this build started.

## Decision

Keep **MIT**, for continuity with v1 and because the patent-grant benefit of Apache-2.0 matters most for projects expecting significant third-party contribution/patent exposure, which does not describe this project's current stage. Can be revisited pre-v1.0 release if that changes.

Repo identity: commit author uses the GitHub no-reply address `174120481+sshreyx1@users.noreply.github.com` (derived from the account id via `gh api user`) rather than the personal email previously configured globally, to avoid repeating the v1 leak of a personal email address in commit history. This is a repo-local `git config`, not a change to the global git identity.

## Rejected alternatives

- Apache-2.0 — deferred, not rejected outright; revisit later if contribution patterns change.
