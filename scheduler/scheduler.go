package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
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
	ID           string
	Names        []string
	Image        string
	Status       string
	Created      int64
	Stats        ContainerStats
	PortMappings string
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

		containerJSON, err := docker_client.ContainerInspect(context.Background(), c.ID)
		if err != nil {
			panic(err)
		}

		portMappingsStr := formatPortMappings(containerJSON.NetworkSettings.Ports)

		container := ContainersOnLocal{
			ID:           c.ID,
			Names:        c.Names,
			Image:        c.Image,
			Status:       c.Status,
			Created:      c.Created,
			Stats:        containerStats,
			PortMappings: portMappingsStr,
		}
		containerList = append(containerList, container)
	}

	for _, c := range containerList {
		// Process container information, including port mappings
		log.Printf("Container ID: %s, Port Mappings: %s", c.ID, c.PortMappings)
	}
}

func formatPortMappings(portBindings nat.PortMap) string {
	var portMappings = make(map[string]string)
	for containerPortBinding, hostPortBinding := range portBindings {
		for _, portBinding := range hostPortBinding {
			log.Printf("Container to Host port mapping: %s:%s\n", containerPortBinding.Port(), portBinding.HostPort)
			portMappings[containerPortBinding.Port()] = portBinding.HostPort
		}
	}

	portMappingString, err := json.Marshal(portMappings)
	if err != nil {
		log.Printf("Failed to marshall portMappings. Error is: %v\n", err.Error())
		return "{}"
	}

	return string(portMappingString)
}
