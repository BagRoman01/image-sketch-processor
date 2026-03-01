package injectors

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/messaging/rabbitmq"
	"github.com/BagRoman01/image-sketch-processor/internal/repositories"
	"github.com/BagRoman01/image-sketch-processor/internal/services"
)

type ServiceInjector struct {
	FileService       *services.FileService
	TaskService       *services.TaskService
	ProcessingService *services.ProcessingService

	redisRepo         *repositories.RedisRepository
	rabbitMQPublisher *rabbitmq.RabbitMQPublisher
	rabbitMQConsumer  *rabbitmq.RabbitMQConsumer
	s3Repo            *repositories.S3Repository
}

func NewServiceInjector(
	ctx context.Context,
	cfg *config.Config,
) (*ServiceInjector, error) {
	slog.Debug("initializing S3 repository",
		"endpoint", cfg.S3StorageConfig.Endpoint,
		"bucket", cfg.S3StorageConfig.Bucket,
	)

	s3repository, err := repositories.NewS3Repository(
		ctx,
		&cfg.S3StorageConfig,
	)

	if err != nil {
		slog.Error("failed to create S3 repository",
			"error", err,
			"endpoint", cfg.S3StorageConfig.Endpoint,
		)
		return nil, err
	}

	slog.Debug(
		"ensuring S3 bucket exists",
		"bucket",
		cfg.S3StorageConfig.Bucket,
	)

	if err := s3repository.CreateBucket(ctx); err != nil {
		slog.Error("failed to create S3 bucket",
			"error", err,
			"bucket", cfg.S3StorageConfig.Bucket,
		)
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	slog.Info("S3 storage service initialized successfully",
		"bucket", cfg.S3StorageConfig.Bucket,
	)

	redisRepo, err := repositories.NewRedisRepository(
		ctx,
		&cfg.RedisConfig,
	)
	if err != nil {
		slog.Error("failed to create redis repository",
			"error", err,
		)
		return nil, err
	}

	rabbitmqPublisher, err := rabbitmq.NewRabbitMQPublisher(
		ctx,
		&cfg.RabbitMQConfig,
	)
	if err != nil {
		slog.Error("failed to create redis repository",
			"error", err,
		)
		return nil, err
	}

	taskService := services.NewTaskService(redisRepo, rabbitmqPublisher)
	fileService := services.NewFileService(s3repository, taskService)

	rabbitmqConsumer, err := rabbitmq.NewRabbitMQConsumer(
		ctx,
		&cfg.RabbitMQConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("create rabbitmq consumer: %w", err)
	}

	processingSrv, err := services.NewProcessingService(
		ctx,
		fileService,
		taskService,
		rabbitmqConsumer,
	)
	if err != nil {
		slog.Error("failed to create processing service!",
			"error", err,
		)
		return nil, err
	}

	return &ServiceInjector{
		FileService:       fileService,
		TaskService:       taskService,
		redisRepo:         redisRepo,
		rabbitMQPublisher: rabbitmqPublisher,
		rabbitMQConsumer:  rabbitmqConsumer,
		s3Repo:            s3repository,
		ProcessingService: processingSrv,
	}, nil
}

func (i *ServiceInjector) Shutdown(ctx context.Context) error {
	var errs []error

	if i.rabbitMQPublisher != nil {
		if err := i.rabbitMQPublisher.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if i.rabbitMQConsumer != nil {
		if err := i.rabbitMQConsumer.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if i.redisRepo != nil {
		if err := i.redisRepo.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
