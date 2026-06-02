# Fintrack Backend MVP Implementation Plan

## Build Status (last updated 2026-06-02)

Legend: ✅ done · 🟡 partial · ⏳ pending · ⏭️ deferred (postponed by choice)

| Phase | Task | Title | Status | Commit / Notes |
|-------|------|-------|--------|----------------|
| 0 | 0  | Git init + planning artifacts + GitHub push | ✅ | `c70062e`, `2c9aef3` — repo live at hafis915/fintrack |
| 1 | 1  | Project skeleton | ✅ | `8fc690c` |
| 1 | 2  | Config loader (viper) | ✅ | `4b4733d` |
| 1 | 3  | Logger (zerolog) | ✅ | `a4be019` |
| 1 | 4  | AppError package | ✅ | `0fdcb5a` |
| 1 | 5  | Response envelope | ✅ | `d26a98a` |
| 1 | 6  | Validator wrapper | ✅ | `f05e99b` |
| 2 | —  | Docker Compose for local Postgres (pre-step) | ✅ | `ae655c0` — port 5433 to avoid host conflict |
| 2 | 7  | Migration 0001 (extensions + enums) | ✅ | `c1e1ba2` |
| 2 | 8  | Migrations 0002–0003 (user_profiles, expense_categories) | ✅ | `2afd5bc` |
| 2 | 9  | Migrations 0004–0005 (budget_plans, budget_items) | ✅ | `71bd80a` |
| 2 | 10 | Migration 0006 (transactions) | ✅ | `6a2f97b` |
| 2 | 11 | Migration 0007 (debt_items) | ✅ | `59370ae` |
| 2 | 12 | Migration 0008 (goals) | ✅ | `e4ec9c8` |
| 2 | 13 | Migrations 0009–0010 (weekly_reports, api_tokens) | ✅ | `4275f45` |
| 2 | 14 | RLS + seed + dev auth shim | ✅ | `5d471bd` — added 000011_dev_auth_shim before RLS (shifted plan's 0011→0012, 0012→0013) |
| 2 | 15 | sqlc setup + first user queries | ✅ | `3dceedd` |
| 2 | 16 | pgx connection pool | ✅ | `ec82a92` |
| 3 | 17 | Echo bootstrap + middleware + stub /health | ✅ | `7099c6b` |
| 4 | 18 | AES-256-GCM income encryption | ✅ | `b219dff` |
| 4 | 19 | JWT auth middleware (Supabase HS256) | ✅ | `6b04791` |
| 5 | 20 | User domain entity + repo interface | ✅ | `99077b5` |
| 5 | 21 | User service | ✅ | `9664f9a` (+ `cb09f82` followup for objx dep) |
| 5 | 22 | User repository (sqlc-backed) | ✅ | `b275bfb` — repo absorbs pgtype↔domain conversions |
| 5 | 23 | Profile handler + DTO + wiring | ✅ | `62a5c5d` |
| 6 | 24 | Category entity + repo + queries | ✅ | `2f807ec` |
| 7 | 25 | Budget engine (pure logic) | ✅ | `b2b4d81` — 4 unit tests all pass |
| 7 | 26 | Budget repo + service + onboarding handler | ✅ | `02ace70` — full onboarding flow verified live |
| 8 | 27 | Transactions CRUD | ⏳ | — |
| 9 | 28 | Fatigue calculator + handler | ⏳ | — |
| 10 | 29 | Anthropic HTTP client | ⏳ | — |
| 10 | 30 | Receipt categorizer (HERO) | ⏳ | — |
| 10 | 31 | Narrative summarizer | ⏳ | — |
| 11 | 32 | Debts domain | ⏳ | — |
| 12 | 33 | Goals CRUD | ⏳ | — |
| 13 | 34 | Reports domain | ⏳ | — |
| 14 | 35 | API Tokens (BYOA) | ⏳ | — |
| 15 | 36 | Worker (weekly cron) | ⏳ | — |
| 16 | 37 | Real /health with DB Ping | ✅ | `3f1cd06` |
| 16 | 38 | Rate limiting middleware | ⏳ | — |
| 16 | 39 | Railway deploy config (Dockerfile + railway.toml + PORT fix) | ⏭️ DEFERRED | Per builder choice 2026-06-02 — Configs unwritten. Needs: Go 1.25-alpine in Dockerfile, PORT fallback in config.go, Railway provisioning |
| 16 | 40 | End-to-end smoke script | ⏳ | — |
| 17 | 41 | Frontend scaffold + 5 hero pages + Vercel | ⏳ | — |
| 18 | 43–46 | Portfolio README + Loom + final verify | ⏳ | — |

**Milestones (per design doc):** Milestone 1 = 🟡 code done / deploy deferred · Milestone 2 = 🟢 done · Milestones 3–6 = ⏳ pending.

**Verified locally (end-to-end on `localhost:8090`):** `/health` returns `{db: ok}`, JWT-gated `/v1/*` routes work, AES income encryption round-trips, onboarding generates per-program budget, GET /v1/budget/current reads back the plan, GET /v1/profile shows the upserted profile.

---

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

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Fintrack Go backend (REST API + worker) covering all 25 MVP endpoints from the PRD: onboarding/budget engine, transactions with AI receipt scan, category fatigue, debts, goals, reports, and BYOA tokens.

**Architecture:** Monorepo Go workspace. Single Echo HTTP service in `apps/api` plus a separate `apps/worker` for cron jobs. Layered: handler → service → repository → DB. Domain interfaces live in `internal/domain/<context>`, implementations in `internal/repository`. sqlc generates type-safe DB code from raw SQL. Manual DI in `main.go` (no Wire/Fx). Income encrypted AES-256-GCM at app layer; Postgres RLS enforces multi-tenancy.

**Tech Stack:** Go 1.22+ · Echo v4 · pgx/v5 · sqlc · golang-migrate · golang-jwt/jwt · viper · zerolog · go-playground/validator · google/uuid · testify · sqlmock · gocron · Anthropic Claude API (REST) · PostgreSQL (Supabase) · Railway (deploy).

---

## Context

This is a greenfield Go backend for Fintrack — a "gym app for your money" targeting Indonesian fresh workers. The PRD (`full_doc.html` at repo root) defines the product, full DB schema (9 tables + 6 enums + RLS policies), and 25 REST endpoints. The frontend is a separate React PWA that will live under `web/`. Deploy target is Railway with ~$20/mo total infra budget.

The plan is organized into 16 phases. Each phase produces independently testable, committable progress. Phases 1–4 build foundation (project, DB, server, auth); 5–14 implement domains; 15–16 cover worker, deploy, and polish. Follow TDD: failing test → minimal impl → green → commit. Repository layer is tested with sqlmock; domain/service with pure unit tests; handlers via httptest. Defer real Postgres integration tests until after MVP.

**Reusable patterns to apply throughout:**
- DTO structs with `validate:` tags + `c.Bind` + central validator
- Service constructors take repo interfaces; never sql/pgx directly
- Errors flow up as `*apperror.Error`; one Echo error handler maps to JSON envelope
- All money is `int64` Rupiah; never float64
- `time.Time` in UTC, formatted RFC3339 at the JSON boundary

---

## File Structure

This mirrors PRD §06 with concrete file responsibilities locked in.

```
fintrack/
├── apps/
│   ├── api/main.go                 # bootstrap REST server
│   └── worker/main.go              # cron jobs (weekly report)
├── internal/
│   ├── config/config.go            # Config struct + viper loader
│   ├── server/server.go            # Echo init, global middleware, error handler
│   ├── server/routes.go            # route table → handlers
│   ├── middleware/auth.go          # JWT (Supabase) validation
│   ├── middleware/apitoken.go      # BYOA bearer token validation
│   ├── middleware/logger.go        # request log
│   ├── middleware/requestid.go     # uuid per request
│   ├── domain/
│   │   ├── user/{entity,repository,service}.go
│   │   ├── category/{entity,repository,service}.go
│   │   ├── budget/{entity,repository,service,engine}.go
│   │   ├── transaction/{entity,repository,service}.go
│   │   ├── fatigue/{entity,service}.go            # no repo — derives from budget+tx
│   │   ├── debt/{entity,repository,service}.go
│   │   ├── goal/{entity,repository,service}.go
│   │   ├── report/{entity,repository,service}.go
│   │   └── token/{entity,repository,service}.go
│   ├── handler/
│   │   ├── onboarding_handler.go
│   │   ├── profile_handler.go
│   │   ├── budget_handler.go
│   │   ├── category_handler.go
│   │   ├── transaction_handler.go
│   │   ├── fatigue_handler.go
│   │   ├── debt_handler.go
│   │   ├── goal_handler.go
│   │   ├── report_handler.go
│   │   ├── token_handler.go
│   │   ├── health_handler.go
│   │   └── dto/{*}.go              # request/response shapes per resource
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── category_repo.go
│   │   ├── budget_repo.go
│   │   ├── transaction_repo.go
│   │   ├── debt_repo.go
│   │   ├── goal_repo.go
│   │   ├── report_repo.go
│   │   └── token_repo.go
│   ├── ai/
│   │   ├── client.go               # generic Anthropic HTTP client
│   │   ├── categorizer.go          # receipt image → {amount, category, note}
│   │   └── summarizer.go           # weekly stats → narrative
│   ├── encryption/aes.go           # EncryptIncome / DecryptIncome / MaskIncome
│   └── worker/
│       ├── scheduler.go            # gocron registration
│       └── weekly_report.go        # Monday 7am job
├── pkg/
│   ├── apperror/errors.go          # typed error codes
│   ├── response/response.go        # JSON envelope helpers
│   ├── logger/logger.go            # zerolog wrapper
│   └── validator/validator.go      # echo Validator interface impl
├── database/
│   ├── migrations/0000NN_*.{up,down}.sql
│   └── sqlc/
│       ├── sqlc.yaml
│       ├── query/{user,category,budget,transaction,debt,goal,report,token}.sql
│       └── generated/              # sqlc output, do not hand-edit
├── scripts/{migrate.sh,generate.sh}
├── .env.example
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

Files that change together stay together (handler/dto next to handler; repo next to its sqlc query file). Domain knows nothing about HTTP or pgx.

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

## Phase 1 — Foundation

### Task 1: Project skeleton

**Files:**
- Create: `go.mod`, `Makefile`, `.env.example`, `.gitignore`, `README.md`
- Create: `apps/api/main.go`, `apps/worker/main.go` (stubs)

- [ ] **Step 1: Init Go module**

```bash
cd <repo-root>
go mod init github.com/<owner>/fintrack
```

- [ ] **Step 2: Create `.gitignore`**

```gitignore
.env
*.local.env
/bin
/tmp
/database/sqlc/generated
*.test
*.out
```

- [ ] **Step 3: Create `.env.example`**

```dotenv
APP_ENV=development
HTTP_PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/fintrack?sslmode=disable
SUPABASE_JWT_SECRET=replace-me
SUPABASE_JWT_AUDIENCE=authenticated
INCOME_ENCRYPTION_KEY=base64-32-bytes-here
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-haiku-4-5-20251001
LOG_LEVEL=debug
```

- [ ] **Step 4: Create `Makefile`**

```makefile
.PHONY: run run-worker test test-race tidy lint sqlc migrate-up migrate-down migrate-create

run:
	go run ./apps/api

run-worker:
	go run ./apps/worker

test:
	go test ./...

test-race:
	go test -race ./...

tidy:
	go mod tidy

sqlc:
	sqlc generate -f database/sqlc/sqlc.yaml

migrate-up:
	migrate -path database/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path database/migrations -database "$$DATABASE_URL" down 1

migrate-create:
	@read -p "Name: " name; migrate create -ext sql -dir database/migrations -seq $$name
```

- [ ] **Step 5: Stub `apps/api/main.go`**

```go
package main

import "fmt"

func main() {
	fmt.Println("fintrack api: not implemented yet")
}
```

- [ ] **Step 6: Stub `apps/worker/main.go`**

```go
package main

import "fmt"

func main() {
	fmt.Println("fintrack worker: not implemented yet")
}
```

- [ ] **Step 7: Verify build**

Run: `go build ./...`
Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add .
git commit -m "chore: bootstrap go module + project skeleton"
```

### Task 2: Config loader

**Files:**
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Add viper dep**

```bash
go get github.com/spf13/viper
```

- [ ] **Step 2: Write failing test `internal/config/config_test.go`**

```go
package config_test

import (
	"os"
	"testing"

	"github.com/<owner>/fintrack/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("SUPABASE_JWT_SECRET", "s")
	t.Setenv("INCOME_ENCRYPTION_KEY", "k")
	t.Setenv("ANTHROPIC_API_KEY", "a")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "test", cfg.AppEnv)
	require.Equal(t, 9090, cfg.HTTPPort)
	require.Equal(t, "postgres://x", cfg.DatabaseURL)
}

func TestLoad_MissingRequired(t *testing.T) {
	os.Clearenv()
	_, err := config.Load()
	require.Error(t, err)
}
```

- [ ] **Step 3: Add testify dep**

```bash
go get github.com/stretchr/testify
```

- [ ] **Step 4: Run test, expect FAIL**

Run: `go test ./internal/config/...`
Expected: package not found / `Load undefined`.

- [ ] **Step 5: Implement `internal/config/config.go`**

```go
package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv               string
	HTTPPort             int
	DatabaseURL          string
	SupabaseJWTSecret    string
	SupabaseJWTAudience  string
	IncomeEncryptionKey  string
	AnthropicAPIKey      string
	AnthropicModel       string
	LogLevel             string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("HTTP_PORT", 8080)
	v.SetDefault("SUPABASE_JWT_AUDIENCE", "authenticated")
	v.SetDefault("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001")
	v.SetDefault("LOG_LEVEL", "info")

	cfg := &Config{
		AppEnv:              v.GetString("APP_ENV"),
		HTTPPort:            v.GetInt("HTTP_PORT"),
		DatabaseURL:         v.GetString("DATABASE_URL"),
		SupabaseJWTSecret:   v.GetString("SUPABASE_JWT_SECRET"),
		SupabaseJWTAudience: v.GetString("SUPABASE_JWT_AUDIENCE"),
		IncomeEncryptionKey: v.GetString("INCOME_ENCRYPTION_KEY"),
		AnthropicAPIKey:     v.GetString("ANTHROPIC_API_KEY"),
		AnthropicModel:      v.GetString("ANTHROPIC_MODEL"),
		LogLevel:            v.GetString("LOG_LEVEL"),
	}

	if cfg.DatabaseURL == "" || cfg.SupabaseJWTSecret == "" ||
		cfg.IncomeEncryptionKey == "" || cfg.AnthropicAPIKey == "" {
		return nil, errors.New("missing required env: DATABASE_URL, SUPABASE_JWT_SECRET, INCOME_ENCRYPTION_KEY, ANTHROPIC_API_KEY")
	}
	return cfg, nil
}
```

- [ ] **Step 6: Run test, expect PASS**

Run: `go test ./internal/config/...`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/config go.mod go.sum
git commit -m "feat(config): viper-backed config loader with required-env validation"
```

### Task 3: Logger package

**Files:**
- Create: `pkg/logger/logger.go`

- [ ] **Step 1: Add zerolog**

```bash
go get github.com/rs/zerolog
```

- [ ] **Step 2: Implement logger**

```go
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(level string) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}
```

- [ ] **Step 3: Verify**

Run: `go build ./...`
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add pkg/logger go.mod go.sum
git commit -m "feat(logger): zerolog initializer"
```

