package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/golang-collections/collections/queue"
	"github.com/shashank-mugiwara/joyboy/database"
	"github.com/shashank-mugiwara/joyboy/task"
	"github.com/shashank-mugiwara/joyboy/utils"
)

type Worker struct {
	Name      string
	Queue     *queue.Queue
	DB        *badger.DB
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
	var marshalledTask string
	err := w.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(t.ID.String()))
		if err != nil {
			log.Println("No entry found for the existing taskId. Considering it as new task with Id: ", t.ID.String())
			return err
		}
		marshalledTask = item.String()
		return err
	})

	var taskPersisted task.Task
	if err == nil {
		taskPersisted = utils.UnmarshallTask(marshalledTask)
	} else {
		taskPersisted = t
		err := w.DB.Update(func(txn *badger.Txn) error {
			err := txn.Set([]byte(taskPersisted.ID.String()), []byte(utils.MarshallTask(taskPersisted)))
			return err
		})

		if err != nil {
			log.Println("Failed to insert task to KV db.")
		}
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

	err = w.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(taskPersisted.ID.String()), []byte(utils.MarshallTask(t)))
		return err
	})
	if err != nil {
		log.Printf("failed to persist task info after running/stopping the task")
	}

	return result
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := t.NewConfig(t)
	d := t.NewDocker(config)

	var marshalledTask string
	err := w.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(t.ID.String()))
		if err != nil {
			log.Println("No entry found for the existing taskId. Considering it as new task with Id: ", t.ID.String())
			return err
		}
		marshalledTask = item.String()
		return err
	})

	if err != nil {
		log.Println("No tasks found running with the given taskId: ", t.ID)
		return task.DockerResult{
			Error: errors.New("no tasks found running with the given task id"),
		}
	}

	runningTask := utils.UnmarshallTask(marshalledTask)
	result := d.Stop(runningTask.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v", t.ContainerID, result.Error)
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed

	var taskPersisted = t
	err = w.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(taskPersisted.ID.String()), []byte(utils.MarshallTask(taskPersisted)))
		return err
	})

	if err != nil {
		log.Println("Failed to insert task to KV db.")
	}

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
		var taskPersisted = t
		err := w.DB.Update(func(txn *badger.Txn) error {
			err := txn.Set([]byte(taskPersisted.ID.String()), []byte(utils.MarshallTask(taskPersisted)))
			return err
		})

		if err != nil {
			log.Println("Failed to insert task to KV db.")
		}
		return result
	}

	t.ContainerID = result.ContainerId
	t.State = task.Running
	var taskPersisted = t
	err := w.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(taskPersisted.ID.String()), []byte(utils.MarshallTask(taskPersisted)))
		return err
	})

	if err != nil {
		log.Println("Failed to insert task to KV db.")
	}

	// Insert entry to db
	err = database.GetDb().Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(t.ID.String()), []byte(utils.MarshallTask(t)))
		return err
	})

	if err != nil {
		log.Println("Failed to insert container details to KV store.")
	}

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
