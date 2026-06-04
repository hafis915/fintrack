// Package auth holds the local-first token minting used by Phase 0 email
// register/login and the mint-jwt CLI. It produces the exact HS256 JWT that
// internal/middleware/jwt.go validates: RegisteredClaims with iss/sub/iat/nbf/exp,
// signed with the shared JWT secret. Real Supabase Auth replaces this in v2
// (ADR-014); the middleware's verification path stays unchanged.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Mint issues a signed HS256 token for the given user. The claim shape mirrors
// cmd/mint-jwt exactly so middleware.JWTAuth accepts it without changes:
// Issuer=issuer, Subject=userID, IssuedAt=NotBefore=now, ExpiresAt=now+ttl.
func Mint(secret, issuer string, userID uuid.UUID, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return signed, nil
}
