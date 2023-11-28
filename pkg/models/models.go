package models

type Warehouse struct {
	ID        int    `json:"warehouse_id" validate:"required,number"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

type Item struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	UniqueCode string `json:"unique_code"`
	Amount     int    `json:"amount"`
}

type Reservation struct {
	WarehouseID int               `json:"warehouse_id" validate:"required,number"`
	ItemsList   []ItemReservation `json:"cart" validate:"required,min=1,dive,required"`
}

type ItemReservation struct {
	UniqueCode int  `json:"unique_code" validate:"required"`
	Amount     *int `json:"amount"`
}
