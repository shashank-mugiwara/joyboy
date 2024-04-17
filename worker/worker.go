package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/shashank-mugiwara/joyboy/task"
	"gorm.io/gorm"
)

type Worker struct {
	Name      string
	Queue     *queue.Queue
	DB        *gorm.DB
	TaskCount int
}

func (w *Worker) CollectStats() {
	fmt.Println("I will collect stats")
}

func (w *Worker) RunTask() task.DockerResult {
	if w.Queue.Len() == 0 {
		log.Println("No tasks in queue")
		return task.DockerResult{Error: nil}
	}

	t := w.Queue.Dequeue().(task.Task)
	var taskPersisted task.Task
	taskPerResult := w.DB.Where(&task.Task{ID: t.ID}).Find(&taskPersisted)
	if taskPerResult.RowsAffected == 0 || taskPerResult.Error != nil {
		log.Println("Failed to fetch existing task record in database")
		return task.DockerResult{
			Error:   taskPerResult.Error,
			Message: "Failed to fetch task from database.",
		}
	}

	var result task.DockerResult
	if task.ValidStateTransition(taskPersisted.State, t.State) {
		switch t.State {
		case task.Scheduled.String():
			result = w.StartTask(&t)
		case task.Completed.String():
			result = w.StopTask(&t)
		default:
			result.Error = errors.New("we should not run this task")
		}
	} else {
		err := fmt.Errorf("invalid transition from %v to %v", taskPersisted.State, t.State)
		result.Error = err
	}

	t.ContainerID = result.ContainerId
	updatedTask := w.DB.Save(&t)
	if updatedTask.Error != nil {
		log.Println("Failed to update task in DB after stopping it.")
		return task.DockerResult{
			Error:   updatedTask.Error,
			Message: "Please check the container might still be running. You can use 'docker stop [container-id]' to stop and 'docker rm [container-id]'",
		}
	}

	return result
}

func (w *Worker) StopTask(t *task.Task) task.DockerResult {
	config := t.NewConfig(t)
	d := t.NewDocker(config)

	var runningTask task.Task
	getResult := w.DB.First(&runningTask, t.ID)
	if getResult.Error != nil {
		log.Println("No tasks found running with the given taskId: ", t.ID)
		return task.DockerResult{
			Error: errors.New("no tasks found running with the given task id"),
		}
	}

	result := d.Stop(runningTask.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v", t.ContainerID, result.Error)
		return task.DockerResult{
			Error:   result.Error,
			Message: "Please check the container might still be running. You can use 'docker stop [container-id]' to stop and 'docker rm [container-id]'",
		}
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed.String()
	t.ID = runningTask.ID
	updatedTask := w.DB.Save(&t)
	if updatedTask.Error != nil {
		log.Println("Failed to update task in DB after stopping it.")
		return task.DockerResult{
			Error:   updatedTask.Error,
			Message: "Please check the container might still be running. You can use 'docker stop [container-id]' to stop and 'docker rm [container-id]'",
		}
	}

	log.Printf("Stopped and removed the container %v for task %v", t.ContainerID, t.ID)
	result.ContainerId = runningTask.ContainerID
	result.Message = "Container stopped successfully!"
	deleteResult := w.DB.Delete(&t)
	if deleteResult.Error != nil {
		log.Println("Failed to delete task in DB after stopping it.")
		return task.DockerResult{
			Error:   deleteResult.Error,
			Message: "Please check the container might still be running. You can use 'docker stop [container-id]' to stop and 'docker rm [container-id]'",
		}
	}

	return result
}

func (w *Worker) StartTask(t *task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	config := t.NewConfig(t)
	d := t.NewDocker(config)

	result := d.Run()

	if result.Error != nil {
		log.Printf("Error running task %v: %v\n", t.ID, result.Error)
		t.State = task.Failed.String()
		return task.DockerResult{
			Error:       result.Error,
			ContainerId: t.ContainerID,
			Action:      "Failed",
		}

	}

	t.State = task.Running.String()
	return result
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func RunTasks(w Worker) {
	for {
		if w.Queue.Len() != 0 {
			result := w.RunTask()
			if result.Error != nil {
				log.Printf("Error running task: %v", result.Error)
			}
		} else {
			log.Printf("No tasks to process currently.")
		}

		log.Println("Sleeping for 5 seconds.")
		time.Sleep(5 * time.Second)
	}
}
