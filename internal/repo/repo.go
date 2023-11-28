package repo

import (
	"github.com/jackc/pgx/v5"

	"go-warehouse-api/pkg/models"
)

type Storage interface {
	Reservater
	StockGetter
	WhValidater
}

type Reservater interface {
	ReserveAtWH(reservation *models.Reservation) error
	ReleaseFromWH(release *models.Reservation) error
}

type StockGetter interface {
	GetStock(id int) (items *[]models.Item, err error)
}
type WhValidater interface {
	ValidateWH(tx pgx.Tx, id int) (isAvailable bool, err error)
}

type Repo struct {
	Storage
}

func New(storage Storage) *Repo {
	return &Repo{
		Storage: storage,
	}
}
