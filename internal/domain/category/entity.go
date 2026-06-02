package category

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Name      string
	Icon      *string
	Type      string
	IsDefault bool
	IsActive  bool
	SortOrder int
	CreatedAt time.Time
}

type CreateInput struct {
	UserID uuid.UUID
	Name   string
	Icon   *string
	Type   string
}
