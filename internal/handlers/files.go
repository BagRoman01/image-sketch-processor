// handlers/files_handler.go
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
	S3storageSrv *services.S3storageService
}

func NewFilesHandler(serviceInjector *injectors.ServiceInjector) *FilesHandler {
	return &FilesHandler{
		S3storageSrv: serviceInjector.S3storageSrv,
	}
}

// uploadFileStreaming godoc
// @Summary      Загрузить файл в S3
// @Description  Загружает файл напрямую в S3 без сохранения на диск
// @Tags         files
// @Accept       multipart/form-data
// @Produce      application/json
// @Param        file formData file true "Файл для загрузки"
// @Success      200  {object}  models.UploadResponse
// @Failure      400
// @Router       /file/streaming [post]
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

	result, key, task, err := h.S3storageSrv.UploadFileStream(
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
		Key:        key,
		Location:   result.Location,
		URL:        h.S3storageSrv.S3Repo.GetFileURL(key),
		Size:       fileHeader.Size,
		TaskID:     task.ID,
		TaskStatus: string(task.Status),
	}

	logger.Info("file uploaded successfully",
		"key", key,
		"task_id", task.ID,
	)

	c.JSON(http.StatusOK, response)
}
