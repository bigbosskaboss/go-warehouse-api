package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"

	"go-warehouse-api/internal/config"
	"go-warehouse-api/internal/lib/logger/sl"
)

func RunHTTP(cfg *config.Config, router *echo.Echo) error {

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start warehouse-service")
		}
	}()

	log.Info("warehouse-service started at ", slog.String("address", cfg.Address))

	<-done
	log.Info("stopping warehouse-service")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop warehouseRepo-service", sl.Err(err))

		return err
	}

	log.Info("warehouse-service stopped")
	return nil
}