### Task 4: AppError package

**Files:**
- Create: `pkg/apperror/errors.go`
- Test: `pkg/apperror/errors_test.go`

- [ ] **Step 1: Write failing test**

```go
package apperror_test

import (
	"errors"
	"testing"

	"github.com/<owner>/fintrack/pkg/apperror"
	"github.com/stretchr/testify/require"
)

func TestNotFoundIs(t *testing.T) {
	e := apperror.NotFound("user", "id=x")
	require.Equal(t, apperror.CodeNotFound, e.Code)
	require.True(t, errors.Is(e, apperror.ErrNotFound))
}

func TestValidationFields(t *testing.T) {
	e := apperror.Validation("bad", map[string]string{"amount": "must be > 0"})
	require.Equal(t, "bad", e.Message)
	require.Equal(t, "must be > 0", e.Fields["amount"])
}
```

- [ ] **Step 2: Run test, expect FAIL**

Run: `go test ./pkg/apperror/...`

- [ ] **Step 3: Implement `pkg/apperror/errors.go`**

```go
package apperror

import "errors"

type Code string

const (
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeInternal     Code = "INTERNAL_ERROR"
	CodeAI           Code = "AI_ERROR"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
)

type Error struct {
	Code    Code
	Message string
	Fields  map[string]string
	cause   error
	HTTP    int
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.cause }

func wrap(code Code, http int, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, HTTP: http, cause: cause}
}

func NotFound(resource, detail string) *Error {
	return wrap(CodeNotFound, 404, resource+" not found: "+detail, ErrNotFound)
}
func Unauthorized(msg string) *Error { return wrap(CodeUnauthorized, 401, msg, ErrUnauthorized) }
func Forbidden(msg string) *Error    { return wrap(CodeForbidden, 403, msg, ErrForbidden) }
func Conflict(msg string) *Error     { return wrap(CodeConflict, 409, msg, ErrConflict) }
func Internal(cause error) *Error    { return wrap(CodeInternal, 500, "internal error", cause) }
func AI(msg string, cause error) *Error { return wrap(CodeAI, 502, msg, cause) }

func Validation(msg string, fields map[string]string) *Error {
	e := wrap(CodeValidation, 400, msg, nil)
	e.Fields = fields
	return e
}
```

- [ ] **Step 4: Run test, expect PASS**

Run: `go test ./pkg/apperror/...`

- [ ] **Step 5: Commit**

```bash
git add pkg/apperror
git commit -m "feat(apperror): typed error codes with HTTP mapping"
```

### Task 5: Response envelope

**Files:**
- Create: `pkg/response/response.go`

- [ ] **Step 1: Implement**

```go
package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/<owner>/fintrack/pkg/apperror"
)

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Total     int    `json:"total,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
}

type Envelope struct {
	Data any   `json:"data,omitempty"`
	Meta *Meta `json:"meta,omitempty"`
}

