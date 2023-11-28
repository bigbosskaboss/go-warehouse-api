package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"go-warehouse-api/internal/repo"
	"go-warehouse-api/pkg/models"
)

type WarehouseService struct {
	Repo repo.Storage
	log  *slog.Logger
}

func New(repo repo.Storage, log *slog.Logger) *WarehouseService {
	return &WarehouseService{
		Repo: repo,
		log:  log,
	}
}

func (service *WarehouseService) Reserve(reservation *models.Reservation) error {
	service.log.Info("Attempt to reserve Warehouse ID: ", reservation.WarehouseID)
	err := service.Repo.ReserveAtWH(reservation)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "warehouse does not exist"):
			return errors.New("warehouse does not exist")
		case strings.Contains(err.Error(), "not enough items for ID"):
			return errors.New("not enough items in the warehouse")
		case strings.Contains(err.Error(), "warehouse is not available"):
			return errors.New("warehouse is not available")
		case strings.Contains(err.Error(), "item with ID"):
			return errors.New("some item were not found on the warehouse")
		default:
			return errors.New("reservation failed")
		}
	}

	service.log.Info("Reservation completed successfully")
	return nil
}

func (service *WarehouseService) Release(release *models.Reservation) error {

	service.log.Info("Attempt to reserve Warehouse ID:", release.WarehouseID)

	err := service.Repo.ReleaseFromWH(release)
	if err != nil {
		fmt.Println("AAAAAAAAAAAA", err)
		switch {
		case strings.Contains(err.Error(), "warehouse does not exist"):
			return errors.New("warehouse does not exist")
		case strings.Contains(err.Error(), "not enough reserved items for ID"):
			return errors.New("not enough reserved items in the warehouse")
		case strings.Contains(err.Error(), "warehouse is not available"):
			return errors.New("warehouse is not available")
		case strings.Contains(err.Error(), "item with ID"):
			return errors.New("some item were not found on the warehouse")
		default:
			return errors.New("release failed")
		}
	}

	service.log.Info("Release completed successfully")
	return nil
}

func (service *WarehouseService) GetWarehouseStock(warehouse *models.Warehouse) (response *[]models.Item, err error) {
	response, err = service.Repo.GetStock(warehouse.ID)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("warehouse is empty")
	}

	return response, nil
}
