package handlers

import (
	"errors"
	"fmt"
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
// @Description  Получить текущий статус задачи по ID
// @Tags         tasks
// @Produce      application/json
// @Param        id  path  string  true  "ID задачи"
// @Success      200  {object}  models.S3FileTask "Задача"
// @Failure      404  {object}  map[string]string "Задача не найдена"
// @Router       /tasks/{id} [get]
func (h *TasksHandler) GetTaskStatus(c *gin.Context) {
	logger := logging.LoggerFromContext(c.Request.Context())
	taskID := c.Param("id")

	task, err := h.TaskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, fmt.Errorf("task not found")) {
			logger.Warn("task not found", "task_id", taskID, "error", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	logger.Info(
		"task status requested",
		"task_id", taskID,
		"status", task.Status,
	)
	c.JSON(http.StatusOK, task)
}