type ErrorEnvelope struct {
	Error errBody `json:"error"`
}
type errBody struct {
	Code    apperror.Code     `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, Envelope{Data: data, Meta: meta(c)})
}
func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, Envelope{Data: data, Meta: meta(c)})
}
func List(c echo.Context, data any, total, page, perPage int) error {
	m := meta(c)
	m.Total, m.Page, m.PerPage = total, page, perPage
	return c.JSON(http.StatusOK, Envelope{Data: data, Meta: m})
}
func meta(c echo.Context) *Meta {
	id, _ := c.Get("request_id").(string)
	return &Meta{RequestID: id}
}

func Error(c echo.Context, e *apperror.Error) error {
	if e.HTTP == 0 {
		e.HTTP = http.StatusInternalServerError
	}
	return c.JSON(e.HTTP, ErrorEnvelope{Error: errBody{
		Code: e.Code, Message: e.Message, Fields: e.Fields,
	}})
}
```

- [ ] **Step 2: Add Echo dep**

```bash
go get github.com/labstack/echo/v4
```

- [ ] **Step 3: Verify build**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```bash
git add pkg/response go.mod go.sum
git commit -m "feat(response): JSON envelope helpers (data/meta + error)"
```

### Task 6: Validator wrapper

**Files:**
- Create: `pkg/validator/validator.go`

- [ ] **Step 1: Add validator dep**

```bash
go get github.com/go-playground/validator/v10
```

- [ ] **Step 2: Implement**

```go
package validator

import (
	"strings"

	gpv "github.com/go-playground/validator/v10"
	"github.com/<owner>/fintrack/pkg/apperror"
)

type V struct{ v *gpv.Validate }

func New() *V { return &V{v: gpv.New()} }

func (x *V) Validate(i any) error { return x.v.Struct(i) }

func ToAppError(err error) *apperror.Error {
	if err == nil {
		return nil
	}
	verrs, ok := err.(gpv.ValidationErrors)
	if !ok {
		return apperror.Validation(err.Error(), nil)
	}
	fields := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		fields[strings.ToLower(fe.Field())] = fe.Tag()
	}
	return apperror.Validation("validation failed", fields)
}
```

- [ ] **Step 3: Verify build & commit**

Run: `go build ./...`

```bash
git add pkg/validator go.mod go.sum
git commit -m "feat(validator): playground validator wrapper + apperror mapping"
```

---

## Phase 2 — Database

### Task 7: Migrations directory + first SQL

**Files:**
- Create: `database/migrations/000001_init_extensions_and_enums.up.sql`
- Create: `database/migrations/000001_init_extensions_and_enums.down.sql`
- Create: `scripts/migrate.sh`

- [ ] **Step 1: Install migrate CLI locally** (developer machine, not committed)

```bash
brew install golang-migrate
```

- [ ] **Step 2: Write `000001_init_extensions_and_enums.up.sql`**

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE expense_category_type AS ENUM ('fixed','variable','debt','want');
CREATE TYPE financial_program     AS ENUM ('pondasi','bebas_utang','goal_chaser','tumbuh','seimbang');
CREATE TYPE fatigue_status        AS ENUM ('fresh','warning','fatigued');
CREATE TYPE debt_method           AS ENUM ('snowball','avalanche');
CREATE TYPE lifestyle_style       AS ENUM ('easy','balanced','strict');
CREATE TYPE housing_type          AS ENUM ('kosan','kpr','keluarga');
```

- [ ] **Step 3: Write `000001_init_extensions_and_enums.down.sql`**

```sql
DROP TYPE IF EXISTS housing_type;
DROP TYPE IF EXISTS lifestyle_style;
DROP TYPE IF EXISTS debt_method;
DROP TYPE IF EXISTS fatigue_status;
DROP TYPE IF EXISTS financial_program;
DROP TYPE IF EXISTS expense_category_type;
```

- [ ] **Step 4: Write `scripts/migrate.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail
cmd=${1:-up}
migrate -path database/migrations -database "${DATABASE_URL}" "$cmd"
```

```bash
chmod +x scripts/migrate.sh
```

- [ ] **Step 5: Run migration locally**

Run: `DATABASE_URL=$DATABASE_URL ./scripts/migrate.sh up`
Expected: applies version 1.

- [ ] **Step 6: Commit**

```bash
git add database/migrations scripts/migrate.sh
git commit -m "feat(db): migration 0001 — uuid extension + enums"
```

### Task 8: user_profiles + expense_categories migrations

**Files:**
- Create: `database/migrations/000002_user_profiles.up.sql` / `.down.sql`
- Create: `database/migrations/000003_expense_categories.up.sql` / `.down.sql`

- [ ] **Step 1: 0002 up**

```sql
CREATE TABLE user_profiles (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL UNIQUE,
    income_encrypted    TEXT,
    income_hint         VARCHAR(20),
    housing_type        housing_type,
    lifestyle_style     lifestyle_style,
    emergency_months    SMALLINT NOT NULL DEFAULT 0,
    active_program      financial_program,
    onboarding_done     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 2: 0002 down**

```sql
DROP TABLE IF EXISTS user_profiles;
```

- [ ] **Step 3: 0003 up**

```sql
CREATE TABLE expense_categories (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID,
    name        VARCHAR(100) NOT NULL,
    icon        VARCHAR(10),
    type        expense_category_type NOT NULL,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  SMALLINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_expense_categories_user ON expense_categories (user_id);
```

- [ ] **Step 4: 0003 down**

```sql
DROP TABLE IF EXISTS expense_categories;
```

- [ ] **Step 5: Run + commit**

Run: `./scripts/migrate.sh up`

```bash
git add database/migrations
git commit -m "feat(db): migrations 0002-0003 — user_profiles + expense_categories"
```

### Task 9: Budget tables migrations

**Files:**
- Create: `database/migrations/000004_budget_plans.up.sql` / `.down.sql`
- Create: `database/migrations/000005_budget_items.up.sql` / `.down.sql`

- [ ] **Step 1: 0004 up**

```sql
CREATE TABLE budget_plans (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL,
    period_year  SMALLINT NOT NULL,
    period_month SMALLINT NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    total_income BIGINT NOT NULL,
    program      financial_program NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, period_year, period_month)
);
CREATE INDEX idx_budget_plans_user_period ON budget_plans (user_id, period_year, period_month);
```

- [ ] **Step 2: 0004 down**

```sql
DROP TABLE IF EXISTS budget_plans;
```

- [ ] **Step 3: 0005 up**

```sql
CREATE TABLE budget_items (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    budget_plan_id   UUID NOT NULL REFERENCES budget_plans(id) ON DELETE CASCADE,
    category_id      UUID NOT NULL REFERENCES expense_categories(id),
    allocated_amount BIGINT NOT NULL,
    percentage       NUMERIC(5,2),
    is_debt_focus    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (budget_plan_id, category_id)
);
```

- [ ] **Step 4: 0005 down**

```sql
DROP TABLE IF EXISTS budget_items;
```

- [ ] **Step 5: Run + commit**

```bash
./scripts/migrate.sh up
git add database/migrations
git commit -m "feat(db): budget_plans + budget_items"
```

### Task 10: Transactions migration

**Files:**
- Create: `database/migrations/000006_transactions.up.sql` / `.down.sql`

- [ ] **Step 1: up**

```sql
CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    budget_plan_id  UUID REFERENCES budget_plans(id),
    category_id     UUID NOT NULL REFERENCES expense_categories(id),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    note            TEXT,
    receipt_url     TEXT,
    ai_categorized  BOOLEAN NOT NULL DEFAULT FALSE,
    ai_confidence   NUMERIC(3,2),
    transacted_at   TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_transactions_user_period   ON transactions (user_id, transacted_at DESC);
CREATE INDEX idx_transactions_user_category ON transactions (user_id, category_id, transacted_at DESC);
```

- [ ] **Step 2: down**

```sql
DROP TABLE IF EXISTS transactions;
```

- [ ] **Step 3: Run + commit**

```bash
./scripts/migrate.sh up
git add database/migrations
git commit -m "feat(db): transactions table + indexes"
```

### Task 11: Debt items migration

**Files:**
- Create: `database/migrations/000007_debt_items.up.sql` / `.down.sql`

- [ ] **Step 1: up**

```sql
CREATE TABLE debt_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    category_id     UUID REFERENCES expense_categories(id),
    name            VARCHAR(100) NOT NULL,
    total_amount    BIGINT NOT NULL,
    current_balance BIGINT NOT NULL,
    interest_rate   NUMERIC(5,2) NOT NULL,
    min_payment     BIGINT NOT NULL,
    method          debt_method NOT NULL DEFAULT 'snowball',
    priority        SMALLINT NOT NULL DEFAULT 1,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    started_at      DATE NOT NULL,
    target_paid_at  DATE,
    paid_at         DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_debt_items_user_active ON debt_items (user_id, is_active, priority);
```

- [ ] **Step 2: down + run + commit**

```sql
-- down
DROP TABLE IF EXISTS debt_items;
```

```bash
./scripts/migrate.sh up
git add database/migrations && git commit -m "feat(db): debt_items"
```

### Task 12: Goals migration

**Files:**
- Create: `database/migrations/000008_goals.up.sql` / `.down.sql`

- [ ] **Step 1: up**

```sql
CREATE TABLE goals (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    name            VARCHAR(100) NOT NULL,
    icon            VARCHAR(10),
    target_amount   BIGINT NOT NULL,
    current_amount  BIGINT NOT NULL DEFAULT 0,
    target_date     DATE,
    is_completed    BOOLEAN NOT NULL DEFAULT FALSE,
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_goals_user ON goals (user_id, is_primary);
```

- [ ] **Step 2: down + run + commit**

```sql
-- down
DROP TABLE IF EXISTS goals;
```

```bash
./scripts/migrate.sh up
git add database/migrations && git commit -m "feat(db): goals"
```

### Task 13: Weekly reports + API tokens

**Files:**
- Create: `database/migrations/000009_weekly_reports.up.sql` / `.down.sql`
- Create: `database/migrations/000010_api_tokens.up.sql` / `.down.sql`

- [ ] **Step 1: 0009 up**

```sql
CREATE TABLE weekly_reports (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL,
    week_start    DATE NOT NULL,
    week_end      DATE NOT NULL,
    total_spent   BIGINT NOT NULL DEFAULT 0,
    total_budget  BIGINT NOT NULL DEFAULT 0,
    narrative     TEXT,
    generated_at  TIMESTAMPTZ,
    email_sent_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, week_start)
);
```

- [ ] **Step 2: 0010 up**

```sql
CREATE TABLE api_tokens (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL,
    name         VARCHAR(100) NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,
    token_hint   VARCHAR(10),
    can_read     BOOLEAN NOT NULL DEFAULT TRUE,
    can_write    BOOLEAN NOT NULL DEFAULT FALSE,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 3: down for both**

```sql
-- 0009 down
DROP TABLE IF EXISTS weekly_reports;
-- 0010 down
DROP TABLE IF EXISTS api_tokens;
```

- [ ] **Step 4: Run + commit**

```bash
./scripts/migrate.sh up
git add database/migrations && git commit -m "feat(db): weekly_reports + api_tokens"
```

### Task 14: RLS policies + seed system categories

**Files:**
- Create: `database/migrations/000011_rls.up.sql` / `.down.sql`
- Create: `database/migrations/000012_seed_categories.up.sql` / `.down.sql`

- [ ] **Step 1: 0011 up**

```sql
ALTER TABLE user_profiles      ENABLE ROW LEVEL SECURITY;
ALTER TABLE expense_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_plans       ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_items       ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions       ENABLE ROW LEVEL SECURITY;
ALTER TABLE debt_items         ENABLE ROW LEVEL SECURITY;
ALTER TABLE goals              ENABLE ROW LEVEL SECURITY;
ALTER TABLE weekly_reports     ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_tokens         ENABLE ROW LEVEL SECURITY;

CREATE POLICY "user owns user_profiles" ON user_profiles
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns budget_plans" ON budget_plans
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns budget_items via plan" ON budget_items
  FOR ALL USING (EXISTS (
    SELECT 1 FROM budget_plans bp WHERE bp.id = budget_plan_id AND bp.user_id = auth.uid()
  ));
CREATE POLICY "user owns transactions" ON transactions
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns debt_items" ON debt_items
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns goals" ON goals
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns weekly_reports" ON weekly_reports
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns api_tokens" ON api_tokens
  FOR ALL USING (user_id = auth.uid());

CREATE POLICY "user reads system + own categories" ON expense_categories
  FOR SELECT USING (user_id IS NULL OR user_id = auth.uid());
CREATE POLICY "user inserts own categories" ON expense_categories
  FOR INSERT WITH CHECK (user_id = auth.uid());
CREATE POLICY "user updates own categories" ON expense_categories
  FOR UPDATE USING (user_id = auth.uid());
CREATE POLICY "user deletes own categories" ON expense_categories
  FOR DELETE USING (user_id = auth.uid());
```

- [ ] **Step 2: 0011 down**

```sql
DROP POLICY IF EXISTS "user owns user_profiles" ON user_profiles;
DROP POLICY IF EXISTS "user owns budget_plans" ON budget_plans;
DROP POLICY IF EXISTS "user owns budget_items via plan" ON budget_items;
DROP POLICY IF EXISTS "user owns transactions" ON transactions;
DROP POLICY IF EXISTS "user owns debt_items" ON debt_items;
DROP POLICY IF EXISTS "user owns goals" ON goals;
DROP POLICY IF EXISTS "user owns weekly_reports" ON weekly_reports;
DROP POLICY IF EXISTS "user owns api_tokens" ON api_tokens;
DROP POLICY IF EXISTS "user reads system + own categories" ON expense_categories;
DROP POLICY IF EXISTS "user inserts own categories" ON expense_categories;
DROP POLICY IF EXISTS "user updates own categories" ON expense_categories;
DROP POLICY IF EXISTS "user deletes own categories" ON expense_categories;
ALTER TABLE user_profiles      DISABLE ROW LEVEL SECURITY;
ALTER TABLE expense_categories DISABLE ROW LEVEL SECURITY;
ALTER TABLE budget_plans       DISABLE ROW LEVEL SECURITY;
ALTER TABLE budget_items       DISABLE ROW LEVEL SECURITY;
ALTER TABLE transactions       DISABLE ROW LEVEL SECURITY;
ALTER TABLE debt_items         DISABLE ROW LEVEL SECURITY;
ALTER TABLE goals              DISABLE ROW LEVEL SECURITY;
ALTER TABLE weekly_reports     DISABLE ROW LEVEL SECURITY;
ALTER TABLE api_tokens         DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 3: 0012 seed system categories (user_id NULL)**

```sql
INSERT INTO expense_categories (user_id, name, icon, type, is_default, sort_order) VALUES
  (NULL, 'Sewa kosan',     '🏠', 'fixed',    true, 10),
  (NULL, 'Cicilan KPR',    '🏗', 'fixed',    true, 11),
  (NULL, 'Listrik & air',  '💡', 'fixed',    true, 12),
  (NULL, 'Transportasi',   '🛵', 'fixed',    true, 13),
  (NULL, 'Internet & HP',  '📶', 'fixed',    true, 14),
  (NULL, 'Makan & minum',  '🍱', 'variable', true, 20),
  (NULL, 'Belanja harian', '🛒', 'variable', true, 21),
  (NULL, 'Hiburan',        '🎬', 'want',     true, 30),
  (NULL, 'Self-care',      '💅', 'want',     true, 31),
  (NULL, 'Kartu kredit',   '💳', 'debt',     true, 40),
  (NULL, 'Paylater',       '💸', 'debt',     true, 41),
  (NULL, 'Tabungan',       '🐷', 'fixed',    true, 50);
```

- [ ] **Step 4: 0012 down**

```sql
DELETE FROM expense_categories WHERE user_id IS NULL AND is_default = true;
```

- [ ] **Step 5: Run + commit**

```bash
./scripts/migrate.sh up
git add database/migrations
git commit -m "feat(db): RLS policies + seed system default categories"
```

### Task 15: sqlc setup

**Files:**
- Create: `database/sqlc/sqlc.yaml`
- Create: `database/sqlc/query/user.sql`
- Create: `scripts/generate.sh`

- [ ] **Step 1: Install sqlc CLI** (developer machine)

```bash
brew install sqlc
```

- [ ] **Step 2: `database/sqlc/sqlc.yaml`**

```yaml
version: "2"
sql:
  - engine: postgresql
    queries: database/sqlc/query
    schema: database/migrations
    gen:
      go:
        package: db
        out: database/sqlc/generated
        sql_package: pgx/v5
        emit_json_tags: false
        emit_pointers_for_null_types: true
        overrides:
          - db_type: housing_type
            go_type: string
          - db_type: lifestyle_style
            go_type: string
          - db_type: financial_program
            go_type: string
          - db_type: expense_category_type
            go_type: string
          - db_type: debt_method
            go_type: string
          - db_type: fatigue_status
            go_type: string
```

- [ ] **Step 3: First query — `database/sqlc/query/user.sql`**

```sql
-- name: GetUserProfileByUserID :one
SELECT * FROM user_profiles WHERE user_id = $1;

-- name: CreateUserProfile :one
INSERT INTO user_profiles (
    user_id, income_encrypted, income_hint, housing_type, lifestyle_style,
    emergency_months, active_program, onboarding_done
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE user_profiles SET
    lifestyle_style  = COALESCE(sqlc.narg('lifestyle_style'), lifestyle_style),
    emergency_months = COALESCE(sqlc.narg('emergency_months'), emergency_months),
    active_program   = COALESCE(sqlc.narg('active_program'), active_program),
    onboarding_done  = COALESCE(sqlc.narg('onboarding_done'), onboarding_done),
    updated_at       = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserIncome :one
UPDATE user_profiles SET
    income_encrypted = $2,
    income_hint      = $3,
    updated_at       = NOW()
WHERE user_id = $1
RETURNING income_hint;
```

- [ ] **Step 4: `scripts/generate.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail
sqlc generate -f database/sqlc/sqlc.yaml
```

```bash
chmod +x scripts/generate.sh
```

- [ ] **Step 5: Generate**

Run: `./scripts/generate.sh`
Expected: files appear under `database/sqlc/generated/`.

- [ ] **Step 6: Add pgx + commit (do NOT commit `generated/` — it's gitignored)**

```bash
go get github.com/jackc/pgx/v5
git add database/sqlc/sqlc.yaml database/sqlc/query scripts/generate.sh go.mod go.sum
git commit -m "feat(db): sqlc config + initial user queries"
```

### Task 16: pgx connection pool

**Files:**
- Create: `internal/database/db.go`

- [ ] **Step 1: Implement**

```go
package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.HealthCheckPeriod = 30 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
```

- [ ] **Step 2: Verify build & commit**

```bash
go build ./...
git add internal/database go.sum
git commit -m "feat(db): pgx connection pool with health check"
```

---

## Phase 3 — HTTP Server

### Task 17: Echo bootstrap + middleware skeletons

**Files:**
- Create: `internal/server/server.go`
- Create: `internal/middleware/requestid.go`
- Create: `internal/middleware/logger.go`

- [ ] **Step 1: `internal/middleware/requestid.go`**

```go
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get("X-Request-ID")
			if id == "" {
				id = uuid.NewString()
			}
			c.Set("request_id", id)
			c.Response().Header().Set("X-Request-ID", id)
			return next(c)
		}
	}
}
```

- [ ] **Step 2: `internal/middleware/logger.go`**

```go
package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			req := c.Request()
			id, _ := c.Get("request_id").(string)
			log.Info().
				Str("request_id", id).
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", c.Response().Status).
				Dur("latency", time.Since(start)).
				Msg("http")
			return err
		}
	}
}
```

- [ ] **Step 3: Add uuid dep**

```bash
go get github.com/google/uuid
```

- [ ] **Step 4: `internal/server/server.go`**

```go
package server

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/<owner>/fintrack/internal/middleware"
	"github.com/<owner>/fintrack/pkg/apperror"
	"github.com/<owner>/fintrack/pkg/response"
	v "github.com/<owner>/fintrack/pkg/validator"
)

type Deps struct {
	// filled in later phases
}

func New(deps Deps) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = v.New()

	e.Use(echomw.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, "X-Request-ID"},
	}))

	e.HTTPErrorHandler = errorHandler
	registerRoutes(e, deps)
	return e
}

func errorHandler(err error, c echo.Context) {
	var ae *apperror.Error
	if errors.As(err, &ae) {
		_ = response.Error(c, ae)
		return
	}
	if he, ok := err.(*echo.HTTPError); ok {
		_ = response.Error(c, &apperror.Error{
			Code: apperror.CodeInternal, Message: he.Error(), HTTP: he.Code,
		})
		return
	}
	_ = response.Error(c, apperror.Internal(err))
}
```

- [ ] **Step 5: Stub `internal/server/routes.go`**

```go
package server

import "github.com/labstack/echo/v4"

func registerRoutes(e *echo.Echo, _ Deps) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]any{
			"data": map[string]string{"status": "ok", "version": "0.1.0", "db": "ok"},
		})
	})
}
```

- [ ] **Step 6: Wire bootstrap in `apps/api/main.go`**

```go
package main

import (
	"context"
	"fmt"

	"github.com/<owner>/fintrack/internal/config"
	"github.com/<owner>/fintrack/internal/database"
	"github.com/<owner>/fintrack/internal/server"
	"github.com/<owner>/fintrack/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger.Init(cfg.LogLevel)

	pool, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	e := server.New(server.Deps{})
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	if err := e.Start(addr); err != nil {
		panic(err)
	}
}
```

- [ ] **Step 7: Run server, smoke test**

Run: `go run ./apps/api`
Then: `curl localhost:8080/health`
Expected: `{"data":{"status":"ok","version":"0.1.0","db":"ok"}}`

- [ ] **Step 8: Commit**

```bash
git add internal/server internal/middleware apps/api/main.go go.mod go.sum
git commit -m "feat(server): echo bootstrap + request id + logger + error handler"
```

---

## Phase 4 — Auth & Encryption

### Task 18: AES-256-GCM income encryption

**Files:**
- Create: `internal/encryption/aes.go`
- Test: `internal/encryption/aes_test.go`

- [ ] **Step 1: Failing test**

```go
package encryption_test

import (
	"encoding/base64"
	"testing"

	"github.com/<owner>/fintrack/internal/encryption"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	enc, err := encryption.New(key)
	require.NoError(t, err)

	cipher, err := enc.EncryptIncome(8_000_000)
	require.NoError(t, err)
	require.NotEmpty(t, cipher)

	got, err := enc.DecryptIncome(cipher)
	require.NoError(t, err)
	require.Equal(t, int64(8_000_000), got)
}

func TestMaskIncome(t *testing.T) {
	require.Equal(t, "Rp 8jt", encryption.MaskIncome(8_000_000))
	require.Equal(t, "Rp 12jt", encryption.MaskIncome(12_500_000))
	require.Equal(t, "Rp 950rb", encryption.MaskIncome(950_000))
}

func TestEncryptIsNonDeterministic(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	enc, _ := encryption.New(key)
	a, _ := enc.EncryptIncome(1_000_000)
	b, _ := enc.EncryptIncome(1_000_000)
	require.NotEqual(t, a, b, "GCM nonce must randomize ciphertext")
}
```

- [ ] **Step 2: Run, expect FAIL**

Run: `go test ./internal/encryption/...`

- [ ] **Step 3: Implement**

```go
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

type Encryptor struct{ aead cipher.AEAD }

func New(b64Key string) (*Encryptor, error) {
	key, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil || len(key) != 32 {
		return nil, errors.New("INCOME_ENCRYPTION_KEY must be base64-encoded 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Encryptor{aead: aead}, nil
}

func (e *Encryptor) EncryptIncome(amount int64) (string, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	plain := []byte(strconv.FormatInt(amount, 10))
	ct := e.aead.Seal(nil, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(append(nonce, ct...)), nil
}

func (e *Encryptor) DecryptIncome(b64 string) (int64, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return 0, err
	}
	ns := e.aead.NonceSize()
	if len(raw) < ns {
		return 0, errors.New("ciphertext too short")
	}
	plain, err := e.aead.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(plain), 10, 64)
}

func MaskIncome(amount int64) string {
	switch {
	case amount >= 1_000_000:
		return fmt.Sprintf("Rp %djt", amount/1_000_000)
	case amount >= 1_000:
		return fmt.Sprintf("Rp %drb", amount/1_000)
	default:
		return fmt.Sprintf("Rp %d", amount)
	}
}
```

- [ ] **Step 4: Run, expect PASS**

Run: `go test ./internal/encryption/...`

- [ ] **Step 5: Commit**

```bash
git add internal/encryption
git commit -m "feat(encryption): AES-256-GCM income encrypt/decrypt + mask"
```

### Task 19: JWT auth middleware (Supabase)

**Files:**
- Create: `internal/middleware/auth.go`
- Test: `internal/middleware/auth_test.go`

- [ ] **Step 1: Add jwt dep**

```bash
go get github.com/golang-jwt/jwt/v5
```

- [ ] **Step 2: Failing test**

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/<owner>/fintrack/internal/middleware"
)

func TestJWT_ValidToken(t *testing.T) {
	secret := "test-secret"
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "123e4567-e89b-12d3-a456-426614174000",
		"aud": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, _ := tok.SignedString([]byte(secret))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	h := middleware.JWT(secret, "authenticated")(func(c echo.Context) error {
		called = true
		require.Equal(t, "123e4567-e89b-12d3-a456-426614174000", c.Get("user_id"))
		return c.NoContent(200)
	})
	require.NoError(t, h(c))
	require.True(t, called)
}

func TestJWT_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := middleware.JWT("s", "authenticated")(func(c echo.Context) error { return nil })(c)
	require.Error(t, err)
}
```

- [ ] **Step 3: Implement `internal/middleware/auth.go`**

```go
package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/<owner>/fintrack/pkg/apperror"
)

