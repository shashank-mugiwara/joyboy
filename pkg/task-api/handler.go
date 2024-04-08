package taskapi

import (
	"github.com/labstack/echo/v4"
	"github.com/shashank-mugiwara/joyboy/worker"
)

type Handler struct {
	worker worker.Worker
}

func NewHandler(w worker.Worker) *Handler {
	return &Handler{
		worker: w,
	}
}

func (h *Handler) InitRoutes(e *echo.Echo) {
	task_route := e.Group("/api/v1/task")
	task_route.POST("/add", h.StartTask)
}
