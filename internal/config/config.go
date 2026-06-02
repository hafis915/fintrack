package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv              string
	HTTPPort            int
	DatabaseURL         string
	SupabaseJWTSecret   string
	SupabaseJWTAudience string
	IncomeEncryptionKey string
	AnthropicAPIKey     string
	AnthropicModel      string
	LogLevel            string
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