func JWT(secret, audience string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				return apperror.Unauthorized("missing bearer token")
			}
			raw := strings.TrimPrefix(h, "Bearer ")
			claims := jwt.MapClaims{}
			tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, apperror.Unauthorized("unexpected signing method")
				}
				return []byte(secret), nil
			})
			if err != nil || !tok.Valid {
				return apperror.Unauthorized("invalid token")
			}
			if aud, ok := claims["aud"].(string); !ok || aud != audience {
				return apperror.Unauthorized("audience mismatch")
			}
			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				return apperror.Unauthorized("missing subject")
			}
			c.Set("user_id", sub)
			return next(c)
		}
	}
}
```

- [ ] **Step 4: Run, expect PASS**

Run: `go test ./internal/middleware/...`

- [ ] **Step 5: Commit**

```bash
git add internal/middleware go.mod go.sum
git commit -m "feat(middleware): JWT auth using Supabase HS256 secret"
```

---

## Phase 5 — User & Profile Domain

### Task 20: User domain entity + repo interface

**Files:**
- Create: `internal/domain/user/entity.go`
- Create: `internal/domain/user/repository.go`

- [ ] **Step 1: `entity.go`**

```go
package user

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	IncomeEncrypted   *string
	IncomeHint        *string
	HousingType       *string
	LifestyleStyle    *string
	EmergencyMonths   int
	ActiveProgram     *string
	OnboardingDone    bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CreateProfileInput struct {
	UserID           uuid.UUID
	IncomeEncrypted  string
	IncomeHint       string
	HousingType      string
	LifestyleStyle   string
	EmergencyMonths  int
	ActiveProgram    string
	OnboardingDone   bool
}
```

- [ ] **Step 2: `repository.go`**

```go
package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	Create(ctx context.Context, in CreateProfileInput) (*Profile, error)
	UpdateLifestyle(ctx context.Context, userID uuid.UUID, lifestyle *string, emergencyMonths *int) (*Profile, error)
	UpdateIncome(ctx context.Context, userID uuid.UUID, encrypted, hint string) (string, error)
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/domain/user
git commit -m "feat(user): domain entity + repository contract"
```

### Task 21: User service

**Files:**
- Create: `internal/domain/user/service.go`
- Test: `internal/domain/user/service_test.go`

- [ ] **Step 1: Failing test (with mock repo)**

```go
package user_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/<owner>/fintrack/internal/domain/user"
)

type repoMock struct{ mock.Mock }

