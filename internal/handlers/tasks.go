// handlers/tasks_handler.go (НОВЫЙ ФАЙЛ)
package handlers

import (
	"net/http"

	"github.com/BagRoman01/image-sketch-processor/internal/injectors"
	"github.com/BagRoman01/image-sketch-processor/internal/logging"
	"github.com/BagRoman01/image-sketch-processor/internal/services"
	"github.com/gin-gonic/gin"
)

type TasksHandler struct {
	TaskService *services.TaskService
}

func NewTasksHandler(
	serviceInjector *injectors.ServiceInjector,
) *TasksHandler {
	return &TasksHandler{
		TaskService: serviceInjector.TaskService,
	}
}

// GetTaskStatus godoc
// @Summary      Получить статус обработки файла
// @Description  Возвращает текущий статус задачи по ID
// @Tags         tasks
// @Produce      application/json
// @Param        task_id  path  string  true  "ID задачи"
// @Success      200  {object}  models.FileTask
// @Failure      404  {object}  map[string]string
// @Router       /tasks/{task_id}/status [get]
func (h *TasksHandler) GetTaskStatus(c *gin.Context) {
	logger := logging.LoggerFromContext(c.Request.Context())
	taskID := c.Param("task_id")

	task, err := h.TaskService.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		logger.Warn("task not found", "task_id", taskID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	logger.Info(
		"task status requested",
		"task_id", taskID,
		"status", task.Status,
	)
	c.JSON(http.StatusOK, task)
}
