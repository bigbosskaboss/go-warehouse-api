package middlewares

import (
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"

	"go-warehouse-api/internal/lib/logger/sl"
	"go-warehouse-api/pkg/models"
)

func ReservationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request models.Reservation
		err := c.Bind(&request)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to decode request body"})
		}
		log.Debug("request body decoded", slog.Any("request", request))

		//Проверяем указано ли количество товаров для освобождения
		for i := range request.ItemsList {
			if request.ItemsList[i].Amount == nil {
				defaultAmount := 1
				request.ItemsList[i].Amount = &defaultAmount
			}
		}

		// Валидируем тело запроса
		v := validator.New()
		if err := v.Struct(request); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		c.Set("request", request)
		return next(c)
	}
}