func (m *repoMock) GetByUserID(ctx context.Context, id uuid.UUID) (*user.Profile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *repoMock) Create(ctx context.Context, in user.CreateProfileInput) (*user.Profile, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *repoMock) UpdateLifestyle(ctx context.Context, id uuid.UUID, ls *string, em *int) (*user.Profile, error) {
	args := m.Called(ctx, id, ls, em)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *repoMock) UpdateIncome(ctx context.Context, id uuid.UUID, enc, hint string) (string, error) {
	args := m.Called(ctx, id, enc, hint)
	return args.String(0), args.Error(1)
}

type encMock struct{ mock.Mock }

func (m *encMock) EncryptIncome(amount int64) (string, error) {
	args := m.Called(amount)
	return args.String(0), args.Error(1)
}

func TestUpdateIncome_EncryptsAndReturnsHint(t *testing.T) {
	repo := &repoMock{}
	enc := &encMock{}
	uid := uuid.New()

	enc.On("EncryptIncome", int64(8_000_000)).Return("CIPHER", nil)
	repo.On("UpdateIncome", mock.Anything, uid, "CIPHER", "Rp 8jt").Return("Rp 8jt", nil)

	svc := user.NewService(repo, enc)
	hint, err := svc.UpdateIncome(context.Background(), uid, 8_000_000)
	require.NoError(t, err)
	require.Equal(t, "Rp 8jt", hint)
}
```

- [ ] **Step 2: Implement `service.go`**

```go
package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/<owner>/fintrack/internal/encryption"
)

type IncomeEncryptor interface {
	EncryptIncome(amount int64) (string, error)
}

type Service interface {
	Get(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpdateLifestyle(ctx context.Context, userID uuid.UUID, lifestyle *string, emergencyMonths *int) (*Profile, error)
	UpdateIncome(ctx context.Context, userID uuid.UUID, amount int64) (string, error)
}

type service struct {
	repo Repository
	enc  IncomeEncryptor
}

func NewService(repo Repository, enc IncomeEncryptor) Service {
	return &service{repo: repo, enc: enc}
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*Profile, error) {
	return s.repo.GetByUserID(ctx, id)
}

func (s *service) UpdateLifestyle(ctx context.Context, id uuid.UUID, ls *string, em *int) (*Profile, error) {
	return s.repo.UpdateLifestyle(ctx, id, ls, em)
}

func (s *service) UpdateIncome(ctx context.Context, id uuid.UUID, amount int64) (string, error) {
	cipher, err := s.enc.EncryptIncome(amount)
	if err != nil {
		return "", err
	}
	hint := encryption.MaskIncome(amount)
	return s.repo.UpdateIncome(ctx, id, cipher, hint)
}
```

- [ ] **Step 3: Run + commit**

```bash
go test ./internal/domain/user/...
git add internal/domain/user
git commit -m "feat(user): service layer encrypting income before persist"
```

### Task 22: User repository implementation (sqlc-backed)

**Files:**
- Create: `internal/repository/user_repo.go`

- [ ] **Step 1: Implement using sqlc generated `db.Queries`**

```go
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/<owner>/fintrack/database/sqlc/generated"
	"github.com/<owner>/fintrack/internal/domain/user"
	"github.com/<owner>/fintrack/pkg/apperror"
)

type userRepo struct {
	q *db.Queries
}

func NewUserRepo(pool *pgxpool.Pool) user.Repository {
	return &userRepo{q: db.New(pool)}
}

func (r *userRepo) GetByUserID(ctx context.Context, id uuid.UUID) (*user.Profile, error) {
	row, err := r.q.GetUserProfileByUserID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("user_profile", id.String())
	}
	return toDomain(row), nil
}

func (r *userRepo) Create(ctx context.Context, in user.CreateProfileInput) (*user.Profile, error) {
	row, err := r.q.CreateUserProfile(ctx, db.CreateUserProfileParams{
		UserID:          in.UserID,
		IncomeEncrypted: ptr(in.IncomeEncrypted),
		IncomeHint:      ptr(in.IncomeHint),
		HousingType:     in.HousingType,
		LifestyleStyle:  in.LifestyleStyle,
		EmergencyMonths: int16(in.EmergencyMonths),
		ActiveProgram:   in.ActiveProgram,
		OnboardingDone:  in.OnboardingDone,
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return toDomain(row), nil
}

func (r *userRepo) UpdateLifestyle(ctx context.Context, id uuid.UUID, ls *string, em *int) (*user.Profile, error) {
	var emInt *int16
	if em != nil {
		v := int16(*em)
		emInt = &v
	}
	row, err := r.q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		UserID:          id,
		LifestyleStyle:  ls,
		EmergencyMonths: emInt,
	})
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return toDomain(row), nil
}

func (r *userRepo) UpdateIncome(ctx context.Context, id uuid.UUID, encrypted, hint string) (string, error) {
	out, err := r.q.UpdateUserIncome(ctx, db.UpdateUserIncomeParams{
		UserID:          id,
		IncomeEncrypted: &encrypted,
		IncomeHint:      &hint,
	})
	if err != nil {
		return "", apperror.Internal(err)
	}
	if out == nil {
		return hint, nil
	}
	return *out, nil
}

func toDomain(r db.UserProfile) *user.Profile {
	return &user.Profile{
		ID:              r.ID,
		UserID:          r.UserID,
		IncomeEncrypted: r.IncomeEncrypted,
		IncomeHint:      r.IncomeHint,
		HousingType:     r.HousingType,
		LifestyleStyle:  r.LifestyleStyle,
		EmergencyMonths: int(r.EmergencyMonths),
		ActiveProgram:   r.ActiveProgram,
		OnboardingDone:  r.OnboardingDone,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func ptr[T any](v T) *T { return &v }
```

- [ ] **Step 2: Build + commit**

```bash
go build ./...
git add internal/repository
git commit -m "feat(user): pgx-backed repository via sqlc"
```

### Task 23: Profile handler + DTO + wiring

**Files:**
- Create: `internal/handler/dto/profile_dto.go`
- Create: `internal/handler/profile_handler.go`
- Modify: `internal/server/server.go`, `internal/server/routes.go`, `apps/api/main.go`

- [ ] **Step 1: DTO**

```go
package dto

type ProfileResponse struct {
	ID              string `json:"id"`
	IncomeHint      string `json:"income_hint,omitempty"`
	HousingType     string `json:"housing_type,omitempty"`
	LifestyleStyle  string `json:"lifestyle_style,omitempty"`
	EmergencyMonths int    `json:"emergency_months"`
	ActiveProgram   string `json:"active_program,omitempty"`
	OnboardingDone  bool   `json:"onboarding_done"`
}

type UpdateProfileRequest struct {
	LifestyleStyle  *string `json:"lifestyle_style"  validate:"omitempty,oneof=easy balanced strict"`
	EmergencyMonths *int    `json:"emergency_months" validate:"omitempty,oneof=0 1 3 6"`
}

type UpdateIncomeRequest struct {
	Income int64 `json:"income" validate:"required,gt=0"`
}
```

- [ ] **Step 2: Handler**

```go
package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/<owner>/fintrack/internal/domain/user"
	"github.com/<owner>/fintrack/internal/handler/dto"
	"github.com/<owner>/fintrack/pkg/apperror"
	"github.com/<owner>/fintrack/pkg/response"
	v "github.com/<owner>/fintrack/pkg/validator"
)

type ProfileHandler struct{ Svc user.Service }

func (h *ProfileHandler) Get(c echo.Context) error {
	uid, err := uuid.Parse(c.Get("user_id").(string))
	if err != nil {
		return apperror.Unauthorized("bad user_id")
	}
	p, err := h.Svc.Get(c.Request().Context(), uid)
	if err != nil {
		return err
	}
	return response.OK(c, toProfileResponse(p))
}

func (h *ProfileHandler) Update(c echo.Context) error {
	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return apperror.Validation(err.Error(), nil)
	}
	if err := c.Validate(&req); err != nil {
		return v.ToAppError(err)
	}
	uid, _ := uuid.Parse(c.Get("user_id").(string))
	p, err := h.Svc.UpdateLifestyle(c.Request().Context(), uid, req.LifestyleStyle, req.EmergencyMonths)
	if err != nil {
		return err
	}
	return response.OK(c, toProfileResponse(p))
}

func (h *ProfileHandler) UpdateIncome(c echo.Context) error {
	var req dto.UpdateIncomeRequest
	if err := c.Bind(&req); err != nil {
		return apperror.Validation(err.Error(), nil)
	}
	if err := c.Validate(&req); err != nil {
		return v.ToAppError(err)
	}
	uid, _ := uuid.Parse(c.Get("user_id").(string))
	hint, err := h.Svc.UpdateIncome(c.Request().Context(), uid, req.Income)
	if err != nil {
		return err
	}
	return response.OK(c, map[string]string{"income_hint": hint})
}

func toProfileResponse(p *user.Profile) dto.ProfileResponse {
	r := dto.ProfileResponse{
		ID:              p.ID.String(),
		EmergencyMonths: p.EmergencyMonths,
		OnboardingDone:  p.OnboardingDone,
	}
	if p.IncomeHint != nil {
		r.IncomeHint = *p.IncomeHint
	}
	if p.HousingType != nil {
		r.HousingType = *p.HousingType
	}
	if p.LifestyleStyle != nil {
		r.LifestyleStyle = *p.LifestyleStyle
	}
	if p.ActiveProgram != nil {
		r.ActiveProgram = *p.ActiveProgram
	}
	return r
}
```

- [ ] **Step 3: Update `server.Deps` + routes**

```go
// in internal/server/server.go
type Deps struct {
	Cfg            *config.Config
	ProfileHandler *handler.ProfileHandler
	// (more added in later phases)
}
```

```go
// in internal/server/routes.go
func registerRoutes(e *echo.Echo, d Deps) {
	e.GET("/health", healthHandler)

	v1 := e.Group("/v1")
	v1.Use(middleware.JWT(d.Cfg.SupabaseJWTSecret, d.Cfg.SupabaseJWTAudience))

	v1.GET("/profile", d.ProfileHandler.Get)
	v1.PATCH("/profile", d.ProfileHandler.Update)
	v1.PUT("/profile/income", d.ProfileHandler.UpdateIncome)
}
```

- [ ] **Step 4: Wire in `apps/api/main.go`**

```go
// after pool creation:
enc, err := encryption.New(cfg.IncomeEncryptionKey)
if err != nil { panic(err) }

userRepo := repository.NewUserRepo(pool)
userSvc  := user.NewService(userRepo, enc)

profileH := &handler.ProfileHandler{Svc: userSvc}

e := server.New(server.Deps{
	Cfg:            cfg,
	ProfileHandler: profileH,
})
```

- [ ] **Step 5: Build, run, smoke test**

```bash
go build ./...
go run ./apps/api
# in another shell, with a valid Supabase JWT:
curl -H "Authorization: Bearer $JWT" localhost:8080/v1/profile
```

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "feat(profile): GET/PATCH /profile + PUT /profile/income"
```

---

## Phase 6 — Categories Domain

### Task 24: Category entity + repo + queries

**Files:**
- Create: `internal/domain/category/{entity,repository}.go`
- Create: `database/sqlc/query/category.sql`
- Create: `internal/repository/category_repo.go`

- [ ] **Step 1: `database/sqlc/query/category.sql`**

```sql
-- name: ListCategoriesForUser :many
SELECT * FROM expense_categories
WHERE (user_id IS NULL OR user_id = $1) AND is_active = TRUE
ORDER BY sort_order, name;

-- name: CreateCategory :one
INSERT INTO expense_categories (user_id, name, icon, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM expense_categories WHERE id = $1 AND user_id = $2;

-- name: GetCategory :one
SELECT * FROM expense_categories WHERE id = $1;
```

