# Fintrack — Working Plan

Repo-local plan log. Canonical roadmap/ADRs live in the Hafis-Brain vault; this file
records in-repo what's being built and why, so it travels with the code.

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

## Feature batch 2 (in progress, 2026-06-04)

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

## Deferred — pre-public-launch gates (from /review 2026-06-04)

Recorded so they aren't forgotten. None are exploitable on the current local single-user setup; all matter before internet exposure.

- **[P1] Rate-limit `/v1/auth/login` + `/register`** before public launch (or when Supabase auth lands in v2). Today they accept unlimited password guesses, each forcing a bcrypt hash → brute-force + CPU-DoS. Deferred by Hafis (Supabase owns auth/throttling in v2).
- **[P2] Receipt serving via signed URLs.** `receipt_url` is stored as a plain object URL; CLAUDE.md mandates 15-min signed URLs when serving. `storage.SignedURL` exists but isn't wired — there's no receipt-viewing UI yet (deferred MVP feature), so no serve path exists. Wire it when that UI is built.
- **[P2] Custom-category uniqueness.** `POST /v1/categories` has no per-user `(user_id, lower(name))` uniqueness, so a user can create duplicate custom categories. Add a constraint + 409 mapping.
- **[P2] JWT `iss` not verified** by the middleware (any valid-signature token passes). Harmless with one local secret; add `jwt.WithIssuer` when Supabase shares/rotates secrets in v2.

## Done (earlier batches, 2026-06-03 → 04)
- All 4 MVP features (onboarding, transactions CRUD, receipt scan, fatigue dashboard).
- Local email **register/login + bcrypt password + logout**; JWT-expiry route guard.
- Onboarding: persist answers on "Ubah jawaban"; allow + coach overspend; custom expense categories;
  finish-to-budget CTA.
- Beranda dashboard (health card removed); budget vs actual graphic + reduction recommendations.
- Transaction input fixes (any amount > 0 valid; layout).
