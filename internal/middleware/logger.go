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
