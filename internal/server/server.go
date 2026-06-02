package server

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/hafis915/fintrack/internal/config"
	"github.com/hafis915/fintrack/internal/handler"
	"github.com/hafis915/fintrack/internal/middleware"
	"github.com/hafis915/fintrack/pkg/apperror"
	"github.com/hafis915/fintrack/pkg/response"
	v "github.com/hafis915/fintrack/pkg/validator"
)

type Deps struct {
	Cfg             *config.Config
	Pool            *pgxpool.Pool
	Version         string
	ProfileHandler  *handler.ProfileHandler
	CategoryHandler *handler.CategoryHandler
}

func New(deps Deps) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = v.New()

	e.Use(echomw.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, "X-Request-ID"},
	}))

	e.HTTPErrorHandler = errorHandler
	registerRoutes(e, deps)
	return e
}

func errorHandler(err error, c echo.Context) {
	var ae *apperror.Error
	if errors.As(err, &ae) {
		_ = response.Error(c, ae)
		return
	}
	if he, ok := err.(*echo.HTTPError); ok {
		_ = response.Error(c, &apperror.Error{
			Code:    apperror.CodeInternal,
			Message: he.Error(),
			HTTP:    he.Code,
		})
		return
	}
	_ = response.Error(c, apperror.Internal(err))
}
