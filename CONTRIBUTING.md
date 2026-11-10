# Contributing

This is currently a solo-maintained portfolio project built in the open. External PRs are welcome once the walking skeleton (Phase 1) lands; until then, expect rapid structural churn.

## Workflow

- Branches: `phase-N/topic` for roadmap work, otherwise `topic/short-description`.
- Commits: [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`, `ci:`, `sec:`), imperative subject ≤ 72 chars, body explains *why* not *what*.
- Before pushing: `task lint test` (once the Taskfile targets exist) must pass locally.
- No `--no-verify`, no force-push to `main`, no committing secrets or real scan data (see `.gitignore` and the `gitleaks` CI check).
- Non-trivial decisions get a short ADR in `docs/adr/`.

## Commit attribution

Commits and PRs in this repository must not contain AI-attribution trailers or footers (`Co-Authored-By: ... Claude ...`, "Generated with Claude Code", `Claude-Session:`, etc.) — this is enforced by a commit-msg hook and CI (see `docs/MASTER_PLAN.md §12`).

## Code of conduct

See [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).
