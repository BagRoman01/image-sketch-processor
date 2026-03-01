package handlers

import (
	"net/http"

	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	"github.com/BagRoman01/image-sketch-processor/internal/services"
	"github.com/gin-gonic/gin"
)

type FilesHandler struct {
	FileSrv *services.FileService
}

func NewFilesHandler(serviceInjector *injectors.ServiceInjector) *FilesHandler {
	return &FilesHandler{
		FileSrv: serviceInjector.FileService,
	}
}

// uploadFileStreaming godoc
// @Summary      Создать задачу на обработку изображения
// @Description  Загружает изображение в S3 и создает задачу на .
// @Tags         files
// @Accept       multipart/form-data
// @Produce      application/json
// @Param        file  formData  file  true  "Изображение (JPG, PNG, max 10MB)"
// @Success      200   {object}  models.UploadResponse  "Task создана, файл в S3"
// @Failure      400   {object}  map[string]string      "Неверный файл"
// @Failure      500   {object}  map[string]string      "Ошибка сервера"
// @Router       /files [post]
func (h *FilesHandler) UploadFileStreaming(c *gin.Context) {
	logger := logging.LoggerFromContext(c.Request.Context())

	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Warn("missing file parameter in upload request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file parameter is required",
		})
		return
	}

	logger.Info("starting file upload",
		"file", fileHeader.Filename,
		"size", fileHeader.Size,
		"ct", fileHeader.Header.Get("Content-Type"),
	)

	result, task, err := h.FileSrv.UploadFileStream(
		c.Request.Context(),
		fileHeader,
	)

	if err != nil {
		logger.Error("failed to upload file to S3",
			"error", err,
			"file", fileHeader.Filename,
			"size", fileHeader.Size,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to upload file",
		})
		return
	}

	response := &models.UploadResponse{
		Message:    "File uploaded successfully",
		Key:        task.S3FileInfo.FileKey,
		URL:        result.Location,
		Size:       fileHeader.Size,
		TaskID:     task.ID,
		TaskStatus: string(task.Status),
	}

	logger.Info("file uploaded successfully",
		"key", task.S3FileInfo.FileKey,
		"task_id", task.ID,
	)

	c.JSON(http.StatusOK, response)
}
