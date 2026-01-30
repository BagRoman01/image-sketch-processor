package routers

import (
	"github.com/BagRoman01/image-sketch-processor/internal/handlers"
	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/gin-gonic/gin"
)

func RegisterTasksRoutes(
	r *gin.RouterGroup,
	serviceInjector *injectors.ServiceInjector,
) {
	handler := handlers.NewTasksHandler(serviceInjector)

	tasks := r.Group("/tasks")
	{
		tasks.GET("/:task_id/status", handler.GetTaskStatus)
	}
}
