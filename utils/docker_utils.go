package utils

import (
	"encoding/json"

	"github.com/docker/go-connections/nat"
)

func CreateExposedPorts(portsMap map[string]interface{}) (nat.PortSet, error) {
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

func CreatePortBindings(jsonBindings string) (nat.PortMap, error) {
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
