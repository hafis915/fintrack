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

	"github.com/hafis915/fintrack/internal/config"
	"github.com/hafis915/fintrack/internal/encryption"
	"github.com/hafis915/fintrack/internal/handler"
	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/responses"
)

// Deps groups everything a server needs. Wired in main.go and passed here —
// no global state, no init().
type Deps struct {
	Config *config.Config
	Logger zerolog.Logger
	DB     *pgxpool.Pool
}

// New returns a configured Echo instance with global middleware mounted
// and the public + protected route groups registered. Returns an error if
// any upstream dependency (today: only the income cipher) fails to construct.
func New(d Deps) (*echo.Echo, error) {
	cipher, err := encryption.NewCipherFromHex(d.Config.IncomeEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("building income cipher: %w", err)
	}

	users := repository.NewUsersRepo(d.DB)
	userProfiles := repository.NewUserProfilesRepo(d.DB)
	categories := repository.NewCategoriesRepo(d.DB)
	budgetPlans := repository.NewBudgetPlansRepo(d.DB)

	onboarding := handler.NewOnboarding(handler.OnboardingDeps{
		Users:        users,
		UserProfiles: userProfiles,
		Categories:   categories,
		BudgetPlans:  budgetPlans,
		Cipher:       cipher,
	})
	categoriesHandler := handler.NewCategories(categories)

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

	// Protected routes live under /v1 and require a valid JWT.
	v1 := e.Group("/v1", middleware.JWTAuth(d.Config.JWTSecret))
	v1.GET("/me", meHandler())
	v1.POST("/onboarding", onboarding.Handle)
	v1.GET("/categories", categoriesHandler.List)

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
