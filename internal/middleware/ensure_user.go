package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/responses"
)

// EnsureUser upserts a users row for the authenticated subject before
// downstream handlers run. Without this, the first /v1 request after a
// fresh JWT mint (local dev / e2e) would hit a foreign-key violation
// when it tries to insert into a table that references users(id).
//
// In production with Supabase, the row will already be created via auth
// hooks — this middleware becomes a no-op (idempotent upsert). It stays
// in the stack so the local dev path and prod path use one code path.
//
// Must be mounted AFTER JWTAuth — relies on UserID() being populated.
func EnsureUser(users repository.UsersRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			uid := UserID(c)
			if uid == uuid.Nil {
				return responses.Err(c, http.StatusUnauthorized, "unauthorized", "auth context missing")
			}
			// Email defaults to <uuid>@local; Supabase fills the real value later.
			if _, err := users.Upsert(c.Request().Context(), uid, uid.String()+"@local"); err != nil {
				return responses.Err(c, http.StatusInternalServerError, "user_bootstrap_failed", err.Error())
			}
			return next(c)
		}
	}
}
