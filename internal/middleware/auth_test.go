package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/hafis915/fintrack/internal/middleware"
)

func TestJWT_ValidToken(t *testing.T) {
	secret := "test-secret"
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "123e4567-e89b-12d3-a456-426614174000",
		"aud": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, _ := tok.SignedString([]byte(secret))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	h := middleware.JWT(secret, "authenticated")(func(c echo.Context) error {
		called = true
		require.Equal(t, "123e4567-e89b-12d3-a456-426614174000", c.Get("user_id"))
		return c.NoContent(200)
	})
	require.NoError(t, h(c))
	require.True(t, called)
}

func TestJWT_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := middleware.JWT("s", "authenticated")(func(c echo.Context) error { return nil })(c)
	require.Error(t, err)
}
