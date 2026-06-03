package integration_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/hafis915/fintrack/test/testhelper"
)

// mintNonUUIDSubToken signs a token whose `sub` claim is intentionally not a
// UUID. The signature is valid (so middleware accepts the token cryptographically)
// — only the subject-parse step should reject it. Used by
// TestIntegration_Me_RejectsNonUUIDSubject.
func mintNonUUIDSubToken(t *testing.T) string {
	t.Helper()
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    testhelper.TestJWTIssuer,
		Subject:   "not-a-uuid",
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(testhelper.TestJWTSecret))
	if err != nil {
		t.Fatalf("signing non-uuid token: %v", err)
	}
	return signed
}
