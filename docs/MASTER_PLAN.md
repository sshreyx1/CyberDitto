# CyberDitto v2: Master Build Plan

> **Tagline:** *Attack your twin, not your production.*
> **Audience of this document:** Claude Code (the builder) and the project owner (the human).
> **Date written:** 10 Nov 2026. Version numbers below were current at that date; **verify every pinned version against official docs/registries before you lock it** (see §0.3).

---

## 0. How to use this document

### 0.1 For the human owner (do these BEFORE starting Claude Code)

**Known facts about your setup:** legacy copy at `C:\Users\shreya\Documents\GitHub\legacy\CyberDitto-v1`; new, empty repo at `https://github.com/sshreyx1/CyberDitto`; new project folder e.g. `C:\Users\shreya\Documents\GitHub\CyberDitto`.

1. **Verify the new GitHub repo really has no old history.** The new repo uses the *same name* as the old one. On GitHub open `Commits`: it must show at most one fresh initial commit (README/licence). If it shows the old commits (e.g. one starting `90964ee`), the repo was only emptied, not recreated, and the old secrets are still public. In that case delete the repo (Settings → Danger Zone) and create it again. Also check the repo has **no forks** (`Insights → Forks`).
2. **Neutralise the legacy folder's `.git`.** The legacy copy still carries the *entire old history* (hardcoded passwords, MACs/IPs, your email) and still has `origin` pointing at `sshreyx1/CyberDitto`. Deleting files from the working tree does not remove them from history. Do one of:
   - `cd` into the legacy folder and run `git remote remove origin` **and** delete the folder's `.git` directory (best; the working files are all Claude Code needs), or
   - at minimum `git remote remove origin` so nothing can ever be pushed to the new repo by accident.
   Also delete the `.vagrant` folders inside `legacy\...\cyberditto-backend\vagrant\` (they contain your Windows username/OneDrive path).
3. **Rotate credentials.** Any password you used in the old `config.ts` (two accounts) and any JWT secret: change them everywhere you reused them. Assume they were scraped while the repo was public.
4. **GitHub privacy:** Settings → Emails → enable *Keep my email addresses private* and *Block command line pushes that expose my email*. Use the no-reply address in git config (§12.4).
5. **Set up commit attribution** exactly as in **§12** before the first commit (no second GitHub account needed).
6. **Where to run Claude Code (important for autonomous mode, see §0.5):** inside **WSL2 or a devcontainer** (or a throwaway VM), with the new repo folder as the only mounted project. Do **not** give an autonomous agent your whole user profile. Keep `ANTHROPIC_API_KEY` for the *product's* AI features in your shell only when that phase needs it.
7. Virtualization: Hyper-V enabled on the Windows host for the real-lab phases (see §5.9 for sizing). Autonomous work does not wait for this: it builds against the `fake` provider first and queues real-lab verification in `docs/OWNER_TODO.md`.

### 0.2 Paste this as your first prompt to Claude Code

```text
You have full autonomy over the CyberDitto v2 build. Working directory is the new, empty repo (origin: https://github.com/sshreyx1/CyberDitto).
1. Read docs/MASTER_PLAN.md completely (put it there first), especially §0.4, §0.5 (autonomous mode), §6.4 (safety invariants) and §12 (git rules).
2. The v1 code is at C:\Users\shreya\Documents\GitHub\legacy\CyberDitto-v1 (WSL: /mnt/c/Users/shreya/Documents/GitHub/legacy/CyberDitto-v1). Read it for ideas only. Never copy code, data, or secrets from it, and never read its .git directory.
3. Initialise state files per §0.5, then execute Phase 0 through Phase 8 in order, continuing on your own without asking for approval between phases. Stop only at the hard stops listed in §0.5.3.
4. Commit and push to main (Phase 0) / phase branches with PRs self-merged after green CI (§12.5, §0.5.4). Keep docs/PROGRESS.md current after every task so you can resume after a context reset.
```

### 0.3 Version policy (important)

Everything marked `(verify)` is a decision made from research on the date above. Claude Code must, at build time: (a) check the latest stable release on the official site/registry, (b) pin an exact version (never `latest`), (c) record the choice and date in `docs/adr/`, (d) choose versions at least 14 days old unless a security fix demands otherwise.

### 0.4 Operating rules for Claude Code (also copy into `CLAUDE.md`, see §12)

1. **Phase discipline.** Work one phase at a time (§10). Do not start phase N+1 until phase N's acceptance criteria pass **and your self-review gate (§0.5.5) is green**.
2. **Plan, then build.** For each phase write a short plan (files, decisions, risks) into `docs/phases/phase-N.md`, then implement it. In autonomous mode you approve your own plan only if it satisfies §0.5.5; otherwise stop at a hard stop.
3. **No fabricated results.** Never hardcode sample data into production code paths. Seeded demo data lives only in `fixtures/` and is only loaded in demo mode. Never claim a check passed unless code actually evaluated it. (The v1 collector hardcoded `firewall=true` and `uac=true`; that must never happen again.)
4. **Safety invariants are law** (§6.4). Attack tooling executes only inside lab twins, never on the host that runs CyberDitto, never on production assets.
5. **Secrets.** Never commit secrets, real scan output, or personal identifiers. Use env vars + `.env.example` (placeholders only). Run `gitleaks` before every commit.
6. **No shell injection.** Never build shell strings from user input. Use `exec.Command(name, args...)` with validated arguments and allow-listed executables.
7. **Tests with code.** Every feature ships with unit tests; every endpoint ships with an integration test; every user flow ships with a Playwright test (§9.4).
8. **ADRs.** Any non-trivial decision gets a short Architecture Decision Record in `docs/adr/NNNN-title.md`.
9. **Hard stops, not guesses,** for licensing, destructive operations, secrets, or anything touching the host OS outside the repo directory (§0.5.3).
10. **Report honestly.** End each phase with: what works, what is stubbed, what you could not verify, written to `docs/PROGRESS.md` and the PR description.

### 0.5 Autonomous mode (owner has granted full control)

**Mandate:** build CyberDitto v2 end to end without waiting for per-phase approval. Full control over architecture details, code, tests, docs, CI, repo structure, UI design, commits and PRs **within the guardrails below**. The safety invariants in §6.4 and the git rules in §12 are not negotiable and are not "delegated" decisions.

#### 0.5.1 State that survives context resets (create in Phase 0, update constantly)
| File | Purpose |
|---|---|
| `docs/PROGRESS.md` | Master checklist copied from §10 with status, dates, PR links, "next action". **Read this first at the start of every session.** |
| `docs/DECISIONS.md` + `docs/adr/` | Decisions taken autonomously, with rationale and rejected alternatives |
| `docs/OWNER_TODO.md` | Things only the owner can do (real-lab verification, GitHub settings, secrets, purchases). Each with exact steps and the phase it unblocks |
| `docs/BLOCKERS.md` | Anything stuck: what was tried, error output, suggested next steps |
| `CLAUDE.md` | Kept short; points to the files above |

#### 0.5.2 Decisions already made for you (do not ask; record in ADRs; change only with a written reason)
| Topic | Decision |
|---|---|
| Licence | ~~MIT (continuity with v1)~~ — **superseded 2026-11-10: All Rights Reserved, owner instruction.** Repo stays public; code is not open source. See `docs/adr/0001-license-and-repo-identity.md`. |
| Repo | `https://github.com/sshreyx1/CyberDitto`, branch `main` protected after Phase 0; default branch stays `main` |
| Host dev OS | Windows 11 host; develop in WSL2/devcontainer; real-lab provider = Hyper-V |
| Phase order | §10 as written. Phases 0 to 5 are the product; 6 to 8 make it shine |
| Compliance engine | `native` first (Phase 3), `wazuh-sca` adapter second behind a compose profile; lite mode (native only) is a first-class supported mode |
| Caldera | ~~Deferred to stretch (§11). Not in v2.0~~ — **superseded 2026-11-10, owner instruction: in scope for Phase 4**, alongside Atomic Red Team, not deferred. See `docs/adr/0011-caldera-in-scope.md`. |
| Docs framework | Fumadocs unless a blocking problem appears (ADR) |
| Analytics | Plausible or Umami, self-hostable, no cookies |
| Website hosting | Static + edge friendly (Vercel/Cloudflare Pages); config for both, deploy only when the owner supplies credentials (OWNER_TODO) |
| AI features | Off by default; fully functional app without an API key |

#### 0.5.3 Hard stops (stop, write to `docs/BLOCKERS.md` or `docs/OWNER_TODO.md`, tell the owner, and continue with unrelated work if any exists)
1. A secret, token, API key, or paid account is needed and is not in the environment.
2. Any action outside the project directory that modifies the host (installing system packages, changing Windows/Hyper-V settings, editing the registry, touching other repos or user folders). Inside a devcontainer/WSL sandbox, installing dev dependencies is fine.
3. **Any attack/emulation execution anywhere except a verified lab twin or the `fake` provider.** Never run Atomic tests, collectors-with-apply, or remediation scripts on the host.
4. Anything that would weaken a §6.4 invariant, disable a security check, skip hooks (`--no-verify`), force-push, rewrite published history, or change branch protection to bypass checks.
5. Licensing/legal doubts (bundling CIS text, redistributing Windows images, third-party assets with unclear licences).
6. Real GitHub/hosting/DNS/billing changes that need the owner's account (prepare the config and instructions instead).
7. The same task fails **three** attempts with different approaches: document in `BLOCKERS.md`, move to another task.
8. Spending: do not run paid external services beyond the AI budget in `.env` and never run loops that can burn tokens unboundedly.

#### 0.5.4 Git and PR flow without a human reviewer
- Phase 0: commit directly to `main` until CI and protection exist; after that **never push to `main` directly**.
- Work on `phase-N/<topic>` branches; open a PR using `gh` if authenticated (else leave the branch pushed and note it in OWNER_TODO); **self-merge only when all required checks are green and the PR description contains the acceptance checklist with evidence** (test output, screenshots).
- Squash-merge with a Conventional-Commit title. Obey §12 attribution rules on every commit and PR body.
- Tag `v0.N.0` after each completed phase; tag `v2.0.0` after Phase 8.

#### 0.5.5 Self-review gate (replaces owner approval at the end of each phase)
A phase is complete only when **all** are true; record the evidence in `docs/PROGRESS.md`:
1. Acceptance criteria in §10 for the phase are met, or the unmet ones are explicitly moved to `OWNER_TODO.md` because they need real hardware/accounts.
2. `task lint test` and CI are green, including `gitleaks`, `govulncheck`, dependency audit, the architecture (`depguard`) check and the commit-hygiene check.
3. The §6.4 invariant tests pass and **no new code path** can execute attacks on a non-twin target.
4. No hardcoded demo data in production paths (§0.4.3); no TODO/FIXME left without an issue entry.
5. Docs updated (architecture, ADRs, runbooks, README sections), screenshots/GIFs refreshed for UI phases.
6. An adversarial self-review pass: re-read your own diff as a security reviewer, list 5 ways it could be abused or fail, and fix or log each.
7. A short "what works / what's stubbed / what I couldn't verify" note in the PR.

#### 0.5.6 Work around missing hardware instead of waiting
Phases 2 and 4 need a real Windows lab. Build and fully test everything against the `fake` provider and recorded fixtures first (synthetic, never from a real machine), write the real-lab adapter and its runbook, then add the real-lab acceptance steps to `OWNER_TODO.md` as a numbered checklist the owner can run later. Continue to the next phase meanwhile. The final report must state clearly which acceptance criteria were verified on real hardware and which were not.

#### 0.5.7 How the owner should run it
- Run Claude Code in the sandbox from §0.1.6. If you enable a permissive mode (e.g. skipping permission prompts) do it **only inside the container/VM** with no personal files, no cloud credentials, and no SSH keys beyond a repo-scoped deploy credential or a `gh` login restricted to this repo.
- Keep the deny rules in §12.1. Add a budget cap in your Anthropic console.
- Check `docs/PROGRESS.md` and `docs/OWNER_TODO.md` whenever you like; to resume after a stop, start a new session and say: *"Read CLAUDE.md and docs/PROGRESS.md and continue."*

---

## 1. Vision and executive summary

**CyberDitto** builds a *configuration-faithful digital twin* of a Windows endpoint (later: small networks), then safely attacks and audits **the twin** to answer three questions about the real machine:

1. **How hardened is it?** (CIS-style configuration assessment, scored)
2. **Would real attacker techniques work, and would we see them?** (MITRE ATT&CK-mapped emulation with *blocked / detected / logged-only / missed* outcomes)
3. **What exactly should I change, and does the change actually work?** (evidence-linked remediation, **re-tested on the twin** before anything touches production)

The v1 project stated these goals but only ever implemented "scan the host, then boot a stock Windows VM". v2 delivers the whole loop:

```
Collect → Build Twin → Assess (CIS) → Attack (ATT&CK) → Correlate → Remediate on Twin → Re-test → Export fix for Prod → Report
```

### 1.1 Differentiators (what makes this portfolio-grade)

| # | Differentiator | Why it matters |
|---|---|---|
| 1 | **Closed-loop validation**: apply a fix on the twin, re-run the *same* attack, show the before/after delta | Most tools list findings. Few prove a fix works. |
| 2 | **Evidence-weighted risk**: a failed control with a *proven working attack* outranks one with no test | Prioritisation that is defensible and explainable |
| 3 | **Twin Fidelity Score** with an honest "known gaps" panel | Credibility. Shows you understand the limits of the approach. |
| 4 | **Ditto Copilot** (Claude-powered): grounded in your own run evidence, tool-using, with human approval on every mutating action | Useful AI, not a gimmick |
| 5 | **MCP server** (`cyberditto-mcp`): use CyberDitto from Claude Desktop/Code | Modern, demonstrates protocol fluency |
| 6 | **Safety-by-design lab**: isolated network, snapshot/revert, signed pinned attack catalog, destructive tests blocked | Security tool that is itself secure |
| 7 | **Public demo mode** that runs with zero backend | Recruiters can click through in 10 seconds |

### 1.2 Non-goals (v2)

- Not a pentest framework or C2. No exploitation of third-party systems. No real malware.
- Not auto-pushing changes to production in v2 (it **exports** reviewed remediation as PowerShell/Ansible/GPO/Intune artifacts instead).
- Not a SIEM replacement. It validates detections; it does not run a SOC.
- Not multi-tenant SaaS in v2 (but data model is workspace-scoped from day one so it can become one).

---

## 2. Audit of the v1 project (what is actually in the zip)

### 2.1 What exists

| Area | What I found |
|---|---|
| Backend | Go 1.23, `gorilla/mux`, `rs/cors`. Endpoints: start scan, scan status/result, deploy (Vagrant VM), deploy status/start/stop/delete/cancel, list deployments. |
| Scan | `scripts/config_collector.ps1` collects OS version, CPU, RAM, C: size, NICs, listening ports, DNS, running services, firewall *rule count*, last 5 hotfixes, and writes JSON. |
| Twin | `pkg/vagrant/manager.go` shells out to `vagrant`/`VBoxManage`, creates a dir per VM, drops the scan JSON onto `C:\` of a stock **`StefanScherer/windows_11` box pinned to a 2021 version**, 4 GB RAM, GUI on. |
| Frontend | React 18 + Vite + TS, react-router, recharts, lucide. Pages: Login, Dashboard, Digital Twin (create/view), Benchmark (run/view), Emulation (run/view). |

### 2.2 What was claimed but is not implemented

The README describes MySQL auth, Nmap, Ansible, GNS3, NetReplica, CIS benchmark automation, Atomic Red Team/Caldera attack runs, PDF reports, gap analysis and one-click remediation. **None of these exist in the code.** The Benchmark and Emulation screens use hardcoded arrays and a `setInterval` fake progress bar. The Dashboard charts are hardcoded sample data (monthly "traffic", fake APT scores).

### 2.3 Defects and security findings (all must be designed out in v2)

| # | Finding | Severity | v2 answer |
|---|---|---|---|
| 1 | Hardcoded fallback credentials and JWT secret in frontend config; JWT is **signed in the browser** with a secret bundled to every visitor; token in `localStorage` | Critical | Server-side auth, httpOnly cookie sessions, passkeys/TOTP (§6.2) |
| 2 | Backend `Auth` middleware is a no-op, so **every API endpoint is unauthenticated** | Critical | AuthN/Z on every route, deny-by-default (§6.2) |
| 3 | Collector hardcodes `firewall.enabled = $true`, `uac.enabled = $true`, and labels any Defender output as "Windows Defender" | High (integrity) | Real checks with evidence; unit tests with golden data |
| 4 | Real device identifiers (MACs, IPs, services, KBs) committed in `security_posture.json` | High (privacy) | Pseudonymisation, `.gitignore`, secret/PII scanning in CI |
| 5 | "Twin" does not mirror the host. It is a stock 2021 image plus a JSON file | High (product) | Real config replication + fidelity score (§5.7) |
| 6 | State only in memory (scans, deployments lost on restart); resource usage values are hardcoded "examples" | Medium | PostgreSQL + real hypervisor metrics |
| 7 | VM names/paths built from IDs by string formatting and passed to `vagrant`/`VBoxManage` | Medium | Validated IDs, allow-listed executables, arg arrays |
| 8 | No tests, no CI, no migrations, no docs, `node_modules`/`.vagrant` state present in archive (check whether tracked) | Medium | Full DevEx baseline (§9) |
| 9 | Vagrant + Packer are now under HashiCorp's source-available BSL licence; the pinned box is stale | Medium | Provider abstraction; Packer-built images optional (§5.9) |
| 10 | Duplicate `ProtectedRoute` (empty file + inline copy), mixed casing in API payloads (`ID` vs `id`) | Low | Generated typed client from OpenAPI (§5.6) |

### 2.4 What is worth reusing

- **Product shape and page map** (Dashboard / Digital Twin / Benchmark / Emulation) as wireframe input only.
- **Data concepts** from `models/types.go` (SystemInfo, NetworkInfo, SecurityInfo, deployment status lifecycle). They become a richer `Snapshot` model.
- **Lifecycle ideas** (preparing → initializing → deploying → running → stopped) as the seed for the Twin state machine.
- Everything else is rewritten. Do not port code; port ideas.

---

## 3. Why v2 is far more capable now (research summary)

| Topic | Then (v1) | Now (verified Nov 2026) | Impact |
|---|---|---|---|
| **Attack library** | Planned "Atomic Red Team" but never wired | Atomic Red Team (MIT, maintained by Red Canary) with the `Invoke-AtomicRedTeam` PowerShell executor; a maintainer-built **MCP server** exposes 1,500+ atomics for search/validate/generate/execute (execution disabled by default, opt-in via env var) | Real ATT&CK-mapped tests, AI-assisted test authoring |
| **CIS benchmarking** | Hand-rolled PowerShell, none implemented | **Wazuh SCA** ships open CIS-style policies for Windows 10, 11, and Server 2012 to 2025, with scoring in the dashboard | Hundreds of checks instead of dozens; continuous drift detection |
| **Adversary emulation (advanced)** | n/a | MITRE Caldera v5 (Apache-2.0) with Sandcat agent and `emu` plugin for CTID plans. **Had a max-severity RCE (CVE-2025-27364), fixed in 5.1.0+** | Optional advanced mode, but must be isolated and patched |
| **VM tooling** | Vagrant + VirtualBox + 2021 box | HashiCorp moved Vagrant/Packer/Terraform to BSL in 2023; OpenTofu/OpenBao are the OSS forks. Hyper-V (Windows), KVM/libvirt (Linux), Proxmox API are first-class | Provider-abstraction, Packer-built or Microsoft-evaluation images |
| **Backend language** | Go 1.23 | **Go 1.26** (Feb 2026; Green Tea GC default; 1.26.x security patch releases ongoing) | Latest stdlib routing, `log/slog`, `govulncheck` |
| **Frontend** | React 18 + Vite SPA | **Next.js 16 (Turbopack stable), React 19.2, Tailwind v4, shadcn/ui v4 CLI, TanStack Query** | Marketing site + app in one codebase, far better UI velocity |
| **LLM integration** | None | Claude Messages API with **structured outputs** (schema-guaranteed JSON), **strict tool use**, **prompt caching**, **citations**, **MCP connector**, **Agent SDK** | Copilot that returns validated plans, not prose guesses |
| **BAS market patterns** | n/a | Commercial BAS (Picus, AttackIQ, SafeBreach, Cymulate) converged on: ATT&CK heatmaps, prevention vs detection coverage, severity per technique, scheduled continuous runs, reports | UX patterns to borrow (without copying branding) |

**Honest limits** (state these in the product too): a twin mirrors *configuration*, not live data or user behaviour; Windows images need valid licensing; running Windows twins plus a Wazuh stack needs real RAM (§5.9).

---

## 4. Product definition

### 4.1 Personas

| Persona | Goal | Key screens |
|---|---|---|
| **Sam, security student/engineer** (primary, also the portfolio viewer) | Understand and demonstrate hardening + attack validation end to end | Twin builder, Assessment, Emulation, Copilot |
| **Priya, IT admin at a small business** | Know what to fix first and get safe scripts | Findings, Remediation, Report |
| **Dev, detection engineer** | Verify detections fire for specific techniques | ATT&CK heatmap, Telemetry, Sigma suggestions |
| **Recruiter/Reviewer** | Evaluate quality in minutes | Marketing site, public demo, README, architecture docs |

### 4.2 Core user journey

1. **Sign in** (first-run wizard creates the owner account; passkey/TOTP offered).
2. **Add an asset**: download the signed **Collector**, run it on a Windows machine (read-only, consent prompt, manifest of what it gathers), upload or auto-send the snapshot.
3. **Build a twin** from the snapshot. Watch live progress; get a **Fidelity Score**.
4. **Assess**: run the compliance engine against the twin. See score, failed controls with evidence.
5. **Emulate**: choose a scenario (preset or ATT&CK technique list or "paste a threat report"). Policy engine pre-flight checks safety. Run on the twin. Watch the live console.
6. **Review outcomes**: ATT&CK heatmap with Blocked / Detected / Logged / Missed, with telemetry evidence per technique.
7. **Correlate**: Gap view links *failed control ↔ successful technique* with an evidence-weighted priority.
8. **Remediate on the twin**: generate a plan (human-readable + scripts + rollback), approve, apply to twin, **re-run the same attack**, see the delta.
9. **Export for production**: signed script bundle (PowerShell/Ansible/GPO backup/Intune), plus PDF/JSON/SARIF/ATT&CK-Navigator/CSV reports.
10. **Ask Ditto** anywhere ("why did T1003.001 succeed?", "what breaks if I enable this?").

### 4.3 Twin fidelity levels

| Level | Meaning | v2 scope |
|---|---|---|
| L0 Inventory | OS, patches, ports, services names (what v1 did) | Baseline |
| **L1 Configuration** | Security policy, audit policy, registry policy keys, services (config + state), firewall profiles/rules, Defender config, local users/groups (no hashes), features, SMB/TLS/RDP/WinRM/PowerShell-logging settings, installed software versions | **MVP** |
| L2 Topology | Multiple hosts, subnets, AD domain | Stretch (§11) |
| L3 Data/behaviour | Synthetic data, user activity | Stretch |

**Fidelity Score** = weighted share of L1 settings that, after the twin is built, are re-collected from the twin and found equal to the source snapshot. Unmatched items are listed with the reason (e.g. "hardware-specific", "needs TPM", "domain-managed").

### 4.4 Scoring (transparent and documented in-app)

- **Posture score (0–100):** `100 × (1 − Σ wᵢ·failᵢ / Σ wᵢ)`; weights: High=5, Medium=3, Low=1 (configurable). Not-applicable and manual checks excluded and counted separately.
- **Resilience score (0–100):** over executed techniques: `Blocked=1.0, Detected=0.6, Logged=0.25, Missed=0`, averaged, optionally weighted by technique prevalence.
- **Gap priority:** `severity(control) × evidence(1 if technique succeeded on twin, else 0.3) × prevalence(technique) × exposure(asset criticality)`. Always show the factor breakdown (explainability).
- **Outcome taxonomy per executed test:** `blocked` (prevented), `detected` (ran but alert raised), `logged` (telemetry exists, no alert), `missed` (no useful telemetry), plus `error/skipped`.

---

## 5. Architecture

### 5.1 Principles

1. **Modular monolith first.** One Go control-plane binary (API + workers) with strict internal module boundaries; extract services only if there is a proven need. Clean/hexagonal architecture: domain has zero infrastructure imports.
2. **Ports and adapters.** Hypervisors, compliance engines, telemetry sources, LLM, object storage are all *ports* with swappable *adapters*. A `Fake` adapter for each makes CI and demo mode possible without a hypervisor.
3. **API-first.** `api/openapi.yaml` is the source of truth; server stubs and the TypeScript client are generated.
4. **Event-sourced where it matters.** Runs and twin builds emit append-only events (progress, evidence, audit). UI streams them via SSE.
5. **Secure by default, safe by design** (§6). Deny-by-default authZ, isolated lab network, signed catalogs.
6. **Everything explainable.** Every score, gap and AI statement links to evidence rows.
7. **Boring tech where possible**: Postgres for state *and* queue; no Redis/Kafka unless measured need.

### 5.2 System context

```mermaid
flowchart LR
  user([Analyst / Admin])
  subgraph Host["CyberDitto host"]
    web["Web app (Next.js 16)"]
    api["Control plane API + workers (Go)"]
    db[("PostgreSQL")]
    obj[("Object storage: S3/MinIO or local")]
    siem["Telemetry/SCA backend (Wazuh)"]
    hv["Hypervisor provider (Hyper-V / KVM / Proxmox)"]
  end
  subgraph Lab["Isolated lab network (no LAN, no internet by default)"]
    twin["Windows twin VM"]
    agent["Twin agent (Go)"]
    wagent["Wazuh agent + Sysmon"]
  end
  prod["Real asset (read-only Collector)"]
  claude["Anthropic API (Claude)"]

  user --> web --> api
  api --> db
  api --> obj
  api --> hv --> twin
  agent -- "mTLS, outbound only" --> api
  twin --- agent
  twin --- wagent --> siem
  api --> siem
  prod -- "signed snapshot upload" --> api
  api -- "redacted context only" --> claude
```

### 5.3 Container / module view (control plane)

```mermaid
flowchart TB
  subgraph transport["transport (adapters in)"]
    http["HTTP/JSON + SSE (generated from OpenAPI)"]
    mcp["MCP server"]
    labgw["Lab gateway listener (mTLS, lab NIC only)"]
  end
  subgraph app["application (use cases)"]
    assets["assets"]
    snaps["snapshots"]
    twins["twins"]
    assess["assessments"]
    emu["emulation"]
    gaps["correlation"]
    rem["remediation"]
    rep["reports"]
    copilot["copilot"]
    iam["iam + audit"]
  end
  subgraph domain["domain (pure Go, no I/O)"]
    ent["entities, value objects, state machines, scoring, policy engine"]
  end
  subgraph ports["ports (interfaces)"]
    p1["LabProvider"]
    p2["ComplianceEngine"]
    p3["TelemetrySource"]
    p4["AttackCatalog"]
    p5["LLM"]
    p6["Repository / Queue / BlobStore"]
  end
  subgraph adapters["adapters (out)"]
    a1["hyperv, libvirt, proxmox, fake"]
    a2["native, wazuh-sca"]
    a3["wazuh-indexer, fake"]
    a4["atomic-yaml (pinned), caldera"]
    a5["anthropic, fake"]
    a6["postgres+sqlc, river, s3, localfs"]
  end
  transport --> app --> domain
  app --> ports
  adapters --> ports
```

**Dependency rule (enforce with `depguard`/`go-arch-lint` in CI):** `domain` imports nothing internal; `app` imports `domain` + `ports`; `adapters` import `ports` + `domain`; `transport` imports `app`. Nothing imports `adapters` except `cmd/*` wiring.

### 5.4 Technology decisions

| Concern | Decision | Notes / ADR topic |
|---|---|---|
| Backend language | **Go 1.26** (verify patch level) | `log/slog`, stdlib `ServeMux` with method+path patterns; add `chi` only if middleware ergonomics demand it |
| API contract | **OpenAPI 3.1** → `oapi-codegen` (Go server types) + generated TS client (`openapi-typescript` + `openapi-fetch`) | One source of truth |
| DB | **PostgreSQL** (current supported major) with `pgx`, **`sqlc`** for typed queries, **`goose`** migrations | RLS on workspace_id |
| Jobs/workflows | **River** (Postgres-backed queue) + explicit state machines | Avoids Redis. Revisit Temporal only if workflows get complex |
| Realtime | **SSE** for run/build/chat streams (simpler than WebSockets, proxy-friendly) | WebSocket only if a feature truly needs bidirectional |
| Auth | Server-side sessions in Postgres, **Argon2id**, **WebAuthn passkeys** (`go-webauthn`), TOTP, optional **OIDC** (`go-oidc`) | See §6.2 |
| Object storage | S3 API (MinIO in dev compose) / local FS adapter | Evidence, reports, snapshots |
| Lab provider | `LabProvider` port: **Hyper-V** (primary on Windows), **libvirt/KVM**, **Proxmox**, `fake`; VirtualBox only as legacy adapter if wanted | Vagrant is *not* required |
| Image building | **Packer** (optional) *or* Microsoft evaluation images + **Ansible** provisioning over WinRM-HTTPS / PowerShell Direct | ADR on BSL licence and image licensing |
| Compliance engine | `ComplianceEngine` port: **`native`** (our checks for walking skeleton) and **`wazuh-sca`** (full CIS-style coverage) | Both emit the same `Finding` shape |
| Telemetry | **Sysmon** (community config as base) + **Wazuh agent/indexer**; `TelemetrySource` port | Evaluate outcomes per technique |
| Attack catalog | **Atomic Red Team** YAML, **pinned to a commit SHA**, parsed into our DB; executor = `Invoke-AtomicRedTeam` inside the twin | MITRE ATT&CK STIX data for technique metadata (pin version) |
| Detection mapping | **SigmaHQ** rules as optional coverage reference (check its licence terms) | Detection engineering stretch |
| LLM | **Anthropic Go SDK**; models configurable via env (defaults: Sonnet-class for chat, Haiku-class for cheap tasks, Opus-class for deep analysis) | Verify model IDs in docs at build time |
| Reports | HTML → PDF via headless Chromium (`chromedp`/Playwright) with print CSS | Beats hand-built PDF libs for design quality |
| Frontend | **Next.js 16** (App Router, Turbopack), **React 19.2**, **TypeScript strict**, **Tailwind v4**, **shadcn/ui + Radix**, **TanStack Query/Table**, **React Flow (xyflow)**, **Recharts** or **visx**, **Motion**, **cmdk** | One app, route groups `(marketing)` and `(console)` |
| Monorepo | **pnpm workspaces + Turborepo**; Go modules under `services/` via `go.work` | |
| Observability | `slog` JSON logs, **OpenTelemetry** traces/metrics, Prometheus endpoint, optional Grafana in compose | |
| Packaging | Docker images (distroless/non-root), docker compose for dev/self-host; Helm chart as stretch | |

### 5.5 Repository layout

```
cyberditto/
├─ apps/
│  └─ web/                         # Next.js 16: (marketing), (docs), (console), /demo
├─ services/
│  ├─ api/                         # Go control plane (API + workers + MCP)
│  │  ├─ cmd/{api,worker,migrate}/
│  │  └─ internal/
│  │     ├─ domain/                # entities, state machines, scoring, policy engine
│  │     ├─ app/                   # use cases per module
│  │     ├─ ports/
│  │     ├─ adapters/{postgres,hyperv,libvirt,proxmox,fake,wazuh,atomic,caldera,anthropic,s3}/
│  │     └─ transport/{http,mcp,labgw}/
│  ├─ collector/                   # Go read-only host collector (signed binary)
│  └─ twin-agent/                  # Go agent that runs inside the twin
├─ packages/
│  ├─ api-client/                  # generated TS client
│  ├─ ui/                          # design system (tokens + components)
│  └─ config/                      # eslint, tsconfig, prettier presets
├─ api/openapi.yaml
├─ content/
│  ├─ mappings/                    # control↔technique mappings with provenance
│  ├─ remediation-templates/       # parametrised, tested fixes + rollbacks
│  └─ scenarios/                   # curated emulation scenarios
├─ deploy/
│  ├─ compose/{dev,selfhost,wazuh}.yml
│  ├─ packer/  ├─ ansible/  └─ hyperv/  # lab provisioning
├─ fixtures/                       # synthetic demo data ONLY (never real scans)
├─ docs/{MASTER_PLAN.md,architecture.md,threat-model.md,adr/,runbooks/}
├─ .github/{workflows,ISSUE_TEMPLATE,dependabot or renovate}
├─ .claude/settings.json  CLAUDE.md  lefthook.yml  Taskfile.yml
└─ README.md  SECURITY.md  LICENSE  CONTRIBUTING.md  CODE_OF_CONDUCT.md
```

### 5.6 Domain model (workspace-scoped; every table has `workspace_id`, `created_at`, `updated_at`)

```mermaid
erDiagram
  WORKSPACE ||--o{ MEMBERSHIP : has
  USER ||--o{ MEMBERSHIP : joins
  WORKSPACE ||--o{ ASSET : owns
  ASSET ||--o{ SNAPSHOT : "has versions"
  SNAPSHOT ||--o{ TWIN : "builds"
  TWIN ||--o{ ASSESSMENT : "is assessed by"
  ASSESSMENT ||--o{ FINDING : yields
  TWIN ||--o{ EMULATION_RUN : "is attacked by"
  EMULATION_RUN ||--o{ TEST_EXECUTION : contains
  TEST_EXECUTION ||--o{ TELEMETRY_EVENT : "evidence"
  TEST_EXECUTION ||--|| DETECTION_RESULT : "outcome"
  FINDING }o--o{ GAP : "correlated in"
  TEST_EXECUTION }o--o{ GAP : "correlated in"
  GAP ||--o{ REMEDIATION_PLAN : "fixed by"
  REMEDIATION_PLAN ||--o{ REMEDIATION_STEP : has
  REMEDIATION_PLAN ||--o{ RETEST : "verified by"
  WORKSPACE ||--o{ REPORT : produces
  WORKSPACE ||--o{ AUDIT_EVENT : records
  USER ||--o{ CHAT_THREAD : owns
  CHAT_THREAD ||--o{ CHAT_MESSAGE : has
  CHAT_THREAD ||--o{ APPROVAL_REQUEST : raises
```

Reference tables (global, versioned, read-only at runtime): `attack_technique`, `attack_test` (Atomic catalog, with `risk_tier`, `catalog_commit`), `control` (+ `control_technique_map` with `source`, `confidence`, `provenance_url`), `remediation_template`.

**Key invariants (enforce in domain + DB constraints):**
- A `TEST_EXECUTION` may only reference a `TWIN` in state `ready|running`; never an `ASSET`.
- `attack_test.risk_tier = destructive` can never be scheduled (hard block, not a setting).
- `AUDIT_EVENT` is append-only and hash-chained (`prev_hash`, `hash`).
- Snapshots are immutable; edits create a new version.

### 5.7 Twin lifecycle and run lifecycle

```mermaid
stateDiagram-v2
  [*] --> requested
  requested --> imaging: provider allocates base image
  imaging --> provisioning: boot + agent enrol (mTLS)
  provisioning --> applying: apply L1 config from snapshot
  applying --> verifying: re-collect from twin
  verifying --> ready: fidelity score computed
  ready --> running: run active
  running --> ready: run finished + revert to clean snapshot
  ready --> stopped
  stopped --> ready
  ready --> destroyed
  requested --> failed
  imaging --> failed
  provisioning --> failed
  applying --> failed
  verifying --> failed
  failed --> destroyed
```

```mermaid
stateDiagram-v2
  [*] --> queued
  queued --> preflight: policy engine (risk tiers, target is twin, lab isolated, consent recorded)
  preflight --> rejected: any invariant fails
  preflight --> executing
  executing --> collecting: per test, per prerequisite, per cleanup
  collecting --> evaluating: telemetry window closed
  evaluating --> completed
  executing --> aborted: kill switch / timeout
  executing --> failed
  completed --> [*]
  aborted --> [*]
  rejected --> [*]
```

### 5.8 Key interfaces (Go ports; keep small and testable)

```go
// ports/lab.go
type LabProvider interface {
    CreateNetwork(ctx context.Context, spec NetworkSpec) (NetworkID, error) // isolated by construction
    CreateTwin(ctx context.Context, spec TwinSpec) (TwinHandle, error)
    Snapshot(ctx context.Context, h TwinHandle, name string) error
    Revert(ctx context.Context, h TwinHandle, name string) error
    Start(ctx context.Context, h TwinHandle) error
    Stop(ctx context.Context, h TwinHandle) error
    Destroy(ctx context.Context, h TwinHandle) error
    Metrics(ctx context.Context, h TwinHandle) (Metrics, error)
    Health(ctx context.Context) error
}

// ports/compliance.go
type ComplianceEngine interface {
    Evaluate(ctx context.Context, target TwinHandle, policy PolicyRef) (<-chan FindingEvent, error)
    ListPolicies(ctx context.Context, os OSFamily) ([]PolicyRef, error)
}

// ports/telemetry.go
type TelemetrySource interface {
    Query(ctx context.Context, q TelemetryQuery) ([]Event, error) // window + host + marker
}

// ports/attack.go
type AttackCatalog interface {
    Search(ctx context.Context, f TestFilter) ([]AttackTest, error)
    Get(ctx context.Context, id TestID) (AttackTest, error)
}

// ports/llm.go
type LLM interface {
    Stream(ctx context.Context, req LLMRequest) (<-chan LLMEvent, error)
    Complete(ctx context.Context, req LLMRequest) (LLMResponse, error) // supports JSON-schema output
}
```

### 5.9 Lab design, isolation and sizing

- **Network:** one dedicated virtual switch per lab, **internal-only** (no NAT, no bridge). Twin → control plane traffic only through the **lab gateway listener** (separate port bound to the lab NIC, mTLS, outbound-initiated by the agent). Egress to LAN/internet blocked at the hypervisor *and* in the twin's firewall. Optional "controlled egress" through a logging proxy is a later opt-in.
- **Per-run hygiene:** snapshot `clean` after provisioning; every run starts from `clean` and reverts afterwards; hard timeouts; resource quotas; kill switch from UI/API/CLI.
- **Provider notes:** Hyper-V is the primary adapter on Windows (the platform is already enabled on machines running WSL2; PowerShell Direct avoids network dependency for bootstrapping). libvirt/KVM for Linux hosts. Proxmox API for a remote lab box. VirtualBox is optional and legacy; it conflicts with Hyper-V on some Windows setups.
- **Images:** do not redistribute Windows images. Document how to obtain Microsoft evaluation media; build with Packer or import a user-supplied VHDX. Record the image's licence/expiry in the twin record.
- **Sizing guide (put in docs):** minimum 1 twin = 4 GB RAM / 2 vCPU / 60 GB disk. Wazuh single node adds roughly 4 to 8 GB depending on settings. A 16 GB laptop can run *one* twin + a trimmed Wazuh, but comfortably running both needs 32 GB or a separate box for Wazuh. Provide a "lite mode" that uses the `native` engine and local event-log collection to avoid the Wazuh stack.
- **Demo mode:** `fake` providers + fixtures so the entire UI works with no VMs (used for CI, screenshots, and the public demo).

### 5.10 Collector (replaces `config_collector.ps1`)

- Single static Go binary (Windows first, Linux later). **Read-only.** No network access except an optional explicit `--upload URL`.
- Requires `--consent` and prints a **manifest of every category it will collect** before running; `--dry-run` lists without collecting.
- **Collects (L1):** OS edition/build/UBR, hotfixes, security policy export (`secedit /export`), audit policy (`auditpol`), policy registry keys, Defender prefs/status, firewall profiles + rules, services (config + state), scheduled tasks (metadata), local users/groups (names/flags only, never hashes), installed software + versions, optional features, BitLocker/Secure Boot/TPM/VBS/HVCI/Credential Guard/LSA protection state, SMB/TLS(SChannel)/RDP/WinRM/PowerShell-logging settings, NICs/listening ports/DNS, shares, AppLocker/WDAC policy presence.
- **Never collects:** files, documents, browser data, credentials, tokens, event-log contents, clipboard, screenshots.
- **Output:** canonical JSON (`snapshot.schema.json` versioned), SHA-256 digest, signed (cosign/minisign), optional **pseudonymisation** of hostnames/usernames/MACs/IPs with a local reversible map that never leaves the machine.
- **Integrity:** each field carries `source` (command/API), `collected_at`, and `error` if it failed. **A failed read is reported as `unknown`, never as a default value.**
- Tests: golden-file tests with recorded outputs; fuzz tests for all parsers.

### 5.11 Twin agent and control channel

- Pull-based: agent in the twin dials the **lab gateway** (mTLS, short-lived client certs issued at enrolment, pinned server cert). No inbound ports on the twin.
- Capabilities (allow-listed RPCs only): `CollectSnapshot`, `ApplyConfig(plan)`, `RunAtomic(testId, inputs, runId)`, `Cleanup(testId)`, `GetEvents(window)`, `Ping`, `Shutdown`.
- The agent **refuses** to execute anything not referenced by a signed catalog entry or a signed remediation plan produced by the control plane. No "run arbitrary command" RPC exists.
- Every execution is tagged with `runId` + `testId` markers so telemetry can be correlated.

### 5.12 Emulation and evaluation pipeline

1. **Preflight** (policy engine): twin is `ready`, lab isolation verified, consent recorded, every test is `risk_tier ∈ {benign, intrusive}` and not on the denylist (disk wipe, boot-record tampering, ransomware-like encryption, mass deletion, network scanning beyond lab, etc.).
2. Revert twin to `clean`, start telemetry window.
3. For each test: check/install prerequisites (inside the twin only) → execute → capture stdout/stderr/exit → cleanup → close window.
4. **Evaluate** via `TelemetrySource`: Defender/ASR blocks and alerts, Sysmon events matching the technique, optional Sigma rule hits, Wazuh alerts → classify `blocked|detected|logged|missed` with the evidence rows attached. Event IDs and heuristics are documented in `docs/outcome-classification.md` and verified against real runs (do not assume; test).
5. Persist, update ATT&CK coverage, trigger **correlation**.

### 5.13 Correlation ("gap analysis") engine

- Inputs: failed `Finding`s, executed `TestExecution`s with outcomes, `control_technique_map`.
- Mapping sources (store provenance and confidence on every row): CTID/MITRE mappings of controls to ATT&CK, curated mappings in `content/mappings/`, and **AI-suggested mappings that stay `unreviewed` until a human approves** (never used for scoring until reviewed).
- Output: `Gap` = {controls[], techniques[], evidence[], priority, factor breakdown, recommended remediation template(s)}.

### 5.14 Remediation engine

- **Templates over free-form generation.** `content/remediation-templates/*.yaml` define parameters, the apply script, **rollback script**, verification check, risk notes, and reboot requirement. Every template has Pester tests and is linted with PSScriptAnalyzer.
- AI can *select and parameterise* templates and *draft* new ones. Drafts remain `proposed` and must pass lint + dry-run on a twin + human approval before use.
- **Apply path:** twin only. **Re-test:** re-run the exact tests linked to the gap; record delta. **Prod path:** export bundle (PowerShell, Ansible, GPO backup, Intune/Settings-catalog JSON) with checksums and rollback; never auto-applied by CyberDitto in v2.

### 5.15 API surface (REST, `/api/v1`, JSON, cursor pagination, idempotency keys on POST)

| Resource | Endpoints (summary) |
|---|---|
| auth | `POST /auth/login`, `/auth/logout`, `/auth/passkeys/*`, `/auth/totp/*`, `GET /me` |
| assets | CRUD; `POST /assets/{id}/snapshots` (upload signed snapshot) |
| twins | `POST /twins` (from snapshot), `GET /twins/{id}`, `POST /twins/{id}:start|stop|revert|destroy`, `GET /twins/{id}/fidelity`, `GET /twins/{id}/events` (SSE) |
| assessments | `POST /assessments`, `GET /assessments/{id}`, `/findings?status=&severity=`, `/events` (SSE) |
| catalog | `GET /attack/techniques`, `GET /attack/tests`, `GET /scenarios` |
| runs | `POST /runs` (preflight inline), `POST /runs/{id}:abort`, `GET /runs/{id}`, `/executions`, `/events` (SSE) |
| gaps | `GET /gaps`, `GET /gaps/{id}` |
| remediation | `POST /gaps/{id}/plans`, `POST /plans/{id}:approve|apply-to-twin|retest|export` |
| reports | `POST /reports` (async), `GET /reports/{id}`, `GET /reports/{id}/download` |
| copilot | `POST /chat/threads`, `POST /chat/threads/{id}/messages` (SSE), `POST /approvals/{id}:approve|deny` |
| admin | users, roles, audit log (`GET /audit`), settings, health, version |

Error format: RFC 9457 problem+json. Every mutating call writes an `AUDIT_EVENT`.

---

## 6. Security (the platform must be as hardened as what it assesses)

> CyberDitto holds a map of its users' weaknesses and runs attack tooling. Treat it as a **high-value target** and a **dangerous capability**. Write `docs/threat-model.md` (STRIDE per container in §5.3) in Phase 0 and keep it current.

### 6.1 Threat model highlights

| Threat | Example | Primary controls |
|---|---|---|
| Stolen session/credentials | Phished admin | Passkeys, TOTP, short sessions, device binding option, anomaly alerts |
| Unauthenticated API access (v1 flaw) | Anyone on LAN triggers VM builds/attacks | Deny-by-default authZ, listener binding, rate limits |
| Attack tooling escapes lab | Test reaches LAN/internet | Hypervisor-level isolation, no NAT, egress deny, preflight verification, kill switch |
| Malicious/poisoned catalog or AI output | Injected payload in "atomic" YAML or LLM-drafted script | Pinned+verified catalog, risk tiers, signed plans, no arbitrary-exec RPC, human approval |
| Prompt injection via telemetry/logs | Log line instructs the copilot to do something | Treat tool output as untrusted data; read-only tools by default; approvals for mutations (§7.4) |
| Data exfiltration to LLM | Hostnames/users/IPs leave the box | Redaction/pseudonymisation layer, opt-in per workspace, "no-AI mode" |
| Supply-chain compromise | Malicious npm/Go dependency, tampered release | Pinning, `govulncheck`, OSV scan, SBOM, signed images/binaries, provenance attestations |
| Sensitive data at rest | DB or evidence leaked | Disk encryption guidance, envelope encryption for stored secrets, retention + delete |
| Insider/audit evasion | Someone runs attacks and hides it | Append-only hash-chained audit log, exported to external sink |

### 6.2 Identity, sessions, access control

- **No default credentials, ever.** First-run setup wizard (only available until the first owner exists) creates the owner; API refuses to start in "open" mode.
- **Passwords:** Argon2id with tuned parameters, breached-password check (k-anonymity API or offline list), lockout/backoff, constant-time compares.
- **Passkeys (WebAuthn)** preferred; **TOTP** fallback; recovery codes hashed.
- **Sessions:** server-side, random 256-bit IDs, stored hashed, `HttpOnly; Secure; SameSite=Lax` cookie (`__Host-` prefix), idle + absolute timeouts, rotation on privilege change, "sign out everywhere". **No tokens in `localStorage`.**
- **CSRF:** SameSite + Origin/Referer check + CSRF token on state-changing browser requests.
- **RBAC:** `owner`, `admin`, `analyst`, `viewer`, plus `approver` capability for attack runs and remediation apply. Check on every handler via a single policy function; write table-driven authZ tests (every route × every role).
- **Postgres RLS** on `workspace_id` as defence in depth; app connects with a non-superuser role.
- **Service-to-service:** lab gateway uses mTLS with short-lived certs; MCP server requires scoped API tokens (hashed at rest, per-scope, expiring).

### 6.3 Application and API hardening

- Strict JSON decoding (`DisallowUnknownFields`), request size/time limits, pagination caps, OpenAPI request validation, output encoding.
- Security headers: strict **CSP with nonces** (no `unsafe-inline`), HSTS, `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy`, COOP/CORP, frame-ancestors none.
- CORS: explicit allow-list from config (no wildcards with credentials).
- Rate limiting + abuse controls on auth, chat, and public "Ask Ditto" endpoints; Cloudflare Turnstile (or similar) on public forms.
- **No shell-string building.** Executables are allow-listed; args are validated against strict regex/enum types (e.g. `TwinID` is a UUID type, not a string).
- SSRF guard for any outbound fetch (deny private ranges unless explicitly configured).
- File uploads (snapshots): size cap, schema validation, signature verification, stored outside web root, content-addressed.
- Dependency hygiene: Renovate (or Dependabot), `govulncheck`, `pnpm audit`/OSV-Scanner, Trivy on images, CodeQL + Semgrep, `gitleaks`, license check. **Next.js/React have had serious advisories historically: pin to the latest patched releases and keep auto-update PRs on.**

### 6.4 Lab safety invariants (non-negotiable; each has a test that fails the build if violated)

1. Attack execution targets only `TWIN` records; the executor has no code path accepting an `ASSET` or arbitrary host.
2. The control plane verifies **lab isolation** before every run (provider reports switch type internal-only; twin egress probe fails as expected).
3. `risk_tier=destructive` tests cannot be scheduled. The denylist is code, not configuration.
4. Every run requires a recorded **authorisation acknowledgement** by a user with `approver` capability.
5. Runs have hard wall-clock and resource limits; **kill switch** reverts and powers off the twin.
6. The twin agent executes only catalog-signed tests or signed remediation plans.
7. The catalog is pinned to a commit SHA and verified (hash manifest). Updating the pin is a reviewed change.
8. AI-generated tests/scripts are `draft` and never auto-executable.
9. Caldera integration (in-scope for Phase 4, `docs/adr/0011`) runs in its own container/VM on the lab network only, **patched to a release that fixes CVE-2025-27364**, never exposed beyond the lab.
10. The product shows a persistent "Authorised testing only" notice and ships an Acceptable Use Policy.

### 6.5 Data protection and privacy

- Minimise: collector manifests, no content collection (§5.10).
- **Pseudonymisation layer** (hostnames, user names, MACs, IPs, domain names) used for exports, demo data, and anything sent to an LLM.
- Encrypt at rest (volume encryption documented; envelope encryption for stored credentials/API keys using a KMS adapter, `age`/SOPS for dev).
- Configurable retention; one-click workspace export and delete.
- Logs never contain secrets or snapshot bodies; log scrubbing tests.

### 6.6 Supply chain and release security

- Reproducible-ish builds in GitHub Actions; actions pinned by commit SHA.
- **SBOM** (CycloneDX via Syft) per image/binary; **cosign keyless signing** for images and collector/agent binaries; **GitHub artifact attestations / SLSA provenance**; OpenSSF Scorecard in CI with badge.
- Branch protection, required checks, CODEOWNERS, signed commits *by the owner's own key* (optional).
- `SECURITY.md` with a disclosure process and security contact; a documented threat model; an "Assurance" page on the website.

### 6.6.1 Windows collector signing

Authenticode requires a code-signing certificate. If not available, publish SHA-256 + cosign signature + SBOM and document the SmartScreen warning honestly.

### 6.7 Compliance and legal hygiene

- CIS Benchmarks are copyrighted and licensed by CIS; **review CIS terms before bundling or redistributing benchmark text, and avoid implying CIS certification**. Prefer referring to control IDs and your own wording; Wazuh SCA content is its own open implementation (check its licence for redistribution).
- Atomic Red Team (MIT), Caldera (Apache-2.0), Wazuh (GPL for agent/manager components; run as separate processes), Sigma (check Detection Rule License terms). Create `THIRD_PARTY_NOTICES.md` automatically in CI.
- Windows images: licence compliance, no redistribution.
- Ship: AUP, `SECURITY.md`, privacy notice for the website, README disclaimer ("for authorised use on systems you own or have written permission to test").

---

## 7. Ditto Copilot (AI layer)

### 7.1 Goals

Help users **understand, decide, and act**: explain findings and attack results, prioritise, draft remediation, compare runs, write reports, answer "what if", and use CyberDitto by conversation, always grounded in the user's own evidence, never hallucinating results.

### 7.2 Capabilities

| Capability | How |
|---|---|
| Explain a finding/technique/outcome | Retrieval of finding + evidence rows + ATT&CK technique metadata; answer with **citations** to evidence IDs |
| Prioritise "what do I fix first?" | Calls `list_gaps` tool; explains factor breakdown from §4.4 |
| Remediation drafting | **Structured output** against a JSON schema `{template_id?, params, steps[], rollback[], verification[], risk, reboot}`; validated, linted, dry-run on twin before it can be approved |
| Run planning | Proposes a scenario (techniques, order, expected outcomes) as an `ApprovalRequest`, never executes directly |
| Compare runs / drift | `compare_runs` tool returns structured diff; model narrates |
| Report writing | Executive summary + technical narrative drafted by a cheaper model, human-editable before export |
| Threat-report → emulation plan | User pastes CTI text; model extracts TTPs (structured) → maps to catalog tests → plan for approval |
| Detection suggestions | For `missed/logged` techniques, drafts Sigma rule candidates (labelled `draft`) |
| Public "Ask Ditto" on the website | Docs-only RAG, Haiku-class model, strict rate/budget limits, no tools, no user data |

### 7.3 Architecture

- Orchestration lives in the **Go control plane** (`app/copilot`) using the Anthropic Go SDK; it is the only place that holds the API key. The browser never sees it.
- **Model routing (config-driven, verify IDs):** fast/cheap model for classification, summaries, public docs bot; default chat model for the in-app copilot; the largest model only for opt-in "deep analysis".
- **Prompt caching** for the large static system prompt + ATT&CK/tool schemas; **streaming** over SSE to the UI.
- **Tools** (JSON-schema, strict tool use):
  - *Read-only (auto-approved):* `get_twin`, `list_findings`, `get_finding`, `list_runs`, `get_run`, `get_execution_evidence`, `list_gaps`, `search_attack`, `get_remediation_template`, `compare_runs`, `explain_score`.
  - *Mutating (always produce an ApprovalRequest, executed only after human click):* `propose_run`, `propose_remediation`, `propose_report`, `propose_twin_build`.
- **Grounding rule** in the system prompt and enforced in code: any claim about the user's environment must cite a tool result; if no evidence exists, say so.
- **MCP:** ship `cyberditto-mcp` (same tool set, scoped tokens, read-only by default) so Claude Desktop/Claude Code can query it. Optionally *consume* the Atomic Red Team MCP server for test search/validation, with **execution disabled** unless running inside the lab VM.
- **Evals:** golden conversations + rubric tests (grounding, refusal on unsafe asks, schema validity, no secret leakage) run in CI against the `fake` LLM and nightly against the real API with a small budget.

### 7.4 LLM-specific security (map to OWASP LLM Top 10)

- **Prompt injection:** all tool results are wrapped and labelled as untrusted data; the system prompt states data never contains instructions; mutating tools can only create approval requests; approval UI shows the *actual* command/plan diff, not the model's description of it.
- **Data minimisation:** redaction/pseudonymisation before any content leaves the host; workspace-level toggle "Allow AI features" (default **off** until enabled), and "Local-only mode" to disable all external calls.
- **Output handling:** model text is rendered as sanitised Markdown (no raw HTML), links rewritten/allow-listed, code blocks never auto-executed.
- **Abuse/cost control:** per-user and per-workspace token budgets, max tool-loop depth, request timeouts, circuit breaker, spend dashboard for admins.
- **Auditability:** store prompts, tool calls, and outputs (redacted) with run/thread IDs; link every approval to the model message that proposed it.
- **No autonomy over safety invariants:** the policy engine (domain code) is the final authority; the model cannot override it.

---

## 8. Frontend, UX and website

### 8.1 Design direction

- **Brand idea:** *Ditto* = the shapeshifter, a perfect copy. Visual motif: two mirrored, slightly morphing shapes (source ↔ twin) with a soft "gooey" blob transition. Use an original mark; **do not use any Pokémon IP/imagery**.
- **Mood:** dark-first "security operations" aesthetic, calm and dense, with one confident accent. Light theme supported.
- **Tokens (starting point, tune in Phase 7):**
  - Background `#09090d`, surface `#111118`, raised `#171722`, border `rgba(255,255,255,.08)`
  - Accent (Ditto violet → fuchsia gradient) `#8b5cf6 → #d946ef`
  - Outcome colours: **Blocked** emerald, **Detected** sky, **Logged** amber, **Missed** rose; severity scale for findings
  - Type: **Geist Sans + Geist Mono** (via `next/font`), tabular numerals for metrics
  - Radius 10 to 14px, 1px hairline borders, subtle inner glow on hover, motion ≤ 200ms, `prefers-reduced-motion` respected
- **Accessibility:** WCAG 2.2 AA, full keyboard nav, visible focus, colour never the only signal (icons + labels on outcomes), `axe` checks in CI.

### 8.2 Inspiration to study (look at them; borrow *patterns*, never assets, copy or branding)

| Reference | Borrow |
|---|---|
| **Linear** | Density, keyboard-first UX, command palette, crisp list/detail layouts |
| **Vercel dashboard + marketing** | Dark theme, mono accents, deploy-log streaming UI, clean docs |
| **Wiz / Vanta / Drata** | Posture score cards, compliance progress, trend lines, finding drill-downs |
| **Picus, AttackIQ, SafeBreach, Cymulate** | ATT&CK heatmap with prevention vs detection layers, per-technique severity, scenario library, scheduled assessments, executive reports |
| **Wazuh dashboard / Grafana** | Panel layouts, time-range controls, event tables |
| **Raycast** | Command palette polish |
| **Resend / Stripe docs / Clerk** | Marketing clarity, docs typography, code-sample blocks |
| **shadcn/ui blocks** (dashboard, sidebar, auth) | Fast, accessible starting layouts |

### 8.3 Information architecture (console)

```
Overview (posture + resilience + recent activity + "next best action")
Assets            └ Asset detail → Snapshots → Diff between snapshots
Twins             └ Twin detail: Fidelity, Topology graph, Config diff (source vs twin), Console/Events, Resources
Assessments       └ Run detail: score, controls table (filter/group/search), evidence drawer
Emulation         └ Scenario library, Run builder (with preflight panel), Live run console, Results
ATT&CK            └ Interactive matrix heatmap (Blocked/Detected/Logged/Missed), technique drawer
Gaps              └ Prioritised list with factor breakdown → Remediation plan → Re-test delta
Reports           └ Generate/preview/download (PDF, JSON, SARIF, Navigator layer, CSV)
Copilot           └ Global drawer (⌘J) + full-page threads
Settings          └ Workspace, members/roles, AI & privacy, integrations, API tokens, audit log, about
```

### 8.4 Key screens and components (each needs loading, empty, error and success states)

- **Overview:** radial posture gauge, resilience gauge, 30-day trend sparklines, "top 5 gaps", recent runs, system health (lab, Wazuh, AI).
- **Twin builder:** stepper (select snapshot → options → review → build) + live event stream + fidelity ring + "known gaps" panel.
- **Twin topology (React Flow):** source ↔ twin ↔ services; later multi-host networks.
- **Config diff:** side-by-side source vs twin (Monaco diff or a custom table), filterable by category.
- **Controls table (TanStack Table):** virtualised, grouped by section, severity chips, bulk actions, saved views, evidence drawer with "Ask Ditto about this" button.
- **ATT&CK matrix:** tactics as columns, technique cells coloured by outcome; toggle layers (prevention / detection / evidence available); click → drawer with test, telemetry timeline, remediation links; export Navigator layer.
- **Run builder:** scenario cards, technique picker with risk-tier badges, **preflight checklist** (twin ready ✓ isolated ✓ consent ✓ tiers ✓), big "Authorise & run".
- **Live run console:** terminal-style stream (virtualised), per-test status timeline, kill switch always visible.
- **Gap detail:** "Why this ranks here" factor bars, linked control + technique + evidence, remediation plan stepper (draft → twin-applied → re-tested → exported), **before/after delta card**.
- **Report preview:** paginated print-style preview identical to the PDF.
- **Command palette (⌘K):** navigate, search assets/techniques/controls, run actions.
- **Copilot drawer:** context-aware (knows the open entity), streaming, collapsible **tool-call trace** ("show work"), citation chips that open evidence, **approval cards** showing the exact command/plan diff with Approve/Deny, suggested prompts, thread history, thumbs feedback, admin-only token/cost meter, clear "AI can be wrong" affordance.
- **Global:** toast system, skeletons not spinners, optimistic UI where safe, URL-synced filters, keyboard shortcuts help (`?`).

### 8.5 Frontend engineering rules

- TypeScript `strict`, no `any` (lint-enforced), Zod at boundaries where API types don't cover.
- Data: TanStack Query with the generated client; SSE hooks with reconnect/backoff and `Last-Event-ID`.
- Server components for static/marketing content; client components only where interactivity is needed. Cookies are httpOnly and handled by the Go API (the web app proxies via same-site rewrites so the browser talks to one origin).
- Design system in `packages/ui` with Storybook (or Ladle) + visual regression (Playwright screenshots).
- Performance budgets: marketing LCP < 2.0s on mid-tier mobile; console route JS budgets; no layout shift; images via `next/image`; fonts self-hosted.
- i18n-ready strings (English only at launch).

### 8.6 Public website (marketing + docs + demo), the portfolio centrepiece

**Pages:** `/` (landing), `/product` (feature deep-dives), `/how-it-works` (animated 4-step flow), `/security` (assurance: threat model summary, safety invariants, supply-chain badges), `/docs/*`, `/demo`, `/changelog`, `/about` (story + roadmap), `/blog` (2 to 3 posts: "Why a digital twin", "Closing the loop: attack → fix → retest", "Building a safe attack lab").

**Landing page structure:**
1. **Hero:** headline "Attack your twin, not your production." + sub-copy + two CTAs (*Try the live demo*, *View on GitHub*) + animated visual: source machine → morphing twin → attack beams → ATT&CK heatmap lighting up.
2. **Trust strip:** MITRE ATT&CK, Atomic Red Team, Wazuh, Claude (use text/logos only where brand guidelines allow; otherwise text).
3. **Problem → loop:** "Collect → Twin → Assess → Attack → Fix → Re-test" interactive stepper.
4. **Bento grid:** Fidelity score, ATT&CK heatmap, evidence-weighted risk, Copilot, safe-by-design lab, reports/exports.
5. **Interactive mini-demo:** scrubbable before/after (failed control + successful technique → fix → re-test blocked).
6. **Copilot section:** short looping chat with approval card.
7. **Security & safety section** (the invariants as a checklist), architecture diagram, ~~"open source" block with stars/CI/Scorecard badges~~ — **superseded 2026-11-10**: the project is not open source (see `docs/adr/0001-license-and-repo-identity.md`); this becomes a "public code, CI/Scorecard/supply-chain transparency" block instead — stars/forks framing dropped, CI/Scorecard/signed-release badges kept since those are about code quality and build integrity, not a license grant.
8. **Footer CTA**, docs links, AUP/Privacy/Security links.

**Docs:** MDX-based (**Fumadocs** or Nextra; pick via ADR) with quickstart, concepts, architecture, API reference (generated from OpenAPI), safety model, runbooks, FAQ. Includes the **"Ask Ditto" docs chatbot** (§7.2) with search fallback.

**`/demo` (no backend needed):** the console running on **MSW (Mock Service Worker) + synthetic fixtures** (`fixtures/`), clearly badged "Demo data". A guided tour (5 to 7 steps) highlights the loop. Resettable. No real data paths exist in the demo bundle. Deployable to Vercel/Netlify/Cloudflare Pages as a static + edge app.

**SEO/meta:** OG images generated per page, sitemap, structured data, fast, privacy-respecting analytics (Plausible/Umami) with no cookie banner needed.

---

## 9. Developer experience, quality and operations

### 9.1 Tooling

- **Task runner:** `Taskfile.yml` with `task dev | test | lint | gen | demo | e2e | build | release`.
- **Local stack:** `deploy/compose/dev.yml` (Postgres, MinIO, Mailpit, optional Wazuh profile). One command: `task dev`.
- **Devcontainer** for contributors (Go, Node, pnpm, Task, gitleaks preinstalled).
- **Pre-commit hooks** (lefthook): gofmt/goimports, golangci-lint, eslint, prettier, tsc, gitleaks, commit-message check (§12.3).
- **Conventional Commits** + `release-please` (changelog + versioning).
- **Go lint set:** `golangci-lint` with `gosec`, `errcheck`, `staticcheck`, `depguard` (architecture rule), `govet`, `revive`.

### 9.2 CI/CD (GitHub Actions; actions pinned to SHAs)

| Workflow | Contents |
|---|---|
| `ci.yml` | Go build/test/race, lint, `govulncheck`; web typecheck/lint/unit; OpenAPI diff + codegen drift check; migrations up/down test against Postgres service |
| `e2e.yml` | Playwright against compose stack using `fake` providers; a11y (axe) assertions |
| `security.yml` | CodeQL, Semgrep, gitleaks, Trivy (fs + image), OSV-Scanner, dependency review, Scorecard |
| `release.yml` | Build multi-arch images, SBOM (Syft), cosign sign, attestations, publish collector/agent binaries with checksums |
| `commit-hygiene.yml` | Fails PRs whose commits contain attribution trailers (§12.3) |
| `deploy-web.yml` | Deploy website + demo to the chosen host on `main` |

### 9.3 Observability

`slog` JSON with request/run/twin IDs; OpenTelemetry traces across API → worker → provider; Prometheus metrics (run duration, outcome counts, queue depth, LLM tokens/cost/latency, provider errors); health endpoints (`/healthz`, `/readyz`); admin "System health" page; optional Grafana dashboards in `deploy/compose`.

### 9.4 Test strategy

| Layer | Tools | What |
|---|---|---|
| Domain unit | Go `testing`, table tests, `rapid`/fuzz | Scoring, policy engine, state machines (property tests for invariants in §5.6/§6.4) |
| Adapter integration | `testcontainers-go` (Postgres, MinIO) | Repos, queue, RLS, migrations |
| Contract | OpenAPI validation + Schemathesis | Server matches spec; fuzz endpoints |
| AuthZ matrix | Generated table tests | Every route × every role |
| Collector/agent | Golden files, fuzz parsers | No hardcoded values; unknown ≠ default |
| Frontend unit/component | Vitest + Testing Library | Components, hooks, SSE client |
| E2E | Playwright | Full journeys in demo + compose modes |
| Visual | Playwright screenshots / Storybook | Design-system regressions |
| A11y | axe-core | Zero serious/critical violations |
| AI evals | Custom harness (fake LLM in CI; real API nightly, tiny budget) | Grounding, schema validity, injection resistance, refusals |
| Lab integration (manual/nightly on a real box) | Hyper-V runner | Twin build, fidelity, one benign Atomic, outcome classification |
| Security | `gosec`, Semgrep, own pen-test checklist (OWASP ASVS L2 subset) | Phase 8 |

**Coverage targets:** domain ≥ 90%, app ≥ 80%, overall ≥ 75%. Don't chase numbers over meaning.

---

## 10. Roadmap (phases, tasks, acceptance criteria)

> Check boxes as you go (mirror into `docs/PROGRESS.md`). In autonomous mode, **do not stop for owner review**; pass the self-review gate (§0.5.5) and continue, stopping only at the hard stops (§0.5.3).

### Phase 0: Foundations and hygiene (≈ 1 week)
- [ ] Fresh repo, license (§14 open question), `README` stub, `SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`
- [ ] Monorepo scaffold (pnpm + Turborepo + `go.work`), Taskfile, devcontainer, compose dev stack
- [ ] `.claude/settings.json`, `CLAUDE.md`, commit-msg hook, lefthook, gitleaks (§12)
- [ ] CI skeleton (`ci`, `security`, `commit-hygiene`) green on an empty app
- [ ] `docs/threat-model.md` v0, `docs/architecture.md`, ADRs 0001 to 0010 (stack, auth, queue, provider abstraction, catalog pinning, LLM routing, licensing, monorepo, design system, docs framework)
- [ ] `api/openapi.yaml` with health + auth stubs; codegen wired for Go + TS
- [ ] State files from §0.5.1; repo-hygiene guard: pre-commit + CI **block** files named `security_posture.json`, `*.vagrant*`, `.vagrant/`, `node_modules/`, `.env*` (except `.env.example`), and any path containing `legacy/`; custom `gitleaks` rules for private keys, JWT secrets, and Windows user-profile paths (`C:\\Users\\`)
- [ ] First-commit audit: confirm `git log` shows only fresh history and the author is the owner's no-reply identity

**Accept:** `task dev` starts everything; CI green; secret scan clean; hook rejects an attribution trailer in a test commit.

### Phase 1: Walking skeleton (≈ 2 weeks)
- [ ] Auth: first-run wizard, Argon2id, sessions, TOTP, passkeys, RBAC, audit log (hash-chained)
- [ ] Postgres schema + migrations + RLS; `sqlc` repos; River queue
- [ ] Assets + Snapshots (upload + schema validation + signature check) using **fixture** snapshots
- [ ] `fake` LabProvider: twin lifecycle state machine with progress events over SSE
- [ ] Console shell in Next.js: layout, nav, command palette, theme, generated client, auth pages, Assets and Twins pages
- [ ] AuthZ matrix tests; e2e: login → add asset → build fake twin → watch progress

**Accept:** full journey works in the browser against the real API with fake lab; authZ matrix passes; no tokens in `localStorage`; CSP enforced without violations.

### Phase 2: Collector and real twin (≈ 3 weeks)
- [ ] Go collector (consent, manifest, dry-run, L1 categories, golden tests, signing), pseudonymisation
- [ ] Hyper-V `LabProvider` (isolated switch, VHDX import/Packer build, snapshots, PowerShell Direct bootstrap), mTLS lab gateway, twin-agent enrolment
- [ ] `ApplyConfig` from snapshot (Ansible/PowerShell), re-collect, **Fidelity Score + known-gaps panel**
- [ ] Twin detail UI: fidelity ring, config diff, resource metrics, console events
- [ ] Docs: lab setup guide, sizing, image licensing

**Accept:** on a real Windows 11 host: collect → build twin → fidelity ≥ agreed threshold (set after first measurement) → revert/start/stop/destroy work; twin has no route to LAN/internet (verified by automated probe).

### Phase 3: Assessment engine (≈ 2 to 3 weeks)
- [ ] `ComplianceEngine` port + `native` engine (≥ 60 high-value Windows 11 controls, each with evidence + remediation link)
- [ ] `wazuh-sca` adapter (compose profile; enrol twin agent; import policy results; map to our `Finding`)
- [ ] Posture score, controls table UI, evidence drawer, assessment history + trend
- [ ] Policy/benchmark selection by OS build

**Accept:** assessment on a real twin matches an independent spot-check of 10 controls; scores reproducible; "unknown" handled honestly.

### Phase 4: Emulation, telemetry, ATT&CK (≈ 3 weeks, now including Caldera — see below)
- [ ] Atomic catalog importer (pinned SHA, hash manifest, risk tiers, denylist) + ATT&CK data import (pinned)
- [ ] Twin-agent `RunAtomic` with prerequisites, cleanup, run/test markers
- [ ] **Caldera integration (in scope per `docs/adr/0011-caldera-in-scope.md`, owner instruction 2026-11-10):** Caldera server pinned to a release ≥ 5.1.0 (verify the actual latest patched release at build time, per §0.3) that fixes CVE-2025-27364, run in its own container/VM on the lab network only, never exposed beyond it; Sandcat agent + `emu` plugin evaluated for inclusion under the same pin/verify rules as the Atomic catalog; a second `AttackCatalog`/execution adapter so Caldera operations and Atomic tests share one preflight, risk-tier, and outcome-classification pipeline
- [ ] Sysmon + Wazuh telemetry pipeline; `TelemetrySource`; **outcome classifier** with documented heuristics validated on real runs, covering both engines
- [ ] Policy engine preflight (all §6.4 invariants, including invariant 9 for Caldera), kill switch, authorisation acknowledgement
- [ ] UI: scenario library, run builder + preflight panel (shows which engine a scenario uses), live console, ATT&CK matrix heatmap, technique drawer, Navigator export
- [ ] 6 to 10 curated **benign/intrusive** scenarios (credential access simulation, persistence, discovery, defence-evasion basics, etc.), drawing on both Atomic Red Team and Caldera/CTID plans where appropriate

**Accept:** a run of ≥ 20 benign tests across both engines produces classified outcomes with evidence; invariant tests prove destructive tests and non-twin targets are impossible for *both* Atomic and Caldera execution paths; kill switch verified for both; Caldera's own isolation (own container/VM, no route beyond the lab network) is verified by an automated probe, not just asserted.

### Phase 5: Correlation, remediation, re-test, reports (≈ 3 weeks)
- [ ] Control↔technique mapping content with provenance; correlation engine; gap priority with factor breakdown
- [ ] Remediation templates (≥ 25 with apply/rollback/verify, Pester-tested, PSScriptAnalyzer-clean)
- [ ] Plan → approve → apply to twin → re-test → delta; export bundles (PS/Ansible/GPO/Intune)
- [ ] Reports: HTML→PDF, JSON, SARIF, Navigator layer, CSV; report preview UI

**Accept:** demonstrable closed loop on a real twin: a control fails → technique succeeds → fix applied → same technique now blocked/detected → delta recorded → report generated.

### Phase 6: Ditto Copilot and MCP (≈ 2 to 3 weeks)
- [ ] LLM port + Anthropic adapter, redaction layer, budgets, audit
- [ ] Tools (read-only + approval-gated), streaming chat API, thread persistence
- [ ] Copilot drawer UI: citations, tool trace, approval cards
- [ ] Structured remediation drafting with validation; threat-report → plan; report narrative
- [ ] `cyberditto-mcp` server with scoped tokens; docs for Claude Desktop/Code setup
- [ ] AI eval suite (injection, grounding, schema, refusals) in CI

**Accept:** eval suite green; injection tests cannot trigger a mutating action without approval; AI toggled off → app fully functional.

### Phase 7: Website, docs, demo, polish (≈ 3 weeks)
- [ ] Marketing site (§8.6), animated hero, bento, interactive mini-demo
- [ ] Docs site + "Ask Ditto" docs bot (rate-limited, budget-capped, Turnstile)
- [ ] `/demo` on MSW + fixtures with guided tour; reset
- [ ] Design polish pass (empty/loading/error states everywhere, motion, dark/light), a11y audit, perf budgets
- [ ] OG images, SEO, analytics, changelog, 3 blog posts

**Accept:** Lighthouse ≥ 95 (perf/a11y/best-practices/SEO) on marketing pages; demo works offline of the API; axe clean; a stranger can understand the product in 60 seconds.

### Phase 8: Hardening and launch (≈ 2 weeks)
- [ ] Self pen-test using an OWASP ASVS L2 checklist; fix findings; publish a short "security assessment" doc
- [ ] Signed releases (cosign, SBOM, provenance), Scorecard ≥ target, branch protection
- [ ] Self-host guide (compose), backup/restore runbook, upgrade guide
- [ ] Demo video (60 to 90s) + README GIFs + architecture diagram
- [ ] Tag `v2.0.0`

**Accept:** fresh-machine install following only the docs works; release artefacts verify with documented commands.

---

## 11. More ideas to make it exceptional (pick after Phase 5)

**Product**
1. **Continuous validation:** nightly/weekly scheduled runs on twin refresh; **drift detection** when the real asset's new snapshot diverges from the twin.
2. **"Twin time machine":** compare twin versions; "what changed and how did resilience move?"
3. **Threat-informed mode:** choose a threat actor/campaign (CTID emulation plans) and see coverage against *that* actor.
4. **Detection engineering loop:** for `missed/logged` techniques, generate candidate Sigma rules, test them against the captured telemetry, show before/after detection.
5. **Attack path graph** (React Flow) for multi-host/AD twins (BloodHound-style), L2 fidelity.
6. **Linux and macOS twins**, then container/Kubernetes twins.
7. **CI/CD integration:** GitHub Action that validates a golden image or IaC baseline against a twin on pull requests, posts a posture/resilience delta comment.
8. **Framework mappings:** NIST 800-53, CIS Controls v8, ISO 27001, and **OSCAL** export for compliance workflows.
9. **Ticketing/chat integrations:** GitHub Issues/Jira/Slack for gaps and run summaries.
10. **Risk quantification** for execs: rough "likelihood × impact" with transparent assumptions.
11. **Benchmark comparisons:** anonymised peer baselines (opt-in) or comparisons against Microsoft security baselines.
12. **Gamified training mode:** "Cyber range" scenarios for students with scoring and hints (great for the education persona).
13. **Browser extension / CLI:** `cyberditto` CLI for scripting (collect, upload, run, wait, export).
14. **Cloud twin (stretch):** Azure/AWS lab provider with strict budget guards, using cloud-native isolation.

**Engineering polish**
15. Feature flags; plugin interface for compliance engines and attack catalogs; public API docs with try-it console.
16. Multi-tenant SaaS hardening (per-workspace encryption keys, quotas) with Helm chart.
17. Accessibility statement and a keyboard-shortcuts cheat sheet; localisation-ready.
18. Public **"Safety Report"**: an auto-generated page proving the lab invariants pass in CI (badge per invariant).

---

## 12. Git, commit identity and attribution rules

**Goal:** commits and PRs are authored by the owner with **no AI attribution lines** (no `Co-Authored-By: Claude`, no "Generated with Claude Code" footer, no `Claude-Session:` trailer). A second GitHub account is **not** needed. Claude Code adds those lines only via its attribution feature, which can be turned off, and a git hook makes it impossible for them to slip in.

> **Owner instruction on record (2026-11-10): no AI attribution, ever, in this repository's commits or PRs — no exceptions, no defaults override this.** This supersedes any later tool-level default that tries to re-add a `Co-Authored-By` or "Generated with" line; every commit and PR produced for this project omits them.

### 12.1 `.claude/settings.json` (commit this file in the repo)

> **Amendment (2026-11-10):** the owner asked for `.claude/` to be removed from the repo; it no longer exists as committed config. The attribution rule it encoded is still enforced independently (the commit-msg hook in §12.3 + `commit-hygiene.yml` + the standing owner instruction above); the `permissions.deny` guardrails this file carried (blocking force-push, remote changes, `gh repo delete`, etc.) are not currently replaced by anything else — see `docs/PROGRESS.md`. The JSON below is kept as a reference for what it contained, not as a file to recreate unless asked.

```json
{
  "attribution": {
    "commit": "",
    "pr": "",
    "sessionUrl": false
  },
  "permissions": {
    "deny": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)",
      "Bash(git push --force:*)",
      "Bash(git push -f:*)",
      "Bash(git push --force-with-lease:*)",
      "Bash(git remote add:*)",
      "Bash(git remote set-url:*)",
      "Bash(gh repo delete:*)",
      "Bash(gh auth token:*)",
      "Bash(rm -rf /:*)"
    ]
  }
}
```

Notes: set **both** `commit` and `pr` to empty strings (setting only one leaves the other default). `sessionUrl: false` stops the separate `Claude-Session:` trailer that newer versions append. Older tutorials mention `includeCoAuthoredBy: false`; that key is deprecated. Permission-rule syntax and attribution keys evolve: **verify against current Claude Code settings docs** and put a personal copy in `~/.claude/settings.json` too. Cloud/remote sessions only read the committed project file.

### 12.2 `CLAUDE.md` (repo root; loaded automatically)

```markdown
# CyberDitto: instructions for Claude Code
- Read docs/MASTER_PLAN.md first. Follow §0.4 operating rules and §6.4 safety invariants.
- Commits: Conventional Commits (feat:, fix:, docs:, test:, refactor:, chore:, ci:, sec:). Imperative, ≤ 72-char subject, body explains *why*.
- NEVER add Co-Authored-By lines, "Generated with Claude Code", Claude-Session trailers, or any mention of Claude/Anthropic/AI tooling in commit messages or PR descriptions.
- Never use --no-verify. Never force-push. Never commit secrets, .env files, real scan data, or personal identifiers.
- One logical change per commit. Run `task lint test` before committing.
- Autonomous mode: at the start of every session read docs/PROGRESS.md, docs/BLOCKERS.md and docs/OWNER_TODO.md, then continue from "next action". Follow MASTER_PLAN §0.5 (hard stops, self-review gate, PR flow).
- Legacy v1 lives outside this repo: never copy from it, never read its .git, never commit anything referencing it.
```

### 12.3 Hook that *enforces* it (`.githooks/commit-msg`)

```bash
#!/usr/bin/env bash
# Rejects AI-attribution trailers/footers. Enable once with:
#   git config core.hooksPath .githooks && chmod +x .githooks/commit-msg
msg_file="$1"
if grep -Eiq '^(co-authored-by:.*(claude|anthropic|noreply@anthropic\.com)|claude-session:)|generated with .?claude|🤖 generated' "$msg_file"; then
  echo "✖ Commit rejected: remove AI attribution lines (see docs/MASTER_PLAN.md §12)." >&2
  exit 1
fi
```

Mirror the same regex in `.github/workflows/commit-hygiene.yml` (scan `git log base..head --format=%B`) so a PR can't merge with such a trailer. If one ever slips into an unpushed commit: `git commit --amend` (last commit) or `git rebase -i` to reword; if already pushed to your *own* branch, rewrite and `git push --force-with-lease` (the deny-rules in §12.1 block Claude from force-pushing and changing remotes, so do that step yourself).

### 12.4 Commit identity and privacy

```bash
git config user.name  "Your Name"
git config user.email "<id>+<username>@users.noreply.github.com"   # GitHub no-reply address
```

On GitHub enable **"Keep my email addresses private"** and **"Block command line pushes that expose my email"**. (The old repo's history exposes a personal email address.) Optional: sign commits with your own SSH/GPG key.

### 12.5 Branching and review

``main` protected (PR + required checks, applied by the owner or by `gh` once authenticated; else listed in OWNER_TODO); short-lived feature branches `phase-N/topic`; squash-merge with a Conventional-Commit title; one PR per phase milestone with the acceptance checklist pasted in the description (self-merge rules in §0.5.4).

---

## 13. Portfolio packaging ("a proper project to show")

- **README** (top-notch): logo, one-line pitch, 60-second GIF of the closed loop, badges (CI, Scorecard, release — **not an open-source license badge, per the 2026-11-10 all-rights-reserved relicensing, `docs/adr/0001`**), "Why", architecture diagram, feature table, quickstart (demo mode in 3 commands; lab mode link), safety model summary, roadmap, a note that the code is public for viewing but all rights reserved, and the **v1 → v2 story** (what was learned; honest limits).
- **Live demo** link (`/demo`) at the very top of the README.
- **Docs:** architecture, ADR index, threat model, safety invariants, outcome-classification methodology, scoring explainer.
- **Demo video** (60 to 90 s, no fluff) + 3 short blog posts.
- **Proof of rigor:** CI badges, test counts, Scorecard, signed releases, SBOM, a short "security assessment" write-up.
- **Be able to explain every module.** Reviewers will ask why the twin is "faithful", how outcomes are classified, and how the lab is isolated. Keep the docs honest about limits.
- Optional transparency note in the README about AI-assisted development; your call, but if you include it, keep it factual.

---

## 14. Risks and open questions

> In autonomous mode the owner-decision items below have defaults in §0.5.2; they remain listed so the reasoning is visible. Items needing the owner are routed to `docs/OWNER_TODO.md`.

| # | Item | Notes |
|---|---|---|
| 1 | **Licence** | **Resolved 2026-11-10**, owner instruction: All Rights Reserved, not MIT/Apache-2.0 — repo public, code not open source. See `docs/adr/0001-license-and-repo-identity.md`. |
| 2 | **Hardware** | v1 scan showed a 16 GB laptop with Hyper-V-capable setup. One twin is fine; Wazuh adds weight. Decide: Wazuh on same host (trimmed) vs a second box vs `native`-only lite mode for MVP. |
| 3 | **Windows images** | Evaluation media expire; licences constrain redistribution. Decide the supported image sources. |
| 4 | **Outcome classification accuracy** | Heuristic by nature. Validate on real runs, publish methodology and known false-positive/negative cases. |
| 5 | **CIS content licensing** | Avoid bundling CIS text; reference IDs; confirm terms. |
| 6 | **Website hosting + domain** | Vercel/Cloudflare Pages + a cheap domain. Plausible/Umami analytics. |
| 7 | **Caldera integration** | **Resolved 2026-11-10**, owner instruction: included in Phase 4's main build (not deferred), isolated + patched per `docs/adr/0011-caldera-in-scope.md`. |
| 8 | **Scope creep** | Hold the line: Phases 0 to 5 are the product; 6 to 8 make it shine; §11 is optional. |
| 9 | **Anthropic API cost** | Budgets/caps in config; Haiku-class for the public bot. |
| 10 | **Legal** | AUP, disclaimer, authorisation acknowledgement are product features, not afterthoughts. |

---

## Appendix A: Configuration (`.env.example`, placeholders only)

```dotenv
# --- Core ---
CD_ENV=dev                       # dev|prod
CD_BASE_URL=http://localhost:3000
CD_API_ADDR=:8080
CD_LAB_GATEWAY_ADDR=             # bind to the lab NIC only, e.g. 10.77.0.1:8443
CD_ALLOWED_ORIGINS=http://localhost:3000

# --- Data ---
CD_DATABASE_URL=postgres://cyberditto:CHANGE_ME@localhost:5432/cyberditto?sslmode=disable
CD_BLOB_BACKEND=localfs          # localfs|s3
CD_BLOB_PATH=./.data/blobs
CD_S3_ENDPOINT=                  # for MinIO/S3
CD_S3_BUCKET=
# Credentials come from the environment/secret manager, never this file in prod.

# --- Security ---
CD_SESSION_SECRET_FILE=./.secrets/session.key   # generated by `task bootstrap`; never committed
CD_ENVELOPE_KEY_FILE=./.secrets/envelope.key
CD_REQUIRE_PASSKEY_FOR_ADMIN=true

# --- Lab ---
CD_LAB_PROVIDER=fake             # fake|hyperv|libvirt|proxmox
CD_COMPLIANCE_ENGINE=native      # native|wazuh
CD_WAZUH_API_URL=
CD_WAZUH_INDEXER_URL=
CD_ATOMIC_CATALOG_COMMIT=        # pinned SHA (verified at import)
CD_ATTACK_DATA_VERSION=          # pinned ATT&CK release

# --- AI (all optional; features off if unset) ---
ANTHROPIC_API_KEY=               # from environment only
CD_AI_ENABLED=false
CD_AI_MODEL_CHAT=claude-sonnet-5-5       # verify IDs in current docs
CD_AI_MODEL_FAST=claude-haiku-5-5
CD_AI_MODEL_DEEP=claude-opus-5-5
CD_AI_MONTHLY_TOKEN_BUDGET=2000000
CD_AI_REDACTION=pseudonymise     # pseudonymise|strict|off(not allowed in prod)

# --- Public site ---
NEXT_PUBLIC_DEMO_MODE=false
TURNSTILE_SITE_KEY=
TURNSTILE_SECRET_KEY=
```

## Appendix B: Glossary

**Twin:** a VM rebuilt from a snapshot's configuration. **Snapshot:** immutable, signed config capture of a real asset. **Atomic:** a small, ATT&CK-mapped attack test. **Outcome:** blocked/detected/logged/missed classification of an executed test. **Gap:** a correlated set of failed controls and successful techniques with a priority. **Fidelity:** how closely the twin's re-collected config matches the source. **ApprovalRequest:** a human-gated action proposed by the copilot or system.

## Appendix C: Research references (checked 10 Nov 2026; re-verify when building)

- Invoke-AtomicRedTeam wiki: https://github.com/redcanaryco/invoke-atomicredteam/wiki
- Atomic Red Team MCP server overview: https://www.zscaler.com/mx/blogs/cybersecurity-best-practices/ai-security-testing and listing https://glama.ai/mcp/servers/s25xgh96zw
- Wazuh available SCA policies: https://documentation.wazuh.com/current/user-manual/capabilities/sec-config-assessment/available-sca-policies.html
- Wazuh Windows hardening + SCA example: https://wazuh.com/blog/automating-windows-endpoint-hardening-with-wazuh/
- MITRE Caldera: https://github.com/apache/caldera ; RCE advisory coverage: https://www.darkreading.com/application-security/max-severity-rce-vuln-all-versions-mitre-caldera
- HashiCorp BSL change and alternatives: https://cloudnativeengineer.substack.com/p/ep-8-hashicorps-open-source-exodus
- Go 1.26 release notes: https://go.dev/doc/go1.26 ; release history: https://go.dev/doc/devel/release
- Next.js 16 / React 19.2 / shadcn / TanStack stack overview: https://dev.to/abhaysinghr1/the-react-ai-stack-in-2026-what-you-should-actually-be-using-bkk
- Claude API: structured outputs https://platform.claude.com/docs/en/build-with-claude/structured-outputs ; MCP connector https://platform.claude.com/docs/en/agents-and-tools/mcp-connector ; Agent SDK https://platform.claude.com/docs/en/agent-sdk/mcp
- BAS/ATT&CK heatmap patterns (Picus datasheet): https://picussecurity.com/hubfs/Operationalize-MITRE-ATT%26CK-with-The-Picus-Platform-Datasheet.pdf
- Claude Code attribution settings: https://dev.classmethod.jp/en/articles/claude-code-attribution-setting/ ; https://pasqualepillitteri.it/en/news/27998/remove-claude-co-authored-by-commits-prs ; session-trailer issue https://github.com/anthropics/claude-code/issues/77830

*End of plan.*
