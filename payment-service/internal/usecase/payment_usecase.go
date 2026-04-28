package usecase

import (
	"context"
	"errors"
	"payment-service/internal/domain"
)

var ErrPaymentNotFound = errors.New("payment not found")
var ErrInvalidAmountRange = errors.New("min_amount must be less than or equal to max_amount")

type PaymentUsecase struct {
	repo PaymentRepository
	ids  IDGenerator
}

func NewPaymentUsecase(repo PaymentRepository, ids IDGenerator) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, ids: ids}
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, orderID string, amount int64) (*domain.Payment, error) {
	payment, err := domain.NewPayment(uc.ids.NewID(), orderID, uc.ids.NewID(), amount)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (uc *PaymentUsecase) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	payment, err := uc.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}
	return payment, nil
}

func (uc *PaymentUsecase) ListPayments(ctx context.Context, minAmount, maxAmount int64) ([]*domain.Payment, error) {
	if minAmount > 0 && maxAmount > 0 && minAmount > maxAmount {
		return nil, ErrInvalidAmountRange
	}
	return uc.repo.ListByAmountRange(ctx, minAmount, maxAmount)
}
