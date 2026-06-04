// mint-jwt: a tiny CLI that issues an HS256 JWT for local development.
//
// Usage:
//
//	go run ./cmd/mint-jwt -sub <user-uuid> [-secret <hex>] [-ttl 24h]
//
// The token is signed with the same secret the API uses for validation,
// so a token minted here is accepted by /v1/* routes immediately.
//
// In production this is replaced by Supabase Auth (see ADR-014).
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/hafis915/fintrack/internal/auth"
	"github.com/hafis915/fintrack/internal/config"
)

func main() {
	sub := flag.String("sub", "", "subject (user UUID). If empty, a fresh UUID is generated.")
	secret := flag.String("secret", "", "signing secret (overrides JWT_SECRET from env/.env)")
	issuer := flag.String("iss", "", "issuer (overrides JWT_ISSUER from env/.env)")
	ttl := flag.Duration("ttl", 24*time.Hour, "token lifetime")
	flag.Parse()

	signingSecret := *secret
	signingIssuer := *issuer
	if signingSecret == "" || signingIssuer == "" {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, "loading config:", err)
			os.Exit(1)
		}
		if signingSecret == "" {
			signingSecret = cfg.JWTSecret
		}
		if signingIssuer == "" {
			signingIssuer = cfg.JWTIssuer
		}
	}

	var subject uuid.UUID
	if *sub == "" {
		subject = uuid.New()
	} else {
		parsed, err := uuid.Parse(*sub)
		if err != nil {
			fmt.Fprintln(os.Stderr, "subject must be a UUID:", err)
			os.Exit(1)
		}
		subject = parsed
	}

	expires := time.Now().UTC().Add(*ttl)
	signed, err := auth.Mint(signingSecret, signingIssuer, subject, *ttl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "sub:      %s\n", subject.String())
	fmt.Fprintf(os.Stderr, "iss:      %s\n", signingIssuer)
	fmt.Fprintf(os.Stderr, "expires:  %s\n", expires.Format(time.RFC3339))
	fmt.Println(signed)
}
