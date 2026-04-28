package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"order-service/internal/domain"
)

type OrderUsecase interface {
	CreateOrder(ctx *gin.Context, customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, int, error)
	GetOrder(ctx *gin.Context, id string) (*domain.Order, error)
	CancelOrder(ctx *gin.Context, id string) (*domain.Order, error)
}

type Handler struct{ uc OrderUsecase }

func NewHandler(uc OrderUsecase) *Handler { return &Handler{uc: uc} }

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/:id", h.GetOrder)
	r.PATCH("/orders/:id/cancel", h.CancelOrder)
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}
	order, status, err := h.uc.CreateOrder(c, req.CustomerID, req.ItemName, req.Amount, c.GetHeader("Idempotency-Key"))
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error(), "order": toOrderResponse(order)})
		return
	}
	c.JSON(status, toOrderResponse(order))
}

func (h *Handler) GetOrder(c *gin.Context) {
	order, err := h.uc.GetOrder(c, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}

func (h *Handler) CancelOrder(c *gin.Context) {
	order, err := h.uc.CancelOrder(c, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}

func toOrderResponse(order *domain.Order) OrderResponse {
	if order == nil {
		return OrderResponse{}
	}
	return OrderResponse{
		ID: order.ID, CustomerID: order.CustomerID, ItemName: order.ItemName,
		Amount: order.Amount, Status: order.Status, CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}
}
