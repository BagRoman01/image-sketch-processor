package routers

import (
	"github.com/BagRoman01/image-sketch-processor/internal/handlers"
	"github.com/BagRoman01/image-sketch-processor/internal/services"
	"github.com/gin-gonic/gin"
)

func RegisterPingRoutes(r *gin.RouterGroup, fileSrv *services.FilesService) {
	handler := handlers.NewFilesHandler(fileSrv)
	r.GET("/ping", handler.Ping)
}
