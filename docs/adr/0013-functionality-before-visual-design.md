# 0013: Build all functionality first, then a dedicated UI design phase

- **Status:** accepted
- **Date:** 2026-11-23

## Context

MASTER_PLAN §8 describes a polished console (design tokens, shadcn/ui, command palette, motion, ATT&CK heatmap visuals) and §10 spreads that UI work across Phases 1–7, with a "design polish pass" in Phase 7. The console that exists after Phase 1 is deliberately plain: functional pages styled with a few Tailwind utilities. The owner asked to keep it that way until every functional phase is done, then do the visual design as its own phase rather than interleaving it.

## Decision

- Phases 1–6 build **functional** console pages only: correct data, correct states (loading/empty/error/success), correct authZ behaviour, keyboard-usable, no layout breakage. No design-system work, no animation, no visual polish beyond legibility.
- A new **UI phase** runs after Phase 6 (all product functionality done) and before Phase 7 (marketing site, docs, demo). It owns everything in MASTER_PLAN §8.1–§8.5: `packages/ui` design system (ADR 0010), tokens, the console information architecture in §8.3, the ATT&CK heatmap and other rich visualisations, command palette, light/dark themes, motion, and the accessibility audit.
- Phase 7 keeps the marketing site, docs site and public demo, built on top of the finished design system instead of inventing one.

## Consequences

- Plain-looking console pages during Phases 1–6 are intentional, not regressions; reviews of those phases judge behaviour, not looks.
- Functional pages should still keep data-fetching and state logic separate from markup (`lib/api.ts`, small components) so the UI phase can restyle without rewriting behaviour.

## Rejected alternatives

- Polishing each page as it's built (the original §10 interleaving) — rejected per owner instruction; it spends design effort on screens whose data and flows are still changing.
