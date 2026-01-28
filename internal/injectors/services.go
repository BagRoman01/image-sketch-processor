package injectors

import (
	"context"
	"fmt"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/repositories"
	"github.com/BagRoman01/image-sketch-processor/internal/services"
)

type ServiceInjector struct {
	S3storageSrv *services.S3storageService
}

func NewServiceInjector(
	ctx context.Context,
	cfg *config.Config,
) (*ServiceInjector, error) {
	s3repository, err := repositories.NewS3Repository(
		&cfg.S3StorageConfig,
		ctx,
	)

	if err != nil {
		return nil, err
	}

	if err := s3repository.CreateBucket(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	S3storageSrv := services.NewS3storageService(s3repository)

	return &ServiceInjector{S3storageSrv: S3storageSrv}, nil
}
