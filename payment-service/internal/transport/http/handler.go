package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-service/internal/domain"
)

type PaymentUsecase interface {
	CreatePayment(ctx *gin.Context, orderID string, amount int64) (*domain.Payment, error)
	GetByOrderID(ctx *gin.Context, orderID string) (*domain.Payment, error)
}

type Handler struct{ uc PaymentUsecase }

func NewHandler(uc PaymentUsecase) *Handler { return &Handler{uc: uc} }

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/payments", h.CreatePayment)
	r.GET("/payments/:orderID", h.GetByOrderID)
}

func (h *Handler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}
	payment, err := h.uc.CreatePayment(c, req.OrderID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toPaymentResponse(payment))
}

func (h *Handler) GetByOrderID(c *gin.Context) {
	payment, err := h.uc.GetByOrderID(c, c.Param("orderID"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, toPaymentResponse(payment))
}

func toPaymentResponse(payment *domain.Payment) PaymentResponse {
	return PaymentResponse{
		ID: payment.ID, OrderID: payment.OrderID, TransactionID: payment.TransactionID,
		Amount: payment.Amount, Status: payment.Status,
	}
}
