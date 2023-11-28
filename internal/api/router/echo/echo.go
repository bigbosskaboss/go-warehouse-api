package echo

import (
	"log/slog"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"go-warehouse-api/internal/api/handlers"
	"go-warehouse-api/internal/api/middlewares"
	"go-warehouse-api/internal/repo"
	"go-warehouse-api/internal/service"
)

func New(log *slog.Logger, repo *repo.Repo) *echo.Echo {
	router := echo.New()

	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())

	whService := service.New(repo.Storage, log)
	whHandler := handlers.New(whService, log)

	router.POST("/reserve", whHandler.ReserveHandler(), middlewares.ReservationMiddleware)
	router.POST("/release", whHandler.ReleaseHandler(), middlewares.ReservationMiddleware)
	router.POST("/stock", whHandler.StockHandler())

	return router
}