- [ ] **Step 2: Generate**

Run: `./scripts/generate.sh`

- [ ] **Step 3: Domain + repo + service** (boilerplate, follow user/ shape — interface, sqlc-backed impl, service that wraps repo and forbids deleting `is_default = true`)

- [ ] **Step 4: Handler** — GET/POST/DELETE on `/v1/categories`. Validate type ∈ {fixed,variable,debt,want}. Return 403 on attempt to delete system default.

- [ ] **Step 5: Wire + smoke + commit**

```bash
git add . && git commit -m "feat(category): list system+custom, create custom, delete with default-protection"
```

---

## Phase 7 — Onboarding & Budget Engine

### Task 25: Budget engine (pure logic)

**Files:**
- Create: `internal/domain/budget/engine.go`
- Test: `internal/domain/budget/engine_test.go`

This is the heart of onboarding — heavy unit tests. The engine takes onboarding answers + expense items and produces a `BudgetPlan` with allocations summing to income, plus a `program` and a `warning` string.

- [ ] **Step 1: Failing tests covering each program path**

```go
package budget_test

import (
	"testing"

	"github.com/<owner>/fintrack/internal/domain/budget"
	"github.com/stretchr/testify/require"
)

func TestEngine_BebasUtang_AssignsDebtFocusFlag(t *testing.T) {
	in := budget.OnboardingInput{
		Income:          8_000_000,
		Goal:            "debt",
		HousingType:     "kpr",
		LifestyleStyle:  "balanced",
		EmergencyMonths: 1,
		DebtTypes:       []string{"cc"},
		ExpenseItems: []budget.OnboardingItem{
			{Name: "Sewa kosan", Type: "fixed", Amount: 1_200_000, CategoryID: idA()},
			{Name: "Cicilan KPR", Type: "fixed", Amount: 1_500_000, CategoryID: idB()},
			{Name: "Makan & minum", Type: "variable", Amount: 1_200_000, CategoryID: idC()},
			{Name: "Kartu kredit", Type: "debt", Amount: 400_000, CategoryID: idD()},
		},
	}
	out, err := budget.GenerateAllocation(in)
	require.NoError(t, err)
	require.Equal(t, "bebas_utang", out.Program)
	require.Equal(t, int64(8_000_000), out.Summary.Total)
	require.True(t, out.HasDebtFocus)
	require.NotEmpty(t, out.Warning, "kebutuhan over 50% should warn")
}

func TestEngine_Pondasi_NoDebt_NoEmergency(t *testing.T) {
	in := budget.OnboardingInput{
		Income: 6_000_000, Goal: "emergency", HousingType: "keluarga",
		LifestyleStyle: "balanced", EmergencyMonths: 0,
		ExpenseItems: []budget.OnboardingItem{
			{Name: "Makan", Type: "variable", Amount: 1_500_000, CategoryID: idA()},
		},
	}
	out, _ := budget.GenerateAllocation(in)
	require.Equal(t, "pondasi", out.Program)
}

func TestEngine_Tumbuh_EmergencyDone(t *testing.T) {
	in := budget.OnboardingInput{
		Income: 12_000_000, Goal: "invest", EmergencyMonths: 6, LifestyleStyle: "balanced",
		ExpenseItems: []budget.OnboardingItem{
			{Name: "Sewa", Type: "fixed", Amount: 2_500_000, CategoryID: idA()},
		},
	}
	out, _ := budget.GenerateAllocation(in)
	require.Equal(t, "tumbuh", out.Program)
}

func TestEngine_RejectsExpensesOverIncome(t *testing.T) {
	in := budget.OnboardingInput{
		Income: 5_000_000, Goal: "balance", LifestyleStyle: "balanced",
		ExpenseItems: []budget.OnboardingItem{
			{Name: "Sewa", Type: "fixed", Amount: 6_000_000, CategoryID: idA()},
		},
	}
	_, err := budget.GenerateAllocation(in)
	require.Error(t, err)
}
```

(`idA/B/C/D` are helpers that return `uuid.New()`.)

- [ ] **Step 2: Run tests, expect FAIL**

Run: `go test ./internal/domain/budget/...`

- [ ] **Step 3: Implement engine**

```go
package budget

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type OnboardingItem struct {
	Name       string
	Icon       string
	Type       string // fixed|variable|debt|want
	Amount     int64
	CategoryID uuid.UUID
}

type OnboardingInput struct {
	Income          int64
	Goal            string   // emergency|debt|goal|invest|balance
	HousingType     string
	LifestyleStyle  string
	EmergencyMonths int
	DebtTypes       []string
	ExpenseItems    []OnboardingItem
}

type AllocationItem struct {
	CategoryID      uuid.UUID
	CategoryName    string
	Icon            string
	Type            string
	AllocatedAmount int64
	Percentage      float64
	IsDebtFocus     bool
}

type SummaryGroup struct {
	Amount     int64
	Percentage float64
}

type AllocationSummary struct {
	Kebutuhan SummaryGroup
	Utang     SummaryGroup
	Keinginan SummaryGroup
	Tabungan  SummaryGroup
	Total     int64
}

type Allocation struct {
	Program      string
	Items        []AllocationItem
	Summary      AllocationSummary
	HasDebtFocus bool
	Warning      string
}

func GenerateAllocation(in OnboardingInput) (*Allocation, error) {
	if in.Income <= 0 {
		return nil, errors.New("income must be > 0")
	}
	var totalExp int64
	for _, it := range in.ExpenseItems {
		if it.Amount < 0 {
			return nil, errors.New("expense amount must be >= 0")
		}
		totalExp += it.Amount
	}
	if totalExp > in.Income {
		return nil, errors.New("expenses exceed income")
	}

	prog := classifyProgram(in)
	out := &Allocation{Program: prog}
	out.Items = make([]AllocationItem, 0, len(in.ExpenseItems))

	var fixed, variable, debt, want int64
	for _, it := range in.ExpenseItems {
		isFocus := prog == "bebas_utang" && it.Type == "debt"
		out.Items = append(out.Items, AllocationItem{
			CategoryID:      it.CategoryID,
			CategoryName:    it.Name,
			Icon:            it.Icon,
			Type:            it.Type,
			AllocatedAmount: it.Amount,
			Percentage:      pct(it.Amount, in.Income),
			IsDebtFocus:     isFocus,
		})
		if isFocus {
			out.HasDebtFocus = true
		}
		switch it.Type {
		case "fixed":
			fixed += it.Amount
		case "variable":
			variable += it.Amount
		case "debt":
			debt += it.Amount
		case "want":
			want += it.Amount
		}
	}

	kebutuhan := fixed + variable
	tabungan := in.Income - kebutuhan - debt - want
	if tabungan < 0 {
		tabungan = 0
	}

	out.Summary = AllocationSummary{
		Kebutuhan: SummaryGroup{kebutuhan, pct(kebutuhan, in.Income)},
		Utang:     SummaryGroup{debt, pct(debt, in.Income)},
		Keinginan: SummaryGroup{want, pct(want, in.Income)},
		Tabungan:  SummaryGroup{tabungan, pct(tabungan, in.Income)},
		Total:     in.Income,
	}

	if pct(kebutuhan, in.Income) > 50 {
		out.Warning = fmt.Sprintf(
			"Kebutuhan pokok %.0f%% — sedikit di atas ideal (50%%), wajar untuk kondisimu.",
			pct(kebutuhan, in.Income))
	}
	return out, nil
}

func classifyProgram(in OnboardingInput) string {
	hasDebt := false
	for _, it := range in.ExpenseItems {
		if it.Type == "debt" && it.Amount > 0 {
			hasDebt = true
			break
		}
	}
	switch in.Goal {
	case "debt":
		if hasDebt {
			return "bebas_utang"
		}
		return "seimbang"
	case "emergency":
		return "pondasi"
	case "goal":
		return "goal_chaser"
	case "invest":
		if in.EmergencyMonths >= 3 {
			return "tumbuh"
		}
		return "pondasi"
	default:
		return "seimbang"
	}
}

func pct(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
```

- [ ] **Step 4: Run tests, expect PASS, commit**

```bash
go test ./internal/domain/budget/...
git add internal/domain/budget
git commit -m "feat(budget): pure allocation engine — program classification + summary + warning"
```

### Task 26: Budget repo + service + onboarding handler

**Files:**
- Create: `database/sqlc/query/budget.sql`
- Create: `internal/domain/budget/{entity,repository,service}.go`
- Create: `internal/repository/budget_repo.go`
- Create: `internal/handler/dto/budget_dto.go`
- Create: `internal/handler/onboarding_handler.go`
- Create: `internal/handler/budget_handler.go`

- [ ] **Step 1: SQL queries** (CreateBudgetPlan, CreateBudgetItem, GetBudgetForPeriod with items via join, UpdateBudgetItem)

```sql
-- name: CreateBudgetPlan :one
INSERT INTO budget_plans (user_id, period_year, period_month, total_income, program)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (user_id, period_year, period_month)
DO UPDATE SET total_income = EXCLUDED.total_income, program = EXCLUDED.program, updated_at = NOW()
RETURNING *;

-- name: CreateBudgetItem :one
INSERT INTO budget_items (budget_plan_id, category_id, allocated_amount, percentage, is_debt_focus)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (budget_plan_id, category_id)
DO UPDATE SET allocated_amount = EXCLUDED.allocated_amount, percentage = EXCLUDED.percentage,
              is_debt_focus = EXCLUDED.is_debt_focus, updated_at = NOW()
RETURNING *;

-- name: GetCurrentBudgetPlan :one
SELECT * FROM budget_plans WHERE user_id = $1 AND period_year = $2 AND period_month = $3;

-- name: ListBudgetItemsWithCategory :many
SELECT bi.*, ec.name AS category_name, ec.icon AS category_icon, ec.type AS category_type
FROM budget_items bi JOIN expense_categories ec ON ec.id = bi.category_id
WHERE bi.budget_plan_id = $1
ORDER BY ec.sort_order;

-- name: UpdateBudgetItem :one
UPDATE budget_items SET allocated_amount = $2, updated_at = NOW()
WHERE id = $1 RETURNING *;
```

- [ ] **Step 2: Generate sqlc**

- [ ] **Step 3: Service: `GenerateFromOnboarding(ctx, userID, in)` →
  1. Call engine.GenerateAllocation
  2. Call userRepo.Create/Update profile (encrypt income, set program, mark onboarding_done=true)
  3. Insert budget_plan + items inside a pgx transaction
  4. Return Allocation with budget_plan_id**

- [ ] **Step 4: Service: `GetCurrent(ctx, userID, year, month)` →
  - Fetch plan + items
  - Sum spent per category from transactions table (delegate via TransactionRepo.SumByCategoryAndPeriod)
  - For each item: compute spent, percentage_used, status (fresh<60, warning<85, fatigued>=85), remaining
  - Return aggregated DTO**

- [ ] **Step 5: Onboarding handler** (`POST /v1/onboarding`, validates body matching PRD §08, returns 201 with allocation envelope).

- [ ] **Step 6: Budget handler** (`GET /v1/budget/current`, `GET /v1/budget/:year/:month`, `PATCH /v1/budget/items/:id`).

- [ ] **Step 7: Wire in main.go + register routes + smoke + commit**

```bash
git add .
git commit -m "feat(budget): onboarding generates plan + GET current with fatigue status"
```

---

## Phase 8 — Transactions Domain

### Task 27: Transaction repo + service + handler (CRUD)

**Files:**
- Create: `database/sqlc/query/transaction.sql`
- Create: `internal/domain/transaction/{entity,repository,service}.go`
- Create: `internal/repository/transaction_repo.go`
- Create: `internal/handler/dto/transaction_dto.go`
- Create: `internal/handler/transaction_handler.go`

