package redisClient

import (
	"L0/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type RedisClient struct {
	client *redis.Client
	log    *zap.Logger
}

func NewRedisClient(ctx context.Context, addr string, password string, db int, log *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	log = log.Named("redis_client")
	if err := client.Ping(ctx).Err(); err != nil {
		log.Error("failed to connect to redis_client", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to redis_client: %w", err)
	}
	return &RedisClient{client: client, log: log}, nil
}

func (rc *RedisClient) SetOrder(ctx context.Context, order *models.Order, ttl time.Duration, key string) error {
	rc.log.Debug("setting order", zap.String("key", key), zap.Duration("ttl", ttl))
	jsonOrder, err := json.Marshal(order)
	if err != nil {
		rc.log.Error("failed to marshal order", zap.Error(err))
		return fmt.Errorf("failed to marshal order: %w", err)
	}
	return rc.client.Set(ctx, key, jsonOrder, ttl).Err()
}

func (rc *RedisClient) GetOrder(ctx context.Context, uid string, key string) (*models.Order, error) {
	data, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		rc.log.Error("failed to get order", zap.String("uid", uid), zap.Error(err))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	var order models.Order
	err = json.Unmarshal(data, &order)
	if err != nil {
		rc.log.Error("failed to unmarshal order", zap.String("uid", uid), zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}
	return &order, nil
}

func (rc *RedisClient) Close() {
	if rc.client != nil {
		rc.client.Close()
	}
}
