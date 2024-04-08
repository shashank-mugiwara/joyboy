package taskapi

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shashank-mugiwara/joyboy/task"
	"github.com/shashank-mugiwara/joyboy/utils"
)

func (h *Handler) StartTask(c echo.Context) error {
	req := TaskRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, errors.New(err.Error()))
	}

	if utils.IsBlank(req.Image) {
		return c.JSON(http.StatusBadRequest, errors.New("image field is mandatory"))
	}

	if utils.IsBlank(req.Name) {
		return c.JSON(http.StatusBadRequest, errors.New("name field is mandatory"))
	}

	newTask := task.Task{
		Image: req.Image,
		Name:  req.Name,
		ID:    uuid.New(),
		State: task.Scheduled,
	}

	h.worker.AddTask(newTask)
	c.Logger().Info("Task successfully submitted to queue.")
	return c.JSON(http.StatusAccepted, newTask)
}
