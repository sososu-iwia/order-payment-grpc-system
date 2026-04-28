package domain

import (
	"errors"
	"time"
)

const (
	OrderStatusPending   = "Pending"
	OrderStatusPaid      = "Paid"
	OrderStatusFailed    = "Failed"
	OrderStatusCancelled = "Cancelled"
)

type Order struct {
	ID         string
	CustomerID string
	ItemName   string
	Amount     int64
	Status     string
	CreatedAt  time.Time
}

func NewOrder(id, customerID, itemName string, amount int64, now time.Time) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("customer_id is required")
	}
	if itemName == "" {
		return nil, errors.New("item_name is required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	return &Order{
		ID:         id,
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     OrderStatusPending,
		CreatedAt:  now,
	}, nil
}

func (o *Order) MarkPaid()   { o.Status = OrderStatusPaid }
func (o *Order) MarkFailed() { o.Status = OrderStatusFailed }
func (o *Order) Cancel() error {
	if o.Status != OrderStatusPending {
		return errors.New("only pending orders can be cancelled")
	}
	o.Status = OrderStatusCancelled
	return nil
}
