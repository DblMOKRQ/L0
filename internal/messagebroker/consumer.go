package messagebroker

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader *kafka.Reader
	log    *zap.Logger
}

func NewConsumer(brokers []string, topic string, log *zap.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	return &Consumer{reader: reader, log: log.Named("consumer")}
}

func (c *Consumer) ReadMessage(ctx context.Context) (*kafka.Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		c.log.Error("Error reading message", zap.Error(err))
		return nil, fmt.Errorf("error reading message: %w", err)
	}
	c.log.Info("received message", zap.ByteString("message", msg.Value))

	return &msg, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
