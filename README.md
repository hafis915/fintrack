# Fintrack

> **"Gym app for your money."** Personal finance PWA where money discipline feels like training, not bookkeeping.

Indonesian-native mobile-first PWA. Currently in pre-development. Built solo by [@hafis915](https://github.com/hafis915) as a vibe-coding learning project — the goal is to practice planning, directing, and reviewing AI-generated code on a production-grade app.

---

## Status

**Pre-development.** Planning artifacts complete, no code yet. Phase 0 (repo skeleton, hello world) is next.

This repo was reset on 2026-06-03 to align with the planning trail. Prior commit history is preserved under the `pre-vault-reset` tag.

## What This Is

- **Product:** Personal finance app — track spending, receipt-scan via AI, "category fatigue" dashboard, coaching narrative
- **User #1:** Hafis. Public launch deferred until personal validation succeeds (30 days daily use + success criteria).
- **Stack:** Go + Echo + sqlc + Supabase (backend), Vue 3 + Vite + Tailwind + shadcn-vue (frontend)
- **AI:** Claude Vision for receipt categorization (Claude Haiku for narrative reports in v2)

## MVP (4 features)

1. **Goal-First Onboarding** — 6-question intake → personalized budget + program selection
2. **Transactions CRUD + manual entry** — foundation
3. **Receipt Photo Categorization** — Claude Vision hero feature
4. **Category Fatigue Dashboard** — Fresh / Warning / Fatigued indicators

Deferred to v2: emergency fund, weekly narrative, debt tracker, BYOA agent tokens.

## Documentation

All planning docs live at repo root:

- **[CLAUDE.md](./CLAUDE.md)** — Project context for AI coding agents. Read first.
- **[DESIGN.md](./DESIGN.md)** — Design system. Typography, color, motion. Read before any UI work.
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** — Tech stack, repo structure, data flows.
- **[DECISIONS.md](./DECISIONS.md)** — ADR log. Every non-obvious decision with rationale.
- **[docs/prd.html](./docs/prd.html)** — Original product & technical PRD (v1.0, April 2026).

Additional planning lives in the author's Obsidian vault (`Hafis-Brain/03 Projects/Fintrack/`) — status logs, roadmap checkboxes, weekly reviews.

## Local Development

Coming in Phase 0. The stack will require:

- Go 1.22+
- Node 20+ (for the Vue frontend)
- Docker (for local MinIO storage)
- Supabase account (DB + Auth)
- Anthropic API key (Claude Vision)

Setup instructions will be added once Phase 0 is complete.

## Goals & Non-Goals

**Goals:**
- Learn vibe-coding: plan, direct AI, review production-grade code solo
- Ship a working app Hafis uses daily for 30+ consecutive days
- Build something usable enough to validate the goal-first / fatigue dashboard concept

**Non-goals (for v1):**
- Public launch / marketing
- Monetization
- OJK regulatory compliance
- Multi-tenant scaling beyond user #1

## License

Not yet decided. Project is currently private use only.

---

*Built solo with Claude (Sonnet 4.5) as the implementation partner. Every commit is human-reviewed before merge.*
