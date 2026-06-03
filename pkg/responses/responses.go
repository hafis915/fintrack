package responses

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Envelope is the shared JSON response shape for the API.
// Either Data or Error is set, never both.
type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, Envelope{Data: data})
}

func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, Envelope{Data: data})
}

func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func Err(c echo.Context, status int, code, message string) error {
	return c.JSON(status, Envelope{Error: &ErrorBody{Code: code, Message: message}})
}
