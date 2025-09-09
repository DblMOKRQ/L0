package messagebroker

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// ====== Тесты ======

// mockReader реализует интерфейс Reader
type mockReader struct {
	message *kafka.Message
	err     error
}

func (m *mockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	if m.err != nil {
		return kafka.Message{}, m.err
	}
	return *m.message, nil
}

func (m *mockReader) Close() error {
	return nil
}

func newTestConsumer(msg *kafka.Message, err error) *Consumer {
	logger := zap.NewExample()
	return &Consumer{
		reader: &mockReader{message: msg, err: err},
		log:    logger.Named("consumer"),
	}
}

func TestConsumer_ReadMessage_Success(t *testing.T) {
	expectedMsg := &kafka.Message{Value: []byte("hello world")}
	consumer := newTestConsumer(expectedMsg, nil)

	msg, err := consumer.ReadMessage(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedMsg.Value, msg.Value)
}

func TestConsumer_ReadMessage_Error(t *testing.T) {
	consumer := newTestConsumer(nil, errors.New("read error"))

	msg, err := consumer.ReadMessage(context.Background())

	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
}

func TestConsumer_Close(t *testing.T) {
	expectedMsg := &kafka.Message{Value: []byte("test")}
	consumer := newTestConsumer(expectedMsg, nil)

	err := consumer.Close()

	assert.NoError(t, err)
}
