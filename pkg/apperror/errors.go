package apperror

import "errors"

type Code string

const (
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeInternal     Code = "INTERNAL_ERROR"
	CodeAI           Code = "AI_ERROR"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
)

type Error struct {
	Code    Code
	Message string
	Fields  map[string]string
	cause   error
	HTTP    int
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.cause }

func wrap(code Code, http int, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, HTTP: http, cause: cause}
}

func NotFound(resource, detail string) *Error {
	return wrap(CodeNotFound, 404, resource+" not found: "+detail, ErrNotFound)
}
func Unauthorized(msg string) *Error    { return wrap(CodeUnauthorized, 401, msg, ErrUnauthorized) }
func Forbidden(msg string) *Error       { return wrap(CodeForbidden, 403, msg, ErrForbidden) }
func Conflict(msg string) *Error        { return wrap(CodeConflict, 409, msg, ErrConflict) }
func Internal(cause error) *Error       { return wrap(CodeInternal, 500, "internal error", cause) }
func AI(msg string, cause error) *Error { return wrap(CodeAI, 502, msg, cause) }

func Validation(msg string, fields map[string]string) *Error {
	e := wrap(CodeValidation, 400, msg, nil)
	e.Fields = fields
	return e
}
