---
title: "Fintrack — Architecture"
type: project-doc
created: 2026-06-03
last_updated: 2026-06-03 (storage layer added — see ADR-008/009/010)
tags: [project, architecture, stack, fintrack]
related:
  - "[[(C) PROJECT.md]]"
  - "[[(C) DECISIONS.md]]"
---

## Stack Summary

| Layer | Choice | Why |
|-------|--------|-----|
| **Backend language** | Go (Golang) | Learning goal — Hafis's stretch language |
| **HTTP framework** | Echo v4 | Clean middleware, explicit errors |
| **DB** | PostgreSQL via Supabase | Free tier, RLS, hosted auth |
| **DB access** | sqlc | Type-safe Go from explicit SQL — financial accuracy |
| **DB driver** | pgx/v5 | Modern Postgres, fastest |
| **Migrations** | golang-migrate | Versioned `.sql` files |
| **Auth** | Supabase Auth → JWT | Validated by `golang-jwt/jwt` in Go |
| **Encryption** | `crypto/aes` (stdlib) | AES-256-GCM for income field |
| **AI** | Claude Vision + Haiku via `net/http` | No SDK — lightweight |
| **Object storage** | S3-compatible via `minio-go` SDK | MinIO local (Docker), Supabase Storage prod — same client, endpoint swap |
| **Logging** | rs/zerolog | Structured JSON, Railway-compatible |
| **Validation** | go-playground/validator | Struct-tag based |
| **IDs** | google/uuid (v4) | Primary keys |
| **Frontend** | Vue 3 + Vite | Lean SPA, fast iteration |
| **PWA** | vite-plugin-pwa | Install, offline shell, service worker |
| **Routing** | Vue Router 4 | Standard |
| **State** | Pinia | Official, lean |
| **Forms** | VeeValidate + Zod | Type-safe validation matching Go backend |
| **HTTP client** | Axios | Or native fetch wrapper |
| **Styling** | Tailwind CSS | Utility-first, fast iteration, design tokens in config |
| **Component library** | shadcn-vue | Copy-paste components into repo (owned, not imported) — community Vue port of shadcn/ui |
| **Backend host** | Railway | TBD — confirm |
| **Frontend host** | Vercel / Cloudflare Pages | TBD — pick one |

## Deliberately Excluded (with reasoning)

| Excluded | Reason |
|----------|--------|
| GORM / any ORM | Financial calc requires explicit SQL — sqlc fits better |
| GraphQL | REST covers all 12 MVP endpoints cleanly |
| Redis | PostgreSQL handles current scale (1 user) |
| gRPC | Single-service monorepo, no inter-service comms |
| Message queues | Goroutines handle async needs |
| Docker (for app deploy) | Railway deploys from source — no need (Docker still used locally for MinIO via Compose) |
| Nuxt 3 | SSR wasted on authenticated dashboard, adds framework magic |
| Vuex | Pinia is the official successor |

## Monorepo Structure

```
fintrack/
├── apps/
│   ├── api/main.go              ← HTTP server entry
│   └── worker/main.go           ← Background jobs (deferred for v2)
├── internal/
│   ├── config/                  ← Viper env loading
│   ├── server/                  ← Echo init, middleware wiring
│   ├── middleware/              ← Auth, logging, rate limit
│   ├── domain/                  ← Business logic + interfaces
│   │   ├── user/
│   │   ├── budget/
│   │   ├── transaction/
│   │   └── fatigue/             ← Category fatigue calculations
│   ├── handler/                 ← HTTP handlers
│   ├── repository/              ← sqlc-generated DB layer
│   ├── ai/                      ← Claude Vision client
│   ├── storage/                 ← S3-compatible object storage (MinIO/Supabase)
│   └── encryption/              ← AES income encryption
├── database/
│   ├── migrations/              ← Numbered .sql files
│   └── sqlc/
│       ├── sqlc.yaml
│       ├── query/               ← Hand-written SQL
│       └── generated/           ← sqlc output — never edit manually
├── pkg/                         ← Public utils (errors, responses, logger)
└── web/                         ← Vue 3 + Vite frontend
    ├── src/
    │   ├── components/
    │   ├── views/
    │   ├── stores/              ← Pinia
    │   ├── router/
    │   └── api/                 ← HTTP client wrappers
    ├── public/
    └── vite.config.ts
```

## Dependency Injection

- Manual wiring from `main.go` downward
- Domain layer exposes only interfaces
- Test environments inject mocks; production injects real impls
- **No Wire / Fx** — explicit > magic for solo learning project

## Security Model

- **JWT auth** — Supabase issues, Go validates
- **RLS (Row Level Security)** — Postgres enforces user-scoped data access
- **Income encryption** — AES-256-GCM server-side before DB insert; plaintext never returned to client
- **API token hashing** — bcrypt for BYOA tokens (v2 feature)
- **Rate limiting** — Echo middleware, per-user buckets

## Data Flow: Receipt Scan (Hero Feature)

```
Frontend: User snaps photo, compresses to JPEG <1MB client-side
   ↓
POST /transactions/scan (multipart/form-data, 2MB body cap)
   ↓
Go handler validates content-type + size
   ↓
Storage.Upload(ctx, "receipts/{user_id}/{txn_id}.jpg", bytes, "image/jpeg")
   → MinIO local / Supabase Storage prod (S3-compatible)
   ↓
Go calls Claude Vision API with image bytes
   ↓
Claude returns: { amount, merchant, suggested_category, confidence }
   ↓
Go writes pending transaction (state: needs_confirmation), receipt_url stored
   ↓
Frontend shows confirmation UI (signed URL if reviewing image)
   ↓
User confirms → POST /transactions/{id}/confirm
   ↓
Transaction marked confirmed, fatigue dashboard updated
```

## Storage Layer

**Interface** (`internal/storage/storage.go`):

```go
type Storage interface {
    Upload(ctx, key string, data io.Reader, contentType string) (string, error)
    Delete(ctx, key string) error
    SignedURL(ctx, key string, ttl time.Duration) (string, error)
}
```

**Conventions:**
- **Bucket structure:** `receipts/{user_id}/{transaction_id}.jpg`
- **Signed URLs:** 15-min TTL when serving images back to user
- **Size cap:** 2MB request body limit (Echo middleware)
- **Format:** Client compresses HEIC/PNG → JPEG before upload via `browser-image-compression`
- **Local:** MinIO container via `docker-compose.yml`, endpoint `http://localhost:9000`
- **Prod:** Supabase Storage S3-compatible endpoint
- **Switching:** Env var `STORAGE_ENDPOINT` only — same `S3Storage` impl in both environments

**Deferred to v2:**
- Cleanup worker for orphaned uploads (snapped but never confirmed)
- Retry-on-Claude-API-failure using stored image
- User-facing "view receipt" history UI

## What's NOT Architected Yet

- Email service for v2 weekly reports (Resend? Postmark?)
- Error tracking (Sentry?)
- Analytics (PostHog?)
- CI/CD pipeline (GitHub Actions → Railway?)

## Connections

- [[(C) PROJECT.md]] — Project overview
- [[(C) DECISIONS.md]] — Why each stack choice was made
- [[(C) ROADMAP.md]] — When each layer gets built
