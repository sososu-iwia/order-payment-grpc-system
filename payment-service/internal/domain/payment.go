package domain

import "errors"

const (
	PaymentStatusAuthorized = "Authorized"
	PaymentStatusDeclined   = "Declined"
)

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
}

func NewPayment(id, orderID, transactionID string, amount int64) (*Payment, error) {
	if orderID == "" {
		return nil, errors.New("order_id is required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	status := PaymentStatusAuthorized
	if amount > 100000 {
		status = PaymentStatusDeclined
	}

	return &Payment{ID: id, OrderID: orderID, TransactionID: transactionID, Amount: amount, Status: status}, nil
}
