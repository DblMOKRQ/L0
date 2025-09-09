package service

import (
	"L0/internal/models"
	"L0/pkg/validator"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"time"
)

type Consumer interface {
	ReadMessage(ctx context.Context) (*kafka.Message, error)
	Close() error
}

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *models.Order) error
	GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error)
	GetRecentOrders(ctx context.Context, limit int) ([]*models.Order, error)
	Close()
}

type RedisClient interface {
	SetOrder(ctx context.Context, order *models.Order, ttl time.Duration, key string) error
	GetOrder(ctx context.Context, orderUID string, key string) (*models.Order, error)
	Close()
}

type OrderService struct {
	consumer    Consumer
	repository  OrderRepository
	redisClient RedisClient
	ttl         time.Duration
	log         *zap.Logger
}

func NewOrderService(consumer Consumer, repository OrderRepository, redisClient RedisClient, ttl time.Duration, log *zap.Logger) *OrderService {
	return &OrderService{consumer: consumer, repository: repository, redisClient: redisClient, ttl: ttl, log: log.Named("OrderService")}
}

func (s *OrderService) SaveOrder(ctx context.Context, order *models.Order) error {
	return s.repository.SaveOrder(ctx, order)
}

func (s *OrderService) GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error) {
	key := fmt.Sprintf("order:%s", orderUID)
	order, err := s.redisClient.GetOrder(ctx, orderUID, key)
	if err == nil {
		return order, nil
	} else {
		if errors.Is(err, redis.Nil) {
			s.log.Warn("Order not found in redis", zap.String("key", key))
		} else {
			s.log.Error("Error getting order in redis", zap.Error(err))
		}
	}
	order, err = s.repository.GetOrderByUID(ctx, orderUID)
	if err != nil {
		if errors.Is(err, models.OrderNotFoundError) {
			s.log.Warn("Order not found", zap.String("order_uid", orderUID))
			return nil, models.OrderNotFoundError
		}
		s.log.Error("Error getting order in postgres", zap.Error(err))
		return nil, fmt.Errorf("error getting order in postgres: %w", err)
	}
	err = s.SetOrder(ctx, order)
	if err != nil {
		s.log.Error("Error setting order in redis", zap.Error(err))
	}
	return order, nil
}

func (s *OrderService) PreloadRecentOrder(ctx context.Context, limit int) error {
	orders, err := s.repository.GetRecentOrders(ctx, limit)
	if err != nil {
		s.log.Error("Error getting recent orders", zap.Error(err))
		return fmt.Errorf("error getting recent orders: %w", err)
	}
	for _, order := range orders {
		if err := s.SetOrder(ctx, order); err != nil {
			s.log.Error("Error setting order in postgres", zap.Error(err))
		} else {
			s.log.Debug("Preloaded order to cache", zap.String("orderUID", order.OrderUID))
		}
	}
	return nil
}

func (s *OrderService) ReadMessage(ctx context.Context) (*models.Order, error) {

	msg, err := s.consumer.ReadMessage(ctx)
	if err != nil {
		s.log.Error("Error reading message", zap.Error(err))
		return nil, fmt.Errorf("error reading message: %w", err)
	}
	var order models.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		s.log.Error("Error unmarshalling message", zap.Error(err))
		return nil, fmt.Errorf("error unmarshalling message: %w", err)
	}

	if err := validator.ValidateOrder(&order); err != nil {
		s.log.Error("Error validating order", zap.Error(err))
		return nil, fmt.Errorf("error validating order: %w", err)
	}
	return &order, nil
}

func (s *OrderService) SetOrder(ctx context.Context, order *models.Order) error {
	key := fmt.Sprintf("order:%s", order.OrderUID)
	return s.redisClient.SetOrder(ctx, order, s.ttl, key)
}

func (s *OrderService) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	key := fmt.Sprintf("order:%s", orderUID)
	return s.redisClient.GetOrder(ctx, orderUID, key)
}

func (s *OrderService) CloseRedisClient() {
	s.redisClient.Close()
}
func (s *OrderService) CloseConsumer() error {
	return s.consumer.Close()
}
func (s *OrderService) CloseRepo() {
	s.repository.Close()
}
