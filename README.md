# Fintrack

> **"Gym app for your money."** Personal finance PWA where money discipline feels like training, not bookkeeping.

Indonesian-native mobile-first PWA. Currently in pre-development. Built solo by [@hafis915](https://github.com/hafis915) as a vibe-coding learning project — the goal is to practice planning, directing, and reviewing AI-generated code on a production-grade app.

---

## Status

**Phase 0 done (2026-06-03).** Repo skeleton + hello world:

- Go backend on Echo with `/health` (public) and `/v1/me` (JWT-protected)
- Vue 3 + Vite + Tailwind frontend rendering `/health` status
- Postgres + MinIO via `docker compose`
- `cmd/mint-jwt` for local-only token issuance ([ADR-014](./DECISIONS.md))

Prior pre-vault-reset commit history is preserved under the `pre-vault-reset` tag.

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

Requires Go 1.22+, Node 20+, Docker, `sqlc`, and `golang-migrate`. No external accounts during MVP (see [ADR-014](./DECISIONS.md)).

```bash
# 1. secrets
cp .env.example .env
# replace JWT_SECRET + INCOME_ENCRYPTION_KEY with: openssl rand -hex 32

# 2. dependencies (postgres on :55432 to avoid clashing with host postgres)
make up
make migrate

# 3. run
make api       # backend on :8080
make web       # frontend on :5173 (in another shell)

# 4. mint a test JWT and hit a protected route
TOKEN=$(make token | tail -1)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/me
```

See `make help` for the full target list.

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
