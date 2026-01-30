// internal/repositories/redis_repository.go
package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
	cfg    *config.RedisConfig
}

func NewRedisRepository(
	cfg *config.RedisConfig,
	ctx context.Context,
) (*RedisRepository, error) {
	slog.Debug("initializing Redis client",
		"addr", cfg.Addr,
		"db", cfg.DB,
		"pool_size", cfg.PoolSize,
	)

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  time.Duration(cfg.DialTimeoutSec) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSec) * time.Second,
		PoolSize:     cfg.PoolSize,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("failed to connect to Redis",
			"error", err,
			"addr", cfg.Addr,
		)
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	slog.Info("Redis client connected successfully",
		"addr", cfg.Addr,
		"db", cfg.DB,
	)

	return &RedisRepository{
		client: client,
		cfg:    cfg,
	}, nil
}

func (r *RedisRepository) SaveTask(
	ctx context.Context,
	task *models.FileTask,
) error {
	logger := logging.LoggerFromContext(ctx)

	key := fmt.Sprintf("task:%s", task.ID)
	data, err := json.Marshal(task)
	if err != nil {
		logger.Error("failed to marshal task", "error", err, "task_id", task.ID)
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// TTL 24 часа для задач
	if err := r.client.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		logger.Error("failed to save task to Redis",
			"error", err,
			"task_id", task.ID,
			"key", key,
		)
		return fmt.Errorf("failed to save task: %w", err)
	}

	logger.Debug("task saved to Redis", "task_id", task.ID, "ttl", "24h")
	return nil
}

func (r *RedisRepository) GetTask(
	ctx context.Context,
	taskID string,
) (*models.FileTask, error) {
	logger := logging.LoggerFromContext(ctx)

	key := fmt.Sprintf("task:%s", taskID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		logger.Warn("task not found in Redis", "task_id", taskID)
		return nil, fmt.Errorf("task not found")
	} else if err != nil {
		logger.Error("failed to get task from Redis",
			"error", err,
			"task_id", taskID,
		)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var task models.FileTask
	if err := json.Unmarshal(data, &task); err != nil {
		logger.Error("failed to unmarshal task",
			"error", err,
			"task_id", taskID,
		)
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	logger.Debug("task retrieved from Redis", "task_id", taskID)
	return &task, nil
}

func (r *RedisRepository) UpdateTaskStatus(
	ctx context.Context,
	taskID string,
	status models.TaskStatus,
	errorMsg string,
) error {
	logger := logging.LoggerFromContext(ctx)

	task, err := r.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	oldStatus := task.Status
	task.Status = status
	task.UpdatedAt = time.Now()
	if errorMsg != "" {
		task.Error = errorMsg
	}

	if err := r.SaveTask(ctx, task); err != nil {
		return err
	}

	logger.Info("task status updated",
		"task_id", taskID,
		"old_status", oldStatus,
		"new_status", status,
	)

	return nil
}

func (r *RedisRepository) Close() error {
	slog.Info("closing Redis connection", "addr", r.cfg.Addr)
	return r.client.Close()
}
