package service_test

import (
	"L0/internal/models"
	"L0/internal/service"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// --- Mocks ---

type MockConsumer struct {
	mock.Mock
}

func (m *MockConsumer) ReadMessage(ctx context.Context) (*kafka.Message, error) {
	args := m.Called(ctx)
	if msg, ok := args.Get(0).(*kafka.Message); ok {
		return msg, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockConsumer) Close() error {
	return m.Called().Error(0)
}

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) SaveOrder(ctx context.Context, order *models.Order) error {
	return m.Called(ctx, order).Error(0)
}
func (m *MockRepo) GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error) {
	args := m.Called(ctx, orderUID)
	if order, ok := args.Get(0).(*models.Order); ok {
		return order, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) GetRecentOrders(ctx context.Context, limit int) ([]*models.Order, error) {
	args := m.Called(ctx, limit)
	if orders, ok := args.Get(0).([]*models.Order); ok {
		return orders, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) Close() { m.Called() }

type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) SetOrder(ctx context.Context, order *models.Order, ttl time.Duration, key string) error {
	return m.Called(ctx, order, ttl, key).Error(0)
}
func (m *MockRedis) GetOrder(ctx context.Context, orderUID string, key string) (*models.Order, error) {
	args := m.Called(ctx, orderUID, key)
	if order, ok := args.Get(0).(*models.Order); ok {
		return order, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRedis) Close() { m.Called() }

// --- Tests ---

func TestGetOrderByUID_FromRedis(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	redisClient := new(MockRedis)
	consumer := new(MockConsumer)
	log := zap.NewNop()

	expectedOrder := &models.Order{OrderUID: "123"}

	redisClient.On("GetOrder", ctx, "123", "order:123").
		Return(expectedOrder, nil)

	svc := service.NewOrderService(consumer, repo, redisClient, time.Minute, log)

	order, err := svc.GetOrderByUID(ctx, "123")

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, order)
	redisClient.AssertExpectations(t)
}

func TestGetOrderByUID_FromRepoWhenRedisNil(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	redisClient := new(MockRedis)
	consumer := new(MockConsumer)
	log := zap.NewNop()

	expectedOrder := &models.Order{OrderUID: "456"}

	redisClient.On("GetOrder", ctx, "456", "order:456").
		Return(nil, redis.Nil)

	repo.On("GetOrderByUID", ctx, "456").
		Return(expectedOrder, nil)

	redisClient.On("SetOrder", ctx, expectedOrder, mock.Anything, "order:456").
		Return(nil)

	svc := service.NewOrderService(consumer, repo, redisClient, time.Minute, log)

	order, err := svc.GetOrderByUID(ctx, "456")

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, order)
	repo.AssertExpectations(t)
	redisClient.AssertExpectations(t)
}

func TestReadMessage_ValidOrder(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	redisClient := new(MockRedis)
	consumer := new(MockConsumer)
	log := zap.NewNop()

	order := &models.Order{OrderUID: "563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  "563feb7b2b84b6test",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Date(2021, 11, 26, 6, 22, 19, 0, time.UTC),
		OofShard:          "1"}
	data, _ := json.Marshal(order)

	consumer.On("ReadMessage", ctx).
		Return(&kafka.Message{Value: data}, nil)

	svc := service.NewOrderService(consumer, repo, redisClient, time.Minute, log)

	got, err := svc.ReadMessage(ctx)

	assert.NoError(t, err)
	assert.Equal(t, order.OrderUID, got.OrderUID)
	consumer.AssertExpectations(t)
}
func TestPreloadRecentOrder_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	redisClient := new(MockRedis)
	consumer := new(MockConsumer)
	log := zap.NewNop()

	orders := []*models.Order{
		{OrderUID: "111"},
		{OrderUID: "222"},
	}

	repo.On("GetRecentOrders", ctx, 2).
		Return(orders, nil)

	redisClient.On("SetOrder", ctx, orders[0], mock.Anything, "order:111").
		Return(nil)
	redisClient.On("SetOrder", ctx, orders[1], mock.Anything, "order:222").
		Return(nil)

	svc := service.NewOrderService(consumer, repo, redisClient, time.Minute, log)

	err := svc.PreloadRecentOrder(ctx, 2)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	redisClient.AssertExpectations(t)
}

func TestPreloadRecentOrder_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	redisClient := new(MockRedis)
	consumer := new(MockConsumer)
	log := zap.NewNop()

	repo.On("GetRecentOrders", ctx, 5).
		Return(nil, errors.New("db error"))

	svc := service.NewOrderService(consumer, repo, redisClient, time.Minute, log)

	err := svc.PreloadRecentOrder(ctx, 5)

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestSetOrder_CallsRedis(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	redisClient := new(MockRedis)
	consumer := new(MockConsumer)
	log := zap.NewNop()

	order := &models.Order{OrderUID: "999"}

	redisClient.On("SetOrder", ctx, order, time.Minute, "order:999").
		Return(nil)

	svc := service.NewOrderService(consumer, repo, redisClient, time.Minute, log)

	err := svc.SetOrder(ctx, order)

	assert.NoError(t, err)
	redisClient.AssertExpectations(t)
}
