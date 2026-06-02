package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/pkg/apperror"
)

func JWT(secret, audience string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				return apperror.Unauthorized("missing bearer token")
			}
			raw := strings.TrimPrefix(h, "Bearer ")
			claims := jwt.MapClaims{}
			tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, apperror.Unauthorized("unexpected signing method")
				}
				return []byte(secret), nil
			})
			if err != nil || !tok.Valid {
				return apperror.Unauthorized("invalid token")
			}
			if aud, ok := claims["aud"].(string); !ok || aud != audience {
				return apperror.Unauthorized("audience mismatch")
			}
			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				return apperror.Unauthorized("missing subject")
			}
			c.Set("user_id", sub)
			return next(c)
		}
	}
}
