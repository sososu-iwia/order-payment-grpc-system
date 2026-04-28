package usecase

import (
	"context"
	"time"

	"order-service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order, idempotencyKey string) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
}

type PaymentClient interface {
	CreatePayment(ctx context.Context, orderID string, amount int64) (*PaymentResult, error)
}

type Clock interface{ Now() time.Time }
type IDGenerator interface{ NewID() string }

type PaymentResult struct {
	Status        string
	TransactionID string
}
