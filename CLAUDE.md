<!--
  Master copy lives here in Hafis-Brain vault.
  Copy to the Fintrack repo as `CLAUDE.md` (no prefix) when syncing.
  Update here first, then sync. Repo file is a working copy, not the source of truth.
-->

# Fintrack — Project Context for Claude Code

> Personal finance PWA. Goal: learn full-cycle production app development solo (Go + Vue).
> Hafis is the developer AND the only user (v1). Public launch deferred until personal validation.

---

## Who You're Working With

- **Developer:** Hafis (Muh Hafidz Tafsani Hamty)
- **Background:** 4+ years fullstack — TypeScript / React / Vue / Node strong; **Go is the stretch language for this project**
- **Address as:** Hafis
- **Communication style:** Direct, bulleted, no fluff. Challenge ideas, don't just validate.

---

## Project Identity

- **Product:** Fintrack — "Gym app for your money"
- **Target user:** Hafis himself (v1). Indonesian fresh workers (Rp 8–10jt/mo income) when public.
- **Type:** PWA, mobile-first, B2C
- **Source PRD:** https://github.com/hafis915/fintrack/blob/main/full_doc.html
- **Status:** Pre-development

---

## MVP Scope (v1 — 4 features)

**In scope:**

1. **Goal-First Onboarding** — 6-question intake → generates personalized budget + program selection (5 programs: Pondasi, Bebas Utang, Goal Chaser, Tumbuh, Seimbang)
2. **Transactions CRUD + manual entry** — foundation; users can log transactions manually with/without receipt
3. **Receipt Photo Categorization** — Claude Vision extracts amount/merchant/category from photo, user confirms
4. **Category Fatigue Dashboard** — Fresh / Warning / Fatigued indicators per category with coaching language

**Out of scope (deferred to v2):**

- Emergency Fund Tracker
- Weekly Narrative Summary (cron + email + LLM reports)
- Debt Tracker (Snowball / Avalanche)
- Progressive Saving Rate suggestions
- BYOA / Agent API tokens
- Storage cleanup worker
- Retry-on-Claude-API-failure logic
- User-facing receipt history UI

**MVP success criteria:**
- Hafis uses Fintrack daily for 30 consecutive days
- ≥80% of transactions logged via receipt scan
- Fatigue dashboard surfaces a "Warning" category before the 20th of the month
- Deployed to production infra (Railway + Vercel/CF Pages)

**If a feature isn't on this list, push back when asked to build it.** Scope creep kills solo projects.

### v1 scope additions (accepted 2026-06-04)

These emerged during the build cycle and were approved by Hafis as the intended product flow:
**login → onboarding → result/budget → compare spending vs budget (with graphic) → reduction recommendations.**

- **Local email register/login UI + route guard** — Phase-0 local auth only. Real Supabase Auth stays deferred to v2.
- **Beranda redesigned into a real dashboard** — snapshot + quick actions. The old API-status/health card is dev-only and removed from the user-facing home.
- **Budget vs actual** — the budget dashboard visually compares spending to the plan (a graphic) and surfaces reduction recommendations ("what to cut"). This is an extension of the **Category Fatigue Dashboard** feature, not a new standalone feature.

### v1 scope additions (accepted 2026-06-04, batch 2)

Approved by Hafis; design + rationale recorded in `PLAN.md` (ADR-2026-06-04).

- **Auth hardening** — register/login now require a **bcrypt password** (≥8 chars); generic 401 on bad credentials (no email-existence leak); **logout** ("Keluar") on the beranda; router guard validates JWT `exp` client-side.
- **Responsive desktop layer** — on `≥lg`, a left sidebar nav replaces the bottom tabs and content widens (multi-column where useful). Mobile-first is unchanged. (Departs from the strict "max-width 420px" rule → see Design System note + `PLAN.md`.)
- **Reports page (`/reports`)** — desktop-optimized spending report: **month filter**, spending-by-category table + chart, and **CSV export** of the selected month's transactions.
- **Month filter** on the Transactions list (date-range, via the existing `from`/`to` API params).

