package repositories

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Repository struct {
	client   *s3.Client
	uploader *manager.Uploader
	cfg      *config.S3StorageConfig
}

func NewS3Repository(
	cfg *config.S3StorageConfig,
	ctx context.Context,
) (*S3Repository, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = true
	})

	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = cfg.ChunkUploadSize
		u.Concurrency = 1
	})

	return &S3Repository{
		client:   client,
		cfg:      cfg,
		uploader: uploader,
	}, nil
}

func (s *S3Repository) CreateBucket(ctx context.Context) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(s.cfg.Bucket),
	}

	isAWS := s.cfg.Endpoint == ""

	if isAWS && s.cfg.Region != "us-east-1" {
		input.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(s.cfg.Region),
		}
	}

	_, err := s.client.CreateBucket(ctx, input)
	if err != nil {
		var bucketExists *s3Types.BucketAlreadyExists
		var bucketOwnedByYou *s3Types.BucketAlreadyOwnedByYou

		if errors.As(err, &bucketExists) {
			if s.isBucketAccessible(ctx) {
				return nil
			}
			return fmt.Errorf(
				"bucket name '%s' already taken by another account: %w",
				s.cfg.Bucket,
				err,
			)
		}

		if errors.As(err, &bucketOwnedByYou) {
			return nil
		}

		return fmt.Errorf("failed to create bucket: %w", err)
	}

	if isAWS {
		waiter := s3.NewBucketExistsWaiter(s.client)
		waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		err = waiter.Wait(waitCtx, &s3.HeadBucketInput{
			Bucket: aws.String(s.cfg.Bucket),
		}, time.Minute)

		if err != nil {
			return fmt.Errorf(
				"bucket '%s' created but not accessible after timeout: %w",
				s.cfg.Bucket,
				err,
			)
		}
	}

	return nil
}

func (s *S3Repository) isBucketAccessible(ctx context.Context) bool {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.cfg.Bucket),
	})
	return err == nil
}

func (s *S3Repository) UploadFileStream(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
	key string,
) (*manager.UploadOutput, error) {
	if s.cfg.MaxUploadSize > 0 && fileHeader.Size > s.cfg.MaxUploadSize {
		return nil, fmt.Errorf(
			"file size %d exceeds maximum allowed size %d",
			fileHeader.Size,
			s.cfg.MaxUploadSize,
		)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	result, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return result, nil
}

func (s *S3Repository) GetFileURL(key string) string {
	if s.cfg.Endpoint != "" {
		return fmt.Sprintf(
			"%s/%s/%s",
			s.cfg.Endpoint,
			s.cfg.Bucket,
			key,
		)
	}
	return fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		s.cfg.Bucket,
		s.cfg.Region,
		key,
	)
}
