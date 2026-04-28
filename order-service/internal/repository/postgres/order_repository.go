package postgres

import (
	"context"
	"database/sql"
	"errors"

	"order-service/internal/domain"
)

type OrderRepository struct{ db *sql.DB }

func NewOrderRepository(db *sql.DB) *OrderRepository { return &OrderRepository{db: db} }

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order, idempotencyKey string) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, order.ID, order.CustomerID, order.ItemName, order.Amount, order.Status, order.CreatedAt, nullIfEmpty(idempotencyKey))
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE id = $1`, id)
	var order domain.Order
	if err := row.Scan(&order.ID, &order.CustomerID, &order.ItemName, &order.Amount, &order.Status, &order.CreatedAt); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *OrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE idempotency_key = $1`, key)
	var order domain.Order
	if err := row.Scan(&order.ID, &order.CustomerID, &order.ItemName, &order.Amount, &order.Status, &order.CreatedAt); err != nil {
		return nil, err
	}
	return &order, nil
}

func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
