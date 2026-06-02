# CEO Review — Fintrack Portfolio Build

**Plan under review:** `docs/superpowers/plans/2026-04-29-backend-mvp.md` (3,378 lines, 16-phase Go backend MVP).
**Companion design doc (just written by /office-hours):** `docs/superpowers/designs/2026-06-01-fintrack-portfolio-sequencing.md` (321 lines).
**Repo state:** No git, no CLAUDE.md, no TODOS.md. Only `docs/` and `full_doc.html` exist.
**Context:** /office-hours just reframed this project from "startup attempt" → "portfolio piece showcasing AI-orchestrated development." Scope locked to all 25 endpoints. Sequencing locked to Approach A (vertical-slice). No interview deadline (slow-burn).

---

## Pre-Review System Audit

- **Plan file** is 3,378 lines, ~92KB. Sequential 16-phase TDD plan. Written 2026-04-29, before the portfolio reframing.
- **Design doc** is the source of truth for scope and sequencing decisions (written today).
- **Greenfield** — no existing code to leverage. Every line is new.
- **No git** — the very first build action will be `git init`.
- **No CI, no deploy infra, no env vars set up yet.**
- **Frontend** mentioned in plan structure (`web/` dir) but is deferred to phase ~15. Design doc wants a thin frontend in Milestone 6.

## Step 0A — Premise Challenge (already done in /office-hours)

Recap from /office-hours session above, no re-litigation:
- Project reframed: portfolio-first, AI-orchestration showcase.
- Status quo: Indonesian fresh workers e-wallet panic on the 20th. Tolerable pain, hard to displace.
- No founder-market fit, no user contact, no demand validation — and the reframe makes that fine.
- The 92KB PRD + 16-phase plan + this CEO doc are themselves portfolio artifacts.

## Step 0B — Existing Code Leverage

N/A. Greenfield project. No code to reuse. Every line is new.

## Step 0C — Dream State Mapping

```
CURRENT STATE             THIS PLAN                    12-MONTH IDEAL
[full_doc.html only,      [25 endpoints,                [Live URL, hiring managers click,
 no code, no git]   ───>   live demo, README]    ───>    portfolio converts to interviews
                                                          or freelance gigs]
```

The plan moves toward the ideal cleanly. No drift.

## Step 0C-bis — Implementation Alternatives

Generated in /office-hours, decision already locked:
- **A) Vertical showcase first, horizontal fill after** — CHOSEN. Foundation+deploy → 3 vertical slices → horizontal fill → thin frontend. Live URL gate at Milestone 1.
- B) Execute 16-phase plan as written — rejected (no live URL for 3 months, high motivation-drop risk).
- C) Frontend demo first, backend backfills — rejected (dilutes Go signal).

No re-evaluation needed. The decision was made with full context less than an hour ago.

## Step 0F — Mode Selected: HOLD SCOPE

Confirmed by user. Focus: bulletproof the plan-as-written and surface gaps between plan-file and design-doc.

---

## Section Verdicts (11 sections)

| # | Section | Verdict |
|---|---------|---------|
| 1 | Architecture | **Finding 1** — plan sequencing conflicts with design doc |
| 2 | Error & Rescue | No issues — plan has detailed error envelope + named exceptions |
| 3 | Security | Minor — rate limit lives in phase 16, deployed publicly before that |
| 4 | Data Flow & Edge Cases | No issues — AI receipt scan has parse-error fallback, multipart validation |
| 5 | Code Quality | No issues — sqlc + manual DI + Echo is a clean Go stack, no DRY violations in 3,378-line plan |
| 6 | Tests | No issues — TDD throughout, test patterns for service/repo/handler all specified |
| 7 | Performance | Minor — token middleware does N+1 bcrypt scan, but plan acknowledges and defers |
| 8 | Observability | Minor — logger + request IDs, no metrics/traces (appropriate for portfolio scope) |
| 9 | Deployment | **Finding 1** (same as Section 1 — deploy is in phase 16, design doc wants it at Milestone 1) |
| 10 | Long-term | No issues — Go + Echo + sqlc choices are stable in 2026, won't age badly |
| 11 | Design/UX | **Finding 2** — frontend (`web/`) and README are missing from the 16 phases |

**Cross-cutting finding 3:** Plan does not mention preserving AI-build artifacts (`docs/superpowers/` to repo, commit-message conventions). For an AI-orchestration portfolio piece, this is the actual differentiator.

---

## Decisions Locked

