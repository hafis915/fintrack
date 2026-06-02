package handler

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/pkg/response"
)

type HealthHandler struct {
	Pool    *pgxpool.Pool
	Version string
}

func (h *HealthHandler) Get(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Second)
	defer cancel()
	dbStatus := "ok"
	if err := h.Pool.Ping(ctx); err != nil {
		dbStatus = "down"
	}
	return response.OK(c, map[string]string{
		"status":  "ok",
		"version": h.Version,
		"db":      dbStatus,
	})
}
