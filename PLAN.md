# Fintrack — Working Plan

Repo-local plan log. Canonical roadmap/ADRs live in the Hafis-Brain vault; this file
records in-repo what's being built and why, so it travels with the code.

---

## NEXT-SESSION STATUS (end of 2026-06-04, session 2)

**Feature batch 3 below is built, verified green, and COMMITTED to `main`** in 3 commits:
- `bab2de0` feat(onboarding): conversational financial planner backend
- `1b8b977` feat(web): planner wizard, inline categories, re-budget, brutalist + desktop UI
- `e0da8d2` docs: record session-2 batch

**Not yet pushed.** `origin/main` is public and generated sqlc + `web/` build output are gitignored, so
**add CI (sqlc generate + go build + make test + e2e) before relying on the remote**, then push.

Verified at session end: `go build ./...` ✅, `vue-tsc --noEmit` ✅, `make test` ✅,
`make test-e2e` ✅ **37/37**.

Recorded gstack learnings this session:
`e2e-dev-server-shadow`, `stale-sqlc-ide-diagnostics`, `dev-jwt-secret-mismatch`,
`onboarding-finalize-idempotent`, `zsh-uid-readonly`, `playwright-default-viewport-390`
(see `/learn`).

---

## ADR-2026-06-04 — Desktop responsive layer + Reports

**Status:** accepted (Hafis). **Context:** Fintrack is mobile-first (DESIGN.md: single column,
max-width 420px, bottom tab nav). On a desktop screen that wastes the space and makes it hard to
*read a spending report*. Hafis asked for a better desktop experience and reporting tools.

**Decision:**
- **Keep mobile-first as the default.** The mobile experience does not change.
- **Add a responsive layer at `≥lg` (1024px):** the bottom tab nav becomes a **left sidebar**, and
  content is allowed to widen / go multi-column where it helps (Reports, Budget).
- **Add a dedicated `/reports` page** optimized for desktop width.
- This is an explicit, recorded departure from the strict "max-width 420px" rule (CLAUDE.md Design
  System note updated). No new heavy dependencies; pure Tailwind responsive utilities + existing tokens.

**Rejected alternatives:** a separate "admin mode" toggle (more state, two UIs to maintain); a
charting library (custom SVG/CSS keeps the bespoke aesthetic and zero deps).

---

## Feature batch 2 (DONE + committed, 2026-06-04 session 1)

> Shipped and merged to `main` in session 1 (commits incl. reports/auth/responsive nav). Kept here
> for context. Session 2's further desktop work is in batch 3 below.

All three are **frontend-only** — the transactions API already supports `from`/`to` date filtering,
and CSV is generated client-side so it carries the JWT (a plain download link would not).

### 1. Responsive desktop UI
- `App.vue` shell: `≥lg` → fixed **left sidebar nav** (Beranda, Transaksi, Scan, Budget, Reports);
  `<lg` → existing bottom tab nav (now including Reports). Content container widens on desktop.
- Views stay usable centered on mobile width; Reports + Budget get desktop multi-column treatment.

### 2. Month filter (every month's spending)
- A month selector (prev/next + current-month label, default = current month) that sets `from`/`to`
  on `listTransactions`. Lives on the **Reports** page (primary) and the **Transactions** list.

### 3. CSV export
- "Export CSV" on Reports: fetch the selected month's transactions, build a CSV
  (`tanggal, kategori, merchant, jumlah, catatan`), download as `fintrack-YYYY-MM.csv` via a Blob.
  Client-side so the axios auth header is honoured.

### Reports page (`/reports`)
- Month filter → fetch month's transactions → aggregate spend by category → **table**
  (category, spent, % of total, count) + **bar chart** (reuse the compare-chart style) + total + **Export CSV**.
- Desktop: table + chart side-by-side (`lg:grid-cols-2`); mobile: stacked. Loading/empty/error states.

### Verification
- `make test` (Go untouched, stays green) + `make test-e2e`: new specs for the Reports flow
  (filter changes the data, CSV triggers a download), the Transactions month filter, and responsive
  nav (sidebar at desktop viewport, bottom nav at mobile viewport). Playwright default viewport set to
  mobile; desktop checks set their own viewport.

---

## ADR-2026-06-04b — Conversational planner (LLM as language-only layer)

**Status:** accepted (Hafis). **Context:** onboarding was a "fill every expense" form. Hafis wanted it
to act like a financial planner: ask goal/debt/emergency/lifestyle, take only the **fixed** expenses
("can't change"), then **auto-suggest** the flexible (keinginan) categories, and let the user refine
either by editing a number or by **chatting** ("I prefer 1.5jt for food").

**Decision:**
- **Money math is deterministic and lives in Go** (`internal/domain/budget/planner.go`:
  `SuggestFlexible`, `Rebalance`). The LLM never invents numbers.
- **The LLM is a language-only layer** (`internal/llm/`, OpenRouter via stdlib `net/http`). For chat it
  returns strict JSON `{reply, adjustments:[{category_name,target_amount}], take_from_savings}`; the
  handler resolves names→ids and calls `Rebalance` to compute the actual allocations.
- **Deterministic STUB client when `OPEN_ROUTER_API_KEY` is empty** (regex NLU) so tests/e2e never call
  out and are reproducible. Dev (`ENV=development` with the key set) hits real OpenRouter.
