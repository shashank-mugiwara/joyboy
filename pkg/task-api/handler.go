package taskapi

import (
	"github.com/labstack/echo/v4"
	"github.com/shashank-mugiwara/joyboy/pkg/logging"
	"github.com/shashank-mugiwara/joyboy/worker"
	"gorm.io/gorm"
)

type Handler struct {
	worker worker.Worker
	DB     *gorm.DB
	logger *logging.Logger
}

func NewHandler(w worker.Worker, db *gorm.DB) *Handler {
	return &Handler{
		worker: w,
		DB:     db,
		logger: logging.GetDefaultLogger(),
	}
}

func (h *Handler) InitRoutes(e *echo.Echo) {
	task_route := e.Group("/api/v1/task")
	task_route.GET("/tasks", h.GetListOfRunningTasks)
	task_route.POST("/add", h.StartTask)
	task_route.POST("/stop", h.StopTask)
	task_route.GET("/:id", h.GetSingleTaskInformation)
}
