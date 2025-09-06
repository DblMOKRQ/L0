package messagebroker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	log    *zap.Logger
}

func NewProducer(broker []string, topic string, log *zap.Logger) *Producer {
	writer := kafka.Writer{
		Addr:  kafka.TCP(broker...),
		Topic: topic,
	}
	return &Producer{writer: &writer, log: log.Named("producer")}
}

func (p *Producer) SendMessage(ctx context.Context, key string, value interface{}) error {
	p.log.Debug("Producer send message", zap.Any("value", value))
	jsonValue, err := json.Marshal(value)
	if err != nil {
		p.log.Error("Failed to marshal value", zap.Any("value", value), zap.Error(err))
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	message := &kafka.Message{
		Key:   []byte(key),
		Value: jsonValue,
	}
	if err := p.writer.WriteMessages(ctx, *message); err != nil {
		p.log.Error("Failed to write message", zap.Any("message", message), zap.Error(err))
		return fmt.Errorf("failed to write message: %w", err)
	}
	p.log.Debug("Producer sent message", zap.Any("message", message))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
