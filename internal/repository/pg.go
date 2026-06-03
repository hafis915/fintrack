// Package repository provides thin, domain-typed wrappers around sqlc's
// generated code. Handlers depend on these interfaces; tests can substitute
// real DB pools for fast in-process integration tests.
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Helpers to bridge pgtype ↔ standard library types so domain code never
// has to import pgx. Kept here so every repository file stays narrow.

func toPgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func fromPgUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return uuid.UUID(p.Bytes)
}

func fromPgTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}
