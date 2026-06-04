package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"

	"github.com/hafis915/fintrack/internal/ai"
	"github.com/hafis915/fintrack/internal/config"
	"github.com/hafis915/fintrack/internal/encryption"
	"github.com/hafis915/fintrack/internal/handler"
	"github.com/hafis915/fintrack/internal/llm"
	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/internal/storage"
	"github.com/hafis915/fintrack/pkg/responses"
)

// Deps groups everything a server needs. Wired in main.go and passed here —
// no global state, no init().
type Deps struct {
	Config *config.Config
	Logger zerolog.Logger
	DB     *pgxpool.Pool

	// ReceiptAnalyzer and Storage are optional overrides. When nil, New builds
	// real implementations from config (or stubs when no API key is set).
	// Integration tests inject stubs so they never hit Claude or MinIO.
	ReceiptAnalyzer ai.ReceiptAnalyzer
	Storage         storage.Storage

	// LLM is an optional override for the financial-planner chat language layer.
	// When nil, New builds the OpenRouter client when OPEN_ROUTER_API_KEY is set,
	// or the deterministic stub otherwise. Integration/e2e tests inject the stub
	// so they never call OpenRouter.
	LLM llm.Client
}

// New returns a configured Echo instance with global middleware mounted
// and the public + protected route groups registered. Returns an error if
// any upstream dependency (today: only the income cipher) fails to construct.
func New(d Deps) (*echo.Echo, error) {
	cipher, err := encryption.NewCipherFromHex(d.Config.IncomeEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("building income cipher: %w", err)
	}

	// Receipt analyzer: caller override → real Claude client when an API key is
	// configured → deterministic stub otherwise (local dev without a key).
	analyzer := d.ReceiptAnalyzer
	if analyzer == nil {
		if d.Config.AnthropicAPIKey == "" {
			analyzer = ai.NewStubAnalyzer()
		} else {
			analyzer = ai.NewClaudeAnalyzer(d.Config.AnthropicAPIKey, d.Config.AnthropicModel, nil)
		}
	}

	// Storage: caller override → real MinIO/S3 from config. A misconfigured
	// store is a hard boot failure rather than a silent degrade — the receipt
	// flow can't persist images without it.
	store := d.Storage
	if store == nil {
		store, err = storage.NewMinIOStorage(storage.MinIOConfig{
			Endpoint:  d.Config.StorageEndpoint,
			AccessKey: d.Config.StorageAccessKey,
			SecretKey: d.Config.StorageSecretKey,
			Bucket:    d.Config.StorageBucket,
			UseSSL:    d.Config.StorageUseSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("building storage: %w", err)
		}
	}

	// Planner language layer: caller override → real OpenRouter client when an
	// API key is configured → deterministic stub otherwise (local dev / tests).
	// The stub does network-free NLU; the budget math is always deterministic.
	plannerLLM := d.LLM
	if plannerLLM == nil {
		// In the test environment we ALWAYS use the deterministic stub so e2e and
		// integration runs never hit the network (and never flake on upstream
		// rate limits). This holds even if a real OPEN_ROUTER_API_KEY leaks in via
		// a developer's .env file, which viper merges ahead of env vars.
		if d.Config.Env == "test" || d.Config.OpenRouterAPIKey == "" {
			plannerLLM = llm.NewStubClient()
		} else {
			plannerLLM = llm.NewOpenRouterClient(d.Config.OpenRouterAPIKey, d.Config.OpenRouterModel, nil)
		}
	}

	users := repository.NewUsersRepo(d.DB)
	userProfiles := repository.NewUserProfilesRepo(d.DB)
	categories := repository.NewCategoriesRepo(d.DB)
	budgetPlans := repository.NewBudgetPlansRepo(d.DB)
	transactionsRepo := repository.NewTransactionsRepo(d.DB)
	fatigueRepo := repository.NewFatigueRepo(d.DB)

	onboarding := handler.NewOnboarding(handler.OnboardingDeps{
		Users:        users,
		UserProfiles: userProfiles,
		Categories:   categories,
		BudgetPlans:  budgetPlans,
		Cipher:       cipher,
		Logger:       d.Logger,
	})
	authHandler := handler.NewAuth(handler.AuthDeps{
		Users:     users,
		JWTSecret: d.Config.JWTSecret,
		JWTIssuer: d.Config.JWTIssuer,
	})
	categoriesHandler := handler.NewCategories(categories)
	transactionsHandler := handler.NewTransactions(transactionsRepo, categories)
	receiptsHandler := handler.NewReceipts(transactionsRepo, categories, analyzer, store)
	budgetHandler := handler.NewBudget(fatigueRepo)
	plannerHandler := handler.NewPlanner(categories, plannerLLM)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Body size cap protects against accidental large uploads on every route.
	// The receipt-scan endpoint will tighten this further when it lands.
	e.Use(echomw.BodyLimit("2M"))
	e.Use(middleware.RequestID())
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: strings.Split(d.Config.CORSAllowedOrigins, ","),
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, middleware.HeaderRequestID},
	}))

	// Public routes (no auth).
	e.GET("/health", healthHandler(d))

	// Phase 0 local-first auth: PUBLIC register/login that mint the same HS256
	// JWT the /v1 group validates. Gated by config so prod (Supabase Auth) can
	// disable them. NOT under the JWTAuth/EnsureUser group — they create the JWT.
	if d.Config.AuthLocalEnabled {
		e.POST("/v1/auth/register", authHandler.Register)
		e.POST("/v1/auth/login", authHandler.Login)
	}

	// Protected routes live under /v1 and require a valid JWT + a bootstrapped users row.
	v1 := e.Group("/v1",
		middleware.JWTAuth(d.Config.JWTSecret),
		middleware.EnsureUser(users),
	)
	v1.GET("/me", meHandler())
	v1.POST("/onboarding", onboarding.Handle)
	// Goal-first planner: deterministic suggestion + multi-turn chat refinement.
	// The LLM is the language layer only — budget math stays deterministic.
	v1.POST("/onboarding/suggest", plannerHandler.Suggest)
	v1.POST("/planner/chat", plannerHandler.Chat)
	v1.GET("/categories", categoriesHandler.List)
	v1.POST("/categories", categoriesHandler.Create)
	v1.POST("/transactions", transactionsHandler.Create)
	v1.GET("/transactions", transactionsHandler.List)
	v1.GET("/transactions/:id", transactionsHandler.Get)
	v1.PATCH("/transactions/:id", transactionsHandler.Update)
	v1.DELETE("/transactions/:id", transactionsHandler.Delete)
	// Receipt routes accept multipart image uploads, so they get a looser body
	// limit than the global 2M cap (multipart adds boundary/header overhead).
	// The handler still enforces a hard 2MB cap on the decoded image itself.
	v1.POST("/receipts/analyze", receiptsHandler.Analyze, echomw.BodyLimit("5M"))
	v1.POST("/receipts/confirm", receiptsHandler.Confirm, echomw.BodyLimit("5M"))
	v1.GET("/budget/current", budgetHandler.Current)

	return e, nil
}

// healthHandler reports app + DB readiness. Used by Railway, load balancers,
// uptime checks, and Phase 0 smoke testing.
func healthHandler(d Deps) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := d.DB.Ping(ctx); err != nil {
			dbStatus = fmt.Sprintf("error: %v", err)
		}
		return responses.OK(c, map[string]any{
			"status":  "ok",
			"db":      dbStatus,
			"version": "0.0.0",
		})
	}
}

// meHandler is a tiny protected endpoint useful for smoke-testing the JWT
// middleware end-to-end. Will be replaced by a real user lookup in Phase 1.
func meHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return responses.OK(c, map[string]any{
			"user_id": middleware.UserID(c).String(),
		})
	}
}
