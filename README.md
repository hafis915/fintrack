# Fintrack

> "Gym app for your money" — Indonesian-native personal finance PWA targeting
> fresh workers. AI scans receipts, categorizes spending, warns when a category
> is "fatigued," and writes a weekly narrative in Bahasa Indonesia.

**Status:** Pre-implementation. Phase 0 (repo + planning artifacts) is done.
Phase 1 (Go backend skeleton) starts next. See the planning trail below for the
full roadmap.

## Planning trail

This project is built solo with Claude Code (Opus 4.7) as the implementation
driver. The planning artifacts are committed to the repo so the workflow is
visible alongside the code.

- [`full_doc.html`](./full_doc.html) — Product & technical PRD (v1.0, Apr 2026).
  Indonesian-native, includes full DB schema, 25 REST endpoints, user stories.
- [`docs/superpowers/plans/2026-04-29-backend-mvp.md`](./docs/superpowers/plans/2026-04-29-backend-mvp.md) —
  16-phase Go backend implementation plan, ~3,400 lines of TDD task breakdowns.
- [`docs/superpowers/designs/2026-06-01-fintrack-portfolio-sequencing.md`](./docs/superpowers/designs/2026-06-01-fintrack-portfolio-sequencing.md) —
  Office-hours design doc. Frames the project as a portfolio piece showcasing
  AI-orchestrated solo development; chooses vertical-slice execution order.
- [`docs/superpowers/reviews/reviews/2026-06-01-ceo-review.md`](./docs/superpowers/reviews/reviews/2026-06-01-ceo-review.md) —
  CEO-mode plan review. Surfaces three gaps between the plan and the design doc;
  locks in the inserts now applied to the plan file.

## Tech stack (planned)

Go 1.22 · Echo v4 · pgx/v5 · sqlc · golang-migrate · Postgres (Supabase) ·
Anthropic Claude API · Railway · Next.js 14 · Vercel.

## Roadmap (milestones, not phase numbers)

1. **Milestone 1** — Foundation + Railway deploy. Live URL with `/health`.
2. **Milestone 2** — Onboarding + budget engine vertical slice.
3. **Milestone 3** — Receipt scan + transactions + fatigue vertical slice.
4. **Milestone 4** — Weekly narrative worker.
5. **Milestone 5** — Debts, goals, reports, BYOA tokens (horizontal fill).
6. **Milestone 6** — Thin Next.js frontend + portfolio-grade README + Loom demo.

This README will be rewritten in Milestone 6 with live URLs, architecture
diagrams, and the embedded walkthrough. For now its only job is to explain
that the repo isn't empty by accident.
