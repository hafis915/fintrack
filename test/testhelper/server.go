// Package testhelper provides shared bootstrap for integration tests.
//
// Each test package calls NewTestServer(t) to get a fully wired Echo instance
// backed by the real fintrack_test Postgres database. We use httptest under
// the hood so tests run without binding a port — same code path as production
// but in-process and parallel-safe.
package testhelper

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/ai"
	"github.com/hafis915/fintrack/internal/config"
	"github.com/hafis915/fintrack/internal/server"
	"github.com/hafis915/fintrack/internal/storage"
	"github.com/hafis915/fintrack/pkg/logger"
)

// Fixed test config — deterministic secrets so tokens minted in tests are
// reproducible across runs. Real secrets live in .env and are loaded in prod.
const (
	TestJWTSecret = "test-secret-do-not-use-in-prod-test-secret-do-not-use-in-prod"
	TestJWTIssuer = "fintrack-test"
	// 32-byte AES key (hex). Required by config validation; unused in Phase 0.
	TestEncryptionKey = "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
)

// TestServer bundles an Echo handler + DB pool + config so tests can use any
// of them. Close() releases the DB pool.
type TestServer struct {
	Echo   *echo.Echo
	DB     *pgxpool.Pool
	Config *config.Config
}

// Close releases connections. Always defer this in tests.
func (s *TestServer) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}

// NewTestServer wires up the API against fintrack_test. It fails the test
// (not the suite) if the DB is unreachable so individual packages can be
// skipped without infecting others.
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	cfg := &config.Config{
		Env:                 "test",
		LogLevel:            "warn",
		HTTPHost:            "127.0.0.1",
		HTTPPort:            0,
		DatabaseURL:         dbURL,
		JWTSecret:           TestJWTSecret,
		JWTIssuer:           TestJWTIssuer,
		IncomeEncryptionKey: TestEncryptionKey,
		CORSAllowedOrigins:  "http://localhost:5173",
		AuthLocalEnabled:    true, // exercise the Phase 0 local register/login routes
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("connecting to test db: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("pinging test db (is `make test-db-ensure` run?): %v", err)
	}

	log := logger.New(cfg.LogLevel)
	e, err := server.New(server.Deps{
		Config: cfg,
		Logger: log,
		DB:     pool,
		// Stub out third-party deps so integration tests never hit Claude or MinIO.
		ReceiptAnalyzer: ai.NewStubAnalyzer(),
		Storage:         storage.NewStubStorage(),
	})
	if err != nil {
		pool.Close()
		t.Fatalf("building test server: %v", err)
	}

	return &TestServer{Echo: e, DB: pool, Config: cfg}
}