Endpoints: `GET/POST /v1/transactions`, `PATCH/DEL /v1/transactions/:id`. POST returns the new transaction PLUS a `fatigue_alert` (computed by injecting fatigue.Service after Phase 9).

- [ ] **Step 1: SQL queries**

```sql
-- name: ListTransactions :many
SELECT t.*, ec.name AS category_name, ec.icon AS category_icon, ec.type AS category_type
FROM transactions t JOIN expense_categories ec ON ec.id = t.category_id
WHERE t.user_id = $1
  AND ($2::uuid IS NULL OR t.category_id = $2)
  AND (sqlc.narg('start')::timestamptz IS NULL OR t.transacted_at >= sqlc.narg('start'))
  AND (sqlc.narg('end')::timestamptz   IS NULL OR t.transacted_at <  sqlc.narg('end'))
ORDER BY t.transacted_at DESC
LIMIT $3 OFFSET $4;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1
  AND ($2::uuid IS NULL OR category_id = $2);

-- name: CreateTransaction :one
INSERT INTO transactions (user_id, budget_plan_id, category_id, amount, note, receipt_url,
                          ai_categorized, ai_confidence, transacted_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING *;

-- name: UpdateTransaction :one
UPDATE transactions SET
    amount        = COALESCE(sqlc.narg('amount'), amount),
    note          = COALESCE(sqlc.narg('note'), note),
    category_id   = COALESCE(sqlc.narg('category_id'), category_id),
    transacted_at = COALESCE(sqlc.narg('transacted_at'), transacted_at),
    updated_at    = NOW()
WHERE id = $1 AND user_id = $2 RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1 AND user_id = $2;

-- name: SumSpentByCategoryForPlan :many
SELECT category_id, SUM(amount)::bigint AS total
FROM transactions
WHERE user_id = $1 AND budget_plan_id = $2
GROUP BY category_id;
```

- [ ] **Step 2: Domain + service tests** (creating tx, list with pagination, sum-by-category)

- [ ] **Step 3: Implement repo + service + handler**

- [ ] **Step 4: Smoke + commit**

```bash
git add .
git commit -m "feat(transaction): CRUD + filtering by month/category + pagination"
```

---

## Phase 9 — Fatigue Domain

### Task 28: Fatigue calculator + handler

**Files:**
- Create: `internal/domain/fatigue/{entity,service}.go`
- Test: `internal/domain/fatigue/service_test.go`
- Create: `internal/handler/fatigue_handler.go`

No repo — fatigue is derived state.

Status thresholds:
- `fresh` if `percentage < 60`
- `warning` if `60 <= percentage < 85`
- `fatigued` if `percentage >= 85`

**Daily budget remaining:** `(allocated - spent) / max(1, days_remaining_in_month)`.

- [ ] **Step 1: Failing test for status thresholds + tip generation**

```go
func TestStatus_Thresholds(t *testing.T) {
	require.Equal(t, "fresh",    fatigue.ComputeStatus(0, 100))
	require.Equal(t, "fresh",    fatigue.ComputeStatus(59, 100))
	require.Equal(t, "warning",  fatigue.ComputeStatus(60, 100))
	require.Equal(t, "warning",  fatigue.ComputeStatus(84, 100))
	require.Equal(t, "fatigued", fatigue.ComputeStatus(85, 100))
}
```

- [ ] **Step 2: Implement service**

```go
package fatigue

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/<owner>/fintrack/internal/domain/budget"
)

type Service interface {
	Snapshot(ctx context.Context, userID uuid.UUID, now time.Time) (*Snapshot, error)
	AlertForCategory(ctx context.Context, userID, categoryID uuid.UUID, now time.Time) (*Alert, error)
}

type CategorySnapshot struct {
	CategoryID           uuid.UUID
	CategoryName         string
	CategoryIcon         string
	Type                 string
	Allocated            int64
	Spent                int64
	Remaining            int64
	Percentage           float64
	Status               string
	DailyBudgetRemaining int64
	Tip                  *string
}

type Snapshot struct {
	Period       string
	DayOfMonth   int
	DaysRemaining int
	Categories   []CategorySnapshot
	Overall      Overall
}

type Overall struct {
	TotalAllocated     int64
	TotalSpent         int64
	Percentage         float64
	ProjectedEOM       int64
	OnTrack            bool
}

type Alert struct {
	Status         string
	CategoryName   string
	PercentageUsed float64
	Message        string
}

func ComputeStatus(spent, allocated int64) string {
	if allocated == 0 {
		return "fresh"
	}
	p := float64(spent) / float64(allocated) * 100
	switch {
	case p >= 85:
		return "fatigued"
	case p >= 60:
		return "warning"
	default:
		return "fresh"
	}
}

// service that depends on budget.Service is wired in main.go.
```

- [ ] **Step 3: Handler `GET /v1/fatigue`** (combines budget.Service.GetCurrent with category tips. For each `fatigued`/`warning` category, generate a contextual tip — for MVP a static template like "Budget {name} {status}. Sisa Rp {remaining} untuk {days} hari ke depan." The AI-generated tip can replace this in v2.)

- [ ] **Step 4: Hook into transaction POST** — after creating a tx, call `fatigueSvc.AlertForCategory(ctx, userID, tx.CategoryID, now)` and embed in response if status != "fresh".

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat(fatigue): status calc + dashboard endpoint + transaction inline alert"
```

---

## Phase 10 — AI Integration

### Task 29: Anthropic HTTP client

**Files:**
- Create: `internal/ai/client.go`
- Test: `internal/ai/client_test.go`

- [ ] **Step 1: Test using httptest stub server** (assert headers `x-api-key`, `anthropic-version`; assert request body shape; return canned response)

- [ ] **Step 2: Implement** (POST to `https://api.anthropic.com/v1/messages`, 30s timeout, 1 retry on 5xx)

```go
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const apiURL = "https://api.anthropic.com/v1/messages"

type Client struct {
	apiKey string
	model  string
	hc     *http.Client
	url    string
}

func New(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey, model: model,
		hc:     &http.Client{Timeout: 30 * time.Second},
		url:    apiURL,
	}
}

type Message struct {
	Role    string  `json:"role"`
	Content []Block `json:"content"`
}

type Block struct {
	Type   string  `json:"type"`
	Text   string  `json:"text,omitempty"`
	Source *Source `json:"source,omitempty"`
}

type Source struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // "image/jpeg"
	Data      string `json:"data"`
}

type request struct {
	Model     string    `json:"model"`
	System    string    `json:"system,omitempty"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type response struct {
	Content []Block `json:"content"`
}

