package main

import (
	"context"
	"fmt"

	"github.com/hafis915/fintrack/internal/config"
	"github.com/hafis915/fintrack/internal/database"
	"github.com/hafis915/fintrack/internal/domain/user"
	"github.com/hafis915/fintrack/internal/encryption"
	"github.com/hafis915/fintrack/internal/handler"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/internal/server"
	"github.com/hafis915/fintrack/pkg/logger"
)

const version = "0.1.0"

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger.Init(cfg.LogLevel)

	pool, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	enc, err := encryption.New(cfg.IncomeEncryptionKey)
	if err != nil {
		panic(err)
	}

	userRepo := repository.NewUserRepo(pool)
	userSvc := user.NewService(userRepo, enc)
	profileH := &handler.ProfileHandler{Svc: userSvc}

	e := server.New(server.Deps{
		Cfg:            cfg,
		Pool:           pool,
		Version:        version,
		ProfileHandler: profileH,
	})
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	if err := e.Start(addr); err != nil {
		panic(err)
	}
}
