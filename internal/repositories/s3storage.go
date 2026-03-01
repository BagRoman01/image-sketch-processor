package repositories

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/config"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Repository struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	uploader      *manager.Uploader
	cfg           *config.S3StorageConfig
}

func NewS3Repository(
	ctx context.Context,
	cfg *config.S3StorageConfig,
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
		u.Concurrency = int(cfg.UploadConcurrency)
	})

	return &S3Repository{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		cfg:           cfg,
		uploader:      uploader,
	}, nil
}

func (s *S3Repository) CreateBucket(ctx context.Context) error {
	logger := logging.LoggerFromContext(ctx)

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
	if err == nil {
		if isAWS {
			if waitErr := s.waitForBucket(ctx); waitErr != nil {
				logger.Warn("bucket created but not immediately accessible",
					"bucket", s.cfg.Bucket, "error", waitErr)
			} else {
				logger.Info("bucket created and accessible", "bucket", s.cfg.Bucket)
			}
		}
		return nil
	}

	var bucketExists *s3Types.BucketAlreadyExists
	var bucketOwnedByYou *s3Types.BucketAlreadyOwnedByYou

	if errors.As(err, &bucketExists) || errors.As(err, &bucketOwnedByYou) {
		if s.isBucketAccessible(ctx) {
			logger.Info("bucket already exists and is accessible",
				"bucket", s.cfg.Bucket)
			return nil
		}

		logger.Warn("bucket exists but not accessible",
			"bucket", s.cfg.Bucket, "error", err)
		return fmt.Errorf(
			"bucket %s exists but inaccessible: %w",
			s.cfg.Bucket,
			err,
		)
	}

	logger.Error("failed to create bucket",
		"bucket", s.cfg.Bucket, "error", err)
	return fmt.Errorf(
		"failed to create/access bucket %s: %w",
		s.cfg.Bucket,
		err,
	)
}

func (s *S3Repository) waitForBucket(ctx context.Context) error {
	logger := logging.LoggerFromContext(ctx)
	logger.Debug(
		"waiting for bucket to become available",
		"bucket",
		s.cfg.Bucket,
	)

	waiter := s3.NewBucketExistsWaiter(s.client)
	waitCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	return waiter.Wait(waitCtx, &s3.HeadBucketInput{
		Bucket: aws.String(s.cfg.Bucket),
	}, time.Minute)
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
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
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
		return nil, fmt.Errorf("failed to upload file %s to S3: %w", key, err)
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

func (s *S3Repository) DownloadFile(
	ctx context.Context,
	key string,
) (io.ReadCloser, *models.Content, error) {
	if key == "" {
		return nil, nil, fmt.Errorf("file key is required")
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to download %q from S3 bucket %q: %w",
			key, s.cfg.Bucket, err,
		)
	}

	contentLength := aws.ToInt64(result.ContentLength)

	content := &models.Content{
		ContentType:   aws.ToString(result.ContentType),
		ContentLength: contentLength,
	}

	return result.Body, content, nil
}

func (s *S3Repository) UploadData(
	ctx context.Context,
	key string,
	data []byte,
	contentType string,
) (*manager.UploadOutput, error) {
	if s.cfg.MaxUploadSize > 0 && int64(len(data)) > s.cfg.MaxUploadSize {
		return nil, fmt.Errorf(
			"size %d exceeds max %d",
			len(data), s.cfg.MaxUploadSize,
		)
	}

	result, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	return result, err
}

func (s *S3Repository) GenerateDownloadURL(
	ctx context.Context,
	fileKey string,
	expiresIn time.Duration,
) (string, error) {
	request, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(fileKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})

	if err != nil {
		return "", fmt.Errorf("failed to presign request: %w", err)
	}

	return request.URL, nil
}
