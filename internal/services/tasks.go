package services

import (
	"context"
	"fmt"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/messaging/rabbitmq"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	"github.com/BagRoman01/image-sketch-processor/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type TaskService struct {
	redisRepo         *repositories.RedisRepository
	rabbitmqPublisher *rabbitmq.RabbitMQPublisher
}

func NewTaskService(
	redisRepo *repositories.RedisRepository,
	rabbitmqPublisher *rabbitmq.RabbitMQPublisher,
) *TaskService {
	return &TaskService{
		redisRepo:         redisRepo,
		rabbitmqPublisher: rabbitmqPublisher,
	}
}

func (s *TaskService) CreateFileProcessingTask(
	ctx context.Context,
	fileInfo models.S3FileInfo,
) (*models.S3FileTask, error) {
	logger := logging.LoggerFromContext(ctx)

	taskID := ulid.Make().String()

	task := &models.S3FileTask{
		Task: models.Task{
			ID:        taskID,
			Status:    models.TaskStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		S3FileInfo: fileInfo,
	}

	if err := s.redisRepo.SaveTask(ctx, task); err != nil {
		logger.Error(
			"failed to save task to Redis",
			"error", err,
			"task_id", taskID,
		)
		return nil, err
	}

	logger.Debug(
		"task metadata saved to Redis",
		"task_id", taskID,
	)

	if err := s.rabbitmqPublisher.PublishTask(ctx, task); err != nil {
		logger.Error(
			"failed to publish task to RabbitMQ",
			"error", err,
			"task_id", taskID,
		)
		return nil, err
	}

	logger.Info(
		"file processing task created",
		"task_id", taskID,
		"file_key", fileInfo.FileKey,
	)

	return task, nil
}

func (s *TaskService) SetTaskProcessing(
	ctx context.Context,
	taskID string,
) error {
	logger := logging.LoggerFromContext(ctx)

	if err := s.redisRepo.UpdateTask(
		ctx,
		taskID,
		func(task *models.S3FileTask) error {
			task.Status = models.TaskStatusProcessing
			task.UpdatedAt = time.Now()
			return nil
		}); err != nil {
		logger.Error(
			"failed to set task processing",
			"task_id",
			taskID,
			"error",
			err,
		)
		return fmt.Errorf("set task %q to processing: %w", taskID, err)
	}

	return nil
}

func (s *TaskService) SetTaskCompleted(
	ctx context.Context,
	taskID, processedKey, downloadURL string,
) error {
	logger := logging.LoggerFromContext(ctx)

	if err := s.redisRepo.UpdateTask(
		ctx,
		taskID,
		func(task *models.S3FileTask) error {
			task.Status = models.TaskStatusCompleted
			task.ProcessedKey = processedKey
			task.DownloadURL = downloadURL
			task.CompletedAt = time.Now()
			task.UpdatedAt = time.Now()
			return nil
		}); err != nil {
		logger.Error("failed to set task completed",
			"task_id", taskID,
			"processed_key", processedKey,
			"error", err,
		)
		return fmt.Errorf("set task %q completed: %w", taskID, err)
	}

	return nil
}

func (s *TaskService) SetTaskFailed(
	ctx context.Context,
	taskID, errorMsg string,
) error {
	logger := logging.LoggerFromContext(ctx)
	logger.Warn("setting task to failed",
		"task_id", taskID,
		"error", errorMsg,
	)

	if err := s.redisRepo.UpdateTask(
		ctx,
		taskID,
		func(task *models.S3FileTask) error {
			task.Status = models.TaskStatusFailed
			task.Error = errorMsg
			task.UpdatedAt = time.Now()
			return nil
		}); err != nil {
		logger.Error("failed to set task failed",
			"task_id", taskID,
			"error_msg", errorMsg,
			"update_error", err,
		)
		return fmt.Errorf("set task %q failed: %w", taskID, err)
	}

	return nil
}

func (s *TaskService) GetTask(
	ctx context.Context,
	taskID string,
) (*models.S3FileTask, error) {
	return s.redisRepo.GetTask(ctx, taskID)
}
