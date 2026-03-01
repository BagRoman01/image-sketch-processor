package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/BagRoman01/image-sketch-processor/internal/messaging/rabbitmq"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	ut "github.com/BagRoman01/image-sketch-processor/pkg/utils"
)

type ProcessingService struct {
	fileService      *FileService
	taskService      *TaskService
	imageProcessor   *ut.ImageProcessor
	rabbitmqConsumer *rabbitmq.RabbitMQConsumer
}

func NewProcessingService(
	ctx context.Context,
	fileService *FileService,
	taskService *TaskService,
	rabbitmqConsumer *rabbitmq.RabbitMQConsumer,
) (*ProcessingService, error) {
	imageProcessor := ut.NewImageProcessor()

	return &ProcessingService{
		imageProcessor:   imageProcessor,
		rabbitmqConsumer: rabbitmqConsumer,
		fileService:      fileService,
		taskService:      taskService,
	}, nil
}

func (w *ProcessingService) Start(ctx context.Context) error {
	return w.rabbitmqConsumer.Consume(ctx, w.processTask)
}

func (w *ProcessingService) processTask(
	ctx context.Context,
	task *models.S3FileTask,
) error {
	slog.Info("processing file task",
		"task_id", task.ID,
		"file_key", task.S3FileInfo.FileKey)

	err := w.taskService.SetTaskProcessing(ctx, task.ID)
	if err != nil {
		return err
	}

	fileData, err := w.fileService.DownloadFile(ctx, task.S3FileInfo.FileKey)
	if err != nil {
		return w.taskService.SetTaskFailed(
			ctx,
			task.ID,
			fmt.Sprintf("download failed: %s", err),
		)
	}

	processedData, err := w.imageProcessor.CreatePencilSketch(ctx, fileData)
	if err != nil {
		return w.taskService.SetTaskFailed(
			ctx,
			task.ID,
			fmt.Sprintf("processing failed: %v", err),
		)
	}

	processedKey, err := w.fileService.UploadProcessedFile(
		ctx,
		task,
		processedData,
	)
	if err != nil {
		return w.taskService.SetTaskFailed(
			ctx,
			task.ID,
			fmt.Sprintf("upload failed: %v", err),
		)
	}

	downloadURL, genErr := w.fileService.GenerateDownloadURL(
		ctx,
		processedKey,
		1*time.Hour,
	)
	if genErr != nil {
		slog.Error(
			"failed to generate download URL",
			"task_id",
			task.ID,
			"error",
			genErr,
		)
	}

	if err := w.taskService.SetTaskCompleted(
		ctx,
		task.ID,
		processedKey,
		downloadURL,
	); err != nil {
		return err
	}

	slog.Info("file processed successfully",
		"task_id", task.ID,
		"input_key", task.S3FileInfo.FileKey,
		"output_key", processedKey)
	return nil
}
