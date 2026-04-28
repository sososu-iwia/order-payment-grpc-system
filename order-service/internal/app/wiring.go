package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"order-service/internal/domain"
	repo "order-service/internal/repository/postgres"
	grpctransport "order-service/internal/transport/grpc"
	httptransport "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

type Config struct {
	// PaymentGRPCAddr is read from env PAYMENT_GRPC_ADDR (never hardcoded).
	PaymentGRPCAddr string
	GRPCPort        string
}

type RouterDeps struct {
	DB     *sql.DB
	Config Config
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type uuidGenerator struct{}

func (uuidGenerator) NewID() string { return uuid.NewString() }

// orderUsecaseAdapter bridges gin.Context to the pure domain usecase.
type orderUsecaseAdapter struct{ uc *usecase.OrderUsecase }

func (a *orderUsecaseAdapter) CreateOrder(ctx *gin.Context, customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, int, error) {
	return a.uc.CreateOrder(ctx.Request.Context(), customerID, itemName, amount, idempotencyKey)
}
func (a *orderUsecaseAdapter) GetOrder(ctx *gin.Context, id string) (*domain.Order, error) {
	return a.uc.GetOrder(ctx.Request.Context(), id)
}
func (a *orderUsecaseAdapter) CancelOrder(ctx *gin.Context, id string) (*domain.Order, error) {
	return a.uc.CancelOrder(ctx.Request.Context(), id)
}

// NewRouter wires REST layer. Order Service calls Payment Service via gRPC
// using the generated PaymentServiceClient from grpc-contracts-artifacts.
func NewRouter(deps RouterDeps) (*gin.Engine, func(), error) {
	paymentClient, err := NewPaymentGRPCClient(deps.Config.PaymentGRPCAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("dial payment grpc: %w", err)
	}
	cleanup := func() {
		if err := paymentClient.Close(); err != nil {
			log.Printf("[app] grpc client close: %v", err)
		}
	}

	orderRepo := repo.NewOrderRepository(deps.DB)
	uc := usecase.NewOrderUsecase(orderRepo, paymentClient, realClock{}, uuidGenerator{})
	handler := httptransport.NewHandler(&orderUsecaseAdapter{uc: uc})

	r := gin.Default()
	handler.RegisterRoutes(r)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r, cleanup, nil
}

// NewGRPCServer creates the streaming gRPC server using generated
// OrderServiceServer interface from grpc-contracts-artifacts.
func NewGRPCServer(db *sql.DB) *grpctransport.Server {
	srv := grpctransport.NewServer(db)
	log.Println("[app] gRPC order streaming server created (using grpc-contracts-artifacts)")
	return srv
}
