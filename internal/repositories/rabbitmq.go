// internal/repositories/rabbitmq_publisher.go
package repositories

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
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     *config.RabbitMQConfig
}

func NewRabbitMQPublisher(
	cfg *config.RabbitMQConfig,
) (*RabbitMQPublisher, error) {
	if cfg == nil {
		cfg = config.NewRabbitMQConfig()
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	err = ch.Qos(
		1,     // prefetch_count: 1 сообщение за раз
		0,     // prefetch_size: 0 = без ограничения по байтам
		false, // global: false = применяется к каналу, не ко всем консьюмерам
	)

	_, err = ch.QueueDeclare(
		cfg.QueueName, // name
		true,          // durable (сохраняется при рестарте)
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)

	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQPublisher{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

func (p *RabbitMQPublisher) PublishTask(
	ctx context.Context,
	task *models.FileTask,
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
			DeliveryMode: amqp.Persistent, // сообщение сохраняется на диск
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

func (p *RabbitMQPublisher) Close() error {
	var errs []error

	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			errs = append(
				errs,
				fmt.Errorf("failed to close connection: %w", err),
			)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}
