package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	*RabbitMQBase
}

func NewRabbitMQPublisher(
	ctx context.Context,
	cfg *config.RabbitMQConfig,
) (*RabbitMQPublisher, error) {
	base, err := NewRabbitMQBase(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &RabbitMQPublisher{RabbitMQBase: base}, nil
}

func (p *RabbitMQPublisher) PublishTask(
	ctx context.Context,
	task *models.S3FileTask,
) error {
	logger := logging.LoggerFromContext(ctx)

	body, err := json.Marshal(task)
	if err != nil {
		logger.Error(
			"failed to marshal task",
			"error", err,
			"task_id", task.ID,
		)
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		"",              // exchange
		p.cfg.QueueName, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		logger.Error("failed to publish task to RabbitMQ",
			"error", err,
			"task_id", task.ID,
			"queue", p.cfg.QueueName,
		)
		return fmt.Errorf("failed to publish task: %w", err)
	}

	logger.Info("task published to RabbitMQ",
		"task_id", task.ID,
		"queue", p.cfg.QueueName,
	)
	return nil
}

