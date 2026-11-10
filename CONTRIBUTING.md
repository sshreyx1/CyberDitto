# Contributing

This is a solo-maintained, all-rights-reserved portfolio project (see [`LICENSE`](LICENSE)) — the code is public for viewing and review, but no open-source license is granted. External pull requests are **not** currently being accepted, since there is no license under which to merge outside contributions.

If you notice a bug or security issue, see [`SECURITY.md`](SECURITY.md) for how to report it privately; that's welcome regardless of the license.

## Internal workflow (for the maintainer)

- Branches: `phase-N/topic` for roadmap work, otherwise `topic/short-description`.
- Commits: [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`, `ci:`, `sec:`), imperative subject ≤ 72 chars, body explains *why* not *what*.
- Before pushing: `task lint test` (once the Taskfile targets exist) must pass locally.
- No `--no-verify`, no force-push to `main`, no committing secrets or real scan data (see `.gitignore` and the `gitleaks` CI check).
- No AI-attribution trailers/footers in commits or PRs (enforced by a commit-msg hook and CI — see `docs/MASTER_PLAN.md §12`).
- Non-trivial decisions get a short ADR in `docs/adr/`.

## Code of conduct

See [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).