| # | Finding | Decision |
|---|---------|----------|
| 1 | Plan sequencing vs design doc | **A** — Add "Execution Order Override" note at top of plan file |
| 2 | Frontend + README not phases | **A** — Add Phase 17 (Frontend) + Phase 18 (Portfolio README + Loom) |
| 3 | AI-build artifacts invisible | **A** — Add Phase 0 (git init + commit planning artifacts) + commit conventions |

---

## Action Items (post-exit-plan-mode, edit the project plan file directly)

The following text needs to be added to `docs/superpowers/plans/2026-04-29-backend-mvp.md`. Three insertions:

### Insert 1 — at TOP of plan file, right after the `# Fintrack Backend MVP Implementation Plan` heading

```markdown
> **Execution order override (2026-06-01):** This plan file is the detailed
> reference but is NOT the execution order. The current execution order is
> defined in `docs/superpowers/designs/2026-06-01-fintrack-portfolio-sequencing.md`
> (Approach A — vertical showcase first). Specifically:
>
> - **Milestone 1 = Phase 0 + Phases 1–4 + Task 37 (health) + Task 39 (Railway deploy).**
>   Deploy is the gate. No feature work until `/health` is live on a public URL.
> - **Milestone 2 = Phases 5, 6, 7** (onboarding + budget engine vertical slice).
> - **Milestone 3 = Phases 8, 9, 10** (receipt scan + transactions + fatigue vertical slice).
> - **Milestone 4 = Phase 15** (weekly narrative worker).
> - **Milestone 5 = Phases 11, 12, 13, 14** (debts, goals, reports, BYOA tokens — horizontal fill).
> - **Milestone 6 = Phase 17 + Phase 18** (frontend + portfolio README).
>
> Do not execute phases in numerical order. Follow the milestone order in the design doc.
```

### Insert 2 — NEW Phase 0, inserted between "Context" section and "Phase 1 — Foundation"

```markdown
---

## Phase 0 — Repo Initialization & AI-Build Artifact Preservation

The very first thing this project ships is the planning trail. Before any Go
code, the repo must be initialized with the PRD, plan, and design doc visible
to anyone who browses it. This is what makes the AI-orchestration story legible
to reviewers.

### Task 0 — Git init, commit planning artifacts, push to GitHub

**Files preserved (NOT gitignored):**
- `full_doc.html` (the PRD)
- `docs/superpowers/plans/2026-04-29-backend-mvp.md` (this file)
- `docs/superpowers/designs/2026-06-01-fintrack-portfolio-sequencing.md` (design doc)

- [ ] **Step 1: `git init` and configure**

```bash
git init
git config user.name "<your name>"
git config user.email "<your email>"
```

- [ ] **Step 2: Create initial `.gitignore`** (Go-flavored)

```gitignore
# Go
/bin
/tmp
*.test
*.out
coverage.*

# Env / secrets
.env
*.local.env

# sqlc generated
/database/sqlc/generated

# OS
.DS_Store

# Editor
.idea/
.vscode/
```

**Critical:** do NOT gitignore `docs/`, `full_doc.html`, or `docs/superpowers/`.
These are the AI-build artifacts and they must be visible in the repo.

- [ ] **Step 3: First commit — planning artifacts only**

```bash
git add full_doc.html docs/ .gitignore
git commit -m "chore: initial commit — PRD, implementation plan, design doc

Planning artifacts written with Claude Code (Opus 4.7) via /office-hours.
Code phases begin in Phase 1."
```

- [ ] **Step 4: Create GitHub repo and push**

```bash
# Create the repo via gh CLI or the GitHub web UI first, then:
git remote add origin https://github.com/<owner>/fintrack.git
git branch -M main
git push -u origin main
```

- [ ] **Step 5: Create README v0**

A 20-line README v0 that says: "Fintrack — gym app for your money. Built with
Claude Code. Status: pre-implementation (Phase 1 starts after deploy infra is
live). See `docs/superpowers/` for the planning trail."

This README gets rewritten in Phase 18. For now its only job is to give a
visitor enough context to understand the repo isn't empty by accident.

### Commit message conventions (apply to ALL subsequent commits)

- Use Conventional Commits format: `feat(scope): ...`, `fix(scope): ...`,
  `chore: ...`, `test(scope): ...`, `docs: ...`.
- When a commit was AI-drafted and you reviewed/adjusted, append
  `(AI-drafted, human-reviewed)` to the first line.
- When a commit was AI-drafted and you accepted with no changes, append
  `(AI-drafted)`.
- When a commit is fully human-authored (rare — usually planning or copy edits),
  no annotation needed.
- Do not fake annotations. The honest record IS the credibility signal.
- Do NOT include `Co-Authored-By: Claude` or similar lines — the workflow note
  in the README + the commit annotations are sufficient.

---
```

