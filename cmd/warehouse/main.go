package main

import (
	"log/slog"
	"os"

	"go-warehouse-api/internal/api/router/echo"
	"go-warehouse-api/internal/api/server"
	"go-warehouse-api/internal/config"
	"go-warehouse-api/internal/lib/logger/sl"
	"go-warehouse-api/internal/repo"
	"go-warehouse-api/internal/repo/postgres"
)

func main() {
	cfg := config.MustLoad()
	log := sl.Setup(cfg.Env)

	log.Info("starting warehouse api", slog.String("env", cfg.Env))

	storage, err := postgres.New(cfg.DatabaseURL, log)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	log.Info("storage initialized")

	repository := repo.New(storage)

	router := echo.New(log, repository)
	if err := server.RunHTTP(cfg, router); err != nil {
		log.Error("failed to start warehouse service", sl.Err(err))
		os.Exit(1)
	}
}
