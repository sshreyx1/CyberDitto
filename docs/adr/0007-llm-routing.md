# 0007: LLM integration and model routing

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v1 had no AI layer. v2's "Ditto Copilot" (MASTER_PLAN §7) must be grounded in the user's own evidence, never autonomous over safety invariants, and off by default so the app is fully functional without an API key.

## Decision

- All LLM orchestration lives server-side in the Go control plane (`app/copilot`) via the Anthropic Go SDK; the browser never sees the API key.
- Model routing is config-driven (env vars `CD_AI_MODEL_FAST` / `CD_AI_MODEL_CHAT` / `CD_AI_MODEL_DEEP`), not hardcoded — exact model IDs are verified against current Anthropic docs at the time Phase 6 is built, per MASTER_PLAN §0.3's version policy.
- Tools are split into read-only (auto-approved) and mutating (always produce an `ApprovalRequest`, executed only after a human click) — the model can propose but never directly execute a run, remediation apply, report send, or twin build.
- `CD_AI_ENABLED=false` by default; all AI features degrade to "not available" rather than the app failing.
- Prompt caching for the static system prompt/tool schemas; SSE streaming to the UI.

## Rejected alternatives

- Giving the model a generic "run shell command" or "apply plan" tool with no approval step — rejected; violates MASTER_PLAN §6.4/§7.4 (no autonomy over safety invariants).
