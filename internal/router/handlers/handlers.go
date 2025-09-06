package handlers

import (
	"L0/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type OrderHandlers struct {
	orderService *service.OrderService
}

func NewOrderHandlers(orderService *service.OrderService) *OrderHandlers {
	return &OrderHandlers{orderService: orderService}
}

func (h *OrderHandlers) GetOrder(c *gin.Context) {
	log := c.Value("logger").(*zap.Logger)
	log.Info("Handling getting order")

	orderID := c.Param("orderUID")

	order, err := h.orderService.GetOrderByUID(c.Request.Context(), orderID)
	if err != nil {
		log.Error("Error getting order", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, order)
}
