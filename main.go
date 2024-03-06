package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/shashank-mugiwara/joyboy/task"
	"github.com/shashank-mugiwara/joyboy/worker"
)

func main() {
	os.Setenv("DOCKER_API_VERSION", "1.44")
	db := make(map[uuid.UUID]*task.Task)
	w := worker.Worker{
		Queue: *queue.New(),
		Db:    db,
	}

	t := task.Task{
		ID:    uuid.New(),
		Name:  "test-container-1",
		State: task.Scheduled,
		Image: "nginx:mainline-alpine-perl",
	}

	fmt.Println("starting worker")
	w.AddTask(t)
	result := w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}

	t.ContainerID = result.ContainerId
	fmt.Printf("task %s is running in container %s\n", t.ID, result.ContainerId)
	fmt.Println("sleeping for 30 seconds / letting container run for 30 seconds")
	time.Sleep(time.Second * 30)

	fmt.Println("stopping tasks")
	result = w.StopTask(t)
	if result.Error != nil {
		fmt.Printf("error in stopping the container %v\n", result.Error)
		panic(result.Error)
	}
}