### Insert 3 — NEW Phase 17 + Phase 18, inserted AFTER current Phase 16 (Polish & Deploy)

```markdown
---

## Phase 17 — Thin Demo Frontend

Per design doc Milestone 6. A reviewer cannot evaluate the receipt-scan flow
from curl examples — they need a live page where they upload a photo and see
the AI response. Frontend is intentionally thin: this is not a frontend
portfolio piece, it's the demo surface for the backend.

### Task 40 — Next.js app scaffold under `web/`

**Files:**
- Create: `web/package.json`, `web/next.config.js`, `web/tsconfig.json`
- Create: `web/.env.example` (just `NEXT_PUBLIC_API_URL`)

- [ ] **Step 1: `npx create-next-app@latest web --typescript --tailwind --app`**
- [ ] **Step 2: Add `NEXT_PUBLIC_API_URL` to `.env.example`** pointing to the Railway URL
- [ ] **Step 3: Commit**

### Task 41 — Five hero pages

Each page hits the live Railway backend; no mocking, no fakes.

- [ ] **`/onboarding`** — 6-question form, POST to `/v1/onboarding`, show resulting budget
- [ ] **`/receipt-scan`** — file upload, POST to `/v1/transactions/scan`, show result
- [ ] **`/dashboard`** — GET `/v1/fatigue/dashboard`, show Fresh/Warning/Fatigued cards
- [ ] **`/weekly-report`** — GET `/v1/reports/weekly`, render the AI narrative
- [ ] **`/debts`** — list debts with snowball order, payment form

Styling: Tailwind defaults, use the gym-metaphor copy from the PRD. No design
system needed — the goal is "looks acceptable and the demo works."

- [ ] **Step 1–5: Build each page incrementally with TDD (vitest + React Testing Library)**
- [ ] **Step 6: Commit each page separately**

### Task 42 — Deploy to Vercel

- [ ] **Step 1: `vercel` CLI deploy or GitHub integration**
- [ ] **Step 2: Set `NEXT_PUBLIC_API_URL` env var to Railway URL**
- [ ] **Step 3: Verify all 5 pages work end-to-end against live Railway backend**
- [ ] **Step 4: Commit `vercel.json` if used**

---

## Phase 18 — Portfolio README + Demo Artifacts

The README is the actual deliverable. A reviewer spends 60 seconds on it before
deciding whether to look at code or close the tab. Treat this phase as the most
important phase in the project.

### Task 43 — README rewrite (the headline artifact)

**File:** `README.md`

Structure:

```markdown
# Fintrack — gym app for your money

> Indonesian-native personal finance PWA. AI scans receipts, categorizes
> spending, warns when a category is "fatigued," and writes a weekly narrative
> in Bahasa Indonesia.

**Live demo:** https://fintrack-web.vercel.app
**API:** https://fintrack-api.railway.app
**Walkthrough Loom:** [2-minute video link]

## What it does
[Gym metaphor — programs, category fatigue, coach voice]
[1-2 paragraphs, plain Indonesian-aware language]

## How it was built
This project was built solo with Claude Code (Opus 4.7) as the primary
implementation driver. I wrote the PRD, the 16-phase plan, and the office-hours
design doc; Claude Code generated the Go code, the SQL schema, the AI prompts,
and the tests against those artifacts. Commit messages mark which work was
AI-drafted vs human-reviewed.

The planning trail is open: see `full_doc.html` for the PRD,
`docs/superpowers/plans/` for the implementation plan, and
`docs/superpowers/designs/` for the office-hours design decisions.

## Architecture
[ASCII diagram of the system: client → Railway API → Postgres + Anthropic API]

## Tech stack
Go 1.22 · Echo v4 · pgx/v5 · sqlc · Postgres (Supabase) · Anthropic Claude API ·
Railway · Next.js 14 · Vercel

## Local setup
[copy from .env.example, make migrate-up, make run]