### v1 scope additions (accepted 2026-06-04, batch 3 — session 2)

Approved by Hafis; design + rationale recorded in `PLAN.md` (ADR-2026-06-04b). **Built, verified green
(`make test` / `make test-e2e` 37/37), and committed to `main` (3 commits, not yet pushed) — see the
"NEXT-SESSION STATUS" banner in `PLAN.md`.**

- **Conversational planner onboarding** — onboarding is a 3-step wizard: 6 questions → **fixed expenses
  only** → app **auto-suggests flexible (keinginan) amounts** + a savings target, refine by editing a
  number OR **chatting** with a planner. Money math is deterministic in Go (`internal/domain/budget`);
  the LLM (`internal/llm`, OpenRouter, stub when `OPEN_ROUTER_API_KEY` empty) is a **language-only**
  layer. New endpoints: `POST /v1/onboarding/suggest`, `POST /v1/planner/chat`.
- **Inline custom-category creation** — `AddCategoryInline.vue` in onboarding step 2 and the
  add-transaction form (so a user can add a category that isn't in the catalog). Uses existing
  `POST /v1/categories`.
- **Re-budget entry point** — "↻ Atur ulang budget" on the Budget dashboard re-runs the planner
  (finalize is idempotent, so it replaces the current-month plan).
- **Neo-brutalist light theme** — see the Design System note below.
- **Desktop UX pass** — Home/Budget/Transactions go multi-column at `≥lg`; **logout moved into the
  sidebar navbar** (home-header `Keluar` is now mobile-only); CTAs + row Edit/Hapus buttons made obvious.

---

## Hafis's Learning Goals (PRIMARY)

This project exists to **teach Hafis the vibe coding skill** — how to plan, orchestrate AI, and ship production-ready apps solo as the *director*, not as the line-by-line writer.

**Operating mode: VIBE CODING**
- **Claude implements features end-to-end.** Hafis reviews, requests changes, approves.
- **Hafis writes the prompts, the scope, the constraints. Claude writes the code.**
- The skills Hafis is developing: scoping, reviewing AI output, spotting AI mistakes, architecting at a high level, setting guardrails (like this file).
- The skills Hafis is *deliberately not* developing here: writing Go from muscle memory, raw debugging without AI assistance.

**Implications for how you help:**
- **Auto-generate whole features when asked.** Don't gatekeep with "let me explain first, then you write it."
- **Surface the "why" behind decisions** in plain language so Hafis can review intelligently.
- **Push back when Hafis takes shortcuts** that compromise production quality (security, data integrity, error handling).
- **No magic abstractions.** Avoid Wire / Fx / heavy frameworks — they make the generated code harder to *review*, which defeats the purpose.
- **Always show what you generated and why** — Hafis reviews everything before it's accepted.

---

## Tech Stack

### Backend (Go)
| Layer | Choice |
|-------|--------|
| Language | Go (Golang) |
| HTTP framework | Echo v4 |
| DB | PostgreSQL via Supabase |
| DB access | sqlc (type-safe Go from explicit SQL) |
| DB driver | pgx/v5 |
| Migrations | golang-migrate (numbered `.sql` files) |
| Auth | Supabase Auth → JWT validated by `golang-jwt/jwt` |
| Encryption | `crypto/aes` (stdlib) — AES-256-GCM for income |
| AI | Claude Vision + Haiku via `net/http` (no SDK) |
| Object storage | S3-compatible via `minio-go` (MinIO local, Supabase Storage prod) |
| Config | viper |
| Validation | go-playground/validator |
| Logging | rs/zerolog (structured JSON) |
| IDs | google/uuid (v4) |
| Testing | stdlib + testify + sqlmock |

### Frontend (Vue)
| Layer | Choice |
|-------|--------|
| Framework | Vue 3 |
| Build | Vite |
| PWA | vite-plugin-pwa |
| Router | Vue Router 4 |
| State | Pinia |
| Forms | VeeValidate + Zod |
| HTTP | Axios |
| Styling | Tailwind CSS |
| Component library | shadcn-vue (community port of shadcn/ui — copy-paste components into repo, not npm-installed) |
| Image compression | browser-image-compression |

### Infra
| Layer | Choice |
|-------|--------|
| Backend host | Railway (TBD confirm) |
| Frontend host | Vercel / Cloudflare Pages (TBD pick) |
| Local containers | Docker Compose (for MinIO) |
| CI/CD | GitHub Actions (TBD) |

---

## Architecture Principles

1. **Manual dependency injection.** Wire from `main.go` downward. Domain exposes interfaces only. **No Wire / Fx.**
2. **Explicit SQL, never ORM.** Financial accuracy requires hand-written queries. Use sqlc.
3. **User isolation is app-layer and TESTED (RLS is not a backstop yet).** Every query filters `user_id` (the `...ForUser` sqlc naming convention); `test/integration/isolation_integration_test.go` proves user B cannot read/edit/delete user A's rows. ⚠️ RLS is `ENABLE`d on every table but has **no policies** and the app connects as the table owner, so RLS is currently a **no-op** — it is NOT the isolation guarantee. Real policy-backed RLS (policies + `FORCE ROW LEVEL SECURITY` + a non-owner role that `SET LOCAL app.user_id`) is a deferred pre-launch hardening (see CSO audit 2026-06-05 / `PLAN.md`). Until then, **any new query MUST scope by `user_id`**, and any new by-id route MUST be covered by the isolation test.
4. **Repository pattern.** Handlers never touch sqlc directly — go through `internal/repository/`.
5. **Domain layer is pure Go.** No HTTP, no DB, no Echo. Just business logic + interfaces.
6. **Errors are values.** Use sentinel errors + `errors.Is` / `errors.As`. No panics in normal flow.
7. **Validation at the boundary.** Validate request payloads in handlers before they reach domain.

---

## Repository Structure

```
fintrack/
├── apps/
│   ├── api/main.go              # HTTP server entry
│   └── worker/main.go           # Background jobs (v2)
├── internal/
│   ├── config/                  # viper env loading
│   ├── server/                  # Echo init, middleware wiring
│   ├── middleware/              # auth, logging, rate limit, body size
│   ├── domain/                  # Business logic + interfaces (pure Go)
│   │   ├── user/
│   │   ├── budget/
│   │   ├── transaction/
│   │   └── fatigue/
│   ├── handler/                 # HTTP request/response
│   ├── repository/              # DB access wrapper over sqlc
│   ├── ai/                      # Claude Vision client
│   ├── storage/                 # S3-compatible storage interface + S3Storage impl
│   └── encryption/              # AES-256-GCM income encryption
├── database/
│   ├── migrations/              # 0001_init.up.sql, 0001_init.down.sql, ...
│   └── sqlc/
│       ├── sqlc.yaml
│       ├── query/               # Hand-written SQL
│       └── generated/           # sqlc output — NEVER edit manually
├── pkg/                         # Public utils (errors, responses, logger)
├── web/                         # Vue 3 + Vite frontend
│   ├── src/
│   │   ├── components/
│   │   ├── views/
│   │   ├── stores/              # Pinia
│   │   ├── router/
│   │   └── api/                 # HTTP client wrappers
│   └── vite.config.ts
├── docker-compose.yml           # MinIO for local dev
└── CLAUDE.md                    # This file
```

---

## Backend Conventions (Go)

- **Package names:** lowercase, single word (`user`, `budget`, not `userService`)
- **File names:** `snake_case.go`
- **Interfaces in domain:** Declared in `internal/domain/<x>/<x>.go`. Implementations elsewhere.
- **Error wrapping:** Always `fmt.Errorf("doing thing: %w", err)`. Never `%v` for errors.
- **Context:** First parameter on every function that crosses I/O. Never store in struct.
- **No init functions** — explicit setup only.
- **Test files:** `_test.go` co-located with code. Table-driven tests preferred.
- **sqlc workflow:** Edit `database/sqlc/query/*.sql` → run `sqlc generate` → use generated code. **Never hand-edit generated files.**
- **Migration naming:** `NNNN_description.up.sql` + `NNNN_description.down.sql`. Both required.

---

## Frontend Conventions (Vue)

- **Components:** `PascalCase.vue` — single-file components with `<script setup lang="ts">`
- **Composition API only** — no Options API
- **TypeScript everywhere** — no plain JS in `web/src/`
- **Pinia stores:** `useXxxStore` naming. One store per domain (user, transactions, budget, fatigue)
- **API calls:** Centralized in `web/src/api/<domain>.ts`. Components never call HTTP directly.
- **Form validation:** Zod schema → VeeValidate. Same schema shape as Go validation tags where possible.
- **Routing:** Named routes. Lazy-load route components.
- **Styling:** Tailwind CSS utility classes. **No inline `style=""`** except for dynamic values that can't be expressed as classes.
- **Components:** Prefer shadcn-vue copy-pasted into `web/src/components/ui/` over external dependencies. Customize freely — components are owned, not imported.
- **Design tokens:** Configure colors/spacing/radius in `tailwind.config.ts`. No hardcoded hex values in components.

---

## Database Conventions

- **All primary keys:** UUID v4 (`uuid_generate_v4()`)
- **All tables:** `user_id` column, and every query MUST filter on it (system tables like `expense_categories` defaults use `user_id is null`)
- **Money:** Stored as `BIGINT` (Rupiah, no decimals). Never `FLOAT` / `DECIMAL` for currency.
- **Timestamps:** `TIMESTAMPTZ` always. Use `now()` default.
- **Soft delete:** Only where needed (transactions yes, categories no).
- **Indexes:** Add explicitly per query pattern. Document in migration comment.
- **RLS:** Enabled on every table but currently un-policied (a no-op behind the owner connection) — NOT a backstop today. Isolation is app-layer + covered by `isolation_integration_test.go`. Adding real policies (+ `FORCE` + non-owner role + `SET LOCAL app.user_id`) is a deferred pre-launch gate.

---

## Security & Privacy Constraints

- **Income encryption:** AES-256-GCM **before** DB insert. Plaintext **never** returned in API responses. UI shows hints only (e.g., "Rp 8jt").
- **JWT:** Validate on every request (except `/health`). Extract `user_id` from claim, put in request context.
- **User isolation (app-layer, tested):** every query scopes by `user_id`; a cross-user isolation integration test is the regression guard. RLS is enabled but un-policied (owner connection bypasses it), so it does **not** currently block cross-user access at the DB — do not rely on it as a backstop. Policy-backed RLS is a deferred pre-launch gate.
- **API tokens (v2):** Bcrypt hashed. Plaintext shown once at creation, never again.
- **Image upload:** Max 2MB. Content-type validation. Stored at `receipts/{user_id}/{txn_id}.jpg`.
- **Signed URLs:** 15-min TTL when serving images.
- **CORS:** Whitelist frontend origin only. No `*`.

---

## What NOT to Do

- ❌ Use GORM or any ORM
- ❌ Hand-edit `database/sqlc/generated/` files
- ❌ Store money as float
- ❌ Write a query (or by-id route) that doesn't filter `user_id` — app-layer scoping is the ONLY isolation control today (RLS is a no-op); cover new by-id routes with the isolation test
- ❌ Return raw income in API responses
- ❌ Use `panic` for expected error paths
- ❌ Add Redis "for caching" — not needed at solo-user scale
- ❌ Add gRPC / message queues — single service, goroutines suffice
- ❌ Adopt Nuxt / SSR — authenticated PWA doesn't need it
- ❌ Hardcode hex colors / px spacing in components — use Tailwind config tokens
- ❌ Install component UI libraries (Vuetify, PrimeVue, Quasar) — use shadcn-vue copy-paste pattern
- ❌ Build features outside the MVP scope list above without explicit Hafis approval
- ❌ Skip code review steps — Hafis reviews everything Claude generates before merge

---

## Definition of Done

A feature is **not complete** — and must not be committed to `main` — until **all** of the following hold:

1. **Integration tests for every API endpoint.** Every HTTP route introduced (or modified) has at least one Go integration test that exercises the full request/response cycle through the assembled Echo handler against a **real Postgres database** (the `fintrack_test` DB, never `fintrack`). Tests live next to the code they cover (e.g. `internal/server/server_integration_test.go`) or in `test/integration/`. **Auth-protected endpoints must cover at minimum:** missing token, malformed header, invalid signature, expired token, valid token. Mocks are allowed only for outbound third-party calls (Claude/OpenRouter, MinIO/Supabase Storage) — never for the database or the Echo handler under test.

2. **E2E tests for every user-facing flow.** Every flow a real user can complete (a sequence of UI actions ending in a meaningful outcome — e.g. "load home", "complete onboarding", "scan receipt and confirm") is covered by a **Playwright** test in `web/e2e/`. Tests run against the full local stack (`docker compose` services + `go run ./apps/api` + `vite dev`) — no API mocking at the network layer. Playwright's `webServer` config is the source of truth for how the stack is started.

3. **Tests run green locally before commit.** Both `make test` (Go unit + integration) and `make test-e2e` (Playwright) are required gates. CI will enforce the same.

4. **Test data is isolated.** Integration tests connect to `fintrack_test`; the development `fintrack` DB is never touched by tests. The test DB is reset between test packages (truncate or migrate-down/up).

5. **New flows update the test matrix.** When a new API or flow is added, the relevant test file is added or extended in the **same commit/PR** as the feature. "I'll add tests later" is not acceptable.

**Operationally, this means:** before claiming a feature is "done", run `make test && make test-e2e`. If either is red, the feature is not done.

---

## Development Workflow

### Local dev
```bash
# Backend
cd fintrack
docker compose up -d minio        # storage
go run apps/api/main.go           # API on :8080

# Frontend
cd web
npm run dev                        # Vite dev server on :5173
```

### Database
```bash
# Run migrations
migrate -path database/migrations -database "$DATABASE_URL" up

# Regenerate sqlc after editing query/*.sql
sqlc generate
```

### Testing
```bash
go test ./...                      # All Go tests
cd web && npm test                 # Vue tests (Vitest)
```

---

## Commit Conventions

- **Format:** Conventional Commits (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`)
- **Scope (optional):** `feat(transactions): add manual entry endpoint`
- **Body:** Why, not what. Code shows what.
- **No commit > 200 lines** unless mechanical (sqlc generation, migration, lockfile).

---

## How Claude Should Help

**Operating mode: VIBE CODING.** Claude writes the code, Hafis reviews and approves.

**When Hafis asks for a feature:**
- Implement end-to-end (backend + frontend + tests + migration if needed).
- Show the full diff or files generated, organized for review.
- Explain *why* you made the choices you made — Hafis is reviewing intent, not just syntax.
- If the feature has architectural impact, write an ADR entry in `(C) DECISIONS.md` proactively.

**When Hafis asks for a fix:**
- Diagnose root cause before patching.
- Show the bug, the fix, and why the original code failed.
- Add a test that would have caught it.

**When Hafis asks for review of existing code:**
- Be blunt. Call out anti-patterns. Reference principles above.
- Suggest test cases that would have caught issues.
- Don't be polite about technical debt — name it.

**When Hafis asks for architecture decisions:**
- Surface 2-3 options with trade-offs. Recommend one with reasoning.
- If non-obvious, write an ADR entry to `(C) DECISIONS.md`.

**When Hafis is stuck or debugging:**
- Investigate systematically — don't guess.
- Propose hypotheses, validate with logs/output before changing code.

**Default scope discipline:**
- Reject feature requests outside the MVP scope list. Ask if Hafis wants to update MVP first.
- Reject suggestions that add dependencies not in the stack table without ADR.
- Reject magic abstractions that hide what's happening — keep generated code review-friendly.

---

## Design System

**ALWAYS read `(C) DESIGN.md` before making any visual or UI decision.**

All font choices, colors, spacing, motion, and aesthetic direction are defined there. Do not deviate without:
1. Explicit Hafis approval
2. A new ADR entry in `(C) DECISIONS.md`

**Hard rules from DESIGN.md:**
- Memorable thing: *"Money discipline that feels like training, not bookkeeping."* Every design decision must serve this.
- Fonts: JetBrains Mono (hero numbers) / General Sans (display) / DM Sans (body) — **NO Inter, Roboto, Space Grotesk**
- Brand accent: Saffron Gold (`#F4A300` dark / `#D97706` light) — sacred, used only for currency prefix, primary CTAs, active states
- Semantic colors: green (Fresh) / amber (Warning) / coral (Fatigued) — **ONLY for state, never decoration**
- Hero numbers: typographic composition (mono digits + saffron Rp + muted decimals), not just "big bold white text"
- Both dark and light modes must be tested
- **Light theme is now NEO-BRUTALIST (accepted 2026-06-04 session 2, ADR-2026-06-04b).** Broken-white bg
  (`#F1F1EF` — deliberately NOT Anthropic/Claude cream; see auto-memory `feedback-avoid-claude-colors`),
  white cards, **thick black borders** (`border-2 border-line`), **hard offset shadows** (`shadow-brutal`),
  **sharp corners** (`rounded-card` = 2px), chunky uppercase buttons with `active:translate` press, solid
  color-block status chips. Saffron, the three semantic state colors, and the font stack are UNCHANGED.
  Tokens live in `tailwind.config.ts`. **Rollout is partial** — Home/Budget/Transactions/Onboarding +
  shared nav are brutalist; `ReportsView`/`ScanView`/`Login`/`Register` still use the old style.
  Canonical DESIGN.md (vault) not yet updated to match.
- Mobile-first, single column max-width 420px, bottom tab nav — **on mobile**. (See ADR-2026-06-04 / `PLAN.md`: desktop (≥`lg`) now ADDS a responsive layer — left sidebar nav instead of bottom tabs, wider/multi-column content, and a desktop-optimized Reports page. Mobile-first remains the default and the mobile experience is unchanged. Session 2 extended this: Home/Budget/Transactions also widen/multi-column at `≥lg`, and **logout lives in the sidebar navbar** on desktop.)
- Motion is dynamic — count-ups, state transitions, signature scan-flow choreography
- `prefers-reduced-motion` must be respected

**In QA mode:** flag any component that doesn't match DESIGN.md.

---

## Related Docs (in Hafis-Brain vault)

These live outside the repo. Hafis maintains them as the project's "second brain."

- `03 Projects/Fintrack/(C) PROJECT.md` — Project overview, MVP scope, success criteria
- `03 Projects/Fintrack/(C) ARCHITECTURE.md` — Full stack rationale, data flows, storage layer
- `03 Projects/Fintrack/(C) ROADMAP.md` — 5 phases, ~12 weeks effort
- `03 Projects/Fintrack/(C) DECISIONS.md` — ADR log of every choice made
- `03 Projects/Fintrack/(C) DESIGN.md` — Design system (typography, color, motion, layout) ⭐ MUST READ BEFORE UI WORK

---

## Status

| Date | Status |
|------|--------|
| 2026-06-03 | CLAUDE.md drafted in vault. No code yet. Phase 0 of roadmap starts next. |
| 2026-06-04 (s1) | All 4 MVP features built; added local auth UI, beranda dashboard, budget-vs-actual insight + reduction recommendations. Committed to `main`. |
| 2026-06-04 (s2) | Conversational planner onboarding (LLM language-layer + deterministic Go math), inline custom-category creation, re-budget button, neo-brutalist light theme, desktop-optimized layouts + sidebar logout. Verified green (`make test` + `make test-e2e` 37/37). **Committed to `main` (3 commits, not pushed) — see `PLAN.md` NEXT-SESSION STATUS banner.** |
