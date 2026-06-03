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
3. **Database is source of truth.** RLS enforces user isolation at DB layer, not app layer.
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
- **All tables:** `user_id` column + RLS policy (except system tables like `expense_categories` defaults)
- **Money:** Stored as `BIGINT` (Rupiah, no decimals). Never `FLOAT` / `DECIMAL` for currency.
- **Timestamps:** `TIMESTAMPTZ` always. Use `now()` default.
- **Soft delete:** Only where needed (transactions yes, categories no).
- **Indexes:** Add explicitly per query pattern. Document in migration comment.
- **RLS:** Policy per table. Test with `SET ROLE` queries.

---

## Security & Privacy Constraints

- **Income encryption:** AES-256-GCM **before** DB insert. Plaintext **never** returned in API responses. UI shows hints only (e.g., "Rp 8jt").
- **JWT:** Validate on every request (except `/health`). Extract `user_id` from claim, put in request context.
- **RLS enforcement:** Even with bugs in app code, DB blocks cross-user data access.
- **API tokens (v2):** Bcrypt hashed. Plaintext shown once at creation, never again.
- **Image upload:** Max 2MB. Content-type validation. Stored at `receipts/{user_id}/{txn_id}.jpg`.
- **Signed URLs:** 15-min TTL when serving images.
- **CORS:** Whitelist frontend origin only. No `*`.

---

## What NOT to Do

- ❌ Use GORM or any ORM
- ❌ Hand-edit `database/sqlc/generated/` files
- ❌ Store money as float
- ❌ Skip RLS policies
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
- Mobile-first, single column max-width 420px, bottom tab nav
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
