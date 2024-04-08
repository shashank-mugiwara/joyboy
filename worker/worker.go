package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/shashank-mugiwara/joyboy/task"
)

type Worker struct {
	Name      string
	Queue     *queue.Queue
	Db        map[uuid.UUID]*task.Task
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

	taskPersisted := w.Db[t.ID]
	if taskPersisted == nil {
		taskPersisted = &t
		w.Db[t.ID] = &t
	}

	var result task.DockerResult
	if task.ValidStateTransition(taskPersisted.State, t.State) {
		switch t.State {
		case task.Scheduled:
			result = w.StartTask(t)
		case task.Completed:
			result = w.StopTask(t)
		default:
			result.Error = errors.New("we should not run this task")
		}
	} else {
		err := fmt.Errorf("invalid transition from %v to %v", taskPersisted.State, t.State)
		result.Error = err
	}

	return result
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := t.NewConfig(t)
	d := t.NewDocker(config)

	result := d.Stop(t.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v", t.ContainerID, result.Error)
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed the container %v for task %v", t.ContainerID, t.ID)
	return result
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	config := t.NewConfig(t)
	d := t.NewDocker(config)

	result := d.Run()

	if result.Error != nil {
		log.Printf("Error running task %v: %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return result
	}

	t.ContainerID = result.ContainerId
	t.State = task.Running
	w.Db[t.ID] = &t
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
