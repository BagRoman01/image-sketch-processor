package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	*RabbitMQBase
}

func NewRabbitMQConsumer(
	ctx context.Context,
	cfg *config.RabbitMQConfig,
) (*RabbitMQConsumer, error) {
	base, err := NewRabbitMQBase(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &RabbitMQConsumer{RabbitMQBase: base}, nil
}

func (c *RabbitMQConsumer) Consume(
	ctx context.Context,
	handler func(ctx context.Context, task *models.S3FileTask) error,
) error {
	msgs, err := c.channel.Consume(
		c.cfg.QueueName,
		"",    // consumer tag
		false, // no auto-ack (manual ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("start consuming: %w", err)
	}

	logger := logging.LoggerFromContext(ctx)
	logger.Info("started consuming messages", "queue", c.cfg.QueueName)

	for {
		select {
		case <-ctx.Done():
			logger.Info("consumer stopped by context")
			return nil
		case delivery, ok := <-msgs:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}

			go func(delivery amqp.Delivery) {
				task, err := c.unmarshalTask(delivery.Body)
				if err != nil {
					slog.Error("failed to unmarshal task", "error", err)
					if err := delivery.Nack(false, false); err != nil {
						slog.Error("failed to Nack message", "error", err)
					}
					return
				}

				taskCtx := context.TODO()
				if err := handler(taskCtx, task); err != nil {
					slog.Error("task handler failed",
						"task_id", task.ID,
						"error", err)
					if nackErr := delivery.Nack(false, false); nackErr != nil {
						slog.Error("failed to Nack message", "error", nackErr)
					}
					return
				}

				// Ack только после успешной обработки
				if err := delivery.Ack(false); err != nil {
					slog.Error("failed to Ack message",
						"task_id", task.ID,
						"error", err)
				}
			}(delivery)
		}
	}
}

func (c *RabbitMQConsumer) unmarshalTask(
	body []byte,
) (*models.S3FileTask, error) {
	var task models.S3FileTask
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, fmt.Errorf("unmarshal task: %w", err)
	}
	return &task, nil
}
