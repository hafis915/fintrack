package server

import "github.com/labstack/echo/v4"

func registerRoutes(e *echo.Echo, _ Deps) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]any{
			"data": map[string]string{"status": "ok", "version": "0.1.0", "db": "ok"},
		})
	})
}
