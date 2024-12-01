package dkrclient

import (
	"encoding/json"
	"log"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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

func ConstructNatPortSet(portsMap map[string]interface{}) (nat.PortSet, error) {
	exposedPorts := make(nat.PortSet)
	for port := range portsMap {
		p, err := nat.NewPort("tcp", port)
		if err != nil {
			return nil, err
		}
		exposedPorts[p] = struct{}{}
	}
	return exposedPorts, nil
}

func ConstructNatPortMap(jsonBindings string) (nat.PortMap, error) {
	var bindings map[string]string
	err := json.Unmarshal([]byte(jsonBindings), &bindings)
	if err != nil {
		return nil, err
	}

	portMap := make(nat.PortMap)
	for containerPort, hostPort := range bindings {
		internalPort, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return nil, err
		}
		portMap[internalPort] = []nat.PortBinding{
			{
				HostPort: hostPort,
			},
		}
	}
	return portMap, nil
}
