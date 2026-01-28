package routers

import (
	"github.com/BagRoman01/image-sketch-processor/internal/handlers"
	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/gin-gonic/gin"
)

func RegisterFilesRoutes(
	r *gin.RouterGroup,
	serviceInjector *injectors.ServiceInjector,
) {
	handler := handlers.NewFilesHandler(serviceInjector)

	upload := r.Group("/file")
	{
		upload.POST("/streaming", handler.UploadFileStreaming)

	}
}
