package dkrclient

import (
	"log"

	"github.com/docker/docker/client"
)

var dockerClient *client.Client

func InitPlainDockerClient() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %s", err)
	}

	dockerClient = cli
}

func GetPlainDockerClient() *client.Client {
	return dockerClient
}
