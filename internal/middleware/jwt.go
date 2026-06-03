package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/pkg/responses"
)

// Context keys for values injected by the JWT middleware.
const (
	CtxKeyUserID = "user_id"
	CtxKeyClaims = "jwt_claims"
)

// FintrackClaims is the JWT body we expect — minimal for now.
// Matches what cmd/mint-jwt produces locally and what Supabase Auth will
// produce in production (Supabase puts the user UUID in the `sub` claim).
type FintrackClaims struct {
	jwt.RegisteredClaims
}

// JWTAuth verifies the Authorization: Bearer <token> header against the
// configured HS256 secret and writes the user UUID + claims into the request
// context. Returns 401 on any failure.
func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Request().Header.Get(echo.HeaderAuthorization)
			if h == "" {
				return responses.Err(c, http.StatusUnauthorized, "missing_token", "Authorization header required")
			}
			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return responses.Err(c, http.StatusUnauthorized, "malformed_token", "expected 'Bearer <token>'")
			}

			claims := &FintrackClaims{}
			tok, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !tok.Valid {
				return responses.Err(c, http.StatusUnauthorized, "invalid_token", "token validation failed")
			}

			uid, err := uuid.Parse(claims.Subject)
			if err != nil {
				return responses.Err(c, http.StatusUnauthorized, "invalid_subject", "sub claim is not a UUID")
			}

			c.Set(CtxKeyUserID, uid)
			c.Set(CtxKeyClaims, claims)
			return next(c)
		}
	}
}

// UserID extracts the authenticated user UUID from the request context.
// Returns the zero UUID if the middleware didn't run — handlers should treat
// that as a programmer error (the route was mounted without JWTAuth).
func UserID(c echo.Context) uuid.UUID {
	if v, ok := c.Get(CtxKeyUserID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}
