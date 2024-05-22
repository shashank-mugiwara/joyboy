package task

type Local struct {
	Name        string
	PortMapping map[string]string
	Cpu         float64
	Memory      float64
	ContainerID string
	Owner       string
}
