// internal/services/task_service.go
package services

import (
	"context"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	"github.com/BagRoman01/image-sketch-processor/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type TaskService struct {
	redisRepo    *repositories.RedisRepository
	rabbitMQRepo *repositories.RabbitMQPublisher
}

func NewTaskService(
	redisRepo *repositories.RedisRepository,
	rabbitMQRepo *repositories.RabbitMQPublisher,
) *TaskService {
	return &TaskService{
		redisRepo:    redisRepo,
		rabbitMQRepo: rabbitMQRepo,
	}
}

func (s *TaskService) CreateFileProcessingTask(
	ctx context.Context,
	fileKey, fileName string,
	fileSize int64,
	contentType string,
) (*models.FileTask, error) {
	logger := logging.LoggerFromContext(ctx)

	taskID := ulid.Make().String()

	task := &models.FileTask{
		ID:          taskID,
		FileKey:     fileKey,
		FileName:    fileName,
		FileSize:    fileSize,
		ContentType: contentType,
		Status:      models.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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

	if err := s.rabbitMQRepo.PublishTask(ctx, task); err != nil {
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
		"file_key", fileKey,
	)

	return task, nil
}

func (s *TaskService) GetTaskStatus(
	ctx context.Context,
	taskID string,
) (*models.FileTask, error) {
	return s.redisRepo.GetTask(ctx, taskID)
}
