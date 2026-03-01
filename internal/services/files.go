package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	"github.com/BagRoman01/image-sketch-processor/internal/repositories"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/oklog/ulid/v2"
)

type FileService struct {
	s3Repo      *repositories.S3Repository
	taskService *TaskService
	entropy     *ulid.LockedMonotonicReader
}

func NewFileService(
	s3Repo *repositories.S3Repository,
	taskService *TaskService,
) *FileService {
	return &FileService{
		s3Repo: s3Repo,
		entropy: &ulid.LockedMonotonicReader{
			MonotonicReader: ulid.Monotonic(rand.Reader, 0),
		},
		taskService: taskService,
	}
}

func (s *FileService) UploadFileStream(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
) (*manager.UploadOutput, *models.S3FileTask, error) {
	logger := logging.LoggerFromContext(ctx)

	fileID := ulid.MustNew(ulid.Timestamp(time.Now()), s.entropy).String()
	key := "upload/" + fileID

	logger.Debug(
		"generated S3 key",
		"key", key,
		"original_file", fileHeader.Filename,
	)

	result, err := s.s3Repo.UploadFileStream(ctx, fileHeader, key)
	if err != nil {
		logger.Error("S3 repository upload failed", "error", err, "key", key)
		return nil, nil, err
	}

	content := models.Content{
		ContentLength: fileHeader.Size,
		ContentType:   fileHeader.Header.Get("Content-Type"),
	}

	fileInfo := models.S3FileInfo{
		FileKey: key,
		FileID:  fileID,
		FileInfo: models.FileInfo{
			FileName: fileHeader.Filename,
			Content:  content,
		},
	}

	task, err := s.taskService.CreateFileProcessingTask(
		ctx,
		fileInfo,
	)

	if err != nil {
		logger.Error(
			"failed to create processing task",
			"error", err,
			"key", key,
		)
		return nil, nil, err
	}

	return result, task, nil
}

func (s *FileService) UploadProcessedFile(
	ctx context.Context,
	task *models.S3FileTask,
	data []byte,
) (string, error) {
	logger := logging.LoggerFromContext(ctx)

	processedKey := "processed/" + task.S3FileInfo.FileID

	logger.Debug(
		"uploading processed file",
		"task_id", task.ID,
		"input_key", task.S3FileInfo.FileKey,
		"output_key", processedKey,
	)

	_, err := s.s3Repo.UploadData(
		ctx,
		processedKey,
		data,
		task.S3FileInfo.Content.ContentType,
	)
	if err != nil {
		logger.Error(
			"failed to upload processed file",
			"task_id", task.ID,
			"key", processedKey,
			"error", err,
		)
		return "", fmt.Errorf("upload processed file: %w", err)
	}

	logger.Debug(
		"processed file uploaded successfully",
		"task_id", task.ID,
		"key", processedKey,
		"size", len(data),
	)

	return processedKey, nil
}

func (s *FileService) DownloadFile(
	ctx context.Context,
	key string,
) ([]byte, error) {
	logger := logging.LoggerFromContext(ctx)
	logger.Debug("downloading file from S3", "key", key)

	fileData, _, err := s.s3Repo.DownloadFile(ctx, key)
	if err != nil {
		logger.Error("failed to download file from S3",
			"key", key,
			"error", err)
		return nil, fmt.Errorf("download file %q: %w", key, err)
	}
	defer fileData.Close()

	data, err := io.ReadAll(fileData)
	if err != nil {
		logger.Error("failed to read S3 file data",
			"key", key,
			"error", err)
		return nil, fmt.Errorf("read S3 file %q: %w", key, err)
	}

	logger.Info("file downloaded successfully",
		"key", key,
		"size_bytes", len(data))
	return data, nil
}

func (s *FileService) GenerateDownloadURL(
	ctx context.Context,
	key string,
	expiry time.Duration,
) (string, error) {
	logger := logging.LoggerFromContext(ctx)

	url, err := s.s3Repo.GenerateDownloadURL(ctx, key, expiry)
	if err != nil {
		logger.Error(
			"failed to generate download URL",
			"key", key,
			"error", err,
		)
		return "", fmt.Errorf("generate download URL for %q: %w", key, err)
	}

	logger.Info(
		"presigned URL generated successfully",
		"key", key,
		"url", url[:50]+"...",
		"expiry", expiry,
	)

	return url, nil
}
