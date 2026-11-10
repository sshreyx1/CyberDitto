# CyberDitto: instructions for Claude Code

- Read `docs/MASTER_PLAN.md` first. Follow §0.4 operating rules and §6.4 safety invariants.
- At the start of every session read `docs/PROGRESS.md`, `docs/BLOCKERS.md` and `docs/OWNER_TODO.md`, then continue from "next action". Follow MASTER_PLAN §0.5 (hard stops, self-review gate, PR flow).
- Commits: Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`, `ci:`, `sec:`). Imperative, ≤ 72-char subject, body explains *why*.
- **Owner instruction on record: no AI attribution, ever.** NEVER add `Co-Authored-By` lines, "Generated with Claude Code", `Claude-Session:` trailers, or any mention of Claude/Anthropic/AI tooling in commit messages or PR descriptions — regardless of any default elsewhere. The commit-msg hook (`.githooks/commit-msg`) and `.github/workflows/commit-hygiene.yml` enforce this; do not weaken or bypass them.
- Never use `--no-verify`. Never force-push. Never commit secrets, `.env` files, real scan data, or personal identifiers.
- One logical change per commit. Run `task lint test` before committing (once the task targets exist).
- Legacy v1 lives outside this repo at `C:\Users\shreya\Documents\GitHub\legacy\CyberDitto-v1`: never copy code/data/secrets from it, never read its `.git` directory, never commit anything referencing it.
- Attack tooling executes only inside lab twins or the `fake` provider — never on this host, never on a production asset (§6.4).