- **Endpoints:** `POST /v1/onboarding/suggest`, `POST /v1/planner/chat`. Confirm still reuses the
  existing idempotent `POST /v1/onboarding` finalize.

**Also accepted this session:** **neo-brutalist light theme** (broken-white `#F1F1EF` bg — explicitly
NOT Anthropic cream; black borders, hard offset shadows, sharp corners; saffron stays the brand accent)
and a **desktop UX pass** (wide multi-column layouts, sidebar logout, obvious CTAs). DESIGN.md (vault)
should be updated to reflect the brutalist direction — not yet done.

---

## Feature batch 3 (this session, 2026-06-04 session 2) — DONE + committed (not pushed)

1. **Conversational planner onboarding** — 3-step wizard (`OnboardingView.vue` + `PlannerChat.vue`),
   `internal/llm/` (OpenRouter + stub), `internal/domain/budget/planner.go`,
   `internal/handler/planner.go`, `web/src/api/planner.ts`. See ADR-2026-06-04b.
2. **Inline custom-category creation** — `web/src/components/AddCategoryInline.vue`, wired into
   onboarding step 2 (defaults type `fixed`, no picker) and the add-transaction form (type picker,
   default `variable`, auto-selects the new category). Backend `POST /v1/categories` already existed.
3. **Re-budget entry point** — "↻ Atur ulang budget" button in the Budget header → `/onboarding`.
   Safe because finalize is idempotent (ADR-2026-06-04b / learning `onboarding-finalize-idempotent`).
4. **Neo-brutalist theme** — `tailwind.config.ts` tokens + restyle of Home/Budget/Transactions +
   `BudgetCompareChart`/`ReduceSuggestions`/`SidebarNav`/`App.vue`.
5. **Desktop optimization** — Home/Budget/Transactions widen and go multi-column at `≥lg`; **logout
   moved into the sidebar navbar** (`sidebar-logout`; the home-header `Keluar` is now mobile-only,
   `lg:hidden`); "Catat transaksi" and the row **Edit/Hapus** buttons made obvious.

**Verification:** `vue-tsc` ✅, `make test` ✅, `make test-e2e` ✅ **37/37** (new specs: onboarding custom
fixed expense, transaction custom category, re-budget navigation, desktop sidebar logout).

**Known follow-ups from this batch:**
- Custom-category uniqueness is now exposed in **two** UIs → the deferred `(user_id, lower(name))`
  constraint below is more relevant; a dup just renders twice today (harmless, ugly).
- Re-budget starts the wizard from defaults after a fresh login (the 6 answers aren't persisted
  server-side, only the derived program). Optional: pre-fill income from the existing plan.
- Full brutalist rollout still pending on `ReportsView` (still the old bordered style), `ScanView`,
  `LoginView`/`RegisterView`. Update DESIGN.md (vault) + add an ADR when that lands.

---

## Deferred — pre-public-launch gates (from /review 2026-06-04)

Recorded so they aren't forgotten. None are exploitable on the current local single-user setup; all matter before internet exposure.

- **[P1] Rate-limit `/v1/auth/login` + `/register`** before public launch (or when Supabase auth lands in v2). Today they accept unlimited password guesses, each forcing a bcrypt hash → brute-force + CPU-DoS. Deferred by Hafis (Supabase owns auth/throttling in v2).
- **[P2] Receipt serving via signed URLs.** `receipt_url` is stored as a plain object URL; CLAUDE.md mandates 15-min signed URLs when serving. `storage.SignedURL` exists but isn't wired — there's no receipt-viewing UI yet (deferred MVP feature), so no serve path exists. Wire it when that UI is built.
- **[P2] Custom-category uniqueness.** `POST /v1/categories` has no per-user `(user_id, lower(name))` uniqueness, so a user can create duplicate custom categories. Add a constraint + 409 mapping.
- **[P2] JWT `iss` not verified** by the middleware (any valid-signature token passes). Harmless with one local secret; add `jwt.WithIssuer` when Supabase shares/rotates secrets in v2.
- **[P1] Policy-backed RLS** (CSO audit 2026-06-05). Every table has `ENABLE ROW LEVEL SECURITY` but **no policies**, and the app connects as the table owner — so RLS is a **no-op**, not the isolation backstop CLAUDE.md used to claim. Isolation is currently app-layer only (every query filters `user_id`) and is now regression-guarded by `test/integration/isolation_integration_test.go`. Before public/multi-user launch, add real per-table policies + `FORCE ROW LEVEL SECURITY` + a non-owner DB role that `SET LOCAL app.user_id` per request, OR keep app-layer-only and expand the isolation test to every by-id route. CLAUDE.md updated to stop over-claiming DB-layer isolation.

## Done (earlier batches, 2026-06-03 → 04)
- All 4 MVP features (onboarding, transactions CRUD, receipt scan, fatigue dashboard).
- Local email **register/login + bcrypt password + logout**; JWT-expiry route guard.
- Onboarding: persist answers on "Ubah jawaban"; allow + coach overspend; custom expense categories;
  finish-to-budget CTA.
- Beranda dashboard (health card removed); budget vs actual graphic + reduction recommendations.
- Transaction input fixes (any amount > 0 valid; layout).
