package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const HeaderRequestID = "X-Request-ID"

// RequestID assigns a UUID per request, echoes it back in the response header,
// and exposes it in the context for downstream logging.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rid := c.Request().Header.Get(HeaderRequestID)
			if rid == "" {
				rid = uuid.NewString()
			}
			c.Response().Header().Set(HeaderRequestID, rid)
			c.Set("request_id", rid)
			return next(c)
		}
	}
}
