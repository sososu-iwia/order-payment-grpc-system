package grpc

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	contractspb "github.com/sososu-iwia/grpc-contracts-artifacts/payment"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

type PaymentUsecase interface {
	CreatePayment(ctx context.Context, orderID string, amount int64) (*domain.Payment, error)
	ListPayments(ctx context.Context, minAmount, maxAmount int64) ([]*domain.Payment, error)
}

type Handler struct {
	contractspb.UnimplementedPaymentServiceServer
	uc PaymentUsecase
}

func NewHandler(uc PaymentUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) ProcessPayment(ctx context.Context, req *contractspb.PaymentRequest) (*contractspb.PaymentResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than 0")
	}

	payment, err := h.uc.CreatePayment(ctx, req.GetOrderId(), req.GetAmount())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process payment: %v", err)
	}

	return toPaymentResponse(payment), nil
}

func (h *Handler) ListPayments(ctx context.Context, req *contractspb.ListPaymentsRequest) (*contractspb.ListPaymentsResponse, error) {
	payments, err := h.uc.ListPayments(ctx, req.GetMinAmount(), req.GetMaxAmount())
	if err != nil {
		if err == usecase.ErrInvalidAmountRange {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to list payments: %v", err)
	}

	resp := &contractspb.ListPaymentsResponse{
		Payments: make([]*contractspb.PaymentResponse, 0, len(payments)),
	}
	for _, p := range payments {
		resp.Payments = append(resp.Payments, toPaymentResponse(p))
	}
	return resp, nil
}

func toPaymentResponse(p *domain.Payment) *contractspb.PaymentResponse {
	return &contractspb.PaymentResponse{
		Id:            p.ID,
		OrderId:       p.OrderID,
		TransactionId: p.TransactionID,
		Amount:        p.Amount,
		Status:        p.Status,
		CreatedAt:     timestamppb.New(time.Now().UTC()),
	}
}

// Server wraps the gRPC server with logging interceptor (bonus requirement).
type Server struct {
	grpcServer *grpc.Server
}

func NewServer(uc PaymentUsecase) *Server {
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(LoggingInterceptor),
	)
	h := NewHandler(uc)
	contractspb.RegisterPaymentServiceServer(grpcSrv, h)
	return &Server{grpcServer: grpcSrv}
}

func (s *Server) Listen(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("[grpc] payment-service listening on %s", addr)
	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() { s.grpcServer.GracefulStop() }
