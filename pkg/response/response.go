package response

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/pkg/apperror"
)

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Total     int    `json:"total,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
}

type Envelope struct {
	Data any   `json:"data,omitempty"`
	Meta *Meta `json:"meta,omitempty"`
}

type ErrorEnvelope struct {
	Error errBody `json:"error"`
}

type errBody struct {
	Code    apperror.Code     `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, Envelope{Data: data, Meta: meta(c)})
}

func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, Envelope{Data: data, Meta: meta(c)})
}

func List(c echo.Context, data any, total, page, perPage int) error {
	m := meta(c)
	m.Total, m.Page, m.PerPage = total, page, perPage
	return c.JSON(http.StatusOK, Envelope{Data: data, Meta: m})
}

func meta(c echo.Context) *Meta {
	id, _ := c.Get("request_id").(string)
	return &Meta{RequestID: id}
}

func Error(c echo.Context, e *apperror.Error) error {
	if e.HTTP == 0 {
		e.HTTP = http.StatusInternalServerError
	}
	return c.JSON(e.HTTP, ErrorEnvelope{Error: errBody{
		Code:    e.Code,
		Message: e.Message,
		Fields:  e.Fields,
	}})
}
