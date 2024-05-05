package utils

import (
	"encoding/json"

	"github.com/docker/go-connections/nat"
)

func CreateExposedPorts(portsMap map[string]interface{}) (nat.PortSet, error) {
	exposedPorts := make(nat.PortSet)

	// Iterate through the map; the keys are the port specifications.
	for port := range portsMap {
		// Create a Port using the nat package, which requires a protocol ("tcp" assumed here)
		p, err := nat.NewPort("tcp", port) // Adjust protocol as necessary if different.
		if err != nil {
			return nil, err // Return an error if the port format is incorrect.
		}
		exposedPorts[p] = struct{}{} // Add the port to the PortSet.
	}
	return exposedPorts, nil // Return the populated PortSet.
}

func CreatePortBindings(jsonBindings string) (nat.PortMap, error) {
	var bindings map[string]string
	err := json.Unmarshal([]byte(jsonBindings), &bindings)
	if err != nil {
		return nil, err
	}

	portMap := make(nat.PortMap)
	for containerPort, hostPort := range bindings {
		internalPort, err := nat.NewPort("tcp", containerPort) // Assuming tcp; adjust as necessary.
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
