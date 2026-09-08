package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/railzwaylabs/github-actions-samples/internal/order/application"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("order.http", fx.Invoke(Register))

type Handler struct {
	orders application.Service
	log    *zap.Logger
}

func Register(router *gin.Engine, orders application.Service, log *zap.Logger) {
	handler := &Handler{orders: orders, log: log}
	router.GET("/orders", handler.List)
	router.POST("/orders", handler.Create)
	router.GET("/orders/:id", handler.Get)
}

func (h *Handler) List(c *gin.Context) {
	orders, err := h.orders.List(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *Handler) Get(c *gin.Context) {
	order, err := h.orders.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, application.ErrOrderNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) Create(c *gin.Context) {
	var request struct {
		CustomerID string `json:"customer_id" binding:"required"`
		Items      []struct {
			ProductID string `json:"product_id" binding:"required"`
			Quantity  int64  `json:"quantity" binding:"required,gt=0"`
			Price     int64  `json:"price" binding:"gte=0"`
		} `json:"items" binding:"required,min=1,dive"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order", "details": err.Error()})
		return
	}
	input := application.CreateOrderInput{CustomerID: request.CustomerID, Items: make([]application.CreateOrderItemInput, 0, len(request.Items))}
	for _, item := range request.Items {
		input.Items = append(input.Items, application.CreateOrderItemInput{ProductID: item.ProductID, Quantity: item.Quantity, Price: item.Price})
	}
	order, err := h.orders.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *Handler) internalError(c *gin.Context, err error) {
	h.log.Error("order request failed", zap.Error(err), zap.String("path", c.Request.URL.Path))
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
