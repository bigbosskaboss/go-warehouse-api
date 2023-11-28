package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"go-warehouse-api/pkg/models"
)

type PostgreSQL struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func New(storagePath string, log *slog.Logger) (*PostgreSQL, error) {
	const fn = "repo.postgres.New"
	pool, err := pgxpool.New(context.Background(), storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	db := &PostgreSQL{
		pool: pool,
		log:  log,
	}

	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("%s: failed to create tables: %w", fn, err)
	}

	return db, nil

}

func (p *PostgreSQL) ReserveAtWH(reservation *models.Reservation) (err error) {
	p.log.Info("Initiating a new reservation...")
	tx, err := p.pool.Begin(context.Background())
	if err != nil {
		p.log.Debug("failed to start a transaction: %v", err)
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer handleTransaction(tx, &err)

	p.log.Debug("Checking if warehouse with ID %d is available...", reservation.WarehouseID)
	isAvailable, err := p.ValidateWH(tx, reservation.WarehouseID)
	if err != nil {
		return err
	}
	if !isAvailable {
		p.log.Debug("Warehouse with ID %d is not available", reservation.WarehouseID)
		return errors.New("warehouse is not available")
	}
	p.log.Debug("Warehouse with ID %d is available", reservation.WarehouseID)

	var itemIDs []int
	for _, item := range reservation.ItemsList {
		itemIDs = append(itemIDs, item.UniqueCode)
	}

	query := `SELECT item_id, (amount - reserved_amount) as available_amount
			 FROM warehouse_items
			 WHERE warehouse_id = $1 AND item_id = ANY($2)
			 FOR UPDATE
			 `

	rows, err := tx.Query(context.Background(), query, reservation.WarehouseID, itemIDs)
	if err != nil {
		p.log.Debug("Failed to fetch warehouse items: %v", err)
		return fmt.Errorf("failed to fetch warehouse items: %w", err)
	}
	defer rows.Close()

	foundItems := make(map[int]int)
	for rows.Next() {
		var itemID, availableAmount int
		err := rows.Scan(&itemID, &availableAmount)
		if err != nil {
			return fmt.Errorf("failed to scan warehouse items: %w", err)
		}
		foundItems[itemID] = availableAmount
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed during rows iteration: %w", err)
	}

	for _, item := range reservation.ItemsList {
		availableAmount, found := foundItems[item.UniqueCode]
		if !found {
			p.log.Debug("Item with ID %d is not found on the warehouse", item.UniqueCode)
			return fmt.Errorf("item with ID %d is not found on the warehouse", item.UniqueCode)
		}
		if *item.Amount > availableAmount {
			p.log.Debug("Not enough items for ID %d in the warehouse", item.UniqueCode)
			return fmt.Errorf("not enough items for ID %d", item.UniqueCode)
		}
		query = `UPDATE warehouse_items
				SET reserved_amount = reserved_amount + $1
				WHERE warehouse_id = $2 AND item_id = $3 AND (amount - reserved_amount) >= $1`

		_, err = tx.Exec(context.Background(), query, item.Amount, reservation.WarehouseID, item.UniqueCode)
		if err != nil {
			log.Printf("Transaction failed: %v", err)
			return fmt.Errorf("failed to update warehouse items: %w", err)
		} else {
			log.Println("Reservation completed successfully")
		}
	}
	return nil
}

func (p *PostgreSQL) ReleaseFromWH(release *models.Reservation) (err error) {
	p.log.Info("Initiating a release...")
	tx, err := p.pool.Begin(context.Background())
	if err != nil {
		p.log.Debug("Failed to start a transaction: %v", err)
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer handleTransaction(tx, &err)

	p.log.Debug("Checking if warehouse with ID %d is available...", release.WarehouseID)
	isAvailable, err := p.ValidateWH(tx, release.WarehouseID)
	if err != nil {
		return err
	}
	if !isAvailable {
		p.log.Debug("Warehouse with ID %d is not available", release.WarehouseID)
		return errors.New("warehouse is not available")
	}
	p.log.Debug("Warehouse with ID %d is available", release.WarehouseID)

	var itemIDs []int
	for _, item := range release.ItemsList {
		itemIDs = append(itemIDs, item.UniqueCode)
	}

	query := `SELECT item_id, reserved_amount
			 FROM warehouse_items
			 WHERE warehouse_id = $1 AND item_id = ANY($2)
			 FOR UPDATE
			 `

	rows, err := tx.Query(context.Background(), query, release.WarehouseID, itemIDs)
	if err != nil {
		p.log.Debug("Failed to fetch warehouse items: %v", err)
		return fmt.Errorf("failed to fetch warehouse items: %w", err)
	}
	defer rows.Close()

	foundItems := make(map[int]int)
	for rows.Next() {
		var itemID, reservedAmount int
		err := rows.Scan(&itemID, &reservedAmount)
		if err != nil {
			return fmt.Errorf("failed to scan warehouse items: %w", err)
		}
		foundItems[itemID] = reservedAmount
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed during rows iteration: %w", err)
	}

	for _, item := range release.ItemsList {
		reservedAmount, found := foundItems[item.UniqueCode]
		if !found {
			p.log.Debug("Item with ID %d is not found on the warehouse", item.UniqueCode)
			return fmt.Errorf("item with ID %d is not found on the warehouse", item.UniqueCode)
		}
		log.Printf("Amount to release: %d, Reserved amount: %d", *item.Amount, reservedAmount)

		if *item.Amount > reservedAmount {
			p.log.Debug("Not enough reserved items for ID %d in the warehouse", item.UniqueCode)
			return fmt.Errorf("not enough reserved items for ID %d", item.UniqueCode)
		}

		// Updated to decrease the reserved_amount and amount
		query = `UPDATE warehouse_items
				 SET reserved_amount = reserved_amount - $1, 
				     amount = amount - $1
				 WHERE warehouse_id = $2 AND item_id = $3 AND reserved_amount >= $1`

		_, err = tx.Exec(context.Background(), query, *item.Amount, release.WarehouseID, item.UniqueCode)
		if err != nil {
			log.Printf("Transaction failed: %v", err)
			return fmt.Errorf("failed to update warehouse items: %w", err)
		} else {
			log.Println("Release completed successfully")
		}
	}
	return nil
}

func (p *PostgreSQL) GetStock(id int) (response *[]models.Item, err error) {
	p.log.Info("Initiating getting stock...")
	tx, err := p.pool.Begin(context.Background())
	if err != nil {
		p.log.Debug("Failed to start a transaction: %v", err)
		return nil, fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer handleTransaction(tx, &err)

	p.log.Debug("Checking if warehouse with ID %d is available...", id)
	isAvailable, err := p.ValidateWH(tx, id)
	if err != nil {
		return nil, err
	}
	if !isAvailable {
		p.log.Debug("Warehouse with ID %d is not available", id)
		return nil, errors.New("warehouse is not available")
	}
	p.log.Debug("Warehouse with ID %d is available", id)

	query := `select item_id, name, size, (amount - reserved_amount) as total
		from warehouse_items join items on warehouse_items.item_id = items.id
		where warehouse_items.warehouse_id = $1
		order by total desc`
	rows, err := tx.Query(context.Background(), query, id)
	if err != nil {
		p.log.Debug("Failed to fetch warehouse items: %v", err)
		return nil, fmt.Errorf("failed to fetch warehouse items: %w", err)
	}
	defer rows.Close()
	var itemsList []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.UniqueCode, &item.Name, &item.Size, &item.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan warehouse items: %w", err)
		}
		itemsList = append(itemsList, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed during rows iteration: %w", err)
	}
	response = &itemsList
	return response, nil
}

func (p *PostgreSQL) ValidateWH(tx pgx.Tx, id int) (isAvailable bool, err error) {
	query := "SELECT is_available FROM warehouses WHERE id = $1 FOR UPDATE"
	err = tx.QueryRow(context.Background(), query, id).Scan(&isAvailable)
	if err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return false, errors.New("warehouse does not exist")
		}
		return false, fmt.Errorf("failed to fetch warehouse availability: %w", err)
	}

	if !isAvailable {
		p.log.Debug("Warehouse with ID %d is not available", id)
		return false, errors.New("warehouse is not available")
	}
	return true, nil
}

func handleTransaction(tx pgx.Tx, err *error) {
	if p := recover(); p != nil {
		_ = tx.Rollback(context.Background())
	} else if *err != nil {
		_ = tx.Rollback(context.Background())
	} else {
		*err = tx.Commit(context.Background())
		if *err != nil {
			log.Printf("failed to commit transaction: %v", *err)
		}
	}
}

func (p *PostgreSQL) createTables() error {
	const fn = "PostgreSQL.createTables"
	query := `
CREATE TABLE if not exists warehouses(
    id SERIAL PRIMARY KEY, 
    name VARCHAR(255),
    is_available BOOLEAN
);
CREATE TABLE if not exists items ( 
	id INT PRIMARY KEY UNIQUE,
    name VARCHAR(255), 
    size VARCHAR(255) 
);
CREATE TABLE if not exists warehouse_items (
    id SERIAL PRIMARY KEY, 
    warehouse_id INT, 
    item_id INT, 
    amount INT,
	reserved_amount INT,
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id), 
    FOREIGN KEY (item_id) REFERENCES items(id)
);
`
	_, err := p.pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("%s: failed to create tables: %w", fn, err)
	}

	return nil

}
