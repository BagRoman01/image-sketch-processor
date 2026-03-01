package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQBase struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     *config.RabbitMQConfig
}

func NewRabbitMQBase(
	ctx context.Context,
	cfg *config.RabbitMQConfig,
) (*RabbitMQBase, error) {
	slog.Debug("init RabbitMQ base", "queue", cfg.QueueName)

	if cfg == nil {
		cfg = config.NewRabbitMQConfig()
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		slog.Error("RabbitMQ dial failed", "error", err, "url", cfg.URL)
		return nil, fmt.Errorf("RabbitMQ dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		slog.Error("RabbitMQ channel failed", "error", err)
		return nil, fmt.Errorf("RabbitMQ channel: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		slog.Error("RabbitMQ QoS failed", "error", err)
		return nil, fmt.Errorf("RabbitMQ QoS: %w", err)
	}

	if _, err := ch.QueueDeclare(
		cfg.QueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		ch.Close()
		conn.Close()
		slog.Error(
			"RabbitMQ queue declare failed",
			"queue", cfg.QueueName,
			"error", err,
		)
		return nil, fmt.Errorf("RabbitMQ queue: %w", err)
	}

	slog.Info("RabbitMQ base ready", "queue", cfg.QueueName)

	return &RabbitMQBase{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

func (p *RabbitMQBase) Close(ctx context.Context) error {
	logger := logging.LoggerFromContext(ctx)
	logger.Info("closing RabbitMQ", "queue", p.cfg.QueueName)

	done := make(chan error, 1)

	go func() {
		var errs []error

		if p.channel != nil && !p.channel.IsClosed() {
			if err := p.channel.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close channel: %w", err))
			}
		}

		if p.conn != nil && !p.conn.IsClosed() {
			if err := p.conn.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close connection: %w", err))
			}
		}

		if len(errs) > 0 {
			done <- fmt.Errorf("rabbitmq close errors: %v", errs)
			return
		}

		done <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
