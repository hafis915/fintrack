package server

import (
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/handler"
)

func registerRoutes(e *echo.Echo, deps Deps) {
	health := &handler.HealthHandler{Pool: deps.Pool, Version: deps.Version}
	e.GET("/health", health.Get)
}
