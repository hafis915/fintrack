package main

import (
	"context"
	"fmt"

	"github.com/hafis915/fintrack/internal/config"
	"github.com/hafis915/fintrack/internal/database"
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

	e := server.New(server.Deps{Pool: pool, Version: version})
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	if err := e.Start(addr); err != nil {
		panic(err)
	}
}
