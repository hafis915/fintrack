---
title: "Fintrack — Decision Log"
type: project-doc
created: 2026-06-03
last_updated: 2026-06-03 (added ADR-008–013 — storage + Tailwind/shadcn-vue + vibe coding + design system)
tags: [project, decisions, adr, fintrack]
related:
  - "[[(C) PROJECT.md]]"
  - "[[(C) ARCHITECTURE.md]]"
  - "[[(C) ROADMAP.md]]"
---

## How to Read This

ADR-style log. One entry per decision. Don't edit past decisions — add new ones that supersede them. This is your memory when future-Hafis asks "why did we pick X?"

---

## ADR-001 — Project framing: personal-first, public-later (2026-06-03)

**Decision:** Build Fintrack as a personal app where Hafis is user #1. Open to public only after personal validation (30 days of daily use + success criteria met).

**Why:**
- Removes pressure to solve monetization, marketing, regulatory questions now
- Forces Hafis to "eat his own dog food" before exposing it to anyone
- Validates the product actually solves a problem Hafis himself has

**Trade-off:** Slower path to public launch. Acceptable because there is no deadline (see ADR-003).

**Supersedes:** Original framing in PRD which implied SaaS-from-day-one.

---

## ADR-002 — Stack chosen for learning, not shipping speed (2026-06-03)

**Decision:** Go + Echo + sqlc + pgx backend. Vue 3 + Vite frontend.

**Why:**
- **Go for learning:** Hafis is strongest in TypeScript/Node. Go is the stretch language. Cost: slower iteration. Benefit: career skill + learning production patterns in a typed compiled language.
- **Vue 3 + Vite (not Nuxt):** Authenticated PWA has no SEO need → SSR is wasted complexity. Vite gives PWA via plugin with minimal framework magic.
- **sqlc over ORM:** Financial accuracy requires explicit SQL. Audit trail visible. Type-safe Go output.

**Trade-off:** Slower MVP than if Hafis used Node + Nuxt (his strongest stack). Acceptable because learning is the primary goal.

---

## ADR-003 — No deadline (2026-06-03)

**Decision:** Fintrack has no launch date. Effort estimates in roadmap are guidance, not commitments.

**Why:**
- Hafis self-identified weakness: "rushes under stress" trades precision for speed
- Personal project + learning goal makes deadline pressure counterproductive
- The cost of slipping is zero (no users waiting)

**Risk:** Project becomes perpetual planning, never ships. **Mitigation:** Weekly review of [[(C) PROJECT.md]] status log. Honest check-in: did I write code this week, or just plan?

---

## ADR-004 — MVP cut from 6 features to 4 (2026-06-03)

**Decision:** MVP = Onboarding + Transactions CRUD + Receipt Scan + Fatigue Dashboard.

**Cut from MVP:**
- Emergency Fund Tracker → v2 (just another dashboard widget, no new learning)
- Weekly Narrative Summary → v2 (separate cron + email + LLM system)
- Debt Tracker → v2 (separate domain, snowball/avalanche algorithms)

**Why:** 6 features × solo evenings × Go-learning-curve = scope death. 4 features is still ambitious but achievable.

**Added (implicit):** Transactions CRUD with manual entry fallback — original plan had this implicit, now explicit because AI categorization will fail sometimes.

---

## ADR-005 — Goal-First Onboarding kept in MVP (2026-06-03)

**Decision:** Keep Goal-First Onboarding in MVP despite the 2-3 week cost.

**Why:**
- It is the strategic differentiation vs. Money Lover, Wallet, Monefy (which are tracking-first)
- Cutting it would remove the moat from the product
- The algorithm (budget generation per program) is real engineering work worth learning

**Implementation note:** Build the onboarding fully, but bypass it for Hafis's own account via a SQL seed migration. Onboarding is built as a product feature, not as a daily workflow.

---

## ADR-006 — Vue 3 + Vite over Nuxt 3 (2026-06-03)

**Decision:** Vue 3 + Vite + vite-plugin-pwa for frontend.

**Why:**
- Fintrack is authenticated-only → SSR wasted
- Vite + plugin gives full PWA capabilities (install, offline shell, SW)
- Less framework magic → faster to debug for solo dev
- Pinia for state, Vue Router 4, VeeValidate + Zod for forms

**Rejected alternatives:**
- Nuxt 3 (over-engineered for this use case)
- React (Hafis knows it but Vue is also in stack and slightly preferred here)

---

## ADR-007 — Supabase for Auth + DB (2026-06-03)

