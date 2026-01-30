package injectors

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/repositories"
	"github.com/BagRoman01/image-sketch-processor/internal/services"
)

type ServiceInjector struct {
	S3storageSrv *services.S3storageService
	TaskService  *services.TaskService
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
		&cfg.S3StorageConfig,
		ctx,
	)

	if err != nil {
		slog.Error("failed to create S3 repository",
			"error", err,
			"endpoint", cfg.S3StorageConfig.Endpoint,
		)
		return nil, err
	}

	slog.Debug("ensuring S3 bucket exists", "bucket", cfg.S3StorageConfig.Bucket)

	if err := s3repository.CreateBucket(ctx); err != nil {
		slog.Error("failed to create S3 bucket",
			"error", err,
			"bucket", cfg.S3StorageConfig.Bucket,
		)
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	s3storageSrv := services.NewS3storageService(s3repository)

	slog.Info("S3 storage service initialized successfully",
		"bucket", cfg.S3StorageConfig.Bucket,
	)

	redisRepo, err := repositories.NewRedisRepository(
		&cfg.RedisConfig,
		ctx,
	)
	if err != nil {
		slog.Error("failed to create redis repository",
			"error", err,
		)
		return nil, err
	}

	rabbitMQRepo, err := repositories.NewRabbitMQPublisher(&cfg.RabbitMQConfig)
	if err != nil {
		slog.Error("failed to create redis repository",
			"error", err,
		)
		return nil, err
	}

	taskService := services.NewTaskService(redisRepo, rabbitMQRepo)
	s3storageSrv.TaskService = taskService

	return &ServiceInjector{
		S3storageSrv: s3storageSrv,
		TaskService:  taskService,
	}, nil
}
