package server

import (
	"github.com/labstack/echo/v4"

	"github.com/hafis915/fintrack/internal/handler"
	"github.com/hafis915/fintrack/internal/middleware"
)

func registerRoutes(e *echo.Echo, d Deps) {
	health := &handler.HealthHandler{Pool: d.Pool, Version: d.Version}
	e.GET("/health", health.Get)

	v1 := e.Group("/v1")
	v1.Use(middleware.JWT(d.Cfg.SupabaseJWTSecret, d.Cfg.SupabaseJWTAudience))

	v1.GET("/profile", d.ProfileHandler.Get)
	v1.PATCH("/profile", d.ProfileHandler.Update)
	v1.PUT("/profile/income", d.ProfileHandler.UpdateIncome)

	v1.GET("/categories", d.CategoryHandler.List)
	v1.POST("/categories", d.CategoryHandler.Create)
	v1.DELETE("/categories/:id", d.CategoryHandler.Delete)

	v1.POST("/onboarding", d.OnboardingHandler.Submit)
	v1.GET("/budget/current", d.BudgetHandler.Current)

	v1.GET("/transactions", d.TransactionHandler.List)
	v1.POST("/transactions", d.TransactionHandler.Create)
	v1.PATCH("/transactions/:id", d.TransactionHandler.Update)
	v1.DELETE("/transactions/:id", d.TransactionHandler.Delete)
}
