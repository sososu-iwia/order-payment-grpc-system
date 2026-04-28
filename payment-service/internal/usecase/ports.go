package usecase

import (
	"context"
	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	ListByAmountRange(ctx context.Context, minAmount, maxAmount int64) ([]*domain.Payment, error)
}
type IDGenerator interface{ NewID() string }

type PaymentResult struct {
	Status        string
	TransactionID string
}
