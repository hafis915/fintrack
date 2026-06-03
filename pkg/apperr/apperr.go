package apperr

import "errors"

// Sentinel errors used across domain + handler boundaries.
// Domain layer returns these; handlers map them to HTTP status + responses.ErrorBody.

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInternal          = errors.New("internal error")
	ErrUpstreamUnavailable = errors.New("upstream unavailable")
)
