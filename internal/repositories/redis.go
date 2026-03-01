package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
	cfg    *config.RedisConfig
}

func NewRedisRepository(
	ctx context.Context,
	cfg *config.RedisConfig,
) (*RedisRepository, error) {
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
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisRepository{
		client: client,
		cfg:    cfg,
	}, nil
}

func (r *RedisRepository) SaveTask(
	ctx context.Context,
	task *models.S3FileTask,
) error {
	key := fmt.Sprintf("task:%s", task.ID)
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := r.client.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	return nil
}

func (r *RedisRepository) GetTask(
	ctx context.Context,
	taskID string,
) (*models.S3FileTask, error) {
	key := fmt.Sprintf("task:%s", taskID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("task %s not found in Redis", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("get task %s from Redis: %w", taskID, err)
	}

	var task models.S3FileTask
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("unmarshal task %s: %w", taskID, err)
	}

	return &task, nil
}

func (r *RedisRepository) UpdateTask(
	ctx context.Context,
	taskID string,
	updateFunc func(*models.S3FileTask) error,
) error {
	currentTask, err := r.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task %s not found, nothing to update", taskID)
	}

	if err := updateFunc(currentTask); err != nil {
		return fmt.Errorf("update task %s: %w", taskID, err)
	}

	if err := r.SaveTask(ctx, currentTask); err != nil {
		return fmt.Errorf("failed to update task %s: %w", taskID, err)
	}

	return nil
}

func (r *RedisRepository) Close(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		done <- r.client.Close()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
