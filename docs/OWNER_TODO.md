# Owner TODO

Things only the project owner can do. Each item lists the phase it unblocks.

## Unblocks: Phase 0 / ongoing hygiene

1. **Verify the legacy folder's `.git` is neutralised** (`C:\Users\shreya\Documents\GitHub\legacy\CyberDitto-v1`). Per MASTER_PLAN §0.1.2, run inside that folder (not from this repo, and not by Claude — this is outside the project directory):
   ```
   git remote remove origin
   ```
   and ideally delete that folder's `.git` directory entirely, plus any `.vagrant` folders under `legacy\...\cyberditto-backend\vagrant\`. This repo's automation never reads or touches that path.
2. **Rotate credentials** that were hardcoded in the old `config.ts` (two accounts) and any JWT secret, if reused anywhere else. Assume they were exposed while the old repo was public.
3. **GitHub privacy settings**: Settings → Emails → enable "Keep my email addresses private" and "Block command line pushes that expose my email". Your GitHub no-reply address is `174120481+sshreyx1@users.noreply.github.com` (derived from your account id via `gh api user`) — this repo's local git config now uses it (see commit below); enabling the two settings above makes GitHub itself enforce it too.
4. **Branch protection on `main`**: once CI is green (end of Phase 0), add a branch protection rule requiring the `ci`, `security`, and `commit-hygiene` checks before merge, and disallow force-pushes. This needs your GitHub account/admin access — not something the automation can do with the current `gh` token scopes (`repo`, `read:org`, `project`, `gist` — no `admin:org`/repo-admin branch-protection scope confirmed yet).
5. **Hyper-V**: confirm Hyper-V is enabled on this Windows 11 host (`Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V` → `Enabled`) ahead of Phase 2's real `LabProvider` work. Not required to start Phase 0/1, which build against the `fake` provider.

## Unblocks: Phase 2 (real twin)

6. Decide/obtain a Windows 11 evaluation image or a user-supplied VHDX (licensing constraints — see MASTER_PLAN §6.7, §14 item 3). Record the chosen source and its licence/expiry once decided.
7. Decide hardware layout for Wazuh: same host (trimmed) vs. a second box vs. `native`-only lite mode for the MVP (MASTER_PLAN §5.9, §14 item 2). Default assumption until you say otherwise: **lite mode (`native` engine only)** for the MVP, Wazuh as an opt-in compose profile later.

## Unblocks: Phase 6 (AI features)

8. Supply `ANTHROPIC_API_KEY` in your own shell/secret manager when you want to test the Copilot. The app is fully functional with `CD_AI_ENABLED=false` (the default) until then.

## Unblocks: Phase 7 (website)

9. Pick and pay for a domain + hosting (Vercel/Cloudflare Pages) if you want the marketing site live publicly; otherwise it stays buildable/deployable locally.