## What's NOT in this build
[Honest list: no real user research, no production traffic, designed to
demonstrate AI-orchestrated solo dev — not a market-validated product]
```

- [ ] **Step 1: Draft each section**
- [ ] **Step 2: Add architecture diagram (ASCII or rendered image)**
- [ ] **Step 3: Embed the Loom (see Task 44)**
- [ ] **Step 4: Commit**

### Task 44 — 2-minute Loom walkthrough

Record:
- 0:00–0:20: open the live demo, show the onboarding flow finishing in 30 seconds
- 0:20–0:50: scan a real Indonesian struk (Indomaret / Alfamart) and show the result
- 0:50–1:20: show the fatigue dashboard reflecting the new transaction in real time
- 1:20–1:50: show the weekly narrative from the cron job
- 1:50–2:00: end on the GitHub repo, point at `docs/superpowers/`

- [ ] **Step 1: Script (rehearse twice before recording)**
- [ ] **Step 2: Record on Loom**
- [ ] **Step 3: Embed in README**

### Task 45 — Optional: LEARNED.md

A running list of non-obvious things you learned about driving AI through this
build. Hiring managers love these because they prove genuine reflection. Skip
if you don't have at least 5 honest entries.

### Task 46 — Final commit + push + verify

- [ ] **Step 1: `git push origin main` and click through the GitHub repo as a stranger would**
- [ ] **Step 2: Open the live URL in a private tab, click through all 5 pages**
- [ ] **Step 3: Check the README renders correctly on GitHub**
- [ ] **Step 4: Add the live URL to your LinkedIn / personal site / CV**

The project is "shipped" when a stranger can find the live URL from your CV,
click it, and grasp what Fintrack does in under 60 seconds.

---
```

---

## Completion Summary

```
+====================================================================+
|            MEGA PLAN REVIEW — COMPLETION SUMMARY                   |
+====================================================================+
| Mode selected        | HOLD SCOPE                                  |
| System Audit         | greenfield, no git, plan written pre-pivot  |
| Step 0               | HOLD SCOPE (office-hours just locked scope) |
| Section 1  (Arch)    | 1 issue (sequencing) — resolved via D2      |
| Section 2  (Errors)  | 0 issues — plan has named exceptions        |
| Section 3  (Security)| 1 minor (rate limit late) — accepted as-is  |
| Section 4  (Data/UX) | 0 issues                                    |
| Section 5  (Quality) | 0 issues                                    |
| Section 6  (Tests)   | 0 issues — TDD throughout                   |
| Section 7  (Perf)    | 1 minor (N+1 bcrypt token scan, deferred)   |
| Section 8  (Observ)  | 1 minor (no metrics/traces) — accepted      |
| Section 9  (Deploy)  | 1 issue — same as Section 1, resolved D2    |
| Section 10 (Future)  | 0 issues — stack is durable                 |
| Section 11 (Design)  | 1 issue (no frontend phase) — resolved D3   |
+--------------------------------------------------------------------+
| Cross-cutting        | AI-build artifacts — resolved via D4        |
| NOT in scope         | Email integration, Supabase Storage upload, |
|                      | progressive saving rate, monitoring         |
|                      | dashboards, OAuth providers beyond Supabase |
| What already exists  | Nothing — greenfield                        |
| Dream state delta    | Plan now produces live URL + portfolio      |
|                      | README + AI-build trail visible in repo     |
| TODOS.md updates     | 0 deferred (HOLD SCOPE — no expansion)      |
| CEO plan             | Skipped (HOLD SCOPE)                        |
| Outside voice        | Skipped (low-leverage given fresh           |
|                      | office-hours session)                       |
| Diagrams produced    | Plan-vs-design-doc sequencing mapping       |
+====================================================================+
```

## Verdict

**CEO REVIEW CLEAR — PLAN READY TO EXECUTE (after applying the 3 inserts above)**

The plan was solid on engineering, but had three blind spots from being
written before the portfolio reframing:
1. Wrong execution order (deploy at the end vs. deploy first)
2. Missing portfolio-critical phases (frontend, README)
3. Invisible AI-build trail (no Phase 0, no commit conventions)

All three are resolved by the additions above. Once you apply them to
`docs/superpowers/plans/2026-04-29-backend-mvp.md`, the plan and the design
doc agree, and you can start Milestone 1 (Phase 0 → Phase 1 → ... → deploy)
with confidence.

## Next Steps

1. **Apply the 3 inserts** to `docs/superpowers/plans/2026-04-29-backend-mvp.md` after exiting plan mode.
2. **Start Milestone 1** by opening a fresh Claude Code session and asking it to execute Phase 0 → Phase 1 → Phase 4 + Tasks 37 & 39.
3. **Optional: run `/plan-eng-review`** for architecture-level scrutiny (sqlc patterns, error envelope, RLS test strategy, Anthropic retry policy). Recommended before Phase 5+ but not blocking.
4. **Skip `/plan-design-review`** for now — frontend doesn't exist yet. Run it when you reach Milestone 6.
