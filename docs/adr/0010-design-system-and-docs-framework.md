# 0010: Design system approach and docs framework

- **Status:** accepted
- **Date:** 2026-11-10

## Context

v2 needs a cohesive design system (tokens, components) shared between the console and the marketing site, plus a docs site with an API reference generated from OpenAPI (MASTER_PLAN §8, §8.6).

## Decision

- Design system lives in `packages/ui`: design tokens (dark-first, per §8.1) plus components built on **shadcn/ui + Radix**, documented with Storybook (or Ladle — final pick deferred to Phase 7 when visual-regression tooling is wired up) and visual-regression via Playwright screenshots.
- Docs framework: **Fumadocs** (MDX-based, Next.js-native) is the default per MASTER_PLAN §0.5.2, used unless a blocking problem appears during Phase 7 — if so, record the switch (e.g. to Nextra) as an amendment to this ADR with the reason.
- Both the marketing site and the console live in the single `apps/web` Next.js app, using route groups (`(marketing)`, `(docs)`, `(console)`, `/demo`) rather than separate deployables, so the design system and generated API client are shared.

## Rejected alternatives

- A separate docs repo/subdomain stack — unnecessary split for the current scope; one Next.js app with route groups covers it.
