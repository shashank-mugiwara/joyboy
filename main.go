package main

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/shashank-mugiwara/joyboy/config"
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
		Image: "strm/helloworld-http",
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

func create_container() (*task.Docker, *task.DockerResult) {
	containerConfig := config.Config{
		Name:  "postgres-001",
		Image: "postgres:12",
		Env: []string{
			"POSPOSTGRES_USER=joyboy",
			"POSTGRES_PASSWORD=SamplePassword",
		},
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv)
	d := task.Docker{
		Client: dc,
		Config: containerConfig,
	}

	result := d.Run()

	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil, nil
	}

	fmt.Printf("%v", result)

	fmt.Printf("Container %s is running with config %v\n", result.ContainerId, containerConfig)
	return &d, &result
}

func stop_container(d *task.Docker) *task.DockerResult {
	result := d.Stop(d.ContainerId)
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil
	}

	fmt.Printf("Container %s has been stopped or removed\n", result.ContainerId)
	return &result
}
