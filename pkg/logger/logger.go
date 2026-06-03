package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// New returns a JSON structured logger that writes to stdout.
// In dev, level can be lowered via env; in prod, INFO is the default.
func New(level string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil || lvl == zerolog.NoLevel {
		lvl = zerolog.InfoLevel
	}
	return zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Str("service", "fintrack-api").
		Logger()
}
