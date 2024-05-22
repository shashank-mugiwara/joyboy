package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/shashank-mugiwara/joyboy/dkrclient"
)

type Scheduler struct {
}

type ContainerStats struct {
	CPU     types.CPUStats                `json:"cpu_stats,omitempty"`
	Memory  types.MemoryStats             `json:"memory_stats,omitempty"`
	Network map[string]types.NetworkStats `json:"networks,omitempty"`
}
type ContainersOnLocal struct {
	ID      string
	Names   []string
	Image   string
	Status  string
	Created int64
	Stats   ContainerStats
}

func InitBackgroundScheduler() {
	scheduler_instance := Scheduler{}
	timer := time.NewTimer(10 * time.Second)
	for {
		<-timer.C
		scheduler_instance.RunningDockerContainersOnMachine()
		timer.Reset(10 * time.Second)
	}
}

func (s *Scheduler) RunningDockerContainersOnMachine() {
	docker_client := dkrclient.GetPlainDockerClient()
	if docker_client == nil {
		log.Printf("Failed to get docker_client instance")
		return
	}

	containers, err := docker_client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		panic(err)
	}

	var containerList []ContainersOnLocal

	for _, c := range containers {
		stats, err := docker_client.ContainerStats(context.Background(), c.ID, false)

		if err != nil {
			panic(err)
		}
		defer stats.Body.Close()

		var containerStats ContainerStats
		if err := json.NewDecoder(stats.Body).Decode(&containerStats); err != nil {
			panic(err)
		}

		container := ContainersOnLocal{
			ID:      c.ID[:10],
			Names:   c.Names,
			Image:   c.Image,
			Status:  c.Status,
			Created: c.Created,
			Stats:   containerStats,
		}
		containerList = append(containerList, container)
	}

	for _, c := range containerList {
		fmt.Printf("ID: %s, Names: %v, Image: %s, Status: %s, Created: %d\n", c.ID, c.Names, c.Image, c.Status, c.Created)
		fmt.Printf("CPU Stats: %+v\n", c.Stats.CPU)
		fmt.Printf("Memory Stats: %+v\n", c.Stats.Memory)
		fmt.Printf("Network Stats: %+v\n", c.Stats.Network)
		fmt.Println()
	}
}
