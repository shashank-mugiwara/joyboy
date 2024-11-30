package taskapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shashank-mugiwara/joyboy/task"
	"github.com/shashank-mugiwara/joyboy/utils"
	"gorm.io/gorm"
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

	var existingTask task.Task
	result := h.DB.Where(&task.Task{Name: req.Name, State: "Scheduled"}).Take(&existingTask)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		result.Error = nil
	}

	if result.Error != nil {
		c.Logger().Info("Failed to fetch entries from db. Error is: ", result.Error.Error())
		return c.JSON(http.StatusBadRequest, result.Error)
	}

	if existingTask.Name == req.Name {
		return c.JSON(http.StatusBadRequest, "Container with name: "+req.Name+" is already Scheduled to run. Please wait for the container to start or remove the scheduled container and try again.")
	}

	result = h.DB.Where(&task.Task{Name: req.Name, State: "Running"}).Take(&existingTask)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		result.Error = nil
	}

	if result.Error != nil {
		c.Logger().Info("Failed to fetch entries from db. Error is: ", result.Error.Error())
		return c.JSON(http.StatusBadRequest, result.Error)
	}

	if existingTask.Name == req.Name {
		return c.JSON(http.StatusBadRequest, "Container with name: "+req.Name+" is already running. Please stop this container and try again")
	}

	port_mapping_string, err := json.Marshal(req.PortMapping)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Failed to marshall portMapping. Error is: "+err.Error())
	}

	newTask := task.Task{
		Image:        req.Image,
		Name:         req.Name,
		ID:           uuid.New(),
		State:        task.Scheduled.String(),
		PortBindings: string(port_mapping_string),
		Memory:       req.Resources.Memory,
		Cpus:         req.Resources.Cpus,
	}

	result = h.DB.Save(newTask)
	if result.Error != nil {
		c.Logger().Info("Failed to save entried to db. Error is: ", result.Error.Error())
		return c.JSON(http.StatusBadRequest, result.Error)
	}

	h.worker.AddTask(newTask)
	c.Logger().Info("Task successfully submitted to queue.")

	taskResponse := TaskResponse{
		Image: newTask.Image,
		Name:  newTask.Name,
		ID:    newTask.ID.String(),
		State: newTask.State,
	}
	return c.JSON(http.StatusAccepted, taskResponse)
}

func (h *Handler) StopTask(c echo.Context) error {
	req := TaskRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, errors.New(err.Error()))
	}

	if utils.IsBlank(req.ID) {
		return c.JSON(http.StatusBadRequest, errors.New("id field is required to stop a task"))
	}

	task_id, err := uuid.Parse(req.ID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("given uuid of task is improper"))
	}

	newTask := task.Task{
		ID: task_id,
	}

	result := h.worker.StopTask(&newTask)

	if result.Error != nil && !utils.IsBlank(result.Error.Error()) {
		return c.JSON(http.StatusBadRequest, result)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetListOfRunningTasks(c echo.Context) error {
	var tasks []task.Task
	state := c.QueryParam("state")
	if utils.IsBlank(state) {
		state = "Running"
	}

	_, ok := task.KnownContainerStateMap[state]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid State"})
	}

	h.DB.Where("state = ?", state).First(&tasks)
	return c.JSON(http.StatusOK, tasks)
}

func (h *Handler) GetSingleTaskInformation(c echo.Context) error {
	var runningTask task.Task
	taskId := c.Param("id")
	taskUUID, err := uuid.Parse(taskId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Failed to parse UUID")
	}

	result := h.DB.Where(&task.Task{ID: taskUUID}).Find(&runningTask)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, result.Error)
	}

	if result.RowsAffected == 0 {
		return c.JSON(http.StatusBadRequest, "No running tasks found for the given taskId.")
	}

	return c.JSON(http.StatusOK, runningTask)
}
