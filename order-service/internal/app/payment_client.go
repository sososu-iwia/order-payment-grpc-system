package app

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	contractspb "github.com/sososu-iwia/grpc-contracts-artifacts/payment"

	"order-service/internal/usecase"
)

type PaymentGRPCClient struct {
	conn    *grpc.ClientConn
	client  contractspb.PaymentServiceClient
	timeout time.Duration
}

func NewPaymentGRPCClient(addr string) (*PaymentGRPCClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	log.Printf("[grpc-client] connected to payment-service at %s", addr)
	return &PaymentGRPCClient{
		conn:    conn,
		client:  contractspb.NewPaymentServiceClient(conn),
		timeout: 5 * time.Second,
	}, nil
}

func (c *PaymentGRPCClient) Close() error { return c.conn.Close() }

func (c *PaymentGRPCClient) CreatePayment(ctx context.Context, orderID string, amount int64) (*usecase.PaymentResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.ProcessPayment(ctx, &contractspb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.InvalidArgument:
				return nil, err
			case codes.Unavailable:
				return nil, err
			}
		}
		return nil, err
	}

	return &usecase.PaymentResult{
		Status:        resp.GetStatus(),
		TransactionID: resp.GetTransactionId(),
	}, nil
}
