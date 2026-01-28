package handlers

import (
	"net/http"

	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/BagRoman01/image-sketch-processor/internal/models"
	"github.com/BagRoman01/image-sketch-processor/internal/services"
	"github.com/gin-gonic/gin"
)

type FilesHandler struct {
	S3storageSrv *services.S3storageService
}

func NewFilesHandler(serviceInjector *injectors.ServiceInjector) *FilesHandler {
	return &FilesHandler{S3storageSrv: serviceInjector.S3storageSrv}
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
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file parameter is required",
		})
		return
	}

	result, key, err := h.S3storageSrv.UploadFileStream(
		c.Request.Context(),
		fileHeader,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to upload file:" + err.Error(),
		})
		return
	}

	response := &models.UploadResponse{
		Message:  "File uploaded successfully",
		Key:      key,
		Location: result.Location,
		URL:      h.S3storageSrv.S3Repo.GetFileURL(key),
		Size:     fileHeader.Size,
	}

	c.JSON(http.StatusOK, response)
}
