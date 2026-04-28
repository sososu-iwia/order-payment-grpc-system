package grpc

import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Import generated server interface from shared contracts repo (Contract-First)
	contractspb "github.com/sososu-iwia/grpc-contracts-artifacts/order"
)

type StreamingHandler struct {
	contractspb.UnimplementedOrderServiceServer
	db *sql.DB
}

func NewStreamingHandler(db *sql.DB) *StreamingHandler {
	return &StreamingHandler{db: db}
}

func (h *StreamingHandler) SubscribeToOrderUpdates(
	req *contractspb.OrderRequest,
	stream grpc.ServerStreamingServer[contractspb.OrderStatusUpdate],
) error {
	if req.GetOrderId() == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	log.Printf("[grpc-stream] client subscribed to order %s", req.GetOrderId())

	var lastStatus string
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("[grpc-stream] client disconnected from order %s", req.GetOrderId())
			return nil

		case <-ticker.C:
			currentStatus, err := h.fetchOrderStatus(stream.Context(), req.GetOrderId())
			if err != nil {
				return status.Errorf(codes.NotFound, "order %s not found", req.GetOrderId())
			}

			// Only push if the status has actually changed in the DB
			if currentStatus == lastStatus {
				continue
			}
			lastStatus = currentStatus

			update := &contractspb.OrderStatusUpdate{
				OrderId:   req.GetOrderId(),
				Status:    currentStatus,
				UpdatedAt: timestamppb.New(time.Now().UTC()),
			}
			if err := stream.Send(update); err != nil {
				log.Printf("[grpc-stream] send error: %v", err)
				return err
			}
			log.Printf("[grpc-stream] order %s → %s", req.GetOrderId(), currentStatus)

			if isTerminal(currentStatus) {
				log.Printf("[grpc-stream] order %s reached terminal state, closing stream", req.GetOrderId())
				return nil
			}
		}
	}
}

func (h *StreamingHandler) fetchOrderStatus(ctx context.Context, orderID string) (string, error) {
	var s string
	err := h.db.QueryRowContext(ctx,
		`SELECT status FROM orders WHERE id = $1`, orderID,
	).Scan(&s)
	return s, err
}

func isTerminal(s string) bool {
	return s == "Paid" || s == "Failed" || s == "Cancelled"
}

// Server wraps the gRPC server for the Order streaming service.
type Server struct {
	grpcServer *grpc.Server
	handler    *StreamingHandler
}

func NewServer(db *sql.DB) *Server {
	grpcSrv := grpc.NewServer()
	h := NewStreamingHandler(db)
	contractspb.RegisterOrderServiceServer(grpcSrv, h)
	return &Server{grpcServer: grpcSrv, handler: h}
}

func (s *Server) Listen(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("[grpc] order-service streaming on %s", addr)
	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() { s.grpcServer.GracefulStop() }
