package handlers

import (
	"L0/internal/models"
	"L0/internal/service"
	"errors"
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

// GetOrder godoc
// @Summary Get an order by UID
// @Description Retrieves order details by its UID
// @Tags orders
// @Accept json
// @Produce json
// @Param orderUID path string true "Order UID"
// @Success 200 {object} models.Order
// @Failure 404 {object} nil "Not Found"
// @Failure 500 {object} nil "Internal Error"
// @Router /orders/{orderUID} [get]
func (h *OrderHandlers) GetOrder(c *gin.Context) {
	log := c.Value("logger").(*zap.Logger)
	log.Info("Handling getting order")

	orderID := c.Param("orderUID")

	order, err := h.orderService.GetOrderByUID(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, models.OrderNotFoundError) {
			log.Warn("Order not found", zap.String("order_uid", orderID))
			c.Status(http.StatusNotFound)
			return
		}
		log.Error("Error getting order", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, order)
}
