package app

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"payment-service/internal/domain"
	repo "payment-service/internal/repository/postgres"
	grpctransport "payment-service/internal/transport/grpc"
	httptransport "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
)

type uuidGenerator struct{}

func (uuidGenerator) NewID() string { return uuid.NewString() }

// httpUsecaseAdapter bridges the gin-aware HTTP interface to the pure usecase.
type httpUsecaseAdapter struct{ uc *usecase.PaymentUsecase }

func (a *httpUsecaseAdapter) CreatePayment(ctx *gin.Context, orderID string, amount int64) (*domain.Payment, error) {
	return a.uc.CreatePayment(ctx.Request.Context(), orderID, amount)
}
func (a *httpUsecaseAdapter) GetByOrderID(ctx *gin.Context, orderID string) (*domain.Payment, error) {
	return a.uc.GetByOrderID(ctx.Request.Context(), orderID)
}

type RouterDeps struct{ DB *sql.DB }

// NewRouter wires the REST layer. Business logic (UseCase) is unchanged.
func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.Default()
	paymentRepo := repo.NewPaymentRepository(deps.DB)
	uc := usecase.NewPaymentUsecase(paymentRepo, uuidGenerator{})
	handler := httptransport.NewHandler(&httpUsecaseAdapter{uc: uc})
	handler.RegisterRoutes(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	return r
}

// NewGRPCServer wires the gRPC layer using generated contracts from
// github.com/sososu-iwia/grpc-contracts-artifacts/payment.
// The UseCase layer is reused without modification (Clean Architecture preserved).
func NewGRPCServer(deps RouterDeps) *grpctransport.Server {
	paymentRepo := repo.NewPaymentRepository(deps.DB)
	uc := usecase.NewPaymentUsecase(paymentRepo, uuidGenerator{})
	srv := grpctransport.NewServer(uc)
	log.Println("[app] gRPC payment server created (using grpc-contracts-artifacts)")
	return srv
}
