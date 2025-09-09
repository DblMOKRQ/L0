package redisClient

import (
	"L0/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

// Мок Redis
type mockRedis struct {
	data   map[string]string
	setErr error
	getErr error
}

func (m *mockRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	if m.setErr != nil {
		return redis.NewStatusResult("", m.setErr)
	}

	// безопасная конвертация любых типов в строку
	switch v := value.(type) {
	case string:
		m.data[key] = v
	case []byte:
		m.data[key] = string(v)
	default:
		m.data[key] = fmt.Sprintf("%v", v)
	}

	return redis.NewStatusResult("OK", nil)
}

func (m *mockRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	if m.getErr != nil {
		return redis.NewStringResult("", m.getErr)
	}
	val, ok := m.data[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(val, nil)
}

func (m *mockRedis) Close() error {
	return nil
}

// helper для создания RedisClient с моками
func newTestRedisClient() *RedisClient {
	logger := zap.NewExample()
	mock := &mockRedis{data: make(map[string]string)}
	return &RedisClient{
		client: mock,
		log:    logger.Named("redis_client"),
	}
}

func TestRedisClient_SetOrder_Success(t *testing.T) {
	rc := newTestRedisClient()
	order := &models.Order{OrderUID: "123"}
	key := "order:123"

	err := rc.SetOrder(context.Background(), order, time.Minute, key)

	assert.NoError(t, err)
	mock := rc.client.(*mockRedis)
	data, _ := json.Marshal(order)
	assert.Equal(t, string(data), mock.data[key])
}

func TestRedisClient_GetOrder_Success(t *testing.T) {
	rc := newTestRedisClient()
	order := &models.Order{OrderUID: "123"}
	key := "order:123"

	data, _ := json.Marshal(order)
	rc.client.(*mockRedis).data[key] = string(data)

	gotOrder, err := rc.GetOrder(context.Background(), order.OrderUID, key)

	assert.NoError(t, err)
	assert.Equal(t, order.OrderUID, gotOrder.OrderUID)
	//assert.Equal(t, order.Name, gotOrder.Name)
}

func TestRedisClient_GetOrder_NotFound(t *testing.T) {
	rc := newTestRedisClient()

	gotOrder, err := rc.GetOrder(context.Background(), "123", "missingKey")

	assert.Nil(t, gotOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get order")
}

func TestRedisClient_Close(t *testing.T) {
	rc := newTestRedisClient()
	rc.Close()
}
