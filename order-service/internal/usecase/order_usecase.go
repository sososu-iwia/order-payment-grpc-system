package usecase

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"order-service/internal/domain"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrServiceUnavailable = errors.New("payment service unavailable")

type OrderUsecase struct {
	repo          OrderRepository
	paymentClient PaymentClient
	clock         Clock
	ids           IDGenerator
}

func NewOrderUsecase(repo OrderRepository, paymentClient PaymentClient, clock Clock, ids IDGenerator) *OrderUsecase {
	return &OrderUsecase{repo: repo, paymentClient: paymentClient, clock: clock, ids: ids}
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, int, error) {
	if idempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(ctx, idempotencyKey)
		if err == nil && existing != nil {
			return existing, http.StatusOK, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, http.StatusInternalServerError, err
		}
	}

	order, err := domain.NewOrder(uc.ids.NewID(), customerID, itemName, amount, uc.clock.Now())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := uc.repo.Create(ctx, order, idempotencyKey); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	paymentResult, err := uc.paymentClient.CreatePayment(ctx, order.ID, order.Amount)
	if err != nil {
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed)
		order.MarkFailed()
		return order, http.StatusServiceUnavailable, ErrServiceUnavailable
	}

	if paymentResult.Status == "Authorized" {
		order.MarkPaid()
	} else {
		order.MarkFailed()
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return order, http.StatusCreated, nil
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (uc *OrderUsecase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if err := order.Cancel(); err != nil {
		return nil, err
	}
	if err := uc.repo.UpdateStatus(ctx, id, order.Status); err != nil {
		return nil, err
	}
	return order, nil
}
