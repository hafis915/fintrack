package testhelper

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// MintToken signs an HS256 JWT with the test secret/issuer. ttl=0 uses 1h.
// Override fields by mutating the returned claims before signing if a test
// needs an expired or otherwise-invalid token; for the common case use
// MintExpiredToken / MintTokenForUser below.
func MintToken(t *testing.T, sub uuid.UUID, ttl time.Duration) string {
	t.Helper()
	if ttl == 0 {
		ttl = time.Hour
	}
	return mint(t, sub, time.Now(), ttl, TestJWTSecret)
}

// MintTokenForUser is the most common form: a fresh, 1h-valid token for a
// caller-supplied user UUID.
func MintTokenForUser(t *testing.T, sub uuid.UUID) string {
	t.Helper()
	return MintToken(t, sub, time.Hour)
}

// MintExpiredToken returns a token whose exp is in the past, for testing
// the rejection path of the JWT middleware.
func MintExpiredToken(t *testing.T, sub uuid.UUID) string {
	t.Helper()
	// IssuedAt + NotBefore in the past, ExpiresAt also in the past.
	past := time.Now().Add(-2 * time.Hour)
	return mint(t, sub, past, time.Hour, TestJWTSecret)
}

// MintTokenWithSecret signs with a caller-supplied secret. Use to verify
// that the API rejects tokens signed by anyone but us.
func MintTokenWithSecret(t *testing.T, sub uuid.UUID, secret string) string {
	t.Helper()
	return mint(t, sub, time.Now(), time.Hour, secret)
}

func mint(t *testing.T, sub uuid.UUID, issued time.Time, ttl time.Duration, secret string) string {
	t.Helper()
	claims := jwt.RegisteredClaims{
		Issuer:    TestJWTIssuer,
		Subject:   sub.String(),
		IssuedAt:  jwt.NewNumericDate(issued),
		NotBefore: jwt.NewNumericDate(issued),
		ExpiresAt: jwt.NewNumericDate(issued.Add(ttl)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}
	return signed
}
