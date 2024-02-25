package main

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/shashank-mugiwara/joyboy/config"
	"github.com/shashank-mugiwara/joyboy/task"
)

func main() {
	os.Setenv("DOCKER_API_VERSION", "1.44")

	fmt.Printf("Testing create_container")
	dockerTask, createResult := create_container()
	if createResult.Error != nil {
		fmt.Println(createResult.Error.Error())
		os.Exit(1)
	}

	time.Sleep(20 * time.Second)
	fmt.Printf("Stopping container %s\n", createResult.ContainerId)
	stop_container(dockerTask)
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
