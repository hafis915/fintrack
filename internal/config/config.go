package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration loaded from env + .env file.
// All keys must be present (or have a sane default) for the app to boot.
type Config struct {
	Env      string `mapstructure:"ENV"`
	LogLevel string `mapstructure:"LOG_LEVEL"`

	HTTPHost string `mapstructure:"HTTP_HOST"`
	HTTPPort int    `mapstructure:"HTTP_PORT"`

	DatabaseURL string `mapstructure:"DATABASE_URL"`

	JWTSecret string `mapstructure:"JWT_SECRET"`
	JWTIssuer string `mapstructure:"JWT_ISSUER"`

	IncomeEncryptionKey string `mapstructure:"INCOME_ENCRYPTION_KEY"`

	StorageEndpoint  string `mapstructure:"STORAGE_ENDPOINT"`
	StorageAccessKey string `mapstructure:"STORAGE_ACCESS_KEY"`
	StorageSecretKey string `mapstructure:"STORAGE_SECRET_KEY"`
	StorageBucket    string `mapstructure:"STORAGE_BUCKET"`
	StorageUseSSL    bool   `mapstructure:"STORAGE_USE_SSL"`

	CORSAllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`
}

// Load reads env vars + optional .env file and returns the parsed Config.
// Unknown keys are ignored; missing required keys produce a clear error.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults.
	v.SetDefault("ENV", "dev")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("HTTP_HOST", "0.0.0.0")
	v.SetDefault("HTTP_PORT", 8080)
	v.SetDefault("JWT_ISSUER", "fintrack-local")
	v.SetDefault("STORAGE_USE_SSL", false)
	v.SetDefault("STORAGE_BUCKET", "receipts")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173")

	// .env is optional — we may run with only real env vars in prod.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	missing := []string{}
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.IncomeEncryptionKey == "" {
		missing = append(missing, "INCOME_ENCRYPTION_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}
