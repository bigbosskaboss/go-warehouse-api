package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo"

	"go-warehouse-api/internal/lib/logger/sl"
	"go-warehouse-api/internal/service"
	"go-warehouse-api/pkg/models"
)

type WarehouseHandler struct {
	Service *service.WarehouseService
	log     *slog.Logger
}

func New(service *service.WarehouseService, log *slog.Logger) *WarehouseHandler {
	return &WarehouseHandler{
		Service: service,
		log:     log,
	}
}

func (wh *WarehouseHandler) ReserveHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		const fn = "handlers.Reserve"
		wh.log = wh.log.With("fn", fn)

		request := c.Get("request").(models.Reservation)

		// Передаем запрос слою сервис
		err := wh.Service.Reserve(&request)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, "items were successfully reserved")
	}
}

func (wh *WarehouseHandler) ReleaseHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		const fn = "handlers.Release"
		wh.log = wh.log.With("fn", fn)
		request := c.Get("request").(models.Reservation)

		wh.log.Info("request validated", slog.Any("request", request))

		// Передаем запрос слою сервиса
		err := wh.Service.Release(&request)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, "items were successfully released")
	}
}

func (wh *WarehouseHandler) StockHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		const fn = "handlers.GetWarehouseStock"
		wh.log = wh.log.With("fn", fn)

		var warehouse models.Warehouse
		err := c.Bind(&warehouse)
		if err != nil {
			wh.log.Error("failed to decode request body", sl.Err(err))
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to decode request body"})
		}
		wh.log.Debug("request body decoded", slog.Any("request", warehouse))

		// Валидируем тело запроса
		v := validator.New()
		if err := v.Struct(warehouse); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		response, err := wh.Service.GetWarehouseStock(&warehouse)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		c.JSON(http.StatusOK, response)

		return nil
	}
}