func (c *Client) Complete(ctx context.Context, system string, msgs []Message, maxTokens int) (string, error) {
	body, _ := json.Marshal(request{
		Model: c.model, System: system, MaxTokens: maxTokens, Messages: msgs,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	res, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 500 {
		return "", errors.New("anthropic 5xx")
	}
	if res.StatusCode >= 400 {
		return "", errors.New("anthropic 4xx")
	}
	var out response
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	for _, b := range out.Content {
		if b.Type == "text" {
			return b.Text, nil
		}
	}
	return "", errors.New("no text in response")
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/ai
git commit -m "feat(ai): minimal Anthropic Messages API client"
```

### Task 30: Receipt categorizer

**Files:**
- Create: `internal/ai/categorizer.go`
- Test: `internal/ai/categorizer_test.go`
- Create: `internal/handler/transaction_handler.go` (extend with `Scan` method)

Returns `{amount int64, suggested_category_id uuid, suggested_category_name, note, confidence float64, alternatives []}`.

- [ ] **Step 1: Define interface and prompt template**

```go
package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type ReceiptScan struct {
	Amount         int64                  `json:"amount"`
	CategoryName   string                 `json:"category_name"`
	Note           string                 `json:"note"`
	Confidence     float64                `json:"confidence"`
	Alternatives   []ReceiptAlternative   `json:"alternatives"`
}

type ReceiptAlternative struct {
	CategoryName string  `json:"category_name"`
	Confidence   float64 `json:"confidence"`
}

type Category struct {
	ID   uuid.UUID
	Name string
	Type string
}

type Categorizer struct{ c *Client }

func NewCategorizer(c *Client) *Categorizer { return &Categorizer{c: c} }

func (cz *Categorizer) Scan(ctx context.Context, image []byte, mimeType string, available []Category) (*ReceiptScan, uuid.UUID, error) {
	cats, _ := json.Marshal(catNames(available))
	system := "Kamu asisten pencatat keuangan. Jawab HANYA JSON valid."
	prompt := fmt.Sprintf(`Baca struk pada gambar. Kembalikan JSON:
{"amount": int_rupiah, "category_name": "...", "note": "merchant atau item utama", "confidence": 0.0-1.0, "alternatives":[{"category_name":"...","confidence":...}]}
Pilih category_name dari daftar ini saja: %s. Tidak ada teks tambahan.`, string(cats))

	msg := Message{Role: "user", Content: []Block{
		{Type: "image", Source: &Source{Type: "base64", MediaType: mimeType, Data: base64.StdEncoding.EncodeToString(image)}},
		{Type: "text", Text: prompt},
	}}
	raw, err := cz.c.Complete(ctx, system, []Message{msg}, 600)
	if err != nil {
		return nil, uuid.Nil, err
	}
	var out ReceiptScan
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, uuid.Nil, fmt.Errorf("parse ai response: %w", err)
	}
	id := matchCategory(out.CategoryName, available)
	return &out, id, nil
}

func catNames(cats []Category) []string {
	out := make([]string, 0, len(cats))
	for _, c := range cats {
		out = append(out, c.Name)
	}
	return out
}

func matchCategory(name string, cats []Category) uuid.UUID {
	for _, c := range cats {
		if c.Name == name {
			return c.ID
		}
	}
	return uuid.Nil
}
```

- [ ] **Step 2: `POST /v1/transactions/scan` handler**
  - `multipart/form-data`, field `image`
  - upload to Supabase Storage (or skip and pass URL through if frontend handles upload — for MVP we skip storage and just pass back a placeholder `receipt_url` from the request multipart, or write to local tmp; a follow-up plan covers Supabase Storage integration)
  - call categorizer with image bytes + categories list
  - return `{amount, suggested_category_id, suggested_category_name, note, confidence, receipt_url, alternatives}`
  - DOES NOT persist — user POSTs `/v1/transactions` after confirmation.

- [ ] **Step 3: Commit**

```bash
git add .
git commit -m "feat(ai): receipt scan endpoint — image -> categorized suggestion"
```

### Task 31: Narrative summarizer (used by reports + worker)

**Files:**
- Create: `internal/ai/summarizer.go`

- [ ] **Step 1: Implement** — input a struct with weekly stats (totals, top categories, vs last week), output a 2-3 sentence Indonesian narrative. Prompt instructs short, friendly, non-judgmental tone matching PRD §02.

- [ ] **Step 2: Commit**

---

## Phase 11 — Debts Domain

### Task 32: Debt entity + payoff calculator + endpoints

**Files:**
- Create: `database/sqlc/query/debt.sql`
- Create: `internal/domain/debt/{entity,repository,service}.go`
- Create: `internal/repository/debt_repo.go`
- Create: `internal/handler/debt_handler.go`
- Test: `internal/domain/debt/service_test.go`

Endpoints: `GET/POST /v1/debts`, `PATCH/DEL /v1/debts/:id`, `POST /v1/debts/:id/payment`.

Snowball: order by smallest `current_balance` first. Avalanche: order by highest `interest_rate` first. `monthly_payment = min_payment + extra_from_focus`. `estimated_payoff_months` from amortization with monthly compounding `r = interest_rate / 12 / 100`:

```
n = ceil( -log(1 - balance*r/payment) / log(1+r) )
```

If `r == 0`, just `ceil(balance/payment)`. If `payment <= balance*r`, return -1 (never pays off — service should warn).

- [ ] **Step 1: Tests for both methods + edge cases (zero interest, over-min payment)**

- [ ] **Step 2: Implement service.PayoffMonths + ranking**

- [ ] **Step 3: Repo + handler**

- [ ] **Step 4: `POST /debts/:id/payment` flow**:
  - decrement `current_balance` by `amount`, set `paid_at` if hits 0, recalc `target_paid_at`
  - return new balance + new estimated months + `is_paid_off`

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat(debt): snowball/avalanche ordering + payoff projection + payment endpoint"
```

---

## Phase 12 — Goals Domain

### Task 33: Goals CRUD + deposit + estimate

**Files:**
- Create: `database/sqlc/query/goal.sql`
- Create: `internal/domain/goal/{entity,repository,service}.go`
- Create: `internal/repository/goal_repo.go`
- Create: `internal/handler/goal_handler.go`

Endpoints: `GET/POST /v1/goals`, `PATCH/DEL /v1/goals/:id`, `POST /v1/goals/:id/deposit`.

`estimated_months = ceil((target - current) / monthly_savings_capacity)` where capacity comes from latest budget plan's `tabungan` summary. `monthly_needed = (target - current) / max(1, months_to_target_date)`.

Milestone messages at 25/50/75/100%.

- [ ] **Step 1: Tests for milestone messages**

- [ ] **Step 2: Implement + commit**

---

## Phase 13 — Reports Domain

### Task 34: Weekly + monthly report generation

**Files:**
- Create: `database/sqlc/query/report.sql`
- Create: `internal/domain/report/{entity,repository,service}.go`
- Create: `internal/repository/report_repo.go`
- Create: `internal/handler/report_handler.go`

Endpoints: `GET /v1/reports/weekly`, `GET /v1/reports/monthly/:year/:month`.

Weekly: aggregate transactions for current ISO week vs prior week, top 3 categories, narrative via `ai.Summarizer`. Cache in `weekly_reports` table; only regenerate if not cached or > 1 hour old.

Monthly: aggregate by category type (`fixed/variable/debt/want`), saving rate, personal record check (compare against best previous month).

- [ ] **Step 1: Service tests for aggregation + saving rate calc**

- [ ] **Step 2: Implement + commit**

---

## Phase 14 — API Tokens (BYOA)

### Task 35: Token generate, list, revoke + middleware

**Files:**
- Create: `database/sqlc/query/token.sql`
- Create: `internal/domain/token/{entity,repository,service}.go`
- Create: `internal/repository/token_repo.go`
- Create: `internal/handler/token_handler.go`
- Create: `internal/middleware/apitoken.go`

Token format: `fnt_live_<base64url 32 random bytes>`. Store bcrypt hash. Display the plaintext **once only** in the create response.

- [ ] **Step 1: Add bcrypt dep**

```bash
go get golang.org/x/crypto/bcrypt
```

- [ ] **Step 2: Service.Generate**: random 32 bytes → format → bcrypt hash → save with `token_hint = "..." + last4`

- [ ] **Step 3: Middleware** that accepts `Authorization: Bearer fnt_live_…`. Look up active tokens for-user (efficient: store a non-secret prefix index). For MVP: scan all active tokens with `is_active = true AND expires_at IS NULL OR > NOW()` and bcrypt-compare each. Acceptable up to ~thousands of tokens; revisit later.
  - Actually better: store token hash with deterministic SHA-256 in addition to bcrypt — query by SHA-256 prefix, then bcrypt-verify. Add column `token_lookup_hash` (SHA256 hex, 64 chars). Add a follow-up migration if not in current schema.

- [ ] **Step 4: Routes**:
  - `/v1/tokens` group uses JWT middleware (BYOA tokens cannot manage themselves — only Supabase JWT)
  - All other v1 routes accept either JWT or `apitoken` middleware (chained — JWT first; if no Bearer JWT, fall through to apitoken). For write endpoints, apitoken middleware checks `can_write`.

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat(tokens): BYOA token generate/list/revoke + bearer middleware with scope check"
```

---

## Phase 15 — Worker

### Task 36: gocron scheduler + weekly email job

**Files:**
- Create: `internal/worker/scheduler.go`
- Create: `internal/worker/weekly_report.go`
- Modify: `apps/worker/main.go`

Run as a separate Railway service so HTTP service stays single-purpose.

Job: every Monday 7am Asia/Jakarta, for each user with `onboarding_done = true`, generate a weekly report via `report.Service.GenerateWeekly`, and (if email is configured later) send via SMTP. For MVP we only cache the report; email integration is a follow-up.

- [ ] **Step 1: Add gocron**

```bash
go get github.com/go-co-op/gocron/v2
```

- [ ] **Step 2: Scheduler**

```go
package worker

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
)

type Job func(ctx context.Context) error

func Run(ctx context.Context, weekly Job) error {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	s, err := gocron.NewScheduler(gocron.WithLocation(loc))
	if err != nil {
		return err
	}
	_, err = s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Monday), gocron.NewAtTimes(gocron.NewAtTime(7, 0, 0))),
		gocron.NewTask(func() {
			if err := weekly(ctx); err != nil {
				log.Error().Err(err).Msg("weekly job failed")
			}
		}),
	)
	if err != nil {
		return err
	}
	s.Start()
	<-ctx.Done()
	return s.Shutdown()
}
```

- [ ] **Step 3: Job: iterate users via repo and call report svc**

- [ ] **Step 4: Wire `apps/worker/main.go`**

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat(worker): gocron scheduler + weekly report job"
```

---

## Phase 16 — Polish & Deploy

### Task 37: Health endpoint with DB check + version

**Files:**
- Create: `internal/handler/health_handler.go`

- [ ] **Step 1: Implement**

```go
package handler

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/<owner>/fintrack/pkg/response"
)

type HealthHandler struct {
	Pool    *pgxpool.Pool
	Version string
}

func (h *HealthHandler) Get(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Second)
	defer cancel()
	dbStatus := "ok"
	if err := h.Pool.Ping(ctx); err != nil {
		dbStatus = "down"
	}
	return response.OK(c, map[string]string{
		"status":  "ok",
		"version": h.Version,
		"db":      dbStatus,
	})
}
```

- [ ] **Step 2: Wire as `GET /health`** (no JWT)

- [ ] **Step 3: Commit**

### Task 38: Rate limiting middleware

**Files:**
- Create: `internal/middleware/ratelimit.go`

In-memory token bucket per `user_id`, 60 req/min. (Acceptable for single Railway instance MVP; promote to Redis later if scaling out.)

- [ ] **Step 1: Implement using `golang.org/x/time/rate`**

```bash
go get golang.org/x/time/rate
```

```go
package middleware

import (
	"sync"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"github.com/<owner>/fintrack/pkg/apperror"
)

type RateLimit struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func NewRateLimit(rps float64, burst int) *RateLimit {
	return &RateLimit{limiters: map[string]*rate.Limiter{}, r: rate.Limit(rps), b: burst}
}

func (rl *RateLimit) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key, _ := c.Get("user_id").(string)
			if key == "" {
				key = c.RealIP()
			}
			rl.mu.Lock()
			lim, ok := rl.limiters[key]
			if !ok {
				lim = rate.NewLimiter(rl.r, rl.b)
				rl.limiters[key] = lim
			}
			rl.mu.Unlock()
			if !lim.Allow() {
				return &apperror.Error{Code: "RATE_LIMITED", Message: "too many requests", HTTP: 429}
			}
			return next(c)
		}
	}
}
```

- [ ] **Step 2: Wire after JWT middleware in routes.go**

- [ ] **Step 3: Commit**

### Task 39: Railway deploy config

**Files:**
- Create: `railway.toml`
- Create: `apps/api/Dockerfile`
- Create: `apps/worker/Dockerfile`

- [ ] **Step 1: API Dockerfile**

```dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./apps/api

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
```

- [ ] **Step 2: Worker Dockerfile** — analogous, builds `./apps/worker`

- [ ] **Step 3: `railway.toml`** wiring two services from same repo

```toml
[build]
builder = "DOCKERFILE"

[[services]]
name = "api"
dockerfile = "apps/api/Dockerfile"

[[services]]
name = "worker"
dockerfile = "apps/worker/Dockerfile"
```

- [ ] **Step 4: README** — document setup (`make` targets, env vars, migrate command, run sqlc)

- [ ] **Step 5: Commit**

```bash
git add railway.toml apps/api/Dockerfile apps/worker/Dockerfile README.md
git commit -m "chore(deploy): Dockerfiles + Railway config"
```

### Task 40: End-to-end smoke test script

**Files:**
- Create: `scripts/smoke.sh`

A bash script that:
1. Mints a local HS256 JWT for a test user (using the configured secret)
2. POST /onboarding with sample body from PRD §08
3. GET /budget/current — assert status 200 and items present
4. POST /transactions — assert fatigue_alert returned
5. GET /fatigue — assert categories array non-empty
6. POST /debts + POST /debts/:id/payment
7. POST /goals + POST /goals/:id/deposit
8. GET /reports/weekly
9. POST /tokens — capture token, then GET /v1/profile with that token instead of JWT
10. DEL /tokens/:id — assert subsequent request 401

- [ ] **Step 1: Write script using `curl` + `jq`** (skip if too time-consuming — leave as TODO marker for full integration test phase)

- [ ] **Step 2: Run end to end against local stack**

- [ ] **Step 3: Final commit**

```bash
git add scripts/smoke.sh
git commit -m "test(smoke): end-to-end happy path against local server"
```

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

## Verification

After Phase 16 is complete, the system should satisfy:

1. **Unit tests pass**: `make test-race` — every domain service, the budget engine, fatigue thresholds, encryption round-trip, JWT parsing.
2. **Local server boots**: `make run` brings up `:8080`; `curl localhost:8080/health` returns `{"data":{"status":"ok","version":"…","db":"ok"}}`.
3. **Migration round-trip**: `make migrate-down` then `make migrate-up` succeeds without errors against a clean Postgres.
4. **Smoke script**: `./scripts/smoke.sh` exits 0 — exercises onboarding → budget → tx → fatigue → debt → goal → report → BYOA token paths.
5. **Worker**: `make run-worker` starts; logs show next-Monday schedule registered.
6. **All 25 endpoints from PRD §08 are reachable** with valid JWT (or token where applicable). Response envelopes match `{"data":...,"meta":...}` shape; errors match `{"error":{"code","message","fields"}}`.

After deploy:
- Railway app responds at health endpoint with `db: "ok"`.
- Logs show structured JSON with `request_id`.
- Calling any non-`/health` endpoint without `Authorization` returns `401 UNAUTHORIZED`.

---

## Notes for the executing agent

- **Replace `<owner>`** in all `import` paths with the actual GitHub owner before/at Task 1.
- **`go get` the module additions when you first need them**; do not pre-fetch everything in Task 1.
- **sqlc generated code is gitignored** (per `.gitignore`). Whoever clones runs `./scripts/generate.sh` before `go build`.
- **Frequent commits** — each Task ends with a commit. Don't batch.
- **Don't fight RLS** — the app connects with the Supabase service role for system jobs (`auth.uid()` is null in worker, so RLS won't allow access). For the worker we either bypass RLS via the service role, or pass `SET LOCAL "request.jwt.claim.sub" = '<user_id>'` per query. Decide once when you start Phase 15; document in the worker's README section.
- **Receipt storage**: `POST /transactions/scan` currently takes the image inline and returns a `receipt_url` placeholder. Real Supabase Storage upload is intentionally deferred to a follow-up plan — call it out in the README as a known MVP gap.
- **Email delivery for weekly report**: also deferred — worker only generates and caches. Document as known MVP gap.
