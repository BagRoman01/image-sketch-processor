package handlers

import (
	"net/http"

	"github.com/BagRoman01/image-sketch-processor/internal/services"
	"github.com/gin-gonic/gin"
)

type FilesHandler struct {
	fileSrv *services.FilesService
}

func NewFilesHandler(fileSrv *services.FilesService) *FilesHandler {
	return &FilesHandler{fileSrv: fileSrv}
}

func (h *FilesHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, h.fileSrv.Pong())
}
