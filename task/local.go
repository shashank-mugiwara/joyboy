package task

type Local struct {
	Name        string
	PortMapping string
	Cpus        float32
	Memory      int64
	ContainerID string
	Owner       string
}
