package main

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/shashank-mugiwara/joyboy/config"
	"github.com/shashank-mugiwara/joyboy/manager"
	"github.com/shashank-mugiwara/joyboy/node"
	"github.com/shashank-mugiwara/joyboy/task"
	"github.com/shashank-mugiwara/joyboy/worker"
)

func main() {
	os.Setenv("DOCKER_API_VERSION", "1.44")
	t := task.Task{
		ID:     uuid.New(),
		Name:   "Task-1",
		State:  task.Pending,
		Image:  "Image-01",
		Memory: 1024,
		Disk:   1,
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Pending,
		Timestamp: time.Now(),
		Task:      t,
	}

	fmt.Printf("task: %v\n", t)
	fmt.Printf("task event: %v\n", te)

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]task.Task),
	}

	fmt.Printf("worker: %v\n", w)
	w.CollectStats()
	w.RunTask()
	w.StartTask()
	w.StopTask()

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb:  make(map[string][]task.Task),
		EventDb: make(map[string][]task.TaskEvent),
		Workers: []string{w.Name},
	}

	fmt.Printf("manager: %v\n", m)
	m.SelectWorker()
	m.UpdateTasks()
	m.SelectWorker()

	n := node.Node{
		Name:   "Node-1",
		Ip:     "192.168.1.1",
		Cores:  4,
		Memory: 1024,
		Disk:   25,
		Role:   "Worker",
	}

	fmt.Printf("node: %v\n", n)

	fmt.Printf("Testing create_container")
	dockerTask, createResult := create_container()
	if createResult.Error != nil {
		fmt.Println(createResult.Error.Error())
		os.Exit(1)
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("Stopping container %s\n", createResult.ContainerId)
	stop_container(dockerTask)
}

func create_container() (*task.Docker, *task.DockerResult) {
	containerConfig := config.Config{
		Name:  "postgres-container-001-test",
		Image: "postgres:13",
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