**Decision:** Use Supabase for both PostgreSQL hosting and Auth.

**Why:**
- Free tier covers personal use
- RLS lets data isolation live at DB layer, not app layer
- Auth UI primitives + JWT issuance handled — Go just validates
- Supabase Storage available for receipt images

**Trade-off:** Vendor lock-in. Acceptable because exit strategy is just "export Postgres dump + reimplement auth" — both are standard.

---

---

## ADR-008 — Object storage layer included in MVP for learning (2026-06-03)

**Decision:** Build a `Storage` interface + S3-compatible implementation in MVP, even though MVP feature requirements alone don't strictly need it.

**Why:**
- Hafis's primary goal for Fintrack is *learning production patterns*, not shipping minimal scope
- DI + storage layer + bucket conventions + signed URLs are real production skills worth practicing
- Future v2 features (receipt history, retry-on-failure, cleanup worker) will all need this layer — building it now means less refactoring later
- **Honesty clause:** This is *not* justified by "we need storage for resilience" — that would be a v2 reason. It is justified by learning intent, named explicitly.

**Trade-off:** Adds ~3-5 days of work to Phase 3 vs. discarding images in-memory. Acceptable given no deadline (ADR-003) and learning goal (ADR-002).

**Alternatives considered:**
- Discard images after Claude Vision parses them (simpler, but no learning of storage layer)
- Defer storage entirely to v2 (cleaner MVP, but loses the learning Hafis wants now)

---

## ADR-009 — Single S3-compatible storage impl, MinIO local + Supabase prod (2026-06-03)

**Decision:** Use `minio-go` SDK with one `S3Storage` implementation. Local dev uses MinIO via Docker Compose. Production uses Supabase Storage's S3-compatible endpoint. Environment selection via `STORAGE_ENDPOINT` env var only.

**Why:**
- Both MinIO and Supabase Storage speak S3 API → one client works against both
- Identical code path in dev and prod → fewer "works on my machine" bugs
- Future swap to Cloudflare R2 / Backblaze B2 / DigitalOcean Spaces is also S3-compatible — zero code change
- Still preserves `Storage` interface (ADR-008) for future non-S3 backends (e.g., local filesystem for offline-first features)

**Trade-off:** Slight coupling to S3 API surface. Acceptable because S3 is the de facto standard and migration cost is low if a non-S3 backend is ever needed.

**Alternatives considered:**
- Two separate impls (`MinioStorage` + `SupabaseStorage`) — rejected as needless duplication
- Use Supabase Storage SDK directly — rejected because it locks out MinIO and other S3-compatible options

---

## ADR-010 — Storage cleanup, retry logic, and receipt history deferred to v2 (2026-06-03)

**Decision:** MVP storage layer only implements the **happy path** (`Upload` on receipt scan). The following are deferred to v2:

- **Cleanup worker:** Orphaned uploads (snapped but never confirmed) live forever in storage during MVP
- **Retry-on-Claude-API-failure:** If Claude Vision errors, frontend re-uploads from browser memory (not from stored image)
- **User-facing receipt history:** No "view past receipt" UI in MVP

**Why:**
- Hafis is sole user → storage cost is trivial (~30 receipts/mo × 500KB = 15MB/mo)
- Cleanup worker requires `apps/worker/main.go`, scheduled job runner, deletion query — entire vertical slice of work that doesn't aid the hero feature
- Retry resilience is a v2 problem (ADR-008 named learning, not resilience, as the reason)
- Receipt history is a public-user feature, not a personal-Hafis feature

**Trade-off:** Storage accumulates orphans during MVP. Acceptable at solo scale. Worker added in v2 before public launch.

**Risk:** Forgetting that cleanup is missing → bill creep when going public. **Mitigation:** Listed explicitly in v2 backlog ([[(C) ROADMAP.md]]) so it's not forgotten.

---

---

## ADR-011 — Tailwind CSS + shadcn-vue for frontend styling (2026-06-03)

**Decision:** Use Tailwind CSS for styling and shadcn-vue (community Vue port of shadcn/ui) for components.

**Why:**
- Tailwind utility classes accelerate iteration vs. writing custom CSS
- Design tokens (colors, spacing, radius) live in `tailwind.config.ts` — single source of truth
- shadcn-vue components are *copy-pasted into the repo*, not npm-installed → owned, customizable, no version conflicts
- Matches modern Vue 3 + Vite + TS ecosystem patterns
- Aligned with "vibe coding" mode (ADR-012): generated Tailwind code is easy to review at a glance

