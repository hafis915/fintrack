package config_test

import (
	"os"
	"testing"

	"github.com/hafis915/fintrack/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("SUPABASE_JWT_SECRET", "s")
	t.Setenv("INCOME_ENCRYPTION_KEY", "k")
	t.Setenv("OPENROUTER_API_KEY", "a")

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
