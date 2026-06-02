package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get("X-Request-ID")
			if id == "" {
				id = uuid.NewString()
			}
			c.Set("request_id", id)
			c.Response().Header().Set("X-Request-ID", id)
			return next(c)
		}
	}
}