**Trade-off:** Less learning of vanilla CSS / design system internals. Acceptable because the learning goal is vibe coding orchestration, not CSS mastery.

**Alternatives considered:**
- PrimeVue (rejected — heavy, npm-installed, less customizable)
- Vuetify (rejected — Material Design opinions don't fit "Gym app for your money" aesthetic)
- Build components from scratch (rejected — slows MVP without proportional learning value)

**Supersedes:** Original CLAUDE.md draft that listed "PrimeVue (TBD)" as the UI library.

---

## ADR-012 — Vibe coding as the primary operating mode (2026-06-03)

**Decision:** Claude implements features end-to-end. Hafis writes prompts, scope, constraints, and reviews all generated code before acceptance.

**Why:**
- Original framing of project (per Hafis): *"vibe code skill how I plan and build the production-ready app"*
- The skill being developed is **orchestration + review**, not raw Go syntax muscle memory
- Vibe coding lets Hafis ship more ambitious scope solo than hand-coding would
- Code review is itself a production skill (PR reviews at work, AI-generated code at scale)

**Trade-off:**
- Hafis won't develop Go writing muscle memory on this project
- Risk: dependence on AI for future debugging if AI isn't available
- Risk: shallow understanding of generated code if review is skimmed

**Mitigations:**
- CLAUDE.md mandates "explain why" for every generated piece
- Architecture principles + conventions in CLAUDE.md keep generated code consistent and reviewable
- ADRs document every non-obvious decision so future-Hafis can understand intent
- Hafis can switch off vibe mode for any subsystem he wants to write by hand (e.g., the fatigue calculation logic)

**Supersedes:** Initial CLAUDE.md draft that mandated "don't auto-generate entire features."

---

---

## ADR-013 — Design system: Performance Utility, mobile-first, dynamic (2026-06-03)

**Decision:** Adopt the design system documented in `(C) DESIGN.md`. Key choices:

- **Memorable thing:** "Money discipline that feels like training, not bookkeeping."
- **Aesthetic:** Industrial Utilitarian with editorial accents
- **Reference apps:** Whoop, Strava, Apple Fitness, Strong/Hevy (NOT fintech apps)
- **Brand accent:** Saffron Gold (#F4A300 dark / #D97706 light)
- **Hero typography:** JetBrains Mono for numbers, General Sans for display, DM Sans for body
- **Hero number treatment:** Mono digits + saffron "Rp" prefix + muted decimals (typographic hierarchy within numbers)
- **Theme:** Dark default, light opt-in (both supported)
- **Motion:** Dynamic — count-up animations, state transitions, signature scan-flow choreography
- **Layout:** Mobile-first single column max-width 420px, bottom tab nav with raised FAB for Scan

**Why:**
- Performance/training aesthetic is the strongest differentiation from generic fintech
- Mobile-first matches actual usage (phone-based, on-the-go financial decisions)
- Dynamic motion is the "alive" feel Hafis specifically requested
- Saffron has Indonesian cultural resonance + avoids the green/blue fintech default
- Numbers-as-designed-objects (vs just "big white numbers") avoids AI-coded aesthetic

**Trade-off:**
- Dark mode default may surprise users expecting light finance apps (mitigation: light mode toggle in MVP)
- JetBrains Mono + dark mode + accent color *combination* risks reading as AI-generated (mitigation: typographic hierarchy treatment elevates beyond default)
- Both themes means ~1-2 extra days work in Phase 2

**Alternatives considered:**
- Direction B (Coach-led, editorial typography front and center) — rejected as secondary
- Direction C (Indonesian visual culture-heavy) — rejected as too narrow; some cultural cues retained (saffron color, IDR-first formatting)
- General Sans for hero numbers (safer) — rejected for being less distinctive
- Light mode only — rejected, dark default better matches "performance" framing
- Dark mode only — rejected per Hafis preference for user choice

**Supersedes:** No prior design decisions. This is the first design ADR.

---

## Decision Template (for future entries)

```
## ADR-XXX — [Title] (YYYY-MM-DD)

**Decision:** What did we decide?

**Why:**
- Reason 1
- Reason 2

**Trade-off:** What does this cost us?

**Alternatives considered:** What did we reject?

**Supersedes:** [Reference earlier ADRs if applicable]
```

## Connections

- [[(C) PROJECT.md]] — Project overview
- [[(C) ARCHITECTURE.md]] — Implementation of these decisions
- [[(C) ROADMAP.md]] — When each decision gets executed
